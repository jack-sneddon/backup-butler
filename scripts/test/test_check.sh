#!/bin/bash
source scripts/test/common.sh

setup_check_test() {
    rm -rf "$TEST_ROOT"

    # Create test directories with source and target
    mkdir -p "$TEST_DIR/dir1/subdir1" \
            "$TEST_DIR/dir2" \
            "$TEST_DIR-target/dir1/subdir1" \
            "$TEST_DIR-target/dir2"

    # Create source files
    dd if=/dev/zero of="$TEST_DIR/dir1/file1.txt" bs=1M count=1 2>/dev/null
    dd if=/dev/zero of="$TEST_DIR/dir1/subdir1/file2.txt" bs=1M count=2 2>/dev/null
    dd if=/dev/zero of="$TEST_DIR/dir2/file3.txt" bs=1M count=3 2>/dev/null

    # Create target files (some matching, some different)
    cp "$TEST_DIR/dir1/file1.txt" "$TEST_DIR-target/dir1/"
    dd if=/dev/zero of="$TEST_DIR-target/dir1/subdir1/file2.txt" bs=1M count=2 2>/dev/null
    dd if=/dev/zero of="$TEST_DIR-target/dir2/file3.txt" bs=1M count=4 2>/dev/null # Different size

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
}

test_validation_levels() {
    printf "  Testing validation levels...\n"

    for level in quick standard deep; do
        printf "    Testing %-10s level... " "$level"
        output=$(./bin/backup-butler check -c "${TEST_DIR}/test_config.yaml" --level "$level" 2>&1)

        if [[ $output =~ "Scan Results" && $output =~ "Results Summary" ]]; then
            printf "${GREEN}PASS${NC}\n"
            $VERBOSE && echo "$output"
        else
            printf "${RED}FAIL${NC}\n"
            $VERBOSE || echo "$output"
            return 1
        fi
    done
    return 0
}

test_comparison_results() {
    printf "  Testing comparison results... "

    output=$(./bin/backup-butler check -c "${TEST_DIR}/test_config.yaml" 2>&1)

    if [[ $output =~ "Matched:" &&
          $output =~ "Modified:" &&
          $output =~ "Results Summary" ]]; then
        printf "${GREEN}PASS${NC}\n"
        $VERBOSE && echo "$output"
        return 0
    else
        printf "${RED}FAIL${NC}\n"
        $VERBOSE || echo "$output"
        return 1
    fi
}

test_summary_output() {
    printf "  Testing summary output... "

    output=$(./bin/backup-butler check -c "${TEST_DIR}/test_config.yaml" 2>&1)

    if [[ $output =~ "Source:" &&
          $output =~ "Target:" &&
          $output =~ "Directories:" ]]; then
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
    printf "Running check command tests...\n"
    
    setup_check_test

    test_validation_levels || failed=1
    test_comparison_results || failed=1
    test_summary_output || failed=1

    rm -rf "$TEST_ROOT"
    
    return $failed
}

main