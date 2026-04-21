.PHONY: all build clean run test install

BINARY_NAME=web-timer-cli
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X main.Version=$(VERSION)"

all: build

build:
	go build $(LDFLAGS) -o $(BINARY_NAME)

clean:
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME)-*

run: build
	./$(BINARY_NAME)

test:
	go test -v ./...

install:
	go install $(LDFLAGS)

# Cross-platform builds
build-linux:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)-linux-amd64

build-darwin-amd64:
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)-darwin-amd64

build-darwin-arm64:
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BINARY_NAME)-darwin-arm64

build-windows:
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)-windows-amd64.exe

build-all: build-linux build-darwin-amd64 build-darwin-arm64 build-windows

# Format code
fmt:
	go fmt ./...

# Run linter
lint:
	golangci-lint run

# Download dependencies
deps:
	go mod download
	go mod tidy
