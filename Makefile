.PHONY: build run clean tidy build-all build-darwin build-linux build-windows help

BINARY=smallcode
VERSION?=0.1.0
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

GOOS=$(shell go env GOOS)
GOARCH=$(shell go env GOARCH)

PLATFORMS := darwin/amd64 darwin/arm64 linux/amd64 linux/arm64
PLATFORMS_WINDOWS := windows/amd64 windows/arm64

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

test: ## Run tests
	CGO_ENABLED=1 go test -v ./...

test-harness: ## Run e2e security harness tests
	CGO_ENABLED=1 go test -v ./api/...

build-all: ## Build for all platforms
	@mkdir -p dist
	@for platform in $(PLATFORMS); do \
		os=$${platform%/*}; \
		arch=$${platform#*/}; \
		output="dist/$(BINARY)-$$os-$$arch"; \
		[ "$$os" = "windows" ] && output="$$output.exe"; \
		echo "Building for $$os/$$arch..."; \
		if [ "$$os" = "linux" ]; then \
			CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch go build $(LDFLAGS) -o $$output .; \
		else \
			CGO_ENABLED=1 GOOS=$$os GOARCH=$$arch go build $(LDFLAGS) -o $$output .; \
		fi \
	done

build-darwin: ## Build for macOS (all architectures)
	@mkdir -p dist
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-darwin-amd64 .
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY)-darwin-arm64 .

build-linux: build-linux-amd64 build-linux-arm64 ## Build for Linux (all architectures)

build-linux-amd64:
	@mkdir -p dist
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-linux-amd64 .

build-linux-arm64: ## Build for Linux ARM64
	@mkdir -p dist
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY)-linux-arm64 .

deploy: build-linux-arm64 ## Deploy ARM64 binary to tart debian instance
	scp dist/$(BINARY)-linux-arm64 admin@$$(tart ip debian):/home/admin/bin/$(BINARY)

build-windows: ## Build for Windows (all architectures)
	@mkdir -p dist
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-windows-amd64.exe .
	CGO_ENABLED=1 GOOS=windows GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY)-windows-arm64.exe .
