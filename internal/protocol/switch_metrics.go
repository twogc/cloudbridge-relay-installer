package protocol

import (
	"github.com/prometheus/client_golang/prometheus"
)

// SwitchMetrics представляет метрики переключения протоколов
type SwitchMetrics struct {
	switches *prometheus.CounterVec
	duration *prometheus.HistogramVec
}

// NewSwitchMetrics создает новые метрики переключения
func NewSwitchMetrics() *SwitchMetrics {
	return &SwitchMetrics{
		switches: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "protocol_switches_total",
				Help: "Total number of protocol switches",
			},
			[]string{"from", "to", "reason"},
		),
		duration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "protocol_switch_duration_seconds",
				Help:    "Duration of protocol switches",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"from", "to"},
		),
	}
}

// RecordSwitch записывает метрику переключения
func (sm *SwitchMetrics) RecordSwitch(from, to Protocol, reason string) {
	sm.switches.WithLabelValues(from.String(), to.String(), reason).Inc()
}

// RecordSwitchDuration записывает длительность переключения
func (sm *SwitchMetrics) RecordSwitchDuration(from, to Protocol, duration float64) {
	sm.duration.WithLabelValues(from.String(), to.String()).Observe(duration)
}

// Register регистрирует метрики в Prometheus
func (sm *SwitchMetrics) Register(registry prometheus.Registerer) error {
	collectors := []prometheus.Collector{
		sm.switches,
		sm.duration,
	}

	for _, collector := range collectors {
		if err := registry.Register(collector); err != nil {
			return err
		}
	}

	return nil
}
