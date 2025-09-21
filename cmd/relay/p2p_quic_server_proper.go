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
	"sync"
	"time"

	"github.com/quic-go/quic-go"
	"go.uber.org/zap"

	"github.com/twogc/cloudbridge-relay-installer/config"
	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
	"github.com/twogc/cloudbridge-relay-installer/internal/p2p"
)

// P2PQUICServerProper представляет правильную реализацию P2P QUIC сервера
// с использованием quic.Transport для работы с одним UDP сокетом
type P2PQUICServerProper struct {
	logger           *logger.Logger
	ctx              context.Context
	quicRelayHandler *QUICRelayHandler
	config           *config.Config

	// QUIC Transport - центральная точка для всех QUIC соединений
	transport *quic.Transport
	listener  *quic.Listener

	// UDP соединение для QUIC
	udpConn *net.UDPConn

	// Управление соединениями
	connections      map[string]*quic.Conn
	connectionsMutex sync.RWMutex

	// StatelessResetKey для корректного завершения соединений
	statelessResetManager *StatelessResetManager

	// ICE/STUN поддержка для NAT traversal
	iceSupport *ICESupport

	// MASQUE поддержка для обхода корпоративных сетей
	masqueServer *MASQUEServer

	// P2P Tunnel Manager для управления туннелями
	tunnelManager *p2p.P2PTunnelManager

	// QUIC Underlay Manager для управления QUIC соединениями
	quicUnderlayManager *p2p.QUICUnderlayManager

	// NAT Traversal Manager для NAT traversal
	natTraversalManager *p2p.NATTraversalManager
}

// NewP2PQUICServerProper создает новый правильный P2P QUIC сервер
func NewP2PQUICServerProper(logger *logger.Logger, ctx context.Context, quicRelayHandler *QUICRelayHandler, config *config.Config) *P2PQUICServerProper {
	// Создаем StatelessResetManager
	statelessResetManager := NewStatelessResetManager(logger, "/app/data/stateless_reset.key")

	// Создаем ICE поддержку
	var iceSupport *ICESupport
	if config.ICEStunTurn.Enabled {
		var err error
		iceSupport, err = NewICESupport(logger, config)
		if err != nil {
			logger.Warn("Failed to create ICE support", zap.Error(err))
		}
	}

	// Создаем MASQUE сервер
	var masqueServer *MASQUEServer
	if config.MASQUE.Enabled {
		masqueServer = NewMASQUEServer(logger, config)
	}

	// Создаем P2P Tunnel Manager
	tunnelManager := p2p.NewP2PTunnelManager(logger, quicRelayHandler.relayManager)

	// Создаем QUIC Underlay Manager
	quicUnderlayManager := p2p.NewQUICUnderlayManager(logger)

	// Создаем NAT Traversal Manager
	natTraversalManager := p2p.NewNATTraversalManager(logger)

	return &P2PQUICServerProper{
		logger:                logger,
		ctx:                   ctx,
		quicRelayHandler:      quicRelayHandler,
		config:                config,
		connections:           make(map[string]*quic.Conn),
		statelessResetManager: statelessResetManager,
		iceSupport:            iceSupport,
		masqueServer:          masqueServer,
		tunnelManager:         tunnelManager,
		quicUnderlayManager:   quicUnderlayManager,
		natTraversalManager:   natTraversalManager,
	}
}

