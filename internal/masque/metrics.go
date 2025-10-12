package masque

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// MASQUEMetrics содержит метрики для MASQUE прокси
type MASQUEMetrics struct {
	// Счетчики соединений
	ConnectionsTotal     prometheus.Counter
	ConnectionsActive    prometheus.Gauge
	ConnectionsFailed    prometheus.Counter
	ConnectionsClosed    prometheus.Counter

	// Метрики трафика
	BytesReceivedTotal   prometheus.Counter
	BytesSentTotal       prometheus.Counter
	PacketsReceivedTotal prometheus.Counter
	PacketsSentTotal     prometheus.Counter

	// Метрики производительности
	RequestDuration      prometheus.Histogram
	ProxyLatency         prometheus.Histogram
	UDPRelayLatency      prometheus.Histogram

	// Метрики ошибок
	ProxyErrors          prometheus.Counter
	UDPErrors            prometheus.Counter
	ConnectionErrors     prometheus.Counter
	TimeoutErrors        prometheus.Counter

	// Метрики конфигурации
	ProxyStatus          prometheus.Gauge
	TargetResolutions    prometheus.Counter
	TargetResolutionErrors prometheus.Counter
}

// NewMASQUEMetrics создает новый экземпляр метрик MASQUE
func NewMASQUEMetrics() *MASQUEMetrics {
	return &MASQUEMetrics{
		// Счетчики соединений
		ConnectionsTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "masque_connections_total",
			Help: "Total number of MASQUE connections established",
		}),
		ConnectionsActive: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "masque_connections_active",
			Help: "Number of active MASQUE connections",
		}),
		ConnectionsFailed: promauto.NewCounter(prometheus.CounterOpts{
			Name: "masque_connections_failed_total",
			Help: "Total number of failed MASQUE connections",
		}),
		ConnectionsClosed: promauto.NewCounter(prometheus.CounterOpts{
			Name: "masque_connections_closed_total",
			Help: "Total number of closed MASQUE connections",
		}),

		// Метрики трафика
		BytesReceivedTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "masque_bytes_received_total",
			Help: "Total bytes received through MASQUE proxy",
		}),
		BytesSentTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "masque_bytes_sent_total",
			Help: "Total bytes sent through MASQUE proxy",
		}),
		PacketsReceivedTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "masque_packets_received_total",
			Help: "Total packets received through MASQUE proxy",
		}),
		PacketsSentTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "masque_packets_sent_total",
			Help: "Total packets sent through MASQUE proxy",
		}),

		// Метрики производительности
		RequestDuration: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "masque_request_duration_seconds",
			Help:    "Duration of MASQUE requests",
			Buckets: prometheus.DefBuckets,
		}),
		ProxyLatency: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "masque_proxy_latency_seconds",
			Help:    "Latency of MASQUE proxy operations",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		}),
		UDPRelayLatency: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "masque_udp_relay_latency_seconds",
			Help:    "Latency of UDP relay operations",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		}),

		// Метрики ошибок
		ProxyErrors: promauto.NewCounter(prometheus.CounterOpts{
			Name: "masque_proxy_errors_total",
			Help: "Total number of MASQUE proxy errors",
		}),
		UDPErrors: promauto.NewCounter(prometheus.CounterOpts{
			Name: "masque_udp_errors_total",
			Help: "Total number of UDP relay errors",
		}),
		ConnectionErrors: promauto.NewCounter(prometheus.CounterOpts{
			Name: "masque_connection_errors_total",
			Help: "Total number of connection errors",
		}),
		TimeoutErrors: promauto.NewCounter(prometheus.CounterOpts{
			Name: "masque_timeout_errors_total",
			Help: "Total number of timeout errors",
		}),

		// Метрики конфигурации
		ProxyStatus: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "masque_proxy_status",
			Help: "MASQUE proxy status (1=active, 0=inactive)",
		}),
		TargetResolutions: promauto.NewCounter(prometheus.CounterOpts{
			Name: "masque_target_resolutions_total",
			Help: "Total number of target address resolutions",
		}),
		TargetResolutionErrors: promauto.NewCounter(prometheus.CounterOpts{
			Name: "masque_target_resolution_errors_total",
			Help: "Total number of target resolution errors",
		}),
	}
}

