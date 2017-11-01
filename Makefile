

build: ## build and lint locally
	./scripts/build.sh

# clean up each time to make sure nothing is cached between runs
#
test: ## build and run integration test
	./scripts/test.sh

init:  ## install gometalinter and msgp locally
	go get -u github.com/alecthomas/gometalinter
	gometalinter --install
	go get -u github.com/tinylib/msgp/msgp
	go get .


clean: ## cleanup
	rm -rf artifacts
	find . -name '*.log' | xargs rm -f
	go clean ./...
	git gc

# https://www.client9.com/self-documenting-makefiles/
help:
	@awk -F ':|##' '/^[^\t].+?:.*?##/ {\
	printf "\033[36m%-30s\033[0m %s\n", $$1, $$NF \
	}' $(MAKEFILE_LIST)
.DEFAULT_GOAL=help
.PHONY=help
