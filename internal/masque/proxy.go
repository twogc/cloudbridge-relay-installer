package masque

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"github.com/quic-go/quic-go/quicvarint"
	"go.uber.org/zap"

	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
)

const (
	uriTemplateTargetHost = "target_host"
	uriTemplateTargetPort = "target_port"
)

var contextIDZero = quicvarint.Append([]byte{}, 0)

// MASQUEProxy представляет MASQUE прокси для CONNECT-UDP (RFC 9298)
type MASQUEProxy struct {
	logger     *logger.Logger
	config     *Config
	quicServer *quic.Listener
	httpServer *http3.Server
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup

	// Proxy entries management (from masque-go reference)
	mx       sync.Mutex
	closed   bool
	refCount sync.WaitGroup
	closers  map[io.Closer]struct{}

	// Metrics
	metrics *MASQUEMetrics
}

// Config конфигурация MASQUE прокси
type Config struct {
	ListenAddr string
	CertFile   string
	KeyFile    string
}

// proxyEntry represents a MASQUE proxy entry (from masque-go reference)
type proxyEntry struct {
	str  *http3.Stream
	conn *net.UDPConn
}

func (e proxyEntry) Close() error {
	e.str.CancelRead(quic.StreamErrorCode(http3.ErrCodeConnectError))
	return errors.Join(e.str.Close(), e.conn.Close())
}

// Request represents a CONNECT-UDP request (from masque-go reference)
type Request struct {
	Host string
}

// NewMASQUEProxy создает новый MASQUE прокси
func NewMASQUEProxy(logger *logger.Logger, config *Config) *MASQUEProxy {
	ctx, cancel := context.WithCancel(context.Background())

	return &MASQUEProxy{
		logger:  logger,
		config:  config,
		ctx:     ctx,
		cancel:  cancel,
		closers: make(map[io.Closer]struct{}),
		metrics: NewMASQUEMetrics(),
	}
}

// Start запускает MASQUE прокси
func (mp *MASQUEProxy) Start() error {
	// Создаем TLS конфигурацию
	cert, err := tls.LoadX509KeyPair(mp.config.CertFile, mp.config.KeyFile)
	if err != nil {
		return fmt.Errorf("failed to load certificate: %v", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		NextProtos:   []string{"h3"},
	}

	// Создаем QUIC сервер
	udpAddr, err := net.ResolveUDPAddr("udp", mp.config.ListenAddr)
	if err != nil {
		return fmt.Errorf("failed to resolve address: %v", err)
	}

	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return fmt.Errorf("failed to listen UDP: %v", err)
	}

	quicConfig := &quic.Config{
		EnableDatagrams: true,
	}

	listener, err := quic.Listen(udpConn, tlsConfig, quicConfig)
	if err != nil {
		return fmt.Errorf("failed to create QUIC listener: %v", err)
	}

	mp.quicServer = listener

	// Создаем HTTP/3 сервер
	mp.httpServer = &http3.Server{
		Handler: mp,
	}

	mp.logger.Info("MASQUE proxy started", zap.String("addr", mp.config.ListenAddr))

	// Устанавливаем статус прокси как активный
	mp.metrics.SetProxyStatus(true)

	// Запускаем обработку соединений
	go mp.handleQUICConnections()

	return nil
}

// Stop останавливает MASQUE прокси
func (mp *MASQUEProxy) Stop() error {
	mp.mx.Lock()
	if mp.closed {
		mp.mx.Unlock()
		return nil
	}
	mp.closed = true

	// Close all proxy entries
	for closer := range mp.closers {
		closer.Close()
	}
	mp.mx.Unlock()

	mp.cancel()
	mp.refCount.Wait() // Wait for all proxy operations to complete
	mp.wg.Wait()

	// Устанавливаем статус прокси как неактивный
	mp.metrics.SetProxyStatus(false)

	if mp.quicServer != nil {
		return mp.quicServer.Close()
	}

	return nil
}

