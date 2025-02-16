#!/bin/bash
source scripts/test/common.sh

# Process arguments first
process_args "$@"

printf "Running Backup Butler Tests (LOG_LEVEL=$LOG_LEVEL)\n"
printf "==========================\n\n"

# Run each test script with the same logging level
for test in scripts/test/test_*.sh; do
    if [[ "$test" != *"run_tests.sh"* ]]; then
        printf "Running ${test#scripts/test/}...\n"
        if ! bash "$test" --log-level="$LOG_LEVEL" $($VERBOSE && echo "--verbose"); then
            printf "${RED}Test failed${NC}\n"
            exit 1
        fi
        printf "\n"
    fi
done

printf "${GREEN}All tests passed${NC}\n"