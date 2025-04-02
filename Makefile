.PHONY: build clean run test

# Build variables
BINARY_NAME=zfs-watcher
BUILD_DIR=build

# Default target
all: build

# Build the ZFS watcher
build:
	@echo "Building ZFS watcher..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/zfs-watcher
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Run the ZFS watcher
run: build
	@echo "Running ZFS watcher..."
	@$(BUILD_DIR)/$(BINARY_NAME) $(ARGS)

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@echo "Clean complete"

# Show help
help:
	@echo "ZFS Tools Makefile"
	@echo ""
	@echo "Targets:"
	@echo "  build       Build the ZFS watcher binary"
	@echo "  run         Build and run the ZFS watcher (use ARGS=\"--pools pool1,pool2\" to pass arguments)"
	@echo "  test        Run all tests"
	@echo "  clean       Remove build artifacts"
	@echo "  help        Show this help message"
	@echo ""
	@echo "Examples:"
	@echo "  make run ARGS=\"--pools pool1,pool2 --interval 10 --output events.log\"" 