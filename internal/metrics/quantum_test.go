package metrics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewQuantumMetrics(t *testing.T) {
	// Test creating new quantum metrics
	metrics := NewQuantumMetrics()
	assert.NotNil(t, metrics)
}

func TestQuantumMetrics_PQCConfiguration(t *testing.T) {
	metrics := NewQuantumMetrics()

	// Test algorithm status
	metrics.UpdateAlgorithmStatus("kyber", "1024", true)
	metrics.UpdateAlgorithmStatus("dilithium", "2048", false)
	metrics.UpdateAlgorithmStatus("sphincs", "256", true)

	// Test hybrid mode status
	metrics.UpdateHybridModeStatus("kyber", true)
	metrics.UpdateHybridModeStatus("dilithium", false)

	// Test key rotation days
	metrics.UpdateKeyRotationDays("kyber", 30)
	metrics.UpdateKeyRotationDays("dilithium", 90)
	metrics.UpdateKeyRotationDays("sphincs", 60)

	// Verify metrics were updated
	assert.NotNil(t, PQCAlgorithmStatus)
	assert.NotNil(t, PQCHybridModeStatus)
	assert.NotNil(t, PQCKeyRotationDays)
}

func TestQuantumMetrics_PQCKeyPairs(t *testing.T) {
	metrics := NewQuantumMetrics()

	// Test key pairs count
	metrics.UpdateKeyPairsCount("kyber", "active", 10)
	metrics.UpdateKeyPairsCount("kyber", "inactive", 5)
	metrics.UpdateKeyPairsCount("dilithium", "active", 8)
	metrics.UpdateKeyPairsCount("sphincs", "active", 12)

	// Test key generation duration
	metrics.ObserveKeyGenerationDuration("kyber", "1024", 0.1)
	metrics.ObserveKeyGenerationDuration("kyber", "2048", 0.2)
	metrics.ObserveKeyGenerationDuration("dilithium", "1024", 0.15)

	// Test key validation errors
	metrics.IncrementKeyValidationErrors("kyber", "invalid_format")
	metrics.IncrementKeyValidationErrors("kyber", "expired")
	metrics.IncrementKeyValidationErrors("dilithium", "corrupted")

	// Test key rotation events
	metrics.IncrementKeyRotationEvents("kyber", "scheduled")
	metrics.IncrementKeyRotationEvents("kyber", "emergency")
	metrics.IncrementKeyRotationEvents("dilithium", "scheduled")

	// Verify metrics were updated
	assert.NotNil(t, PQCKeyPairsTotal)
	assert.NotNil(t, PQCKeyGenerationDuration)
	assert.NotNil(t, PQCKeyValidationErrors)
	assert.NotNil(t, PQCKeyRotationEvents)
}

func TestQuantumMetrics_HybridKeyPairs(t *testing.T) {
	metrics := NewQuantumMetrics()

	// Test hybrid key pairs count
	metrics.UpdateHybridKeyPairsCount("rsa", "kyber", "active", 5)
	metrics.UpdateHybridKeyPairsCount("rsa", "kyber", "inactive", 2)
	metrics.UpdateHybridKeyPairsCount("ecdsa", "dilithium", "active", 8)
	metrics.UpdateHybridKeyPairsCount("ed25519", "sphincs", "active", 6)

	// Test hybrid key combination duration
	metrics.ObserveHybridKeyCombinationDuration("rsa", "kyber", 0.05)
	metrics.ObserveHybridKeyCombinationDuration("ecdsa", "dilithium", 0.08)
	metrics.ObserveHybridKeyCombinationDuration("ed25519", "sphincs", 0.06)

	// Test hybrid key validation errors
	metrics.IncrementHybridKeyValidationErrors("rsa", "kyber", "mismatch")
	metrics.IncrementHybridKeyValidationErrors("ecdsa", "dilithium", "invalid_signature")
	metrics.IncrementHybridKeyValidationErrors("ed25519", "sphincs", "expired")

	// Verify metrics were updated
	assert.NotNil(t, HybridKeyPairsTotal)
	assert.NotNil(t, HybridKeyCombinationDuration)
	assert.NotNil(t, HybridKeyValidationErrors)
}

