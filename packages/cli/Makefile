.PHONY: clean format lint test build install

BIN_DIR  := bin
BIN_NAME := postmanify
SRC      := ./main.go
OWNER    := github.com/hieudoanm/postmanify

VERSION      := $(shell date +%Y.%m.%d.%H.%M.%S)
COMMIT       := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
FULL_VERSION := $(VERSION)-$(COMMIT)
LDFLAGS      := -ldflags="-s -w \
	-X '$(OWNER)/src/cmd.Version=$(FULL_VERSION)'"

## clean: Remove build artifacts
clean:
	@rm -rf $(BIN_DIR)
	@echo "🧹 Cleaned $(BIN_DIR)/"

## format: Run go fmt
format:
	go fmt ./...

## lint: Run go vet
lint:
	go vet ./...

## test: Run all tests
test:
	go test ./...

## build: Format, lint, then compile binary into ./bin
build: clean format lint test
	@mkdir -p $(BIN_DIR)
	go build $(LDFLAGS) -o $(BIN_DIR)/$(BIN_NAME) $(SRC)
	@echo "✅ Binary:  $$(pwd)/$(BIN_DIR)/$(BIN_NAME)"
	@echo "📦 Version: $(FULL_VERSION)"
	@echo "📏 Size:    $$(du -h $(BIN_DIR)/$(BIN_NAME) | cut -f1)"

## install: Build then install binary into ~/bin
install: build
	@mkdir -p $(HOME)/bin
	cp $(BIN_DIR)/$(BIN_NAME) $(HOME)/bin/
	chmod +x $(HOME)/bin/$(BIN_NAME)
	@echo "✅ Installed to $$HOME/bin/$(BIN_NAME)"
	@echo "   Make sure $$HOME/bin is in your PATH:"
	@echo '   export PATH="$$HOME/bin:$$PATH"'

## help: List available targets
help:
	@grep -E '^## ' Makefile | sed 's/^## //'
