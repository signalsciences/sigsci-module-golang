package main

import (
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	sigsci "github.com/signalsciences/sigsci-module-golang"
	"github.com/signalsciences/sigsci-module-golang/schema"
)

// convertHeaders converts http.Header to the format expected by sigsci
func convertHeaders(h http.Header) [][2]string {
	out := make([][2]string, 0, len(h))
	for key, values := range h {
		for _, value := range values {
			out = append(out, [2]string{key, value})
		}
	}
	return out
}

func GetHeader(headers [][2]string, key string) (string, bool) {
	key = strings.ToLower(key)
	for _, h := range headers {
		if strings.ToLower(h[0]) == key {
			return h[1], true
		}
	}
	return "", false
}

// Configure the Signal Sciences module and attach it as global middleware
func Middleware(r *gin.Engine, options ...sigsci.ModuleConfigOption) (*sigsci.Module, error) {
	// Create a dummy handler since we won't use it directly
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {})

	m, err := sigsci.NewModule(dummyHandler, options...)
	if err != nil {
		return nil, err
	}

	// Add the actual gin middleware
	r.Use(func(c *gin.Context) {
		start := time.Now()
		finiwg := sync.WaitGroup{}

		// Create input for PreRequest call
		msgIn := sigsci.NewRPCMsgIn(m.ModuleConfig(), c.Request, nil, 0, 0, 0)
		msgOut := &sigsci.RPCMsgOut{}

		// Call PreRequest inspection
		if err := m.Inspector().PreRequest(msgIn, msgOut); err != nil {
			if m.ModuleConfig().Debug() {
				log.Printf("ERROR: PreRequest call failed (failing open): %s", err.Error())
			}
			c.Next()
			return
		}

		// Add RespAction response headers
		for _, a := range msgOut.RespActions {
			switch a.Code {
			case schema.AddHdr:
				c.Writer.Header().Add(a.Args[0], a.Args[1])
			case schema.SetHdr:
				c.Writer.Header().Set(a.Args[0], a.Args[1])
			case schema.SetNEHdr:
				if len(c.Writer.Header().Get(a.Args[0])) == 0 {
					c.Writer.Header().Set(a.Args[0], a.Args[1])
				}
			case schema.DelHdr:
				c.Writer.Header().Del(a.Args[0])
			}
		}
		// Check WAF response
		wafResponse := msgOut.WAFResponse
		switch {
		case m.ModuleConfig().IsAllowCode(int(wafResponse)):
			// Continue with gin request processing
			c.Next()
		case m.ModuleConfig().IsBlockCode(int(wafResponse)):
			status := int(wafResponse)
			// Handle redirects
			if status >= 300 && status <= 399 {
				if v, ok := GetHeader(msgOut.RequestHeaders, "X-Sigsci-Redirect"); ok {
					c.Redirect(status, v)
					break
				}
			}
			c.AbortWithStatusJSON(status, gin.H{
				"error":   http.StatusText(status),
				"blocked": "Request blocked by Fastly",
			})

		default:
			log.Printf("ERROR: Received invalid response code from inspector (failing open): %d", wafResponse)
			c.Next()
		}

		// After request processing, handle PostRequest/UpdateRequest
		// Only do this after c.Next() completes successfully
		duration := time.Since(start)
		responseCode := c.Writer.Status()
		responseSize := int64(c.Writer.Size())

		// Do post-processing in background to avoid interfering with response
		// go func() {
		if len(msgOut.RequestID) > 0 {
			// Use UpdateRequest if we have a RequestID
			updateMsg := &sigsci.RPCMsgIn2{
				RequestID:      msgOut.RequestID,
				ResponseCode:   int32(responseCode),
				ResponseSize:   responseSize,
				ResponseMillis: int64(duration / time.Millisecond),
				HeadersOut:     convertHeaders(c.Writer.Header()),
			}

			var updateOut sigsci.RPCMsgOut
			finiwg.Add(1) // Inspection finializer will wait for this goroutine
			go func() {
				defer finiwg.Done()
				if err := m.Inspector().UpdateRequest(updateMsg, &updateOut); err != nil && m.ModuleConfig().Debug() {
					log.Printf("ERROR: UpdateRequest call failed: %s", err.Error())
				}
			}()
		} else if responseCode >= 300 || responseSize >= m.ModuleConfig().AnomalySize() || duration >= m.ModuleConfig().AnomalyDuration() {

			// Use PostRequest for anomalies
			postMsg := sigsci.NewRPCMsgIn(m.ModuleConfig(), c.Request, nil, responseCode, responseSize, duration)
			postMsg.WAFResponse = wafResponse
			postMsg.HeadersOut = convertHeaders(c.Writer.Header())

			var postOut sigsci.RPCMsgOut
			finiwg.Add(1) // Inspection finializer will wait for this goroutine
			go func() {
				defer finiwg.Done()
				if err := m.Inspector().PostRequest(postMsg, &postOut); err != nil && m.ModuleConfig().Debug() {
					log.Printf("ERROR: PostRequest call failed: %s", err.Error())
				}
			}()
		}
	})

	return m, nil
}
