package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// BackupTotal is a counter for total backup attempts, labeled by model and status
	BackupTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "gobackup",
			Name:      "backup_total",
			Help:      "Total number of backup attempts",
		},
		[]string{"model", "status"},
	)

	// BackupDurationSeconds is a histogram for backup duration
	BackupDurationSeconds = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "gobackup",
			Name:      "backup_duration_seconds",
			Help:      "Duration of backup in seconds",
			Buckets:   prometheus.ExponentialBuckets(1, 2, 15), // 1s to ~9h
		},
		[]string{"model"},
	)

	// BackupFileSize is a gauge for the last backup file size in bytes
	BackupFileSize = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "gobackup",
			Name:      "backup_file_size_bytes",
			Help:      "Size of the last backup file in bytes",
		},
		[]string{"model"},
	)

	// BackupLastTimestamp is a gauge for the last backup timestamp (Unix epoch)
	BackupLastTimestamp = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "gobackup",
			Name:      "backup_last_timestamp",
			Help:      "Timestamp of the last backup attempt (Unix epoch)",
		},
		[]string{"model", "status"},
	)
)
