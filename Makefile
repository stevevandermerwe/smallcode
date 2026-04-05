.PHONY: build run clean tidy build-all build-darwin build-linux build-windows help

BINARY=smallcode
VERSION?=0.1.0
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

GOOS=$(shell go env GOOS)
GOARCH=$(shell go env GOARCH)

PLATFORMS := darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64 windows/arm64

help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build for current platform
	@mkdir -p dist
	go build $(LDFLAGS) -o dist/$(BINARY) .

run: build ## Build and run
	./dist/$(BINARY)

clean: ## Remove dist directory
	rm -rf dist

tidy: ## Tidy go modules
	go mod tidy

build-all: ## Build for all platforms
	@mkdir -p dist
	@for platform in $(PLATFORMS); do \
		os=$${platform%/*}; \
		arch=$${platform#*/}; \
		output="dist/$(BINARY)-$$os-$$arch"; \
		[ "$$os" = "windows" ] && output="$$output.exe"; \
		echo "Building for $$os/$$arch..."; \
		GOOS=$$os GOARCH=$$arch go build $(LDFLAGS) -o $$output .; \
	done

build-darwin: ## Build for macOS (all architectures)
	@mkdir -p dist
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY)-darwin-arm64 .

build-linux: ## Build for Linux (all architectures)
	@mkdir -p dist
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY)-linux-arm64 .

build-windows: ## Build for Windows (all architectures)
	@mkdir -p dist
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-windows-amd64.exe .
	GOOS=windows GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY)-windows-arm64.exe .
