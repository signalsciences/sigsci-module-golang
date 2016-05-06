

lint:
	find . -name '*.go' | grep -v _gen.go | xargs golint
	gofmt -w -s *.go
	goimports -w *.go

test:
	go build ./...
	go test ./...

clean:
	rm -f *~
	go clean ./...