// Start запускает правильный P2P QUIC сервер с quic.Transport
func (pqs *P2PQUICServerProper) Start() error {
	if pqs.quicRelayHandler == nil {
		return fmt.Errorf("QUIC relay handler is nil")
	}

	// Инициализируем StatelessResetManager
	if err := pqs.statelessResetManager.LoadOrGenerateKey(); err != nil {
		pqs.logger.Error("Failed to load or generate StatelessResetKey", zap.Error(err))
		return fmt.Errorf("failed to initialize StatelessResetManager: %v", err)
	}

	// Запускаем ICE поддержку
	if pqs.iceSupport != nil {
		if err := pqs.iceSupport.Start(pqs.ctx); err != nil {
			pqs.logger.Error("Failed to start ICE support", zap.Error(err))
			return fmt.Errorf("failed to start ICE support: %v", err)
		}
		pqs.logger.Info("ICE support started successfully")
	}

	// Запускаем MASQUE сервер
	if pqs.masqueServer != nil {
		if err := pqs.masqueServer.Start(); err != nil {
			pqs.logger.Error("Failed to start MASQUE server", zap.Error(err))
			return fmt.Errorf("failed to start MASQUE server: %v", err)
		}
		pqs.logger.Info("MASQUE server started successfully")
	}

	// Запускаем P2P Tunnel Manager
	if err := pqs.tunnelManager.Start(); err != nil {
		pqs.logger.Error("Failed to start P2P Tunnel Manager", zap.Error(err))
		return fmt.Errorf("failed to start P2P Tunnel Manager: %v", err)
	}
	pqs.logger.Info("P2P Tunnel Manager started successfully")

	// Запускаем QUIC Underlay Manager
	if err := pqs.quicUnderlayManager.Start("0.0.0.0:5554"); err != nil {
		pqs.logger.Error("Failed to start QUIC Underlay Manager", zap.Error(err))
		return fmt.Errorf("failed to start QUIC Underlay Manager: %v", err)
	}
	pqs.logger.Info("QUIC Underlay Manager started successfully")

	go func() {
		// Создаем UDP соединение для P2P QUIC
		udpAddr, err := net.ResolveUDPAddr("udp4", "0.0.0.0:5553")
		if err != nil {
			pqs.logger.Error("Failed to resolve UDP address for P2P QUIC", zap.Error(err))
			return
		}

		udpConn, err := net.ListenUDP("udp4", udpAddr)
		if err != nil {
			pqs.logger.Error("Failed to start P2P QUIC UDP server", zap.Error(err))
			return
		}
		pqs.udpConn = udpConn

		// Настраиваем UDP буферы для QUIC (обязательно для производительности)
		if err := udpConn.SetReadBuffer(7 * 1024 * 1024); err != nil {
			pqs.logger.Warn("Failed to set UDP read buffer size", zap.Error(err))
		}
		if err := udpConn.SetWriteBuffer(7 * 1024 * 1024); err != nil {
			pqs.logger.Warn("Failed to set UDP write buffer size", zap.Error(err))
		}

		pqs.logger.Info("P2P QUIC server started on port 5553")
		pqs.logger.Debug("P2P QUIC server details",
			zap.String("local_addr", udpConn.LocalAddr().String()),
			zap.Int("read_buffer_size", 7*1024*1024),
			zap.Int("write_buffer_size", 7*1024*1024))

		// Создаем QUIC Transport - центральная точка для всех QUIC соединений
		pqs.transport = &quic.Transport{
			Conn: udpConn,
		}

		// Устанавливаем StatelessResetKey
		pqs.statelessResetManager.SetTransport(pqs.transport)

		// Создаем QUIC Listener через Transport
		pqs.listener, err = pqs.transport.Listen(pqs.generateTLSConfig(), &quic.Config{
			MaxIdleTimeout:  30 * time.Second,
			KeepAlivePeriod: 15 * time.Second,
		})
		if err != nil {
			pqs.logger.Error("Failed to create QUIC listener for P2P relay", zap.Error(err))
			udpConn.Close()
			return
		}

		pqs.logger.Info("QUIC Transport and Listener created successfully")

		// Запускаем обработку входящих QUIC соединений
		pqs.handleIncomingConnections()

		// Запускаем обработку non-QUIC пакетов (STUN/TURN/DTLS/SRTP)
		go pqs.handleNonQUICPackets()

		// Ждем завершения контекста сервера
		<-pqs.ctx.Done()
		pqs.logger.Info("Shutting down P2P QUIC server")
		pqs.listener.Close()
		udpConn.Close()
	}()

	return nil
}

// handleIncomingConnections обрабатывает входящие QUIC соединения
func (pqs *P2PQUICServerProper) handleIncomingConnections() {
	for {
		select {
		case <-pqs.ctx.Done():
			pqs.logger.Info("P2P QUIC connection handler stopped")
			return
		default:
			pqs.logger.Debug("Waiting for QUIC connection...")

			// Принимаем входящее QUIC соединение
			conn, err := pqs.listener.Accept(pqs.ctx)
			if err != nil {
				if pqs.ctx.Err() != nil {
					pqs.logger.Info("Context cancelled, stopping QUIC connection handler")
					return
				}
				pqs.logger.Error("Failed to accept QUIC connection", zap.Error(err))
				continue
			}

			pqs.logger.Info("QUIC connection accepted",
				zap.String("remote_addr", conn.RemoteAddr().String()),
				zap.String("local_addr", conn.LocalAddr().String()))

			// Сохраняем соединение
			addrStr := conn.RemoteAddr().String()
			pqs.connectionsMutex.Lock()
			pqs.connections[addrStr] = conn
			pqs.connectionsMutex.Unlock()

			// Обрабатываем соединение в отдельной горутине
			go pqs.handleQUICConnection(conn, addrStr)
		}
	}
}

