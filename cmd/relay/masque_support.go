package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"go.uber.org/zap"

	"github.com/twogc/cloudbridge-relay-installer/config"
	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
)

// MASQUEServer представляет MASQUE сервер для обхода корпоративных сетей
type MASQUEServer struct {
	logger      *logger.Logger
	config      *config.Config
	http3Server *http3.Server
	udpProxy    *MASQUEUDPProxy
	ipProxy     *MASQUEIPProxy

	// Управление соединениями
	connections      map[string]*MASQUEConnection
	connectionsMutex sync.RWMutex

	// Состояние
	isRunning bool
	ctx       context.Context
	cancel    context.CancelFunc
}

// MASQUEConnection представляет MASQUE соединение
type MASQUEConnection struct {
	ID           string
	Stream       *http3.Stream
	UDPConn      *net.UDPConn
	CreatedAt    time.Time
	LastActivity time.Time
}

// MASQUEUDPProxy представляет UDP прокси для CONNECT-UDP
type MASQUEUDPProxy struct {
	logger *logger.Logger
}

// MASQUEIPProxy представляет IP прокси для CONNECT-IP
type MASQUEIPProxy struct {
	logger *logger.Logger
}

// NewMASQUEServer создает новый MASQUE сервер
func NewMASQUEServer(logger *logger.Logger, config *config.Config) *MASQUEServer {
	ctx, cancel := context.WithCancel(context.Background())

	return &MASQUEServer{
		logger:      logger,
		config:      config,
		udpProxy:    &MASQUEUDPProxy{logger: logger},
		ipProxy:     &MASQUEIPProxy{logger: logger},
		connections: make(map[string]*MASQUEConnection),
		ctx:         ctx,
		cancel:      cancel,
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
	tlsConfig, err := ms.createTLSConfig()
	if err != nil {
		return fmt.Errorf("failed to create TLS config: %v", err)
	}

	// Создаем HTTP/3 сервер
	ms.http3Server = &http3.Server{
		Handler:   ms.createHTTPHandler(),
		TLSConfig: tlsConfig,
		QUICConfig: &quic.Config{
			MaxIdleTimeout: 30 * time.Second,
		},
	}

	// Запускаем сервер
	go func() {
		addr := ":8443" // MASQUE порт по умолчанию
		ms.logger.Info("MASQUE HTTP/3 server starting", zap.String("addr", addr))

		if err := ms.http3Server.ListenAndServe(); err != nil {
			ms.logger.Error("MASQUE HTTP/3 server error", zap.Error(err))
		}
	}()

	ms.isRunning = true
	ms.logger.Info("MASQUE server started successfully")

	return nil
}

// createTLSConfig создает TLS конфигурацию для MASQUE
func (ms *MASQUEServer) createTLSConfig() (*tls.Config, error) {
	// Для демонстрации используем самоподписанный сертификат
	// В продакшене следует использовать реальные сертификаты
	cert, err := ms.generateSelfSignedCert()
	if err != nil {
		return nil, fmt.Errorf("failed to generate certificate: %v", err)
	}

	return &tls.Config{
		MinVersion: tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
		Certificates: []tls.Certificate{cert},
		NextProtos:   ms.config.MASQUE.ALPNProtocols,
	}, nil
}

// generateSelfSignedCert генерирует самоподписанный сертификат
func (ms *MASQUEServer) generateSelfSignedCert() (tls.Certificate, error) {
	// Генерируем приватный ключ
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to generate private key: %v", err)
	}

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

	// Подписываем сертификат
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to create certificate: %v", err)
	}

	// Создаем TLS сертификат
	return tls.Certificate{
		Certificate: [][]byte{certDER},
		PrivateKey:  priv,
	}, nil
}

// createHTTPHandler создает HTTP обработчик для MASQUE
func (ms *MASQUEServer) createHTTPHandler() http.Handler {
	mux := http.NewServeMux()

	// CONNECT-UDP обработчик
	mux.HandleFunc("/.well-known/masque/udp/", ms.handleConnectUDP)

	// CONNECT-IP обработчик
	mux.HandleFunc("/.well-known/masque/ip/", ms.handleConnectIP)

	// Health check
	mux.HandleFunc("/health", ms.handleHealth)

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
	streamer, ok := r.Context().Value("http3.streamer").(http3.HTTPStreamer)
	if !ok {
		return nil, fmt.Errorf("not an HTTP/3 stream")
	}

	stream := streamer.HTTPStream()

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
		ID:           fmt.Sprintf("masque_%d", time.Now().UnixNano()),
		Stream:       stream,
		UDPConn:      udpConn,
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
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
	go p.proxyToTarget(conn)
	go p.proxyFromTarget(conn)

	return nil
}

