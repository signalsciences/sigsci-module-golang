package sigsci

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestNewRPCMsgInWithModuleConfigFromRequest(t *testing.T) {

	c, err := NewModuleConfig(
		AllowUnknownContentLength(true),
		ServerFlavor("SugarAndSpice"),
		AltResponseCodes(403),
		AnomalyDuration(10*time.Second),
		AnomalySize(8192),
		CustomInspector(&RPCInspector{}, func(_ *http.Request) bool { return true }, func(_ *http.Request) {}),
		CustomHeaderExtractor(func(_ *http.Request) (http.Header, error) { return nil, nil }),
		Debug(true),
		MaxContentLength(500000),
		Socket("tcp", "0.0.0.0:1234"),
		Timeout(10*time.Millisecond),
	)
	if err != nil {
		t.Fatalf("Failed to create module config: %s", err)
	}

	b := bytes.Buffer{}
	b.WriteString("test")
	r, err := http.NewRequest("GET", "http://localhost/", &b)
	if err != nil {
		t.Fatal(err)
	}
	r.RemoteAddr = "127.0.0.1"
	r.Header.Add("If-None-Match", `W/"wyzzy"`)
	r.RequestURI = "http://localhost/"
	r.TLS = &tls.ConnectionState{}

	want := RPCMsgIn{
		ServerName:   "localhost",
		ServerFlavor: "SugarAndSpice",
		Method:       "GET",
		Scheme:       "https",
		URI:          "http://localhost/",
		Protocol:     "HTTP/1.1",
		RemoteAddr:   "127.0.0.1",
		HeadersIn:    [][2]string{{"Host", "localhost"}, {"If-None-Match", `W/"wyzzy"`}},
	}
	eq := func(got, want RPCMsgIn) (ne string, equal bool) {
		switch {
		case got.ServerName != want.ServerName:
			return "ServerHostname", false
		case got.Method != want.Method:
			return "Method", false
		case got.Scheme != want.Scheme:
			return "Scheme", false
		case got.URI != want.URI:
			return "URI", false
		case got.Protocol != want.Protocol:
			return "Protocol", false
		case got.RemoteAddr != want.RemoteAddr:
			return "RemoteAddr", false
		case !reflect.DeepEqual(got.HeadersIn, want.HeadersIn):
			return "HeadersIn", false
		case got.ServerFlavor != want.ServerFlavor:
			return "ServerFlavor", false
		default:
			return "", true
		}
	}

	got := NewRPCMsgInWithModuleConfig(c, r, nil)
	if ne, equal := eq(*got, want); !equal {
		t.Errorf("NewRPCMsgInWithModuleConfig: incorrect %q", ne)
	}
}

func TestNewRPCMsgFromRequest(t *testing.T) {
	b := bytes.Buffer{}
	b.WriteString("test")
	r, err := http.NewRequest("GET", "http://localhost/", &b)
	if err != nil {
		t.Fatal(err)
	}
	r.RemoteAddr = "127.0.0.1"
	r.Header.Add("If-None-Match", `W/"wyzzy"`)
	r.RequestURI = "http://localhost/"
	r.TLS = &tls.ConnectionState{}

	want := RPCMsgIn{
		ServerName: "localhost",
		Method:     "GET",
		Scheme:     "https",
		URI:        "http://localhost/",
		Protocol:   "HTTP/1.1",
		RemoteAddr: "127.0.0.1",
		HeadersIn:  [][2]string{{"Host", "localhost"}, {"If-None-Match", `W/"wyzzy"`}},
	}
	eq := func(got, want RPCMsgIn) (ne string, equal bool) {
		switch {
		case got.ServerName != want.ServerName:
			return "ServerHostname", false
		case got.Method != want.Method:
			return "Method", false
		case got.Scheme != want.Scheme:
			return "Scheme", false
		case got.URI != want.URI:
			return "URI", false
		case got.Protocol != want.Protocol:
			return "Protocol", false
		case got.RemoteAddr != want.RemoteAddr:
			return "RemoteAddr", false
		case !reflect.DeepEqual(got.HeadersIn, want.HeadersIn):
			return "HeadersIn", false
		default:
			return "", true
		}
	}

	got := NewRPCMsgIn(r, nil, -1, -1, -1, "", "")
	if ne, equal := eq(*got, want); !equal {
		t.Errorf("NewRPCMsgIn: incorrect %q", ne)
	}
}

// helper functions

