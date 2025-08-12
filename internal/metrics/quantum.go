package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// PQC Configuration Metrics
	PQCAlgorithmStatus = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "pqc_algorithm_status",
			Help: "Post-Quantum Cryptography algorithm status (1 = enabled, 0 = disabled)",
		},
		[]string{"algorithm", "key_size"},
	)

	PQCHybridModeStatus = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "pqc_hybrid_mode_status",
			Help: "PQC hybrid mode status (1 = enabled, 0 = disabled)",
		},
		[]string{"algorithm"},
	)

	PQCKeyRotationDays = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "pqc_key_rotation_days",
			Help: "Number of days until next key rotation",
		},
		[]string{"algorithm"},
	)

	// PQC Key Pair Metrics
	PQCKeyPairsTotal = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "pqc_key_pairs_total",
			Help: "Total number of PQC key pairs",
		},
		[]string{"algorithm", "status"},
	)

	PQCKeyGenerationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "pqc_key_generation_duration_seconds",
			Help:    "Time taken to generate PQC key pairs",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"algorithm", "key_size"},
	)

	PQCKeyValidationErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pqc_key_validation_errors_total",
			Help: "Total number of PQC key validation errors",
		},
		[]string{"algorithm", "error_type"},
	)

	PQCKeyRotationEvents = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pqc_key_rotation_events_total",
			Help: "Total number of PQC key rotation events",
		},
		[]string{"algorithm", "trigger"},
	)

	// Hybrid Key Pair Metrics
	HybridKeyPairsTotal = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "hybrid_key_pairs_total",
			Help: "Total number of hybrid key pairs (Classical + Quantum)",
		},
		[]string{"classical_algorithm", "quantum_algorithm", "status"},
	)

	HybridKeyCombinationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "hybrid_key_combination_duration_seconds",
			Help:    "Time taken to combine classical and quantum keys",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"classical_algorithm", "quantum_algorithm"},
	)

	HybridKeyValidationErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "hybrid_key_validation_errors_total",
			Help: "Total number of hybrid key validation errors",
		},
		[]string{"classical_algorithm", "quantum_algorithm", "error_type"},
	)

	// Quantum WireGuard Metrics
	QuantumWireGuardConnections = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "quantum_wireguard_connections",
			Help: "Number of active quantum-secured WireGuard connections",
		},
		[]string{"interface", "algorithm"},
	)

	QuantumWireGuardKeyExchanges = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "quantum_wireguard_key_exchanges_total",
			Help: "Total number of quantum key exchanges in WireGuard",
		},
		[]string{"interface", "algorithm", "status"},
	)

	QuantumWireGuardKeyExchangeDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "quantum_wireguard_key_exchange_duration_seconds",
			Help:    "Time taken for quantum key exchange in WireGuard",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"interface", "algorithm"},
	)

	QuantumWireGuardSignatureValidations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "quantum_wireguard_signature_validations_total",
			Help: "Total number of quantum signature validations in WireGuard",
		},
		[]string{"interface", "algorithm", "status"},
	)

	QuantumWireGuardErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "quantum_wireguard_errors_total",
			Help: "Total number of quantum WireGuard errors",
		},
		[]string{"interface", "error_type"},
	)

	// Quantum Enhanced QUIC Metrics
	QuantumEnhancedQUICConnections = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "quantum_enhanced_quic_connections",
			Help: "Number of active quantum-secured Enhanced QUIC connections",
		},
		[]string{"algorithm"},
	)

	QuantumEnhancedQUICCertificates = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "quantum_enhanced_quic_certificates",
			Help: "Number of quantum certificates for Enhanced QUIC",
		},
		[]string{"algorithm", "status"},
	)

	QuantumEnhancedQUICCertificateGenerationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "quantum_enhanced_quic_certificate_generation_duration_seconds",
			Help:    "Time taken to generate quantum certificates for Enhanced QUIC",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"algorithm"},
	)

	QuantumEnhancedQUICCertificateValidations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "quantum_enhanced_quic_certificate_validations_total",
			Help: "Total number of quantum certificate validations in Enhanced QUIC",
		},
		[]string{"algorithm", "status"},
	)

	QuantumEnhancedQUICErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "quantum_enhanced_quic_errors_total",
			Help: "Total number of quantum Enhanced QUIC errors",
		},
		[]string{"error_type"},
	)

	// PQC Algorithm Performance Metrics
	PQCEncryptionDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "pqc_encryption_duration_seconds",
			Help:    "Time taken for PQC encryption operations",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"algorithm", "key_size"},
	)

	PQCDecryptionDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "pqc_decryption_duration_seconds",
			Help:    "Time taken for PQC decryption operations",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"algorithm", "key_size"},
	)

	PQCSignatureGenerationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "pqc_signature_generation_duration_seconds",
			Help:    "Time taken for PQC signature generation",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"algorithm", "key_size"},
	)

	PQCSignatureVerificationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "pqc_signature_verification_duration_seconds",
			Help:    "Time taken for PQC signature verification",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"algorithm", "key_size"},
	)

	// PQC Security Metrics
	PQCKeyCompromiseEvents = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pqc_key_compromise_events_total",
			Help: "Total number of PQC key compromise events",
		},
		[]string{"algorithm", "compromise_type"},
	)

	PQCAlgorithmMigrationEvents = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pqc_algorithm_migration_events_total",
			Help: "Total number of PQC algorithm migration events",
		},
		[]string{"from_algorithm", "to_algorithm", "trigger"},
	)

	PQCFallbackUsage = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pqc_fallback_usage_total",
			Help: "Total number of PQC fallback to classical cryptography",
		},
		[]string{"algorithm", "fallback_reason"},
	)
)

