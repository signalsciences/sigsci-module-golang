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
	config    *ModuleConfig
	handler   http.Handler
	inspector Inspector
	inspInit  InspectorInitFunc
	inspFini  InspectorFiniFunc
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
		handler:   h,
		config:    config,
		inspector: config.Inspector(),
		inspInit:  config.InspectorInit(),
		inspFini:  config.InspectorFini(),
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
		ServerFlavor:  m.config.ServerFlavor(),
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
		// Continue with normal request
		m.handler.ServeHTTP(rw, req)
	case m.config.IsBlockCode(int(wafresponse)):
		status := int(wafresponse)

		// Only redirect if it is a redirect status (3xx) AND there is a redirect URL
		if status >= 300 && status <= 399 {
			redirect := req.Header.Get("X-Sigsci-Redirect")
			if len(redirect) > 0 {
				http.Redirect(rw, req, redirect, status)
				break
			}
		}

		// Block
		http.Error(rw, fmt.Sprintf("%d %s\n", status, http.StatusText(status)), status)
	default:
		log.Printf("ERROR: Received invalid response code from inspector (failing open): %d", wafresponse)
		// Continue with normal request
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
		inspin := NewRPCMsgIn(m.config, req, nil, code, size, duration)
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

	inspin := NewRPCMsgIn(m.config, req, reqbody, -1, -1, 0)

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

	if out.RequestID != "" {
		req.Header.Set("X-Sigsci-Requestid", out.RequestID)
	} else {
		req.Header.Del("X-Sigsci-Requestid")
	}

	wafresponse := out.WAFResponse
	req.Header.Set("X-Sigsci-Agentresponse", strconv.Itoa(int(wafresponse)))

	// Add request headers from the WAF response to the request
	req.Header.Del("X-Sigsci-Tags")
	req.Header.Del("X-Sigsci-Redirect")
	for _, kv := range out.RequestHeaders {
		// For X-Sigsci-* headers, use Set to override, but for custom headers, use Add to append
		if strings.HasPrefix(http.CanonicalHeaderKey(kv[0]), "X-Sigsci-") {
			req.Header.Set(kv[0], kv[1])
		} else {
			req.Header.Add(kv[0], kv[1])
		}
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
func NewRPCMsgIn(mcfg *ModuleConfig, r *http.Request, postbody []byte, code int, size int64, dur time.Duration) *RPCMsgIn {
	now := time.Now()

	msgIn := RPCMsgIn{
		ModuleVersion:  mcfg.ModuleIdentifier(),
		ServerVersion:  mcfg.ServerIdentifier(),
		ServerFlavor:   mcfg.ServerFlavor(),
		ServerName:     r.Host,
		Timestamp:      now.Unix(),
		NowMillis:      now.UnixMilli(),
		RemoteAddr:     stripPort(r.RemoteAddr),
		Method:         r.Method,
		URI:            r.RequestURI,
		Protocol:       r.Proto,
		ResponseCode:   int32(code),
		ResponseMillis: dur.Milliseconds(),
		ResponseSize:   size,
		PostBody:       string(postbody),
	}

	if r.TLS != nil {
		// convert golang/spec integers into something human readable
		msgIn.Scheme = "https"
		msgIn.TLSProtocol = tlstext.Version(r.TLS.Version)
		msgIn.TLSCipher = tlstext.CipherSuite(r.TLS.CipherSuite)
	} else {
		msgIn.Scheme = "http"
	}

	if hdrs := mcfg.RawHeaderExtractor(); hdrs != nil {
		msgIn.HeadersIn = hdrs(r)
	}
	if msgIn.HeadersIn == nil {
		msgIn.HeadersIn = requestHeader(r)
	}
	return &msgIn
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

	if m.config.extendContentTypes {
		return true
	}

	// only read certain types of content
	if inspectableContentType(req.Header.Get("Content-Type")) {
		return true
	}

	// read custom configured content type(s)
	if m.config.IsExpectedContentType(req.Header.Get("Content-Type")) {
		return true
	}

	// read the body if there are multiple Content-Type headers
	if len(req.Header.Values("Content-Type")) > 1 {
		return true
	}

	// Check for comma separated Content-Types
	if len(strings.SplitN(req.Header.Get("Content-Type"), ",", 2)) > 1 {
		return true
	}

	return false
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

	// GraphQL
	case strings.HasPrefix(s, "application/graphql"):
		return true

	// No type provided
	case s == "":
		return true
	}

	return false
}

// requestHeader returns request headers with host header
func requestHeader(r *http.Request) [][2]string {
	out := make([][2]string, 0, len(r.Header)+1)
	// golang removes Host header from req.Header map and
	// promotes it to r.Host field. Add it back as the first header.
	if len(r.Host) > 0 {
		out = append(out, [2]string{"Host", r.Host})
	}
	for key, values := range r.Header {
		for _, value := range values {
			out = append(out, [2]string{key, value})
		}
	}
	return out
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