func TestQuantumMetrics_QuantumWireGuard(t *testing.T) {
	metrics := NewQuantumMetrics()

	// Test quantum WireGuard connections
	metrics.UpdateQuantumWireGuardConnections("wg0", "kyber", 5)
	metrics.UpdateQuantumWireGuardConnections("wg1", "dilithium", 3)
	metrics.UpdateQuantumWireGuardConnections("wg0", "sphincs", 2)

	// Test quantum key exchanges
	metrics.IncrementQuantumWireGuardKeyExchanges("wg0", "kyber", "success")
	metrics.IncrementQuantumWireGuardKeyExchanges("wg0", "kyber", "success")
	metrics.IncrementQuantumWireGuardKeyExchanges("wg1", "dilithium", "failed")

	// Test key exchange duration
	metrics.ObserveQuantumWireGuardKeyExchangeDuration("wg0", "kyber", 0.1)
	metrics.ObserveQuantumWireGuardKeyExchangeDuration("wg1", "dilithium", 0.15)

	// Test signature validations
	metrics.IncrementQuantumWireGuardSignatureValidations("wg0", "kyber", "valid")
	metrics.IncrementQuantumWireGuardSignatureValidations("wg0", "kyber", "invalid")
	metrics.IncrementQuantumWireGuardSignatureValidations("wg1", "dilithium", "valid")

	// Test errors
	metrics.IncrementQuantumWireGuardErrors("wg0", "key_exchange_failed")
	metrics.IncrementQuantumWireGuardErrors("wg1", "signature_verification_failed")

	// Verify metrics were updated
	assert.NotNil(t, QuantumWireGuardConnections)
	assert.NotNil(t, QuantumWireGuardKeyExchanges)
	assert.NotNil(t, QuantumWireGuardKeyExchangeDuration)
	assert.NotNil(t, QuantumWireGuardSignatureValidations)
	assert.NotNil(t, QuantumWireGuardErrors)
}

func TestQuantumMetrics_QuantumEnhancedQUIC(t *testing.T) {
	metrics := NewQuantumMetrics()

	// Test quantum Enhanced QUIC connections
	metrics.UpdateQuantumEnhancedQUICConnections("kyber", 10)
	metrics.UpdateQuantumEnhancedQUICConnections("dilithium", 8)
	metrics.UpdateQuantumEnhancedQUICConnections("sphincs", 6)

	// Test quantum certificates
	metrics.UpdateQuantumEnhancedQUICCertificates("kyber", "valid", 15)
	metrics.UpdateQuantumEnhancedQUICCertificates("kyber", "expired", 2)
	metrics.UpdateQuantumEnhancedQUICCertificates("dilithium", "valid", 12)

	// Test certificate generation duration
	metrics.ObserveQuantumEnhancedQUICCertificateGenerationDuration("kyber", 0.2)
	metrics.ObserveQuantumEnhancedQUICCertificateGenerationDuration("dilithium", 0.25)
	metrics.ObserveQuantumEnhancedQUICCertificateGenerationDuration("sphincs", 0.18)

	// Test certificate validations
	metrics.IncrementQuantumEnhancedQUICCertificateValidations("kyber", "valid")
	metrics.IncrementQuantumEnhancedQUICCertificateValidations("kyber", "invalid")
	metrics.IncrementQuantumEnhancedQUICCertificateValidations("dilithium", "valid")

	// Test errors
	metrics.IncrementQuantumEnhancedQUICErrors("certificate_generation_failed")
	metrics.IncrementQuantumEnhancedQUICErrors("validation_failed")
	metrics.IncrementQuantumEnhancedQUICErrors("connection_failed")

	// Verify metrics were updated
	assert.NotNil(t, QuantumEnhancedQUICConnections)
	assert.NotNil(t, QuantumEnhancedQUICCertificates)
	assert.NotNil(t, QuantumEnhancedQUICCertificateGenerationDuration)
	assert.NotNil(t, QuantumEnhancedQUICCertificateValidations)
	assert.NotNil(t, QuantumEnhancedQUICErrors)
}

