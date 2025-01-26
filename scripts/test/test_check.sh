# scripts/test/test_check.sh (Check command tests)
#!/bin/bash
source scripts/test/common.sh

setup_comparison() {
    mkdir -p "${TEST_DIR}-target/dir1/subdir1" "${TEST_DIR}-target/dir2" 2>/dev/null
    cp "${TEST_DIR}/dir1/file1" "${TEST_DIR}-target/dir1/" 2>/dev/null
    dd if=/dev/zero of="${TEST_DIR}-target/dir1/subdir1/file2" bs=1M count=1 2>/dev/null
    touch "${TEST_DIR}-target/dir2/extra.txt" 2>/dev/null
}

test_comparison() {
    printf "Testing file comparison... "
    output=$(./bin/backup-butler check -c "${TEST_DIR}/test_config.yaml" --log-level debug 2>&1)
    
    patterns=("= dir1/file1" "\* dir1/subdir1/file2" "\+ dir2/file3" "- dir2/extra.txt")
    for pattern in "${patterns[@]}"; do
        if [[ ! $output =~ $pattern ]]; then
            printf "${RED}FAIL${NC}\n"
            echo "$output"
            return 1
        fi
    done
    
    printf "${GREEN}PASS${NC}\n"
    $VERBOSE && echo "$output"
}

setup_test_dir
setup_comparison
test_comparison