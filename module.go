package sigsci

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/signalsciences/tlstext"
)

const moduleVersion = "sigsci-module-golang " + version

// Module is an http.Handler that wraps inbound communication and
// sends it to the Signal Sciences Agent.
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

// NewModule wraps an http.Handler use the Signal Sciences Agent
// Configuration is based on 'functional options' as mentioned in:
// https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis
func NewModule(h http.Handler, options ...func(*Module) error) (*Module, error) {
	// the following are the defaults
	// you over-ride them by passing in function arguments
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

	// override defaults from function args
	for _, opt := range options {
		err := opt(&m)
		if err != nil {
			return nil, err
		}
	}

	if m.inspector == nil {
		m.inspector = &RPCInspector{
			Network: m.rpcNetwork,
			Address: m.rpcAddress,
			Timeout: m.timeout,
			Debug:   m.debug,
		}
	}

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
func Debug(enable bool) func(*Module) error {
	return func(m *Module) error {
		m.debug = enable
		return nil
	}
}

// Socket is a function argument to set where send data to the
// Signal Sciences Agent
func Socket(network, address string) func(*Module) error {
	return func(m *Module) error {
		m.rpcNetwork = network
		m.rpcAddress = address

		if network != "tcp" && network != "unix" {
			return fmt.Errorf("Network must be 'tcp' or 'unix', got %q", network)
		}

		// TODO: check that if TCP, address is an ip/port
		// TODO: check if UNIX (Domain Socket), then address is a path

		return nil
	}
}

// AnomalySize is a function argument to sent data to the inspector if the
// response was abnormally large.
func AnomalySize(size int64) func(*Module) error {
	return func(m *Module) error {
		m.anomalySize = size
		return nil
	}
}

// AnomalyDuration is a function argument to send data to the inspector if
// the response was abnormally slow
func AnomalyDuration(dur time.Duration) func(*Module) error {
	return func(m *Module) error {
		m.anomalyDuration = dur
		return nil
	}
}

// MaxContentLength is a functional argument to set the maximum post
// body length that will be processed
func MaxContentLength(size int64) func(*Module) error {
	return func(m *Module) error {
		m.maxContentLength = size
		return nil
	}
}

// Timeout is a function argument that sets the time to wait until
// receiving a reply from the inspector
func Timeout(t time.Duration) func(*Module) error {
	return func(m *Module) error {
		m.timeout = t
		return nil
	}
}

// ModuleIdentifier is a function argument that sets the module name
// and version for custom setups.
// The version should be a sem-version (e.g., "1.2.3")
func ModuleIdentifier(name, version string) func(*Module) error {
	return func(m *Module) error {
		m.moduleVersion = name + " " + version
		return nil
	}
}

// ServerIdentifier is a function argument that sets the serveru
// identifier for custom setups
func ServerIdentifier(id string) func(*Module) error {
	return func(m *Module) error {
		m.serverVersion = id
		return nil
	}
}

// CustomInspector is a function argument that sets a custom inspector,
// an optional inspector initializer to decide if inspection should occur, and
// an optional inspector finializer that can perform and post-inspection steps
func CustomInspector(insp Inspector, init InspectorInitFunc, fini InspectorFiniFunc) func(*Module) error {
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
		// Fail open
		m.handler.ServeHTTP(w, req)
		return
	}
	if m.inspFini != nil {
		defer m.inspFini(req)
	}

	inspin2, out, err := m.inspectorPreRequest(req)
	if err != nil {
		// Fail open
		if m.debug {
			log.Println("Error in prerequest to inspector: ", err.Error())
		}
		m.handler.ServeHTTP(w, req)
		return
	}

	// NOTE: according to net/http docs, if WriteHeader is not called explicitly,
	// the first call to Write will trigger an implicit WriteHeader(http.StatusOK).
	// this is why the default code is 200 and it only changes if WriteHeader is called.
	rr := &responseRecorder{w, 200, 0}

	wafresponse := out.WAFResponse
	switch wafresponse {
	case 406:
		status := int(wafresponse)
		http.Error(rr, fmt.Sprintf("%d %s\n", status, http.StatusText(status)), status)
	case 200:
		// continue with normal request
		m.handler.ServeHTTP(rr, req)
	default:
		log.Printf("ERROR: Received invalid response code from inspector: %d", wafresponse)
		// continue with normal request
		m.handler.ServeHTTP(rr, req)
	}

	end := time.Now().UTC()
	code := rr.code
	size := rr.size
	duration := end.Sub(start)

	if inspin2.RequestID != "" {
		inspin2.ResponseCode = int32(code)
		inspin2.ResponseSize = int64(size)
		inspin2.ResponseMillis = int64(duration / time.Millisecond)
		inspin2.HeadersOut = filterHeaders(rr.Header())
		if err := m.inspectorUpdateRequest(inspin2); err != nil && m.debug {
			log.Printf("ERROR: 'RPC.UpdateRequest' call failed: %s", err.Error())
		}
		return
	}

	if code >= 300 || size >= m.anomalySize || duration >= m.anomalyDuration {
		if err := m.inspectorPostRequest(req, wafresponse, code, size, duration, rr.Header()); err != nil && m.debug {
			log.Printf("ERROR: 'RPC.PostRequest' request failed:%s", err.Error())
		}
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
	if readPost(req, m) {
		// Read all of it and close
		// if error, just keep going
		// It's possible that is is error event
		// but not sure what it is.  Likely
		// the client disconnected.
		postbody, _ = ioutil.ReadAll(req.Body)
		req.Body.Close()

		// make a new reader so the next handler
		// can still read the post normally as if
		// nothing happened
		req.Body = ioutil.NopCloser(bytes.NewBuffer(postbody))
	}

	inspin := NewRPCMsgIn(req, postbody, -1, -1, -1, m.moduleVersion, m.serverVersion)

	err = m.inspector.PreRequest(inspin, &out)
	if err != nil {
		return
	}

	// set any request headers
	if out.RequestID != "" {
		req.Header.Add("X-SigSci-RequestID", out.RequestID)
	}

	wafresponse := out.WAFResponse
	req.Header.Add("X-SigSci-AgentResponse", strconv.Itoa(int(wafresponse)))
	for _, kv := range out.RequestHeaders {
		req.Header.Add(kv[0], kv[1])
	}

	inspin2 = RPCMsgIn2{
		RequestID:      out.RequestID,
		ResponseCode:   -1,
		ResponseMillis: -1,
		ResponseSize:   -1,
	}

	return
}

// inspectorPostRequest makes a postrequest call to the inspector
func (m *Module) inspectorPostRequest(req *http.Request, wafResponse int32,
	code int, size int64, millis time.Duration, hout http.Header) error {
	// Create message to agent from the input request
	inspin := NewRPCMsgIn(req, nil, code, size, millis, m.moduleVersion, m.serverVersion)
	inspin.WAFResponse = wafResponse
	inspin.HeadersOut = filterHeaders(hout)

	// TBD: Actually use the output
	return m.inspector.PostRequest(inspin, &RPCMsgOut{})
}

// inspectorUpdateRequest makes an updaterequest call to the inspector
func (m *Module) inspectorUpdateRequest(inspin RPCMsgIn2) error {
	// TBD: Actually use the output
	return m.inspector.UpdateRequest(&inspin, &RPCMsgOut{})
}

// NewRPCMsgIn creates a message from a go http.Request object
// End-users of the golang module never need to use this
// directly and it is only exposed for performance testing
func NewRPCMsgIn(r *http.Request, postbody []byte, code int, size int64, dur time.Duration, module, server string) *RPCMsgIn {
	// TODO: change to when request came in
	now := time.Now().UTC()

	// assemble an message to send to inspector
	tlsProtocol := ""
	tlsCipher := ""
	scheme := "http"
	if r.TLS != nil {
		// convert golang/spec integers into something human readable
		scheme = "https"
		tlsProtocol = tlstext.Version(r.TLS.Version)
		tlsCipher = tlstext.CipherSuite(r.TLS.CipherSuite)
	}
	return &RPCMsgIn{
		ModuleVersion:  module,
		ServerVersion:  server,
		ServerFlavor:   "", /* not sure what should be here */
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
		HeadersIn:      filterHeaders(r.Header),
	}
}

// stripPort removes any ending port from an IP address
// Go appears to add this when the client is from localhost ( e.g. "127.0.0.1:12312" )
func stripPort(ipdots string) string {
	host, _, err := net.SplitHostPort(ipdots)
	if err != nil {
		return ipdots
	}
	return host
}

type responseRecorder struct {
	w    http.ResponseWriter
	code int
	size int64
}

func (l *responseRecorder) Header() http.Header {
	return l.w.Header()
}

func (l *responseRecorder) WriteHeader(status int) {
	l.code = status
	l.w.WriteHeader(status)
}

func (l *responseRecorder) Write(b []byte) (int, error) {
	l.size += int64(len(b))
	return l.w.Write(b)
}

func (l *responseRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if h, ok := l.w.(http.Hijacker); ok {
		return h.Hijack()
	}
	// Required for WebSockets to work
	return nil, nil, fmt.Errorf("response writer (%T) does not implement http.Hijacker", l.w)
}

// readPost returns True if we should read a postbody or not
func readPost(req *http.Request, m *Module) bool {
	// nothing to do
	if req.Body == nil {
		return false
	}

	// skip reading post if too long
	if req.ContentLength < 0 {
		return false
	}
	if req.ContentLength > m.maxContentLength {
		return false
	}

	// only read certain types of content
	return checkContentType(req.Header.Get("Content-Type"))
}

func checkContentType(s string) bool {
	s = strings.ToLower(s)
	if strings.HasPrefix(s, "application/x-www-form-urlencoded") {
		return true
	}

	if strings.HasPrefix(s, "multipart/form-data") {
		return true
	}

	if strings.Contains(s, "json") {
		return true
	}

	if strings.Contains(s, "javascript") {
		return true
	}

	if strings.HasSuffix(s, "/xml") {
		return true
	}

	if strings.HasSuffix(s, "+xml") {
		return true
	}

	return false
}

// converts a http.Header map to a [][2]string
func filterHeaders(h http.Header) [][2]string {
	// get headers
	out := make([][2]string, 0, len(h)+1)

	// interestingly golang appears to remove the Host header
	// headersin = append(headersin, [2]string{"Host", r.Host})

	for key, values := range h {
		for _, value := range values {
			out = append(out, [2]string{key, value})
		}
	}
	return out
}
