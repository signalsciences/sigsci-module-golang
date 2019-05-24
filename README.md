# sigsci-module-golang

Signal Sciences module in golang

## Dependencies

The Golang module requires two prerequisite packages to be installed: [MessagePack Code Generator](https://github.com/tinylib/msgp/) and the Signal Sciences custom [tlstext](https://github.com/signalsciences/tlstext) package.

The easiest way to install these packages is by using the go get command to download and install these packages directly from their GitHub repositories:

```bash
go get -u -t github.com/tinylib/msgp/msgp
go get -u -t github.com/signalsciences/tlstext
```

## Usage

The golang module is written as a http.Handler wrapper. To use the module, you
will need to wrap your existing handler with the module handler.

Example:
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
