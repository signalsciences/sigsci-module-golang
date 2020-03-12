package sigsci

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/signalsciences/tlstext"
)

// Module is an http.Handler that wraps an existing handler with
// data collection and sends it to the Signal Sciences Agent for
// inspection.
type Module struct {
	config          *ModuleConfig
	handler         http.Handler
	inspector       Inspector
	inspInit        InspectorInitFunc
	inspFini        InspectorFiniFunc
	headerExtractor HeaderExtractorFunc
}

// NewModule wraps an existing http.Handler with one that extracts data and
// sends it to the Signal Sciences Agent for inspection. The module is configured
// via functional options.
func NewModule(h http.Handler, options ...ModuleConfigOption) (*Module, error) {
	// Configure
	config, err := NewModuleConfig(options...)
	if err != nil {
		return nil, err
	}

	// The following are the defaults, overridden by passing in functional options
	m := Module{
		handler:         h,
		config:          config,
		inspector:       config.Inspector(),
		inspInit:        config.InspectorInit(),
		inspFini:        config.InspectorFini(),
		headerExtractor: config.HeaderExtractor(),
	}

	// By default, use an RPC based inspector if not configured externally
	if m.inspector == nil {
		m.inspector = &RPCInspector{
			Network: m.config.RPCNetwork(),
			Address: m.config.RPCAddress(),
			Timeout: m.config.Timeout(),
			Debug:   m.config.Debug(),
		}
	}

	// Call ModuleInit to initialize the module data, so that the agent is
	// registered on module creation
	now := time.Now()
	in := RPCMsgIn{
		ModuleVersion: m.config.ModuleIdentifier(),
		ServerVersion: m.config.ServerIdentifier(),
		ServerFlavor:  "",
		Timestamp:     now.Unix(),
		NowMillis:     now.UnixNano() / 1e6,
	}
	out := RPCMsgOut{}
	if err := m.inspector.ModuleInit(&in, &out); err != nil {
		if m.config.Debug() {
			log.Println("Error in moduleinit to inspector: ", err.Error())
		}
	}

	return &m, nil
}

// Version returns a SemVer version string
func Version() string {
	return version
}

// ServeHTTP satisfies the http.Handler interface
func (m *Module) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	start := time.Now()
	finiwg := sync.WaitGroup{}

	// Use the inspector init/fini functions if available
	if m.inspInit != nil && !m.inspInit(req) {
		// No inspection is desired, so just defer to the underlying handler
		m.handler.ServeHTTP(w, req)
		return
	}
	if m.inspFini != nil {
		defer func() {
			// Delay the finalizer call until inspection (any pending Post
			// or Update call) is complete
			go func() {
				finiwg.Wait()
				m.inspFini(req)
			}()
		}()
	}

	if m.config.Debug() {
		log.Printf("DEBUG: calling 'RPC.PreRequest' for inspection: method=%s host=%s url=%s", req.Method, req.Host, req.URL)
	}
	inspin2, out, err := m.inspectorPreRequest(req)
	if err != nil {
		// Fail open
		if m.config.Debug() {
			log.Printf("ERROR: 'RPC.PreRequest' call failed (failing open): %s", err.Error())
		}
		m.handler.ServeHTTP(w, req)
		return
	}

	rw := NewResponseWriter(w)

	wafresponse := out.WAFResponse
	switch {
	case m.config.IsAllowCode(int(wafresponse)):
		// continue with normal request
		m.handler.ServeHTTP(rw, req)
	case m.config.IsBlockCode(int(wafresponse)):
		status := int(wafresponse)
		http.Error(rw, fmt.Sprintf("%d %s\n", status, http.StatusText(status)), status)
	default:
		log.Printf("ERROR: Received invalid response code from inspector (failing open): %d", wafresponse)
		// continue with normal request
		m.handler.ServeHTTP(rw, req)
	}

	duration := time.Since(start)
	code := rw.StatusCode()
	size := rw.BytesWritten()

	if len(inspin2.RequestID) > 0 {
		// Do the UpdateRequest inspection in the background while the foreground hurries the response back to the end-user.
		inspin2.ResponseCode = int32(code)
		inspin2.ResponseSize = size
		inspin2.ResponseMillis = int64(duration / time.Millisecond)
		inspin2.HeadersOut = convertHeaders(rw.Header())
		if m.config.Debug() {
			log.Printf("DEBUG: calling 'RPC.UpdateRequest' due to returned requestid=%s: method=%s host=%s url=%s code=%d size=%d duration=%s", inspin2.RequestID, req.Method, req.Host, req.URL, code, size, duration)
		}
		finiwg.Add(1) // Inspection finializer will wait for this goroutine
		go func() {
			defer finiwg.Done()
			if err := m.inspectorUpdateRequest(inspin2); err != nil && m.config.Debug() {
				log.Printf("ERROR: 'RPC.UpdateRequest' call failed: %s", err.Error())
			}
		}()
	} else if code >= 300 || size >= m.config.AnomalySize() || duration >= m.config.AnomalyDuration() {
		// Do the PostRequest inspection in the background while the foreground hurries the response back to the end-user.
		if m.config.Debug() {
			log.Printf("DEBUG: calling 'RPC.PostRequest' due to anomaly: method=%s host=%s url=%s code=%d size=%d duration=%s", req.Method, req.Host, req.URL, code, size, duration)
		}
		inspin := NewRPCMsgIn(req, nil, code, size, duration, m.config.ModuleIdentifier(), m.config.ServerIdentifier())
		m.extractHeaders(req, inspin)
		inspin.WAFResponse = wafresponse
		inspin.HeadersOut = convertHeaders(rw.Header())

		finiwg.Add(1) // Inspection finializer will wait for this goroutine
		go func() {
			defer finiwg.Done()
			if err := m.inspectorPostRequest(inspin); err != nil && m.config.Debug() {
				log.Printf("ERROR: 'RPC.PostRequest' call failed: %s", err.Error())
			}
		}()
	}
}

