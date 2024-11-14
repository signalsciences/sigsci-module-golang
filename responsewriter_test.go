package sigsci

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/signalsciences/sigsci-module-golang/schema"
)

// testResponseRecorder is a httptest.ResponseRecorder without the Flusher interface
type testResponseRecorder struct {
	Recorder *httptest.ResponseRecorder
}

func (w *testResponseRecorder) Header() http.Header {
	return w.Recorder.Header()
}

func (w *testResponseRecorder) WriteHeader(status int) {
	w.Recorder.WriteHeader(status)
}

func (w *testResponseRecorder) Write(b []byte) (int, error) {
	return w.Recorder.Write(b)
}

func (w *testResponseRecorder) ReadFrom(r io.Reader) (n int64, err error) {
	return io.Copy(w.Recorder, r)
}

// testResponseRecorderFlusher is a httptest.ResponseRecorder with the Flusher interface
type testResponseRecorderFlusher struct {
	Recorder *httptest.ResponseRecorder
}

func (w *testResponseRecorderFlusher) Header() http.Header {
	return w.Recorder.Header()
}

func (w *testResponseRecorderFlusher) WriteHeader(status int) {
	w.Recorder.WriteHeader(status)
}

func (w *testResponseRecorderFlusher) Write(b []byte) (int, error) {
	return w.Recorder.Write(b)
}

func (w *testResponseRecorderFlusher) ReadFrom(r io.Reader) (n int64, err error) {
	return io.Copy(w.Recorder, r)
}

func (w *testResponseRecorderFlusher) Flush() {
	w.Recorder.Flush()
}

func testResponseWriter(t *testing.T, w ResponseWriter, flusher bool) {
	status := 200
	respbody := []byte("123456")

	req, err := http.NewRequest(http.MethodGet, "http://example.com/", nil)
	if err != nil {
		t.Fatalf("Failed to generate request: %s", err)
	}

	// Grab the recorder from the base response writer
	var recorder *httptest.ResponseRecorder
	switch rec := w.BaseResponseWriter().(type) {
	case *testResponseRecorder:
		recorder = rec.Recorder
	case *testResponseRecorderFlusher:
		recorder = rec.Recorder
	default:
		panic(fmt.Sprintf("unhandled recorder type: %T", w))
	}

	// This handler writes header/body and then flushes if the writer implements a http.Flusher
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(status)
		w.Write(respbody)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	})
	handler.ServeHTTP(w, req)

	// Verify the response
	resp := recorder.Result()
	defer resp.Body.Close()

	if resp.StatusCode != status {
		t.Errorf("Unexpected status code=%d, expected=%d", resp.StatusCode, status)
	}
	if w.StatusCode() != status {
		t.Errorf("Unexpected recorder status code=%d, expected=%d", w.StatusCode(), status)
	}

	// Verify body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to generate request: %s", err)
	}
	if string(body) != string(respbody) {
		t.Errorf("Unexpected response body=%q, expected=%q", body, respbody)
	}
	if w.BytesWritten() != int64(len(respbody)) {
		t.Errorf("Unexpected response size=%d, expected=%d", w.BytesWritten(), len(respbody))
	}

	// Verify expected flushed value
	if recorder.Flushed != flusher {
		t.Errorf("Unexpected flush=%v, expected %v w=%T recorder=%T", recorder.Flushed, flusher, w, recorder)
	}
}

// TestResponseWriter tests a non-flusher ResponseWriter
func TestResponseWriter(t *testing.T) {
	testResponseWriter(t, NewResponseWriter(&testResponseRecorder{httptest.NewRecorder()}), false)
}

// TestResponseWriterFlusher tests a flusher ResponseWriter
func TestResponseWriterFlusher(t *testing.T) {
	testResponseWriter(t, NewResponseWriter(&testResponseRecorderFlusher{httptest.NewRecorder()}), true)
}

func TestResponseHeaders(t *testing.T) {

	resp := &httptest.ResponseRecorder{
		HeaderMap: http.Header{
			"X-Powered-By": []string{"aa"},
			"Content-Type": []string{"text/plain"},
			"X-Report":     []string{"bb"},
		},
	}
	actions := []schema.Action{
		{schema.AddHdr, []string{"csp", "src=abc"}},
		{schema.SetHdr, []string{"content-type", "text/json"}},
		{schema.DelHdr, []string{"x-powered-by"}},
		{schema.SetNEHdr, []string{"x-report", "cc"}},
	}
	newResponseWriter(resp, actions).Write([]byte("foo"))

	got := resp.Header()
	expected := http.Header{
		"Csp":          []string{"src=abc"},
		"Content-Type": []string{"text/json"},
		"X-Report":     []string{"bb"},
	}
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("expected %v, got %v", expected, got)
	}
}
