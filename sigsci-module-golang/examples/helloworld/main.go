package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	sigsci "github.com/signalsciences/sigsci-module-golang"
)

func helloworld(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %q\n", "world")
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/helloworld", helloworld)

	h, err := sigsci.NewModule(mux, sigsci.Timeout(50 * time.Millisecond))
	if err != nil {
		log.Fatal(err)
	}
	s := &http.Server{
		Handler: h,
		Addr:    "127.0.0.1:9999",
	}
	log.Fatal(s.ListenAndServe())
}
