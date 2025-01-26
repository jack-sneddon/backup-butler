# scripts/test/run_tests.sh
#!/bin/bash
source scripts/test/common.sh

printf "Running Backup Butler Tests\n"
printf "==========================\n\n"

# Run each test script
for test in scripts/test/test_*.sh; do
   if [[ "$test" != *"run_tests.sh"* ]]; then
       printf "Running ${test#scripts/test/}...\n"
       if ! bash "$test" $([[ "$VERBOSE" == "true" ]] && echo "-v"); then
           printf "${RED}Test failed${NC}\n"
           exit 1
       fi
       printf "\n"
   fi
done

printf "${GREEN}All tests passed${NC}\n"