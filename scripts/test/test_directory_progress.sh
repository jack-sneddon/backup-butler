# scripts/test/test_directory_progress.sh
#!/bin/bash
source scripts/test/common.sh
LOG_LEVEL=${LOG_LEVEL:-"debug"}  # Default to debug if not set

# This will show the output in real-time and also save it

# For Debug, when running the test, use:
# ./scripts/test/test_directory_progress.sh --log-level debug 2>&1 | tee out/output.txt
# or Or if you want to see the debug output directly:
# LOG_LEVEL=debug ./scripts/test/test_directory_progress.sh

setup_progress_test() {
    rm -rf "$TEST_ROOT"
    
    # Create nested directory structure with files of varying sizes
    mkdir -p "$TEST_DIR/dir1/subdir1" \
             "$TEST_DIR/dir2/subdir2" \
             "$TEST_DIR/dir3/subdir3" \
             "$TEST_DIR-target"

    # Create test files in different directories
    # dir1: 3 files, 5MB total
    dd if=/dev/urandom of="$TEST_DIR/dir1/file1.dat" bs=1M count=1 2>/dev/null
    dd if=/dev/urandom of="$TEST_DIR/dir1/file2.dat" bs=1M count=2 2>/dev/null
    dd if=/dev/urandom of="$TEST_DIR/dir1/subdir1/file3.dat" bs=1M count=2 2>/dev/null

    # dir2: 2 files, 3MB total
    dd if=/dev/urandom of="$TEST_DIR/dir2/file4.dat" bs=1M count=1 2>/dev/null
    dd if=/dev/urandom of="$TEST_DIR/dir2/subdir2/file5.dat" bs=1M count=2 2>/dev/null

    # dir3: 1 file, 2MB total
    dd if=/dev/urandom of="$TEST_DIR/dir3/subdir3/file6.dat" bs=1M count=2 2>/dev/null

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
}

test_directory_progress_display() {
    printf "  Testing directory progress display... "
    
    # Run sync command and capture output
    output=$(./bin/backup-butler sync -c "$TEST_DIR/test_config.yaml" 2>&1)
    
    # Check for required progress elements
    if [[ $output =~ "Processing:" && \
          $output =~ "files)" && \
          $output =~ "Currently Processing:" && \
          $output =~ "Statistics:" ]]; then
        printf "${GREEN}PASS${NC}\n"
        $VERBOSE && echo "$output"
        return 0
    else
        printf "${RED}FAIL${NC}\n"
        $VERBOSE || echo "$output"
        return 1
    fi
}

test_directory_order() {
    printf "  Testing directory processing order... "
    
    # Run sync command and capture output with stderr
    output=$(./bin/backup-butler sync -c "$TEST_DIR/test_config.yaml" --log-level "$LOG_LEVEL" 2>&1)
    
    if $VERBOSE; then
        echo "Command output:"
        echo "-------------"
        echo "$output"
        echo "-------------"
    fi
    
    # Create temporary files to store directory appearances
    dir1_line=$(echo "$output" | grep -n "dir1" | head -1 | cut -d: -f1)
    dir2_line=$(echo "$output" | grep -n "dir2" | head -1 | cut -d: -f1)
    dir3_line=$(echo "$output" | grep -n "dir3" | head -1 | cut -d: -f1)
    
    # Debug information
    echo "Directory line numbers:"
    echo "dir1: $dir1_line"
    echo "dir2: $dir2_line"
    echo "dir3: $dir3_line"
    
    # Check if directories are processed in order
    if [[ $dir1_line -lt $dir2_line && $dir2_line -lt $dir3_line ]]; then
        printf "${GREEN}PASS${NC}\n"
        return 0
    else
        printf "${RED}FAIL${NC}\n"
        echo "Directory processing order incorrect:"
        echo "Expected: dir1 < dir2 < dir3"
        echo "Got: dir1($dir1_line), dir2($dir2_line), dir3($dir3_line)"
        echo "Full output:"
        echo "$output"
        return 1
    fi
}

test_progress_statistics() {
    printf "  Testing progress statistics... "
    
    # Run sync command and capture output with stderr
    output=$(./bin/backup-butler sync -c "$TEST_DIR/test_config.yaml" --log-level "$LOG_LEVEL" 2>&1)
    
    if $VERBOSE; then
        echo "Command output:"
        echo "-------------"
        echo "$output"
        echo "-------------"
    fi
    
    # Debug information
    echo "Looking for statistics in output:"
    echo "- Processed:"
    echo "- Remaining:"
    echo "- Total Time:"
    
    # Check for statistics in the output
    if [[ $output =~ "Processed:" && \
          $output =~ "Remaining:" && \
          $output =~ "Total Time:" ]]; then
        printf "${GREEN}PASS${NC}\n"
        return 0
    else
        printf "${RED}FAIL${NC}\n"
        echo "Missing required statistics. Found:"
        [[ $output =~ "Processed:" ]] && echo "✓ Processed" || echo "✗ Processed"
        [[ $output =~ "Remaining:" ]] && echo "✓ Remaining" || echo "✗ Remaining"
        [[ $output =~ "Total Time:" ]] && echo "✓ Total Time" || echo "✗ Total Time"
        echo "Full output:"
        echo "$output"
        return 1
    fi
}

main() {
    local failed=0
    printf "Running directory progress tests...\n"
    
    setup_progress_test
    
    test_directory_progress_display || failed=1
    test_directory_order || failed=1
    test_progress_statistics || failed=1
    
    rm -rf "$TEST_ROOT"
    return $failed
}

main
