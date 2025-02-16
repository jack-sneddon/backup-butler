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

## 5. File Validation and Comparison

### 5.1 Validation Strategy

```yaml
validation:
  # Default validation level
  level: "quick"      # quick, standard, deep
  
  # Performance tuning
  buffer_size: 32768         # bytes for partial content validation
  hash_algorithm: "sha256"   # md5, sha1, sha256
```

### 5.2 Validation Levels

#### Quick Validation

- **Purpose**: Fast detection of obvious changes
- **Checks**: File size, modification time (mtime), file mode/type
- **Performance**: ~0.1ms per file
- **Use Case**: Default for most files
- **Limitations**: Cannot detect content changes that preserve metadata

#### Standard Validation

- **Purpose**: Balance of performance and integrity
- **Checks**: Quick validation + partial content hash (first 32KB)
- **Performance**: ~1.9ms per file
- **Use Case**: Files failing quick validation
- **Limitations**: May miss changes in large files beyond first 32KB

#### Deep Validation

- **Purpose**: Complete integrity verification
- **Checks**: Quick validation + full file hash
- **Performance**: ~12s per GB
- **Use Case**: periodic full verification
- **Limitations**: Resource intensive, time-consuming

### 5.3 Validation Algorithm

```bash
For each file:
1. Check metadata (all levels):
   - Compare file sizes
   - Compare modification times (2s tolerance)
   - Return differ if any mismatch

2. Based on configured level:
   QUICK:
     Return match if metadata matches

   STANDARD:
     Calculate 32KB hash
     Return differ if hashes mismatch
     Return match if hashes match

   DEEP:
     Calculate full content hash
     Return differ if hashes mismatch
     Return match if hashes match

3. Record results:
   - Validation level used
   - Status (match/differ)
   - Time taken
```


### 5.4 Validation Output

Validation results show file status and validation level:
```bash
./photos/2024/
  = vacation/img001.jpg [100MB matched]      # Identical
  + vacation/img002.jpg [50MB new]           # To be copied
  * vacation/img003.jpg [75MB changed]       # Content differs
  - vacation/old/img004.jpg [25MB removed]   # Only in target
  ! vacation/corrupt.jpg [ERROR: read failed] # Error occurred

Validation Summary:
├── Quick Checks:   150 files
├── Standard Checks: 45 files
├── Deep Checks:     5 files
└── Total Time:    2m 15s
```

## 6. Command Interface

### 6.1 Global Commands

backup-butler provides these operations:

- **check**:   Compare source and target (see Section 5)
- **sync**:    Perform backup operation (see Section 9)
- **history**: View operation history (see Section 10)
- **version**: Show version information

Global options:

- --**config**:    Configuration file path
- --**log-level**: Log level (debug|info|warn|error)

### 6.2 Command Behavior

All commands follow these principles:

- Configuration validation (Section 7)
- Progress tracking (Section 8)
- Storage optimization (Section 9)
- Report generation (Section 10)

### 6.3 Operation Flow

1. **Check** Command:

- Validates configuration
- Scans directories
- Performs validation
- Shows projected changes
- Generates reports

2. **Sync** Command:

- Performs check steps
- Creates needed directories
- Copies/updates files
- Handles deletions
- Logs operations


3. **History** Command:

- Shows operation logs
- Lists available reports
- Provides operation summary

## 7. Configuration

### 7.1 Example Configuration

```yaml
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
  - "._*"

#storage types (optional)
storage:
  source:
    type: "ssd"              # hdd, ssd, network
    buffer_size: 262144      # Optional: 256KB for SSD
    max_threads: 16          # Optional: higher concurrency for SSD
  target:
    type: "network"          # Network storage (e.g., NAS)
    buffer_size: 1048576     # Optional: 1MB for network
    max_threads: 8           # Optional: limit network concurrency

# Validation settings as defined in section 5.1

# Device optimization (HDD focused)
storage:
  device_type: "hdd"  # currently only hdd supported
  max_threads: 4
  io_priority: "conservative"  # focused on reliability

# Deleted file handling
deleted_files:
  action: "report"           # report, delete
  report_format: "text"      # text only in current version

# Logging
logging:
  level: "info"  # debug, info, warn, error
```