// ServeHTTP обрабатывает HTTP запросы (CONNECT-UDP protocol)
func (mp *MASQUEProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mp.logger.Info("Received HTTP request",
		zap.String("method", r.Method),
		zap.String("url", r.URL.String()),
		zap.String("protocol", r.Header.Get("Capsule-Protocol")))

	if r.Method == http.MethodConnect {
		// Parse CONNECT-UDP request
		req := &Request{Host: r.URL.Host}
		if err := mp.Proxy(w, req); err != nil {
			mp.logger.Error("MASQUE proxy failed", zap.Error(err))
		}
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Proxy proxies a CONNECT-UDP request (based on masque-go reference implementation)
func (mp *MASQUEProxy) Proxy(w http.ResponseWriter, r *Request) error {
	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime).Seconds()
		mp.metrics.ObserveRequestDuration(duration)
	}()

	mp.mx.Lock()
	if mp.closed {
		mp.mx.Unlock()
		w.WriteHeader(http.StatusServiceUnavailable)
		mp.metrics.IncrementConnectionErrors()
		return net.ErrClosed
	}
	mp.mx.Unlock()

	// Simplified proxy status handling (without httpsfv dependency)
	writeProxyStatus := func(err error) error {
		if err != nil {
			w.Header().Set("Proxy-Status", fmt.Sprintf("%s; details=%s", r.Host, err.Error()))
		} else {
			w.Header().Set("Proxy-Status", r.Host)
		}
		return err
	}

	// Resolve target address
	udpAddr, err := net.ResolveUDPAddr("udp", r.Host)
	if err != nil {
		mp.logger.Error("Failed to resolve target", zap.String("target", r.Host), zap.Error(err))
		mp.metrics.IncrementTargetResolutionErrors()
		w.WriteHeader(errToStatus(err))
		return writeProxyStatus(err)
	}
	mp.metrics.IncrementTargetResolutions()

	// Create UDP connection to target
	udpConn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		mp.logger.Error("Failed to dial target", zap.String("target", r.Host), zap.Error(err))
		mp.metrics.IncrementConnectionErrors()
		w.WriteHeader(errToStatus(err))
		return writeProxyStatus(err)
	}

	// Set required headers and send successful response
	w.Header().Set("Capsule-Protocol", "?1")
	w.WriteHeader(http.StatusOK)

	if err := writeProxyStatus(nil); err != nil {
		udpConn.Close()
		return err
	}

	// Get HTTP/3 stream
	str := w.(http3.HTTPStreamer).HTTPStream()

	// Create proxy entry
	entry := &proxyEntry{str: str, conn: udpConn}

	// Register proxy entry for cleanup
	mp.mx.Lock()
	if mp.closed {
		mp.mx.Unlock()
		entry.Close()
		return net.ErrClosed
	}
	mp.closers[entry] = struct{}{}
	mp.refCount.Add(1)
	mp.mx.Unlock()

	// Increment connection metrics
	mp.metrics.IncrementConnections()

	// Start data relaying
	go func() {
		defer func() {
			mp.mx.Lock()
			delete(mp.closers, entry)
			mp.mx.Unlock()
			entry.Close()
			mp.refCount.Done()
			// Decrement active connections when proxy entry is closed
			mp.metrics.DecrementConnections()
			mp.metrics.IncrementClosedConnections()
		}()

		mp.relayData(str, udpConn)
	}()

	mp.logger.Info("MASQUE proxy established", zap.String("target", r.Host))
	return nil
}

// errToStatus converts error to HTTP status code (from masque-go reference)
func errToStatus(err error) int {
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return http.StatusGatewayTimeout
	}
	var dnsError *net.DNSError
	if errors.As(err, &dnsError) {
		return http.StatusBadGateway
	}
	var addrErr *net.AddrError
	var parseError *net.ParseError
	if errors.As(err, &addrErr) || errors.As(err, &parseError) {
		return http.StatusBadRequest
	}
	return http.StatusInternalServerError
}

