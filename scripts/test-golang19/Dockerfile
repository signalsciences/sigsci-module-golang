FROM golang:1.9.2-alpine3.6

RUN apk --update add git
RUN /bin/true \
    && go get github.com/signalsciences/tlstext \
    && go get github.com/tinylib/msgp/...

# we will mount the current directory here
VOLUME [ "/go/src/github.sigsci.in/engineering/sigsci-module-golang" ]
WORKDIR /go/src/github.sigsci.in/engineering/sigsci-module-golang