// handleNonQUICPackets обрабатывает non-QUIC пакеты (STUN/TURN/DTLS/SRTP)
// Использует Transport.ReadNonQUICPacket для демультиплексирования
func (pqs *P2PQUICServerProper) handleNonQUICPackets() {
	buffer := make([]byte, 1500) // Стандартный MTU

	for {
		select {
		case <-pqs.ctx.Done():
			pqs.logger.Info("Non-QUIC packet handler stopped")
			return
		default:
			// Читаем non-QUIC пакеты
			n, addr, err := pqs.transport.ReadNonQUICPacket(pqs.ctx, buffer)
			if err != nil {
				if pqs.ctx.Err() != nil {
					return
				}
				pqs.logger.Debug("No non-QUIC packets available", zap.Error(err))
				time.Sleep(100 * time.Millisecond) // Небольшая пауза
				continue
			}

			logBytes := n
			if logBytes > 8 {
				logBytes = 8
			}
			pqs.logger.Debug("Non-QUIC packet received",
				zap.String("from", addr.String()),
				zap.Int("size", n),
				zap.String("first_bytes", fmt.Sprintf("%x", buffer[:logBytes])))

			// Определяем тип пакета и обрабатываем
			go pqs.processNonQUICPacket(addr, buffer[:n])
		}
	}
}

// processNonQUICPacket обрабатывает non-QUIC пакет
func (pqs *P2PQUICServerProper) processNonQUICPacket(addr net.Addr, data []byte) {
	if len(data) < 4 {
		return
	}

	// Проверяем, является ли это STUN пакетом
	if pqs.isSTUNPacket(data) {
		pqs.logger.Debug("STUN packet detected", zap.String("from", addr.String()))
		// Здесь можно добавить обработку STUN пакетов для ICE
		return
	}

	// Проверяем, является ли это TURN пакетом
	if pqs.isTURNPacket(data) {
		pqs.logger.Debug("TURN packet detected", zap.String("from", addr.String()))
		// Здесь можно добавить обработку TURN пакетов
		return
	}

	// Проверяем, является ли это DTLS пакетом
	if pqs.isDTLSPacket(data) {
		pqs.logger.Debug("DTLS packet detected", zap.String("from", addr.String()))
		// Здесь можно добавить обработку DTLS пакетов
		return
	}

	logBytes := len(data)
	if logBytes > 8 {
		logBytes = 8
	}
	pqs.logger.Debug("Unknown non-QUIC packet type",
		zap.String("from", addr.String()),
		zap.String("first_bytes", fmt.Sprintf("%x", data[:logBytes])))
}

// isSTUNPacket проверяет, является ли пакет STUN пакетом
func (pqs *P2PQUICServerProper) isSTUNPacket(data []byte) bool {
	if len(data) < 20 {
		return false
	}
	// STUN пакеты начинаются с определенных битов
	return (data[0] & 0xC0) == 0x00
}

// isTURNPacket проверяет, является ли пакет TURN пакетом
func (pqs *P2PQUICServerProper) isTURNPacket(data []byte) bool {
	if len(data) < 20 {
		return false
	}
	// TURN пакеты начинаются с определенных битов
	return (data[0] & 0xC0) == 0x40
}

// isDTLSPacket проверяет, является ли пакет DTLS пакетом
func (pqs *P2PQUICServerProper) isDTLSPacket(data []byte) bool {
	if len(data) < 13 {
		return false
	}
	// DTLS пакеты имеют определенную структуру
	return data[0] >= 0x10 && data[0] <= 0x17
}

// handleQUICConnection обрабатывает QUIC соединение
func (pqs *P2PQUICServerProper) handleQUICConnection(conn *quic.Conn, addrStr string) {
	defer func() {
		// Удаляем соединение из списка
		pqs.connectionsMutex.Lock()
		delete(pqs.connections, addrStr)
		pqs.connectionsMutex.Unlock()

		// Закрываем соединение
		conn.CloseWithError(0, "connection closed")
		pqs.logger.Info("QUIC connection closed", zap.String("remote_addr", addrStr))
	}()

	pqs.logger.Info("New QUIC connection established",
		zap.String("remote_addr", conn.RemoteAddr().String()),
		zap.String("local_addr", conn.LocalAddr().String()))

	// Запускаем обработку соединения через QUIC Relay Handler
	if pqs.quicRelayHandler != nil {
		// Используем handleQUICConnection для обработки соединения
		go pqs.quicRelayHandler.handleQUICConnection(conn)
	}

	// Ждем завершения соединения
	<-conn.Context().Done()
}

