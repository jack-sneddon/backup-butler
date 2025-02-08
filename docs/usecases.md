# Backup Butler: Technical Use Cases

## Core Operations

### Case 1: Directory Processing
```
GIVEN: Source directory with multiple subdirectories
WHEN:  Processing starts
THEN:  Process one directory at a time
AND:   Complete current directory before moving to next
AND:   Track progress at directory level
```

### Case 2: File Validation Levels
```
GIVEN: Files exist in both source and target
WHEN:  Validation is performed
THEN:  Apply appropriate validation level:
      - Quick for initial check
      - Standard if quick fails
      - Deep based on configuration
```

### Case 3: Progress Tracking
```
GIVEN: Operation is in progress
THEN:  Show current directory being processed
AND:   Show overall progress percentage
AND:   Calculate remaining time based on:
      - Files processed
      - Average processing time
      - Remaining files
```

## Operation History

### Case 4: Operation Logging
```
GIVEN: Backup operation completes
THEN:  Log to .backup-butler/logs:
      - Operation timestamps
      - Files processed
      - Data transferred
      - Duration
      - Any errors encountered
```

### Case 5: Report Generation
```
GIVEN: Files exist only in target
AND:   deleted_files.action = "report"
THEN:  Generate deleted_files.txt containing:
      - File paths
      - File sizes
      - Last modification times
      - Total affected files/size
```

## Error Handling

### Case 6: Directory Error
```
GIVEN: Error occurs during directory processing
THEN:  Complete current file if possible
AND:   Log error details
AND:   Move to next directory if possible
```

### Case 7: File Access Error
```
GIVEN: File cannot be read or written
THEN:  Log specific error
AND:   Mark file as error in report
AND:   Continue with next file
```

## Performance Optimization

### Case 8: HDD Optimization
```
GIVEN: Target is on HDD
THEN:  Process files sequentially
AND:   Minimize directory switches
AND:   Use conservative I/O settings
```

### Case 9: Large Directory
```
GIVEN: Directory contains many files
THEN:  Show frequent progress updates
AND:   Maintain accurate ETA
AND:   Process files in manageable chunks
```

## Configuration Validation

### Case 10: Invalid Source Directory
```
GIVEN: Source directory in config doesn't exist
WHEN:  Any command is run
THEN:  Validate configuration before starting
AND:   Display clear error message:
      - Actual error (directory not found)
      - Config file location
      - Invalid setting
AND:   Exit with non-zero status
```

### Case 11: Invalid Configuration Format
```
GIVEN: Configuration file has syntax errors
WHEN:  Any command is run
THEN:  Display clear error message:
      - Location of syntax error
      - Expected format
      - Actual content
AND:   Exit with non-zero status
```

### Case 12: Invalid Configuration Values
```
GIVEN: Configuration has invalid values
EXAMPLES:
      - Negative buffer size
      - Unknown validation level
      - Invalid hash algorithm
THEN:  Display clear error message:
      - Invalid setting
      - Allowed values
      - Current value
AND:   Exit with non-zero status
```

## Validation Operations

### Case 13: Quick Validation
```
GIVEN: File exists in source and target
AND:   Quick validation is configured
THEN:  Compare:
      - File sizes
      - Modification times (2-second tolerance)
IF:    Both match
THEN:  Mark as valid (=)
ELSE:  Escalate if configured
      OR Mark as different (*)
```

### Case 14: Standard Validation
```
GIVEN: File exists in source and target
AND:   Standard validation is configured
THEN:  First perform quick validation
IF:    Quick validation passes
THEN:  Read first 32KB of both files
AND:   Calculate hash of these buffers
IF:    Hashes match
THEN:  Mark as valid (=)
ELSE:  Escalate if configured
      OR Mark as different (*)
```

### Case 15: Deep Validation
```
GIVEN: File exists in source and target
AND:   Deep validation is configured
THEN:  First perform quick validation
IF:    Quick validation passes
THEN:  Calculate hash of entire file contents
IF:    Hashes match
THEN:  Mark as valid (=)
ELSE:  Mark as different (*)
```

## Error Handling

### Case 16: Hash Calculation Errors
```
GIVEN: File being validated
WHEN:  Hash calculation fails
POSSIBLE CAUSES:
      - Read error during calculation
      - Memory allocation failure
      - Disk error
THEN:  Log specific error
AND:   Mark file with error status (!)
AND:   Include in error summary
```

### Case 17: Partial Read Errors
```
GIVEN: File being validated
WHEN:  Partial content read fails
THEN:  Log specific error
AND:   Mark file with error status (!)
AND:   Continue with next file
```

### Case 18: Permission Errors
```
GIVEN: File requires specific permissions
WHEN:  File cannot be accessed
THEN:  Log permission error
AND:   Mark file with error status (!)
AND:   Include permission details in report
```

### Case 19: Disk Space Errors
```
GIVEN: Operation is in progress
WHEN:  Disk space becomes full
THEN:  Complete current file if possible
AND:   Log error with space requirements
AND:   Exit gracefully with error status
```

### Case 20: Quick to Standard Escalation
```
GIVEN: File exists in source and target
AND:   Quick validation is configured
AND:   on_mismatch = "standard"
WHEN:  Quick validation fails (size/time mismatch)
THEN:  Automatically escalate to standard validation
AND:   Read first 32KB of both files
AND:   Calculate partial content hashes
AND:   Mark final status with [standard] level indicator
```

### Case 21: Standard to Deep Escalation
```
GIVEN: File exists in source and target
AND:   Standard validation is configured
AND:   on_mismatch = "deep"
WHEN:  Standard validation fails (partial content differs)
THEN:  Automatically escalate to deep validation
AND:   Calculate full file hashes
AND:   Mark final status with [deep] level indicator
```

### Case 22: Multiple Level Escalation
```
GIVEN: File exists in source and target
AND:   Quick validation is configured
AND:   on_mismatch = "deep"
WHEN:  Quick validation fails
THEN:  Skip standard validation
AND:   Escalate directly to deep validation
AND:   Mark final status with [deep] level indicator
```

### Case 23: Escalation with Errors
```
GIVEN: Validation escalation is configured
WHEN:  Error occurs during escalated check
THEN:  Mark file with error status (!)
AND:   Log original validation level
AND:   Log level where error occurred
AND:   Include in error summary
```

## Test Coverage Recommendations

1. File Status Tests:
   - New files
   - Missing files
   - Matching files
   - Different files

2. Quick Validation Tests:
   - Identical metadata
   - Different sizes
   - Different modification times
   - Modification times within tolerance
   - Modification times outside tolerance

3. Standard Validation Tests:
   - Identical first 32KB, different later content
   - Different first 32KB
   - Files exactly 32KB
   - Files smaller than 32KB

4. Deep Validation Tests:
   - Completely identical files
   - Files differing in last byte
   - Empty files
   - Very large files

5. Escalation Tests:
   - Quick to Standard escalation
   - Standard to Deep escalation
   - Failed escalation handling

6. Error Handling Tests:
   - Inaccessible files
   - Read errors
   - Hash calculation failures
   - Memory constraints
