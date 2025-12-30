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
	BackupTotal.Describe(ch)
	desc := <-ch
	assert.NotNil(t, desc)

	// BackupDurationSeconds
	BackupDurationSeconds.Describe(ch)
	desc = <-ch
	assert.NotNil(t, desc)

	// BackupFileSize
	BackupFileSize.Describe(ch)
	desc = <-ch
	assert.NotNil(t, desc)

	// BackupLastTimestamp
	BackupLastTimestamp.Describe(ch)
	desc = <-ch
	assert.NotNil(t, desc)
}

func TestMetricsLabels(t *testing.T) {
	// Test that metrics can be created with labels
	BackupTotal.WithLabelValues("test_model", "success").Add(0)
	BackupTotal.WithLabelValues("test_model", "failure").Add(0)

	BackupDurationSeconds.WithLabelValues("test_model").Observe(0)

	BackupFileSize.WithLabelValues("test_model").Set(0)

	BackupLastTimestamp.WithLabelValues("test_model", "success").Set(0)
	BackupLastTimestamp.WithLabelValues("test_model", "failure").Set(0)
}
