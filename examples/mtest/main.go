package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	sigsci "github.sigsci.in/engineering/sigsci-module-golang"
)

func helloworld(w http.ResponseWriter, r *http.Request) {
	delay := 0
	body := []byte("OK")
	code := 200

	var err error
	var qs url.Values
	if r.URL != nil {
		qs, err = url.ParseQuery(r.URL.RawQuery)
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
		body, err = ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("ioutil.ReadAll erred: %s", err)
		}
	}

	if delay > 0 {
		time.Sleep(time.Millisecond * time.Duration(delay))
	}
	if code >= 300 && code < 400 {
		w.Header().Set("Location", "/foo")
	}
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	w.WriteHeader(code)
	w.Write(body)
}

func main() {
	mux := http.NewServeMux()

	// "/" is handle everything
	mux.HandleFunc("/response", helloworld)

	h, err := sigsci.NewModule(mux,
		sigsci.Socket("tcp", "agent:9090"),
		sigsci.Timeout(1000*time.Millisecond),
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
