#!/bin/sh

set -ex

(cd ./scripts/test-golang110 && ./test.sh)
(cd ./scripts/test-golang111 && ./test.sh)
(cd ./scripts/test-golang114 && ./test.sh)
