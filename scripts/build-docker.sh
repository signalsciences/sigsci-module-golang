#!/bin/sh -ex

docker build -t foo  .

echo "***********************************"
echo `whoami`
ls -lah goroot
echo "***********************************"

rm -rf goroot
docker run -v ${PWD}:/go/src/github.com/signalsciences/sigsci-module-golang --rm foo ./scripts/build.sh
