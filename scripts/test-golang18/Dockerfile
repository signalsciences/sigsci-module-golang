FROM golang:1.8.5-alpine3.6

RUN apk --update add git
RUN /bin/true \
    && go get github.com/signalsciences/tlstext \
    && go get github.com/tinylib/msgp/... \
    && go get github.com/alecthomas/gometalinter \
    && gometalinter --install 

# we will mount the current directory here
VOLUME [ "/go/src/github.com/signalsciences/sigsci-module-golang" ]
WORKDIR /go/src/github.com/signalsciences/sigsci-module-golang
