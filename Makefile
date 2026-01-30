.PHONY: help all build static install uninstall clean test fmt vet lint coverage dist check-linux

# Variables
OS := $(shell uname -s)
PREFIX ?= /usr/local
BINDIR = $(PREFIX)/bin
BINARY = redway
GO = go
GOFLAGS = -ldflags "-s -w"

# Build flags
LDFLAGS = -ldflags "-s -w"

# Linux check helper
check-linux:
	@if [ "$(OS)" != "Linux" ]; then \
		echo "Error: Redway is only available for Linux systems"; \
		echo "Detected OS: $(OS)"; \
		exit 1; \
	fi

help:
	@echo "Redway Makefile Targets"
	@echo ""
	@echo "Build Targets:"
	@echo "  make build          - Build the binary (default)"
	@echo "  make static         - Build a static binary (no CGO)"
	@echo "  make dist           - Build binaries for multiple platforms"
	@echo ""
	@echo "Installation:"
	@echo "  make install        - Build and install to $(PREFIX)/bin"
	@echo "  make uninstall      - Remove installed binary"
	@echo ""
	@echo "Development:"
	@echo "  make fmt            - Format code with gofmt"
	@echo "  make vet            - Run go vet for static analysis"
	@echo "  make lint           - Run golangci-lint (if installed)"
	@echo "  make test           - Run tests with verbose output"
	@echo "  make coverage       - Run tests with coverage report"
	@echo ""
	@echo "Maintenance:"
	@echo "  make clean          - Remove built binaries"
	@echo "  make help           - Show this help message"
	@echo ""
	@echo "Variables:"
	@echo "  PREFIX              - Installation prefix (default: /usr/local)"
	@echo "  DESTDIR             - Staging directory for install"
	@echo ""

all: build

build: check-linux
	@echo "Building Redway..."
	$(GO) build $(LDFLAGS) -o $(BINARY) .
	@echo "Build complete: $(BINARY)"

static: check-linux
	@echo "Building static binary..."
	CGO_ENABLED=0 $(GO) build $(LDFLAGS) -o $(BINARY) .
	@echo "Static build complete: $(BINARY)"

dist:
	@echo "Building distribution binaries..."
	@mkdir -p dist
	@for os in linux darwin windows; do \
		for arch in amd64 arm64; do \
			echo "  Building $$os/$$arch..."; \
			GOOS=$$os GOARCH=$$arch $(GO) build $(LDFLAGS) -o dist/$(BINARY)-$$os-$$arch . || true; \
		done; \
	done
	@echo "Distribution builds complete in dist/"

install: check-linux build
	@echo "Installing Redway..."
	install -Dm755 $(BINARY) $(DESTDIR)$(BINDIR)/$(BINARY)
	@echo ""
	@echo "Installation complete!"
	@echo ""
	@echo "Usage:"
	@echo "  sudo redway init                                    				# Initialize with default image"
	@echo "  sudo redway init docker://redroid/redroid:16.0.0_64only-latest		# Custom OCI image"
	@echo "  sudo redway start                                   				# Start container"
	@echo "  sudo redway adb-connect                             				# Get ADB info"
	@echo ""

uninstall:
	@echo "Uninstalling Redway..."
	rm -f $(DESTDIR)$(BINDIR)/$(BINARY)
	@echo "Uninstall complete"
	@echo ""
	@echo "Note: Config and data preserved. Remove manually if needed:"
	@echo "  rm -rf ~/.config/redway"
	@echo "  rm -rf ~/data-redroid"
	@echo ""

fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...
	@echo "Code formatted"

vet:
	@echo "Running go vet..."
	$(GO) vet ./...
	@echo "Vet check passed"

lint:
	@echo "Running golangci-lint..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" && exit 1)
	golangci-lint run ./...
	@echo "Lint check passed"

test:
	@echo "Running tests..."
	$(GO) test -v ./...
	@echo "Tests passed"

coverage:
	@echo "Running tests with coverage..."
	$(GO) test -v -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

clean:
	@echo "Cleaning..."
	rm -f $(BINARY)
	rm -rf dist/
	rm -f coverage.out coverage.html
	@echo "Clean complete"
