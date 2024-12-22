#!/bin/zsh

# Define source and destination paths for clarity
SOURCE="/Users/jack/tmp/2024-11-27"
DEST="/Users/jack/tmp/2024-11-27-back"

echo "\n================ 1. Initial Setup ================\n"
echo "Source: $SOURCE"
echo "Destination: $DEST"
echo "\nExecuting: rm -rf $DEST"
rm -rf $DEST

echo "\nChecking destination is empty..."
echo "Executing: ls -R $DEST"
ls -R $DEST

echo "\n================ 2. Dry Run ================\n"
echo "Executing dry run to preview changes..."
echo "Executing: go run cmd/main.go -config configs/test_config.yaml --dry-run"
go run cmd/main.go -config configs/test_config.yaml --dry-run

echo "\nCleaning destination folder..."
echo "Executing: rm -rf $DEST"
rm -rf $DEST

echo "\n================ 3. Full Copy ================\n"
echo "Performing full copy operation..."
echo "Executing: go run cmd/main.go -config configs/test_config.yaml"
go run cmd/main.go -config configs/test_config.yaml

echo "\nComparing source and destination sizes..."
echo "Executing: du -sh $SOURCE $DEST"
du -sh $SOURCE $DEST

echo "\n================ 4. Packers Subdirectory Copy ================\n"
echo "Removing Packers subdirectory from destination..."
echo "Executing: rm -rf $DEST/Packers"
rm -rf $DEST/Packers

echo "Copying Packers subdirectory..."
echo "Executing: go run cmd/main.go -config configs/test_config.yaml"
go run cmd/main.go -config configs/test_config.yaml

echo "\n================ 5. Version Management ================\n"
echo "Listing all versions..."
echo "Executing: go run cmd/main.go -config configs/test_config.yaml --list-versions"
go run cmd/main.go -config configs/test_config.yaml --list-versions

echo "\nShowing latest version..."
echo "Executing: go run cmd/main.go -config configs/test_config.yaml --latest-version"
go run cmd/main.go -config configs/test_config.yaml --latest-version

echo "\n================ 6. Final Size Comparison ================\n"
echo "Comparing final sizes of source and destination..."
echo "Executing: du -sh $SOURCE $DEST"
du -sh $SOURCE $DEST