// Dial создает исходящее QUIC соединение (для P2P)
func (pqs *P2PQUICServerProper) Dial(ctx context.Context, addr string) (*quic.Conn, error) {
	if pqs.transport == nil {
		return nil, fmt.Errorf("transport not initialized")
	}

	raddr, err := net.ResolveUDPAddr("udp4", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve address: %v", err)
	}

	tlsConf := &tls.Config{
		InsecureSkipVerify: false,
		NextProtos:         []string{"cloudbridge-p2p", "h3", "h3-29", "h3-28", "h3-27"},
		MinVersion:         tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
	}

	conn, err := pqs.transport.Dial(ctx, raddr, tlsConf, &quic.Config{
		MaxIdleTimeout:  30 * time.Second,
		KeepAlivePeriod: 15 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to dial QUIC connection: %v", err)
	}

	pqs.logger.Info("QUIC connection dialed successfully",
		zap.String("remote_addr", conn.RemoteAddr().String()),
		zap.String("local_addr", conn.LocalAddr().String()))

	return conn, nil
}

// generateTLSConfig генерирует TLS конфигурацию для P2P QUIC
func (pqs *P2PQUICServerProper) generateTLSConfig() *tls.Config {
	// Создаем самоподписанный сертификат для QUIC
	cert, err := pqs.generateSelfSignedCert()
	if err != nil {
		pqs.logger.Error("Failed to generate self-signed certificate", zap.Error(err))
		return nil
	}

	return &tls.Config{
		MinVersion: tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
		Certificates: []tls.Certificate{cert},
		NextProtos:   []string{"cloudbridge-p2p", "h3", "h3-29", "h3-28", "h3-27"},
	}
}

// generateSelfSignedCert генерирует самоподписанный сертификат
func (pqs *P2PQUICServerProper) generateSelfSignedCert() (tls.Certificate, error) {
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

// Stop останавливает P2P QUIC сервер и все его компоненты
func (pqs *P2PQUICServerProper) Stop() error {
	pqs.logger.Info("Stopping P2P QUIC server")

	// Закрываем QUIC listener
	if pqs.listener != nil {
		if err := pqs.listener.Close(); err != nil {
			pqs.logger.Warn("Failed to close QUIC listener", zap.Error(err))
		}
	}

	// Закрываем QUIC Transport
	if pqs.transport != nil {
		if err := pqs.transport.Close(); err != nil {
			pqs.logger.Warn("Failed to close QUIC transport", zap.Error(err))
		}
	}

	// Закрываем UDP соединение
	if pqs.udpConn != nil {
		if err := pqs.udpConn.Close(); err != nil {
			pqs.logger.Warn("Failed to close UDP connection", zap.Error(err))
		}
	}

	// Закрываем все QUIC соединения
	pqs.connectionsMutex.Lock()
	for connID, conn := range pqs.connections {
		if err := conn.CloseWithError(0, "server shutdown"); err != nil {
			pqs.logger.Warn("Failed to close QUIC connection",
				zap.String("conn_id", connID),
				zap.Error(err))
		}
	}
	pqs.connections = make(map[string]*quic.Conn)
	pqs.connectionsMutex.Unlock()

	// Закрываем ICE поддержку
	if pqs.iceSupport != nil {
		if err := pqs.iceSupport.Close(); err != nil {
			pqs.logger.Warn("Failed to close ICE support", zap.Error(err))
		}
	}

	// Закрываем MASQUE сервер
	if pqs.masqueServer != nil {
		if err := pqs.masqueServer.Close(); err != nil {
			pqs.logger.Warn("Failed to close MASQUE server", zap.Error(err))
		}
	}

	// Останавливаем P2P Tunnel Manager
	if err := pqs.tunnelManager.Stop(); err != nil {
		pqs.logger.Warn("Failed to stop P2P Tunnel Manager", zap.Error(err))
	}

	// Останавливаем QUIC Underlay Manager
	if err := pqs.quicUnderlayManager.Stop(); err != nil {
		pqs.logger.Warn("Failed to stop QUIC Underlay Manager", zap.Error(err))
	}

	pqs.logger.Info("P2P QUIC server stopped successfully")
	return nil
}
