package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	sigsci "github.com/signalsciences/sigsci-module-golang"
)

var (
	debug = false
)

func helloworld(w http.ResponseWriter, r *http.Request) {
	if debug {
		reqbytes, _ := httputil.DumpRequest(r, true)
		reqstr := string(reqbytes)
		if !strings.HasSuffix(reqstr, "\n") {
			reqstr += "\n"
		}
		log.Printf("REQUEST %s:\n%s", r.Header.Get("Content-Type"), reqstr)
	}
	delay := 0
	body := []byte("OK")
	code := 200

	var err error
	var qs url.Values
	if r.URL != nil {
		qs, err = url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			log.Println("url.ParseQuery err:", err)
		}
	}
	if qs == nil {
		qs = make(url.Values)
	}

	if num, err := strconv.Atoi(qs.Get("response_time")); err == nil {
		delay = num
	}
	if num, err := strconv.Atoi(qs.Get("response_code")); err == nil {
		code = num
	}
	if num, err := strconv.Atoi(qs.Get("size")); err == nil {
		body = bytes.Repeat([]byte{'a'}, num)
	}
	if len(qs.Get("echo")) > 0 {
		body, err = io.ReadAll(r.Body)
		if err != nil {
			log.Printf("ioutil.ReadAll erred: %s", err)
		}
		r.Body = io.NopCloser(bytes.NewReader(body))
	}

	if delay > 0 {
		time.Sleep(time.Millisecond * time.Duration(delay))
	}
	if code >= 300 && code < 400 {
		w.Header().Set("Location", "/foo")
	}
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))

	// Populate varX response headers from form values for mtest form processing
	if varq := r.FormValue("varq"); len(varq) > 0 {
		w.Header().Add("varq", varq)
	}
	if varb := r.PostFormValue("varb"); len(varb) > 0 {
		w.Header().Add("varb", varb)
	}
	if place := r.PostFormValue("place"); len(place) > 0 {
		w.Header().Add("place", place)
	}

	if debug {
		// Record the response so it can be logged
		wrec := httptest.NewRecorder()
		for k := range w.Header() {
			v := w.Header().Get(k)
			wrec.Header().Set(k, v)
		}
		wrec.WriteHeader(code)
		wrec.Write(body)
		resp := wrec.Result()
		respbytes, _ := httputil.DumpResponse(resp, true)
		respstr := string(respbytes)
		if !strings.HasSuffix(respstr, "\n") {
			respstr += "\n"
		}
		log.Printf("RESPONSE:\n%s", respstr)
	}
	w.WriteHeader(code)
	w.Write(body)
}

func main() {
	if dbg := os.Getenv("DEBUG"); len(dbg) > 0 && dbg != "0" {
		debug = true
	}

	mux := http.NewServeMux()

	// "/" is handle everything
	mux.HandleFunc("/response", helloworld)

	h, err := sigsci.NewModule(mux,
		sigsci.Socket("tcp", "agent:9090"),
		sigsci.Timeout(1500*time.Millisecond),
		// Match agent defaults and better deal with mtest behavior test defaults
		sigsci.MaxContentLength(300*1024),
		sigsci.AnomalySize(512*1024),
	)
	if err != nil {
		log.Fatal(err)
	}
	s := &http.Server{
		Handler: h,
		Addr:    "0.0.0.0:8085",
	}
	log.Fatal(s.ListenAndServe())
}