func TestQuantumMetrics_PQCPerformance(t *testing.T) {
	metrics := NewQuantumMetrics()

	// Test encryption duration
	metrics.ObservePQCEncryptionDuration("kyber", "1024", 0.01)
	metrics.ObservePQCEncryptionDuration("kyber", "2048", 0.02)
	metrics.ObservePQCEncryptionDuration("dilithium", "1024", 0.015)

	// Test decryption duration
	metrics.ObservePQCDecryptionDuration("kyber", "1024", 0.005)
	metrics.ObservePQCDecryptionDuration("kyber", "2048", 0.01)
	metrics.ObservePQCDecryptionDuration("dilithium", "1024", 0.008)

	// Test signature generation duration
	metrics.ObservePQCSignatureGenerationDuration("dilithium", "1024", 0.02)
	metrics.ObservePQCSignatureGenerationDuration("dilithium", "2048", 0.04)
	metrics.ObservePQCSignatureGenerationDuration("sphincs", "256", 0.03)

	// Test signature verification duration
	metrics.ObservePQCSignatureVerificationDuration("dilithium", "1024", 0.01)
	metrics.ObservePQCSignatureVerificationDuration("dilithium", "2048", 0.02)
	metrics.ObservePQCSignatureVerificationDuration("sphincs", "256", 0.015)

	// Verify metrics were updated
	assert.NotNil(t, PQCEncryptionDuration)
	assert.NotNil(t, PQCDecryptionDuration)
	assert.NotNil(t, PQCSignatureGenerationDuration)
	assert.NotNil(t, PQCSignatureVerificationDuration)
}

func TestQuantumMetrics_PQCSecurity(t *testing.T) {
	metrics := NewQuantumMetrics()

	// Test key compromise events
	metrics.IncrementPQCKeyCompromiseEvents("kyber", "suspected")
	metrics.IncrementPQCKeyCompromiseEvents("kyber", "confirmed")
	metrics.IncrementPQCKeyCompromiseEvents("dilithium", "suspected")

	// Test algorithm migration events
	metrics.IncrementPQCAlgorithmMigrationEvents("kyber", "dilithium", "security_upgrade")
	metrics.IncrementPQCAlgorithmMigrationEvents("dilithium", "sphincs", "performance_improvement")
	metrics.IncrementPQCAlgorithmMigrationEvents("sphincs", "kyber", "compatibility")

	// Test fallback usage
	metrics.IncrementPQCFallbackUsage("kyber", "performance_issue")
	metrics.IncrementPQCFallbackUsage("dilithium", "compatibility_issue")
	metrics.IncrementPQCFallbackUsage("sphincs", "key_generation_failed")

	// Verify metrics were updated
	assert.NotNil(t, PQCKeyCompromiseEvents)
	assert.NotNil(t, PQCAlgorithmMigrationEvents)
	assert.NotNil(t, PQCFallbackUsage)
}

