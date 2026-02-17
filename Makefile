BINARY_NAME=portless
VERSION?=1.0.0
BUILD_DIR=bin
INSTALL_DIR=/usr/local/bin

.PHONY: all build clean install uninstall test

all: build

build:
	@echo "Building portless..."
	@mkdir -p $(BUILD_DIR)
	go build -ldflags="-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/portless
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

build-all:
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -ldflags="-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/portless
	GOOS=linux GOARCH=arm64 go build -ldflags="-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/portless
	GOOS=darwin GOARCH=amd64 go build -ldflags="-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/portless
	GOOS=darwin GOARCH=arm64 go build -ldflags="-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/portless
	@echo "Multi-platform build complete"

install: build
	@echo "Installing portless to $(INSTALL_DIR)..."
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/
	@sudo chmod +x $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "Installation complete. Run 'portless help' to get started."

uninstall:
	@echo "Uninstalling portless..."
	@sudo rm -f $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "Uninstall complete"

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@echo "Clean complete"

test:
	@echo "Running tests..."
	go test -v ./...

.DEFAULT_GOAL := build
