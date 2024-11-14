package sigsci

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/signalsciences/sigsci-module-golang/schema"
)

// ResponseWriter is a http.ResponseWriter allowing extraction of data needed for inspection
type ResponseWriter interface {
	http.ResponseWriter
	BaseResponseWriter() http.ResponseWriter
	StatusCode() int
	BytesWritten() int64
}

// ResponseWriterFlusher is a ResponseWriter with a http.Flusher interface
type ResponseWriterFlusher interface {
	ResponseWriter
	http.Flusher
}

// NewResponseWriter returns a ResponseWriter or ResponseWriterFlusher depending on the base http.ResponseWriter.
func NewResponseWriter(base http.ResponseWriter) ResponseWriter {
	return newResponseWriter(base, nil)
}

func newResponseWriter(base http.ResponseWriter, actions []schema.Action) ResponseWriter {
	// NOTE: according to net/http docs, if WriteHeader is not called explicitly,
	// the first call to Write will trigger an implicit WriteHeader(http.StatusOK).
	// this is why the default code is 200 and it only changes if WriteHeader is called.
	w := &responseRecorder{
		base:    base,
		code:    200,
		actions: actions,
	}
	if _, ok := w.base.(http.Flusher); ok {
		return &responseRecorderFlusher{w}
	}
	return w
}

// responseRecorder wraps a base http.ResponseWriter allowing extraction of additional inspection data
type responseRecorder struct {
	base    http.ResponseWriter
	code    int
	size    int64
	actions []schema.Action
}

// BaseResponseWriter returns the base http.ResponseWriter allowing access if needed
func (w *responseRecorder) BaseResponseWriter() http.ResponseWriter {
	return w.base
}

// StatusCode returns the status code that was used
func (w *responseRecorder) StatusCode() int {
	return w.code
}

// BytesWritten returns the number of bytes written
func (w *responseRecorder) BytesWritten() int64 {
	return w.size
}

// Header returns the header object
func (w *responseRecorder) Header() http.Header {
	return w.base.Header()
}

// WriteHeader writes the header, recording the status code for inspection
func (w *responseRecorder) WriteHeader(status int) {
	if w.actions != nil {
		w.mergeHeader()
	}
	w.code = status
	w.base.WriteHeader(status)
}

func (w *responseRecorder) mergeHeader() {
	hdr := w.base.Header()
	for _, a := range w.actions {
		switch a.Code {
		case schema.AddHdr:
			hdr.Add(a.Args[0], a.Args[1])
		case schema.SetHdr:
			hdr.Set(a.Args[0], a.Args[1])
		case schema.SetNEHdr:
			if len(hdr.Get(a.Args[0])) == 0 {
				hdr.Set(a.Args[0], a.Args[1])
			}
		case schema.DelHdr:
			hdr.Del(a.Args[0])
		}
	}
	w.actions = nil
}

// Write writes data, tracking the length written for inspection
func (w *responseRecorder) Write(b []byte) (int, error) {
	if w.actions != nil {
		w.mergeHeader()
	}
	w.size += int64(len(b))
	return w.base.Write(b)
}

func (w *responseRecorder) ReadFrom(r io.Reader) (n int64, err error) {
	if rf, ok := w.base.(io.ReaderFrom); ok {
		return rf.ReadFrom(r)
	}
	return io.Copy(w.base, r)
}

// Hijack hijacks the connection from the HTTP handler so that it can be used directly (websockets, etc.)
// NOTE: This will fail if the wrapped http.responseRecorder is not a http.Hijacker.
func (w *responseRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if h, ok := w.base.(http.Hijacker); ok {
		return h.Hijack()
	}
	// Required for WebSockets to work
	return nil, nil, fmt.Errorf("response writer (%T) does not implement http.Hijacker", w.base)
}

// CloseNotify wraps the underlying CloseNotify or returns a dummy channel if the CloseNotifier interface is not implemented
func (w *responseRecorder) CloseNotify() <-chan bool {
	if cn, ok := w.base.(http.CloseNotifier); ok {
		return cn.CloseNotify()
	}

	// Return a dummy channel that will never get used
	return make(<-chan bool)
}

// responseRecorderFlusher wraps a base http.ResponseWriter/http.Flusher allowing extraction of additional inspection data
type responseRecorderFlusher struct {
	*responseRecorder
}

// Flush flushes data if the underlying http.ResponseWriter is capable of flushing
func (w *responseRecorderFlusher) Flush() {
	if f, ok := w.responseRecorder.base.(http.Flusher); ok {
		f.Flush()
	}
}

// ensure our writers satisfy the intended interfaces
var _ http.Hijacker = (*responseRecorder)(nil)
var _ io.ReaderFrom = (*responseRecorder)(nil)
var _ http.Flusher = (*responseRecorderFlusher)(nil)
