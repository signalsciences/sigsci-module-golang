#!/bin/sh
set -ex

echo "package sigsci" > version.go
echo "" >> version.go
echo "const version = \"$(cat VERSION)\"" >> version.go
find . -name "goroot" -type d | xargs rm -rf
go generate ./...

# make sure files made in docker are readable by all
chmod a+r *.go

go build .
go test .

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
        --enable=vet \
        --exclude=_gen.go \
	--exclude=/usr/local/go/src/net/lookup_unix.go \
        .

rm -rf artifacts/
mkdir -p artifacts/sigsci-module-golang
cp -rf \
  VERSION CHANGELOG.md LICENSE.md README.md \
  clientcodec.go rpc.go rpc_gen.go rpcinspector.go inspector.go module.go version.go \
  module_test.go \
  examples \
  artifacts/sigsci-module-golang/

# mtest is internal only
rm -fr artifacts/sigsci-module-golang/examples/mtest

(cd artifacts; tar -czvf sigsci-module-golang.tar.gz sigsci-module-golang)
chmod a+rw artifacts
