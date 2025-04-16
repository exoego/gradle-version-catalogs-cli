.PHONY: build check test
build:
	go build

check:
	goreleaser check

test:
	go test -v ./...
