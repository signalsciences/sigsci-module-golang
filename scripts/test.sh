#!/bin/sh

set -ex

(cd scripts/test-golang19 && ./test.sh)
(cd scripts/test-golang18 && ./test.sh)

