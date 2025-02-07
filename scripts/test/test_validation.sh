#!/bin/bash
# scripts/test/test_validation.sh
source scripts/test/common.sh
# Parse command line arguments
VERBOSE=false
LOG_LEVEL=""

while [[ $# -gt 0 ]]; do
    case $1 in
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        --log-level)
            LOG_LEVEL="$2"
            shift 2
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Function to run backup-butler with appropriate flags
run_backup_butler() {
    cmd="./bin/backup-butler $*"
    if [[ -n "$LOG_LEVEL" ]]; then
        cmd="$cmd --log-level $LOG_LEVEL"
    fi
    eval "$cmd"
}

setup_test_env() {
    printf "Setting up test environment...\n"
    if $VERBOSE; then
        printf "  TEST_ROOT: %s\n" "$TEST_ROOT"
        printf "  Creating test directories and files...\n"
    fi
    
    rm -rf "$TEST_ROOT"

    # Create test directories
    if $VERBOSE; then
        printf "  Creating directory structure:\n"
    fi
    
    mkdir -p "$TEST_DIR/identical" \
             "$TEST_DIR/metadata" \
             "$TEST_DIR/content" \
             "$TEST_DIR-target/identical" \
             "$TEST_DIR-target/metadata" \
             "$TEST_DIR-target/content" \
             "$TEST_DIR/config"

    if $VERBOSE; then
        printf "  Creating test files:\n"
    fi

    # 1. Create identical files (both content and metadata)
    dd if=/dev/urandom of="$TEST_DIR/identical/same1.dat" bs=1M count=1 2>/dev/null
    cp "$TEST_DIR/identical/same1.dat" "$TEST_DIR-target/identical/"

    # 2. Create files with same content but different metadata
    dd if=/dev/urandom of="$TEST_DIR/metadata/file1.dat" bs=1M count=1 2>/dev/null
    cp "$TEST_DIR/metadata/file1.dat" "$TEST_DIR-target/metadata/"
    
    # Get current time in seconds since epoch
    current_time=$(date +%s)
    # Subtract 2 days worth of seconds (2 * 24 * 60 * 60 = 172800)
    old_time=$((current_time - 172800))
    
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        touch -mt $(date -r $old_time "+%Y%m%d%H%M.%S") "$TEST_DIR-target/metadata/file1.dat"
    else
        # Linux
        touch --date="@$old_time" "$TEST_DIR-target/metadata/file1.dat"
    fi

    # 3. Create files with different content
    # 3a. Different content, same size
    dd if=/dev/urandom of="$TEST_DIR/content/diff_content.dat" bs=1M count=1 2>/dev/null
    dd if=/dev/urandom of="$TEST_DIR-target/content/diff_content.dat" bs=1M count=1 2>/dev/null
    
    # 3b. First 32KB same, rest different (for testing standard validation)
    dd if=/dev/urandom of="$TEST_DIR/content/partial.dat" bs=1M count=2 2>/dev/null
    cp "$TEST_DIR/content/partial.dat" "$TEST_DIR-target/content/"
    dd if=/dev/urandom of="$TEST_DIR-target/content/partial.dat" bs=1M count=2 seek=1 conv=notrunc 2>/dev/null

    if $VERBOSE; then
        printf "  Creating test configurations:\n"
    fi

    # Create test configurations
    cat > "${TEST_DIR}/config/config_quick.yaml" << EOL
source: "${TEST_DIR}"
target: "${TEST_DIR}-target"
validation:
  default_level: "quick"
  on_mismatch: "none"
  buffer_size: 32768
  hash_algorithm: "sha256"
exclude:
  - "config/*"
  - "*.tmp"
  - "*.bak"
EOL

    cat > "${TEST_DIR}/config/config_standard.yaml" << EOL
source: "${TEST_DIR}"
target: "${TEST_DIR}-target"
validation:
  default_level: "standard"
  on_mismatch: "none"
  buffer_size: 32768
  hash_algorithm: "sha256"
exclude:
  - "config/*"
  - "*.tmp"
  - "*.bak"
EOL

    cat > "${TEST_DIR}/config/config_deep.yaml" << EOL
source: "${TEST_DIR}"
target: "${TEST_DIR}-target"
validation:
  default_level: "deep"
  on_mismatch: "none"
  buffer_size: 1048576
  hash_algorithm: "sha256"
exclude:
  - "config/*"
  - "*.tmp"
  - "*.bak"
EOL

}

test_quick_comparison() {
    printf "Testing quick comparison...\n"
    
    # Test identical files
    printf "  Testing identical files... "
    output=$(run_backup_butler check -c "${TEST_DIR}/config/config_quick.yaml" 2>&1)
    if [[ $output =~ "identical/same1.dat" && $output =~ "=" ]]; then
        printf "${GREEN}PASS${NC}\n"
        if $VERBOSE; then
            echo "Command output:"
            echo "-------------"
            echo "$output"
            echo "-------------"
        fi
    else
        printf "${RED}FAIL${NC}\n"
        echo "Command output:"
        echo "-------------"
        echo "$output"
        echo "-------------"
        return 1
    fi

    # Test metadata differences
    printf "  Testing metadata differences... "
    if [[ $output =~ "metadata/file1.dat" && $output =~ "*" ]]; then
        printf "${GREEN}PASS${NC}\n"
        if $VERBOSE; then
            echo "Command output:"
            echo "-------------"
            echo "$output"
            echo "-------------"
        fi
    else
        printf "${RED}FAIL${NC}\n"
        echo "Command output:"
        echo "-------------"
        echo "$output"
        echo "-------------"
        return 1
    fi

    # Quick check should not detect content differences if size is same
    printf "  Testing content blindness... "
    if [[ $output =~ "content/diff_content.dat" && $output =~ "=" ]]; then
        printf "${GREEN}PASS${NC}\n"
        if $VERBOSE; then
            echo "Command output:"
            echo "-------------"
            echo "$output"
            echo "-------------"
        fi
    else
        printf "${RED}FAIL${NC}\n"
        echo "Command output:"
        echo "-------------"
        echo "$output"
        echo "-------------"
        return 1
    fi

    return 0
}

test_standard_comparison() {
    printf "Testing standard comparison...\n"
    
    # Test content differences
    printf "  Testing content differences... "
    #output=$(./bin/backup-butler check -c "${TEST_DIR}/config/config_standard.yaml" 2>&1)
    output=$(run_backup_butler check -c "${TEST_DIR}/config/config_standard.yaml" 2>&1)
    if [[ $output =~ "content/diff_content.dat" && $output =~ "*" ]]; then
        printf "${GREEN}PASS${NC}\n"
        if $VERBOSE; then
            echo "Command output:"
            echo "-------------"
            echo "$output"
            echo "-------------"
        fi
    else
        printf "${RED}FAIL${NC}\n"
        echo "Command output:"
        echo "-------------"
        echo "$output"
        echo "-------------"
        return 1
    fi

    # Test partial content differences
    printf "  Testing partial content... "
    if [[ $output =~ "content/partial.dat" && $output =~ "=" ]]; then
        printf "${GREEN}PASS${NC}\n"
        if $VERBOSE; then
            echo "Command output:"
            echo "-------------"
            echo "$output"
            echo "-------------"
        fi
    else
        printf "${RED}FAIL${NC}\n"
        echo "Command output:"
        echo "-------------"
        echo "$output"
        echo "-------------"
        return 1
    fi

    return 0
}

test_deep_comparison() {
    printf "Testing deep comparison...\n"
    
    printf "  Testing full content verification... "
    output=$(run_backup_butler check -c "${TEST_DIR}/config/config_deep.yaml" 2>&1)
    if [[ $output =~ "content/partial.dat" && $output =~ "*" ]]; then
        printf "${GREEN}PASS${NC}\n"
        if $VERBOSE; then
            echo "Command output:"
            echo "-------------"
            echo "$output"
            echo "-------------"
        fi
    else
        printf "${RED}FAIL${NC}\n"
        echo "Command output:"
        echo "-------------"
        echo "$output"
        echo "-------------"
        return 1
    fi

    return 0
}

test_file_status() {
    printf "Testing file status handling...\n"

    # Test new files (exist in source but not target)
    printf "  Testing new file detection... "
    dd if=/dev/urandom of="$TEST_DIR/source_only.dat" bs=1M count=1 2>/dev/null
    output=$(run_backup_butler check -c "${TEST_DIR}/config/config_quick.yaml" 2>&1)

    if [[ $output =~ "    + source_only.dat [quick]" ]]; then
        printf "${GREEN}PASS${NC}\n"
        $VERBOSE && echo "$output"
    else
        printf "${RED}FAIL${NC}\n"
        echo "Expected pattern: '    + source_only.dat [quick]'"
        echo "Got output:"
        echo "$output"
        return 1
    fi

    # Test missing files (exist in target but not source)
    printf "  Testing missing file detection... "
    dd if=/dev/urandom of="$TEST_DIR-target/target_only.dat" bs=1M count=1 2>/dev/null
    output=$(run_backup_butler check -c "${TEST_DIR}/config/config_quick.yaml" 2>&1)

    if [[ $output =~ "    - target_only.dat" ]]; then
        printf "${GREEN}PASS${NC}\n"
        $VERBOSE && echo "$output"
    else
        printf "${RED}FAIL${NC}\n"
        echo "Expected pattern: '    - target_only.dat'"
        echo "Got output:"
        echo "$output"
        return 1
    fi

    # Test error status (unreadable files in both source and target)
    printf "  Testing error status... "
    # Create a specific config for standard validation
    cat > "${TEST_DIR}/config/config_standard_test.yaml" << EOL
source: "${TEST_DIR}"
target: "${TEST_DIR}-target"
validation:
  default_level: "standard"
  buffer_size: 32768
  hash_algorithm: "sha256"
exclude:
  - "config/*"
  - "*.tmp"
  - "*.bak"
EOL

    # Create identical files first
    dd if=/dev/urandom of="$TEST_DIR/unreadable.dat" bs=1M count=1 2>/dev/null
    cp "$TEST_DIR/unreadable.dat" "$TEST_DIR-target/unreadable.dat"
    # Make both unreadable
    chmod 000 "$TEST_DIR/unreadable.dat"
    chmod 000 "$TEST_DIR-target/unreadable.dat"

    output=$(run_backup_butler check -c "${TEST_DIR}/config/config_standard_test.yaml" 2>&1)

    if [[ $output =~ "    ! unreadable.dat" ]]; then
        printf "${GREEN}PASS${NC}\n"
        $VERBOSE && echo "$output"
    else
        printf "${RED}FAIL${NC}\n"
        echo "Expected pattern: '    ! unreadable.dat'"
        echo "Got output:"
        echo "$output"
        return 1
    fi

    # Cleanup test files
    chmod 644 "$TEST_DIR/unreadable.dat" 2>/dev/null
    chmod 644 "$TEST_DIR-target/unreadable.dat" 2>/dev/null
    rm -f "$TEST_DIR/source_only.dat"
    rm -f "$TEST_DIR-target/target_only.dat"
    rm -f "$TEST_DIR/unreadable.dat"
    rm -f "$TEST_DIR-target/unreadable.dat"
    rm -f "$TEST_DIR/config/config_standard_test.yaml"

    return 0
}

test_validation_escalation() {
    printf "Testing validation level escalation...\n"

    # Create config for quick->standard escalation
    cat > "${TEST_DIR}/config/config_quick_standard.yaml" << EOL
source: "${TEST_DIR}"
target: "${TEST_DIR}-target"
validation:
  default_level: "quick"
  on_mismatch: "standard"
  buffer_size: 32768
  hash_algorithm: "sha256"
exclude:
  - "config/*"
  - "*.tmp"
  - "*.bak"
EOL

    # Create config for standard->deep escalation
    cat > "${TEST_DIR}/config/config_standard_deep.yaml" << EOL
source: "${TEST_DIR}"
target: "${TEST_DIR}-target"
validation:
  default_level: "standard"
  on_mismatch: "deep"
  buffer_size: 32768
  hash_algorithm: "sha256"
exclude:
  - "config/*"
  - "*.tmp"
  - "*.bak"
EOL

    # Test quick to standard escalation
    printf "  Testing quick to standard escalation... "
    # Create file with same size but different content
    dd if=/dev/urandom of="$TEST_DIR/escalate_quick.dat" bs=1M count=1 2>/dev/null
    dd if=/dev/urandom of="$TEST_DIR-target/escalate_quick.dat" bs=1M count=1 2>/dev/null

    output=$(run_backup_butler check -c "${TEST_DIR}/config/config_quick_standard.yaml" 2>&1)

    if [[ $output =~ "    * escalate_quick.dat [standard]" ]]; then
        printf "${GREEN}PASS${NC}\n"
        $VERBOSE && echo "$output"
    else
        printf "${RED}FAIL${NC}\n"
        echo "Expected pattern: '    * escalate_quick.dat [standard]'"
        echo "Got output:"
        echo "$output"
        return 1
    fi

    # Test standard to deep escalation
    printf "  Testing standard to deep escalation... "
    # Create file that matches in first 32KB but differs after
    dd if=/dev/urandom of="$TEST_DIR/escalate_standard.dat" bs=1M count=2 2>/dev/null
    cp "$TEST_DIR/escalate_standard.dat" "$TEST_DIR-target/"
    # Modify the file after the first 32KB
    dd if=/dev/urandom of="$TEST_DIR-target/escalate_standard.dat" bs=1K seek=32 count=32 conv=notrunc 2>/dev/null

    output=$(run_backup_butler check -c "${TEST_DIR}/config/config_standard_deep.yaml" 2>&1)

    if [[ $output =~ "    * escalate_standard.dat [deep]" ]]; then
        printf "${GREEN}PASS${NC}\n"
        $VERBOSE && echo "$output"
    else
        printf "${RED}FAIL${NC}\n"
        echo "Expected pattern: '    * escalate_standard.dat [deep]'"
        echo "Got output:"
        echo "$output"
        return 1
    fi

    # Cleanup test files
    rm -f "$TEST_DIR/escalate_quick.dat"
    rm -f "$TEST_DIR-target/escalate_quick.dat"
    rm -f "$TEST_DIR/escalate_standard.dat"
    rm -f "$TEST_DIR-target/escalate_standard.dat"
    rm -f "$TEST_DIR/config/config_quick_standard.yaml"
    rm -f "$TEST_DIR/config/config_standard_deep.yaml"

    return 0
}


main() {
    local failed=0
    printf "Running validation tests...\n\n"

    setup_test_env

    test_file_status || failed=1
    test_validation_escalation || failed=1

    test_quick_comparison || failed=1
    test_standard_comparison || failed=1
    test_deep_comparison || failed=1

    rm -rf "$TEST_ROOT"

    return $failed
}

main