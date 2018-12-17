#!/bin/sh -ex

docker build -t foo  .

docker run -v ${PWD}:/go/src/github.com/signalsciences/sigsci-module-golang --rm foo ./scripts/build.sh
