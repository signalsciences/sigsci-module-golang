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
	"sync"
	"time"

	"github.com/client9/tlstext"
	pool "gopkg.in/fatih/pool.v2"
)

// Module is an http.Handler that wraps inbound communication and
// sends it to the Signal Sciences Agent.
type Module struct {
	handler          http.Handler
	rpcNetwork       string
	rpcAddress       string
	pool             pool.Pool
	debug            bool
	timeout          time.Duration
	anomalySize      int64
	anomalyDuration  time.Duration
	maxContentLength int64
	ignoredMethods   map[string]bool

	minBodyThreshold int64
	maxBodyThreshold int64
	asyncTimeout     time.Duration
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
		pool:             nil,
		timeout:          100 * time.Millisecond,
		anomalySize:      512 * 1024,
		anomalyDuration:  2 * time.Second,
		maxContentLength: 300000,
		ignoredMethods: map[string]bool{
			"OPTIONS": true,
			"CONNECT": true,
		},
		minBodyThreshold: 1024,
		maxBodyThreshold: 1024 * 1024,
		asyncTimeout:     1000 * time.Millisecond,
	}

	// override defaults from function args
	for _, opt := range options {
		err := opt(&m)
		if err != nil {
			return nil, err
		}
	}

	return &m, nil
}

// Version returns a SemVer version string
func Version() string {
	return version
}

// Debug turns on debug logging
func Debug() func(*Module) error {
	return func(m *Module) error {
		m.debug = true
		return nil
	}
}

// Socket is a function argument to set where send data to the agent
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

// ServeHTTP satisfies the http.Handler interface
func (m *Module) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	start := time.Now().UTC()

	// Do not process ignored methods.
	if m.ignoredMethods[req.Method] {
		m.handler.ServeHTTP(w, req)
		return
	}

	asyncAgentIn, agentin2, out, err := m.agentPreRequest(req)
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
		if asyncAgentIn != nil {
			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				out, err = m.callRPCPre(asyncAgentIn, m.asyncTimeout)
				if out.RequestID != "" {
					agentin2.RequestID = out.RequestID
				}
				wg.Done()
				if err != nil {
					log.Println("Error pushing async prerequest to agent: ", err.Error())
				}
				// NOTE: The extra meta data (e.g. post body copy) is kept in memory release it
				asyncAgentIn = nil
			}()
			// continue with normal request
			m.handler.ServeHTTP(rr, req)
			wg.Wait()
			// NOTE This does not try to block - for most app servers it's not possible to change
			// http status code, body, etc.  Once control is handed off to the app, the module is
			// passive and can only log data for further decisions.
			// If we really wanted to explore blocking (would be messy and a lot of details for each impl), we'd do something like:

			//			fakeReq := httptest.NewRecorder()
			//			m.handler.ServeHTTP(rr, fakeReq)
			//			if out.WAFResponse.Int() == 406 {
			//				block
			//			} else {
			//				copy(fakeReq, req)
			//			}

		} else {
			// continue with normal request
			m.handler.ServeHTTP(rr, req)
		}
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
		if err := m.agentUpdateRequest(req, agentin2); err != nil && m.debug {
			log.Printf("ERROR: 'RPC.UpdateRequest' call failed: %s", err.Error())
		}
		return
	}

	if code >= 300 || size >= m.anomalySize || duration >= m.anomalyDuration {
		if err := m.agentPostRequest(req, int32(wafresponse), code, size, duration, rr.Header()); err != nil && m.debug {
			log.Printf("ERROR: 'RPC.PostRequest' request failed:%s", err.Error())
		}
	}
}

