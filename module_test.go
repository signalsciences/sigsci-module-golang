package sigsci

import (
	"bytes"
	"crypto/tls"
	"net/http"
	"reflect"
	"testing"
)

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
		HeadersIn:  [][2]string{{"If-None-Match", `W/"wyzzy"`}},
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
		t.Errorf("newWafMsgFromRequest: incorrect %q", ne)
	}
}

// helper functions
func TestStripPort(t *testing.T) {
	got := stripPort("127.0.0.1:8000")
	if got != "127.0.0.1" {
		t.Errorf("StripPort(%q) = %q, want %q", "127.0.0.1:8000", got, "127.0.0.1")
	}
}

func TestCheckContentType(t *testing.T) {
	cases := []struct {
		want    bool
		content string
	}{
		{true, "application/x-www-form-urlencoded"},
		{true, "application/x-www-form-urlencoded; charset=UTF-8"},
		{true, "multipart/form-data"},
		{true, "text/xml"},
		{true, "application/xml"},
		{true, "application/json"},
		{true, "application/x-javascript"},
		{true, "text/javascript"},
		{true, "text/x-javascript"},
		{true, "text/x-json"},
		{true, "application/javascript"},
		{false, "octet/stream"},
		{false, "junk/yard"},
	}

	for pos, tt := range cases {
		got := checkContentType(tt.content)
		if got != tt.want {
			t.Errorf("[%d] case %q expected %v got %v", pos, tt.want, got)
		}
	}
}
