#!/bin/bash
source scripts/test/common.sh

test_directory_scan() {
    printf "  Testing directory scan... "
    output=$(./bin/backup-butler check -c "${TEST_DIR}/test_config.yaml" --log-level debug 2>&1)
    
    if [[ $output =~ "Directories: 4" && 
          $output =~ "Files: 4" && 
          $output =~ "Total Size: 6.0 MB" ]]; then
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

    setup_test_env
    test_directory_scan || failed=1
    rm -rf "$TEST_ROOT"

    return $failed
}

main