.PHONY: help all build static install uninstall clean test fmt vet lint coverage dist check-linux run

# Configuration
BINARY = reddock
VERSION = 2.22.5
OS := $(shell uname -s)
PREFIX ?= /usr/local
BINDIR = $(PREFIX)/bin
GO = go

# LDFLAGS for size optimization and version injection
LDFLAGS = -ldflags "-s -w -X reddock/cmd.Version=$(VERSION)"

check-linux:
	@if [ "$(OS)" != "Linux" ]; then \
		echo "Error: Reddock is only available for Linux systems"; \
		echo "Detected OS: $(OS)"; \
		exit 1; \
	fi

help:
	@echo "Reddock Makefile"
	@echo ""
	@echo "Build Targets:"
	@echo "  make build          - Build the binary (default)"
	@echo "  make static         - Build a static binary (no CGO)"
	@echo "  make dist           - Build distribution package"
	@echo ""
	@echo "Installation:"
	@echo "  make install        - Build and install to $(PREFIX)/bin"
	@echo "  make uninstall      - Remove installed binary"
	@echo ""
	@echo "Development:"
	@echo "  make run            - Run with original arguments (e.g. make run ARGS='list')"
	@echo "  make fmt            - Format code with gofmt"
	@echo "  make vet            - Run go vet for static analysis"
	@echo "  make lint           - Run golangci-lint"
	@echo "  make test           - Run tests"
	@echo "  make coverage       - Generate coverage report"
	@echo ""
	@echo "Maintenance:"
	@echo "  make clean          - Remove build artifacts"
	@echo ""
	@echo "Variables:"
	@echo "  PREFIX              - Installation prefix (default: /usr/local)"

all: build

build: check-linux
	@echo "Building Reddock..."
	$(GO) build $(LDFLAGS) -o $(BINARY) .
	@echo "Build complete: ./$(BINARY)"

static: check-linux
	@echo "Building static binary..."
	CGO_ENABLED=0 $(GO) build $(LDFLAGS) -o $(BINARY) .
	@echo "Static build complete: ./$(BINARY)"

dist: check-linux build
	@echo "Creating distribution package..."
	@mkdir -p dist
	tar -czf dist/$(BINARY)-linux-amd64.tar.gz $(BINARY) README.md LICENSE
	@echo "Distribution package: dist/$(BINARY)-linux-amd64.tar.gz"

run: build
	./$(BINARY) $(ARGS)

install: check-linux build
	@echo "Installing Reddock to $(DESTDIR)$(BINDIR)..."
	install -Dm755 $(BINARY) $(DESTDIR)$(BINDIR)/$(BINARY)
	@echo "Installation complete!"

uninstall:
	@echo "Uninstalling Reddock from $(DESTDIR)$(BINDIR)..."
	rm -f $(DESTDIR)$(BINDIR)/$(BINARY)
	@echo "Uninstall complete"

fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...

vet:
	@echo "Running go vet..."
	$(GO) vet ./...

lint:
	@echo "Running golangci-lint..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Install it for better checks." && exit 1)
	golangci-lint run ./...

test:
	@echo "Running tests..."
	$(GO) test -v ./...

coverage:
	@echo "Running tests with coverage..."
	$(GO) test -v -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

clean:
	@echo "Cleaning artifacts..."
	rm -f $(BINARY)
	rm -rf dist/
	rm -f coverage.out coverage.html
	@echo "Done"
