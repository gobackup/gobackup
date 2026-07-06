package metrics

import (
	"testing"

	"github.com/longbridgeapp/assert"
	"github.com/prometheus/client_golang/prometheus"
)

func TestMetricsRegistration(t *testing.T) {
	// Verify all metrics are registered by attempting to describe them
	ch := make(chan *prometheus.Desc, 10)

	// BackupTotal
	TotalAttempts.Describe(ch)
	desc := <-ch
	assert.NotNil(t, desc)

	// BackupDurationSeconds
	DuractionSeconds.Describe(ch)
	desc = <-ch
	assert.NotNil(t, desc)

	// BackupFileSize
	FileSizes.Describe(ch)
	desc = <-ch
	assert.NotNil(t, desc)

	// BackupLastTimestamp
	LastTimestamp.Describe(ch)
	desc = <-ch
	assert.NotNil(t, desc)
}

func TestMetricsLabels(t *testing.T) {
	// Test that metrics can be created with labels
	TotalAttempts.WithLabelValues("test_model", "success").Add(0)
	TotalAttempts.WithLabelValues("test_model", "failure").Add(0)

	DuractionSeconds.WithLabelValues("test_model").Observe(0)

	FileSizes.WithLabelValues("test_model").Set(0)

	LastTimestamp.WithLabelValues("test_model", "success").Set(0)
	LastTimestamp.WithLabelValues("test_model", "failure").Set(0)
}
