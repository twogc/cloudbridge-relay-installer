package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"github.com/quic-go/quic-go/quicvarint"
	"go.uber.org/zap"
	"golang.org/x/net/http2"

	"github.com/twogc/cloudbridge-relay-installer/config"
	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
	"github.com/twogc/cloudbridge-relay-installer/internal/masque"
)

// errToStatus конвертирует ошибку в HTTP статус код (из masque-go)
func errToStatus(err error) int {
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		// Consistent with RFC 9209 Section 2.3.1.
		return http.StatusGatewayTimeout
	}
	var dnsError *net.DNSError
	if errors.As(err, &dnsError) {
		// Recommended by RFC 9209 Section 2.3.2.
		return http.StatusBadGateway
	}
	var addrErr *net.AddrError
	var parseError *net.ParseError
	if errors.As(err, &addrErr) || errors.As(err, &parseError) {
		return http.StatusBadRequest
	}
	return http.StatusInternalServerError
}

// MASQUEServer представляет MASQUE сервер для обхода корпоративных сетей
type MASQUEServer struct {
	logger *logger.Logger
	config *config.Config

	// H3 сервер (UDP/QUIC)
	http3Server *http3.Server

	// H2/H1 сервер (TCP/TLS) - fallback
	httpServer *http.Server

	// MASQUE protocol components
	udpProxy       *masque.MASQUEProxy
	connectIPProxy *masque.ConnectIPProxy
	capsuleHandler *masque.CapsuleHandler

	// Управление соединениями
	connections      map[string]*MASQUEConnection
	connectionsMutex sync.RWMutex

	// Состояние
	isRunning bool
	closed    bool
	ctx       context.Context
	cancel    context.CancelFunc
}

// MASQUEConnection представляет MASQUE соединение
type MASQUEConnection struct {
	ID              string
	Stream          *http3.Stream
	UDPConn         *net.UDPConn
	DatagramChannel chan []byte
	CreatedAt       time.Time
	LastActivity    time.Time
}

// MASQUEUDPProxy представляет UDP прокси для CONNECT-UDP
type MASQUEUDPProxy struct {
	logger *logger.Logger
}

// MASQUEIPProxy представляет IP прокси для CONNECT-IP
type MASQUEIPProxy struct {
	logger *logger.Logger
}

// SendDatagramToChannel отправляет datagram через DatagramChannel
func (mc *MASQUEConnection) SendDatagramToChannel(data []byte) error {
	if mc.DatagramChannel == nil {
		return fmt.Errorf("datagram channel not initialized")
	}

	select {
	case mc.DatagramChannel <- data:
		mc.LastActivity = time.Now()
		return nil
	default:
		return fmt.Errorf("datagram channel is full")
	}
}

// SendDatagramWithContextID отправляет datagram с Context ID (из masque-go)
func (mc *MASQUEConnection) SendDatagramWithContextID(data []byte) error {
	if mc.Stream == nil {
		return fmt.Errorf("stream not initialized")
	}

	// Создаем datagram с Context ID согласно RFC 9297 (как в masque-go)
	// Формат: [Quarter Stream ID (varint)][HTTP Datagram Payload]
	contextIDZero := quicvarint.Append([]byte{}, 0)
	datagram := make([]byte, 0, len(contextIDZero)+len(data))
	datagram = append(datagram, contextIDZero...) // Context ID = 0 для CONNECT-UDP
	datagram = append(datagram, data...)

	// Отправляем через HTTP/3 stream
	if err := mc.Stream.SendDatagram(datagram); err != nil {
		return fmt.Errorf("failed to send datagram: %w", err)
	}

	mc.LastActivity = time.Now()
	return nil
}

// SendDatagram отправляет datagram через MASQUE соединение
// Использует HTTP/3 stream с правильным Context ID
func (mc *MASQUEConnection) SendDatagram(data []byte) error {
	return mc.SendDatagramWithContextID(data)
}