// Inspector returns the configured inspector
func (m *Module) Inspector() Inspector {
	return m.inspector
}

// Version returns the module version string
func (m *Module) Version() string {
	return m.config.ModuleIdentifier()
}

// ServerVersion returns the server version string
func (m *Module) ServerVersion() string {
	return m.config.ServerIdentifier()
}

// ModuleConfig returns the module configuration
func (m *Module) ModuleConfig() *ModuleConfig {
	return m.config
}

// inspectorPreRequest reads the body if required and makes a prerequest call to the inspector
func (m *Module) inspectorPreRequest(req *http.Request) (inspin2 RPCMsgIn2, out RPCMsgOut, err error) {
	// Create message to the inspector from the input request
	// see if we can read-in the post body

	var reqbody []byte
	if shouldReadBody(req, m) {
		// Read all of it and close
		// if error, just keep going
		// It's possible that it is an error event
		// but not sure what it is. Likely
		// the client disconnected.
		reqbody, _ = ioutil.ReadAll(req.Body)
		req.Body.Close()

		// make a new reader so the next handler
		// can still read the post normally as if
		// nothing happened
		req.Body = ioutil.NopCloser(bytes.NewBuffer(reqbody))
	}

	inspin := NewRPCMsgIn(req, reqbody, -1, -1, -1, m.config.ModuleIdentifier(), m.config.ServerIdentifier())
	m.extractHeaders(req, inspin)

	if m.config.Debug() {
		log.Printf("DEBUG: Making PreRequest call to inspector: %s %s", inspin.Method, inspin.URI)
	}

	err = m.inspector.PreRequest(inspin, &out)
	if err != nil {
		if m.config.Debug() {
			log.Printf("DEBUG: PreRequest call error (%s %s): %s", inspin.Method, inspin.URI, err)
		}
		return
	}

	// set any request headers
	if out.RequestID != "" {
		req.Header.Add("X-Sigsci-Requestid", out.RequestID)
	}

	wafresponse := out.WAFResponse
	req.Header.Add("X-Sigsci-Agentresponse", strconv.Itoa(int(wafresponse)))
	for _, kv := range out.RequestHeaders {
		req.Header.Add(kv[0], kv[1])
	}

	inspin2 = RPCMsgIn2{
		RequestID:      out.RequestID,
		ResponseCode:   -1,
		ResponseMillis: -1,
		ResponseSize:   -1,
	}

	if m.config.Debug() {
		tags := req.Header.Get("X-Sigsci-Tags")
		log.Printf("DEBUG: PreRequest call (%s %s): %d RequestID=%s Tags=%v", inspin.Method, inspin.URI, wafresponse, out.RequestID, tags)
	}

	return
}

func (m *Module) extractHeaders(req *http.Request, inspin *RPCMsgIn) {
	// If the user supplied a custom header extractor, use it to unpack the
	// headers. If there no custom header extractor or it returns an error,
	// fallback to the native headers on the request.
	if m.headerExtractor != nil {
		hin, err := m.headerExtractor(req)
		if err == nil {
			inspin.HeadersIn = convertHeaders(hin)
		} else if m.config.Debug() {
			log.Printf("DEBUG: Error extracting custom headers, using native headers: %s", err)
		}
	}
}

