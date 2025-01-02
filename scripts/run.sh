#!/bin/bash

# test_backup.sh
# Script to test backup-butler functionality and validate content
# Usage: ./test_backup.sh <config_file>

set -e  # Exit on any error

if [ -z "$1" ]; then
    echo "Usage: $0 <config_file>"
    echo "Example: $0 configs/config.yaml"
    exit 1
fi

CONFIG_FILE=$1
SOURCE_DIR=$(grep "source_directory" "$CONFIG_FILE" | cut -d':' -f2 | tr -d ' "')
TARGET_DIR=$(grep "target_directory" "$CONFIG_FILE" | cut -d':' -f2 | tr -d ' "')

print_stats() {
    local dir=$1
    local desc=$2

    # Count files excluding .versions directory
    local file_count=$(find "$dir" -type f -not -path "*/.versions/*" | wc -l)
    # Get directory size excluding .versions
    local dir_size=$(du -sh $(find "$dir" -not -path "*/.versions*" -not -name ".versions") | tail -1 | cut -f1)

    echo "$desc Statistics:"
    echo "  Files: $file_count"
    echo "  Size:  $dir_size"
    echo "------------------------"
}

echo "Starting backup-butler test sequence..."
echo "Using config: $CONFIG_FILE"
echo "Source directory: $SOURCE_DIR"
echo "Target directory: $TARGET_DIR"

echo -e "\n1. Source directory statistics"
print_stats "$SOURCE_DIR" "Source"

echo -e "\n2. Clean start - removing target directory"
rm -rf "$TARGET_DIR"
echo "Target directory removed"

echo -e "\n3. Initial backup - should copy all files"
./backup-butler -config "$CONFIG_FILE"

echo -e "\n4. Target directory statistics after initial backup"
print_stats "$TARGET_DIR" "Target"

echo -e "\n5. Checking version structure"
tree "$TARGET_DIR/.versions"

echo -e "\n6. Displaying version history"
./backup-butler -config "$CONFIG_FILE" -versions

echo -e "\n7. Simulating partial deletion (removing Packers directory)"
rm -rf "$TARGET_DIR/Packers"
echo "Packers directory removed"

echo -e "\n8. Target directory statistics after deletion"
print_stats "$TARGET_DIR" "Target"

echo -e "\n9. Running incremental backup - should copy only Packers"
./backup-butler -config "$CONFIG_FILE"

echo -e "\n10. Final target directory statistics"
print_stats "$TARGET_DIR" "Target"

echo -e "\n11. Final version history"
./backup-butler -config "$CONFIG_FILE" -versions

echo -e "\n12. Validation Summary"
src_count=$(find "$SOURCE_DIR" -type f | wc -l)
tgt_count=$(find "$TARGET_DIR" -type f -not -path "*/.versions/*" | wc -l)

if [ "$src_count" -eq "$tgt_count" ]; then
    echo "✅ File count matches: $src_count files"
else
    echo "❌ File count mismatch: Source=$src_count, Target=$tgt_count"
fi

src_size=$(du -s "$SOURCE_DIR" | cut -f1)
tgt_size=$(du -s $(find "$TARGET_DIR" -not -path "*/.versions*" -not -name ".versions") | cut -f1)

if [ "$src_size" -eq "$tgt_size" ]; then
    echo "✅ Total size matches: $src_size blocks"
else
    echo "❌ Size mismatch: Source=$src_size, Target=$tgt_size blocks"
fi

echo -e "\nTest sequence completed!"