.PHONY: help all build static install uninstall clean test fmt vet lint coverage dist dist-pack check-linux run

# Configuration
BINARY = reddock
# Version: override with make VERSION=…
# Default: git describe from annotated/lightweight tags (matches GitHub tags after push); else <commit-count>-<ddmmyy>
GIT_COUNT := $(shell git rev-list --count HEAD 2>/dev/null || echo 0)
BUILD_DATE := $(shell date +%d%m%y)
GIT_DESC := $(shell git describe --tags --always --dirty 2>/dev/null)
VERSION ?= $(if $(GIT_DESC),$(GIT_DESC),$(GIT_COUNT)-$(BUILD_DATE))
OS := $(shell uname -s)
PREFIX ?= /usr/local
BINDIR ?= $(PREFIX)/bin
DESTDIR ?=
GO ?= go

# Installation path (what install/uninstall use)
INSTALLED = $(DESTDIR)$(BINDIR)/$(BINARY)

# Verbose builds: by default lists compiled packages (-v). QUIET=1 disables that; VERBOSE=1 or V=1 adds -x.
V ?= 0
VERBOSE ?= $(V)
GO_VERBOSE :=
ifeq ($(QUIET),1)
GO_VERBOSE :=
else ifeq ($(VERBOSE),1)
GO_VERBOSE += -v -x
else
GO_VERBOSE += -v
endif

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
	@echo "  make dist           - static + .tar.xz (binary + README + LICENSE)"
	@echo "  make dist-pack      - .tar.xz only (requires ./$(BINARY) from make static)"
	@echo ""
	@echo "Installation:"
	@echo "  make install        - Build and install to $(INSTALLED)"
	@echo "  make uninstall      - Remove $(INSTALLED)"
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
	@echo "  VERSION             - Embedded version (git describe from tags, else count+date)"
	@echo "  PREFIX              - Installation prefix (default: /usr/local) → bin at PREFIX/bin"
	@echo "  BINDIR              - Binary directory (default: PREFIX/bin)"
	@echo "  DESTDIR             - Prepended path for staged installs (e.g. packaging)"
	@echo "  VERBOSE / V         - Set to 1 for maximum detail (go build -v -x)"
	@echo "  QUIET               - Set to 1 to silence go build -v (default shows package list)"

all: build

build: check-linux
	@echo "Building Reddock (packages listed below)..."
	$(GO) build $(GO_VERBOSE) $(LDFLAGS) -o $(BINARY) .
	@echo "Build complete: ./$(BINARY)"

static: check-linux
	@echo "Building static binary..."
	CGO_ENABLED=0 $(GO) build $(GO_VERBOSE) $(LDFLAGS) -o $(BINARY) .
	@echo "Static build complete: ./$(BINARY)"

dist: check-linux static
	@$(MAKE) dist-pack

dist-pack:
	@test -f $(BINARY) || (echo "Missing ./$(BINARY); run make static first." && exit 1)
	@mkdir -p dist
	tar -cJf dist/$(BINARY)-$(VERSION)-linux-amd64.tar.xz $(BINARY) README.md LICENSE
	@echo "Tarball: dist/$(BINARY)-$(VERSION)-linux-amd64.tar.xz"

run: build
	./$(BINARY) $(ARGS)

install: check-linux build
	@echo "Installing $(BINARY) → $(INSTALLED)"
	install -Dm755 $(BINARY) $(INSTALLED)
	@test -x "$(INSTALLED)" || { echo "Error: install did not produce an executable at $(INSTALLED)"; exit 1; }
	@ls -l "$(INSTALLED)"
	@echo "Installed and verified at $(INSTALLED)"

uninstall:
	@if [ ! -e "$(INSTALLED)" ]; then \
		echo "Nothing to remove: $(INSTALLED) is not present"; \
		exit 0; \
	fi
	@echo "Removing $(INSTALLED)"
	rm -f "$(INSTALLED)"
	@if [ -e "$(INSTALLED)" ]; then \
		echo "Error: could not remove $(INSTALLED)"; \
		exit 1; \
	fi
	@echo "Uninstall complete ($(INSTALLED) removed)"

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
	rm -rf dist/ .tools/
	rm -f coverage.out coverage.html
	@echo "Done"
