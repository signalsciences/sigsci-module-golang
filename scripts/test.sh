#!/bin/sh
set -xe
rm -f *.log
docker-compose build
docker-compose pull

cleanup() {
  echo "shutting down"
  docker-compose logs --no-color agent >& agent.log
  docker-compose logs --no-color web >& web.log
  docker-compose down
}
trap cleanup 0 1 2 3 6

docker-compose run \
	--entrypoint ./scripts/build.sh web

docker-compose up --no-color -d

docker-compose run \
        -e DISABLE_HTTP_OPTIONS=1 \
        -e MTEST_BASEURL=web:8085 \
        -e MTEST_AGENT=agent:12345 \
	--entrypoint /bin/wait-for mtest web:8085 -- /bin/mtest -test.v
