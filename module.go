package sigsci

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/signalsciences/tlstext"
)

const moduleVersion = "sigsci-module-golang " + version

// Module is an http.Handler that wraps an existing handler with
// data collection and sends it to the Signal Sciences Agent for
// inspection.
type Module struct {
	handler          http.Handler
	rpcNetwork       string
	rpcAddress       string
	debug            bool
	timeout          time.Duration
	anomalySize      int64
	anomalyDuration  time.Duration
	maxContentLength int64
	moduleVersion    string
	serverVersion    string
	inspector        Inspector
	inspInit         InspectorInitFunc
	inspFini         InspectorFiniFunc
}

// ModuleConfigOption is a functional config option for configuring the module
// See: https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis
type ModuleConfigOption func(*Module) error

// NewModule wraps an existing http.Handler with one that extracts data and
// sends it to the Signal Sciences Agent for inspection. The module is configured
// via functional options.
func NewModule(h http.Handler, options ...ModuleConfigOption) (*Module, error) {
	// The following are the defaults, overridden by passing in functional options
	m := Module{
		handler:          h,
		rpcNetwork:       "unix",
		rpcAddress:       "/var/run/sigsci.sock",
		debug:            false,
		timeout:          100 * time.Millisecond,
		anomalySize:      512 * 1024,
		anomalyDuration:  1 * time.Second,
		maxContentLength: 100000,
		moduleVersion:    moduleVersion,
		serverVersion:    runtime.Version(),
	}

	// Override defaults from functional options
	for _, opt := range options {
		err := opt(&m)
		if err != nil {
			return nil, err
		}
	}

	// By default, use an RPC based inspector
	if m.inspector == nil {
		m.inspector = &RPCInspector{
			Network: m.rpcNetwork,
			Address: m.rpcAddress,
			Timeout: m.timeout,
			Debug:   m.debug,
		}
	}

	// Call ModuleInit to initialize the module data, so that the agent is
	// registered on module creation
	now := time.Now().UTC()
	in := RPCMsgIn{
		ModuleVersion: m.moduleVersion,
		ServerVersion: m.serverVersion,
		ServerFlavor:  "",
		Timestamp:     now.Unix(),
		NowMillis:     now.UnixNano() / 1e6,
	}
	out := RPCMsgOut{}
	if err := m.inspector.ModuleInit(&in, &out); err != nil {
		if m.debug {
			log.Println("Error in moduleinit to inspector: ", err.Error())
		}
	}

	return &m, nil
}

// Version returns a SemVer version string
func Version() string {
	return version
}

// Debug turns on debug logging
func Debug(enable bool) ModuleConfigOption {
	return func(m *Module) error {
		m.debug = enable
		return nil
	}
}

// Socket is a function argument to set where to send data to the
// Signal Sciences Agent. The network argument should be `unix`
// or `tcp` and the corresponding address should be either an absolute
// path or an `address:port`, respectively.
func Socket(network, address string) ModuleConfigOption {
	return func(m *Module) error {
		switch network {
		case "unix":
			if !filepath.IsAbs(address) {
				return errors.New(`address must be an absolute path for network="unix"`)
			}
		case "tcp":
			if _, _, err := net.SplitHostPort(address); err != nil {
				return fmt.Errorf(`address must be in "address:port" form for network="tcp": %s`, err)
			}
		default:
			return errors.New(`network must be "tcp" or "unix"`)
		}

		m.rpcNetwork = network
		m.rpcAddress = address

		return nil
	}
}

// AnomalySize is a function argument to indicate when to send data to
// the inspector if the response was abnormally large
func AnomalySize(size int64) ModuleConfigOption {
	return func(m *Module) error {
		m.anomalySize = size
		return nil
	}
}

// AnomalyDuration is a function argument to indicate when to send data
// to the inspector if the response was abnormally slow
func AnomalyDuration(dur time.Duration) ModuleConfigOption {
	return func(m *Module) error {
		m.anomalyDuration = dur
		return nil
	}
}

