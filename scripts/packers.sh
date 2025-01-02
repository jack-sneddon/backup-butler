#!/bin/bash

# Define target folder
target_folder="/Users/jack/tmp/2024-11-27-backup"

# Exclude .versions directory from calculations
exclude_dir="${target_folder}/.versions"

# Function to calculate directory size and file count
get_dir_stats() {
  local dir="$1"
  local size=$(du -sh "$dir" | awk '{print $1}')
  local count=$(find "$dir" -type f -not -path "*/.versions/*" | wc -l)
  echo "Size: ${size}, File Count: ${count}"
}

# Get initial directory stats
echo "Before Backup:"
echo "Target Directory: ${target_folder}"
echo "Excluding: ${exclude_dir}"
echo "Initial Stats:"
get_dir_stats "${target_folder}"

# Execute backup command
echo
echo "Running Backup Command:"
echo "./backup-butler -config configs/config.yaml"
./backup-butler -config configs/config.yaml

# Get final directory stats
echo
echo "After Backup:"
echo "Target Directory: ${target_folder}"
echo "Excluding: ${exclude_dir}"
echo "Final Stats:"
get_dir_stats "${target_folder}"