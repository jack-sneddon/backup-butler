// internal/version/manager.go
package version

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/jack-sneddon/backup-butler/internal/config"
	"github.com/jack-sneddon/backup-butler/internal/types"
)

// Manager handles backup version tracking
type Manager struct {
	targetDir string
	current   *BackupVersion
}

// NewManager creates a new version manager
func NewManager(targetDir string) (*Manager, error) {
	// Create versions directory if it doesn't exist
	versionsDir := filepath.Join(targetDir, ".versions")
	if err := os.MkdirAll(versionsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create versions directory: %w", err)
	}

	return &Manager{
		targetDir: targetDir,
	}, nil
}

// StartNewVersion begins tracking a new backup operation
func (m *Manager) StartNewVersion(cfg *config.Config) *BackupVersion {
	version := &BackupVersion{
		ID:        time.Now().Format("20060102-150405"),
		StartTime: time.Now(),
		Status:    "running",
		Config:    cfg,
		Files:     make(map[string]FileResult),
		Stats:     BackupStats{},
	}
	m.current = version
	return version
}

// RecordFile records the result of backing up a single file
func (m *Manager) RecordFile(relPath string, result FileResult) error {
	if m.current == nil {
		return fmt.Errorf("no backup version in progress")
	}

	m.current.Files[relPath] = result

	// Update statistics
	switch result.Status {
	case "copied":
		m.current.Stats.FilesCopied++
		m.current.Stats.BytesCopied += result.Size
	case "skipped":
		m.current.Stats.FilesSkipped++
		m.current.Stats.BytesSkipped += result.Size
	case "failed":
		m.current.Stats.FilesFailed++
	}
	m.current.Stats.TotalFiles++

	return nil
}

// CompleteVersion finalizes the current backup version
func (m *Manager) CompleteVersion(status string) error {
	if m.current == nil {
		return fmt.Errorf("no backup version in progress")
	}

	m.current.EndTime = time.Now()
	m.current.Status = status

	// Save version file
	versionFile := filepath.Join(m.targetDir, ".versions", m.current.ID+".json")
	data, err := json.MarshalIndent(m.current, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal version data: %w", err)
	}

	if err := os.WriteFile(versionFile, data, 0644); err != nil {
		return fmt.Errorf("failed to save version file: %w", err)
	}

	return nil
}

// GetVersions returns a list of all backup versions
// GetVersions returns a list of all backup versions
func (m *Manager) GetVersions() ([]VersionSummary, error) {
	versionsDir := filepath.Join(m.targetDir, ".versions")
	entries, err := os.ReadDir(versionsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read versions directory: %w", err)
	}

	var summaries []VersionSummary
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(versionsDir, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed to read version file %s: %w", entry.Name(), err)
		}

		var version BackupVersion
		if err := json.Unmarshal(data, &version); err != nil {
			return nil, fmt.Errorf("failed to parse version file %s: %w", entry.Name(), err)
		}

		summary := VersionSummary{
			ID:           version.ID,
			StartTime:    version.StartTime,
			EndTime:      version.EndTime,
			Status:       version.Status,
			TotalFiles:   version.Stats.TotalFiles,
			CopiedFiles:  version.Stats.FilesCopied,
			CopiedBytes:  version.Stats.BytesCopied,
			SkippedFiles: version.Stats.FilesSkipped,
			SkippedBytes: version.Stats.BytesSkipped,
			FailedFiles:  version.Stats.FilesFailed,
		}

		// Calculate total bytes
		summary.TotalBytes = version.Stats.BytesCopied + version.Stats.BytesSkipped

		summaries = append(summaries, summary)
	}

	// Sort by start time, newest first
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].StartTime.After(summaries[j].StartTime)
	})

	return summaries, nil
}

// GetVersion retrieves a specific backup version
func (m *Manager) GetVersion(id string) (*BackupVersion, error) {
	versionFile := filepath.Join(m.targetDir, ".versions", id+".json")
	data, err := os.ReadFile(versionFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("version %s not found", id)
		}
		return nil, fmt.Errorf("failed to read version file: %w", err)
	}

	var version BackupVersion
	if err := json.Unmarshal(data, &version); err != nil {
		return nil, fmt.Errorf("failed to parse version file: %w", err)
	}

	return &version, nil
}

// GetLatestVersion retrieves the most recent backup version
func (m *Manager) GetLatestVersion() (*BackupVersion, error) {
	summaries, err := m.GetVersions()
	if err != nil {
		return nil, err
	}

	if len(summaries) == 0 {
		return nil, fmt.Errorf("no backup versions found")
	}

	return m.GetVersion(summaries[0].ID)
}

// internal/version/manager.go

func (m *Manager) GetFileLastVersion(path string) (*types.FileVersionInfo, error) {
	// Get all versions sorted by time (newest first)
	versions, err := m.GetVersions()
	if err != nil {
		return nil, fmt.Errorf("failed to get versions: %w", err)
	}

	// Look through versions for the file's last backup
	for _, ver := range versions {
		// Get full version details
		fullVer, err := m.GetVersion(ver.ID)
		if err != nil {
			continue // Skip problematic versions
		}

		// Check if file exists in this version
		if fileResult, exists := fullVer.Files[path]; exists {
			return &types.FileVersionInfo{
				ID:        fullVer.ID,
				Path:      path,
				Size:      fileResult.Size,
				ModTime:   fileResult.ModTime,
				Checksum:  fileResult.Checksum,
				QuickHash: fileResult.QuickHash,
			}, nil
		}
	}

	return nil, fmt.Errorf("no version history found for file: %s", path)
}
