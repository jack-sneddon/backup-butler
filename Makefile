.PHONY: all fmt tidy vet build test acceptance-test clean help test-deps

BINARY=bin/backup-butler
TEST_BINARY=${BINARY}.test
COVER_PROFILE=coverage.out


# Default target
all: fmt tidy vet build

# Show help
help:
	@echo "Available targets:"
	@echo "  make all             - Format, tidy, vet and build"
	@echo "  make build           - Build the binary"
	@echo "  make test            - Run unit tests"
	@echo "  make acceptance-test - Run acceptance tests"
	@echo "  make test-all        - Run all tests (unit + acceptance)"
	@echo "  make cover           - Run tests with coverage"
	@echo "  make test-deps       - Verify test dependencies"
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
	go build -o ${BINARY} ./cmd/backup-butler

# Test targets
test-deps:
	@echo "Checking test dependencies..."
	@if [ ! -d "test/acceptance/testdata" ]; then \
		mkdir -p test/acceptance/testdata; \
	fi

test: test-deps
	go test ./internal/...

cover:
	go test -coverprofile=${COVER_PROFILE} ./...
	go tool cover -html=${COVER_PROFILE}

# Acceptance test targets
acceptance-test: build test-deps
	@echo "Running acceptance tests..."
	@mkdir -p /tmp/backup-butler-tests/source
	@mkdir -p /tmp/backup-butler-tests/target
	PATH="${PWD}/bin:${PATH}" go test -v ./test/acceptance/...

acceptance-test-%: build test-deps
	@echo "Running acceptance test $*..."
	@mkdir -p /tmp/backup-butler-tests/source
	@mkdir -p /tmp/backup-butler-tests/target
	PATH="${PWD}/bin:${PATH}" go test -v ./test/acceptance -run $*

test-all: test acceptance-test

check: fmt tidy vet

clean:
	rm -f ${BINARY}
	rm -f ${TEST_BINARY}
	rm -f ${COVER_PROFILE}
	rm -rf /tmp/backup-butler-tests