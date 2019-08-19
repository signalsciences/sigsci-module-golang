package sigsci_test

import (
	"fmt"
	"log"
	"net/http"
	"time"

	sigsci "github.com/signalsciences/sigsci-module-golang"
)

func helloWorld(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %s!", r.URL.Path[1:])
}

func Example() {
	// Existing http.Handler
	mux := http.NewServeMux()
	mux.HandleFunc("/", helloWorld)

	// Wrap the existing http.Handler with the SigSci module handler
	wrapped, err := sigsci.NewModule(
		// Existing handler to wrap
		mux,

		// Any additional module options:
		sigsci.Socket("unix", "/var/run/sigsci.sock"),
		sigsci.Timeout(100*time.Millisecond),
		sigsci.AnomalySize(512*1024),
		sigsci.AnomalyDuration(1*time.Second),
		sigsci.MaxContentLength(100000),
	)

	if err != nil {
		log.Fatal(err)
	}

	// Listen and Serve as usual using the wrapped sigsci handler
	s := &http.Server{
		Handler: wrapped,
		Addr:    "localhost:8000",
	}

	log.Fatal(s.ListenAndServe())
}
