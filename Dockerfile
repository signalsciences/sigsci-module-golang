FROM golang:1.9.0-alpine3.6

RUN apk --update add git
RUN /bin/true \
    && go get github.com/signalsciences/tlstext \
    && go get gopkg.in/fatih/pool.v2 \
    && go get github.com/tinylib/msgp/msgp \
    && go get github.com/alecthomas/gometalinter \
    && gometalinter --install 

# why copy this in when we can just mount it?
RUN mkdir -p /go/src/github.com/signalsciences/sigsci-module-golang/examples/mtest
WORKDIR /go/src/github.com/signalsciences/sigsci-module-golang
COPY VERSION CHANGELOG.md LICENSE.md README.md clientcodec.go rpc.go rpc_gen.go module.go version.go module_test.go ./
COPY examples/mtest/main.go ./examples/mtest