// handleQUICConnections обрабатывает входящие QUIC соединения
func (mp *MASQUEProxy) handleQUICConnections() {
	for {
		select {
		case <-mp.ctx.Done():
			return
		default:
			conn, err := mp.quicServer.Accept(mp.ctx)
			if err != nil {
				if mp.ctx.Err() != nil {
					return
				}
				mp.logger.Error("Failed to accept QUIC connection", zap.Error(err))
				continue
			}

			go mp.handleQUICConnection(conn)
		}
	}
}

// handleQUICConnection обрабатывает QUIC соединение
func (mp *MASQUEProxy) handleQUICConnection(conn *quic.Conn) {
	defer conn.CloseWithError(0, "connection closed")

	mp.logger.Info("New QUIC connection",
		zap.String("remote_addr", conn.RemoteAddr().String()))

	// Serve HTTP/3 over this QUIC connection
	mp.httpServer.ServeQUICConn(conn)
}

// relayData релеирует данные между HTTP/3 stream и UDP соединением
func (mp *MASQUEProxy) relayData(stream *http3.Stream, udpConn *net.UDPConn) {
	mp.wg.Add(2)

	// Relay from HTTP/3 datagrams to UDP
	go func() {
		defer mp.wg.Done()
		mp.relayFromStreamToUDP(stream, udpConn)
	}()

	// Relay from UDP to HTTP/3 datagrams
	go func() {
		defer mp.wg.Done()
		mp.relayFromUDPToStream(udpConn, stream)
	}()
}

// relayFromStreamToUDP релеирует данные из HTTP/3 datagrams в UDP
func (mp *MASQUEProxy) relayFromStreamToUDP(stream *http3.Stream, udpConn *net.UDPConn) {
	for {
		select {
		case <-mp.ctx.Done():
			return
		default:
			// Receive HTTP datagram
			data, err := stream.ReceiveDatagram(mp.ctx)
			if err != nil {
				if !errors.Is(err, context.Canceled) {
					mp.logger.Debug("Failed to receive datagram", zap.Error(err))
					mp.metrics.IncrementUDPErrors()
				}
				return
			}

			// Parse context ID (RFC 9297 HTTP Datagrams)
			contextID, n, err := quicvarint.Parse(data)
			if err != nil {
				mp.logger.Warn("Malformed HTTP datagram", zap.Error(err))
				continue
			}

			// For CONNECT-UDP, context ID should be 0
			if contextID != 0 {
				mp.logger.Debug("Dropping datagram with non-zero context ID", zap.Uint64("context_id", contextID))
				continue
			}

			// Extract UDP payload and send to target
			payload := data[n:]
			_, err = udpConn.Write(payload)
			if err != nil {
				mp.logger.Debug("Failed to write to UDP", zap.Error(err))
				mp.metrics.IncrementUDPErrors()
				return
			}

			// Update metrics
			mp.metrics.AddBytesSent(int64(len(payload)))
			mp.metrics.IncrementPacketsSent()
		}
	}
}

// relayFromUDPToStream релеирует данные из UDP в HTTP/3 datagrams
func (mp *MASQUEProxy) relayFromUDPToStream(udpConn *net.UDPConn, stream *http3.Stream) {
	buffer := make([]byte, 1500) // Standard MTU size

	for {
		select {
		case <-mp.ctx.Done():
			return
		default:
			// Read from UDP
			n, err := udpConn.Read(buffer)
			if err != nil {
				if !errors.Is(err, net.ErrClosed) {
					mp.logger.Debug("Failed to read from UDP", zap.Error(err))
					mp.metrics.IncrementUDPErrors()
				}
				return
			}

			// Prepare HTTP datagram with context ID 0 for CONNECT-UDP
			data := make([]byte, 0, len(contextIDZero)+n)
			data = append(data, contextIDZero...)
			data = append(data, buffer[:n]...)

			// Send HTTP datagram
			err = stream.SendDatagram(data)
			if err != nil {
				mp.logger.Debug("Failed to send datagram", zap.Error(err))
				mp.metrics.IncrementUDPErrors()
				return
			}

			// Update metrics
			mp.metrics.AddBytesReceived(int64(n))
			mp.metrics.IncrementPacketsReceived()
		}
	}
}