// ReceiveDatagram получает datagram из MASQUE соединения
func (mc *MASQUEConnection) ReceiveDatagram(ctx context.Context) ([]byte, error) {
	if mc.DatagramChannel == nil {
		return nil, fmt.Errorf("datagram channel not initialized")
	}

	select {
	case data := <-mc.DatagramChannel:
		mc.LastActivity = time.Now()
		return data, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// NewMASQUEServer создает новый MASQUE сервер
func NewMASQUEServer(logger *logger.Logger, config *config.Config) *MASQUEServer {
	ctx, cancel := context.WithCancel(context.Background())

	// Create MASQUE proxy configuration - PORT 443 for corporate proxy bypass
	masqueConfig := &masque.Config{
		ListenAddr: ":443", // CRITICAL: Use port 443 to bypass corporate firewalls
		CertFile:   config.Security.CertFile,
		KeyFile:    config.Security.KeyFile,
	}

	// Create MASQUE components based on reference implementations
	udpProxy := masque.NewMASQUEProxy(logger, masqueConfig)
	connectIPProxy := masque.NewConnectIPProxy(logger.GetZapLogger())
	capsuleHandler := masque.NewCapsuleHandler(logger.GetZapLogger())

	return &MASQUEServer{
		logger:         logger,
		config:         config,
		udpProxy:       udpProxy,
		connectIPProxy: connectIPProxy,
		capsuleHandler: capsuleHandler,
		connections:    make(map[string]*MASQUEConnection),
		ctx:            ctx,
		cancel:         cancel,
	}
}

// Start запускает MASQUE сервер
func (ms *MASQUEServer) Start() error {
	if !ms.config.MASQUE.Enabled {
		ms.logger.Info("MASQUE support is disabled")
		return nil
	}

	ms.logger.Info("Starting MASQUE server",
		zap.String("proxy_url", ms.config.MASQUE.ProxyURL),
		zap.Strings("alpn_protocols", ms.config.MASQUE.ALPNProtocols))

	// Создаем TLS конфигурацию
	ms.logger.Info("Creating TLS configuration...")
	tlsConfig, err := ms.createTLSConfig()
	if err != nil {
		ms.logger.Error("Failed to create TLS config", zap.Error(err))
		return fmt.Errorf("failed to create TLS config: %v", err)
	}
	ms.logger.Info("TLS configuration created successfully")

	// Создаем HTTP/3 сервер (UDP/QUIC) на порту 443
	ms.logger.Info("Creating HTTP/3 server (UDP/QUIC) on port 443...")
	ms.http3Server = &http3.Server{
		Addr:            ":443",               // CRITICAL: Port 443 to bypass corporate firewalls
		Handler:         ms.createH3Handler(), // Только H3/MASQUE + datagrams
		TLSConfig:       tlsConfig,
		EnableDatagrams: true, // Критично для MASQUE HTTP Datagrams
		QUICConfig: &quic.Config{
			MaxIdleTimeout:  30 * time.Second,
			EnableDatagrams: true, // Включаем QUIC datagrams
			// Anti-amplification protection
			MaxIncomingStreams:    100,
			MaxIncomingUniStreams: 100,
		},
	}
	ms.logger.Info("HTTP/3 server created successfully")

	// Создаем HTTP/2/1 сервер (TCP/TLS) на том же порту 443
	ms.logger.Info("Creating HTTP/2/1 server (TCP/TLS) on port 443...")
	tcpMux := http.NewServeMux()
	tcpMux.HandleFunc("/.well-known/masque/udp/", ms.handleConnectUDPH2)
	tcpMux.HandleFunc("/.well-known/masque/ip/", ms.handleConnectIPH2)
	tcpMux.HandleFunc("/api/v1/connect", ms.handleFallbackConnect)
	tcpMux.HandleFunc("/ws", ms.handleWebSocketFallback)
	tcpMux.HandleFunc("/tcp", ms.handleTCPFallback)
	tcpMux.HandleFunc("/health", ms.handleHealth)
	tcpMux.HandleFunc("/", ms.handleRoot)

	ms.httpServer = &http.Server{
		Addr:      ":443",
		Handler:   tcpMux,
		TLSConfig: tlsConfig,
	}

	// Включаем HTTP/2
	if err := http2.ConfigureServer(ms.httpServer, &http2.Server{}); err != nil {
		ms.logger.Error("Failed to configure HTTP/2", zap.Error(err))
		return fmt.Errorf("failed to configure HTTP/2: %v", err)
	}
	ms.logger.Info("HTTP/2/1 server created successfully")

	// Запускаем H3 сервер (UDP/QUIC)
	go func() {
		ms.logger.Info("Starting HTTP/3 server (UDP/QUIC) on port 443...")
		if err := ms.http3Server.ListenAndServe(); err != nil {
			ms.logger.Error("HTTP/3 server error", zap.Error(err))
			ms.isRunning = false
		}
	}()

	// Запускаем H2/H1 сервер (TCP/TLS)
	go func() {
		ms.logger.Info("Starting HTTP/2/1 server (TCP/TLS) on port 443...")
		if err := ms.httpServer.ListenAndServeTLS("", ""); err != nil && !errors.Is(err, http.ErrServerClosed) {
			ms.logger.Error("HTTP/2/1 server error", zap.Error(err))
			ms.isRunning = false
		}
	}()

	ms.isRunning = true
	ms.logger.Info("MASQUE server started successfully")

	return nil
}

// createTLSConfig создает TLS конфигурацию для MASQUE
func (ms *MASQUEServer) createTLSConfig() (*tls.Config, error) {
	ms.logger.Info("Creating TLS config for MASQUE server...")

	// Пытаемся загрузить реальный сертификат Let's Encrypt
	ms.logger.Info("Attempting to load Let's Encrypt certificate...")
	cert, err := ms.loadLetsEncryptCert()
	if err != nil {
		ms.logger.Warn("Failed to load Let's Encrypt certificate, falling back to self-signed", zap.Error(err))
		// Fallback к самоподписанному сертификату
		ms.logger.Info("Generating self-signed certificate...")
		cert, err = ms.generateSelfSignedCert()
		if err != nil {
			ms.logger.Error("Failed to generate self-signed certificate", zap.Error(err))
			return nil, fmt.Errorf("failed to generate certificate: %v", err)
		}
		ms.logger.Info("Self-signed certificate generated successfully")
	} else {
		ms.logger.Info("Let's Encrypt certificate loaded successfully")
	}

	ms.logger.Info("Creating TLS configuration with HTTPS masquerading...")
	tlsConfig := &tls.Config{
		// CRITICAL: TLS 1.3 required for H3
		MinVersion: tls.VersionTLS13,
		// Remove manual cipher suites - let TLS 1.3 choose automatically
		Certificates: []tls.Certificate{cert},
		// CRITICAL: ALPN protocols for HTTPS masquerading to bypass corporate proxies
		NextProtos: []string{"h3", "h2", "http/1.1"}, // Support HTTP/3, HTTP/2, and HTTP/1.1
	}

	ms.logger.Info("TLS configuration created successfully",
		zap.Strings("alpn_protocols", ms.config.MASQUE.ALPNProtocols),
		zap.Int("cipher_suites_count", len(tlsConfig.CipherSuites)),
		zap.Int("certificates_count", len(tlsConfig.Certificates)))

	return tlsConfig, nil
}

// loadLetsEncryptCert загружает сертификат Let's Encrypt
func (ms *MASQUEServer) loadLetsEncryptCert() (tls.Certificate, error) {
	ms.logger.Info("Loading Let's Encrypt certificate for auth.2gc.ru...")

	// Пути к сертификатам Let's Encrypt
	certFile := "/etc/letsencrypt/live/auth.2gc.ru/fullchain.pem"
	keyFile := "/etc/letsencrypt/live/auth.2gc.ru/privkey.pem"

	ms.logger.Info("Certificate paths",
		zap.String("cert_file", certFile),
		zap.String("key_file", keyFile))

	// Загружаем сертификат
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		ms.logger.Error("Failed to load Let's Encrypt certificate", zap.Error(err))
		return tls.Certificate{}, fmt.Errorf("failed to load certificate: %v", err)
	}

	ms.logger.Info("Let's Encrypt certificate loaded successfully")
	return cert, nil
}

// generateSelfSignedCert генерирует самоподписанный сертификат
func (ms *MASQUEServer) generateSelfSignedCert() (tls.Certificate, error) {
	ms.logger.Info("Generating RSA private key (2048 bits)...")
	// Генерируем приватный ключ
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		ms.logger.Error("Failed to generate RSA private key", zap.Error(err))
		return tls.Certificate{}, fmt.Errorf("failed to generate private key: %v", err)
	}
	ms.logger.Info("RSA private key generated successfully")

	ms.logger.Info("Creating X.509 certificate template...")
	// Создаем сертификат
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization:  []string{"CloudBridge"},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{"San Francisco"},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		DNSNames:    []string{"localhost", "b1.2gc.space"},
	}
	ms.logger.Info("X.509 certificate template created",
		zap.Strings("dns_names", template.DNSNames),
		zap.Strings("ip_addresses", []string{"127.0.0.1", "::1"}))

	ms.logger.Info("Signing X.509 certificate...")
	// Подписываем сертификат
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		ms.logger.Error("Failed to sign X.509 certificate", zap.Error(err))
		return tls.Certificate{}, fmt.Errorf("failed to create certificate: %v", err)
	}
	ms.logger.Info("X.509 certificate signed successfully")

	ms.logger.Info("Creating TLS certificate...")
	// Создаем TLS сертификат
	tlsCert := tls.Certificate{
		Certificate: [][]byte{certDER},
		PrivateKey:  priv,
	}
	ms.logger.Info("TLS certificate created successfully",
		zap.Int("certificate_chain_length", len(tlsCert.Certificate)),
		zap.Bool("private_key_present", tlsCert.PrivateKey != nil))

	return tlsCert, nil
}

