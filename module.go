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

var (
	debug           = false
	timeout         = 100 * time.Millisecond
	anomalySize     = 512 * 1024
	anomalyDuration = 1000
)

// StripPort removes any ending port from an IP address
// GoLang appears to add this when the client is from localhost (127.0.0.1:12312)
func StripPort(ipdots string) string {
	host, _, err := net.SplitHostPort(ipdots)
	if err != nil {
		return ipdots
	}
	return host
}

const maxContentLength = 300000

var ignoredMethods = map[string]bool{
	"OPTIONS": true,
	"CONNECT": true,
}

type responseRecorder struct {
	w    http.ResponseWriter
	code int
	size int
}

func (l *responseRecorder) Header() http.Header {
	return l.w.Header()
}

func (l *responseRecorder) WriteHeader(status int) {
	l.code = status
	l.w.WriteHeader(status)
}

func (l *responseRecorder) Write(b []byte) (int, error) {
	l.size += len(b)
	return l.w.Write(b)
}

func validContentLength(req *http.Request) bool {
	if req.ContentLength < 0 {
		return false
	}

	return req.ContentLength <= maxContentLength
}

func validContentType(req *http.Request) bool {
	contentType := req.Header.Get("Content-Type")
	if contentType == "application/x-www-form-urlencoded" {
		return true
	}

	if strings.Contains(contentType, "json") || strings.Contains(contentType, "javascript") {
		return true
	}

	return false
}

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

// NewRPCMsgIn creates a agent message from a go http.Request object
//  This is would part of a Go-lang module
func NewRPCMsgIn(r *http.Request, readpost bool, code int, size int, millis int) *RPCMsgIn {
	// get post body
	postbody := ""
	if readpost && validContentLength(r) && validContentType(r) {
		if r.Body != nil {
			buf, _ := ioutil.ReadAll(r.Body)
			reader := ioutil.NopCloser(bytes.NewBuffer(buf))
			postbody = string(buf)
			r.Body.Close()
			r.Body = reader
		}
	}

	// assemble an message to send to agent
	tlsProtocol := ""
	tlsCipher := ""
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
		tlsProtocol = tlstext.Version(r.TLS.Version)
		tlsCipher = tlstext.CipherSuite(r.TLS.CipherSuite)
	}
	return &RPCMsgIn{
		ModuleVersion:  "sigsci-module-golang 1.0", /* should come from somewhere else */
		ServerVersion:  runtime.Version(),
		ServerFlavor:   "", /* not sure what should be here */
		ServerName:     r.Host,
		Timestamp:      time.Now().UTC().Unix(), /* TODO: change to when request came in */
		NowMillis:      time.Now().UTC().UnixNano() / 1e6,
		RemoteAddr:     StripPort(r.RemoteAddr),
		Method:         r.Method,
		Scheme:         scheme,
		URI:            r.RequestURI,
		Protocol:       r.Proto,
		TLSProtocol:    tlsProtocol,
		TLSCipher:      tlsCipher,
		ResponseCode:   int32(code),
		ResponseMillis: int64(millis),
		ResponseSize:   int64(size),
		PostBody:       postbody,
		HeadersIn:      filterHeaders(r.Header),
	}
}

// AgentHandler is our agent handler
type AgentHandler struct {
	rpcAddress string
	handler    http.Handler
	pool       pool.Pool
}

func (a *AgentHandler) makeConnection() (net.Conn, error) {
	if debug {
		log.Printf("Making a new RPC connection.")
	}
	conn, err := net.DialTimeout("unix", a.rpcAddress, timeout)
	if err != nil {
		return nil, err
	}
	conn.SetDeadline(time.Now().Add(timeout))
	return conn, nil
}

func (a *AgentHandler) getConnection() (net.Conn, error) {
	if a.pool != nil {
		c, err := a.pool.Get()
		if err == nil {
			c.SetDeadline(time.Now().Add(timeout))
			return c, err
		}
		log.Printf("ERROR: failed to get a RPC connection from the pool - %s", err.Error())
		// Fall through to non-pool method.
	}

	return a.makeConnection()
}