func TestStripPort(t *testing.T) {
	cases := []struct {
		want    string
		content string
	}{
		// Invalid, should not change
		{"", ""},
		{"foo:bar:baz", "foo:bar:baz"},
		// Valid, should have port removed if exists
		{"127.0.0.1", "127.0.0.1"},
		{"127.0.0.1", "127.0.0.1:8000"},
	}
	for pos, tt := range cases {
		got := stripPort(tt.content)
		if got != tt.want {
			t.Errorf("test %d: StripPort(%q) = %q, want %q", pos, tt.content, got, tt.want)
		}
	}
}

func TestShouldReadBody(t *testing.T) {
	m, err := NewModule(
		http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			status := http.StatusOK
			http.Error(w, fmt.Sprintf("%d %s\n", status, http.StatusText(status)), status)
		}),
		MaxContentLength(20),
	)
	if err != nil {
		t.Fatalf("Failed to create module: %s", err)
	}

	cases := []struct {
		want   bool
		genreq func() []byte
	}{
		// No C-T
		{false, func() []byte {
			return genTestRequest("GET", "http://example.com/", "", "")
		}},
		// Invalid C-T
		{false, func() []byte {
			return genTestRequest("GET", "http://example.com/", "bad/type", `{}`)
		}},
		// Zero length
		{false, func() []byte {
			return genTestRequest("GET", "http://example.com/", "application/json", ``)
		}},
		// Too long
		{false, func() []byte {
			return genTestRequest("GET", "http://example.com/", "application/json", `{"foo":"12345678901234567890"}`)
		}},
		// Good to read
		{true, func() []byte {
			return genTestRequest("GET", "http://example.com/", "application/json", `{}`)
		}},
	}

	for pos, tt := range cases {
		req, err := requestParseRaw("127.0.0.1:59000", tt.genreq())
		if err != nil {
			t.Fatalf("Failed to generate request: %s", err)
		}
		got := shouldReadBody(req, m)
		if got != tt.want {
			t.Errorf("test %d: expected %v got %v", pos, tt.want, got)
		}
	}
}

func TestConvertHeaders(t *testing.T) {
	cases := []struct {
		want    [][2]string // Only the order of like keys matters
		content http.Header // Order of values matter
	}{
		// Empty
		{
			[][2]string{},
			http.Header{},
		},
		// Single values
		{
			[][2]string{
				{http.CanonicalHeaderKey("a"), "val a"},
				{http.CanonicalHeaderKey("b"), "val b"},
			}, http.Header{
				http.CanonicalHeaderKey("a"): {"val a"},
				http.CanonicalHeaderKey("b"): {"val b"},
			},
		},
		// Multiple values
		{
			[][2]string{
				{http.CanonicalHeaderKey("a"), "val a"},
				{http.CanonicalHeaderKey("b"), "val b1"},
				{http.CanonicalHeaderKey("b"), "val b2"},
			}, http.Header{
				http.CanonicalHeaderKey("a"): {"val a"},
				http.CanonicalHeaderKey("b"): {"val b1", "val b2"},
			},
		},
	}

	for pos, tt := range cases {
		got := convertHeaders(tt.content)

		// Convert result back to a http.Header for comparison
		hmap := http.Header{}
		for _, v := range got {
			hmap.Add(v[0], v[1])
		}
		if !reflect.DeepEqual(tt.content, hmap) {
			t.Errorf("test %d: expected %#v, got %#v", pos, tt.content, hmap)
		}
	}
}

func TestInspectableContentType(t *testing.T) {
	cases := []struct {
		want    bool
		content string
	}{
		{true, "application/x-www-form-urlencoded"},
		{true, "application/x-www-form-urlencoded; charset=UTF-8"},
		{true, "multipart/form-data"},
		{true, "text/xml"},
		{true, "application/xml"},
		{true, "text/xml;charset=UTF-8"},
		{true, "application/xml; charset=iso-2022-kr"},
		{true, "application/rss+xml"},
		{true, "application/json"},
		{true, "application/x-javascript"},
		{true, "text/javascript"},
		{true, "text/x-javascript"},
		{true, "text/x-json"},
		{true, "application/javascript"},
		{true, "application/graphql"},
		{false, "octet/stream"},
		{false, "junk/yard"},
	}

	for pos, tt := range cases {
		got := inspectableContentType(tt.content)
		if got != tt.want {
			t.Errorf("test %d: expected %v got %v", pos, tt.want, got)
		}
	}
}

