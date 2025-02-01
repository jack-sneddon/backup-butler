# test/README.md

# Backup Butler Testing Guide

## Test Structure

```sh

test/
├── acceptance/     # Black-box testing of CLI commands
│   ├── testdata/  # Test config files and fixtures
│   └── *_test.go  # Test files
├── helpers/       # Shared test utilities
└── integration/   # Integration tests (if needed)
```

## Running Tests

### Using Make Targets

```bash
# Run all acceptance tests
make acceptance-test

# Run specific test suite
make acceptance-test-Config    # Runs TestConfig
make acceptance-test-Check     # Runs TestCheck

# Run unit tests
make test

# Run all tests and build
make all test acceptance-test
```

### Using Go Test Directly

```bash
# Run all acceptance tests


# Run specific test file


# Run specific test


# Run with verbose output


# Run specific test case

```

## Writing Tests

### Test Naming Conventions

- Test files: `TBD`
- Test functions: `TestXxx` where Xxx is the feature being tested
- Test cases: Use descriptive names in the test table

### Using Test Helpers