// createHTTPHandler создает HTTP обработчик для MASQUE с fallback поддержкой
func (ms *MASQUEServer) createHTTPHandler() http.Handler {
	ms.logger.Info("Creating HTTP handler for MASQUE server with fallback support...")
	mux := http.NewServeMux()

	// CONNECT-UDP обработчик
	mux.HandleFunc("/.well-known/masque/udp/", ms.handleConnectUDP)
	ms.logger.Info("CONNECT-UDP handler registered", zap.String("path", "/.well-known/masque/udp/"))

	// CONNECT-IP обработчик
	mux.HandleFunc("/.well-known/masque/ip/", ms.handleConnectIP)
	ms.logger.Info("CONNECT-IP handler registered", zap.String("path", "/.well-known/masque/ip/"))

	// HTTP Datagrams обработчик
	mux.HandleFunc("/.well-known/masque/datagrams/", ms.handleHTTPDatagrams)
	ms.logger.Info("HTTP Datagrams handler registered", zap.String("path", "/.well-known/masque/datagrams/"))

	// Fallback handlers for corporate proxy bypass
	mux.HandleFunc("/api/v1/connect", ms.handleFallbackConnect) // HTTP/2 CONNECT fallback
	mux.HandleFunc("/ws", ms.handleWebSocketFallback)           // WebSocket fallback
	mux.HandleFunc("/tcp", ms.handleTCPFallback)                // TCP fallback

	// Health check
	mux.HandleFunc("/health", ms.handleHealth)
	ms.logger.Info("Health check handler registered", zap.String("path", "/health"))

	// Root handler for HTTPS masquerading
	mux.HandleFunc("/", ms.handleRoot)

	ms.logger.Info("HTTP handler created successfully with fallback support")
	return mux
}

