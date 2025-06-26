package tls

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"strings"
)

// Config represents TLS configuration
type Config struct {
	Enabled      bool
	CertFile     string
	KeyFile      string
	MinVersion   string
	CipherSuites []string
	MTLS         MTLSConfig
}

// MTLSConfig represents mTLS configuration
type MTLSConfig struct {
	Enabled     bool
	CACert      string
	ClientCert  string
	ClientKey   string
}

// NewTLSConfig creates a new TLS configuration
func NewTLSConfig(config Config) (*tls.Config, error) {
	if !config.Enabled {
		return nil, nil
	}

	// Load certificate
	cert, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS certificate: %v", err)
	}

	// Create TLS config
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   getTLSVersion(config.MinVersion),
		CipherSuites: getCipherSuites(config.CipherSuites),
	}

	// Configure mTLS if enabled
	if config.MTLS.Enabled {
		if err := configureMTLS(tlsConfig, config.MTLS); err != nil {
			return nil, fmt.Errorf("failed to configure mTLS: %v", err)
		}
	}

	return tlsConfig, nil
}

// configureMTLS configures mutual TLS
func configureMTLS(tlsConfig *tls.Config, mtlsConfig MTLSConfig) error {
	// Load CA certificate
	caCert, err := os.ReadFile(mtlsConfig.CACert)
	if err != nil {
		return fmt.Errorf("failed to read CA certificate: %v", err)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return fmt.Errorf("failed to append CA certificate")
	}

	// Load client certificate
	clientCert, err := tls.LoadX509KeyPair(mtlsConfig.ClientCert, mtlsConfig.ClientKey)
	if err != nil {
		return fmt.Errorf("failed to load client certificate: %v", err)
	}

	tlsConfig.ClientCAs = caCertPool
	tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
	tlsConfig.Certificates = append(tlsConfig.Certificates, clientCert)

	return nil
}

// getTLSVersion converts string version to tls.Version
func getTLSVersion(version string) uint16 {
	switch strings.ToUpper(version) {
	case "1.0":
		return tls.VersionTLS10
	case "1.1":
		return tls.VersionTLS11
	case "1.2":
		return tls.VersionTLS12
	case "1.3":
		return tls.VersionTLS13
	default:
		return tls.VersionTLS12
	}
}

// getCipherSuites converts string cipher suites to tls.CipherSuite
func getCipherSuites(cipherSuites []string) []uint16 {
	var suites []uint16
	cipherMap := map[string]uint16{
		"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384":   tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256":   tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384": tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256": tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		"TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305":    tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		"TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305":  tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
	}

	for _, cipher := range cipherSuites {
		if suite, ok := cipherMap[cipher]; ok {
			suites = append(suites, suite)
		}
	}

	return suites
}

// ValidateCertificate validates TLS certificate
func ValidateCertificate(certFile, keyFile string) error {
	_, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return fmt.Errorf("invalid certificate: %v", err)
	}
	return nil
}

// GetCertificateInfo returns certificate information
func GetCertificateInfo(certFile string) (map[string]interface{}, error) {
	certPEM, err := os.ReadFile(certFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate: %v", err)
	}

	block, _ := pem.Decode(certPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode certificate")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %v", err)
	}

	return map[string]interface{}{
		"subject":      cert.Subject.String(),
		"issuer":       cert.Issuer.String(),
		"not_before":   cert.NotBefore,
		"not_after":    cert.NotAfter,
		"serial_number": cert.SerialNumber.String(),
		"dns_names":    cert.DNSNames,
		"ip_addresses": cert.IPAddresses,
	}, nil
} 