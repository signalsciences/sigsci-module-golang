
[![grc][grc-img]][grc] [![GoDoc][doc-img]][doc]

# sigsci-module-golang

The Signal Sciences module in Golang allows for integrating your Golang
application directly with the Signal Sciences agent at the source code
level. It is written as a `http.Handler` wrapper. To
integrate your application with the module, you will need to wrap your
existing handler with the module handler.

## :rotating_light: NOTICE :rotating_light:

Effective **May 17th 2021** the default branch will change from `master` to `main`. Run the following commands to update a local clone:
```
git branch -m master main
git fetch origin
git branch -u origin/main main
git remote set-head origin -a
```

## Installation

```console
go get github.com/signalsciences/sigsci-module-golang@latest
```

## Example Code Snippet
```go
// Example existing http.Handler
mux := http.NewServeMux()
mux.HandleFunc("/", helloworld)

// Wrap the existing http.Handler with the SigSci module handler
wrapped, err := sigsci.NewModule(
    // Existing handler to wrap
    mux,

    // Any additional module options:
    //sigsci.Socket("unix", "/var/run/sigsci.sock"),
    //sigsci.Timeout(100 * time.Millisecond),
    //sigsci.AnomalySize(512 * 1024),
    //sigsci.AnomalyDuration(1 * time.Second),
    //sigsci.MaxContentLength(100000),
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
```

## Examples

The [examples/helloworld](examples/helloworld) directory contains complete example code.

To run the simple [helloworld](examples/helloworld/main.go) example:
```shell
# Syntax:
#   go run ./examples/helloworld <listener-address:port> [<sigsci-agent-rpc-address>]
#
# Run WITHOUT sigsci enabled
go run ./examples/helloworld localhost:8000
# Run WITH sigsci-agent listening via a UNIX Domain socket file
go run ./examples/helloworld localhost:8000 /var/run/sigsci.sock
# Run WITH sigsci-agent listening via a TCP address:port
go run ./examples/helloworld localhost:8000 localhost:9999
```

The above will run a HTTP listener on `localhost:8000`, which will send any
traffic to this listener to a running sigsci-agent for inspection (if
an agent address is configured).

[doc-img]: https://godoc.org/github.com/signalsciences/sigsci-module-golang?status.svg
[doc]: https://godoc.org/github.com/signalsciences/sigsci-module-golang
[grc-img]: https://goreportcard.com/badge/github.com/signalsciences/sigsci-module-golang 
[grc]: https://goreportcard.com/report/github.com/signalsciences/sigsci-module-golang
