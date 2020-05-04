#!/bin/sh

set -ex

# build / lint agent in a container
find . -name "goroot" -type d | xargs rm -rf
mkdir goroot


docker build -f Dockerfile.git -t golang-git:1.10.6-alpine3.8 .
docker run --user $(id -u ${USER}):$(id -g ${USER}) -v ${PWD}/goroot:/go/ --rm golang-git:1.10.6-alpine3.8 /bin/sh -c 'go get github.com/signalsciences/tlstext && go get github.com/tinylib/msgp && go get github.com/alecthomas/gometalinter'
./scripts/build-docker.sh

# run module tests
./scripts/test.sh


BASE=$PWD
## setup our package properties by distro
PKG_NAME="sigsci-module-golang"
DST_BUCKET="s3://package-build-artifacts/${PKG_NAME}/${GITHUB_RUN_NUMBER}"
VERSION=$(cat ./VERSION)


cd ${BASE}
echo "DONE"

# Main package
aws s3api put-object \
  --bucket "${DEST_BUCKET}" \
  --cache-control="max-age=300" \
  --content-type="application/octet-stream" \
  --body "./artifacts/${PKG_NAME}.tar.gz" \
  --key "${DST_BUCKET}/${PKG_NAME}_${VERSION}.tar.gz" \
  --grant-full-control id="${SIGSCI_PROD_CANONICAL_ID}"

# Metadata files
aws s3api put-object \
  --bucket "${DEST_BUCKET}" \
  --cache-control="max-age=300" \
  --content-type="text/plain; charset=UTF-8" \
  --body "VERSION" \
  --key "${DST_BUCKET}/VERSION" \
  --grant-full-control id="${SIGSCI_PROD_CANONICAL_ID}"

aws s3api put-object \
  --bucket "${DEST_BUCKET}" \
  --cache-control="max-age=300" \
  --content-type="text/plain; charset=UTF-8" \
  --body "CHANGELOG.md" \
  --key "${DST_BUCKET}/CHANGELOG.md" \
  --grant-full-control id="${SIGSCI_PROD_CANONICAL_ID}"






