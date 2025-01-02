package version

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/jack-sneddon/backup-butler/internal/config"
	"github.com/jack-sneddon/backup-butler/internal/types"
)

type Manager struct {
	baseDir     string
	index       *FileIndex
	indexLock   sync.RWMutex
	currentVer  *BackupVersion
	maxVersions int // For version retention
}

func NewManager(baseDir string) (*Manager, error) {
	m := &Manager{
		baseDir:     baseDir,
		maxVersions: 30, // Keep last 30 versions by default
	}

	// Create version directory structure
	dirs := []string{
		filepath.Join(baseDir, ".versions"),
		filepath.Join(baseDir, ".versions", "backups"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Load or create index
	if err := m.loadIndex(); err != nil {
		return nil, err
	}

	return m, nil
}

func (m *Manager) loadIndex() error {
	indexPath := filepath.Join(m.baseDir, ".versions", "index.json")
	data, err := os.ReadFile(indexPath)
	if err != nil {
		if os.IsNotExist(err) {
			m.index = &FileIndex{
				LastUpdated: time.Now(),
				Files:       make(map[string]FileMetadata),
			}
			return m.saveIndex()
		}
		return fmt.Errorf("failed to read index: %w", err)
	}

	m.index = &FileIndex{}
	if err := json.Unmarshal(data, m.index); err != nil {
		return fmt.Errorf("failed to parse index: %w", err)
	}

	return nil
}

func (m *Manager) saveIndex() error {
	m.indexLock.Lock()
	defer m.indexLock.Unlock()

	m.index.LastUpdated = time.Now()
	indexPath := filepath.Join(m.baseDir, ".versions", "index.json")
	tempPath := indexPath + ".tmp"

	data, err := json.MarshalIndent(m.index, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal index: %w", err)
	}

	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temporary index: %w", err)
	}

	if err := os.Rename(tempPath, indexPath); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to save index: %w", err)
	}

	return nil
}

func (m *Manager) StartNewVersion(config *config.Config) *BackupVersion {
	version := &BackupVersion{
		ID:        time.Now().Format("20060102-150405"),
		StartTime: time.Now(),
		Changes:   make([]FileChange, 0),
	}

	// Initialize stats
	version.Stats.Directories = make(map[string]DirectoryStats)
	for _, folder := range config.FoldersToBackup {
		version.Stats.Directories[folder] = DirectoryStats{}
	}

	m.currentVer = version
	return version
}

// internal/version/manager.go
func (m *Manager) RecordFile(path string, status string, size int64, modTime time.Time, checksum string) error {
	if m.currentVer == nil {
		return fmt.Errorf("no backup version in progress")
	}

	// Record the change
	change := FileChange{
		Path:      path,
		Action:    status,
		Size:      size,
		Timestamp: time.Now(),
	}
	if status == "copied" {
		change.Checksum = checksum
	}
	m.currentVer.Changes = append(m.currentVer.Changes, change)

	// Update directory stats
	dirPath := filepath.Dir(path)
	for dir, stats := range m.currentVer.Stats.Directories {
		if strings.HasPrefix(dirPath, dir) {
			switch status {
			case "copied":
				stats.TotalFiles++
				stats.TotalBytes += size
				stats.CopiedFiles++
				stats.CopiedBytes += size
			case "skipped":
				stats.TotalFiles++
				stats.TotalBytes += size
				stats.SkippedFiles++
				stats.SkippedBytes += size
			case "failed":
				stats.TotalFiles++
				stats.FailedFiles++
			}
			m.currentVer.Stats.Directories[dir] = stats
			break
		}
	}

	// Update total stats
	switch status {
	case "copied":
		m.currentVer.Stats.Total.TotalFiles++
		m.currentVer.Stats.Total.FilesCopied++
		m.currentVer.Stats.Total.BytesCopied += size
	case "skipped":
		m.currentVer.Stats.Total.TotalFiles++
		m.currentVer.Stats.Total.FilesSkipped++
		m.currentVer.Stats.Total.BytesSkipped += size
	case "failed":
		m.currentVer.Stats.Total.TotalFiles++
		m.currentVer.Stats.Total.FilesFailed++
	}

	// Update file index
	m.indexLock.Lock()
	m.index.Files[path] = FileMetadata{
		LastBackupID: m.currentVer.ID,
		Size:         size,
		ModTime:      modTime,
		Checksum:     checksum,
	}
	m.indexLock.Unlock()

	return m.saveIndex()
}

func (m *Manager) CompleteVersion() error {
	if m.currentVer == nil {
		return fmt.Errorf("no backup version in progress")
	}

	m.currentVer.EndTime = time.Now()

	// Save version file
	versionFile := filepath.Join(m.baseDir, ".versions", "backups", m.currentVer.ID+".json")
	data, err := json.MarshalIndent(m.currentVer, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal version data: %w", err)
	}

	if err := os.WriteFile(versionFile, data, 0644); err != nil {
		return fmt.Errorf("failed to save version file: %w", err)
	}

	// Cleanup old versions
	if err := m.cleanupOldVersions(); err != nil {
		return fmt.Errorf("failed to cleanup old versions: %w", err)
	}

	m.currentVer = nil
	return nil
}

func (m *Manager) cleanupOldVersions() error {
	backupsDir := filepath.Join(m.baseDir, ".versions", "backups")
	entries, err := os.ReadDir(backupsDir)
	if err != nil {
		return fmt.Errorf("failed to read backups directory: %w", err)
	}

	var versions []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			versions = append(versions, entry.Name())
		}
	}

	// Sort versions by name (timestamp format ensures chronological order)
	sort.Sort(sort.Reverse(sort.StringSlice(versions)))

	// Remove excess versions
	if len(versions) > m.maxVersions {
		for _, v := range versions[m.maxVersions:] {
			if err := os.Remove(filepath.Join(backupsDir, v)); err != nil {
				return fmt.Errorf("failed to remove old version %s: %w", v, err)
			}
		}
	}

	return nil
}

func (m *Manager) GetVersions() ([]VersionSummary, error) {
	backupsDir := filepath.Join(m.baseDir, ".versions", "backups")
	entries, err := os.ReadDir(backupsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read backups directory: %w", err)
	}

	var summaries []VersionSummary
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			data, err := os.ReadFile(filepath.Join(backupsDir, entry.Name()))
			if err != nil {
				return nil, fmt.Errorf("failed to read version file %s: %w", entry.Name(), err)
			}

			var version BackupVersion
			if err := json.Unmarshal(data, &version); err != nil {
				return nil, fmt.Errorf("failed to parse version file %s: %w", entry.Name(), err)
			}

			summary := VersionSummary{
				ID:        version.ID,
				StartTime: version.StartTime,
				EndTime:   version.EndTime,
				Stats:     version.Stats.Total,
				DirStats:  version.Stats.Directories,
			}
			summaries = append(summaries, summary)
		}
	}

	// Sort by start time, newest first
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].StartTime.After(summaries[j].StartTime)
	})

	return summaries, nil
}

// internal/version/manager.go
func (m *Manager) GetFileLastVersion(path string) (*types.FileVersionInfo, error) {
	m.indexLock.RLock()
	metadata, exists := m.index.Files[path]
	m.indexLock.RUnlock()

	if !exists {
		return nil, fmt.Errorf("no version history found for file: %s", path)
	}

	return &types.FileVersionInfo{
		ID:       metadata.LastBackupID,
		Path:     path,
		Size:     metadata.Size,
		ModTime:  metadata.ModTime,
		Checksum: metadata.Checksum,
	}, nil
}