// SetRPCConnPoolSize sets the min and max RPC connection
// pool size. Setting max to 0 disables pooling (the default).
func (a *AgentHandler) SetRPCConnPoolSize(min int, max int) {
	if min < 0 {
		min = 0
	}
	if max < 0 {
		max = 0
	}
	if min > max {
		min = max
	}
	if a.pool != nil {
		a.pool.Close()
	}
	if max == 0 {
		if a.pool != nil && debug {
			log.Printf("Disabling RPC connection pooling.")
		}
		a.pool = nil
	} else {
		var err error
		if debug {
			log.Printf("Creating a new RPC connection pool min=%d max=%d", min, max)
		}
		a.pool, err = pool.NewChannelPool(min, max, a.makeConnection)
		if err != nil {
			log.Printf("ERROR: failed to create a new RPC connection pool min=%d max=%d - %s", min, max, err.Error())
			a.pool = nil
		}
	}
}

// NewAgentHandler wraps an http.Handler and passes the requests to
// a Signal Sciences agent running on the given agent URI defaults
// socket to /tmp/sigsci-lua
func NewAgentHandler(h http.Handler) *AgentHandler {
	return &AgentHandler{"/tmp/sigsci-lua", h, nil}
}

// NewAgentHandlerWithSocket temporary additional constructor to let
// you set the path to the socket. Module needs a little love to
// make things easier to configure
func NewAgentHandlerWithSocket(sock string, h http.Handler) *AgentHandler {
	return &AgentHandler{sock, h, nil}
}

// AgentPreRequest makes a prerequest RPC call to the agent
func (a *AgentHandler) AgentPreRequest(req *http.Request) (agentin2 RPCMsgIn2, out RPCMsgOut, err error) {
	conn, err := a.getConnection()
	if err != nil {
		return agentin2, out, fmt.Errorf("unable to get connection : %s", err.Error())
	}

	rpcCodec := newMsgpClientCodec(conn)
	client := rpc.NewClientWithCodec(rpcCodec)
	defer client.Close()

	// Create message to agent from the input request
	agentin := NewRPCMsgIn(req, true, -1, -1, -1)
	err = client.Call("RPC.PreRequest", agentin, &out)
	if err != nil {
		return agentin2, out, fmt.Errorf("unable to make RPC.PreRequest call: %s", err.Error())
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
func (a *AgentHandler) AgentPostRequest(req *http.Request, agentResponse int32, code int, size int, millis int, hout http.Header) error {
	conn, err := a.getConnection()
	if err != nil {
		return err
	}

	rpcCodec := newMsgpClientCodec(conn)
	client := rpc.NewClientWithCodec(rpcCodec)
	defer client.Close()

	// Create message to agent from the input request
	agentin := NewRPCMsgIn(req, false, code, size, millis)
	agentin.WAFResponse = agentResponse
	agentin.HeadersOut = filterHeaders(hout)
	var out int
	err = client.Call("RPC.PostRequest", agentin, &out)
	return err
}

// AgentUpdateRequest makes an updaterequest RPC call to the agent
func (a *AgentHandler) AgentUpdateRequest(req *http.Request, agentin RPCMsgIn2) error {
	conn, err := a.getConnection()
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

func (a *AgentHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	start := time.Now().UTC()

	// Do not process ignored methods.
	if ignoredMethods[req.Method] {
		a.handler.ServeHTTP(w, req)
		return
	}

	agentin2, out, err := a.AgentPreRequest(req)
	if err != nil {
		// Fail open
		if debug {
			log.Println("Error pushing prerequest to agent: ", err.Error())
		}
		a.handler.ServeHTTP(w, req)
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
		a.handler.ServeHTTP(rr, req)
	default:
		log.Printf("ERROR: Received invalid response code from agent: %d", wafresponse)
		// continue with normal request
		a.handler.ServeHTTP(rr, req)
	}

	end := time.Now().UTC()
	code := rr.code
	size := rr.size
	duration := int(end.Sub(start) / time.Millisecond)

	if agentin2.RequestID != "" {
		agentin2.ResponseCode = int32(code)
		agentin2.ResponseSize = int64(size)
		agentin2.ResponseMillis = int64(duration)
		agentin2.HeadersOut = filterHeaders(rr.Header())
		if err := a.AgentUpdateRequest(req, agentin2); err != nil && debug {
			log.Printf("ERROR: 'RPC.UpdateRequest' call failed: %s", err.Error())
		}
		return
	}

	if code >= 300 || size >= anomalySize || duration >= anomalyDuration {
		if err := a.AgentPostRequest(req, int32(wafresponse), code, size, duration, rr.Header()); err != nil && debug {
			log.Printf("ERROR: 'RPC.PostRequest' request failed: %s", err.Error())
		}
	}
}
