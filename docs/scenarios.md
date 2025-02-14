# Backup Butler: User Scenarios

## Common Daily Operations

### Checking Backup Status

```yaml
GIVEN: User wants to see what changes will occur
WHEN:  User runs 'backup-butler check'
THEN:  Tool displays:
      - Files to be copied
      - Files to be validated
      - Files only in target
      - Estimated time to complete
AND:   No files are modified
```

### Regular Backup

```yaml
GIVEN: Source and target directories exist
WHEN:  User runs 'backup-butler sync'
THEN:  Tool performs directory-by-directory backup
AND:   Shows progress:
      - Current directory being processed
      - Overall progress percentage
      - Files processed and remaining
      - Estimated time to complete
```

### Viewing Last Operation

```yaml
GIVEN: A backup has been performed
WHEN:  User runs 'backup-butler history'
THEN:  Tool shows:
      - Time of last operation
      - Files processed
      - Data transferred
      - Duration
AND:   Shows location of any generated reports
```

## Weekly/Monthly Operations

### Large Backup After Many Changes

```yaml
GIVEN: Many files have been added/changed in source
WHEN:  User runs 'backup-butler sync'
THEN:  Tool processes all changes
AND:   Shows detailed progress
AND:   Maintains directory-based operation
```

### Reviewing Deleted Files

```yaml
GIVEN: Files have been deleted from source
AND:   deleted_files.action = "report" in config
WHEN:  User runs 'backup-butler sync'
THEN:  Tool generates deleted_files.txt report
AND:   Reports location of this file
AND:   Lists all files that exist only in target
```

## Occasional Operations

### Cleanup of Deleted Files

```yaml
GIVEN: User has reviewed deleted_files.txt
AND:   Wants to remove files from target
WHEN:  User updates config deleted_files.action = "delete"
AND:   Runs 'backup-butler sync'
THEN:  Files only in target are removed
AND:   Operation is logged
```

### Error Recovery

```yaml
GIVEN: A backup operation encounters an error
THEN:  Current directory operation is completed
AND:   Error is clearly reported
AND:   User can see which directory had the error
```

## Administrative Operations

### Version Check
```yaml
GIVEN: User wants to verify tool version
WHEN:  User runs 'backup-butler version'
THEN:  Tool displays version information
AND:   Shows build details
```

# Add to scenarios.md under "Common Daily Operations"

### Quick Validation Check

```yaml
GIVEN: User wants to quickly verify backup integrity
WHEN:  User runs 'backup-butler check --level quick'
THEN:  Tool performs fast metadata comparison
AND:   Completes quickly (≈0.1ms per file)
```

### Standard Validation Check

```yaml
GIVEN: User wants regular backup verification
WHEN:  User runs 'backup-butler check --level standard'
THEN:  Tool performs metadata and partial content check
AND:   Shows detailed progress
```

### Deep Validation Check

```yaml
GIVEN: User requires complete backup verification
WHEN:  User runs 'backup-butler check --level deep'
THEN:  Tool performs full content validation
AND:   Shows validation progress and statistics
```

## Error Scenarios

### Configuration Error Handling

```yaml
GIVEN: User has invalid configuration
WHEN:  User runs any backup-butler command
THEN:  Tool displays clear error message
AND:   Suggests how to fix the issue
AND:   Points to configuration documentation
```

### Validation Error Recovery

```yaml
GIVEN: Error occurs during validation
EXAMPLES:
      - Read error
      - Permission denied
      - Disk full
THEN:  Tool shows clear error message
AND:   Continues with next file if possible
AND:   Includes errors in final report
```

### Quick Validation with Standard Escalation

```yaml
GIVEN: User has configured quick validation
WHEN:  User runs 'backup-butler check'
THEN:  Tool starts with quick validation
AND:   When metadata differences found:
      - Automatically performs standard validation
      - Shows "[standard]" level in output
      - Includes escalation in summary stats
```

### Standard Validation with Deep Escalation

```yaml
GIVEN: User has configured standard validation
WHEN:  User runs 'backup-butler check'
THEN:  Tool starts with standard validation
AND:   When partial content differs:
      - Automatically performs deep validation
      - Shows "[deep]" level in output
      - Includes escalation in summary stats
```

### Validation Level Summary Display

```yaml
GIVEN: Mixed validation levels were used
WHEN:  Operation completes
THEN:  Summary shows:
      - Initial validation counts
      - Escalated validation counts
      - Final status with validation levels:
        = file1.txt [quick]
        * file2.txt [standard]
        * file3.txt [deep]
```

### Escalation Performance Impact

```yaml
GIVEN: Many files require escalated validation
WHEN:  User runs 'backup-butler check'
THEN:  Tool adjusts progress and ETA
AND:   Shows when validation level changes
AND:   Updates time estimates accordingly
```

### Storage-Specific Operations

#### SSD Backup

```yaml
GIVEN: Source is on SSD and target is on network storage
WHEN:  User runs 'backup-butler sync'
THEN:  Tool optimizes:
- Higher thread count for SSD reads
- Larger buffer size for network transfers
- Balanced concurrent operations
```

#### Mixed Storage Types

```yaml
GIVEN: Different storage types for source and target
THEN:  Tool automatically:
- Uses appropriate buffer sizes
- Adjusts thread counts
- Optimizes for slowest device
```
