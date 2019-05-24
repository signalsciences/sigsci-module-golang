package sigsci

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
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
	// NOTE: according to net/http docs, if WriteHeader is not called explicitly,
	// the first call to Write will trigger an implicit WriteHeader(http.StatusOK).
	// this is why the default code is 200 and it only changes if WriteHeader is called.
	w := &responseRecorder{
		base: base,
		code: 200,
	}
	if _, ok := w.base.(http.Flusher); ok {
		return &responseRecorderFlusher{w}
	}
	return w
}

// responseRecorder wraps a base http.ResponseWriter allowing extraction of additional inspection data
type responseRecorder struct {
	base http.ResponseWriter
	code int
	size int64
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
	w.code = status
	w.base.WriteHeader(status)
}

// Write writes data, tracking the length written for inspection
func (w *responseRecorder) Write(b []byte) (int, error) {
	w.size += int64(len(b))
	return w.base.Write(b)
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
