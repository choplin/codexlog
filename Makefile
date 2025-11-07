BINARY := agentlog
CMD_DIR := ./cmd/$(BINARY)
VERSION ?= $(shell git describe --tags --dirty --always 2>/dev/null || echo dev)
LDFLAGS := -X main.version=$(VERSION)

.PHONY: all build run test fmt lint clean

all: build

build:
	go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY) $(CMD_DIR)

run: build
	bin/$(BINARY)

test:
	go test ./...

fmt:
	golangci-lint run --fix ./...

lint:
	golangci-lint run ./...

clean:
	rm -rf bin
