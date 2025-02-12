#!/bin/bash
source scripts/test/common.sh

setup_test_env() {
    rm -rf "$TEST_ROOT"
    mkdir -p "$TEST_DIR/source" "$TEST_DIR/target"

    # Create test config
    cat > "${TEST_DIR}/test_config.yaml" << EOL
source: "${TEST_DIR}/source"
target: "${TEST_DIR}/target"
comparison:
  algorithm: "sha256"
  level: "standard"
storage:
  device_type: "hdd"
  max_threads: 4
EOL
}

run_log_test() {
    local level=$1

    printf "  Testing %-30s" "${level:-default} level..."

    cmd="./bin/backup-butler sync -c ${TEST_DIR}/test_config.yaml"
    [[ $level != "default" ]] && cmd="$cmd --log-level $level"

    # Capture stderr separately since that's where our logs go
    output=$($cmd 2>&1)
    success=false

    case "$level" in
        "default"|"error")
            # Should see no DEBUG or INFO messages
            if [[ ! $output =~ "DEBUG" && ! $output =~ "INFO" ]]; then
                success=true
            fi
            ;;
        "debug")
            # Should see DEBUG messages
            if [[ $output =~ "DEBUG" ]]; then
                success=true
            fi
            ;;
        "info")
            # Should see INFO but no DEBUG messages
            if [[ $output =~ "INFO" && ! $output =~ "DEBUG" ]]; then
                success=true
            fi
            ;;
        "warn")
            # Should see only WARN or higher messages
            if [[ ! $output =~ "INFO" && ! $output =~ "DEBUG" ]]; then
                success=true
            fi
            ;;
        "error")
            # Should see only ERROR messages
            if [[ ! $output =~ "WARN" && ! $output =~ "INFO" && ! $output =~ "DEBUG" ]]; then
                success=true
            fi
            ;;
    esac

    if $success; then
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
    printf "Running logging tests...\n"

    setup_test_env

    run_log_test "default" || failed=1
    run_log_test "debug" || failed=1
    run_log_test "info" || failed=1
    run_log_test "warn" || failed=1
    run_log_test "error" || failed=1


    # cleanup
    rm -rf "$TEST_ROOT"

    return $failed
}

main