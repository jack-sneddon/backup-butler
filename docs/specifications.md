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
  default_level: "quick"      # quick, standard, deep
  
  # Validation level when metadata differs
  on_mismatch: "standard"     # none, standard, deep
  
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
1. Determine validation level based on:
   - Scheduled deep validation status
   - Default level setting

2. Perform validation:
   QUICK:
     Check metadata
     If differs AND on_mismatch = STANDARD:
       Perform standard validation
     If differs AND on_mismatch = DEEP:
       Perform deep validation

   STANDARD:
     Check metadata
     Check first 32KB hash

   DEEP:
     Check metadata
     Check full file hash

3. Record results:
   - Validation level used
   - Time taken
   - Status
   - Bytes processed
```

## 6. Command Interface

### 6.1 Global Commands

```bash
backup-butler [command] [options]

Commands:
  check     Compare source and target without copying
  sync      Synchronize source to target (with copy)
  backup    Create/update backup with resume capability
  version   Manage versions and history

Global Options:
  --config string     Configuration file path
  --dry-run          Show what would happen without changes
  --log-level string Log level (debug|info|warn|error)
```

### 6.2 Check Command

```bash
backup-butler check [options]
  --level string    Validation level (quick|standard|deep)
  --output string   Report format (text|csv|html)
```

### 6.3 Sync/Backup Commands

```bash
backup-butler sync [options]
backup-butler backup [options]
  --resume          Resume from last checkpoint
  --force          Override safety checks
```

### 6.4 Version Command

```bash
backup-butler version [command]
Commands:
  list     Show version history
  show     Display version details
  clean    Clean old versions
  size     Show version storage usage
```

### 6.5 Dry Run Behavior

The --dry-run flag:

- Shows changes without modifying files
- Performs all configured validation checks
- Reports validation levels used
- Estimates time and I/O impact
- Produces detailed report

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

# Device optimization
storage:
  device_type: "hdd"  # hdd, ssd
  max_threads: 4
  io_priority: "balanced"  # balanced, performance, conservative
  sequential_threshold: 100

# Progress tracking
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
    case_sensitive: false
  encoding: "utf-8"

# Logging
logging:
  level: "info"  # debug, info, warn, error
```

## 8. Progress and Reporting

### 8.1 Real-time Display

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

### 8.2 Validation Report

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

- Thread limiting
- I/O scheduling
- Sequential access optimization
- Head movement minimization

### 9.2 Performance Optimization

1. Group files by validation level
2. Process in directory order
3. Implement appropriate caching
4. Enable parallel validation for SSDs
5. Batch small files

## 10. State Management

### 10.1 Metadata Storage

Store in `<target>/.backup-butler/`:

- Version history
- File metadata
- Operation logs
- Resume state
- Validation history

### 10.2 Resume State

```yaml
resume:
  version_id: "20250125-143022"
  completed_files: 156
  total_files: 200
  last_file: "IMG_4567.jpg"
  validation_state:
    quick_complete: 140
    standard_complete: 12
    deep_complete: 4
```

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
