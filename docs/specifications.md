# Backup Butler Specification

## 1. Overview and Purpose

Backup Butler is a command-line utility designed for reliable media backup with high data integrity validation. It provides users with confidence in backup completeness through thorough verification and detailed reporting.

## 2. Objectives

- MUST ensure data integrity through configurable validation levels
- MUST support large media collections (200K+ files)
- MUST prevent storage device damage through intelligent I/O
- MUST provide clear progress and validation reporting
- MUST support resumable operations
- SHOULD minimize unnecessary file operations
- SHOULD optimize for different storage types
- MAY support cloud/NAS destinations

## 3. Terminology

### 3.1 RFC Keywords

- MUST/REQUIRED - Absolute requirement
- SHOULD/RECOMMENDED - May be valid reasons to ignore
- MAY/OPTIONAL - Truly optional

### 3.2 File Status Indicators

- `=` - File identical in source/destination
- `-` - File only in destination
- `+` - File only in source
- `*` - File differs between source/destination
- `!` - Error reading/hashing file

## 4. System Requirements

### 4.1 File Profile Support

MUST handle:

```sh
Total files: 207,163
- Small (<2MB): 37,352
- Medium (2MB-6MB): 105,848
- Large (6MB-30MB): 61,686
- Extra Large (>30MB): 2,277
```

### 4.2 Platform Support

- MUST support macOS
- SHOULD support Windows
- MUST handle cross-platform paths

## 5. Configuration

### 5.1 Example Configuration

```yaml
# Backup Butler Configuration

# Core paths
source: "/Users/media/photos"
target: "/Volumes/backup/photos"

# Directories to process
folders:
  - "2023"
  - "2024/vacation"

# File patterns to exclude
exclude:
  - "*.tmp"
  - ".DS_Store"
  - ".Trashes"
  - "._*" # Mac resource files

# Comparison settings
comparison:
  algorithm: "sha256" # md5, sha1, sha256
  level: "standard" # quick, standard, deep
  buffer_size: 32768 # 32KB

# Device optimization
storage:
  device_type: "hdd" # hdd, ssd
  max_threads: 4 # conservative for HDDs
  io_priority: "balanced" # balanced, performance, conservative
  sequential_threshold: 100 # files before switching to sequential mode

# Progress and checkpointing
auto_save:
  enabled: true
  thresholds:
    files: 100
    data: "1GB"
    time: "5m"

# Platform specific
platform:
  paths:
    windows_separator: "\\"
    case_sensitive: false # Windows handling
  encoding: "utf-8"

# Performance tuning
performance:
  read_buffer: "32MB"
  write_buffer: "32MB"
  batch_size: 100
  max_retries: 3

# Reporting
reporting:
  default_format: "text" # text, csv, html
  colors: true # Terminal colors
  progress_interval: "1s" # Progress update frequency

# Version management
versions:
  location: ".backup-butler"
  max_versions: 10
  cleanup:
    enabled: true
    keep_last: 5

# logging level
logging:
  level: "debug" # debug, info, warn, error
```

### 5.2 Configuration Field Definitions

| Field                              | Description                            | Required | Default        |
| ---------------------------------- | -------------------------------------- | -------- | -------------- |
| `source`                           | Source directory path                  | Yes      | -              |
| `target`                           | Backup destination path                | Yes      | -              |
| `folders`                          | List of subdirectories to backup       | No       | All folders    |
| `exclude`                          | File patterns to ignore                | No       | Empty          |
| `comparison.algorithm`             | Hash algorithm (md5/sha1/sha256)       | No       | sha256         |
| `comparison.level`                 | Validation depth (quick/standard/deep) | No       | standard       |
| `comparison.buffer_size`           | Read/write buffer size in bytes        | No       | 32768          |
| `storage.device_type`              | Storage device type (hdd/ssd)          | Yes      | hdd            |
| `storage.max_threads`              | Maximum concurrent operations          | No       | 4              |
| `storage.io_priority`              | I/O scheduling priority                | No       | balanced       |
| `storage.sequential_threshold`     | Files before sequential mode           | No       | 100            |
| `auto_save.enabled`                | Enable checkpointing                   | No       | true           |
| `auto_save.thresholds.files`       | Files between saves                    | No       | 100            |
| `auto_save.thresholds.data`        | Data volume between saves              | No       | 1GB            |
| `auto_save.thresholds.time`        | Time between saves                     | No       | 5m             |
| `platform.paths.windows_separator` | Windows path separator                 | No       | \\             |
| `platform.paths.case_sensitive`    | Case-sensitive paths                   | No       | false          |
| `platform.encoding`                | File name encoding                     | No       | utf-8          |
| `performance.read_buffer`          | Read buffer size                       | No       | 32MB           |
| `performance.write_buffer`         | Write buffer size                      | No       | 32MB           |
| `performance.batch_size`           | Files per batch                        | No       | 100            |
| `performance.max_retries`          | Operation retry attempts               | No       | 3              |
| `reporting.default_format`         | Report format                          | No       | text           |
| `reporting.colors`                 | Use terminal colors                    | No       | true           |
| `reporting.progress_interval`      | Progress update frequency              | No       | 1s             |
| `versions.location`                | Version storage directory              | No       | .backup-butler |
| `versions.max_versions`            | Maximum versions to keep               | No       | 10             |
| `versions.cleanup.enabled`         | Enable version cleanup                 | No       | true           |
| `versions.cleanup.keep_last`       | Versions to retain                     | No       | 5              |

