

lint:
	echo "package sigsci" > version.go
	echo "" >> version.go
	echo "const version = \"$(shell cat VERSION)\"" >> version.go
	go build .
	gometalinter \
		--vendor \
		--deadline=60s \
		--disable-all \
		--enable=vetshadow \
		--enable=ineffassign \
		--enable=deadcode \
		--enable=golint \
		--enable=gofmt \
		--enable=gosimple \
		--enable=unused \
		--exclude=_gen.go \
		.

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
