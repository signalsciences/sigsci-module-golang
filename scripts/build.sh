#!/bin/sh
set -ex

which go
go version
eval $(go env | grep GOROOT)
export GOROOT
export CGO_ENABLED=0

go build .
go test .

#### Run the linter
#if [ -z "$(which gometalinter)" ]; then
#  go get github.com/alecthomas/gometalinter
#fi
#gometalinter \
#        --vendor \
#        --deadline=60s \
#        --disable-all \
#        --enable=vetshadow \
#        --enable=ineffassign \
#        --enable=deadcode \
#        --enable=golint \
#        --enable=gofmt \
#        --enable=vet \
#        --exclude=_gen.go \
#        --exclude=/usr/local/go/src/net/lookup_unix.go \
#        .

rm -rf artifacts/
mkdir -p artifacts/sigsci-module-golang
cp -rf \
  VERSION CHANGELOG.md LICENSE.md README.md \
  clientcodec.go rpc.go rpc_gen.go rpcinspector.go inspector.go responsewriter.go module.go version.go \
  responsewriter_test.go module_test.go \
  examples \
  artifacts/sigsci-module-golang/

# mtest is internal only
rm -fr artifacts/sigsci-module-golang/examples/mtest

(cd artifacts; tar -czvf sigsci-module-golang.tar.gz sigsci-module-golang)
chmod a+rw artifacts
