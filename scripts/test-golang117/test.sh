#!/bin/bash
set -e

DOCKERCOMPOSE="docker-compose"

# run at end no matter what
cleanup() {
  echo "shutting down"
  # capture log output
  $DOCKERCOMPOSE logs --no-color agent >& agent.log
  $DOCKERCOMPOSE logs --no-color web >& web.log
  $DOCKERCOMPOSE logs --no-color mtest >& mtest.log
  $DOCKERCOMPOSE logs --no-color punchingbag >& punchingbag.log

  # delete everything
  $DOCKERCOMPOSE down

  # show output of module testing
  cat mtest.log
}
trap cleanup 0 1 2 3 6

set -x

# attempt to clean up any leftover junk
$DOCKERCOMPOSE down

$DOCKERCOMPOSE pull --ignore-pull-failures

# start everything, run tests
#
# --no-color --> safe for ci
# --build    --> alway build test server/module container
# --abort-on-container-exit --> without this, the other servers keep the process running
# --exit-code-from mtest -->  make exit code be the result of module test
#
# > /dev/null  --> output of all servers is mixed together and ugly
#                  we get the individual logs at end
#
$DOCKERCOMPOSE up --no-color --build  --abort-on-container-exit --exit-code-from mtest > /dev/null

