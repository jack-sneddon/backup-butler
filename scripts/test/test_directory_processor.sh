#!/bin/bash
source scripts/test/common.sh

setup_processor_test() {
    rm -rf "$TEST_ROOT"
    
    # Create test directories and test files
    mkdir -p "$TEST_DIR/source" \
             "$TEST_DIR/source/nested" \
             "$TEST_DIR/target"

    # Create test files with different sizes
    dd if=/dev/urandom of="$TEST_DIR/source/file1.txt" bs=1024 count=1 2>/dev/null
    dd if=/dev/urandom of="$TEST_DIR/source/file2.txt" bs=1024 count=10 2>/dev/null
    dd if=/dev/urandom of="$TEST_DIR/source/nested/file3.txt" bs=1024 count=5 2>/dev/null

    # Set specific permissions and timestamps
    chmod 644 "$TEST_DIR/source/file1.txt"
    chmod 755 "$TEST_DIR/source/file2.txt"
    touch -t 202001010000 "$TEST_DIR/source/file1.txt"
    touch -t 202001010000 "$TEST_DIR/source/file2.txt"

    # Create test config
    cat > "$TEST_DIR/config.yaml" << EOL
source: "${TEST_DIR}/source"
target: "${TEST_DIR}/target"
comparison:
  algorithm: "sha256"
  level: "standard"
  buffer_size: 32768
storage:
  device_type: "hdd"
  max_threads: 4
EOL
}

test_source_missing() {
    printf "  Testing missing source directory... "
    
    # Create config with non-existent source
    cat > "$TEST_DIR/bad_source.yaml" << EOL
source: "${TEST_DIR}/not_exists"
target: "${TEST_DIR}/target"
comparison:
  algorithm: "sha256"
  level: "standard"
storage:
  device_type: "hdd"
  max_threads: 4
EOL

    ./bin/backup-butler sync -c "$TEST_DIR/bad_source.yaml" 2>/dev/null
    if [ $? -eq 1 ]; then
        printf "${GREEN}PASS${NC}\n"
        return 0
    else
        printf "${RED}FAIL${NC}\n"
        return 1
    fi
}

test_basic_copy() {
    printf "  Testing basic file copy... "
    
    if ! ./bin/backup-butler sync -c "$TEST_DIR/config.yaml" > /dev/null 2>&1; then
        printf "${RED}FAIL${NC} (command failed)\n"
        return 1
    fi
    
    # Check files exist
    for file in "file1.txt" "file2.txt"; do
        if [ ! -f "$TEST_DIR/target/$file" ]; then
            printf "${RED}FAIL${NC} (missing $file)\n"
            return 1
        fi
    done

    # Check file contents match
    for file in "file1.txt" "file2.txt"; do
        if ! cmp -s "$TEST_DIR/source/$file" "$TEST_DIR/target/$file"; then
            printf "${RED}FAIL${NC} (content mismatch in $file)\n"
            return 1
        fi
    done

    printf "${GREEN}PASS${NC}\n"
    return 0
}

test_metadata() {
    printf "  Testing metadata preservation... "
    local failed=0

    # Test permissions
    src_perm=$(stat -f%p "$TEST_DIR/source/file1.txt")
    tgt_perm=$(stat -f%p "$TEST_DIR/target/file1.txt")
    if [ "$src_perm" != "$tgt_perm" ]; then
        printf "${RED}FAIL${NC} (permission mismatch)\n"
        return 1
    fi

    # Test timestamps
    src_time=$(stat -f%m "$TEST_DIR/source/file1.txt")
    tgt_time=$(stat -f%m "$TEST_DIR/target/file1.txt")
    if [ "$src_time" != "$tgt_time" ]; then
        printf "${RED}FAIL${NC} (timestamp mismatch)\n"
        return 1
    fi

    printf "${GREEN}PASS${NC}\n"
    return 0
}

main() {
    local failed=0
    printf "Running directory processor tests...\n"
    
    setup_processor_test
    
    test_source_missing || failed=1
    test_basic_copy || failed=1
    test_metadata || failed=1
    
    rm -rf "$TEST_ROOT"
    return $failed
}

main