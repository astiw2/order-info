# Variables
BINARY_NAME=order-info
MAIN_PATH=./cmd/batch-processor/main.go
BUILD_DIR=./bin

.PHONY: all
all: clean build

.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

.PHONY: run
run:
	@echo "Running $(BINARY_NAME)..."
	go run $(MAIN_PATH)

.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@echo "Clean complete"

.PHONY: test
test:
	@echo "Running tests..."
	go test ./cmd/...

.PHONY: dev
dev: clean build run