// createH3Handler создает HTTP/3 обработчик (только H3/MASQUE + datagrams)
func (ms *MASQUEServer) createH3Handler() http.Handler {
	ms.logger.Info("Creating HTTP/3 handler for MASQUE...")
	mux := http.NewServeMux()

	// CONNECT-UDP обработчик (H3 native datagrams)
	mux.HandleFunc("/.well-known/masque/udp/", ms.handleConnectUDPH3)
	ms.logger.Info("CONNECT-UDP H3 handler registered", zap.String("path", "/.well-known/masque/udp/"))

	// CONNECT-IP обработчик (H3 native datagrams)
	mux.HandleFunc("/.well-known/masque/ip/", ms.handleConnectIPH3)
	ms.logger.Info("CONNECT-IP H3 handler registered", zap.String("path", "/.well-known/masque/ip/"))

	// HTTP Datagrams обработчик (H3 native datagrams)
	mux.HandleFunc("/.well-known/masque/datagrams/", ms.handleHTTPDatagramsH3)
	ms.logger.Info("HTTP Datagrams H3 handler registered", zap.String("path", "/.well-known/masque/datagrams/"))

	// Health check
	mux.HandleFunc("/health", ms.handleHealth)
	ms.logger.Info("Health check handler registered", zap.String("path", "/health"))

	ms.logger.Info("HTTP/3 handler created successfully")
	return mux
}

// handleConnectUDPH3 обрабатывает CONNECT-UDP запросы для H3 (native datagrams)
func (ms *MASQUEServer) handleConnectUDPH3(w http.ResponseWriter, r *http.Request) {
	ms.logger.Info("CONNECT-UDP H3 request received",
		zap.String("method", r.Method),
		zap.String("url", r.URL.String()),
		zap.String("host", r.Host))

	// CRITICAL: CONNECT parsing из :authority header (RFC 9298)
	target := r.Host
	if target == "" {
		// Fallback на path parsing для совместимости
		target = strings.TrimPrefix(r.URL.Path, "/.well-known/masque/udp/")
		if target == "" {
			http.Error(w, "Missing target in :authority header or URL path", http.StatusBadRequest)
			return
		}
	}

	ms.logger.Info("CONNECT-UDP H3 target resolved", zap.String("target", target))

	// H3 native datagrams - не нужен Capsule-Protocol
	w.WriteHeader(http.StatusOK)

	// Создаем MASQUE соединение для H3
	connection, err := ms.createMASQUEConnection(r, target, "udp")
	if err != nil {
		ms.logger.Error("Failed to create MASQUE connection", zap.Error(err))
		http.Error(w, "Failed to create connection", http.StatusInternalServerError)
		return
	}

	ms.logger.Info("CONNECT-UDP H3 connection established", zap.String("connection_id", connection.ID))
}

// handleConnectIPH3 обрабатывает CONNECT-IP запросы для H3
func (ms *MASQUEServer) handleConnectIPH3(w http.ResponseWriter, r *http.Request) {
	ms.logger.Info("CONNECT-IP H3 request received",
		zap.String("method", r.Method),
		zap.String("url", r.URL.String()),
		zap.String("host", r.Host))

	// CONNECT parsing из :authority header
	target := r.Host
	if target == "" {
		target = strings.TrimPrefix(r.URL.Path, "/.well-known/masque/ip/")
		if target == "" {
			http.Error(w, "Missing target in :authority header or URL path", http.StatusBadRequest)
			return
		}
	}

	ms.logger.Info("CONNECT-IP H3 target resolved", zap.String("target", target))

	// H3 native datagrams
	w.WriteHeader(http.StatusOK)

	connection, err := ms.createMASQUEConnection(r, target, "ip")
	if err != nil {
		ms.logger.Error("Failed to create MASQUE connection", zap.Error(err))
		http.Error(w, "Failed to create connection", http.StatusInternalServerError)
		return
	}

	ms.logger.Info("CONNECT-IP H3 connection established", zap.String("connection_id", connection.ID))
}

// handleHTTPDatagramsH3 обрабатывает HTTP Datagrams для H3
func (ms *MASQUEServer) handleHTTPDatagramsH3(w http.ResponseWriter, r *http.Request) {
	ms.logger.Info("HTTP Datagrams H3 request received",
		zap.String("method", r.Method),
		zap.String("url", r.URL.String()))

	// H3 native datagrams - не требуем Capsule-Protocol
	w.WriteHeader(http.StatusOK)

	ms.logger.Info("HTTP Datagrams H3 handler completed")
}

// handleConnectUDPH2 обрабатывает CONNECT-UDP запросы для H2/H1 (Capsule-Protocol)
func (ms *MASQUEServer) handleConnectUDPH2(w http.ResponseWriter, r *http.Request) {
	ms.logger.Info("CONNECT-UDP H2/H1 request received",
		zap.String("method", r.Method),
		zap.String("url", r.URL.String()),
		zap.String("host", r.Host))

	// Проверяем Capsule-Protocol header для H2/H1
	if r.Header.Get("Capsule-Protocol") != "?1" {
		http.Error(w, "Capsule-Protocol header required for H2/H1", http.StatusBadRequest)
		return
	}

	// CONNECT parsing из :authority header
	target := r.Host
	if target == "" {
		target = strings.TrimPrefix(r.URL.Path, "/.well-known/masque/udp/")
		if target == "" {
			http.Error(w, "Missing target in :authority header or URL path", http.StatusBadRequest)
			return
		}
	}

	ms.logger.Info("CONNECT-UDP H2/H1 target resolved", zap.String("target", target))

	// H2/H1 с Capsule-Protocol
	w.Header().Set("Capsule-Protocol", "?1")
	w.WriteHeader(http.StatusOK)

	ms.logger.Info("CONNECT-UDP H2/H1 connection established")
}

