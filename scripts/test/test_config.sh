#!/bin/bash
source scripts/test/common.sh

setup_configs() {
    # Create test directories
    rm -rf "$TEST_ROOT"
    mkdir -p "$TEST_DIR/source" "$TEST_DIR/target" "$CONFIG_DIR"

    # Create test configs
    cat > "${CONFIG_DIR}/valid.yaml" << EOL
source: "${TEST_DIR}/source"
target: "${TEST_DIR}/target"
folders:
  - "test1"
comparison:
  algorithm: "sha256"
  level: "standard"
storage:
  device_type: "hdd"
  max_threads: 4
EOL


    # Valid config
    cat > "${CONFIG_DIR}/valid.yaml" << EOL
source: "${TEST_DIR}/source"
target: "${TEST_DIR}/target"
folders:
  - "test1"
comparison:
  algorithm: "sha256"
  level: "standard"
storage:
  device_type: "hdd"
  max_threads: 4
EOL

cat > "${CONFIG_DIR}/valid.yaml" << EOL
source: "${TEST_DIR}/source"
target: "${TEST_DIR}/target"
folders:
  - "test1"
comparison:
  algorithm: "sha256"
  level: "standard"
storage:
  device_type: "hdd"
  max_threads: 4
EOL

cat > "${CONFIG_DIR}/invalid_algo.yaml" << EOL
source: "${TEST_DIR}/source"
target: "${TEST_DIR}/target"
comparison:
  algorithm: "invalid"
EOL

cat > "${CONFIG_DIR}/invalid_threads.yaml" << EOL
source: "${TEST_DIR}/source"
target: "${TEST_DIR}/target"
comparison:
  algorithm: "sha256"
storage:
  max_threads: -1
EOL

cat > "${CONFIG_DIR}/missing_source.yaml" << EOL
target: "${TEST_DIR}/target"
comparison:
  algorithm: "sha256"
EOL

cat > "${CONFIG_DIR}/non_existent_source.yaml" << EOL
source: "${TEST_DIR}/not_exists"
target: "${TEST_DIR}/target"
comparison:
 algorithm: "sha256"
storage:
 device_type: "hdd"
 max_threads: 4
EOL

cat > "${CONFIG_DIR}/invalid_yaml.yaml" << 'EOL'
source: ${TEST_DIR}/source
target: [missing bracket
EOL

cat > "${CONFIG_DIR}/missing_comparison.yaml" << EOL
source: "${TEST_DIR}/source"
target: "${TEST_DIR}/target"
storage:
 device_type: "invalid"
EOL

cat > "${CONFIG_DIR}/invalid_device.yaml" << EOL
source: "${TEST_DIR}/source"
target: "${TEST_DIR}/target"
comparison:
 algorithm: "sha256"
storage:
 device_type: "unknown"
 max_threads: 4
EOL

cat > "${CONFIG_DIR}/excess_threads.yaml" << EOL
source: "${TEST_DIR}/source"
target: "${TEST_DIR}/target"
comparison:
 algorithm: "sha256"
storage:
 device_type: "hdd"
 max_threads: 32
EOL

cat > "${CONFIG_DIR}/defaults.yaml" << EOL
source: "${TEST_DIR}/source"
target: "${TEST_DIR}/target"
EOL
}


run_test() {
    local name=$1
    local config=$2
    local expected_status=$3

    printf "  Testing %-30s" "$name..."
    output=$(./bin/backup-butler sync -c "$config" 2>&1)
    status=$?

    if [[ $status -eq $expected_status ]]; then
        printf "${GREEN}PASS${NC}\n"
        $VERBOSE && echo "$output"
        return 0
    else
        printf "${RED}FAIL${NC}\n"
        $VERBOSE || echo "$output"  # Show output on failure unless verbose
        return 1
    fi
}

main() {
    local failed=0

    printf "Running configuration tests...\n"
    setup_test_env
    setup_configs

    # Run tests
    run_test "valid config" "${CONFIG_DIR}/valid.yaml" 0 || failed=1
    run_test "invalid algorithm" "${CONFIG_DIR}/invalid_algo.yaml" 1 || failed=1
    run_test "excess thread count" "${CONFIG_DIR}/excess_threads.yaml" 1 || failed=1
    run_test "missing source directory" "${CONFIG_DIR}/missing_source.yaml" 1 || failed=1
    run_test "non-existent source directory" "${CONFIG_DIR}/non_existent_source.yaml" 1 || failed=1
    run_test "invalid YAML syntax" "${CONFIG_DIR}/invalid_yaml.yaml" 1 || failed=1
    run_test "missing comparison settings" "${CONFIG_DIR}/missing_comparison.yaml" 1 || failed=1
    run_test "invalid device type" "${CONFIG_DIR}/invalid_device.yaml" 1 || failed=1
    run_test "minimal config with defaults" "${CONFIG_DIR}/defaults.yaml" 0 || failed=1

    # Cleanup
    rm -rf "$TEST_ROOT"

    return $failed
}

main