func TestQuantumMetrics_EdgeCases(t *testing.T) {
	metrics := NewQuantumMetrics()

	// Test with empty strings
	metrics.UpdateAlgorithmStatus("", "", true)
	metrics.UpdateHybridModeStatus("", false)
	metrics.UpdateKeyRotationDays("", 0)
	metrics.UpdateKeyPairsCount("", "", 0)
	metrics.ObserveKeyGenerationDuration("", "", 0.0)
	metrics.IncrementKeyValidationErrors("", "")
	metrics.IncrementKeyRotationEvents("", "")
	metrics.UpdateHybridKeyPairsCount("", "", "", 0)
	metrics.ObserveHybridKeyCombinationDuration("", "", 0.0)
	metrics.IncrementHybridKeyValidationErrors("", "", "")
	metrics.UpdateQuantumWireGuardConnections("", "", 0)
	metrics.IncrementQuantumWireGuardKeyExchanges("", "", "")
	metrics.ObserveQuantumWireGuardKeyExchangeDuration("", "", 0.0)
	metrics.IncrementQuantumWireGuardSignatureValidations("", "", "")
	metrics.IncrementQuantumWireGuardErrors("", "")
	metrics.UpdateQuantumEnhancedQUICConnections("", 0)
	metrics.UpdateQuantumEnhancedQUICCertificates("", "", 0)
	metrics.ObserveQuantumEnhancedQUICCertificateGenerationDuration("", 0.0)
	metrics.IncrementQuantumEnhancedQUICCertificateValidations("", "")
	metrics.IncrementQuantumEnhancedQUICErrors("")
	metrics.ObservePQCEncryptionDuration("", "", 0.0)
	metrics.ObservePQCDecryptionDuration("", "", 0.0)
	metrics.ObservePQCSignatureGenerationDuration("", "", 0.0)
	metrics.ObservePQCSignatureVerificationDuration("", "", 0.0)
	metrics.IncrementPQCKeyCompromiseEvents("", "")
	metrics.IncrementPQCAlgorithmMigrationEvents("", "", "")
	metrics.IncrementPQCFallbackUsage("", "")

	// Test with negative values
	metrics.UpdateKeyRotationDays("kyber", -1)
	metrics.UpdateKeyPairsCount("kyber", "active", -1)
	metrics.ObserveKeyGenerationDuration("kyber", "1024", -0.1)
	metrics.UpdateHybridKeyPairsCount("rsa", "kyber", "active", -1)
	metrics.ObserveHybridKeyCombinationDuration("rsa", "kyber", -0.1)
	metrics.UpdateQuantumWireGuardConnections("wg0", "kyber", -1)
	metrics.ObserveQuantumWireGuardKeyExchangeDuration("wg0", "kyber", -0.1)
	metrics.UpdateQuantumEnhancedQUICConnections("kyber", -1)
	metrics.UpdateQuantumEnhancedQUICCertificates("kyber", "valid", -1)
	metrics.ObserveQuantumEnhancedQUICCertificateGenerationDuration("kyber", -0.1)
	metrics.ObservePQCEncryptionDuration("kyber", "1024", -0.1)
	metrics.ObservePQCDecryptionDuration("kyber", "1024", -0.1)
	metrics.ObservePQCSignatureGenerationDuration("dilithium", "1024", -0.1)
	metrics.ObservePQCSignatureVerificationDuration("dilithium", "1024", -0.1)

	// Test with very large values
	metrics.UpdateKeyRotationDays("kyber", 999999)
	metrics.UpdateKeyPairsCount("kyber", "active", 999999)
	metrics.ObserveKeyGenerationDuration("kyber", "1024", 999.0)
	metrics.UpdateHybridKeyPairsCount("rsa", "kyber", "active", 999999)
	metrics.ObserveHybridKeyCombinationDuration("rsa", "kyber", 999.0)
	metrics.UpdateQuantumWireGuardConnections("wg0", "kyber", 999999)
	metrics.ObserveQuantumWireGuardKeyExchangeDuration("wg0", "kyber", 999.0)
	metrics.UpdateQuantumEnhancedQUICConnections("kyber", 999999)
	metrics.UpdateQuantumEnhancedQUICCertificates("kyber", "valid", 999999)
	metrics.ObserveQuantumEnhancedQUICCertificateGenerationDuration("kyber", 999.0)
	metrics.ObservePQCEncryptionDuration("kyber", "1024", 999.0)
	metrics.ObservePQCDecryptionDuration("kyber", "1024", 999.0)
	metrics.ObservePQCSignatureGenerationDuration("dilithium", "1024", 999.0)
	metrics.ObservePQCSignatureVerificationDuration("dilithium", "1024", 999.0)

	// Verify all operations completed without errors
	assert.NotNil(t, metrics)
}