// MaxContentLength is a function argument to set the maximum post
// body length that will be processed
func MaxContentLength(size int64) ModuleConfigOption {
	return func(m *Module) error {
		m.maxContentLength = size
		return nil
	}
}

// Timeout is a function argument that sets the maximum time to wait until
// receiving a reply from the inspector. Once this timeout is reached, the
// module will fail open.
func Timeout(t time.Duration) ModuleConfigOption {
	return func(m *Module) error {
		m.timeout = t
		return nil
	}
}

// ModuleIdentifier is a function argument that sets the module name
// and version for custom setups.
// The version should be a sem-version (e.g., "1.2.3")
func ModuleIdentifier(name, version string) ModuleConfigOption {
	return func(m *Module) error {
		m.moduleVersion = name + " " + version
		return nil
	}
}

// ServerIdentifier is a function argument that sets the server
// identifier for custom setups
func ServerIdentifier(id string) ModuleConfigOption {
	return func(m *Module) error {
		m.serverVersion = id
		return nil
	}
}

// CustomInspector is a function argument that sets a custom inspector,
// an optional inspector initializer to decide if inspection should occur, and
// an optional inspector finalizer that can perform any post-inspection steps
func CustomInspector(insp Inspector, init InspectorInitFunc, fini InspectorFiniFunc) ModuleConfigOption {
	return func(m *Module) error {
		m.inspector = insp
		m.inspInit = init
		m.inspFini = fini
		return nil
	}
}

