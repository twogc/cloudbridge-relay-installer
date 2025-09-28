package api

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"time"

	"go.uber.org/zap"
)

// TLSConfig конфигурация для TLS/HTTPS
type TLSConfig struct {
	// Certificates
	CertFile string `json:"cert_file"`
	KeyFile  string `json:"key_file"`
	
	// TLS settings
	MinVersion       uint16   `json:"min_version"`        // tls.VersionTLS12
	MaxVersion       uint16   `json:"max_version"`        // tls.VersionTLS13
	CipherSuites     []uint16 `json:"cipher_suites"`
	CurvePreferences []tls.CurveID `json:"curve_preferences"`
	
	// Security settings
	PreferServerCipherSuites bool `json:"prefer_server_cipher_suites"`
	InsecureSkipVerify       bool `json:"insecure_skip_verify"`
	
	// OCSP and certificate settings
	NextProtos []string `json:"next_protos"`
	
	// Auto-redirect settings
	HTTPPort  int  `json:"http_port"`
	HTTPSPort int  `json:"https_port"`
	AutoRedirect bool `json:"auto_redirect"`
}

// DefaultTLSConfig возвращает безопасную конфигурацию TLS по умолчанию
func DefaultTLSConfig() *TLSConfig {
	return &TLSConfig{
		CertFile:   "/etc/ssl/certs/cloudbridge.crt",
		KeyFile:    "/etc/ssl/private/cloudbridge.key",
		MinVersion: tls.VersionTLS12,
		MaxVersion: tls.VersionTLS13,
		CipherSuites: []uint16{
			// TLS 1.3 cipher suites (предпочтительные)
			tls.TLS_AES_256_GCM_SHA384,
			tls.TLS_AES_128_GCM_SHA256,
			tls.TLS_CHACHA20_POLY1305_SHA256,
			
			// TLS 1.2 cipher suites (для совместимости)
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		},
		CurvePreferences: []tls.CurveID{
			tls.X25519,
			tls.CurveP256,
			tls.CurveP384,
		},
		PreferServerCipherSuites: true,
		InsecureSkipVerify:       false,
		NextProtos: []string{
			"h2",    // HTTP/2
			"http/1.1", // HTTP/1.1
		},
		HTTPPort:     8080,
		HTTPSPort:    8443,
		AutoRedirect: true,
	}
}

// DevelopmentTLSConfig возвращает конфигурацию для разработки
func DevelopmentTLSConfig() *TLSConfig {
	config := DefaultTLSConfig()
	
	// Используем самоподписанные сертификаты для разработки
	config.CertFile = "./certs/dev-cert.pem"
	config.KeyFile = "./certs/dev-key.pem"
	config.AutoRedirect = false // В dev режиме не принуждаем HTTPS
	
	return config
}

// TLSManager управляет TLS конфигурацией
type TLSManager struct {
	config *TLSConfig
	logger *zap.Logger
}

// NewTLSManager создает новый TLS manager
func NewTLSManager(config *TLSConfig, logger *zap.Logger) *TLSManager {
	if config == nil {
		config = DefaultTLSConfig()
	}
	
	return &TLSManager{
		config: config,
		logger: logger,
	}
}

// GetTLSConfig возвращает настроенную TLS конфигурацию
func (tm *TLSManager) GetTLSConfig() (*tls.Config, error) {
	// Проверяем наличие сертификатов
	if !tm.certificatesExist() {
		return nil, fmt.Errorf("TLS certificates not found: cert=%s, key=%s", 
			tm.config.CertFile, tm.config.KeyFile)
	}
	
	// Загружаем сертификат
	cert, err := tls.LoadX509KeyPair(tm.config.CertFile, tm.config.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS certificate: %w", err)
	}
	
	tlsConfig := &tls.Config{
		Certificates:             []tls.Certificate{cert},
		MinVersion:               tm.config.MinVersion,
		MaxVersion:               tm.config.MaxVersion,
		CipherSuites:             tm.config.CipherSuites,
		CurvePreferences:         tm.config.CurvePreferences,
		PreferServerCipherSuites: tm.config.PreferServerCipherSuites,
		InsecureSkipVerify:       tm.config.InsecureSkipVerify,
		NextProtos:               tm.config.NextProtos,
	}
	
	tm.logger.Info("TLS configuration loaded",
		zap.String("cert_file", tm.config.CertFile),
		zap.Uint16("min_version", tm.config.MinVersion),
		zap.Uint16("max_version", tm.config.MaxVersion),
		zap.Int("cipher_suites_count", len(tm.config.CipherSuites)),
	)
	
	return tlsConfig, nil
}

