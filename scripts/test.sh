#!/bin/sh

set -ex

go clean ./...
(cd scripts/test-golang19 && ./test.sh)

go clean ./...
(cd scripts/test-golang18 && ./test.sh)

