#!/bin/zsh

# Function to classify file sizes
classify_file_size() {
    local size=$1
    if (( size < 2000000 )); then
        echo "small"
    elif (( size >= 2000000 && size <= 6000000 )); then
        echo "medium"
    elif (( size > 6000000 && size <= 30000000 )); then
        echo "large"
    else
        echo "extra_large"
    fi
}

# Initialize counters
total_files=0
small_files=0
medium_files=0
large_files=0
extra_large_files=0

# Specify the directory to analyze
directory="${1:-.}"  # Default to the current directory if not specified

# Traverse the directory
while IFS= read -r -d '' file; do
    ((total_files++))
    file_size=$(stat -f%z "$file")  # Get the file size in bytes
    case $(classify_file_size "$file_size") in
        small) ((small_files++)) ;;
        medium) ((medium_files++)) ;;
        large) ((large_files++)) ;;
        extra_large) ((extra_large_files++)) ;;
    esac
done < <(find "$directory" -type f -print0)

# Output results
echo "Directory: $directory"
echo "Total files: $total_files"
echo "Small files (< 2MB): $small_files"
echo "Medium files (2MB - 6MB): $medium_files"
echo "Large files (6MB - 30MB): $large_files"
echo "Extra Large files (> 30MB): $extra_large_files"