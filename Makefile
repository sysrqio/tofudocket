.PHONY: build test lint clean

VERSION ?= 0.1.0
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

build:
	go build $(LDFLAGS) -o bin/tofudocket ./cmd/tofudocket

test:
	go test ./...

lint:
	go vet ./...

clean:
	rm -rf bin/
