# Backup Butler 📂

A sophisticated backup utility providing intelligent, incremental file backup with versioning and verification. Built using Clean Architecture principles to support both CLI and GUI interfaces.

[![Go Version](https://img.shields.io/github/go-mod/go-version/jack-sneddon/backup-butler)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

## ✨ Features

### Core Functionality

- 🔄 **Intelligent Copying**: Only copies changed files, using size and checksum verification
- 📚 **Version Control**: Maintains detailed history of all backup operations
- 📊 **Real-time Progress**: Shows backup progress, speed, and statistics
- 🔐 **Data Integrity**: Verifies all copies with checksums
- ⚡ **Concurrent Operations**: Multi-threaded file operations for optimal performance

### Example Configuration

```yaml
# Basic directory settings
source_directory: "/path/to/source"
target_directory: "/path/to/target"
folders_to_backup:
  - "Folder1"
  - "Folder2"

# File comparison settings
deep_duplicate_check: true # Checksum verification
checksum_algorithm: "sha256" # Integrity verification

# Performance settings
concurrency: 2 # Concurrent file operations
buffer_size: 32768 # Copy buffer size (32KB)
retry_attempts: 3 # Retry count for failures
retry_delay: "1s" # Delay between retries

# Optional patterns to exclude
exclude_patterns:
  - "*.tmp"
  - ".DS_Store"
  - "Thumbs.db"
```

### 🚀 Usage

#### Basic Operations

```bash
# Perform backup
backup-butler -config config.yaml

# Dry run (simulate backup)
backup-butler -config config.yaml --dry-run

# List backup versions
backup-butler -config config.yaml --list-versions

# Show specific version details
backup-butler -config config.yaml --show-version <version-id>
```

### 🏗️ Project Structure

```
.
├── cmd/                        # Entry points
│   ├── cli/                    # CLI interface
│   └── gui/                    # GUI interface (future)
├── internal/                   # Internal packages
│   ├── domain/                 # Core business logic
│   ├── core/                   # Use cases
│   ├── adapters/               # Interface implementations
│   └── ui/                     # User interface components
├── pkg/                        # Public packages
├── configs/                    # Configuration files
└── tests/                      # Integration tests
```

### 📊 Current Status

- ✅ Core backup functionality
- ✅ CLI interface
- ✅ Version tracking
- ✅ Concurrent operations
- 🔄 Clean Architecture migration
- 📋 GUI interface (planned)

### 🚀 Future Enhancements

- 🖥️ GUI Interface
- ☁️ Remote storage support (S3, Google Drive)
- 🔒 Encryption
- 🔄 Cloud configuration sync
- 📅 Backup retention policies
- 🔧 Pre/post backup hooks
- 📬 Notification system
