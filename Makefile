.PHONY: build test clean install lint

BINARY_NAME=caddy-mcp-server
GO=go
GOFLAGS=-ldflags="-s -w"

build:
	$(GO) build $(GOFLAGS) -o $(BINARY_NAME) .

test:
	./test.sh

clean:
	rm -f $(BINARY_NAME)
	$(GO) clean

install: build
	cp $(BINARY_NAME) $(GOPATH)/bin/

lint:
	$(GO) vet ./...
	gofmt -l .

dev:
	$(GO) run .

all: clean build test
