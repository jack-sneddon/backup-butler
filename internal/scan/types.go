// internal/scan/types.go
package scan

// FileStatus represents the comparison status between source and target files
type FileStatus byte

const (
	StatusMatch   FileStatus = '=' // File identical
	StatusNew     FileStatus = '+' // Only in source
	StatusMissing FileStatus = '-' // Only in target
	StatusDiffer  FileStatus = '*' // Content differs
	StatusError   FileStatus = '!' // Error reading/comparing
)

// FileInfo represents metadata for a single file
type FileInfo struct {
	Path    string
	Size    int64
	ModTime int64
	IsDir   bool
	Parent  string
	Status  FileStatus
	Source  *FileInfo
	Target  *FileInfo
}

// DirectoryStats contains aggregated information about a directory
type DirectoryStats struct {
	Path      string
	FileCount int
	TotalSize int64
	Files     []*FileInfo
}

// FileComparison represents the result of comparing a file between source and target
type FileComparison struct {
	Path   string
	Status FileStatus
	Source *FileInfo
	Target *FileInfo
}