// ServeHTTP satisfies the http.Handler interface
func (m *Module) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	start := time.Now().UTC()

	// Use the inspector init/fini functions if available
	if m.inspInit != nil && !m.inspInit(req) {
		// No inspection is desired, so just defer to the underlying handler
		m.handler.ServeHTTP(w, req)
		return
	}
	if m.inspFini != nil {
		defer m.inspFini(req)
	}

	if m.debug {
		log.Printf("DEBUG: calling 'RPC.PreRequest' for inspection: method=%s host=%s url=%s", req.Method, req.Host, req.URL)
	}
	inspin2, out, err := m.inspectorPreRequest(req)
	if err != nil {
		// Fail open
		if m.debug {
			log.Printf("ERROR: 'RPC.PreRequest' call failed (failing open): %s", err.Error())
		}
		m.handler.ServeHTTP(w, req)
		return
	}

	rw := NewResponseWriter(w)

	wafresponse := out.WAFResponse
	switch wafresponse {
	case 406:
		status := int(wafresponse)
		http.Error(rw, fmt.Sprintf("%d %s\n", status, http.StatusText(status)), status)
	case 200:
		// continue with normal request
		m.handler.ServeHTTP(rw, req)
	default:
		log.Printf("ERROR: Received invalid response code from inspector (failing open): %d", wafresponse)
		// continue with normal request
		m.handler.ServeHTTP(rw, req)
	}

	duration := time.Since(start)
	code := rw.StatusCode()
	size := rw.BytesWritten()

	if inspin2.RequestID != "" {
		// Do the UpdateRequest inspection in the background while the foreground hurries the response back to the end-user.
		inspin2.ResponseCode = int32(code)
		inspin2.ResponseSize = size
		inspin2.ResponseMillis = int64(duration / time.Millisecond)
		inspin2.HeadersOut = convertHeaders(rw.Header())
		if m.debug {
			log.Printf("DEBUG: calling 'RPC.UpdateRequest' due to returned requestid=%s: method=%s host=%s url=%s code=%d size=%d duration=%s", inspin2.RequestID, req.Method, req.Host, req.URL, code, size, duration)
		}
		go func() {
			if err := m.inspectorUpdateRequest(inspin2); err != nil && m.debug {
				log.Printf("ERROR: 'RPC.UpdateRequest' call failed: %s", err.Error())
			}
		}()
	} else if code >= 300 || size >= m.anomalySize || duration >= m.anomalyDuration {
		// Do the PostRequest inspection in the background while the foreground hurries the response back to the end-user.
		if m.debug {
			log.Printf("DEBUG: calling 'RPC.PostRequest' due to anomaly: method=%s host=%s url=%s code=%d size=%d duration=%s", req.Method, req.Host, req.URL, code, size, duration)
		}
		inspin := NewRPCMsgIn(req, nil, code, size, duration, m.moduleVersion, m.serverVersion)
		inspin.WAFResponse = wafresponse
		inspin.HeadersOut = convertHeaders(rw.Header())

		go func() {
			if err := m.inspectorPostRequest(inspin); err != nil && m.debug {
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
	return m.moduleVersion
}

// ServerVersion returns the server version string
func (m *Module) ServerVersion() string {
	return m.serverVersion
}

// inspectorPreRequest makes a prerequest call to the inspector
func (m *Module) inspectorPreRequest(req *http.Request) (inspin2 RPCMsgIn2, out RPCMsgOut, err error) {
	// Create message to the inspector from the input request
	// see if we can read-in the post body

	var postbody []byte
	if shouldReadBody(req, m) {
		// Read all of it and close
		// if error, just keep going
		// It's possible that it is an error event
		// but not sure what it is. Likely
		// the client disconnected.
		postbody, _ = ioutil.ReadAll(req.Body)
		req.Body.Close()

		// make a new reader so the next handler
		// can still read the post normally as if
		// nothing happened
		req.Body = ioutil.NopCloser(bytes.NewBuffer(postbody))
	}

	inspin := NewRPCMsgIn(req, postbody, -1, -1, -1, m.moduleVersion, m.serverVersion)

	if m.debug {
		log.Printf("DEBUG: Making PreRequest call to inspector: %s %s", inspin.Method, inspin.URI)
	}

	err = m.inspector.PreRequest(inspin, &out)
	if err != nil {
		if m.debug {
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

	if m.debug {
		tags := req.Header.Get("X-Sigsci-Tags")
		log.Printf("DEBUG: PreRequest call (%s %s): %d RequestID=%s Tags=%v", inspin.Method, inspin.URI, wafresponse, out.RequestID, tags)
	}

	return
}

// inspectorPostRequest makes a postrequest call to the inspector
func (m *Module) inspectorPostRequest(inspin *RPCMsgIn) error {
	// Create message to agent from the input request

	if m.debug {
		log.Printf("DEBUG: Making PostRequest call to inspector: %s %s", inspin.Method, inspin.URI)
	}

	// NOTE: Currently the output argument is not used
	err := m.inspector.PostRequest(inspin, &RPCMsgOut{})
	if err != nil {
		if m.debug {
			log.Printf("DEBUG: PostRequest call error (%s %s): %s", inspin.Method, inspin.URI, err)
		}
	}

	return err
}

// inspectorUpdateRequest makes an updaterequest call to the inspector
func (m *Module) inspectorUpdateRequest(inspin RPCMsgIn2) error {
	if m.debug {
		log.Printf("DEBUG: Making UpdateRequest call to inspector: RequestID=%s", inspin.RequestID)
	}

	// NOTE: Currently the output argument is not used
	err := m.inspector.UpdateRequest(&inspin, &RPCMsgOut{})
	if err != nil {
		if m.debug {
			log.Printf("DEBUG: UpdateRequest call error (RequestID=%s): %s", inspin.RequestID, err)
		}
	}

	return err
}

// NewRPCMsgIn creates a message from a go http.Request object
// End-users of the golang module never need to use this
// directly and it is only exposed for performance testing
func NewRPCMsgIn(r *http.Request, postbody []byte, code int, size int64, dur time.Duration, module, server string) *RPCMsgIn {
	now := time.Now().UTC()

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

	// skip reading - post is invalid or too long
	if req.ContentLength < 0 || req.ContentLength > m.maxContentLength {
		return false
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
