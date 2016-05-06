package sigsci

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/client9/tlstext"
	pool "gopkg.in/fatih/pool.v2"
)

// Version is the semantic version of this module
const Version = "sigsci-module-golang 1.0.0"

// Module is our agent handler
type Module struct {
	handler          http.Handler
	rpcAddress       string
	pool             pool.Pool
	debug            bool
	timeout          time.Duration
	anomalySize      int64
	anomalyDuration  time.Duration
	maxContentLength int64
	ignoredMethods   map[string]bool
}

// NewModule wraps an http.Handler use the Signal Sciences Agent
func NewModule(h http.Handler, options ...func(*Module) error) (*Module, error) {
	// the following are the defaults
	// you over-ride them by passing in function arguments
	m := Module{
		rpcAddress:       "unix:/var/run/sigsci.sock",
		debug:            false,
		pool:             nil,
		timeout:          100 * time.Millisecond,
		anomalySize:      512 * 1024,
		anomalyDuration:  2 * time.Second,
		maxContentLength: 300000,
		ignoredMethods: map[string]bool{
			"OPTIONS": true,
			"CONNECT": true,
		},
	}
	for _, opt := range options {
		err := opt(&m)
		if err != nil {
			return nil, err
		}
	}
	return &m, nil
}

// Socket is a function argument to set where send data to the agent
func Socket(address string) func(*Module) error {
	return func(m *Module) error {
		m.rpcAddress = address
		return nil
	}
}

// AnomalySize is a function argument to sent data to the agent if the
// response was abnormally large.
func AnomalySize(size int64) func(*Module) error {
	return func(m *Module) error {
		m.anomalySize = size
		return nil
	}
}

// AnomalyDuration is a function argument to send data to the agent if
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
// receiving a reply from the agent
func Timeout(t time.Duration) func(*Module) error {
	return func(m *Module) error {
		m.timeout = t
		return nil
	}
}

// ConnectionPoolSize sets the min and max RPC connection
// pool size. Setting max to 0 disables pooling (the default).
func ConnectionPoolSize(min, max int) func(*Module) error {
	return func(m *Module) error {
		if min < 0 || max < 0 || min > max {
			return fmt.Errorf("invalid values for min=%d, max=%d", min, max)
		}
		if m.pool != nil {
			m.pool.Close()
		}
		if max == 0 {
			m.pool = nil
			return nil
		}
		pool, err := pool.NewChannelPool(min, max, m.makeConnection)
		if err != nil {
			m.pool = nil
			return fmt.Errorf("failed to create connection pool min=%d max=%d - %s",
				min, max, err)
		}
		m.pool = pool
		return nil
	}
}

func (m *Module) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	start := time.Now().UTC()

	// Do not process ignored methods.
	if m.ignoredMethods[req.Method] {
		m.handler.ServeHTTP(w, req)
		return
	}

	agentin2, out, err := m.AgentPreRequest(req)
	if err != nil {
		// Fail open
		if m.debug {
			log.Println("Error pushing prerequest to agent: ", err.Error())
		}
		m.handler.ServeHTTP(w, req)
		return
	}

	// NOTE: according to net/http docs, if WriteHeader is not called explicitly,
	// the first call to Write will trigger an implicit WriteHeader(http.StatusOK).
	// this is why the default code is 200 and it only changes if WriteHeader is called.
	rr := &responseRecorder{w, 200, 0}

	wafresponse, _ := out.WAFResponse.Int()
	switch wafresponse {
	case 406:
		http.Error(rr, "you lose", int(wafresponse))
	case 200:
		// continue with normal request
		m.handler.ServeHTTP(rr, req)
	default:
		log.Printf("ERROR: Received invalid response code from agent: %d", wafresponse)
		// continue with normal request
		m.handler.ServeHTTP(rr, req)
	}

	end := time.Now().UTC()
	code := rr.code
	size := rr.size
	duration := end.Sub(start)

	if agentin2.RequestID != "" {
		agentin2.ResponseCode = int32(code)
		agentin2.ResponseSize = int64(size)
		agentin2.ResponseMillis = int64(duration / time.Millisecond)
		agentin2.HeadersOut = filterHeaders(rr.Header())
		if err := m.AgentUpdateRequest(req, agentin2); err != nil && m.debug {
			log.Printf("ERROR: 'RPC.UpdateRequest' call failed: %s", err.Error())
		}
		return
	}

	if code >= 300 || size >= m.anomalySize || duration >= m.anomalyDuration {
		if err := m.AgentPostRequest(req, int32(wafresponse), code, size, duration, rr.Header()); err != nil && m.debug {
			log.Printf("ERROR: 'RPC.PostRequest' request failed:%s", err.Error())
		}
	}
}

func (m *Module) makeConnection() (net.Conn, error) {
	if m.debug {
		log.Printf("Making a new RPC connection.")
	}
	conn, err := net.DialTimeout("unix", m.rpcAddress, m.timeout)
	if err != nil {
		return nil, err
	}
	conn.SetDeadline(time.Now().Add(m.timeout))
	return conn, nil
}

