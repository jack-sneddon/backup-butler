# scripts/test/common.sh
#!/bin/bash
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

TEST_DIR="/tmp/backup-butler-test-scan"

setup_test_dir() {
    rm -rf "$TEST_DIR"
    mkdir -p "${TEST_DIR}/dir1/subdir1" "${TEST_DIR}/dir2"
    dd if=/dev/zero of="${TEST_DIR}/dir1/file1" bs=1M count=1 2>/dev/null
    dd if=/dev/zero of="${TEST_DIR}/dir1/subdir1/file2" bs=1M count=2 2>/dev/null
    dd if=/dev/zero of="${TEST_DIR}/dir2/file3" bs=1M count=3 2>/dev/null

    # Create test config
    cat > "${TEST_DIR}/test_config.yaml" << EOL
source: "${TEST_DIR}"
target: "${TEST_DIR}-target"
comparison:
  algorithm: "sha256"
  level: "standard"
storage:
  device_type: "hdd"
  max_threads: 4
EOL
}

VERBOSE=false
while getopts "v" opt; do
    case $opt in v) VERBOSE=true ;; esac
done