// QuantumMetrics предоставляет методы для обновления метрик Post-Quantum Cryptography
type QuantumMetrics struct{}

// NewQuantumMetrics создает новый экземпляр QuantumMetrics
func NewQuantumMetrics() *QuantumMetrics {
	return &QuantumMetrics{}
}

// UpdateAlgorithmStatus обновляет статус алгоритма
func (qm *QuantumMetrics) UpdateAlgorithmStatus(algorithm, keySize string, enabled bool) {
	status := 0.0
	if enabled {
		status = 1.0
	}
	PQCAlgorithmStatus.WithLabelValues(algorithm, keySize).Set(status)
}

// UpdateHybridModeStatus обновляет статус гибридного режима
func (qm *QuantumMetrics) UpdateHybridModeStatus(algorithm string, enabled bool) {
	status := 0.0
	if enabled {
		status = 1.0
	}
	PQCHybridModeStatus.WithLabelValues(algorithm).Set(status)
}

// UpdateKeyRotationDays обновляет количество дней до ротации ключей
func (qm *QuantumMetrics) UpdateKeyRotationDays(algorithm string, days int) {
	PQCKeyRotationDays.WithLabelValues(algorithm).Set(float64(days))
}

// UpdateKeyPairsCount обновляет количество пар ключей
func (qm *QuantumMetrics) UpdateKeyPairsCount(algorithm, status string, count int) {
	PQCKeyPairsTotal.WithLabelValues(algorithm, status).Set(float64(count))
}

// ObserveKeyGenerationDuration наблюдает время генерации ключей
func (qm *QuantumMetrics) ObserveKeyGenerationDuration(algorithm, keySize string, duration float64) {
	PQCKeyGenerationDuration.WithLabelValues(algorithm, keySize).Observe(duration)
}

// IncrementKeyValidationErrors увеличивает счетчик ошибок валидации ключей
func (qm *QuantumMetrics) IncrementKeyValidationErrors(algorithm, errorType string) {
	PQCKeyValidationErrors.WithLabelValues(algorithm, errorType).Inc()
}

// IncrementKeyRotationEvents увеличивает счетчик событий ротации ключей
func (qm *QuantumMetrics) IncrementKeyRotationEvents(algorithm, trigger string) {
	PQCKeyRotationEvents.WithLabelValues(algorithm, trigger).Inc()
}

// UpdateHybridKeyPairsCount обновляет количество гибридных пар ключей
func (qm *QuantumMetrics) UpdateHybridKeyPairsCount(classicalAlg, quantumAlg, status string, count int) {
	HybridKeyPairsTotal.WithLabelValues(classicalAlg, quantumAlg, status).Set(float64(count))
}

// ObserveHybridKeyCombinationDuration наблюдает время комбинирования ключей
func (qm *QuantumMetrics) ObserveHybridKeyCombinationDuration(classicalAlg, quantumAlg string, duration float64) {
	HybridKeyCombinationDuration.WithLabelValues(classicalAlg, quantumAlg).Observe(duration)
}

// IncrementHybridKeyValidationErrors увеличивает счетчик ошибок валидации гибридных ключей
func (qm *QuantumMetrics) IncrementHybridKeyValidationErrors(classicalAlg, quantumAlg, errorType string) {
	HybridKeyValidationErrors.WithLabelValues(classicalAlg, quantumAlg, errorType).Inc()
}

// UpdateQuantumWireGuardConnections обновляет количество квантовых WireGuard соединений
func (qm *QuantumMetrics) UpdateQuantumWireGuardConnections(interfaceName, algorithm string, count int) {
	QuantumWireGuardConnections.WithLabelValues(interfaceName, algorithm).Set(float64(count))
}

// IncrementQuantumWireGuardKeyExchanges увеличивает счетчик квантовых обменов ключей
func (qm *QuantumMetrics) IncrementQuantumWireGuardKeyExchanges(interfaceName, algorithm, status string) {
	QuantumWireGuardKeyExchanges.WithLabelValues(interfaceName, algorithm, status).Inc()
}

