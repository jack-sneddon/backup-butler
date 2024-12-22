// internal/adapters/metrics/collector/adapter.go
package collector

import (
	"context"

	"github.com/jack-sneddon/backup-butler/internal/core/monitoring"
	"github.com/jack-sneddon/backup-butler/internal/domain/backup"
)

type MetricsCollector struct {
	metrics *monitoring.Metrics
}

func NewMetricsCollector(totalFiles int, quiet bool) *MetricsCollector {
	metrics := monitoring.NewMetrics(quiet)
	metrics.SetTotalFiles(totalFiles)
	return &MetricsCollector{
		metrics: metrics,
	}
}

func (m *MetricsCollector) StartTracking(ctx context.Context) {
	m.metrics.StartTracking(ctx)
}

func (m *MetricsCollector) IncrementCompleted(bytes int64) {
	m.metrics.IncrementCompleted(bytes)
}

func (m *MetricsCollector) IncrementSkipped(bytes int64) {
	m.metrics.IncrementSkipped(bytes)
}

func (m *MetricsCollector) IncrementFailed() {
	m.metrics.IncrementFailed()
}

func (m *MetricsCollector) GetStats() backup.BackupStats {
	return m.metrics.GetStats()
}

func (m *MetricsCollector) DisplayProgress() {
	m.metrics.DisplayProgress()
}

func (m *MetricsCollector) DisplayFinalSummary() {
	m.metrics.DisplayFinalSummary()
}

func (m *MetricsCollector) SetTotalFiles(total int) {
	m.metrics.SetTotalFiles(total)
}
