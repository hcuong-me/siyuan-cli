.PHONY: build test lint clean install

BINARY_NAME := siyuan-cli
BUILD_DIR := ./dist
MAIN_PACKAGE := ./cmd/siyuan

build:
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)

test:
	go test ./... -v -race

test-coverage:
	go test ./... -race -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

lint:
	golangci-lint run ./...

clean:
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

install:
	go install $(MAIN_PACKAGE)

# Cross-compilation builds
build-all:
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PACKAGE)
	GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PACKAGE)
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PACKAGE)
	GOOS=linux GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PACKAGE)
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PACKAGE)