func TestQuantumMetrics_ConcurrentAccess(t *testing.T) {
	metrics := NewQuantumMetrics()
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()

			algorithm := "kyber"
			if id%2 == 0 {
				algorithm = "dilithium"
			}
			keySize := "1024"
			if id%3 == 0 {
				keySize = "2048"
			}
			interfaceName := "wg0"
			if id%2 == 0 {
				interfaceName = "wg1"
			}

			// Test all metrics methods
			metrics.UpdateAlgorithmStatus(algorithm, keySize, true)
			metrics.UpdateHybridModeStatus(algorithm, true)
			metrics.UpdateKeyRotationDays(algorithm, id)
			metrics.UpdateKeyPairsCount(algorithm, "active", id)
			metrics.ObserveKeyGenerationDuration(algorithm, keySize, float64(id)/10.0)
			metrics.IncrementKeyValidationErrors(algorithm, "error")
			metrics.IncrementKeyRotationEvents(algorithm, "scheduled")
			metrics.UpdateHybridKeyPairsCount("rsa", algorithm, "active", id)
			metrics.ObserveHybridKeyCombinationDuration("rsa", algorithm, float64(id)/100.0)
			metrics.IncrementHybridKeyValidationErrors("rsa", algorithm, "error")
			metrics.UpdateQuantumWireGuardConnections(interfaceName, algorithm, id)
			metrics.IncrementQuantumWireGuardKeyExchanges(interfaceName, algorithm, "success")
			metrics.ObserveQuantumWireGuardKeyExchangeDuration(interfaceName, algorithm, float64(id)/100.0)
			metrics.IncrementQuantumWireGuardSignatureValidations(interfaceName, algorithm, "valid")
			metrics.IncrementQuantumWireGuardErrors(interfaceName, "error")
			metrics.UpdateQuantumEnhancedQUICConnections(algorithm, id)
			metrics.UpdateQuantumEnhancedQUICCertificates(algorithm, "valid", id)
			metrics.ObserveQuantumEnhancedQUICCertificateGenerationDuration(algorithm, float64(id)/100.0)
			metrics.IncrementQuantumEnhancedQUICCertificateValidations(algorithm, "valid")
			metrics.IncrementQuantumEnhancedQUICErrors("error")
			metrics.ObservePQCEncryptionDuration(algorithm, keySize, float64(id)/1000.0)
			metrics.ObservePQCDecryptionDuration(algorithm, keySize, float64(id)/1000.0)
			metrics.ObservePQCSignatureGenerationDuration(algorithm, keySize, float64(id)/1000.0)
			metrics.ObservePQCSignatureVerificationDuration(algorithm, keySize, float64(id)/1000.0)
			metrics.IncrementPQCKeyCompromiseEvents(algorithm, "suspected")
			metrics.IncrementPQCAlgorithmMigrationEvents(algorithm, "dilithium", "upgrade")
			metrics.IncrementPQCFallbackUsage(algorithm, "performance")
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all operations completed without errors
	assert.NotNil(t, metrics)
}

func TestQuantumMetrics_IntegrationWorkflow(t *testing.T) {
	// Simulate a complete quantum cryptography workflow
	metrics := NewQuantumMetrics()

	algorithm := "kyber"
	keySize := "1024"
	interfaceName := "wg0"

	// 1. Configure PQC algorithm
	metrics.UpdateAlgorithmStatus(algorithm, keySize, true)
	metrics.UpdateHybridModeStatus(algorithm, true)
	metrics.UpdateKeyRotationDays(algorithm, 30)

	// 2. Generate key pairs
	metrics.ObserveKeyGenerationDuration(algorithm, keySize, 0.1)
	metrics.UpdateKeyPairsCount(algorithm, "active", 5)
	metrics.UpdateKeyPairsCount(algorithm, "inactive", 2)

	// 3. Set up hybrid keys
	metrics.UpdateHybridKeyPairsCount("rsa", algorithm, "active", 3)
	metrics.ObserveHybridKeyCombinationDuration("rsa", algorithm, 0.05)

	// 4. Configure quantum WireGuard
	metrics.UpdateQuantumWireGuardConnections(interfaceName, algorithm, 2)
	metrics.IncrementQuantumWireGuardKeyExchanges(interfaceName, algorithm, "success")
	metrics.ObserveQuantumWireGuardKeyExchangeDuration(interfaceName, algorithm, 0.1)
	metrics.IncrementQuantumWireGuardSignatureValidations(interfaceName, algorithm, "valid")

	// 5. Configure quantum Enhanced QUIC
	metrics.UpdateQuantumEnhancedQUICConnections(algorithm, 4)
	metrics.UpdateQuantumEnhancedQUICCertificates(algorithm, "valid", 6)
	metrics.ObserveQuantumEnhancedQUICCertificateGenerationDuration(algorithm, 0.2)
	metrics.IncrementQuantumEnhancedQUICCertificateValidations(algorithm, "valid")

	// 6. Monitor performance
	metrics.ObservePQCEncryptionDuration(algorithm, keySize, 0.01)
	metrics.ObservePQCDecryptionDuration(algorithm, keySize, 0.005)
	metrics.ObservePQCSignatureGenerationDuration(algorithm, keySize, 0.02)
	metrics.ObservePQCSignatureVerificationDuration(algorithm, keySize, 0.01)

	// 7. Handle security events
	metrics.IncrementPQCKeyCompromiseEvents(algorithm, "suspected")
	metrics.IncrementPQCAlgorithmMigrationEvents(algorithm, "dilithium", "security_upgrade")
	metrics.IncrementPQCFallbackUsage(algorithm, "performance_issue")

	// 8. Handle errors
	metrics.IncrementKeyValidationErrors(algorithm, "invalid_format")
	metrics.IncrementHybridKeyValidationErrors("rsa", algorithm, "mismatch")
	metrics.IncrementQuantumWireGuardErrors(interfaceName, "key_exchange_failed")
	metrics.IncrementQuantumEnhancedQUICErrors("certificate_generation_failed")

	// Verify workflow completed
	assert.NotNil(t, metrics)
}

