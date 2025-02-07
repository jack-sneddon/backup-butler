# Backup Butler: User Scenarios

## Initial Backup Scenarios

### First Time Backup
```bash
GIVEN: A source directory containing files and subdirectories
AND:   An empty target directory
WHEN:  User runs 'backup-butler sync'
THEN:  All files are copied to target
AND:   Directory structure is preserved
AND:   File metadata is preserved
AND:   Progress is displayed during operation
AND:   Summary report shows all files as new
```

### Backup with New Files

```bash
GIVEN: A previous backup exists in target directory
AND:   Source directory contains new files
WHEN:  User runs 'backup-butler sync'
THEN:  Only new files are copied
AND:   Existing files are validated
AND:   Summary report shows:
      - Count of new files copied
      - Count of existing files validated
AND:   Operation completes faster than full backup
```

## Deleted Source File Scenarios

### Backup with Deleted Source Files - Safe Mode (Default Behavior)

```bash
GIVEN: A previous backup exists in target directory
AND:   Files have been deleted from source
WHEN:  User runs 'backup-butler sync'
THEN:  Extra files in target are identified
AND:   A report is generated listing files present only in target
AND:   No files are automatically deleted
AND:   Summary report shows count of files present only in target
         - Count of files present only in target
         - Location of detailed report
```

### Backup with Deleted Files - Mirror Mode

```bash
GIVEN: A previous backup exists in target directory
AND:   Files have been deleted from source
WHEN:  User runs 'backup-butler sync --mirror'
THEN:  Extra files in target are identified
AND:   Files present only in target are automatically deleted
AND:   Summary report shows:
      - Count of files deleted from target
      - Total space reclaimed
```

### Backup with Deleted Files - Custom Configuration

```bash
GIVEN: A previous backup exists in target directory
AND:   Files have been deleted from source
AND:   Deletion policy is configured in config file
WHEN:  User runs 'backup-butler sync'
THEN:  Extra files are handled according to policy:
      - "report": Generate detailed report only
      - "delete": Remove files automatically
      - "archive": Move to specified archive location
AND:   Protected paths defined in config are never deleted
AND:   Summary shows actions taken based on configuration
```

## Storage Device Scenarios

### HDD to HDD Backup

```
GIVEN: Source is on HDD
AND:   Target is on HDD
WHEN:  User runs 'backup-butler sync'
THEN:  Operations are optimized for HDD:
      - Files are processed in directory order
      - Read/write operations are sequential
      - Thread count is limited appropriately
AND:   Progress indicates current directory being processed
AND:   Performance metrics show HDD-optimized throughput
```

### SSD to HDD Backup
```
GIVEN: Source is on SSD
AND:   Target is on HDD
WHEN:  User runs 'backup-butler sync'
THEN:  Operations are hybrid-optimized:
      - Source reads use parallel operations
      - Target writes are sequential
      - Thread count is optimized for configuration
AND:   Performance metrics show:
      - Fast read speeds from SSD
      - Write speeds limited by HDD capability
```

## Future Cloud Scenarios

### Local to Cloud Backup
```
GIVEN: Source is local drive (SSD/HDD)
AND:   Target is cloud storage
AND:   Cloud credentials are configured
WHEN:  User runs 'backup-butler sync'
THEN:  Operations are cloud-optimized:
      - Files are uploaded in parallel
      - Bandwidth limits are respected
      - Resume capability is enabled
AND:   Progress shows:
      - Upload speed
      - Remaining time
      - Network usage
```

### Cloud to Local Backup
```
GIVEN: Source is cloud storage
AND:   Target is local drive
AND:   Cloud credentials are configured
WHEN:  User runs 'backup-butler sync'
THEN:  Operations are download-optimized:
      - Files are downloaded in parallel
      - Local write operations are optimized for device type
      - Resume capability is enabled
AND:   Progress shows:
      - Download speed
      - Remaining time
      - Network usage
```

## Advanced Scenarios

### Interrupted Backup Resume
```
GIVEN: A backup operation was interrupted
AND:   Resume state file exists
WHEN:  User runs 'backup-butler sync --resume'
THEN:  Operation resumes from last checkpoint
AND:   Already processed files are skipped
AND:   Summary includes:
      - Files processed before interruption
      - Files processed after resume
      - Total operation statistics
```

### Network Share Backup
```
GIVEN: Source is local drive
AND:   Target is network share
WHEN:  User runs 'backup-butler sync'
THEN:  Operations are network-optimized:
      - Larger buffer sizes for network transfers
      - Network latency is considered
      - Connection stability is monitored
AND:   Progress shows:
      - Network transfer speed
      - Remaining time
      - Network path status
```

### Scheduled Verification (Future Enhancement)

```bash
GIVEN: A backup exists
AND:   Scheduled verification is configured
WHEN:  User runs 'backup-butler check'
THEN:  Verification is performed based on schedule:
      - Quick validation for recent backups
      - Standard validation for weekly checks
      - Deep validation for monthly checks
AND:   Report shows:
      - Validation level used
      - Files checked
      - Any integrity issues found
```


### Large Media Collection Backup

```bash
GIVEN: Source contains >100K media files
AND:   Files range from <1MB to >10GB
WHEN:  User runs 'backup-butler sync'
THEN:  Operations are optimized for large collections:
      - Files are grouped by size
      - Memory usage is managed
      - Regular progress updates
AND:   Summary includes:
      - Size distribution statistics
      - Performance metrics by file size
      - Overall operation summary
```

## Error Scenarios

### Network Interruption
```
GIVEN: Backup is in progress
AND:   Network connection is lost
WHEN:  Connection is restored
THEN:  Operation resumes automatically
AND:   Partial transfers are validated
AND:   Summary shows:
      - Interruption duration
      - Recovery actions taken
      - Final status of affected files
```

### Storage Space Exhaustion
```
GIVEN: Backup is in progress
AND:   Target device runs out of space
THEN:  Operation is paused gracefully
AND:   User is notified of:
      - Space required to complete
      - Files remaining to be copied
      - Options for continuing
```

### Device Performance Degradation
```
GIVEN: Backup is in progress
AND:   Device shows performance issues
THEN:  Operations are adjusted:
      - Reduced thread count
      - Smaller buffer sizes
      - Increased timeouts
AND:   User is notified of:
      - Performance adjustment made
      - Impact on completion time
      - Recommended actions
```