# sigsci-module-golang

[![Go Report Card](https://goreportcard.com/badge/github.com/signalsciences/sigsci-module-golang)](https://goreportcard.com/report/github.com/signalsciences/sigsci-module-golang)

The Signal Sciences module in golang allows for integrating your golang
application directly with the Signal Sciences agent at the source code
level. This golang module is written as a `http.Handler` wrapper. To
integrate your application with the module, you will need to wrap your
existing handler with the module handler.

Example Code Snippet:
```go
// Existing http.Handler
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

## Dependencies

The golang module requires two prerequisite packages to be installed:
[MessagePack Code Generator](https://github.com/tinylib/msgp/) and the
Signal Sciences custom [tlstext](https://github.com/signalsciences/tlstext)
package.

The easiest way to install these packages is by using the `go get`
command to download and install these packages directly from their
public GitHub repositories:

```bash
go get -u -t github.com/tinylib/msgp/msgp
go get -u -t github.com/signalsciences/tlstext
```

## Examples

The [examples](examples/) directory contains complete example code.

To run the simple [helloworld](examples/helloworld/main.go) example:

```bash
go run examples/helloworld/main.go
```

Or, if your agent is running with a non-default `rpc-address`, you can
pass the sigsci-agent address as an argument such as one of the following:

```bash
# Another UNIX Domain socket
go run examples/helloworld/main.go /tmp/sigsci.sock
# A TCP address:port
go run examples/helloworld/main.go localhost:9999
```

This will run a HTTP listener on `localhost:8000`, which will send any
traffic to this listener to a running sigsci-agent for inspection.