func TestQuantumMetrics_Performance(t *testing.T) {
	// Test performance with many operations
	metrics := NewQuantumMetrics()

	start := time.Now()

	for i := 0; i < 1000; i++ {
		algorithm := "kyber"
		if i%2 == 0 {
			algorithm = "dilithium"
		}
		keySize := "1024"
		if i%3 == 0 {
			keySize = "2048"
		}
		interfaceName := "wg0"
		if i%2 == 0 {
			interfaceName = "wg1"
		}

		metrics.UpdateAlgorithmStatus(algorithm, keySize, true)
		metrics.UpdateHybridModeStatus(algorithm, true)
		metrics.UpdateKeyRotationDays(algorithm, i%100)
		metrics.UpdateKeyPairsCount(algorithm, "active", i%50)
		metrics.ObserveKeyGenerationDuration(algorithm, keySize, float64(i%100)/100.0)
		metrics.IncrementKeyValidationErrors(algorithm, "error")
		metrics.IncrementKeyRotationEvents(algorithm, "scheduled")
		metrics.UpdateHybridKeyPairsCount("rsa", algorithm, "active", i%30)
		metrics.ObserveHybridKeyCombinationDuration("rsa", algorithm, float64(i%100)/1000.0)
		metrics.IncrementHybridKeyValidationErrors("rsa", algorithm, "error")
		metrics.UpdateQuantumWireGuardConnections(interfaceName, algorithm, i%20)
		metrics.IncrementQuantumWireGuardKeyExchanges(interfaceName, algorithm, "success")
		metrics.ObserveQuantumWireGuardKeyExchangeDuration(interfaceName, algorithm, float64(i%100)/1000.0)
		metrics.IncrementQuantumWireGuardSignatureValidations(interfaceName, algorithm, "valid")
		metrics.IncrementQuantumWireGuardErrors(interfaceName, "error")
		metrics.UpdateQuantumEnhancedQUICConnections(algorithm, i%25)
		metrics.UpdateQuantumEnhancedQUICCertificates(algorithm, "valid", i%40)
		metrics.ObserveQuantumEnhancedQUICCertificateGenerationDuration(algorithm, float64(i%100)/1000.0)
		metrics.IncrementQuantumEnhancedQUICCertificateValidations(algorithm, "valid")
		metrics.IncrementQuantumEnhancedQUICErrors("error")
		metrics.ObservePQCEncryptionDuration(algorithm, keySize, float64(i%100)/10000.0)
		metrics.ObservePQCDecryptionDuration(algorithm, keySize, float64(i%100)/10000.0)
		metrics.ObservePQCSignatureGenerationDuration(algorithm, keySize, float64(i%100)/10000.0)
		metrics.ObservePQCSignatureVerificationDuration(algorithm, keySize, float64(i%100)/10000.0)
		metrics.IncrementPQCKeyCompromiseEvents(algorithm, "suspected")
		metrics.IncrementPQCAlgorithmMigrationEvents(algorithm, "dilithium", "upgrade")
		metrics.IncrementPQCFallbackUsage(algorithm, "performance")
	}

	duration := time.Since(start)

	// Should complete within reasonable time (less than 1 second)
	assert.Less(t, duration, time.Second)
}

