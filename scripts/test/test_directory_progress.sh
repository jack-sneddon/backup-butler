#!/bin/bash
source scripts/test/common.sh

setup_progress_test() {
    test_log "Setting up progress test"
    
    # Create directory structure
    mkdir -p "$TEST_DIR/dir1/subdir1" \
             "$TEST_DIR/dir2/subdir2" \
             "$TEST_DIR/dir3/subdir3" \
             "$TEST_DIR-target"

    test_log "Created directory structure"

    # Create test files with specific sizes
    local files=(
        "dir1/file1.dat:1"
        "dir1/file2.dat:2"
        "dir1/subdir1/file3.dat:2"
        "dir2/file4.dat:1"
        "dir2/subdir2/file5.dat:2"
        "dir3/subdir3/file6.dat:2"
    )

    for file_spec in "${files[@]}"; do
        IFS=':' read -r file size <<< "$file_spec"
        test_log "Creating $file ($size MB)"
        dd if=/dev/urandom of="$TEST_DIR/$file" bs=1M count=$size 2>/dev/null
    done

    # Create test config
    cat > "$TEST_DIR/test_config.yaml" << EOL
source: "${TEST_DIR}"
target: "${TEST_DIR}-target"
comparison:
  algorithm: "sha256"
  level: "standard"
storage:
  source:
    type: "hdd"
    max_threads: 4
  target:
    type: "hdd"
    max_threads: 4
EOL

    test_log "Created test configuration"
}

test_directory_progress_display() {
    test_section "progress display"

    # Run command and capture output
    local output
    output=$(run_backup_butler sync -c "$TEST_DIR/test_config.yaml" 2>&1)

    # Define required elements
    local success=true
    local required_elements=(
        "Processing:"
        "files)"
        "Currently Processing:"
        "Statistics:"
    )

    # Check for all required elements
    for element in "${required_elements[@]}"; do
        if ! assert_contains "$output" "$element" "Progress element: $element"; then
            success=false
        fi
    done

    report_result "$success" "$output" "Missing required progress elements"
}

test_directory_order() {
    printf "  Testing directory processing order... "
    
    # Run sync command and capture output with stderr
    output=$(run_backup_butler sync -c "$TEST_DIR/test_config.yaml" 2>&1)
    
    # Create temporary files to store directory processing lines
    # Look for either INFO level logs or progress output
    dir1_line=$(echo "$output" | grep -n "Processing.*dir1" | head -1 | cut -d: -f1)
    dir2_line=$(echo "$output" | grep -n "Processing.*dir2" | head -1 | cut -d: -f1)
    dir3_line=$(echo "$output" | grep -n "Processing.*dir3" | head -1 | cut -d: -f1)

    if $VERBOSE; then
        echo "Directory processing lines:"
        echo "dir1: line $dir1_line"
        echo "dir2: line $dir2_line"
        echo "dir3: line $dir3_line"
        echo "Full output:"
        echo "$output"
    fi
    
    # Check if directories are processed in order
    if [[ -n $dir1_line && -n $dir2_line && -n $dir3_line && \
          $dir1_line -lt $dir2_line && $dir2_line -lt $dir3_line ]]; then
        printf "${GREEN}PASS${NC}\n"
        return 0
    else
        printf "${RED}FAIL${NC}\n"
        echo "Directory processing order incorrect:"
        echo "Expected: dir1 < dir2 < dir3"
        echo "Got: dir1($dir1_line), dir2($dir2_line), dir3($dir3_line)"
        if $VERBOSE; then
            echo "Full output:"
            echo "$output"
        fi
        return 1
    fi
}

test_progress_statistics() {
    test_section "progress statistics"
    
    local output
    output=$(run_backup_butler sync -c "$TEST_DIR/test_config.yaml" 2>&1)
    
    # Check for statistics
    local success=true
    local stats=(
        "Processed:"
        "Remaining:"
        "Total Time:"
    )
    
    for stat in "${stats[@]}"; do
        if ! assert_contains "$output" "$stat" "Statistic: $stat"; then
            success=false
        fi
    done
    
    report_result "$success" "$output" "Missing required statistics"
}

main() {
    test_header "directory progress tests"

    local failed=0
    
    setup_progress_test
    
    test_directory_progress_display || failed=1
    test_directory_order || failed=1
    test_progress_statistics || failed=1
    
    cleanup_test_env
    return $failed
}

# Process args and run
process_args "$@"
main