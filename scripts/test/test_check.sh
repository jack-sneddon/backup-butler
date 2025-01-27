#!/bin/bash
source scripts/test/common.sh

test_file_comparison() {
    printf "  Testing file comparison... "

    output=$(./bin/backup-butler check -c "${TEST_DIR}/test_config.yaml" 2>&1)

    if [[ $output =~ "File Status" ]]; then
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
    
    setup_test_env
    test_file_comparison || failed=1
    rm -rf "$TEST_ROOT"
    
    return $failed
}

main