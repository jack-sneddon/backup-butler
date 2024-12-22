#!/bin/bash

# Setup script for Clean Architecture migration
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Helper function for creating directories
create_dir() {
    if [ ! -d "$1" ]; then
        mkdir -p "$1"
        echo -e "${GREEN}Created directory:${NC} $1"
    else
        echo -e "${YELLOW}Directory already exists:${NC} $1"
    fi
}

# Helper function for creating placeholder files
create_file() {
    if [ ! -f "$1" ]; then
        touch "$1"
        echo -e "${GREEN}Created file:${NC} $1"
        
        # Add package declaration if it's a Go file
        if [[ "$1" == *.go ]]; then
            package_name=$(basename $(dirname "$1"))
            echo "package ${package_name}" > "$1"
        fi
    else
        echo -e "${YELLOW}File already exists:${NC} $1"
    fi
}

echo -e "${GREEN}Setting up Clean Architecture directory structure...${NC}"

# Create domain layer directories
create_dir "internal/domain/backup"
create_dir "internal/domain/versioning"

# Create domain layer files
create_file "internal/domain/backup/service.go"
create_file "internal/domain/backup/types.go"
create_file "internal/domain/backup/ports.go"
create_file "internal/domain/versioning/manager.go"
create_file "internal/domain/versioning/types.go"
create_file "internal/domain/versioning/ports.go"

# Create core layer directories
create_dir "internal/core/backup"
create_dir "internal/core/storage"
create_dir "internal/core/monitoring"
create_dir "internal/core/config"
create_dir "internal/core/worker"

# Create core layer files
create_file "internal/core/backup/service.go"
create_file "internal/core/backup/operations.go"
create_file "internal/core/backup/validator.go"
create_file "internal/core/storage/manager.go"
create_file "internal/core/storage/copy.go"
create_file "internal/core/storage/checksum.go"
create_file "internal/core/storage/errors.go"
create_file "internal/core/monitoring/logger.go"
create_file "internal/core/monitoring/metrics.go"
create_file "internal/core/monitoring/types.go"
create_file "internal/core/config/loader.go"
create_file "internal/core/config/validator.go"
create_file "internal/core/config/types.go"
create_file "internal/core/worker/pool.go"
create_file "internal/core/worker/task.go"

# Create adapters layer directories
create_dir "internal/adapters/storage/filesystem"
create_dir "internal/adapters/storage/mock"
create_dir "internal/adapters/metrics/collector"
create_dir "internal/adapters/metrics/publishers/console"
create_dir "internal/adapters/metrics/publishers/version"
create_dir "internal/adapters/logger/filesystem"

# Create adapters layer files
create_file "internal/adapters/storage/filesystem/adapter.go"
create_file "internal/adapters/storage/filesystem/helper.go"
create_file "internal/adapters/storage/mock/adapter.go"
create_file "internal/adapters/metrics/collector/adapter.go"
create_file "internal/adapters/metrics/publishers/console/adapter.go"
create_file "internal/adapters/metrics/publishers/version/adapter.go"
create_file "internal/adapters/logger/filesystem/adapter.go"

# Create UI layer directories
create_dir "internal/ui/shared/viewmodels"
create_dir "internal/ui/shared/controllers"
create_dir "internal/ui/shared/events"
create_dir "internal/ui/cli/commands"
create_dir "internal/ui/cli/formatter"
create_dir "internal/ui/gui/views"
create_dir "internal/ui/gui/widgets"
create_dir "internal/ui/gui/state"

# Create UI layer files
create_file "internal/ui/cli/commands/root.go"
create_file "internal/ui/cli/commands/backup.go"
create_file "internal/ui/cli/commands/version.go"
create_file "internal/ui/cli/formatter/progress.go"
create_file "internal/ui/cli/formatter/output.go"

echo -e "${GREEN}Directory structure created successfully!${NC}"
echo -e "\nNext steps:"
echo -e "1. Copy domain layer interfaces and types"
echo -e "2. Implement core layer services"
echo -e "3. Create adapter implementations"
echo -e "4. Update main.go to use new structure"

# Create .gitkeep files for empty directories
find . -type d -empty -exec touch {}/.gitkeep \;