// IncrementConnections увеличивает счетчик соединений
func (m *MASQUEMetrics) IncrementConnections() {
	m.ConnectionsTotal.Inc()
	m.ConnectionsActive.Inc()
}

// DecrementConnections уменьшает счетчик активных соединений
func (m *MASQUEMetrics) DecrementConnections() {
	m.ConnectionsActive.Dec()
}

// IncrementFailedConnections увеличивает счетчик неудачных соединений
func (m *MASQUEMetrics) IncrementFailedConnections() {
	m.ConnectionsFailed.Inc()
}

// IncrementClosedConnections увеличивает счетчик закрытых соединений
func (m *MASQUEMetrics) IncrementClosedConnections() {
	m.ConnectionsClosed.Inc()
}

// AddBytesReceived добавляет байты к счетчику полученных байтов
func (m *MASQUEMetrics) AddBytesReceived(bytes int64) {
	m.BytesReceivedTotal.Add(float64(bytes))
}

// AddBytesSent добавляет байты к счетчику отправленных байтов
func (m *MASQUEMetrics) AddBytesSent(bytes int64) {
	m.BytesSentTotal.Add(float64(bytes))
}

// IncrementPacketsReceived увеличивает счетчик полученных пакетов
func (m *MASQUEMetrics) IncrementPacketsReceived() {
	m.PacketsReceivedTotal.Inc()
}

// IncrementPacketsSent увеличивает счетчик отправленных пакетов
func (m *MASQUEMetrics) IncrementPacketsSent() {
	m.PacketsSentTotal.Inc()
}

// ObserveRequestDuration записывает длительность запроса
func (m *MASQUEMetrics) ObserveRequestDuration(duration float64) {
	m.RequestDuration.Observe(duration)
}

// ObserveProxyLatency записывает задержку прокси
func (m *MASQUEMetrics) ObserveProxyLatency(latency float64) {
	m.ProxyLatency.Observe(latency)
}

// ObserveUDPRelayLatency записывает задержку UDP релея
func (m *MASQUEMetrics) ObserveUDPRelayLatency(latency float64) {
	m.UDPRelayLatency.Observe(latency)
}

// IncrementProxyErrors увеличивает счетчик ошибок прокси
func (m *MASQUEMetrics) IncrementProxyErrors() {
	m.ProxyErrors.Inc()
}

// IncrementUDPErrors увеличивает счетчик ошибок UDP
func (m *MASQUEMetrics) IncrementUDPErrors() {
	m.UDPErrors.Inc()
}

// IncrementConnectionErrors увеличивает счетчик ошибок соединения
func (m *MASQUEMetrics) IncrementConnectionErrors() {
	m.ConnectionErrors.Inc()
}

// IncrementTimeoutErrors увеличивает счетчик ошибок таймаута
func (m *MASQUEMetrics) IncrementTimeoutErrors() {
	m.TimeoutErrors.Inc()
}

// SetProxyStatus устанавливает статус прокси
func (m *MASQUEMetrics) SetProxyStatus(active bool) {
	if active {
		m.ProxyStatus.Set(1)
	} else {
		m.ProxyStatus.Set(0)
	}
}

// IncrementTargetResolutions увеличивает счетчик разрешений адресов
func (m *MASQUEMetrics) IncrementTargetResolutions() {
	m.TargetResolutions.Inc()
}

// IncrementTargetResolutionErrors увеличивает счетчик ошибок разрешения адресов
func (m *MASQUEMetrics) IncrementTargetResolutionErrors() {
	m.TargetResolutionErrors.Inc()
}