# File Validation Strategy

## Overview

The file validation system ensures data integrity between source and target locations using a tiered approach that balances performance with confidence.

## Configuration

```yaml
validation:
  # Default validation level for all files
  default_level: "quick"      # quick, standard, deep
  
  # Validation level when metadata differs
  on_mismatch: "standard"     # none, standard, deep
  
  # Critical paths requiring stronger validation
  critical_paths:
    - path: "/financial/**"
      level: "deep"
    - path: "/photos/**"
      level: "standard"
  
  # Scheduled deep validation
  scheduled_deep:
    enabled: true
    frequency: "monthly"      # never, daily, weekly, monthly, yearly
    last_run: "2024-01-01"   # ISO date
    paths:
      - "/important/**"
    exclude:
      - "*.tmp"

  # Performance tuning
  buffer_size: 32768         # bytes for partial content validation
  hash_algorithm: "sha256"   # md5, sha1, sha256
```

## Validation Levels

### Quick Validation

- **Purpose**: Fast detection of obvious changes
- **Checks**:
  - File size
  - Modification time (mtime)
  - File mode/type
- **Performance**: ~0.1ms per file
- **Use Case**: Default for all files
- **Limitations**:
  - Misses content changes that preserve metadata
  - Can miss changes due to timestamp precision

### Standard Validation

- **Purpose**: Balance of performance and integrity
- **Checks**:
  - All Quick validation checks
  - Partial content hash (first 32KB)
- **Performance**: ~1.9ms per file
- **Use Case**:
  - Files that fail Quick validation
  - Critical paths requiring content verification
  - When metadata timestamps are unreliable

### Deep Validation

- **Purpose**: Complete integrity verification
- **Checks**:
  - All Quick validation checks
  - Full file content hash
- **Performance**: Proportional to file size (~12s per GB)
- **Use Case**:
  - Critical financial/legal documents
  - Periodic full integrity checks
  - Initial backup verification

## Validation Algorithm

```yaml
For each file:
1. Determine validation level:
   - Check scheduled deep validation
   - Check critical path configuration
   - Use default level if no special cases

2. Perform validation:
   if level == QUICK:
     Check metadata
     if metadata differs && on_mismatch == STANDARD:
       Perform standard validation
     if metadata differs && on_mismatch == DEEP:
       Perform deep validation

   if level == STANDARD:
     Check metadata
     Check first 32KB hash

   if level == DEEP:
     Check metadata
     Check full file hash

3. Record validation results:
   - Validation level used
   - Time taken
   - Result status
   - Bytes read
```

## Metadata Preservation

When copying files, the following metadata must be preserved:

- Modification time (mtime)
- File mode/permissions
- File type information

Platform-specific considerations:

- Windows: NTFS timestamps use 100ns precision
- Unix: Timestamp precision varies by filesystem
- Cloud: May have limited metadata support

## Performance Characteristics

Theoretical performance for 100MB file:

- Quick: ~0.1ms
- Standard: ~1.9ms
- Deep: ~1.2s

For 200,000 files (all 100MB):

- Quick: ~20s
- Standard: ~6.3m
- Deep: ~66h

## Implementation Guidelines

1. Use buffered I/O for content reads
2. Implement filesystem-appropriate caching
3. Group files by validation level for efficiency
4. Process files in directory order for HDD optimization
5. Enable parallel validation for SSD targets

## Critical File Marking

Files can be marked as critical via:

1. Path patterns in config
2. Directory markers
3. Explicit file lists

## Dry Run Behavior