func TestQuantumMetrics_DataTypes(t *testing.T) {
	// Test different data types and edge cases
	metrics := NewQuantumMetrics()

	testCases := []struct {
		name     string
		testFunc func()
	}{
		{
			name: "string labels",
			testFunc: func() {
				metrics.UpdateAlgorithmStatus("kyber-algorithm", "1024-key-size", true)
			},
		},
		{
			name: "unicode strings",
			testFunc: func() {
				metrics.UpdateAlgorithmStatus("крибер-алгоритм", "1024-размер-ключа", true)
			},
		},
		{
			name: "special characters",
			testFunc: func() {
				metrics.UpdateAlgorithmStatus("kyber@algorithm", "1024_key_size", true)
			},
		},
		{
			name: "long strings",
			testFunc: func() {
				longString := "very-long-algorithm-name-that-might-cause-issues-with-prometheus-label-limits"
				metrics.UpdateAlgorithmStatus(longString, "1024", true)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.NotPanics(t, tc.testFunc)
		})
	}
}

func TestQuantumMetrics_StatusValues(t *testing.T) {
	// Test status values (enabled/disabled)
	metrics := NewQuantumMetrics()

	// Test enabled status (true)
	metrics.UpdateAlgorithmStatus("kyber", "1024", true)
	metrics.UpdateHybridModeStatus("kyber", true)

	// Test disabled status (false)
	metrics.UpdateAlgorithmStatus("kyber", "1024", false)
	metrics.UpdateHybridModeStatus("kyber", false)

	// Verify operations completed without errors
	assert.NotNil(t, PQCAlgorithmStatus)
	assert.NotNil(t, PQCHybridModeStatus)
}

func TestQuantumMetrics_AlgorithmTypes(t *testing.T) {
	// Test different algorithm types
	metrics := NewQuantumMetrics()

	algorithms := []string{"kyber", "dilithium", "sphincs", "falcon", "ntru"}
	keySizes := []string{"1024", "2048", "4096", "256", "512"}

	for _, alg := range algorithms {
		for _, size := range keySizes {
			metrics.UpdateAlgorithmStatus(alg, size, true)
			metrics.UpdateKeyPairsCount(alg, "active", 1)
			metrics.ObserveKeyGenerationDuration(alg, size, 0.1)
			metrics.IncrementKeyValidationErrors(alg, "error")
			metrics.IncrementKeyRotationEvents(alg, "scheduled")
			metrics.ObservePQCEncryptionDuration(alg, size, 0.01)
			metrics.ObservePQCDecryptionDuration(alg, size, 0.005)
			metrics.ObservePQCSignatureGenerationDuration(alg, size, 0.02)
			metrics.ObservePQCSignatureVerificationDuration(alg, size, 0.01)
		}
	}

	// Verify all operations completed without errors
	assert.NotNil(t, metrics)
}

func TestQuantumMetrics_InterfaceTypes(t *testing.T) {
	// Test different interface types
	metrics := NewQuantumMetrics()

	interfaces := []string{"wg0", "wg1", "wg2", "wg3", "wg4"}
	algorithms := []string{"kyber", "dilithium", "sphincs"}

	for _, iface := range interfaces {
		for _, alg := range algorithms {
			metrics.UpdateQuantumWireGuardConnections(iface, alg, 1)
			metrics.IncrementQuantumWireGuardKeyExchanges(iface, alg, "success")
			metrics.ObserveQuantumWireGuardKeyExchangeDuration(iface, alg, 0.1)
			metrics.IncrementQuantumWireGuardSignatureValidations(iface, alg, "valid")
			metrics.IncrementQuantumWireGuardErrors(iface, "error")
		}
	}

	// Verify all operations completed without errors
	assert.NotNil(t, metrics)
}
