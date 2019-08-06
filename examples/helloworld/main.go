package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	// Import the module
	sigsci "github.com/signalsciences/sigsci-module-golang"
)

func main() {
	// Process sigsci-agent rpc-address if passed
	sigsciAgentNetwork := "unix"
	sigsciAgentAddress := "/var/run/sigsci.sock"
	if len(os.Args) > 1 {
		sigsciAgentAddress = os.Args[1]
	}
	if !strings.Contains(sigsciAgentAddress, "/") {
		sigsciAgentNetwork = "tcp"
	}
	log.Printf("Using sigsci-agent address (pass address as program argument to change): %s:%s", sigsciAgentNetwork, sigsciAgentAddress)

	// Existing handler, in this case a simple http.ServeMux,
	// but could be any http.Handler in the application
	mux := http.NewServeMux()
	mux.HandleFunc("/", helloworld)

	// Wrap the existing http.Handler with the SigSci module handler
	wrapped, err := sigsci.NewModule(
		// Existing handler to wrap
		mux,

		// Any additional module options:
		sigsci.Socket(sigsciAgentNetwork, sigsciAgentAddress),
		//sigsci.Timeout(100 * time.Millisecond),
		//sigsci.AnomalySize(512 * 1024),
		//sigsci.AnomalyDuration(1 * time.Second),
		//sigsci.MaxContentLength(100000),

		// Turn on debug logging for this example (do not use in production)
		sigsci.Debug(true),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Listen and Serve as usual using the wrapped sigsci handler
	s := &http.Server{
		Handler: wrapped,
		Addr:    "localhost:8000",
	}
	log.Printf("Server URL: http://%s/", s.Addr)
	log.Fatal(s.ListenAndServe())
}

// helloworld just displays a banner message for testing
func helloworld(w http.ResponseWriter, r *http.Request) {
	status := http.StatusOK
	w.WriteHeader(status)
	w.Write([]byte(`<!DOCTYPE html>
<html lang="en">
<head><title>Hello World</title></head>
<body><h1>Hello World!</h1></body>
</html>
`))
}
