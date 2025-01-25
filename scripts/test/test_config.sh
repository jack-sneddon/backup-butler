# scripts/test/test_config.sh
#!/bin/bash

GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

TEST_DIR="/tmp/backup-butler-test"
CONFIG_DIR="${TEST_DIR}/configs"

# Setup
mkdir -p "${TEST_DIR}/source" "${TEST_DIR}/target" "$CONFIG_DIR"

# Test configs
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


function run_test() {
    printf "Testing %s... " "$1"
    output=$(./bin/backup-butler sync -c "$2" 2>&1)
    ret=$?
    echo "Output: $output" # Debug line
    if [ $ret -eq $3 ]; then
        printf "${GREEN}PASS${NC}\n"
    else
        printf "${RED}FAIL${NC}\n"
    fi
}

function check_defaults() {
    printf "Testing Default values... "
    output=$(./bin/backup-butler sync -c "${CONFIG_DIR}/defaults.yaml" 2>&1)
    if [[ $output == *"Algorithm: sha256"* && 
          $output == *"Level: standard"* && 
          $output == *"Device: hdd"* && 
          $output == *"Threads: 4"* ]]; then
        printf "${GREEN}PASS${NC}\n"
    else
        printf "${RED}FAIL${NC}\n"
    fi
}

echo "Testing Backup Butler Configuration..."
run_test "Valid config" "${CONFIG_DIR}/valid.yaml" 0
run_test "Invalid algorithm" "${CONFIG_DIR}/invalid_algo.yaml" 1
# run_test "Invalid thread count (-1)" "${CONFIG_DIR}/invalid_threads.yaml" 1
run_test "Excess thread count" "${CONFIG_DIR}/excess_threads.yaml" 1
run_test "Missing source directory" "${CONFIG_DIR}/missing_source.yaml" 1
run_test "Non-existent source directory" "${CONFIG_DIR}/non_existent_source.yaml" 1
run_test "Invalid YAML syntax" "${CONFIG_DIR}/invalid_yaml.yaml" 1
run_test "Missing comparison settings" "${CONFIG_DIR}/missing_comparison.yaml" 1
run_test "Invalid device type" "${CONFIG_DIR}/missing_comparison.yaml" 1
run_test "Minimal config with defaults" "${CONFIG_DIR}/defaults.yaml" 0
check_defaults

# Cleanup
rm -rf "$TEST_DIR"