- Uses same validation logic
- Reports potential changes without copying
- Shows validation level used for each file
- Estimates time for full sync

## Performance Optimization

1. Parallel validation for SSDs
2. Sequential validation for HDDs
3. Batch processing for small files
4. Caching of previous validation results

## 6. Command Interface

### 6.1 Check Mode

```bash
backup-butler check [options]
  --config string    Configuration file
  --source string   Source directory
  --target string   Target directory
  --output string   Report format (text|csv|html)
  --level string    Check level (quick|standard|deep)
```

### 6.2 Backup Mode

```bash
backup-butler backup [options]
  --config string     Configuration file
  --resume           Resume from last checkpoint
  --force            Override safety checks
```

### 6.3 Version Mode

```bash
backup-butler version [command]
Commands:
  list     Show version history
  show     Display version details
  clean    Clean old versions
  size     Show version storage usage
```

## 7. Progress Display

### 7.1 Real-time Display

MUST show:

```sh
Directory: /Photos/Vacation2024
[================>    ] 78% (156/200 files)
Currently Processing:
  IMG_4567.jpg (156.2 MB)
  Speed: 45.3 MB/s
  ETA: 2m 15s

Statistics:
├── Processed: 156 files (2.1 GB)
├── Remaining: 44 files (0.6 GB)
└── Total Time: 5m 32s
```

### 7.2 Check Report Output

```sh
# Text format
./photos/2024/
  = vacation/img001.jpg      [100MB matched]
  + vacation/img002.jpg      [50MB new]
  * vacation/img003.jpg      [75MB changed]
  - vacation/old/img004.jpg  [25MB removed]
  ! vacation/corrupt.jpg     [ERROR: read failed]
```

## 8. Storage Device Protection

### 8.1 HDD Protection Requirements

MUST implement:

- Thread limiting
- I/O scheduling
- Sequential access optimization
- Head movement minimization

### 8.2 Device-Specific Settings

```yaml
storage:
  device:
    type: "hdd"
    max_concurrent_ops: 4
    read_buffer: "32MB"
    write_buffer: "32MB"
  optimization:
    mode: "sequential"
    batch_size: 100
```

## 9. Metadata Storage

### 9.1 Structure

MUST store in `<target>/.backup-butler/`:

- Version history
- File metadata
- Operation logs
- Resume state

### 9.2 Database Support

MAY implement:

- SQLite for local storage
- PostgreSQL for scalability
- Version/metadata tables
- Performance metrics

## 10. Error Recovery

### 10.1 Auto-save Points

MUST save state after:

- Configured number of files processed
- Specified data volume processed
- Time interval elapsed

### 10.2 Resume State Format

```yaml
resume:
  version_id: "20250125-143022"
  completed_files: 156
  total_files: 200
  last_file: "IMG_4567.jpg"
  checkpoints:
    - timestamp: "2025-01-25T14:32:15Z"
      files_done: 150
```

## 11. Testing Strategy

### 11.1 CLI Testing

MUST implement bash-based testing:

```bash
#!/bin/bash
setup_test_files() {
  mkdir -p test/source test/target
  # Create test files
}

test_backup() {
  backup-butler backup --config test.yaml
  verify_results
}
```

## 12. Logging

### 12.1 library

Use of Uber's zap logger

### 12.2 Logging levels

Log Levels (lowest to highest):

- DEBUG: Verbose information for debugging issues
- INFO: General operational events
- WARN: Potentially harmful situations
- ERROR: Error events that might still allow the application to continue
- FATAL: Very severe error events that will lead to application termination

### 12.3 setting the logging level

- Default (if not provided) is "ERROR"

- Setting in configuration file

  - example:
        ```yaml
        logging:
        level: "debug"  # debug, info, warn, error
        ```

- Flag in terminal command
  - example: $ backup-butler sync -c config.yaml --log-level debug

### Order of precidence

Order:

1. terminal flag
2. configuration file
3. default

## 13. Future Extensions

### 13.1 Cloud Storage Support

```yaml
target:
  # AWS S3
  type: "s3"
  bucket: "media-backup"
  region: "us-west-2"

  # NAS
  type: "nas"
  protocol: "smb"
  path: "//server/share"
```

### 13.2 Cloud Requirements

MUST implement:

- Concurrent uploads
- Chunked transfers
- Resume capability
- Bandwidth control

## 13. Success Criteria

MUST meet:

- Zero data loss
- Accurate reporting
- Resumable operations
- Performance targets
- Device protection
- Clear error handling