// ObserveQuantumWireGuardKeyExchangeDuration наблюдает время квантового обмена ключей
func (qm *QuantumMetrics) ObserveQuantumWireGuardKeyExchangeDuration(interfaceName, algorithm string, duration float64) {
	QuantumWireGuardKeyExchangeDuration.WithLabelValues(interfaceName, algorithm).Observe(duration)
}

// IncrementQuantumWireGuardSignatureValidations увеличивает счетчик валидаций подписей
func (qm *QuantumMetrics) IncrementQuantumWireGuardSignatureValidations(interfaceName, algorithm, status string) {
	QuantumWireGuardSignatureValidations.WithLabelValues(interfaceName, algorithm, status).Inc()
}

// IncrementQuantumWireGuardErrors увеличивает счетчик ошибок квантового WireGuard
func (qm *QuantumMetrics) IncrementQuantumWireGuardErrors(interfaceName, errorType string) {
	QuantumWireGuardErrors.WithLabelValues(interfaceName, errorType).Inc()
}

// UpdateQuantumEnhancedQUICConnections обновляет количество квантовых Enhanced QUIC соединений
func (qm *QuantumMetrics) UpdateQuantumEnhancedQUICConnections(algorithm string, count int) {
	QuantumEnhancedQUICConnections.WithLabelValues(algorithm).Set(float64(count))
}

// UpdateQuantumEnhancedQUICCertificates обновляет количество квантовых сертификатов
func (qm *QuantumMetrics) UpdateQuantumEnhancedQUICCertificates(algorithm, status string, count int) {
	QuantumEnhancedQUICCertificates.WithLabelValues(algorithm, status).Set(float64(count))
}

// ObserveQuantumEnhancedQUICCertificateGenerationDuration наблюдает время генерации сертификатов
func (qm *QuantumMetrics) ObserveQuantumEnhancedQUICCertificateGenerationDuration(algorithm string, duration float64) {
	QuantumEnhancedQUICCertificateGenerationDuration.WithLabelValues(algorithm).Observe(duration)
}

// IncrementQuantumEnhancedQUICCertificateValidations увеличивает счетчик валидаций сертификатов
func (qm *QuantumMetrics) IncrementQuantumEnhancedQUICCertificateValidations(algorithm, status string) {
	QuantumEnhancedQUICCertificateValidations.WithLabelValues(algorithm, status).Inc()
}

// IncrementQuantumEnhancedQUICErrors увеличивает счетчик ошибок квантового Enhanced QUIC
func (qm *QuantumMetrics) IncrementQuantumEnhancedQUICErrors(errorType string) {
	QuantumEnhancedQUICErrors.WithLabelValues(errorType).Inc()
}

// ObservePQCEncryptionDuration наблюдает время PQC шифрования
func (qm *QuantumMetrics) ObservePQCEncryptionDuration(algorithm, keySize string, duration float64) {
	PQCEncryptionDuration.WithLabelValues(algorithm, keySize).Observe(duration)
}

// ObservePQCDecryptionDuration наблюдает время PQC расшифрования
func (qm *QuantumMetrics) ObservePQCDecryptionDuration(algorithm, keySize string, duration float64) {
	PQCDecryptionDuration.WithLabelValues(algorithm, keySize).Observe(duration)
}

// ObservePQCSignatureGenerationDuration наблюдает время генерации PQC подписи
func (qm *QuantumMetrics) ObservePQCSignatureGenerationDuration(algorithm, keySize string, duration float64) {
	PQCSignatureGenerationDuration.WithLabelValues(algorithm, keySize).Observe(duration)
}

// ObservePQCSignatureVerificationDuration наблюдает время верификации PQC подписи
func (qm *QuantumMetrics) ObservePQCSignatureVerificationDuration(algorithm, keySize string, duration float64) {
	PQCSignatureVerificationDuration.WithLabelValues(algorithm, keySize).Observe(duration)
}

// IncrementPQCKeyCompromiseEvents увеличивает счетчик событий компрометации ключей
func (qm *QuantumMetrics) IncrementPQCKeyCompromiseEvents(algorithm, compromiseType string) {
	PQCKeyCompromiseEvents.WithLabelValues(algorithm, compromiseType).Inc()
}

// IncrementPQCAlgorithmMigrationEvents увеличивает счетчик событий миграции алгоритмов
func (qm *QuantumMetrics) IncrementPQCAlgorithmMigrationEvents(fromAlg, toAlg, trigger string) {
	PQCAlgorithmMigrationEvents.WithLabelValues(fromAlg, toAlg, trigger).Inc()
}

// IncrementPQCFallbackUsage увеличивает счетчик использования fallback
func (qm *QuantumMetrics) IncrementPQCFallbackUsage(algorithm, fallbackReason string) {
	PQCFallbackUsage.WithLabelValues(algorithm, fallbackReason).Inc()
}
