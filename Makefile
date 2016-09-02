

lint:
	echo "package sigsci" > version.go
	echo "const Version = \"sigsci-module-golang $(shell cat VERSION)\"" >> version.go
	go build .
	golint clientcodec.go module.go module_test.go rpc.go
	gofmt -w -s *.go
	goimports -w *.go

rpc_gen.go: rpc.go
	go generate ./...

test:
	go build ./...
	go test ./...

clean:
	go clean ./...
	rm -fr sigsci-module-golang sigsci-module-golang.tar.gz
	git gc

release:
	rm -rf sigsci-module-golang
	mkdir sigsci-module-golang
	cp -rf clientcodec.go rpc.go rpc_gen.go module.go clientcodec.go examples sigsci-module-golang/
	tar -czvf sigsci-module-golang.tar.gz sigsci-module-golang
