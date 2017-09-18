package main

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	sigsci "github.com/signalsciences/sigsci-module-golang"
)

func helloworld(w http.ResponseWriter, r *http.Request) {
	delay := 0
	body := "OK"
	code := 200

	err := r.ParseForm()
	if err != nil {

	}

	if num, err := strconv.Atoi(r.Form.Get("response_time")); err == nil {
		delay = num
	}
	if num, err := strconv.Atoi(r.Form.Get("response_code")); err == nil {
		code = num
	}
	if num, err := strconv.Atoi(r.Form.Get("size")); err == nil {
		body = strings.Repeat("a", num)
	}
	if len(r.Form.Get("echo")) > 0 {
		// TBD
	}
	if delay > 0 {
		time.Sleep(time.Millisecond * time.Duration(delay))
	}

	if code >= 300 && code < 400 {
		w.Header().Set("Location", "/foo")
	}
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(code)
	w.Write([]byte(body))
}

func main() {
	mux := http.NewServeMux()

	// "/" is handle everything
	mux.HandleFunc("/", helloworld)

	h, err := sigsci.NewModule(mux,
		sigsci.Socket("tcp", "agent:9090"),
		sigsci.Timeout(50*time.Millisecond),
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