// HandleConnection обрабатывает IP прокси соединение
func (p *MASQUEIPProxy) HandleConnection(w http.ResponseWriter, conn *MASQUEConnection) error {
	p.logger.Info("Handling IP proxy connection", zap.String("conn_id", conn.ID))

	// Устанавливаем заголовки ответа
	w.Header().Set("Capsule-Protocol", "?1")
	w.WriteHeader(http.StatusOK)

	// Запускаем проксирование данных
	go p.proxyToTarget(conn)
	go p.proxyFromTarget(conn)

	return nil
}

// proxyToTarget проксирует данные от клиента к целевому серверу
func (p *MASQUEUDPProxy) proxyToTarget(conn *MASQUEConnection) {
	defer conn.UDPConn.Close()

	for {
		// Читаем данные из HTTP/3 stream
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

// proxyFromTarget проксирует данные от целевого сервера к клиенту
func (p *MASQUEUDPProxy) proxyFromTarget(conn *MASQUEConnection) {
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

		// Отправляем данные в HTTP/3 stream
		if err := conn.Stream.SendDatagram(buffer[:n]); err != nil {
			p.logger.Debug("Failed to send datagram to stream",
				zap.String("conn_id", conn.ID),
				zap.Error(err))
			break
		}

		conn.LastActivity = time.Now()
	}
}

// proxyToTarget проксирует данные от клиента к целевому серверу (IP)
func (p *MASQUEIPProxy) proxyToTarget(conn *MASQUEConnection) {
	// Аналогично UDP прокси, но для IP туннелирования
	p.proxyToTarget(conn)
}

// proxyFromTarget проксирует данные от целевого сервера к клиенту (IP)
func (p *MASQUEIPProxy) proxyFromTarget(conn *MASQUEConnection) {
	// Аналогично UDP прокси, но для IP туннелирования
	p.proxyFromTarget(conn)
}

// Close закрывает MASQUE сервер
func (ms *MASQUEServer) Close() error {
	ms.logger.Info("Closing MASQUE server")

	ms.cancel()

	// Закрываем все соединения
	ms.connectionsMutex.Lock()
	for connID, conn := range ms.connections {
		conn.Stream.Close()
		conn.UDPConn.Close()
		ms.logger.Debug("Closed MASQUE connection", zap.String("conn_id", connID))
	}
	ms.connections = make(map[string]*MASQUEConnection)
	ms.connectionsMutex.Unlock()

	// Закрываем HTTP/3 сервер
	if ms.http3Server != nil {
		if err := ms.http3Server.Close(); err != nil {
			ms.logger.Warn("Failed to close HTTP/3 server", zap.Error(err))
		}
	}

	ms.isRunning = false
	ms.logger.Info("MASQUE server closed successfully")

	return nil
}

// GetConnectionCount возвращает количество активных соединений
func (ms *MASQUEServer) GetConnectionCount() int {
	ms.connectionsMutex.RLock()
	defer ms.connectionsMutex.RUnlock()

	return len(ms.connections)
}

// GetConnectionInfo возвращает информацию о соединениях
func (ms *MASQUEServer) GetConnectionInfo() map[string]interface{} {
	ms.connectionsMutex.RLock()
	defer ms.connectionsMutex.RUnlock()

	connections := make([]map[string]interface{}, 0, len(ms.connections))
	for connID, conn := range ms.connections {
		connections = append(connections, map[string]interface{}{
			"id":            connID,
			"created_at":    conn.CreatedAt,
			"last_activity": conn.LastActivity,
			"target":        conn.UDPConn.RemoteAddr().String(),
		})
	}

	return map[string]interface{}{
		"total_connections": len(ms.connections),
		"is_running":        ms.isRunning,
		"connections":       connections,
	}
}