## 7.2 Deleted File Handling

### 7.2.1 Action Modes

- **report** (safe-mode / default):
  - Generates detailed report of files only in target
  - Takes no action on files
  - Report includes file path, size, and last modification time
  
- **delete** (mirror):
  - Automatically removes files present only in target
  - Honors protected_paths configuration
  - Logs all deletions with timestamps
  
- **archive**:
  - Moves files to specified archive location
  - Preserves original directory structure
  - Creates timestamped archive directories
  

### 7.2.3 Protected Paths
- Patterns use same syntax as exclude patterns
- Supports glob patterns and directory wildcards
- Takes precedence over action mode
- Protection reason included in reports

### 7.2.4 Report Format
- **text**: Human-readable formatted text
- **csv**: Machine-parseable CSV format
- **html**: Web-viewable report with sorting/filtering

### 7.2.5 Example Config
```yaml
# Deleted file handling
deleted_files:
  action: "delete"           # will remove files not found in source but in target
  archive_location: ""       # Required if action = "archive"
  report_format: "text"      # text, csv, html
  protected_paths:          # Never delete these paths
    - "*.important"
    - "tax/*"
    - "**/*.original"
```


## 8. Progress and Reporting

### 8.1 Progress Model
Progress tracking occurs at two levels:

1. Directory Level
   - Files in current directory
   - Bytes processed
   - Completion percentage
   - Current operation status


2. Overall Progress
   - Total files processed
   - Total bytes complete
   - Elapsed time
   - Remaining work


### 8.2 Progress Display
Standard display format:

```bash
Processing: /Photos/Vacation2024
[================>    ] 78% (156/200 files)
Currently Processing:
  Vacation2024/Summer (156.2 MB)

Statistics:
├── Processed: 156 files (2.1 GB)
├── Remaining: 44 files (0.6 GB)
└── Total Time: 5m 32s
```

### 8.3 Implementation
Progress tracking is:

1. Event-driven: Updates on file operations
2. Directory-based: One directory at a time
3. Thread-safe: Mutex-protected updates
4. Resource-efficient: No background processes

## 9. Storage Device Protection

### 9.1 Directory-Based Processing

#### Core Principles

1. **Sequential Processing**
   - Process one directory completely before moving to next
   - Process parent directories before children
   - Complete directory tree before moving to next

2. **Performance Optimization**
   - Minimize disk head movement
   - Optimize for sequential access
   - Thread limiting based on storage type
   - Buffer sizes tuned to device type

3. **Error Handling**
   - Directory-level recovery points
   - Clear error reporting
   - Easy resume from last directory
   - Comprehensive error logs

### 9.2 HDD Requirements

#### Directory-Based Processing

- Process one directory at a time completely
- Directory ordering optimized for sequential access
- Complete current directory before moving to next
- Directory-level error handling and recovery

#### I/O Optimization

- Thread limiting (max 4 threads)
- Sequential access within directories
- Minimize head movement by completing directories
- Conservative I/O scheduling

#### Storage Types
- HDD (Hard Disk Drive):
  - Buffer Size: 32KB
  - Max Threads: 4
  - Focus: Sequential access, minimize head movement

- SSD (Solid State Drive):
  - Buffer Size: 256KB
  - Max Threads: 16
  - Focus: Parallel operations, larger buffer sizes

- Network Storage:
  - Buffer Size: 1MB
  - Max Threads: 8
  - Focus: Balance between throughput and connection limits

### 9.3 Performance Optimization

1. Group files by directory
   - Process all files in current directory
   - Minimize directory switching
   - Clear progress tracking per directory

2. Directory Processing Order
   - Process parent directories before children
   - Complete one directory tree before moving to next
   - Optimize for sequential disk access

3. Error Handling
   - Directory-level recovery points
   - Clear reporting of directory status
   - Easy resume from last directory

### 9.4 Storage Configuration Defaults

