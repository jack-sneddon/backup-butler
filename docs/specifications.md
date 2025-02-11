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

## 5. File Validation and Comparison

### 5.1 Validation Strategy

Each validation level provides a different balance of performance vs thoroughness:

```yaml
validation:
  # Default validation level
  level: "quick"      # quick, standard, deep
  buffer_size: 32768         # bytes for partial content validation
  hash_algorithm: "sha256"   # md5, sha1, sha256

### 5. Validation Levels

Validation progresses from lightweight to thorough checks, balancing performance against accuracy:

Configuration will define what level to go for, but it will check the lower levels first.  If they pass, it will /onlate to the next level until it reaches what was defined in configuration.

Quick:
- Metadata comparison only (size, mtime)
- Performance: ~0.1ms per file
- Use case: Detecting obvious changes

Standard:
- Metadata + partial content hash (first 32KB)
- Performance: ~1.9ms per file
- Use case: Balanced validation for most files

Deep:
- Metadata + full content hash
- Performance: ~12s per GB
- Use case: Complete verification when needed

### 5.2 Validation Algorithm

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

### 5.3 Validation Algorithm

Quick:

- Metadata mismatch → fail
- Metadata match → pass
Standard:
- Metadata mismatch → fail
- Metadata match + content mismatch → fail
- Metadata match + content match → pass
Deep:
- Metadata mismatch → fail
- Metadata match + partial content mismatch → fail
- Metadata match + partial content match + full content mismatch → fail
- All checks pass → pass

For each file:

1. Check metadata (always):
   - Compare file sizes
   - Compare modification times (2s tolerance)
   - If different: return StatusDiffer

2. Based on configured level:
   QUICK:
     - Return StatusMatch (metadata matches)

   STANDARD:
     - Calculate hash of first 32KB
     - Return StatusDiffer if hashes different
     - Return StatusMatch if hashes match

   DEEP:
     - Calculate full content hash
     - Return StatusDiffer if hashes different
     - Return StatusMatch if hashes match

## 6. Command Interface

### 6.1 Global Commands

```bash
backup-butler [command] [options]

Commands:
  check     Compare source and target without copying
  sync      Synchronize source to target with copy
  history   Show operation logs and reports
  version   Display program version information

Global Options:
  --config string     Configuration file path
  --log-level string Log level (debug|info|warn|error)
```

### 6.2 Check Command

Purpose: Preview all changes that would be made during a sync operation

```bash
backup-butler check [options]
  --level string    Validation level (quick|standard|deep)
```

Output Format:

```bash
Backup Butler v0.1.0
====================

Scan Results:
├── Locations
│   ├── Source: /Photos/2024
│   └── Target: /Volumes/backup/Photos/2024
├── Summary
│   ├── Directories: 5
│   ├── Files: 207,163
│   ├── Total Size: 1.2TB
└── File Status
    = photos/img001.jpg [quick]         # File matches
    + photos/img002.jpg [quick]         # Will be copied
    * photos/img003.jpg [standard]      # Will be updated
    - photos/old/img004.jpg             # Only in target
    ! photos/corrupt.jpg                # Error reading file

Results Summary:
├── Matched:  180,245 files
├── New:      25,890 files
├── Missing:  1,028 files
├── Modified: 0 files
└── Errors:   0 files

Actions to be taken:
├── Files to copy: 25,890 (156.2 GB)
├── Files to update: 0
└── Files only in target: 1,028 (review deleted_files.txt)

Validation to be performed:
├── Quick:    180,245 files
├── Standard: 25,890 files
└── Deep:     1,028 files

Estimated time: 2h 15m
```

Key Features:
- Shows status of ALL files (not just changes)
- Indicates validation level used for each file
- Provides clear summary of actions to be taken
- Same format used in both 'check' and 'sync'
- Identifies files that only exist in target
- Groups files by status for easy review
- Estimates operation time

The check command:
1. Performs all validation but makes no changes
2. Creates the same reports as 'sync' would
3. Shows validation levels that would be used
4. Gives users chance to review before proceeding


### 6.3 Sync Command
Purpose: Perform actual backup operation

```bash
backup-butler sync [options]
  --level string    Validation level (quick|standard|deep)
