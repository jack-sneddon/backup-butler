#!/bin/bash

# Configuration
declare -r TEST_ROOT="/tmp/backup-butler-tests"
declare -r TEST_DIR="$TEST_ROOT/test-files"
declare -r CONFIG_DIR="$TEST_ROOT/configs"

# Colors
declare -r GREEN='\033[0;32m'
declare -r RED='\033[0;31m'
declare -r YELLOW='\033[1;33m'
declare -r NC='\033[0m'

# Default settings
LOG_LEVEL=${LOG_LEVEL:-"error"}  # Application logging level
VERBOSE=${VERBOSE:-false}        # Test framework verbosity

# Process command line arguments
process_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -v|--verbose)
                VERBOSE=true
                shift
                ;;
            --log-level)
                LOG_LEVEL="$2"
                shift 2
                ;;
            *)
                echo "Unknown option: $1"
                exit 1
                ;;
        esac
    done

    export LOG_LEVEL
    export VERBOSE

    if $VERBOSE; then
        test_log "Test settings:"
        test_log "  LOG_LEVEL: $LOG_LEVEL"
        test_log "  VERBOSE: $VERBOSE"
        test_log "  TEST_ROOT: $TEST_ROOT"
    fi
}

# Run backup-butler with consistent logging
run_backup_butler() {
    local cmd="./bin/backup-butler $* --log-level $LOG_LEVEL"

    if $VERBOSE; then
        echo "Executing: $cmd"
    fi

    eval "$cmd"
}

# Test framework logging
test_log() {
    if $VERBOSE; then
        echo "→ $1"
    fi
}

# Test header
test_header() {
    local title="$1"
    echo "Running $title..."
}

# Test section
test_section() {
    local title="$1"
    printf "  Testing %s... " "$title"
}

# Assert contains pattern
assert_contains() {
    local output="$1"
    local pattern="$2"
    local description="$3"

    if [[ $output =~ $pattern ]]; then
        test_log "✓ Found: $description"
        return 0
    else
        test_log "✗ Missing: $description"
        test_log "Expected pattern: $pattern"
        return 1
    fi
}

# Assert file exists
assert_file_exists() {
    local file="$1"
    local description="$2"

    if [[ -f "$file" ]]; then
        test_log "✓ File exists: $description"
        return 0
    else
        test_log "✗ File missing: $description"
        return 1
    fi
}

# Report test result
report_result() {
    local success="$1"
    local output="$2"
    local description="${3:-}"

    if $success; then
        printf "${GREEN}PASS${NC}\n"
        $VERBOSE && echo "$output"
        return 0
    else
        printf "${RED}FAIL${NC}\n"
        if [ -n "$description" ]; then
            echo "$description"
        fi
        echo "$output"
        return 1
    fi
}

# Setup test environment
setup_test_env() {
    test_log "Setting up test environment"
    rm -rf "$TEST_ROOT"
    mkdir -p "$TEST_DIR" "$CONFIG_DIR"
    test_log "Created test directories"
}

# Cleanup test environment
cleanup_test_env() {
    test_log "Cleaning up test environment"
    rm -rf "$TEST_ROOT"
}