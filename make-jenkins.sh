#!/bin/sh

if [ -z "${BUILD_NUMBER}" ]; then
	echo "Must be run in Jenkins with BUILD_NUMBER set"
	exit 2
fi

set -ex
export GOPATH=$WORKSPACE
export GOROOT="/opt/go"
export PATH="/opt/go/bin:$WORKSPACE/bin:/opt/local/bin:$PATH"


BASE=$PWD
## setup our package properties by distro
PKG_NAME="sigsci-module-golang"
DST_BUCKET="s3://package-build-artifacts/${PKG_NAME}/${BUILD_NUMBER}"
VERSION=$(cat ./VERSION)

# make init
go get -u github.com/client9/tlstext
go get -u gopkg.in/fatih/pool.v2

make build
make release

cd ${BASE}
aws s3 cp \
        --no-follow-symlinks \
        --cache-control="max-age=300" \
        ./artifacts/${PKG_NAME}.tar.gz ${DST_BUCKET}/${PKG_NAME}_${VERSION}.tar.gz

aws s3 cp \
        --no-follow-symlinks \
        --cache-control="max-age=300" \
        --content-type="text/plain; charset=UTF-8" \
        VERSION ${DST_BUCKET}/VERSION

aws s3 cp \
        --no-follow-symlinks \
        --cache-control="max-age=300" \
        --content-language="en-US" \
        --content-type="text/markdown; charset=UTF-8" \
        CHANGELOG.md ${DST_BUCKET}/CHANGELOG.md

