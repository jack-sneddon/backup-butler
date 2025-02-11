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

```bash
GIVEN: File exists in source and target
AND:   Quick validation is configured
THEN:  Compare:
      - File sizes
      - Modification times (2-second tolerance)
RETURN: StatusMatch if metadata matches
```

### Case 14: Standard Validation

```bash
GIVEN: File exists in source and target
AND:   Standard validation is configured
THEN:  Check metadata (file size, modification time) as per quick validation
IF:    Quick validation mismatch
THEN : RETURN StatusDiffer as mismatch
ELSE:  Calculate hash of first 32KB if metadata matches
RETURN: StatusDiffer on any mismatch
```

### Case 15: Deep Validation

```bash
GIVEN: File exists in source and target
AND:   Deep validation is configured
THEN:  Check metadata
IF:    Quick validation mismatch
THEN:  RETURN StatusDiffer as mismatch

ELSE IF:  Calculate hash of first 32KB mismatch
THEN:  RETURN StatusDiffer as mismatch

ELSE:  Calculate full content hash matche
IF:    Hashes do not match
RETURN: StatusDiffer on any mismatch
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

```bash
GIVEN: File exists in source and target
AND:   Standard validation is configured
WHEN:  Validation runs
THEN:  Quick validation runs first
IF:    Quick validation shows mismatch
THEN:  Return mismatch immediately
ELSE:  Proceed with standard (partial content) validation
```

### Case 21: Standard to Deep Escalation

```bash
GIVEN: File exists in source and target
AND:   Deep validation is configured
WHEN:  Validation runs
THEN:  Quick validation runs first
IF:    Quick validation shows mismatch
THEN:  Return mismatch immediately
ELSE:  Standard validation runs
IF:    Standard validation shows mismatch
THEN:  Return mismatch immediately
ELSE:  Proceed with deep (full content) validation
```

### Case 22: Multiple Level Escalation

```bash
GIVEN: Files are being validated
WHEN:  Results are displayed
THEN:  Show validation level used for final result:
= file1.txt [quick]      # Matched at quick check
* file2.txt [quick]      # Mismatched at quick check
= file3.txt [standard]   # Matched after standard check
* file4.txt [standard]   # Mismatched at standard check
= file5.txt [deep]       # Matched after deep check
* file6.txt [deep]       # Mismatched at deep check
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