func (m *Module) makeConnection() (net.Conn, error) {
	if m.debug {
		log.Printf("Making a new RPC connection.")
	}
	conn, err := net.DialTimeout(m.rpcNetwork, m.rpcAddress, m.timeout)
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

// SendRawPreRequest sends a preformatted RPCMsgIn to the agent
// End-users never need to use this function is expose for
// performance testing
func (m *Module) SendRawPreRequest(msg *RPCMsgIn) (out RPCMsgOut, err error) {
	conn, err := m.getConnection()
	if err != nil {
		return out, err
	}
	rpcCodec := newMsgpClientCodec(conn)
	client := rpc.NewClientWithCodec(rpcCodec)
	err = client.Call("RPC.PreRequest", msg, &out)
	client.Close()
	return out, err
}

// agentPreRequest makes a prerequest RPC call to the agent
// In general this is never to be used by end-users and is
// only exposed for use in performance testing
func (m *Module) agentPreRequest(req *http.Request) (asyncAgentIn *RPCMsgIn, agentin2 RPCMsgIn2, out RPCMsgOut, err error) {
	// if we can't get a connection, then we should not
	// do any work.  Maybe agent is down
	// TODO: does getConnection actually open a connection?
	conn, err := m.getConnection()
	if err != nil {
		return nil, agentin2, out, fmt.Errorf("unable to get connection : %s", err)
	}

	// Create message to agent from the input request
	// see if we can read-in the post body

	postbody := ""
	if readPost(req, m) {
		// Read all of it and close
		// if error, just keep going
		// It's possible that is is error event
		// but not sure what it is.  Likely
		// the client disconnected.
		buf, _ := ioutil.ReadAll(req.Body)
		req.Body.Close()

		// save a copy
		postbody = string(buf)

		// make a new reader so the next handler
		// can still read the post normally as if
		// nothing happened
		req.Body = ioutil.NopCloser(bytes.NewBuffer(buf))
	}

	// If the post body is larger than our max - throw it away
	if int64(len(postbody)) > m.maxBodyThreshold {
		postbody = ""
	}
	asyncPostBodyCopy := ""
	// If the post body is larger than our max - throw it away
	if int64(len(postbody)) > m.minBodyThreshold {
		asyncPostBodyCopy = postbody
		postbody = ""
	}

	// TODO: we make a full copy of the postbody, but it would
	//  appear we don't really need to do this as it's going to be
	//  encoded before.  Can we change NewRPCMsgIn to accept a []byte?
	//
	agentin := NewRPCMsgIn(req, postbody, -1, -1, -1)

	rpcCodec := newMsgpClientCodec(conn)
	client := rpc.NewClientWithCodec(rpcCodec)

	err = client.Call("RPC.PreRequest", agentin, &out)
	client.Close()
	if err != nil {
		return nil, agentin2, out, fmt.Errorf("unable to make RPC.PreRequest call: %s", err)
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

	if asyncPostBodyCopy != "" {
		asyncAgentIn = NewRPCMsgIn(req, asyncPostBodyCopy, -1, -1, -1)
	}
	return asyncAgentIn, agentin2, out, nil
}

// callRPCPre returns the agent RPCMsgOut
// TODO async vs normal might have different timeouts
func (m *Module) callRPCPre(agentin *RPCMsgIn, timeout time.Duration) (out RPCMsgOut, err error) {
	// if we can't get a connection, then we should not
	// do any work.  Maybe agent is down
	// TODO: does getConnection actually open a connection?
	conn, err := m.getConnection()
	if err != nil {
		return out, fmt.Errorf("unable to get connection : %s", err)
	}
	conn.SetDeadline(time.Now().Add(timeout))

	rpcCodec := newMsgpClientCodec(conn)
	client := rpc.NewClientWithCodec(rpcCodec)

	err = client.Call("RPC.PreRequest", agentin, &out)
	client.Close()
	if err != nil {
		return out, fmt.Errorf("unable to make RPC.PreRequest call: %s", err)
	}
	return out, nil
}

// agentPostRequest makes a postrequest RPC call to the agent
func (m *Module) agentPostRequest(req *http.Request, agentResponse int32,
	code int, size int64, millis time.Duration, hout http.Header) error {
	conn, err := m.getConnection()
	if err != nil {
		return err
	}

	// Create message to agent from the input request
	agentin := NewRPCMsgIn(req, "", code, size, millis)
	agentin.WAFResponse = agentResponse
	agentin.HeadersOut = filterHeaders(hout)

	rpcCodec := newMsgpClientCodec(conn)
	client := rpc.NewClientWithCodec(rpcCodec)
	var out int
	err = client.Call("RPC.PostRequest", agentin, &out)
	client.Close()
	return err
}

// agentUpdateRequest makes an updaterequest RPC call to the agent
func (m *Module) agentUpdateRequest(req *http.Request, agentin RPCMsgIn2) error {
	conn, err := m.getConnection()
	if err != nil {
		return err
	}

	rpcCodec := newMsgpClientCodec(conn)
	client := rpc.NewClientWithCodec(rpcCodec)

	var out int
	err = client.Call("RPC.UpdateRequest", &agentin, &out)
	client.Close()
	return err
}

const moduleVersion = "sigsci-module-golang " + version

// NewRPCMsgIn creates a agent message from a go http.Request object
// End-users of the golang module never need to use this
// directly and it is only exposed for performance testing
func NewRPCMsgIn(r *http.Request, postbody string, code int, size int64, dur time.Duration) *RPCMsgIn {
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
		ModuleVersion:  moduleVersion,
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
	return checkContentType(req.Header.Get("Content-Type"))
}

func checkContentType(s string) bool {
	s = strings.ToLower(s)
	if strings.HasPrefix(s, "application/x-www-form-urlencoded") {
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
