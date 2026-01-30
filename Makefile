.PHONY: help all build static install uninstall clean test fmt vet lint coverage dist check-linux

# Variables
PREFIX ?= /usr/local
BINDIR = $(PREFIX)/bin
BINARY = redway
GO = go
GOFLAGS = -ldflags="-s -w"

# Build flags
LDFLAGS = -ldflags="-s -w"

# Colors for output
RED = \033[0;31m
GREEN = \033[0;32m
YELLOW = \033[0;33m
NC = \033[0m # No Color

# Linux check helper
check-linux:
	@if [ "$(uname -s)" != "Linux" ]; then \
		echo "$(RED)Error: Redway is only available for Linux systems$(NC)"; \
		echo "$(YELLOW)Detected OS: $(uname -s)$(NC)"; \
		exit 1; \
	fi

help:
	@echo "$(GREEN)Redway Makefile Targets$(NC)"
	@echo ""
	@echo "$(YELLOW)Build Targets:$(NC)"
	@echo "  make build          - Build the binary (default)"
	@echo "  make static         - Build a static binary (no CGO)"
	@echo "  make dist           - Build binaries for multiple platforms"
	@echo ""
	@echo "$(YELLOW)Installation:$(NC)"
	@echo "  make install        - Build and install to $(PREFIX)/bin"
	@echo "  make uninstall      - Remove installed binary"
	@echo ""
	@echo "$(YELLOW)Development:$(NC)"
	@echo "  make fmt            - Format code with gofmt"
	@echo "  make vet            - Run go vet for static analysis"
	@echo "  make lint           - Run golangci-lint (if installed)"
	@echo "  make test           - Run tests with verbose output"
	@echo "  make coverage       - Run tests with coverage report"
	@echo ""
	@echo "$(YELLOW)Maintenance:$(NC)"
	@echo "  make clean          - Remove built binaries"
	@echo "  make help           - Show this help message"
	@echo ""
	@echo "$(YELLOW)Variables:$(NC)"
	@echo "  PREFIX              - Installation prefix (default: /usr/local)"
	@echo "  DESTDIR             - Staging directory for install"
	@echo ""

all: build

build: check-linux
	@echo "$(GREEN)Building Redway...$(NC)"
	$(GO) build $(LDFLAGS) -o $(BINARY) .
	@echo "$(GREEN)Build complete: $(BINARY)$(NC)"

static: check-linux
	@echo "$(GREEN)Building static binary...$(NC)"
	CGO_ENABLED=0 $(GO) build $(LDFLAGS) -o $(BINARY) .
	@echo "$(GREEN)Static build complete: $(BINARY)$(NC)"

dist:
	@echo "$(GREEN)Building distribution binaries...$(NC)"
	@mkdir -p dist
	@for os in linux darwin windows; do \
		for arch in amd64 arm64; do \
			echo "  Building $os/$arch..."; \
			GOOS=$os GOARCH=$arch $(GO) build $(LDFLAGS) -o dist/$(BINARY)-$os-$arch . || true; \
		done; \
	done
	@echo "$(GREEN)Distribution builds complete in dist/$(NC)"

install: check-linux build
	@echo "$(GREEN)Installing Redway...$(NC)"
	install -Dm755 $(BINARY) $(DESTDIR)$(BINDIR)/$(BINARY)
	@echo ""
	@echo "$(GREEN)Installation complete!$(NC)"
	@echo ""
	@echo "$(YELLOW)Usage:$(NC)"
	@echo "  sudo redway init                                    				# Initialize with default image"
	@echo "  sudo redway init docker://redroid/redroid:16.0.0_64only-latest		# Custom OCI image"
	@echo "  sudo redway start                                   				# Start container"
	@echo "  sudo redway adb-connect                             				# Get ADB info"
	@echo ""

uninstall:
	@echo "$(YELLOW)Uninstalling Redway...$(NC)"
	rm -f $(DESTDIR)$(BINDIR)/$(BINARY)
	@echo "$(GREEN)Uninstall complete$(NC)"
	@echo ""
	@echo "$(YELLOW)Note:$(NC) Config and data preserved. Remove manually if needed:"
	@echo "  rm -rf ~/.config/redway"
	@echo "  rm -rf ~/data-redroid"
	@echo ""

fmt:
	@echo "$(GREEN)Formatting code...$(NC)"
	$(GO) fmt ./...
	@echo "$(GREEN)Code formatted$(NC)"

vet:
	@echo "$(GREEN)Running go vet...$(NC)"
	$(GO) vet ./...
	@echo "$(GREEN)Vet check passed$(NC)"

lint:
	@echo "$(GREEN)Running golangci-lint...$(NC)"
	@which golangci-lint > /dev/null || (echo "$(RED)golangci-lint not found. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest$(NC)" && exit 1)
	golangci-lint run ./...
	@echo "$(GREEN)Lint check passed$(NC)"

test:
	@echo "$(GREEN)Running tests...$(NC)"
	$(GO) test -v ./...
	@echo "$(GREEN)Tests passed$(NC)"

coverage:
	@echo "$(GREEN)Running tests with coverage...$(NC)"
	$(GO) test -v -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report generated: coverage.html$(NC)"

clean:
	@echo "$(YELLOW)Cleaning...$(NC)"
	rm -f $(BINARY)
	rm -rf dist/
	rm -f coverage.out coverage.html
	@echo "$(GREEN)Clean complete$(NC)"