// certificatesExist проверяет наличие файлов сертификатов
func (tm *TLSManager) certificatesExist() bool {
	if _, err := os.Stat(tm.config.CertFile); os.IsNotExist(err) {
		tm.logger.Warn("Certificate file not found", zap.String("file", tm.config.CertFile))
		return false
	}
	
	if _, err := os.Stat(tm.config.KeyFile); os.IsNotExist(err) {
		tm.logger.Warn("Private key file not found", zap.String("file", tm.config.KeyFile))
		return false
	}
	
	return true
}

// CreateHTTPSServer создает HTTPS сервер с правильной конфигурацией
func (tm *TLSManager) CreateHTTPSServer(handler http.Handler) (*http.Server, error) {
	tlsConfig, err := tm.GetTLSConfig()
	if err != nil {
		return nil, err
	}
	
	server := &http.Server{
		Addr:      fmt.Sprintf(":%d", tm.config.HTTPSPort),
		Handler:   handler,
		TLSConfig: tlsConfig,
		
		// Security timeouts
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		
		// Disable HTTP/2 push (security best practice)
		// TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}
	
	return server, nil
}

// CreateHTTPRedirectServer создает HTTP сервер для редиректа на HTTPS
func (tm *TLSManager) CreateHTTPRedirectServer() *http.Server {
	if !tm.config.AutoRedirect {
		return nil
	}
	
	redirectHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		httpsURL := fmt.Sprintf("https://%s:%d%s", 
			r.Host, tm.config.HTTPSPort, r.RequestURI)
		
		tm.logger.Debug("Redirecting HTTP to HTTPS",
			zap.String("original", r.URL.String()),
			zap.String("redirect", httpsURL),
		)
		
		http.Redirect(w, r, httpsURL, http.StatusMovedPermanently)
	})
	
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", tm.config.HTTPPort),
		Handler:      redirectHandler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  15 * time.Second,
	}
	
	return server
}

// GenerateSelfSignedCert генерирует самоподписанный сертификат для разработки
func (tm *TLSManager) GenerateSelfSignedCert() error {
	// Эта функция должна генерировать самоподписанный сертификат
	// Для простоты, пока что просто логируем
	tm.logger.Info("Self-signed certificate generation requested",
		zap.String("cert_file", tm.config.CertFile),
		zap.String("key_file", tm.config.KeyFile),
	)
	
	// TODO: Реализовать генерацию самоподписанного сертификата
	return fmt.Errorf("self-signed certificate generation not implemented yet")
}

// ValidateTLSConfig проверяет корректность TLS конфигурации
func (tm *TLSManager) ValidateTLSConfig() error {
	if tm.config.MinVersion > tm.config.MaxVersion {
		return fmt.Errorf("min TLS version (%d) cannot be greater than max version (%d)",
			tm.config.MinVersion, tm.config.MaxVersion)
	}
	
	if tm.config.MinVersion < tls.VersionTLS12 {
		tm.logger.Warn("TLS version below 1.2 is not recommended for security",
			zap.Uint16("min_version", tm.config.MinVersion))
	}
	
	if len(tm.config.CipherSuites) == 0 {
		tm.logger.Warn("No cipher suites specified, using Go defaults")
	}
	
	return nil
}

// GetConfig возвращает текущую конфигурацию
func (tm *TLSManager) GetConfig() *TLSConfig {
	return tm.config
}

// UpdateConfig обновляет конфигурацию
func (tm *TLSManager) UpdateConfig(config *TLSConfig) error {
	if err := tm.ValidateTLSConfig(); err != nil {
		return err
	}
	
	tm.config = config
	tm.logger.Info("TLS configuration updated",
		zap.String("cert_file", config.CertFile),
		zap.Uint16("min_version", config.MinVersion),
		zap.Uint16("max_version", config.MaxVersion),
	)
	
	return nil
}