```

Operation Steps:
1. Directory Analysis
   - Scan source and target
   - Build operation plan
   - Show initial summary

2. Directory Processing
   ```
   Processing: /Photos/Vacation2024
   [================>    ] 78% (156/200 files)
   Currently Processing:
     Vacation2024/Summer (156.2 MB)
     ETA: 2m 15s
   ```

3. File Operations:
   - Create missing directories
   - Copy new files
   - Validate existing files
   - Handle deleted files per config
   - Process one directory at a time

4. Progress Display:
   ```
   Statistics:
   ├── Processed: 156 files (2.1 GB)
   ├── Remaining: 44 files (0.6 GB)
   └── Total Time: 5m 32s
   ```

5. Completion Summary:
   ```
   Operation Complete
   =================
   Start Time: 2025-02-06 14:30:22
   End Time:   2025-02-06 14:35:54
   Duration:   5m 32s

   Results:
   ├── Directories Processed: 5
   ├── Files Copied: 25
   ├── Files Validated: 175
   ├── Data Transferred: 2.7 GB
   └── Average Speed: 8.3 MB/s

   Validation Summary:
   ├── Quick: 150 files
   ├── Standard: 45 files
   └── Deep: 5 files

   Reports Generated:
   └── /Volumes/backup/Photos/.backup-butler/reports/deleted_files.txt
   ```

Behavioral Notes:
- Same validation and checks as 'check' command
- Actually performs file operations
- Processes directory by directory
- Creates operation logs
- Generates status reports
- Handles errors gracefully


### 6.4 Version Command

```bash
backup-butler version
Backup Butler v0.1.0
Git commit: abc123
Built: 2025-02-06
```

### 6.5 History Command

```bash
backup-butler history
Last Operation: 2025-02-06 14:30:22
Status: Complete
Results:
├── Files Processed: 207,163
├── Data Transferred: 1.2TB
└── Duration: 1h 15m

Reports Available:
└── /Volumes/backup/Photos/.backup-butler/reports/deleted_files.txt
```

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

# Add to Section 7.1 Example Configuration

```yaml
# Deleted file handling
deleted_files:
  action: "report"           # report, delete, archive
  archive_location: ""       # Required if action = "archive"
  report_format: "text"      # text, csv, html
  protected_paths:          # Never delete these paths
    - "*.important"
    - "tax/*"
    - "**/*.original"
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

## 8. Progress and Reporting

### 8.1 Real-time Display

```sh
Directory: /Photos/Vacation2024
[================>    ] 78% (156/200 files)
Currently Processing:
  Vacation2024/Summer (156.2 MB)
  ETA: 2m 15s

Statistics:
├── Processed: 156 files (2.1 GB)
├── Remaining: 44 files (0.6 GB)
└── Total Time: 5m 32s
```

### 8.2 Directory-Based Processing

- Files are processed one directory at a time
- For each directory:
  1. Create directory structure if needed
  2. Process all files in current directory
  3. Move to next directory
- Benefits:
  - Optimized for HDD sequential access
  - Clear progress reporting
  - Better error handling (directory level)
  - Simpler resume points

### 8.3 Progress Calculation

- Progress based on total files and directories
- ETA calculated using:
  - Completed files count
  - Total elapsed time
  - Remaining files count
- Progress updated:
  - After each file completion
  - When starting new directory
  - At regular intervals (every second)

### 8.4 Validation Report

```sh
./photos/2024/
  = vacation/img001.jpg      [100MB matched]
  + vacation/img002.jpg      [50MB new]
  * vacation/img003.jpg      [75MB changed]
  - vacation/old/img004.jpg  [25MB removed]
  ! vacation/corrupt.jpg     [ERROR: read failed]

Validation Summary:
├── Quick Checks:   150 files
├── Standard Checks: 45 files
├── Deep Checks:     5 files
└── Total Time:      2m 15s
```

## 9. Storage Device Protection

### 9.1 HDD Requirements

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

### 9.2 Performance Optimization

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
