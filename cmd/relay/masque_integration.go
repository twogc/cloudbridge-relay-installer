package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"net"
	"time"

	"go.uber.org/zap"

	"github.com/twogc/cloudbridge-relay-installer/config"
	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
	"github.com/twogc/cloudbridge-relay-installer/internal/masque"
)

// MASQUEIntegration represents MASQUE integration for the main server
type MASQUEIntegration struct {
	logger         *logger.Logger
	config         *config.Config
	masqueProxy    *masque.MASQUEProxy
	connectIPProxy *masque.ConnectIPProxy
	ctx            context.Context
	cancel         context.CancelFunc
}

// NewMASQUEIntegration creates a new MASQUE integration
func NewMASQUEIntegration(logger *logger.Logger, config *config.Config) (*MASQUEIntegration, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// Create MASQUE proxy configuration
	masqueConfig := &masque.Config{
		ListenAddr: ":8443", // MASQUE HTTP/3 port
		CertFile:   config.Security.CertFile,
		KeyFile:    config.Security.KeyFile,
	}

	// If certificates don't exist, generate self-signed ones for development
	if masqueConfig.CertFile == "" || masqueConfig.KeyFile == "" {
		certFile, keyFile, err := generateSelfSignedCert()
		if err != nil {
			cancel()
			return nil, fmt.Errorf("failed to generate self-signed certificate: %v", err)
		}
		masqueConfig.CertFile = certFile
		masqueConfig.KeyFile = keyFile
	}

	// Create MASQUE proxy
	masqueProxy := masque.NewMASQUEProxy(logger, masqueConfig)

	// Create CONNECT-IP proxy
	connectIPProxy := masque.NewConnectIPProxy(logger.GetZapLogger())

	return &MASQUEIntegration{
		logger:         logger,
		config:         config,
		masqueProxy:    masqueProxy,
		connectIPProxy: connectIPProxy,
		ctx:            ctx,
		cancel:         cancel,
	}, nil
}

// Start starts the MASQUE integration
func (mi *MASQUEIntegration) Start() error {
	mi.logger.Info("Starting MASQUE integration")

	// Start MASQUE proxy
	if err := mi.masqueProxy.Start(); err != nil {
		return fmt.Errorf("failed to start MASQUE proxy: %v", err)
	}

	mi.logger.Info("MASQUE integration started",
		zap.Bool("masque_enabled", mi.config.MASQUE.Enabled),
		zap.String("masque_proxy_url", mi.config.MASQUE.ProxyURL))

	return nil
}

// Stop stops the MASQUE integration
func (mi *MASQUEIntegration) Stop() error {
	mi.logger.Info("Stopping MASQUE integration")

	mi.cancel()

	if mi.masqueProxy != nil {
		if err := mi.masqueProxy.Stop(); err != nil {
			mi.logger.Error("Failed to stop MASQUE proxy", zap.Error(err))
			return err
		}
	}

	mi.logger.Info("MASQUE integration stopped")
	return nil
}

// IsEnabled returns whether MASQUE is enabled
func (mi *MASQUEIntegration) IsEnabled() bool {
	return mi.config.MASQUE.Enabled
}

// GetMASQUEProxy returns the MASQUE proxy instance
func (mi *MASQUEIntegration) GetMASQUEProxy() *masque.MASQUEProxy {
	return mi.masqueProxy
}

// GetConnectIPProxy returns the CONNECT-IP proxy instance
func (mi *MASQUEIntegration) GetConnectIPProxy() *masque.ConnectIPProxy {
	return mi.connectIPProxy
}

// generateSelfSignedCert generates a self-signed certificate for development
func generateSelfSignedCert() (string, string, error) {
	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate private key: %v", err)
	}

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization:  []string{"CloudBridge Relay"},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{""},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour), // Valid for 1 year
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IPAddresses:           []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		DNSNames:              []string{"localhost", "auth.2gc.ru"},
	}

	// Create certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to create certificate: %v", err)
	}

	// Write certificate to temporary file
	certFile := "/tmp/masque_cert.pem"
	keyFile := "/tmp/masque_key.pem"

	// Save certificate
	certPEM := fmt.Sprintf("-----BEGIN CERTIFICATE-----\n%s\n-----END CERTIFICATE-----\n",
		string(certDER))
	if err := writeTempFile(certFile, []byte(certPEM)); err != nil {
		return "", "", fmt.Errorf("failed to write certificate file: %v", err)
	}

	// Save private key
	keyPEM := fmt.Sprintf("-----BEGIN RSA PRIVATE KEY-----\n%s\n-----END RSA PRIVATE KEY-----\n",
		string(x509.MarshalPKCS1PrivateKey(privateKey)))
	if err := writeTempFile(keyFile, []byte(keyPEM)); err != nil {
		return "", "", fmt.Errorf("failed to write key file: %v", err)
	}

	return certFile, keyFile, nil
}

// writeTempFile writes data to a temporary file
func writeTempFile(filename string, data []byte) error {
	// This is a simplified implementation
	// In production, use proper file handling
	return nil // Placeholder - implement proper file writing
}
