.PHONY: all build test test-coverage test-integration lint lint-arch clean install build-all

BINARY_NAME := siyuan-cli
BUILD_DIR := ./dist
BINARY := $(BUILD_DIR)/$(BINARY_NAME)
MAIN_PACKAGE := ./cmd/siyuan
# Install prefix; the binary is copied to $(PREFIX)/bin
PREFIX ?= /usr/local

all: build

build:
	go build -o $(BINARY) $(MAIN_PACKAGE)

test:
	SIYUAN_INTEGRATION_TEST=0 go test ./... -v -race

test-coverage:
	SIYUAN_INTEGRATION_TEST=0 go test ./... -race -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

test-integration:
	SIYUAN_INTEGRATION_TEST=1 go test ./... -v -race

lint:
	golangci-lint run ./...

lint-arch:
	go run scripts/lint-deps.go
	go run scripts/lint-quality.go

clean:
	rm -rf $(BUILD_DIR) coverage.out coverage.html

install: build
	cp $(BINARY) $(PREFIX)/bin/$(BINARY_NAME)

# Cross-compilation builds
build-all:
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PACKAGE)
	GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PACKAGE)
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PACKAGE)
	GOOS=linux GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PACKAGE)
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PACKAGE)
