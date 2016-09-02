# sigsci-module-golang
Signal Sciences Module in golang

This is currently in beta, mostly due to packaging and support issues.
In time, this will be posted on github.

```
// use your existing http.Handler
mux := http.NewServeMux()
mux.HandleFunc("/helloworld", helloworld)

// wrap it 
wrapped, err := sigsci.NewModule(mux,
    sigsci.Timeout(50 * time.Millisecond)
)

// Listen and Serve as usual using the wrapper
// sigsci handler
s := &http.Server{
    Handler: wrapped,
    Addr:    "127.0.0.1:9999",
}
log.Fatal(s.ListenAndServe())
```
