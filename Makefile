

build:
	echo "package sigsci" > version.go
	echo "" >> version.go
	echo "const version = \"$(shell cat VERSION)\"" >> version.go
	go build .
	go test .

lint:
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

test: build

# not clear on vendoring in a library
# for now add a step for jenkins to load it in
init:
	go get -u github.com/alecthomas/gometalinter
	gometalinter --install
	go get .

generate:
	go generate ./...

clean:
	rm -rf artifacts
	go clean ./...
	git gc

release:
	rm -rf artifacts/
	mkdir -p artifacts/sigsci-module-golang
	cp -rf \
		VERSION CHANGELOG.md LICENSE.md README.md \
		clientcodec.go rpc.go rpc_gen.go module.go version.go \
		module_test.go \
		examples \
		artifacts/sigsci-module-golang/
	(cd artifacts; tar -czvf sigsci-module-golang.tar.gz sigsci-module-golang)
