#!/bin/sh -ex

docker build -t foo  .

docker run -v ${PWD}:/go/src/github.sigsci.in/engineering/sigsci-module-golang --rm foo ./scripts/build.sh
