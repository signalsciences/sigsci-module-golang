

lint:
	golint clientcodec.go module.go module_test.go rpc.go
	gofmt -w -s *.go
	goimports -w *.go

rpc_gen.go: rpc.go
	go generate ./...

test:
	go build ./...
	go test ./...

clean:
	rm -f *~
	go clean ./...
