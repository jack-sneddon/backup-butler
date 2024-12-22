// internal/domain/backup/types.go
package backup

import "time"

// BackupConfig represents the validated configuration for a backup operation
type BackupConfig struct {
	SourceDirectory    string         `yaml:"source_directory" json:"source_directory"`
	TargetDirectory    string         `yaml:"target_directory" json:"target_directory"`
	FoldersToBackup    []string       `yaml:"folders_to_backup" json:"folders_to_backup"`
	DeepDuplicateCheck bool           `yaml:"deep_duplicate_check" json:"deep_duplicate_check"`
	Concurrency        int            `yaml:"concurrency" json:"concurrency"`
	BufferSize         int            `yaml:"buffer_size" json:"buffer_size"`
	RetryAttempts      int            `yaml:"retry_attempts" json:"retry_attempts"`
	RetryDelay         time.Duration  `yaml:"retry_delay" json:"retry_delay"`
	ExcludePatterns    []string       `yaml:"exclude_patterns" json:"exclude_patterns"`
	ChecksumAlgorithm  string         `yaml:"checksum_algorithm" json:"checksum_algorithm"`
	LogLevel           string         `yaml:"log_level" json:"log_level"`
	Options            *ConfigOptions `yaml:"-" json:"-"`
}

// ConfigOptions represents runtime configuration options
type ConfigOptions struct {
	Verbose  bool   `yaml:"verbose" json:"verbose"`
	Quiet    bool   `yaml:"quiet" json:"quiet"` // Keep this for controlling progress display
	LogLevel string `yaml:"log_level" json:"log_level"`
}

// BackupTask represents a single file backup operation
type BackupTask struct {
	Source      string    // Source file path
	Destination string    // Destination file path
	Size        int64     // File size in bytes
	ModTime     time.Time // File modification time
}

// FileMetadata represents file-level information
type FileMetadata struct {
	Path     string
	Size     int64
	ModTime  time.Time
	Checksum string
}

// BackupStats holds metrics about a backup operation
type BackupStats struct {
	TotalFiles       int   // Total number of files processed
	FilesBackedUp    int   // Number of files actually copied
	FilesSkipped     int   // Number of unchanged files
	FilesFailed      int   // Number of files that failed to backup
	TotalBytes       int64 // Total bytes processed
	BytesTransferred int64 // Actual bytes copied
}

// BackupVersion represents a completed backup operation
type BackupVersion struct {
	ID         string                  // Unique identifier (timestamp-based)
	Timestamp  time.Time               // When backup was performed
	Files      map[string]FileMetadata // Map of path to file metadata
	Size       int64                   // Total size of backup
	Status     string                  // Success, Failed, Partial
	Duration   time.Duration           // How long the backup took
	Stats      BackupStats             // Additional statistics
	ConfigUsed BackupConfig            // Configuration used for this backup
}
