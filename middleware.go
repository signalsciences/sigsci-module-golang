package sigsci

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Configure the Fastly module and attach it as global middleware
func Middleware(r *gin.Engine, options ...ModuleConfigOption) (*Module, error) {

	m, err := NewModule(nil, options...)
	if err != nil {
		return nil, err
	}
	r.Use(m.serveHTTP)
	return m, nil
}

// gin middleware handler
func (m *Module) serveHTTP(c *gin.Context) {

	// create copy to be used inside goroutine
	preReqCtx := c.Copy()
	req := preReqCtx.Request

	start := time.Now()
	finiwg := sync.WaitGroup{}

	// Use the inspector init/fini functions if available
	if m.inspInit != nil && !m.inspInit(req) {
		// No inspection is desired, so just defer to the underlying handler
		c.Next()
		return
	}
	if m.inspFini != nil {
		defer func() {
			// Delay the finalizer call until inspection (any pending Post
			// or Update call) is complete
			go func() {
				finiwg.Wait()
				m.inspFini(req)
			}()
		}()
	}

	if m.config.Debug() {
		log.Printf("DEBUG: calling 'RPC.PreRequest' for inspection: method=%s host=%s url=%s", req.Method, req.Host, req.URL)
	}
	inspin2, out, err := m.inspectorPreRequest(req)
	if err != nil {
		// Fail open
		if m.config.Debug() {
			log.Printf("ERROR: 'RPC.PreRequest' call failed (failing open): %s", err.Error())
		}
		c.Next()
		return
	}

	// w.Header()
	rw := NewResponseWriter(c.Writer)

	wafresponse := out.WAFResponse
	switch {
	case m.config.IsAllowCode(int(wafresponse)):
		// Continue with normal request
		c.Next()
	case m.config.IsBlockCode(int(wafresponse)):
		status := int(wafresponse)

		// Only redirect if it is a redirect status (3xx) AND there is a redirect URL
		if status >= 300 && status <= 399 {
			redirect := req.Header.Get("X-Sigsci-Redirect")
			if len(redirect) > 0 {
				http.Redirect(rw, req, redirect, status)
				break
			}
		}
		// Block
		http.Error(rw, fmt.Sprintf("%d %s\n", status, http.StatusText(status)), status)
	default:
		log.Printf("ERROR: Received invalid response code from inspector (failing open): %d", wafresponse)
		// Continue with normal request
		c.Next()
	}

	duration := time.Since(start)
	code := c.Writer.Status()
	size := int64(c.Writer.Size())

	if len(inspin2.RequestID) > 0 {
		// Do the UpdateRequest inspection in the background while the foreground hurries the response back to the end-user.
		inspin2.ResponseCode = int32(code)
		inspin2.ResponseSize = size
		inspin2.ResponseMillis = int64(duration / time.Millisecond)
		inspin2.HeadersOut = convertHeaders(c.Writer.Header())
		if m.config.Debug() {
			log.Printf("DEBUG: calling 'RPC.UpdateRequest' due to returned requestid=%s: method=%s host=%s url=%s code=%d size=%d duration=%s", inspin2.RequestID, req.Method, req.Host, req.URL, code, size, duration)
		}
		finiwg.Add(1) // Inspection finalizer will wait for this goroutine
		go func() {
			defer finiwg.Done()
			if err := m.inspectorUpdateRequest(inspin2); err != nil && m.config.Debug() {
				log.Printf("ERROR: 'RPC.UpdateRequest' call failed: %s", err.Error())
			}
		}()
	} else if code >= 300 || size >= m.config.AnomalySize() || duration >= m.config.AnomalyDuration() {
		// Do the PostRequest inspection in the background while the foreground hurries the response back to the end-user.
		if m.config.Debug() {
			log.Printf("DEBUG: calling 'RPC.PostRequest' due to anomaly: method=%s host=%s url=%s code=%d size=%d duration=%s", req.Method, req.Host, req.URL, code, size, duration)
		}
		inspin := NewRPCMsgIn(m.config, req, nil, code, size, duration)
		inspin.WAFResponse = wafresponse
		inspin.HeadersOut = convertHeaders(c.Writer.Header())

		finiwg.Add(1) // Inspection finializer will wait for this goroutine
		go func() {
			defer finiwg.Done()
			if err := m.inspectorPostRequest(inspin); err != nil && m.config.Debug() {
				log.Printf("ERROR: 'RPC.PostRequest' call failed: %s", err.Error())
			}
		}()
	}
}
