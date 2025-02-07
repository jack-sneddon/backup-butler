# Validation Level Analysis

## Initial File Status Checks

### Case 1: Missing Target File

```bash
GIVEN: Source file exists
AND:   Target file does not exist
THEN:  Status = StatusNew ('+')
AND:   No validation performed (file needs to be copied)
```

### Case 2: Extra Target File

```bash
GIVEN: Source file does not exist
AND:   Target file exists
THEN:  Status = StatusMissing ('-')
AND:   No validation performed (file may need to be removed)
```

## Quick Validation Cases

Default starting point when default_level = "quick"

### Case 3: Basic Metadata Check

```bash
GIVEN: Source and target files exist
AND:   default_level = "quick"
THEN:  Compare:
       - File sizes
       - Modification times (with 2-second tolerance)
IF:    Sizes match AND ModTimes within tolerance
THEN:  Status = StatusMatch ('=')
ELSE:  IF on_mismatch = "standard"
       THEN Escalate to Standard validation
       ELSE Status = StatusDiffer ('*')
```

## Standard Validation Cases

Used when default_level = "standard" or escalated from quick

### Case 4: Standard Content Check

```bash
GIVEN: Source and target files exist
AND:   (default_level = "standard" OR escalated from quick)
THEN:  1. Perform quick validation first
       2. If quick validation passes:
          - Read first 32KB of both files
          - Calculate hash of buffers
          - Compare hashes
IF:    Hashes match
THEN:  Status = StatusMatch ('=')
ELSE:  IF on_mismatch = "deep"
       THEN Escalate to Deep validation
       ELSE Status = StatusDiffer ('*')
```

## Deep Validation Cases

Used when default_level = "deep" or escalated from standard

### Case 5: Deep Content Check

```bash
GIVEN: Source and target files exist
AND:   (default_level = "deep" OR escalated from standard)
THEN:  1. Perform quick validation first
       2. If quick validation passes:
          - Read entire content of both files
          - Calculate full file hashes
          - Compare hashes
IF:    Hashes match
THEN:  Status = StatusMatch ('=')
ELSE:  Status = StatusDiffer ('*')
```

## Validation Level Escalation Rules

### Case 6: Quick to Standard Escalation

```bash
GIVEN: initial_level = "quick"
AND:   on_mismatch = "standard"
WHEN:  Quick validation fails
THEN:  Perform standard validation
```

### Case 7: Standard to Deep Escalation

```bash
GIVEN: initial_level = "standard"
AND:   on_mismatch = "deep"
WHEN:  Standard validation fails
THEN:  Perform deep validation
```

## Error Handling Cases

### Case 8: File Access Errors

```bash
GIVEN: Any validation level
WHEN:  File cannot be opened or read
THEN:  Status = StatusError ('!')
AND:   Error logged
AND:   No further validation attempted
```

### Case 9: Hash Calculation Errors

```bash
GIVEN: Standard or Deep validation
WHEN:  Hash calculation fails
THEN:  Status = StatusError ('!')
AND:   Error logged
AND:   No further validation attempted
```

## Performance Considerations

### Case 10: Large File Handling

```bash
GIVEN: Large files (> 1GB)
WHEN:  Deep validation requested
THEN:  Process in chunks to manage memory
AND:   Use buffered reading
```

## Deleted File Handling Cases

### Case 11: Default Report Mode

```bash
GIVEN: Source file has been deleted
AND:   File exists in target
AND:   No specific deletion policy configured
THEN:  File is marked as StatusMissing ('-')
AND:   File is included in detailed report
AND:   No action taken on target file
```

### Case 12: Mirror Mode Deletion

```
GIVEN: Source file has been deleted
AND:   File exists in target
AND:   Mirror mode is enabled (--mirror flag)
THEN:  File is marked as StatusMissing ('-')
AND:   File is automatically deleted from target
AND:   Deletion is logged in operation report
```

### Case 13: Configured Archive Mode

```
GIVEN: Source file has been deleted
AND:   File exists in target
AND:   Archive mode is configured with valid path
THEN:  File is marked as StatusMissing ('-')
AND:   File is moved to archive location
AND:   Original path structure is preserved in archive
AND:   Move operation is logged
```

### Case 14: Protected Path Handling

```
GIVEN: Source file has been deleted
AND:   File exists in target
AND:   File matches protected path pattern
THEN:  File is marked as StatusMissing ('-')
AND:   No action is taken regardless of deletion policy
AND:   Protection reason is noted in report
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