func (m *Module) getConnection() (net.Conn, error) {
	if m.pool != nil {
		c, err := m.pool.Get()
		if err == nil {
			c.SetDeadline(time.Now().Add(m.timeout))
			return c, err
		}
		log.Printf("ERROR: failed to get a RPC connection from the pool - %s", err)
		// Fall through to non-pool method.
	}

	return m.makeConnection()
}

// AgentPreRequest makes a prerequest RPC call to the agent
func (m *Module) AgentPreRequest(req *http.Request) (agentin2 RPCMsgIn2, out RPCMsgOut, err error) {
	conn, err := m.getConnection()
	if err != nil {
		return agentin2, out, fmt.Errorf("unable to get connection : %s", err)
	}

	rpcCodec := newMsgpClientCodec(conn)
	client := rpc.NewClientWithCodec(rpcCodec)
	defer client.Close()

	// Create message to agent from the input request
	// see if we can read-in the post body

	postbody := ""
	if readPost(req, m) {
		buf, _ := ioutil.ReadAll(req.Body)
		reader := ioutil.NopCloser(bytes.NewBuffer(buf))
		postbody = string(buf)
		req.Body.Close()
		req.Body = reader
	}
	agentin := newRPCMsgIn(req, postbody, -1, -1, -1)
	err = client.Call("RPC.PreRequest", agentin, &out)
	if err != nil {
		return agentin2, out, fmt.Errorf("unable to make RPC.PreRequest call: %s", err)
	}

	// set any request headers
	if out.RequestID != "" {
		req.Header.Add("X-SigSci-RequestID", out.RequestID)
	}

	wafresponse, _ := out.WAFResponse.Int()
	req.Header.Add("X-SigSci-AgentResponse", strconv.Itoa(int(wafresponse)))
	for _, kv := range out.RequestHeaders {
		req.Header.Add(kv[0], kv[1])
	}

	agentin2 = RPCMsgIn2{
		RequestID:      out.RequestID,
		ResponseCode:   -1,
		ResponseMillis: -1,
		ResponseSize:   -1,
	}

	return agentin2, out, nil
}

// AgentPostRequest makes a postrequest RPC call to the agent
func (m *Module) AgentPostRequest(req *http.Request, agentResponse int32,
	code int, size int64, millis time.Duration, hout http.Header) error {
	conn, err := m.getConnection()
	if err != nil {
		return err
	}

	rpcCodec := newMsgpClientCodec(conn)
	client := rpc.NewClientWithCodec(rpcCodec)
	defer client.Close()

	// Create message to agent from the input request
	agentin := newRPCMsgIn(req, "", code, size, millis)
	agentin.WAFResponse = agentResponse
	agentin.HeadersOut = filterHeaders(hout)
	var out int
	err = client.Call("RPC.PostRequest", agentin, &out)
	return err
}

// AgentUpdateRequest makes an updaterequest RPC call to the agent
func (m *Module) AgentUpdateRequest(req *http.Request, agentin RPCMsgIn2) error {
	conn, err := m.getConnection()
	if err != nil {
		return err
	}

	rpcCodec := newMsgpClientCodec(conn)
	client := rpc.NewClientWithCodec(rpcCodec)
	defer client.Close()

	var out int
	err = client.Call("RPC.UpdateRequest", &agentin, &out)
	return err
}

// NewRPCMsgIn creates a agent message from a go http.Request object
//  This is would part of a Go-lang module
func newRPCMsgIn(r *http.Request, postbody string, code int, size int64, dur time.Duration) *RPCMsgIn {
	// assemble an message to send to agent
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
		ModuleVersion:  Version,
		ServerVersion:  runtime.Version(),
		ServerFlavor:   "", /* not sure what should be here */
		ServerName:     r.Host,
		Timestamp:      time.Now().UTC().Unix(), /* TODO: change to when request came in */
		NowMillis:      time.Now().UTC().UnixNano() / 1e6,
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
		PostBody:       postbody,
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
	contentType := req.Header.Get("Content-Type")
	if contentType == "application/x-www-form-urlencoded" {
		return true
	}

	if strings.Contains(contentType, "json") {
		return true
	}

	if strings.Contains(contentType, "javascript") {
		return true
	}

	return false
}

//
// WARNING -- filterheaders will be removed, as functionality will move to agent
//
func filterHeaders(h http.Header) [][2]string {
	// get headers
	out := make([][2]string, 0, len(h)+1)

	// interestingly golang appears to remove the Host header
	// headersin = append(headersin, [2]string{"Host", r.Host})

	for key, values := range h {
		lowerkey := strings.ToLower(key)
		if lowerkey != "cookie" && lowerkey != "set-cookie" && lowerkey != "authorization" && lowerkey != "x-auth-token" {
			for _, value := range values {
				out = append(out, [2]string{key, value})
			}
		}
	}
	return out
}
