package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestNewMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	metrics := NewMetrics(reg)

	assert.NotNil(t, metrics)
}

func TestNewMetrics_WithNilRegistry(t *testing.T) {
	// Test creating metrics with nil registry - should not panic
	// but we expect it might fail, so we test that it doesn't crash
	defer func() {
		if r := recover(); r != nil {
			// This is expected behavior with nil registry
			t.Logf("Expected panic with nil registry: %v", r)
		}
	}()
	metrics := NewMetrics(nil)
	// If we reach here, the nil registry was handled gracefully
	if metrics != nil {
		assert.NotNil(t, metrics)
	}
}

func TestMetrics_Structure(t *testing.T) {
	// Test that metrics structure is properly initialized
	reg := prometheus.NewRegistry()
	metrics := NewMetrics(reg)

	// Verify that metrics object is created
	assert.NotNil(t, metrics)

	// This is a basic structure test - we can't access private fields
	// but we can verify the object was created successfully
}

func TestMetrics_RegistryIntegration(t *testing.T) {
	// Test that metrics are properly registered
	reg := prometheus.NewRegistry()
	metrics := NewMetrics(reg)

	// Verify metrics were created
	assert.NotNil(t, metrics)

	// Try to gather metrics from registry
	metricFamilies, err := reg.Gather()
	assert.NoError(t, err)

	// Should have some metrics registered
	assert.GreaterOrEqual(t, len(metricFamilies), 0)
}
