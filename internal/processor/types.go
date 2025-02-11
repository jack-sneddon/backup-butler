// internal/processor/types.go
package processor

type DirectoryProcessor interface {
	// ProcessDirectory handles all files in a single directory
	ProcessDirectory(sourcePath, targetPath string) error
}

type ProcessorOptions struct {
	PreserveMetadata bool
	BufferSize       int
}

// Basic result tracking
type DirectoryResult struct {
	FilesProcessed int
	BytesProcessed int64
	Errors         []error
}
