BINARY_NAME=portless
VERSION?=1.0.0
BUILD_DIR=bin
DIST_DIR=dist
INSTALL_DIR=/usr/local/bin

.PHONY: all build clean install uninstall test release release-checksums release-clean help

all: build

help:
	@echo "Available targets:"
	@echo "  make build              - Build portless for current platform"
	@echo "  make build-all          - Build portless for all platforms (darwin/linux, amd64/arm64)"
	@echo "  make release            - Create Homebrew release archives (.tar.gz) for all platforms"
	@echo "  make release-checksums  - Create release archives and generate SHA256 checksums"
	@echo "  make install            - Install portless to $(INSTALL_DIR)"
	@echo "  make uninstall          - Remove portless from $(INSTALL_DIR)"
	@echo "  make test               - Run all tests"
	@echo "  make clean              - Remove build artifacts ($(BUILD_DIR)/)"
	@echo "  make release-clean      - Remove release artifacts ($(DIST_DIR)/)"
	@echo ""
	@echo "Variables:"
	@echo "  VERSION=$(VERSION)      - Version to embed in binary"
	@echo "  BUILD_DIR=$(BUILD_DIR)  - Directory for build output"
	@echo "  DIST_DIR=$(DIST_DIR)    - Directory for release archives"

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

release: build-all
	@echo "Creating Homebrew release archives..."
	@mkdir -p $(DIST_DIR)
	@cd $(BUILD_DIR) && for file in $(BINARY_NAME)-*; do \
		echo "Creating archive for $$file..."; \
		tar czf ../$(DIST_DIR)/$${file}.tar.gz $$file; \
	done
	@echo "Release archives created in $(DIST_DIR)/"
	@ls -lh $(DIST_DIR)/

release-checksums: release
	@echo "Generating SHA256 checksums..."
	@cd $(DIST_DIR) && \
		if command -v sha256sum >/dev/null 2>&1; then \
			sha256sum *.tar.gz > checksums.txt; \
		else \
			shasum -a 256 *.tar.gz > checksums.txt; \
		fi
	@echo "Checksums generated:"
	@cat $(DIST_DIR)/checksums.txt

release-clean:
	@echo "Cleaning release artifacts..."
	@rm -rf $(DIST_DIR)
	@echo "Release artifacts cleaned"

test:
	@echo "Running tests..."
	go test -v ./...

.DEFAULT_GOAL := build
