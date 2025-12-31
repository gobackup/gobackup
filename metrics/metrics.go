package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// TotalAttempts is a counter for total backup attempts, labeled by model and status
	TotalAttempts = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "gobackup",
			Name:      "total_attempts",
			Help:      "Total number of backup attempts",
		},
		[]string{"model", "status"},
	)

	// DuractionSeconds is a histogram for backup duration
	DuractionSeconds = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "gobackup",
			Name:      "duration_seconds",
			Help:      "Duration of backup in seconds",
			Buckets:   prometheus.ExponentialBuckets(1, 2, 15), // 1s to ~9h
		},
		[]string{"model"},
	)

	// FileSizes is a gauge for the last backup file size in bytes
	FileSizes = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "gobackup",
			Name:      "file_size_bytes",
			Help:      "Size of the last backup file in bytes",
		},
		[]string{"model"},
	)

	// LastTimestamp is a gauge for the last backup timestamp (Unix epoch)
	LastTimestamp = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "gobackup",
			Name:      "last_timestamp",
			Help:      "Timestamp of the last backup attempt (Unix epoch)",
		},
		[]string{"model", "status"},
	)
)