func TestModule(t *testing.T) {
	cases := []struct {
		req  []byte // Raw HTTP request
		resp int32  // Inspection response
		tags string // Any tags in the PreRequest call
	}{
		{genTestRequest("GET", "http://example.com/", "", ""), 200, ""},
		{genTestRequest("GET", "http://example.com/", "", ""), 406, "XSS"},
		{genTestRequest("GET", "http://example.com/", "", ""), 403, "XSS"},
		{genTestRequest("GET", "http://example.com/", "", ""), 500, "XSS"},
		{genTestRequest("GET", "http://example.com/", "", ""), 200, ""},
		{genTestRequest("OPTIONS", "http://example.com/", "", ""), 200, ""},
		{genTestRequest("OPTIONS", "http://example.com/", "", ""), 406, "XSS"},
		{genTestRequest("OPTIONS", "http://example.com/", "", ""), 403, "XSS"},
		{genTestRequest("OPTIONS", "http://example.com/", "", ""), 500, "XSS"},
		{genTestRequest("OPTIONS", "http://example.com/", "", ""), 200, ""},
		{genTestRequest("CONNECT", "http://example.com/", "", ""), 200, ""},
		{genTestRequest("CONNECT", "http://example.com/", "", ""), 406, "XSS"},
		{genTestRequest("CONNECT", "http://example.com/", "", ""), 403, "XSS"},
		{genTestRequest("CONNECT", "http://example.com/", "", ""), 500, "XSS"},
		{genTestRequest("CONNECT", "http://example.com/", "", ""), 200, ""},
		{genTestRequest("POST", "http://example.com/", "application/x-www-form-urlencoded", "a=1"), 200, ""},
		{genTestRequest("POST", "http://example.com/", "application/x-www-form-urlencoded", "a=1"), 406, "XSS"},
		{genTestRequest("POST", "http://example.com/", "application/x-www-form-urlencoded", "a=1"), 403, "XSS"},
		{genTestRequest("POST", "http://example.com/", "application/x-www-form-urlencoded", "a=1"), 500, "XSS"},
		{genTestRequest("POST", "http://example.com/", "application/x-www-form-urlencoded", "a=1"), 200, ""},
		{genTestRequest("PUT", "http://example.com/", "application/x-www-form-urlencoded", "a=1"), 200, ""},
		{genTestRequest("PUT", "http://example.com/", "application/x-www-form-urlencoded", "a=1"), 406, "XSS"},
		{genTestRequest("PUT", "http://example.com/", "application/x-www-form-urlencoded", "a=1"), 403, "XSS"},
		{genTestRequest("PUT", "http://example.com/", "application/x-www-form-urlencoded", "a=1"), 500, "XSS"},
		{genTestRequest("PUT", "http://example.com/", "application/x-www-form-urlencoded", "a=1"), 200, ""},
		{genTestRequest("POST", "http://example.com/", "text/xml;charset=UTF-8", `<a>1</a>`), 200, ""},
		{genTestRequest("POST", "http://example.com/", "text/xml;charset=UTF-8", `<a>1</a>`), 406, "XSS"},
		{genTestRequest("POST", "http://example.com/", "text/xml;charset=UTF-8", `<a>1</a>`), 403, "XSS"},
		{genTestRequest("POST", "http://example.com/", "text/xml;charset=UTF-8", `<a>1</a>`), 500, "XSS"},
		{genTestRequest("POST", "http://example.com/", "text/xml;charset=UTF-8", `<a>1</a>`), 200, ""},
		{genTestRequest("POST", "http://example.com/", "application/xml; charset=iso-2022-kr", `<a>1</a>`), 200, ""},
		{genTestRequest("POST", "http://example.com/", "application/xml; charset=iso-2022-kr", `<a>1</a>`), 406, "XSS"},
		{genTestRequest("POST", "http://example.com/", "application/xml; charset=iso-2022-kr", `<a>1</a>`), 403, "XSS"},
		{genTestRequest("POST", "http://example.com/", "application/xml; charset=iso-2022-kr", `<a>1</a>`), 500, "XSS"},
		{genTestRequest("POST", "http://example.com/", "application/xml; charset=iso-2022-kr", `<a>1</a>`), 200, ""},
		{genTestRequest("POST", "http://example.com/", "application/rss+xml", `<a>1</a>`), 200, ""},
		{genTestRequest("POST", "http://example.com/", "application/rss+xml", `<a>1</a>`), 406, "XSS"},
		{genTestRequest("POST", "http://example.com/", "application/rss+xml", `<a>1</a>`), 403, "XSS"},
		{genTestRequest("POST", "http://example.com/", "application/rss+xml", `<a>1</a>`), 500, "XSS"},
		{genTestRequest("POST", "http://example.com/", "application/rss+xml", `<a>1</a>`), 200, ""},
		{genTestRequest("POST", "http://example.com/", "application/graphql", `{}`), 200, ""},
	}

	for pos, tt := range cases {
		respstr := strconv.Itoa(int(tt.resp))
		m, err := NewModule(
			http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				status := http.StatusOK
				http.Error(w, fmt.Sprintf("%d %s\n", status, http.StatusText(status)), status)
			}),
			Timeout(500*time.Millisecond),
			AltResponseCodes(403, 500),
			Debug(true),
			CustomInspector(newTestInspector(tt.resp, tt.tags), nil, nil),
		)
		if err != nil {
			t.Fatalf("test %d: Failed to create module: %s", pos, err)
		}

		req, err := requestParseRaw("127.0.0.1:12345", tt.req)
		if err != nil {
			t.Fatalf("test %d: Failed to parse request: %s\n%s", pos, err, tt.req)
		}

		if dump, err := httputil.DumpRequest(req, true); err == nil {
			t.Log("CLIENT REQUEST:\n" + string(dump))
		}

		if hv := req.Header.Get(`X-Sigsci-Agentresponse`); hv != "" {
			t.Errorf("test %d: unexpected request header %s=%s", pos, `X-Sigsci-Agentresponse`, hv)
		}

		w := httptest.NewRecorder()
		m.ServeHTTP(w, req)
		resp := w.Result()

		if dump, err := httputil.DumpRequest(req, true); err == nil {
			t.Log("SERVER REQUEST:\n" + string(dump))
		}
		if hv := req.Header.Get(`X-Sigsci-Agentresponse`); hv == "" || hv != respstr {
			t.Errorf("test %d: unexpected request header %s=%s, expected %q", pos, `X-Sigsci-Agentresponse`, hv, respstr)
		}
		if len(tt.tags) > 0 {
			if hv := req.Header.Get(`X-Sigsci-Requestid`); hv == "" {
				t.Errorf("test %d: expected request header %s=%s", pos, `X-Sigsci-Requestid`, hv)
			}
		}

		if dump, err := httputil.DumpResponse(resp, true); err == nil {
			t.Log("SERVER RESPONSE:\n" + string(dump))
		}
		if resp.StatusCode != int(tt.resp) {
			t.Errorf("test %d: unexpected status code=%d, expected=%d", pos, resp.StatusCode, tt.resp)
		}

	}
}

