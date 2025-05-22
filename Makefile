.PHONY: build clean test run

BINARY_NAME=backup-butler
BUILD_DIR=bin


# Default target
all: fmt tidy vet build

# Show help
help:
	@echo "Available targets:"
	@echo "  make all             - Format, tidy, vet and build"
	@echo "  make build           - Build the binary"
	@echo "  make test            - Run unit tests"
	@echo "  make clean           - Clean build artifacts"
	@echo "  make fmt             - Format code"
	@echo "  make check           - Run all checks (fmt, tidy, vet)"


fmt:
	go fmt ./...

tidy:
	go mod tidy

vet:
	go vet ./...

build:
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/backup-butler

build-local:
	@echo "Building $(BINARY_NAME) to project root..."
	@go build -o $(BINARY_NAME) ./cmd/backup-butler

clean:
	@echo "Cleaning up..."
	@rm -rf $(BUILD_DIR)/$(BINARY_NAME)
	@rm -f $(BINARY_NAME)

test:
	@echo "Running tests..."
	@go test -v ./...

run: build-local
	@echo "Running $(BINARY_NAME)..."
	@./$(BINARY_NAME)

install: build
	@echo "Installing $(BINARY_NAME)..."
	@cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/