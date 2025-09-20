package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// TLSConfig represents TLS configuration
type TLSConfig struct {
	Enabled       bool       `json:"tls_enabled"`
	CertFile      string     `json:"cert_file"`
	KeyFile       string     `json:"key_file"`
	CipherSuites  []string   `json:"cipher_suites"`
	MinTLSVersion string     `json:"min_tls_version"`
	MTLS          MTLSConfig `json:"mtls"`
}

// MTLSConfig represents mutual TLS configuration
type MTLSConfig struct {
	Enabled            bool   `json:"enabled"`
	CAFile             string `json:"ca_file"`
	ClientCertRequired bool   `json:"client_cert_required"`
}

// TLSServer handles TLS server configuration and startup
type TLSServer struct {
	logger *zap.Logger
	config TLSConfig
}

// NewTLSServer creates a new TLS server instance
func NewTLSServer(config TLSConfig, logger *zap.Logger) *TLSServer {
	return &TLSServer{
		logger: logger,
		config: config,
	}
}

// CreateTLSConfig creates TLS configuration based on the provided settings
func (ts *TLSServer) CreateTLSConfig() (*tls.Config, error) {
	tlsConfig := &tls.Config{
		MinVersion:               ts.getMinTLSVersion(), // #nosec G402 -- getMinTLSVersion() returns safe TLS version
		CipherSuites:             ts.getCipherSuites(),
		PreferServerCipherSuites: true,
		SessionTicketsDisabled:   false,
		NextProtos:               []string{"h2", "http/1.1"},
	}

	// Configure mTLS if enabled
	if ts.config.MTLS.Enabled {
		if err := ts.configureMTLS(tlsConfig); err != nil {
			return nil, fmt.Errorf("failed to configure mTLS: %v", err)
		}
	}

	return tlsConfig, nil
}

// StartHTTPServer starts HTTPS server with TLS configuration
func (ts *TLSServer) StartHTTPServer(addr string, handler http.Handler) (*http.Server, error) {
	tlsConfig, err := ts.CreateTLSConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create TLS config: %v", err)
	}

	server := &http.Server{
		Addr:         addr,
		Handler:      handler,
		TLSConfig:    tlsConfig,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	ts.logger.Info("Starting HTTPS server",
		zap.String("addr", addr),
		zap.String("cert_file", ts.config.CertFile),
		zap.String("key_file", ts.config.KeyFile),
		zap.Bool("mtls_enabled", ts.config.MTLS.Enabled),
		zap.String("min_tls_version", ts.config.MinTLSVersion))

	return server, nil
}

// StartHTTPSServer starts HTTPS server with TLS
func (ts *TLSServer) StartHTTPSServer(addr string, handler http.Handler) error {
	server, err := ts.StartHTTPServer(addr, handler)
	if err != nil {
		return err
	}

	return server.ListenAndServeTLS(ts.config.CertFile, ts.config.KeyFile)
}

// configureMTLS configures mutual TLS authentication
func (ts *TLSServer) configureMTLS(tlsConfig *tls.Config) error {
	// Set client authentication mode
	if ts.config.MTLS.ClientCertRequired {
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
	} else {
		tlsConfig.ClientAuth = tls.VerifyClientCertIfGiven
	}

	// Load CA certificate
	caCert, err := ioutil.ReadFile(ts.config.MTLS.CAFile)
	if err != nil {
		return fmt.Errorf("failed to read CA certificate from %s: %v", ts.config.MTLS.CAFile, err)
	}

	// Create certificate pool and add CA certificate
	tlsConfig.ClientCAs = x509.NewCertPool()
	if !tlsConfig.ClientCAs.AppendCertsFromPEM(caCert) {
		return fmt.Errorf("failed to parse CA certificate from %s", ts.config.MTLS.CAFile)
	}

	ts.logger.Info("mTLS configured successfully",
		zap.String("ca_file", ts.config.MTLS.CAFile),
		zap.Bool("client_cert_required", ts.config.MTLS.ClientCertRequired))

	return nil
}

// getMinTLSVersion returns the minimum TLS version
func (ts *TLSServer) getMinTLSVersion() uint16 {
	switch ts.config.MinTLSVersion {
	case "1.0":
		return tls.VersionTLS10
	case "1.1":
		return tls.VersionTLS11
	case "1.2":
		return tls.VersionTLS12
	case "1.3":
		return tls.VersionTLS13
	default:
		ts.logger.Warn("Unknown TLS version, defaulting to TLS 1.2", zap.String("version", ts.config.MinTLSVersion))
		return tls.VersionTLS12
	}
}

