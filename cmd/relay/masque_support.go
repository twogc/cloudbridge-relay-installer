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
	"io"
	"math/big"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"github.com/quic-go/quic-go/quicvarint"
	"go.uber.org/zap"

	"github.com/twogc/cloudbridge-relay-installer/config"
	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
)

// Context ID для MASQUE datagrams (из masque-go)
var contextIDZero = quicvarint.Append([]byte{}, 0)

// proxyEntry представляет запись прокси для Context ID управления
type proxyEntry struct {
	str  *http3.Stream
	conn *net.UDPConn
}

// Close закрывает proxy entry
func (e proxyEntry) Close() error {
	e.str.CancelRead(quic.StreamErrorCode(http3.ErrCodeConnectError))
	return errors.Join(e.str.Close(), e.conn.Close())
}

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
	logger      *logger.Logger
	config      *config.Config
	http3Server *http3.Server
	udpProxy    *MASQUEUDPProxy
	ipProxy     *MASQUEIPProxy

	// IP Capsule поддержка
	ipCapsuleManager *IPCapsuleManager

	// Управление соединениями
	connections      map[string]*MASQUEConnection
	connectionsMutex sync.RWMutex

	// Context ID управление (из masque-go)
	proxyEntries      map[io.Closer]struct{}
	proxyEntriesMutex sync.Mutex
	refCount          sync.WaitGroup

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

	// Создаем datagram с Context ID согласно RFC 9297
	// Формат: [Quarter Stream ID (varint)][HTTP Datagram Payload]
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

	return &MASQUEServer{
		logger:           logger,
		config:           config,
		udpProxy:         &MASQUEUDPProxy{logger: logger},
		ipProxy:          &MASQUEIPProxy{logger: logger},
		ipCapsuleManager: NewIPCapsuleManager(logger),
		connections:      make(map[string]*MASQUEConnection),
		proxyEntries:     make(map[io.Closer]struct{}),
		ctx:              ctx,
		cancel:           cancel,
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

	// Создаем HTTP/3 сервер
	ms.logger.Info("Creating HTTP/3 server...")
	ms.http3Server = &http3.Server{
		Addr:            ":8443",
		Handler:         ms.createHTTPHandler(),
		TLSConfig:       tlsConfig,
		EnableDatagrams: true, // Критично для MASQUE HTTP Datagrams
		QUICConfig: &quic.Config{
			MaxIdleTimeout:  30 * time.Second,
			EnableDatagrams: true, // Включаем QUIC datagrams
		},
	}
	ms.logger.Info("HTTP/3 server created successfully")

	// Запускаем сервер
	go func() {
		addr := ":8443" // MASQUE порт по умолчанию
		ms.logger.Info("MASQUE HTTP/3 server starting", zap.String("addr", addr))

		// Запускаем HTTP/3 сервер
		ms.logger.Info("Starting HTTP/3 server ListenAndServe...")
		ms.logger.Info("HTTP/3 server configuration",
			zap.String("addr", addr),
			zap.Bool("tls_config_present", ms.http3Server.TLSConfig != nil),
			zap.Bool("handler_present", ms.http3Server.Handler != nil))

		// Используем ListenAndServe с явным указанием адреса
		if err := ms.http3Server.ListenAndServe(); err != nil {
			ms.logger.Error("MASQUE HTTP/3 server error", zap.Error(err))
			ms.isRunning = false
		} else {
			ms.logger.Info("HTTP/3 server ListenAndServe completed without error")
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

	ms.logger.Info("Creating TLS configuration...")
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
		Certificates: []tls.Certificate{cert},
		NextProtos:   ms.config.MASQUE.ALPNProtocols,
	}

	ms.logger.Info("TLS configuration created successfully",
		zap.Strings("alpn_protocols", ms.config.MASQUE.ALPNProtocols),
		zap.Int("cipher_suites_count", len(tlsConfig.CipherSuites)),
		zap.Int("certificates_count", len(tlsConfig.Certificates)))

	return tlsConfig, nil
}

// loadLetsEncryptCert загружает сертификат Let's Encrypt
func (ms *MASQUEServer) loadLetsEncryptCert() (tls.Certificate, error) {
	ms.logger.Info("Loading Let's Encrypt certificate for edge.2gc.ru...")

	// Пути к сертификатам Let's Encrypt
	certFile := "/etc/letsencrypt/live/edge.2gc.ru/fullchain.pem"
	keyFile := "/etc/letsencrypt/live/edge.2gc.ru/privkey.pem"

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

// createHTTPHandler создает HTTP обработчик для MASQUE
func (ms *MASQUEServer) createHTTPHandler() http.Handler {
	ms.logger.Info("Creating HTTP handler for MASQUE server...")
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

	// Health check
	mux.HandleFunc("/health", ms.handleHealth)
	ms.logger.Info("Health check handler registered", zap.String("path", "/health"))

	ms.logger.Info("HTTP handler created successfully")
	return mux
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

	// Создаем MASQUE соединение
	conn, err := ms.createMASQUEConnection(r, target, "udp")
	if err != nil {
		ms.logger.Error("Failed to create MASQUE connection", zap.Error(err))
		http.Error(w, "Failed to create connection", http.StatusInternalServerError)
		return
	}

	// Обрабатываем соединение
	if err := ms.udpProxy.HandleConnection(w, conn); err != nil {
		ms.logger.Error("Failed to handle UDP proxy connection", zap.Error(err))
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

	// Создаем MASQUE соединение
	conn, err := ms.createMASQUEConnection(r, target, "ip")
	if err != nil {
		ms.logger.Error("Failed to create MASQUE connection", zap.Error(err))
		http.Error(w, "Failed to create connection", http.StatusInternalServerError)
		return
	}

	// Обрабатываем соединение
	if err := ms.ipProxy.HandleConnection(w, conn); err != nil {
		ms.logger.Error("Failed to handle IP proxy connection", zap.Error(err))
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

	// Проверяем, инициализирован ли IP Capsule Manager
	if ms.ipCapsuleManager == nil {
		ms.logger.Warn("IP Capsule Manager not initialized, skipping IP capsule processing")
		return fmt.Errorf("IP Capsule Manager not initialized")
	}

	// Обрабатываем IP Capsule
	capsule, err := ms.ipCapsuleManager.ProcessCapsule(data)
	if err != nil {
		ms.logger.Error("Failed to process IP capsule", zap.Error(err))
		return fmt.Errorf("failed to process IP capsule: %w", err)
	}

	// Валидируем IP пакет
	if err := capsule.Validate(); err != nil {
		ms.logger.Error("Invalid IP packet in capsule", zap.Error(err))
		return fmt.Errorf("invalid IP packet: %w", err)
	}

	ms.logger.Debug("IP Capsule processed successfully",
		zap.String("remote_addr", remoteAddr),
		zap.Uint64("context_id", contextID),
		zap.Int("ip_packet_len", len(capsule.IPPacket)),
		zap.Uint16("checksum", capsule.Checksum))

	// Здесь можно добавить логику маршрутизации IP пакетов
	// Например, отправка в соответствующий туннель или обработка локально

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

// Close закрывает MASQUE сервер с поддержкой Context ID управления (из masque-go)
func (ms *MASQUEServer) Close() error {
	ms.proxyEntriesMutex.Lock()
	ms.closed = true
	var errs []error

	// Закрываем все proxy entries
	for closer := range ms.proxyEntries {
		errs = append(errs, closer.Close())
	}
	ms.proxyEntriesMutex.Unlock()

	// Ждем завершения всех горутин
	ms.refCount.Wait()

	// Отменяем контекст
	ms.cancel()

	// Закрываем HTTP/3 сервер
	if ms.http3Server != nil {
		if err := ms.http3Server.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	// Закрываем все соединения
	ms.connectionsMutex.Lock()
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
	ms.proxyEntries = nil

	return errors.Join(errs...)
}

// GetConnectionCount возвращает количество активных соединений
func (ms *MASQUEServer) GetConnectionCount() int {
	ms.connectionsMutex.RLock()
	defer ms.connectionsMutex.RUnlock()
	return len(ms.connections)
}
