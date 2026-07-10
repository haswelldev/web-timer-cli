.PHONY: all build clean run test install update uninstall

BINARY_NAME=web-timer-cli
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X main.Version=$(VERSION)"

# Install location for system-wide install/update/uninstall.
PREFIX ?= /usr/local
BINDIR = $(PREFIX)/bin

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

# install builds the current tree and copies the binary into $(BINDIR),
# escalating with sudo only when that directory is not writable.
install: build
	@if [ -w "$(BINDIR)" ]; then \
		install -m 755 $(BINARY_NAME) "$(BINDIR)/$(BINARY_NAME)"; \
	else \
		echo "Installing to $(BINDIR) (sudo required)"; \
		sudo install -m 755 $(BINARY_NAME) "$(BINDIR)/$(BINARY_NAME)"; \
	fi
	@echo "Installed $(BINARY_NAME) to $(BINDIR)/$(BINARY_NAME)"

# update rebuilds from the current tree and overwrites the installed binary.
update: install

uninstall:
	@if [ -w "$(BINDIR)" ]; then \
		rm -f "$(BINDIR)/$(BINARY_NAME)"; \
	else \
		echo "Removing from $(BINDIR) (sudo required)"; \
		sudo rm -f "$(BINDIR)/$(BINARY_NAME)"; \
	fi
	@echo "Removed $(BINARY_NAME) from $(BINDIR)/$(BINARY_NAME)"

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
