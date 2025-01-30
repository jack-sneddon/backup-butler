#!/bin/bash
source scripts/test/common.sh

setup_scan_test() {
    rm -rf "$TEST_ROOT"

    # Create test directories
    mkdir -p "$TEST_DIR/dir1/subdir1" \
            "$TEST_DIR/dir2" \
            "$TEST_DIR/excluded"

    # Create test files
    dd if=/dev/zero of="$TEST_DIR/dir1/file1.txt" bs=1M count=1 2>/dev/null
    dd if=/dev/zero of="$TEST_DIR/dir1/subdir1/file2.txt" bs=1M count=2 2>/dev/null
    dd if=/dev/zero of="$TEST_DIR/dir2/file3.txt" bs=1M count=3 2>/dev/null
    dd if=/dev/zero of="$TEST_DIR/excluded/temp.tmp" bs=1M count=1 2>/dev/null

    # Create test config
    cat > "${TEST_DIR}/test_config.yaml" << EOL
source: "${TEST_DIR}"
target: "${TEST_DIR}-target"
folders:
  - "dir1"
  - "dir2"
exclude:
  - "*.tmp"
  - "*.yaml"
  - ".DS_Store"
comparison:
  algorithm: "sha256"
  level: "standard"
  buffer_size: 32768
storage:
  device_type: "hdd"
  max_threads: 4
logging:
  level: "debug"
EOL

    if $VERBOSE; then
        echo "Test setup:"
        echo "Directory structure:"
        find "${TEST_DIR}" -type f
        echo "Config contents:"
        cat "${TEST_DIR}/test_config.yaml"
    fi
}

test_directory_scan() {
    printf "  Testing directory scan... "
    output=$(./bin/backup-butler check -c "${TEST_DIR}/test_config.yaml" --log-level debug 2>&1)
    
    if $VERBOSE; then
        echo "Directory count test:"
        echo "Files found:"
        find "${TEST_DIR}" -type f -not -name "test_config.yaml"
        echo "Actual output:"
        echo "$output"
    fi

    if [[ $output =~ "├── Directories: 4" &&
          $output =~ "├── Files: 3" &&
          $output =~ "├── Total Size: 6.0 MB" ]]; then
        printf "${GREEN}PASS${NC}\n"
        $VERBOSE && echo "$output"
        return 0
    else
        printf "${RED}FAIL${NC}\n"
        echo "Expected: Directories: 4, Files: 3, Total Size: 6.0 MB"
        echo "Got: $(echo "$output" | grep -E 'Directories:|Files:|Total Size:')"
        return 1
    fi
}

test_exclusion_patterns() {
    printf "  Testing file exclusion... "
    output=$(./bin/backup-butler check -c "${TEST_DIR}/test_config.yaml" --log-level debug 2>&1)

    if $VERBOSE; then
        echo "Exclusion test:"
        echo "Expected excluded: temp.tmp"
        echo "Config contents:"
        cat "${TEST_DIR}/test_config.yaml"
        echo "Actual output:"
        echo "$output"
    fi

    if [[ $output =~ "├── Excluded Files: 1" &&
          ! $output =~ "/excluded/temp.tmp" ]]; then
        printf "${GREEN}PASS${NC}\n"
        $VERBOSE && echo "$output"
        return 0
    else
        printf "${RED}FAIL${NC}\n"
        echo "Expected: Excluded Files: 1"
        echo "Got: $(echo "$output" | grep 'Excluded Files:')"
        return 1
    fi
}

test_folder_filtering() {
    printf "  Testing folder filtering... "
    output=$(./bin/backup-butler check -c "${TEST_DIR}/test_config.yaml" 2>&1)

    # Only specified folders should be included
    if [[ $output =~ "dir1" &&
          $output =~ "dir2" &&
          ! $output =~ "/excluded/" ]]; then
        printf "${GREEN}PASS${NC}\n"
        $VERBOSE && echo "$output"
        return 0
    else
        printf "${RED}FAIL${NC}\n"
        $VERBOSE || echo "$output"
        return 1
    fi
}

test_progress_tracking() {
    printf "  Testing progress tracking... "
    output=$(./bin/backup-butler check -c "${TEST_DIR}/test_config.yaml" 2>&1)

    # Check for basic progress information
    if [[ $output =~ "Summary" &&
          $output =~ "Total Size:" ]]; then
        printf "${GREEN}PASS${NC}\n"
        $VERBOSE && echo "$output"
        return 0
    else
        printf "${RED}FAIL${NC}\n"
        $VERBOSE || echo "$output"
        return 1
    fi
}

main() {
    local failed=0
    printf "Running scanner tests...\n"

    setup_scan_test

    test_directory_scan || failed=1
    test_exclusion_patterns || failed=1
    test_folder_filtering || failed=1
    test_progress_tracking || failed=1

    rm -rf "$TEST_ROOT"

    return $failed
}

main