// inspectorPostRequest makes a postrequest call to the inspector
func (m *Module) inspectorPostRequest(inspin *RPCMsgIn) error {
	// Create message to agent from the input request

	if m.config.Debug() {
		log.Printf("DEBUG: Making PostRequest call to inspector: %s %s", inspin.Method, inspin.URI)
	}

	// NOTE: Currently the output argument is not used
	err := m.inspector.PostRequest(inspin, &RPCMsgOut{})
	if err != nil {
		if m.config.Debug() {
			log.Printf("DEBUG: PostRequest call error (%s %s): %s", inspin.Method, inspin.URI, err)
		}
	}

	return err
}

// inspectorUpdateRequest makes an updaterequest call to the inspector
func (m *Module) inspectorUpdateRequest(inspin RPCMsgIn2) error {
	if m.config.Debug() {
		log.Printf("DEBUG: Making UpdateRequest call to inspector: RequestID=%s", inspin.RequestID)
	}

	// NOTE: Currently the output argument is not used
	err := m.inspector.UpdateRequest(&inspin, &RPCMsgOut{})
	if err != nil {
		if m.config.Debug() {
			log.Printf("DEBUG: UpdateRequest call error (RequestID=%s): %s", inspin.RequestID, err)
		}
	}

	return err
}

// NewRPCMsgIn creates a message from a go http.Request object
// End-users of the golang module never need to use this
// directly and it is only exposed for performance testing
func NewRPCMsgIn(r *http.Request, postbody []byte, code int, size int64, dur time.Duration, module, server string) *RPCMsgIn {
	now := time.Now()

	// assemble a message to send to inspector
	tlsProtocol := ""
	tlsCipher := ""
	scheme := "http"
	if r.TLS != nil {
		// convert golang/spec integers into something human readable
		scheme = "https"
		tlsProtocol = tlstext.Version(r.TLS.Version)
		tlsCipher = tlstext.CipherSuite(r.TLS.CipherSuite)
	}

	// golang removes Host header from req.Header map and
	// promotes it to r.Host field. Add it back as the first header.
	hin := convertHeaders(r.Header)
	if len(r.Host) > 0 {
		hin = append([][2]string{{"Host", r.Host}}, hin...)
	}

	return &RPCMsgIn{
		ModuleVersion:  module,
		ServerVersion:  server,
		ServerName:     r.Host,
		Timestamp:      now.Unix(),
		NowMillis:      now.UnixNano() / 1e6,
		RemoteAddr:     stripPort(r.RemoteAddr),
		Method:         r.Method,
		Scheme:         scheme,
		URI:            r.RequestURI,
		Protocol:       r.Proto,
		TLSProtocol:    tlsProtocol,
		TLSCipher:      tlsCipher,
		ResponseCode:   int32(code),
		ResponseMillis: int64(dur / time.Millisecond),
		ResponseSize:   size,
		PostBody:       string(postbody),
		HeadersIn:      hin,
	}
}

// stripPort removes any port from an address (e.g., the client port from the RemoteAddr)
func stripPort(ipdots string) string {
	host, _, err := net.SplitHostPort(ipdots)
	if err != nil {
		return ipdots
	}
	return host
}

// shouldReadBody returns true if the body should be read
func shouldReadBody(req *http.Request, m *Module) bool {
	// nothing to do
	if req.Body == nil {
		return false
	}

	// A ContentLength of -1 is an unknown length (streamed) and is only
	// allowed if explicitly configured. In this case the max content length
	// check is bypassed.
	if !(m.config.AllowUnknownContentLength() && req.ContentLength == -1) {
		// skip reading if post is invalid or too long
		if req.ContentLength <= 0 || req.ContentLength > m.config.MaxContentLength() {
			return false
		}
	}

	// only read certain types of content
	return inspectableContentType(req.Header.Get("Content-Type"))
}

// inspectableContentType returns true for an inspectable content type
func inspectableContentType(s string) bool {
	s = strings.ToLower(s)
	switch {

	// Form
	case strings.HasPrefix(s, "application/x-www-form-urlencoded"):
		return true
	case strings.HasPrefix(s, "multipart/form-data"):
		return true

	// JSON
	case strings.Contains(s, "json") ||
		strings.Contains(s, "javascript"):
		return true

	// XML
	case strings.HasPrefix(s, "text/xml") ||
		strings.HasPrefix(s, "application/xml") ||
		strings.Contains(s, "+xml"):
		return true

	// gRPC (protobuf)
	case strings.HasPrefix(s, "application/grpc"):
		return true
	}

	return false
}

// converts a http.Header map to a [][2]string
func convertHeaders(h http.Header) [][2]string {
	// get headers
	out := make([][2]string, 0, len(h)+1)

	for key, values := range h {
		for _, value := range values {
			out = append(out, [2]string{key, value})
		}
	}
	return out
}
