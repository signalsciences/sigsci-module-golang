#!/bin/sh

#if [ -z "${BUILD_NUMBER}" ]; then
#	echo "Must be run in Jenkins with BUILD_NUMBER set"
#	exit 2
#fi

BUILD_NUMBER=999

set -ex

# build / lint agent in a container
find . -name "goroot" -type d | xargs rm -rf
mkdir goroot


docker build -f Dockerfile.git -t golang-git:1.10.6-alpine3.8 .
docker run --user $(id -u ${USER}):$(id -g ${USER}) -v ${PWD}/goroot:/go/ --rm golang-git:1.10.6-alpine3.8 /bin/sh -c 'go get github.com/signalsciences/tlstext && go get github.com/tinylib/msgp && go get github.com/alecthomas/gometalinter'
./scripts/build-docker.sh

# run module tests

docker pull 803688608479.dkr.ecr.us-west-2.amazonaws.com/local-dev/module-testing:latest
docker tag 803688608479.dkr.ecr.us-west-2.amazonaws.com/local-dev/module-testing:latest local-dev/module-testing:latest
docker pull 803688608479.dkr.ecr.us-west-2.amazonaws.com/local-dev/sigsci-agent:latest
docker tag 803688608479.dkr.ecr.us-west-2.amazonaws.com/local-dev/sigsci-agent:latest  local-dev/sigsci-agent:latest
./scripts/test.sh


BASE=$PWD
## setup our package properties by distro
PKG_NAME="sigsci-module-golang"
DST_BUCKET="s3://package-build-artifacts/${PKG_NAME}/${BUILD_NUMBER}"
VERSION=$(cat ./VERSION)


cd ${BASE}
echo "DONE"
#aws s3 cp \
#        --no-follow-symlinks \
#        --cache-control="max-age=300" \
#        ./artifacts/${PKG_NAME}.tar.gz ${DST_BUCKET}/${PKG_NAME}_${VERSION}.tar.gz
#
#aws s3 cp \
#        --no-follow-symlinks \
#        --cache-control="max-age=300" \
#        --content-type="text/plain; charset=UTF-8" \
#        VERSION ${DST_BUCKET}/VERSION
#
#aws s3 cp \
#        --no-follow-symlinks \
#        --cache-control="max-age=300" \
#        --content-language="en-US" \
#        --content-type="text/markdown; charset=UTF-8" \
#        CHANGELOG.md ${DST_BUCKET}/CHANGELOG.md
#
