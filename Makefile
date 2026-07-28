# Binary names
BINS := bleeder bleeder-wav bleeder-irp bleeder-midi

# Build output directory
BIN_DIR := bin

# Install location
INSTALL_PATH := $(HOME)/bin

.PHONY: all build clean install help test

# Default target
all: build

# Build all binaries using a loop
build:
	@echo "Building binaries..."
	@mkdir -p $(BIN_DIR)
	@for bin in $(BINS); do \
		echo "  Building $$bin..."; \
		go build -o $(BIN_DIR)/$$bin ./cmd/$$bin || exit 1; \
	done
	@echo "Done! Binaries in $(BIN_DIR)/"

# Clean all binaries
clean:
	@echo "Cleaning..."
	@rm -rf $(BIN_DIR)

# Install binaries
install: build
	@echo "Installing to $(INSTALL_PATH)..."
	@mkdir -p $(INSTALL_PATH)
	@cp $(BIN_DIR)/* $(INSTALL_PATH)/
	@echo "Installed!"

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Show help
help:
	@echo "Available targets:"
	@echo "  build    - Build all binaries to $(BIN_DIR)/"
	@echo "  clean    - Remove $(BIN_DIR)/ directory"
	@echo "  install  - Install binaries to $(INSTALL_PATH)"
	@echo "  test     - Run all tests"
	@echo ""
	@echo "Binaries: $(BINS)"
