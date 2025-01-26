# scripts/test/test_scan.sh (Scanner unit tests)
#!/bin/bash
source scripts/test/common.sh

setup_test_dir() {
    rm -rf "$TEST_DIR"
    mkdir -p "${TEST_DIR}/dir1/subdir1" "${TEST_DIR}/dir2"
    dd if=/dev/zero of="${TEST_DIR}/dir1/file1" bs=1M count=1 2>/dev/null
    dd if=/dev/zero of="${TEST_DIR}/dir1/subdir1/file2" bs=1M count=2 2>/dev/null
    dd if=/dev/zero of="${TEST_DIR}/dir2/file3" bs=1M count=3 2>/dev/null
}

test_directory_scan() {
    printf "Testing directory scan... "
    output=$(./bin/backup-butler check -c "${TEST_DIR}/test_config.yaml" --log-level debug 2>&1)
    
    if [[ $output =~ "Directories: 4" && 
          $output =~ "Files: 4" && 
          $output =~ "Total Size: 6.0 MB" ]]; then
        printf "${GREEN}PASS${NC}\n"
        $VERBOSE && echo "$output"
    else
        printf "${RED}FAIL${NC}\n"
        echo "$output"
    fi
}

setup_test_dir
test_directory_scan