func genTestRequest(meth, uri, ctype, payload string) []byte {
	var err error
	var req *http.Request

	if len(payload) > 0 {
		req, err = http.NewRequest(meth, uri, strings.NewReader(payload))
		if err != nil {
			panic(err)
		}
	} else {
		req, err = http.NewRequest(meth, uri, nil)
		if err != nil {
			panic(err)
		}
	}

	req.Header.Set(`User-Agent`, `SigSci Module Tester/0.1`)
	if len(ctype) > 0 {
		req.Header.Set(`Content-Type`, ctype)
	}

	// This will add some extra headers typically added by the client
	dump, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		panic(err)
	}

	return dump
}

// requestParseRaw creates a request from the given raw HTTP data
func requestParseRaw(raddr string, raw []byte) (*http.Request, error) {
	req, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(raw)))
	if err != nil {
		return nil, err
	}

	// Set fields typically set by the server
	req.RemoteAddr = raddr

	return req, nil
}

// testInspector is a custom inspector that calls the simulator
// harness within the golang module
type testInspector struct {
	resp int32  // Response code (200, 406, etc.)
	tags string // EX: "XSS" (csv)
}

func newTestInspector(resp int32, tags string) *testInspector {
	return &testInspector{
		resp: resp,
		tags: tags,
	}
}

func (insp *testInspector) ModuleInit(in *RPCMsgIn, out *RPCMsgOut) error {
	out.WAFResponse = 200
	out.RequestID = ""
	out.RequestHeaders = nil
	return nil
}

func (insp *testInspector) PreRequest(in *RPCMsgIn, out *RPCMsgOut) error {
	out.WAFResponse = insp.resp
	if len(insp.tags) > 0 {
		out.RequestID = "0123456789abcdef01234567"
		out.RequestHeaders = [][2]string{
			{"X-SigSci-Tags", insp.tags},
		}
	} else {
		out.RequestID = ""
		out.RequestHeaders = nil
	}

	return nil
}

func (insp *testInspector) PostRequest(in *RPCMsgIn, out *RPCMsgOut) error {
	out.WAFResponse = insp.resp
	out.RequestID = ""
	out.RequestHeaders = nil

	return nil
}

func (insp *testInspector) UpdateRequest(in *RPCMsgIn2, out *RPCMsgOut) error {
	out.WAFResponse = insp.resp
	out.RequestID = ""
	out.RequestHeaders = nil

	return nil
}
