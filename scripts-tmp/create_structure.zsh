#!/bin/zsh

# Define the base directory
BASE_DIR="backup-butler"

# Create the directory structure
mkdir -p $BASE_DIR/cmd
mkdir -p $BASE_DIR/internal/config
mkdir -p $BASE_DIR/internal/backup
mkdir -p $BASE_DIR/internal/storage
mkdir -p $BASE_DIR/internal/version
mkdir -p $BASE_DIR/internal/display
mkdir -p $BASE_DIR/internal/worker
mkdir -p $BASE_DIR/scripts

# Create empty source files
touch $BASE_DIR/cmd/main.go

touch $BASE_DIR/internal/config/loader.go
touch $BASE_DIR/internal/config/types.go

touch $BASE_DIR/internal/backup/service.go
touch $BASE_DIR/internal/backup/scanner.go
touch $BASE_DIR/internal/backup/copier.go
touch $BASE_DIR/internal/backup/validator.go

touch $BASE_DIR/internal/storage/copy.go
touch $BASE_DIR/internal/storage/checksum.go
touch $BASE_DIR/internal/storage/metadata.go

touch $BASE_DIR/internal/version/manager.go
touch $BASE_DIR/internal/version/types.go

touch $BASE_DIR/internal/display/progress.go
touch $BASE_DIR/internal/display/formatter.go

touch $BASE_DIR/internal/worker/pool.go
touch $BASE_DIR/internal/worker/task.go

# Print success message
echo "Folder and file structure for '$BASE_DIR' created successfully."