// handleConnectIPH2 обрабатывает CONNECT-IP запросы для H2/H1
func (ms *MASQUEServer) handleConnectIPH2(w http.ResponseWriter, r *http.Request) {
	ms.logger.Info("CONNECT-IP H2/H1 request received",
		zap.String("method", r.Method),
		zap.String("url", r.URL.String()),
		zap.String("host", r.Host))

	// Проверяем Capsule-Protocol header
	if r.Header.Get("Capsule-Protocol") != "?1" {
		http.Error(w, "Capsule-Protocol header required for H2/H1", http.StatusBadRequest)
		return
	}

	// CONNECT parsing из :authority header
	target := r.Host
	if target == "" {
		target = strings.TrimPrefix(r.URL.Path, "/.well-known/masque/ip/")
		if target == "" {
			http.Error(w, "Missing target in :authority header or URL path", http.StatusBadRequest)
			return
		}
	}

	ms.logger.Info("CONNECT-IP H2/H1 target resolved", zap.String("target", target))

	// H2/H1 с Capsule-Protocol
	w.Header().Set("Capsule-Protocol", "?1")
	w.WriteHeader(http.StatusOK)

	ms.logger.Info("CONNECT-IP H2/H1 connection established")
}

// handleConnectUDP обрабатывает CONNECT-UDP запросы
func (ms *MASQUEServer) handleConnectUDP(w http.ResponseWriter, r *http.Request) {
	ms.logger.Info("CONNECT-UDP request received",
		zap.String("method", r.Method),
		zap.String("url", r.URL.String()),
		zap.String("remote_addr", r.RemoteAddr))

	// Проверяем метод
	if r.Method != http.MethodConnect {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Извлекаем целевой адрес из URL
	target := r.URL.Path[len("/.well-known/masque/udp/"):]
	if target == "" {
		http.Error(w, "Target address required", http.StatusBadRequest)
		return
	}

	// Create MASQUE request with target address
	req := &masque.Request{Host: target}

	// Handle connection using MASQUE proxy
	if err := ms.udpProxy.Proxy(w, req); err != nil {
		ms.logger.Error("Failed to handle UDP proxy connection", zap.Error(err))
		http.Error(w, "Failed to create connection", http.StatusInternalServerError)
		return
	}
}

// handleConnectIP обрабатывает CONNECT-IP запросы
func (ms *MASQUEServer) handleConnectIP(w http.ResponseWriter, r *http.Request) {
	ms.logger.Info("CONNECT-IP request received",
		zap.String("method", r.Method),
		zap.String("url", r.URL.String()),
		zap.String("remote_addr", r.RemoteAddr))

	// Проверяем метод
	if r.Method != http.MethodConnect {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Извлекаем целевой адрес из URL
	target := r.URL.Path[len("/.well-known/masque/ip/"):]
	if target == "" {
		http.Error(w, "Target address required", http.StatusBadRequest)
		return
	}

	// Create CONNECT-IP request
	connectIPReq := &masque.ConnectIPRequest{TargetIP: target}

	// Handle connection using CONNECT-IP proxy
	if _, err := ms.connectIPProxy.Proxy(w, connectIPReq); err != nil {
		ms.logger.Error("Failed to handle IP proxy connection", zap.Error(err))
		http.Error(w, "Failed to create connection", http.StatusInternalServerError)
		return
	}
}

// handleHTTPDatagrams обрабатывает HTTP Datagrams запросы
func (ms *MASQUEServer) handleHTTPDatagrams(w http.ResponseWriter, r *http.Request) {
	ms.logger.Info("HTTP Datagrams request received",
		zap.String("method", r.Method),
		zap.String("url", r.URL.String()),
		zap.String("remote_addr", r.RemoteAddr))

	// Проверяем Capsule-Protocol заголовок
	if r.Header.Get("Capsule-Protocol") != "?1" {
		http.Error(w, "Capsule-Protocol header required", http.StatusBadRequest)
		return
	}

	// Получаем HTTP/3 stream из контекста запроса
	// Используем правильный способ получения stream для HTTP/3
	streamer, ok := r.Context().Value("http3.streamer").(http3.HTTPStreamer)
	if !ok {
		// Альтернативный способ получения stream
		if stream, exists := r.Context().Value("http3.stream").(*http3.Stream); exists {
			// Устанавливаем заголовки ответа
			w.Header().Set("Capsule-Protocol", "?1")
			w.WriteHeader(http.StatusOK)

			// Запускаем обработку datagrams
			go ms.handleDatagramStream(stream, r.RemoteAddr)
			return
		}
		http.Error(w, "Not an HTTP/3 stream", http.StatusBadRequest)
		return
	}

	stream := streamer.HTTPStream()

	// Устанавливаем заголовки ответа
	w.Header().Set("Capsule-Protocol", "?1")
	w.WriteHeader(http.StatusOK)

	// Запускаем обработку datagrams
	go ms.handleDatagramStream(stream, r.RemoteAddr)
}

// handleDatagramStream обрабатывает поток datagrams
func (ms *MASQUEServer) handleDatagramStream(stream *http3.Stream, remoteAddr string) {
	defer stream.Close()

	ms.logger.Info("Starting datagram stream processing", zap.String("remote_addr", remoteAddr))

	for {
		// Читаем datagram из HTTP/3 stream (который уже интегрирован с QUIC datagrams)
		data, err := stream.ReceiveDatagram(context.Background())
		if err != nil {
			ms.logger.Debug("Failed to receive datagram from stream",
				zap.String("remote_addr", remoteAddr),
				zap.Error(err))
			break
		}

		// Обрабатываем datagram
		if err := ms.processDatagram(data, remoteAddr); err != nil {
			ms.logger.Error("Failed to process datagram",
				zap.String("remote_addr", remoteAddr),
				zap.Error(err))
		}
	}

	ms.logger.Info("Datagram stream processing ended", zap.String("remote_addr", remoteAddr))
}

// processDatagram обрабатывает полученный datagram с Context ID поддержкой
// Основано на masque-go/conn.go:ReadFrom
func (ms *MASQUEServer) processDatagram(data []byte, remoteAddr string) error {
	ms.logger.Debug("Processing datagram",
		zap.String("remote_addr", remoteAddr),
		zap.Int("data_len", len(data)))

	// HTTP Datagrams согласно RFC 9297 имеют формат:
	// [Quarter Stream ID (varint)][HTTP Datagram Payload]
	// Для MASQUE Context ID может быть 0 (CONNECT-UDP) или != 0 (CONNECT-IP)

	if len(data) == 0 {
		ms.logger.Warn("Empty datagram received")
		return fmt.Errorf("empty datagram")
	}

	// Парсим Quarter Stream ID (Context ID для MASQUE) как в masque-go
	contextID, n, err := quicvarint.Parse(data)
	if err != nil {
		ms.logger.Warn("Failed to parse context ID from HTTP datagram", zap.Error(err))
		return fmt.Errorf("masque: malformed datagram: %w", err)
	}

	// Извлекаем payload после Context ID
	payload := data[n:]

	ms.logger.Debug("Processing HTTP datagram",
		zap.Uint64("context_id", contextID),
		zap.Int("payload_len", len(payload)),
		zap.String("remote_addr", remoteAddr))

	// Проверяем Context ID как в masque-go
	if contextID != 0 {
		// Для CONNECT-IP (Context ID != 0) обрабатываем как IP Capsule
		ms.logger.Debug("Processing IP Capsule for CONNECT-IP", zap.Uint64("context_id", contextID))
		return ms.processIPCapsule(payload, remoteAddr, contextID)
	}

	// Для CONNECT-UDP (Context ID = 0) - это UDP payload
	// В отличие от нашей предыдущей реализации, здесь мы не добавляем маршрутизацию
	// Вместо этого мы просто передаем UDP payload как есть
	// Маршрутизация должна происходить на уровне MASQUE соединений

	ms.logger.Debug("Processing UDP payload for CONNECT-UDP",
		zap.Int("payload_len", len(payload)),
		zap.String("remote_addr", remoteAddr))

	// Здесь мы можем добавить логику для маршрутизации UDP payload
	// Пока что просто логируем получение
	return nil
}

// processIPCapsule обрабатывает IP Capsule для CONNECT-IP
func (ms *MASQUEServer) processIPCapsule(data []byte, remoteAddr string, contextID uint64) error {
	ms.logger.Debug("Processing IP Capsule",
		zap.String("remote_addr", remoteAddr),
		zap.Uint64("context_id", contextID),
		zap.Int("data_len", len(data)))

	// Проверяем, инициализирован ли Capsule Handler
	if ms.capsuleHandler == nil {
		ms.logger.Warn("Capsule Handler not initialized, skipping capsule processing")
		return fmt.Errorf("Capsule Handler not initialized")
	}

	// Для простоты, мы просто логируем получение IP capsule
	// В полной реализации здесь была бы обработка IP пакетов
	ms.logger.Debug("IP capsule received",
		zap.String("remote_addr", remoteAddr),
		zap.Uint64("context_id", contextID),
		zap.Int("data_len", len(data)))

	// В полной реализации здесь была бы валидация IP пакета и его обработка
	// Например, маршрутизация в соответствующий туннель или обработка локально

	return nil
}

// handleHealth обрабатывает health check запросы
func (ms *MASQUEServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	ms.connectionsMutex.RLock()
	connectionCount := len(ms.connections)
	ms.connectionsMutex.RUnlock()

	response := fmt.Sprintf(`{
		"status": "healthy",
		"masque_enabled": %t,
		"connections": %d,
		"timestamp": "%s"
	}`, ms.config.MASQUE.Enabled, connectionCount, time.Now().UTC().Format(time.RFC3339))

	w.Write([]byte(response))
}

// handleRoot обрабатывает корневые запросы для HTTPS маскировки
func (ms *MASQUEServer) handleRoot(w http.ResponseWriter, r *http.Request) {
	ms.logger.Info("Root request received for HTTPS masquerading",
		zap.String("method", r.Method),
		zap.String("url", r.URL.String()),
		zap.String("user_agent", r.UserAgent()))

	// Возвращаем обычный HTTP ответ для маскировки под веб-сервер
	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Server", "nginx/1.20.2") // Маскируемся под nginx

	// CRITICAL: Alt-Svc для H3 клиентов
	w.Header().Set("Alt-Svc", `h3=":443"; ma=86400`)

	w.WriteHeader(http.StatusOK)

	html := `<!DOCTYPE html>
<html>
<head>
    <title>CloudBridge Relay</title>
</head>
<body>
    <h1>CloudBridge Relay Service</h1>
    <p>Service is running normally.</p>
</body>
</html>`

	w.Write([]byte(html))
}

// handleFallbackConnect обрабатывает HTTP/2 CONNECT fallback
func (ms *MASQUEServer) handleFallbackConnect(w http.ResponseWriter, r *http.Request) {
	ms.logger.Info("HTTP/2 CONNECT fallback request received",
		zap.String("method", r.Method),
		zap.String("url", r.URL.String()))

	if r.Method != http.MethodConnect {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Обрабатываем как обычный CONNECT-UDP
	ms.handleConnectUDP(w, r)
}

// handleWebSocketFallback обрабатывает WebSocket fallback
func (ms *MASQUEServer) handleWebSocketFallback(w http.ResponseWriter, r *http.Request) {
	ms.logger.Info("WebSocket fallback request received",
		zap.String("method", r.Method),
		zap.String("url", r.URL.String()))

	// Проверяем WebSocket upgrade
	if r.Header.Get("Upgrade") != "websocket" {
		http.Error(w, "WebSocket upgrade required", http.StatusBadRequest)
		return
	}

	// Здесь должна быть логика WebSocket fallback
	// Пока что возвращаем ошибку
	http.Error(w, "WebSocket fallback not implemented", http.StatusNotImplemented)
}

// handleTCPFallback обрабатывает TCP fallback
func (ms *MASQUEServer) handleTCPFallback(w http.ResponseWriter, r *http.Request) {
	ms.logger.Info("TCP fallback request received",
		zap.String("method", r.Method),
		zap.String("url", r.URL.String()))

	// Здесь должна быть логика TCP fallback
	// Пока что возвращаем ошибку
	http.Error(w, "TCP fallback not implemented", http.StatusNotImplemented)
}

// createMASQUEConnection создает новое MASQUE соединение
func (ms *MASQUEServer) createMASQUEConnection(r *http.Request, target, protocol string) (*MASQUEConnection, error) {
	// Получаем HTTP/3 stream из контекста запроса
	// Используем правильный способ получения stream для HTTP/3
	streamer, ok := r.Context().Value("http3.streamer").(http3.HTTPStreamer)
	if !ok {
		// Альтернативный способ получения stream
		if stream, exists := r.Context().Value("http3.stream").(*http3.Stream); exists {
			return ms.createConnectionWithStream(stream, target, protocol)
		}
		return nil, fmt.Errorf("not an HTTP/3 stream")
	}

	stream := streamer.HTTPStream()
	return ms.createConnectionWithStream(stream, target, protocol)
}

// createConnectionWithStream создает MASQUE соединение с переданным stream
func (ms *MASQUEServer) createConnectionWithStream(stream *http3.Stream, target, protocol string) (*MASQUEConnection, error) {
	// Создаем UDP соединение к целевому адресу
	udpAddr, err := net.ResolveUDPAddr("udp", target)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve target address: %v", err)
	}

	udpConn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to dial target: %v", err)
	}

	// Создаем MASQUE соединение
	conn := &MASQUEConnection{
		ID:              fmt.Sprintf("masque_%d", time.Now().UnixNano()),
		Stream:          stream,
		UDPConn:         udpConn,
		DatagramChannel: make(chan []byte, 100), // Буфер на 100 datagrams
		CreatedAt:       time.Now(),
		LastActivity:    time.Now(),
	}

	// Сохраняем соединение
	ms.connectionsMutex.Lock()
	ms.connections[conn.ID] = conn
	ms.connectionsMutex.Unlock()

	ms.logger.Info("MASQUE connection created",
		zap.String("conn_id", conn.ID),
		zap.String("target", target),
		zap.String("protocol", protocol))

	return conn, nil
}

// HandleConnection обрабатывает UDP прокси соединение
func (p *MASQUEUDPProxy) HandleConnection(w http.ResponseWriter, conn *MASQUEConnection) error {
	p.logger.Info("Handling UDP proxy connection", zap.String("conn_id", conn.ID))

	// Устанавливаем заголовки ответа
	w.Header().Set("Capsule-Protocol", "?1")
	w.WriteHeader(http.StatusOK)

	// Запускаем проксирование данных
	go p.proxyUDPToTarget(conn)
	go p.proxyUDPFromTarget(conn)

	return nil
}

// HandleConnection обрабатывает IP прокси соединение
func (p *MASQUEIPProxy) HandleConnection(w http.ResponseWriter, conn *MASQUEConnection) error {
	p.logger.Info("Handling IP proxy connection", zap.String("conn_id", conn.ID))

	// Устанавливаем заголовки ответа
	w.Header().Set("Capsule-Protocol", "?1")
	w.WriteHeader(http.StatusOK)

	// Запускаем проксирование данных
	go p.proxyIPToTarget(conn)
	go p.proxyIPFromTarget(conn)

	return nil
}

// proxyUDPToTarget проксирует данные от клиента к целевому серверу
// Основано на masque-go/conn.go:ReadFrom
func (p *MASQUEUDPProxy) proxyUDPToTarget(conn *MASQUEConnection) {
	defer conn.UDPConn.Close()

	for {
		// Читаем данные из HTTP/3 stream (который уже интегрирован с QUIC datagrams)
		data, err := conn.Stream.ReceiveDatagram(context.Background())
		if err != nil {
			p.logger.Debug("Failed to receive datagram from stream",
				zap.String("conn_id", conn.ID),
				zap.Error(err))
			break
		}

		// Парсим Context ID как в masque-go
		contextID, n, err := quicvarint.Parse(data)
		if err != nil {
			p.logger.Warn("Failed to parse context ID from datagram",
				zap.String("conn_id", conn.ID),
				zap.Error(err))
			continue
		}

		// Проверяем Context ID как в masque-go
		if contextID != 0 {
			// Drop datagram для CONNECT-IP (не CONNECT-UDP)
			p.logger.Debug("Dropping datagram with non-zero context ID",
				zap.String("conn_id", conn.ID),
				zap.Uint64("context_id", contextID))
			continue
		}

		// Извлекаем UDP payload после Context ID
		udpPayload := data[n:]

		// Отправляем UDP payload в UDP соединение
		if _, err := conn.UDPConn.Write(udpPayload); err != nil {
			p.logger.Debug("Failed to write to UDP connection",
				zap.String("conn_id", conn.ID),
				zap.Error(err))
			break
		}

		conn.LastActivity = time.Now()
	}
}

// proxyUDPFromTarget проксирует данные от целевого сервера к клиенту
// Основано на masque-go/conn.go:WriteTo
func (p *MASQUEUDPProxy) proxyUDPFromTarget(conn *MASQUEConnection) {
	defer conn.Stream.Close()

	buffer := make([]byte, 1500)
	for {
		// Читаем данные из UDP соединения
		n, err := conn.UDPConn.Read(buffer)
		if err != nil {
			p.logger.Debug("Failed to read from UDP connection",
				zap.String("conn_id", conn.ID),
				zap.Error(err))
			break
		}

		// Создаем datagram с Context ID как в masque-go
		// Формат: [Quarter Stream ID (varint)][HTTP Datagram Payload]
		udpPayload := buffer[:n]
		contextIDZero := quicvarint.Append([]byte{}, 0)
		datagram := make([]byte, 0, len(contextIDZero)+len(udpPayload))
		datagram = append(datagram, contextIDZero...) // Context ID = 0 для CONNECT-UDP
		datagram = append(datagram, udpPayload...)

		// Отправляем datagram через HTTP/3 stream
		if err := conn.Stream.SendDatagram(datagram); err != nil {
			p.logger.Debug("Failed to send datagram to stream",
				zap.String("conn_id", conn.ID),
				zap.Error(err))
			break
		}

		conn.LastActivity = time.Now()
	}
}

// proxyIPToTarget проксирует данные от клиента к целевому серверу (IP)
func (p *MASQUEIPProxy) proxyIPToTarget(conn *MASQUEConnection) {
	defer conn.UDPConn.Close()

	for {
		// Читаем данные из HTTP/3 stream (который уже интегрирован с QUIC datagrams)
		data, err := conn.Stream.ReceiveDatagram(context.Background())
		if err != nil {
			p.logger.Debug("Failed to receive datagram from stream",
				zap.String("conn_id", conn.ID),
				zap.Error(err))
			break
		}

		// Отправляем данные в UDP соединение
		if _, err := conn.UDPConn.Write(data); err != nil {
			p.logger.Debug("Failed to write to UDP connection",
				zap.String("conn_id", conn.ID),
				zap.Error(err))
			break
		}

		conn.LastActivity = time.Now()
	}
}

// proxyIPFromTarget проксирует данные от целевого сервера к клиенту (IP)
func (p *MASQUEIPProxy) proxyIPFromTarget(conn *MASQUEConnection) {
	defer conn.Stream.Close()

	buffer := make([]byte, 1500)
	for {
		// Читаем данные из UDP соединения
		n, err := conn.UDPConn.Read(buffer)
		if err != nil {
			p.logger.Debug("Failed to read from UDP connection",
				zap.String("conn_id", conn.ID),
				zap.Error(err))
			break
		}

		// Отправляем данные через HTTP/3 stream (который уже интегрирован с QUIC datagrams)
		if err := conn.Stream.SendDatagram(buffer[:n]); err != nil {
			p.logger.Debug("Failed to send datagram to stream",
				zap.String("conn_id", conn.ID),
				zap.Error(err))
			break
		}

		conn.LastActivity = time.Now()
	}
}

// GetConnectionInfo возвращает информацию о соединениях
func (ms *MASQUEServer) GetConnectionInfo() map[string]interface{} {
	ms.connectionsMutex.RLock()
	defer ms.connectionsMutex.RUnlock()

	connections := make([]map[string]interface{}, 0, len(ms.connections))
	for connID, conn := range ms.connections {
		target := ""
		if conn.UDPConn != nil {
			target = conn.UDPConn.RemoteAddr().String()
		}
		connections = append(connections, map[string]interface{}{
			"id":            connID,
			"created_at":    conn.CreatedAt,
			"last_activity": conn.LastActivity,
			"target":        target,
		})
	}

	return map[string]interface{}{
		"total_connections": len(ms.connections),
		"is_running":        ms.isRunning,
		"connections":       connections,
	}
}

// Close закрывает MASQUE сервер
func (ms *MASQUEServer) Close() error {
	ms.connectionsMutex.Lock()
	ms.closed = true
	var errs []error

	// Отменяем контекст
	ms.cancel()

	// Закрываем HTTP/3 сервер
	if ms.http3Server != nil {
		if err := ms.http3Server.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	// Закрываем все соединения
	for _, conn := range ms.connections {
		if conn.Stream != nil {
			conn.Stream.Close()
		}
		if conn.UDPConn != nil {
			conn.UDPConn.Close()
		}
	}
	ms.connections = make(map[string]*MASQUEConnection)
	ms.connectionsMutex.Unlock()

	ms.isRunning = false

	return errors.Join(errs...)
}

// GetConnectionCount возвращает количество активных соединений
func (ms *MASQUEServer) GetConnectionCount() int {
	ms.connectionsMutex.RLock()
	defer ms.connectionsMutex.RUnlock()
	return len(ms.connections)
}
