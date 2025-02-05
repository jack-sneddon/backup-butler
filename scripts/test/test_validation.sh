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


main() {
    local failed=0
    printf "Running validation tests...\n\n"

    setup_test_env

    test_quick_comparison || failed=1
    test_standard_comparison || failed=1
    test_deep_comparison || failed=1

    rm -rf "$TEST_ROOT"

    return $failed
}

main