Default Behavior:

If storage type is not specified, assumes "hdd"
If buffer_size is not specified, uses type-specific default
If max_threads is not specified, uses type-specific default
When source and target have different types:

Uses the more conservative thread count
Adjusts buffer sizes independently for read/write

When storage configuration is not fully specified, the following defaults are applied:

```yaml
# Full configuration with defaults shown
storage:
  source:
    type: "hdd"           # Default if not specified
    buffer_size: 32768    # 32KB for HDD
    max_threads: 4        # Conservative default for HDD
  target:
    type: "hdd"
    buffer_size: 32768
    max_threads: 4

# Optimized defaults by storage type:
HDD:
  buffer_size: 32768     # 32KB - optimized for mechanical drives
  max_threads: 4         # Limited to prevent excessive seeking

SSD:
  buffer_size: 262144    # 256KB - leverages faster random access
  max_threads: 16        # Higher concurrency for better performance

Network:
  buffer_size: 1048576   # 1MB - optimized for network transfers
  max_threads: 8         # Balanced for network connections
  ```

## 10. Reports and Logs

### 10.1 Directory Structure

```bash
<target>/.backup-butler/
├── logs/
│   ├── current.log        # Operation metrics for current run
│   └── previous.log       # Last operation metrics
└── reports/
    └── deleted_files.txt  # List of files only in target
```

### 10.2 Operation Logs

Purpose: Track performance and execution metrics

Content:

```bash
Operation Start: 2025-02-06T14:30:22Z
Operation End:   2025-02-06T15:45:33Z
Total Files Processed: 207,163
Total Data: 1.2TB
Validation Summary:
  - Quick: 180,245
  - Standard: 25,890
  - Deep: 1,028
Errors: None
```

### 10.3 File Status Reports
Purpose: Provide actionable information about file discrepancies

Content:

```bash
Files Present Only in Target (Run: 2025-02-06T14:30:22Z)
=====================================================
/Photos/2023/
  - vacation/IMG_1234.jpg (5.2MB, Last Modified: 2023-06-15)
  - vacation/IMG_1235.jpg (4.8MB, Last Modified: 2023-06-15)
/Photos/2024/
  - summer/DSC_9876.jpg (8.1MB, Last Modified: 2024-07-01)

Summary:
- Total files: 3
- Total size: 18.1MB
- Affected directories: 2

Actions:
1. Review these files as they no longer exist in source
2. Use 'backup-butler sync' with deleted_files.action = "delete" to remove
3. Or manually preserve/remove these files as needed
```

### 10.4 File Status Report Configuration

```yaml
deleted_files:
  action: "report"           # report, delete
  report_format: "text"      # text only in current version
```

### 10.5 Retention

- Operation Logs:
  - Keep only current and previous run
  - Automatically rotated on new run
- File Status Reports:
  - New report generated each run
  - Previous report overwritten
  - User should save reports manually if needed

### 10.6 Report Generation

- File status report generated when:
  1. Files exist only in target
  2. Action is set to "report"
- Report location provided in operation summary
- Report remains until next run

## 11. Testing Strategy

### 11.1 Test Categories

1. Unit tests for core functionality
2. Integration tests for command workflow
3. Performance benchmarks
4. Cross-platform compatibility
5. Recovery and resume scenarios

### 11.2 Test Framework

```bash
#!/bin/bash
test_validation() {
  # Test each validation level
  test_quick_validation
  test_standard_validation
  test_deep_validation
  test_hybrid_validation
}

test_device_protection() {
  # Test storage optimizations
  test_hdd_optimization
  test_ssd_optimization
}
```

## 12. Future Extensions

### 12.1 Cloud Storage Support

```yaml
target:
  type: "s3"
  bucket: "media-backup"
  region: "us-west-2"
```

### 12.2 Database Integration

- SQLite for local storage
- PostgreSQL for scalability
- Performance metrics
- Validation history

## 13. Success Criteria

- Zero data loss
- Accurate reporting
- Resumable operations
- Performance targets met
- Device protection verified
- Clear error handling
