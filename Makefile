.PHONY: build test lint fmt install clean ci

BINARY=localvault
VERSION=0.1.0

build:
	go build -ldflags "-X github.com/xilistudios/localvault/cmd.version=$(VERSION)" -o bin/$(BINARY) .

test:
	go test ./... -v -count=1

lint:
	golangci-lint run ./...

fmt:
	gofmt -s -w .
	goimports -w .

install:
	go install -ldflags "-X github.com/xilistudios/localvault/cmd.version=$(VERSION)" .

clean:
	rm -rf bin/

ci: fmt lint test build
