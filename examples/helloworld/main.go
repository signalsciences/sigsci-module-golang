package main

import (
	"net/http"
	"net/http/httputil"

	sigsci "github.com/signalsciences/sigsci-module-golang"
)

func helloworld(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %q", "world")
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/helloworld", helloworld)

	s := &http.Server{
		Handler:        sigsci.NewAgentHandler(mux),
		Addr:           serverAddr,
	}
	log.Fatal(s.ListenAndServe())
}