// getCipherSuites returns the list of cipher suites
func (ts *TLSServer) getCipherSuites() []uint16 {
	var cipherSuites []uint16

	for _, suite := range ts.config.CipherSuites {
		switch suite {
		case "TLS_AES_256_GCM_SHA384":
			cipherSuites = append(cipherSuites, tls.TLS_AES_256_GCM_SHA384)
		case "TLS_CHACHA20_POLY1305_SHA256":
			cipherSuites = append(cipherSuites, tls.TLS_CHACHA20_POLY1305_SHA256)
		case "TLS_AES_128_GCM_SHA256":
			cipherSuites = append(cipherSuites, tls.TLS_AES_128_GCM_SHA256)
		case "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384":
			cipherSuites = append(cipherSuites, tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384)
		case "TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305":
			cipherSuites = append(cipherSuites, tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305)
		case "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256":
			cipherSuites = append(cipherSuites, tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256)
		default:
			ts.logger.Warn("Unknown cipher suite, skipping", zap.String("suite", suite))
		}
	}

	// If no cipher suites specified, use default secure ones
	if len(cipherSuites) == 0 {
		cipherSuites = []uint16{
			tls.TLS_AES_256_GCM_SHA384,
			tls.TLS_CHACHA20_POLY1305_SHA256,
			tls.TLS_AES_128_GCM_SHA256,
		}
	}

	return cipherSuites
}

// ValidateTLSConfig validates TLS configuration
func (ts *TLSServer) ValidateTLSConfig() error {
	if !ts.config.Enabled {
		return nil // TLS is disabled, no validation needed
	}

	// Check if certificate files exist
	if ts.config.CertFile == "" {
		return fmt.Errorf("cert_file is required when TLS is enabled")
	}
	if ts.config.KeyFile == "" {
		return fmt.Errorf("key_file is required when TLS is enabled")
	}

	// Validate certificate files
	if err := ts.validateCertificateFile(ts.config.CertFile); err != nil {
		return fmt.Errorf("invalid certificate file %s: %v", ts.config.CertFile, err)
	}
	if err := ts.validatePrivateKeyFile(ts.config.KeyFile); err != nil {
		return fmt.Errorf("invalid private key file %s: %v", ts.config.KeyFile, err)
	}

	// Validate mTLS configuration if enabled
	if ts.config.MTLS.Enabled {
		if ts.config.MTLS.CAFile == "" {
			return fmt.Errorf("ca_file is required when mTLS is enabled")
		}
		if err := ts.validateCertificateFile(ts.config.MTLS.CAFile); err != nil {
			return fmt.Errorf("invalid CA certificate file %s: %v", ts.config.MTLS.CAFile, err)
		}
	}

	ts.logger.Info("TLS configuration validated successfully")
	return nil
}

// validateCertificateFile validates a certificate file
func (ts *TLSServer) validateCertificateFile(filename string) error {
	certPEM, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read certificate file: %v", err)
	}

	block, _ := pem.Decode(certPEM)
	if block == nil {
		return fmt.Errorf("failed to decode PEM block from certificate file")
	}

	if block.Type != "CERTIFICATE" {
		return fmt.Errorf("invalid certificate type: %s", block.Type)
	}

	_, err = x509.ParseCertificate(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse certificate: %v", err)
	}

	return nil
}

// validatePrivateKeyFile validates a private key file
func (ts *TLSServer) validatePrivateKeyFile(filename string) error {
	keyPEM, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read private key file: %v", err)
	}

	block, _ := pem.Decode(keyPEM)
	if block == nil {
		return fmt.Errorf("failed to decode PEM block from private key file")
	}

	switch block.Type {
	case "RSA PRIVATE KEY":
		_, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	case "PRIVATE KEY":
		_, err = x509.ParsePKCS8PrivateKey(block.Bytes)
	case "EC PRIVATE KEY":
		_, err = x509.ParseECPrivateKey(block.Bytes)
	default:
		return fmt.Errorf("unsupported private key type: %s", block.Type)
	}

	if err != nil {
		return fmt.Errorf("failed to parse private key: %v", err)
	}

	return nil
}

// GetTLSInfo returns information about the TLS configuration
func (ts *TLSServer) GetTLSInfo() map[string]interface{} {
	info := map[string]interface{}{
		"enabled":              ts.config.Enabled,
		"cert_file":            ts.config.CertFile,
		"key_file":             ts.config.KeyFile,
		"min_tls_version":      ts.config.MinTLSVersion,
		"cipher_suites":        ts.config.CipherSuites,
		"mtls_enabled":         ts.config.MTLS.Enabled,
		"client_cert_required": ts.config.MTLS.ClientCertRequired,
	}

	if ts.config.MTLS.Enabled {
		info["ca_file"] = ts.config.MTLS.CAFile
	}

	return info
}
