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
  level: "quick"
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
  level: "standard"
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
  level: "deep"
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
    
    # Test content differences within first 32KB
    printf "  Testing early content differences... "
    dd if=/dev/urandom of="$TEST_DIR/content/diff_content.dat" bs=1K count=16 2>/dev/null
    dd if=/dev/urandom of="$TEST_DIR-target/content/diff_content.dat" bs=1K count=16 2>/dev/null

    output=$(run_backup_butler check -c "${TEST_DIR}/config/config_standard.yaml" 2>&1)
    if [[ $output =~ "content/diff_content.dat" && $output =~ "*" ]]; then
        printf "${GREEN}PASS${NC}\n"
    else
        printf "${RED}FAIL${NC}\n"
        return 1
    fi

    # Test identical first 32KB
    printf "  Testing identical first 32KB... "
    dd if=/dev/urandom of="$TEST_DIR/content/partial.dat" bs=1K count=32 2>/dev/null
    cp "$TEST_DIR/content/partial.dat" "$TEST_DIR-target/"
    dd if=/dev/urandom of="$TEST_DIR-target/content/partial.dat" bs=1K seek=32 count=32 conv=notrunc 2>/dev/null

    if [[ $output =~ "content/partial.dat" && $output =~ "=" ]]; then
        printf "${GREEN}PASS${NC}\n"
    else
        printf "${RED}FAIL${NC}\n"
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
  level: "standard"
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

test_deep_validation_scenarios() {
    printf "Testing deep validation failure scenarios...\n"

    # Setup for deep validation tests
    local test_dir="${TEST_DIR}/deep_scenarios"
    mkdir -p "$test_dir"
    mkdir -p "$test_dir-target"

    # Create deep validation config
    cat > "${TEST_DIR}/config/config_deep_test.yaml" << EOL
source: "${test_dir}"
target: "${test_dir}-target"
validation:
  level: "deep"
  buffer_size: 32768
  hash_algorithm: "sha256"
exclude:
  - "config/*"
  - "*.tmp"
  - "*.bak"
EOL

    # Test 1: Metadata mismatch → fail
    printf "  Testing metadata mismatch (deep)... "
    dd if=/dev/urandom of="$test_dir/meta_diff.dat" bs=1M count=1 2>/dev/null
    cp "$test_dir/meta_diff.dat" "$test_dir-target/"
    # Change modification time of target file
    touch -t 202001010000 "$test_dir-target/meta_diff.dat"

    output=$(./bin/backup-butler check -c "${TEST_DIR}/config/config_deep_test.yaml" 2>&1)
    if [[ $output =~ "meta_diff.dat" && $output =~ "*" && $output =~ "[deep]" ]]; then
        printf "${GREEN}PASS${NC}\n"
        $VERBOSE && echo "$output"
    else
        printf "${RED}FAIL${NC}\n"
        echo "Expected: File should be marked as different (*) with deep validation level"
        echo "Got: $output"
        return 1
    fi

    # Test 2: Metadata match + partial content mismatch
    printf "  Testing partial content mismatch (deep)... "
    dd if=/dev/urandom of="$test_dir/partial_diff.dat" bs=1K count=32 2>/dev/null
    cp "$test_dir/partial_diff.dat" "$test_dir-target/"
    # Modify first 32KB of target file
    dd if=/dev/urandom of="$test_dir-target/partial_diff.dat" bs=1K count=32 conv=notrunc 2>/dev/null

    output=$(./bin/backup-butler check -c "${TEST_DIR}/config/config_deep_test.yaml" 2>&1)
    if [[ $output =~ "partial_diff.dat" && $output =~ "*" && $output =~ "[deep]" ]]; then
        printf "${GREEN}PASS${NC}\n"
        $VERBOSE && echo "$output"
    else
        printf "${RED}FAIL${NC}\n"
        echo "Expected: File should fail at standard validation level"
        echo "Got: $output"
        return 1
    fi

    # Test 3: Metadata match + partial content match + full content mismatch
    printf "  Testing full content mismatch (deep)... "
    # Create 2MB file
    dd if=/dev/urandom of="$test_dir/full_diff.dat" bs=1M count=2 2>/dev/null
    # Copy first 32KB to target to ensure partial match
    dd if="$test_dir/full_diff.dat" of="$test_dir-target/full_diff.dat" bs=1K count=32 2>/dev/null
    # Add different content after 32KB
    dd if=/dev/urandom of="$test_dir-target/full_diff.dat" bs=1K seek=32 count=2016 2>/dev/null

    output=$(./bin/backup-butler check -c "${TEST_DIR}/config/config_deep_test.yaml" 2>&1)
    if [[ $output =~ "full_diff.dat" && $output =~ "*" && $output =~ "[deep]" ]]; then
        printf "${GREEN}PASS${NC}\n"
        $VERBOSE && echo "$output"
    else
        printf "${RED}FAIL${NC}\n"
        echo "Expected: File should fail at deep validation level"
        echo "Got: $output"
        return 1
    fi

    # Cleanup
    rm -rf "$test_dir" "$test_dir-target"
}


main() {
    local failed=0
    printf "Running validation tests...\n\n"

    setup_test_env

    test_file_status || failed=1

    test_quick_comparison || failed=1
    test_standard_comparison || failed=1
    test_deep_comparison || failed=1

    test_deep_validation_scenarios || failed=1

    rm -rf "$TEST_ROOT"

    return $failed
}

main