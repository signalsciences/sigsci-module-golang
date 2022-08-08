#!/bin/sh

set -ex

(cd ./scripts/test-golang117 && ./test.sh)
(cd ./scripts/test-golang118 && ./test.sh)
(cd ./scripts/test-golang119 && ./test.sh)
