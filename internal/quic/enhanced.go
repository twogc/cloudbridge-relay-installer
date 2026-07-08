package quic

import (
	"crypto/tls"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/twogc/cloudbridge-relay-installer/internal/quic/congestion"
	"go.uber.org/zap"
)

// EnhancedConfig конфигурация Enhanced QUIC
type EnhancedConfig struct {
	// Базовые настройки
	MaxStreams            int
	KeepAlive             bool
	IdleTimeout           time.Duration
	HandshakeTimeout      time.Duration
	MaxIncomingStreams    int
	MaxIncomingUniStreams int
	TLSConfig             *tls.Config

	// Flow control
	InitialStreamReceiveWindow     uint64
	MaxStreamReceiveWindow         uint64
	InitialConnectionReceiveWindow uint64
	MaxConnectionReceiveWindow     uint64

	// Congestion control
	CongestionControlAlgorithm string

	// Enhanced features
	TenantManager           *TenantManager
	EnableDDoSProtection    bool
	EnableGeographicRouting bool

	// Experimental features managers
	ExperimentalFeatures *ExperimentalFeaturesConfig
	CCManager            *congestion.CongestionControlManager // BBRv2 congestion control manager
	ACKFreqManager       *ACKFrequencyManager
	FECManager           *FECManager
	GreasingManager      *GreasingManager
	logger               *zap.Logger

	// Prometheus metrics for experimental features
	ExperimentalMetrics  *ExperimentalMetrics
}

// TenantManager управляет мультиарендностью
type TenantManager struct {
	tenants map[string]*TenantInfo
	mu      sync.RWMutex
}

// TenantInfo информация о тенанте
type TenantInfo struct {
	ID                string
	BandwidthLimit    int64
	ConnectionLimit   int
	ActiveConnections int32
	BytesReceived     int64
	BytesSent         int64
	StartTime         time.Time
	LastActivity      time.Time
}

// EnhancedConnection представляет Enhanced QUIC политики и телеметрию
type EnhancedConnection struct {
	tenantID       string
	bytesReceived  int64
	bytesSent      int64
	startTime      time.Time
	lastActivity   atomic.Value // time.Time
	BandwidthLimit int64
	config         *EnhancedConfig
	tenantInfo     *TenantInfo
	ddosProtection *DDoSProtection
	geoRouter      *GeographicRouter
}

// DDoSProtection обеспечивает защиту от DDoS атак
type DDoSProtection struct {
	rateLimiters    map[string]*RateLimiter
	blacklist       map[string]time.Time
	anomalyDetector *AnomalyDetector
	mu              sync.RWMutex
}

// RateLimiter ограничитель скорости
type RateLimiter struct {
	requests  int
	window    time.Duration
	lastReset time.Time
	mu        sync.Mutex
}

// AnomalyDetector детектор аномалий
type AnomalyDetector struct {
	thresholds map[string]float64
}

// GeographicRouter географический роутер
type GeographicRouter struct {
	routes map[string]string
}

// NewTenantManager создает новый менеджер тенантов
func NewTenantManager() *TenantManager {
	return &TenantManager{
		tenants: make(map[string]*TenantInfo),
	}
}

// InitializeExperimentalFeatures инициализирует экспериментальные функции для EnhancedConfig
func (ec *EnhancedConfig) InitializeExperimentalFeatures(logger *zap.Logger) error {
	if ec.ExperimentalFeatures == nil {
		return nil // Экспериментальные функции не включены
	}

	ec.logger = logger

	// Инициализация BBRv2 congestion control
	if ec.ExperimentalFeatures.EnableBBRv2 {
		logger.Info("Initializing BBRv2 congestion control")
		ec.CCManager = congestion.NewCongestionControlManager(
			logger,
			congestion.AlgorithmBBRv2,
			1460, // MTU
		)
	}

	// Инициализация ACK Frequency manager
	if ec.ExperimentalFeatures.EnableACKFrequency {
		logger.Info("Initializing ACK Frequency optimization")
		ec.ACKFreqManager = NewACKFrequencyManager(
			logger,
			ec.ExperimentalFeatures.ACKFrequencyConfig,
		)
	}

	// Инициализация FEC manager
	if ec.ExperimentalFeatures.EnableFEC {
		logger.Info("Initializing Forward Error Correction",
			zap.Float64("redundancy", ec.ExperimentalFeatures.FECRedundancy))

		fecConfig := &FECConfig{
			RedundancyRatio: ec.ExperimentalFeatures.FECRedundancy,
			Algorithm:       "xor",
			GroupSize:       10,
			EnableAdaptive:  true,
			TargetRecoveryRate: 0.95,
		}
		ec.FECManager = NewFECManager(logger, fecConfig)
	}

	// Инициализация QUIC Bit Greasing
	if ec.ExperimentalFeatures.EnableGreasing {
		logger.Info("Initializing QUIC Bit Greasing (RFC 9287)")
		ec.GreasingManager = NewGreasingManager(logger, DefaultGreasingConfig())
	}

	logger.Info("Experimental features initialized successfully",
		zap.Bool("bbrv2", ec.ExperimentalFeatures.EnableBBRv2),
		zap.Bool("ack_frequency", ec.ExperimentalFeatures.EnableACKFrequency),
		zap.Bool("fec", ec.ExperimentalFeatures.EnableFEC),
		zap.Bool("greasing", ec.ExperimentalFeatures.EnableGreasing))

	return nil
}

// StopExperimentalFeatures останавливает все экспериментальные функции
func (ec *EnhancedConfig) StopExperimentalFeatures() {
	if ec.FECManager != nil && ec.FECManager.IsActive() {
		ec.FECManager.Stop()
	}
	if ec.GreasingManager != nil && ec.GreasingManager.IsActive() {
		ec.GreasingManager.Stop()
	}
}

// AddTenant добавляет нового тенанта
func (tm *TenantManager) AddTenant(id string, bandwidthLimit int64, connectionLimit int) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.tenants[id] = &TenantInfo{
		ID:              id,
		BandwidthLimit:  bandwidthLimit,
		ConnectionLimit: connectionLimit,
		StartTime:       time.Now(),
		LastActivity:    time.Now(),
	}
}

// GetTenant получает информацию о тенанте
func (tm *TenantManager) GetTenant(id string) (*TenantInfo, bool) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	tenant, exists := tm.tenants[id]
	return tenant, exists
}

// CheckTenantLimits проверяет лимиты тенанта
func (tm *TenantManager) CheckTenantLimits(tenantID string) error {
	tenant, exists := tm.GetTenant(tenantID)
	if !exists {
		return fmt.Errorf("tenant %s not found", tenantID)
	}

	if atomic.LoadInt32(&tenant.ActiveConnections) >= int32(tenant.ConnectionLimit) {
		return fmt.Errorf("connection limit exceeded for tenant %s", tenantID)
	}

	return nil
}

// DialWithTenant создает соединение с поддержкой мультиарендности
// NewEnhancedPolicies создает Enhanced политики для QUIC соединения
func NewEnhancedPolicies(tenantID string, config *EnhancedConfig) (*EnhancedConnection, error) {
	// Проверяем лимиты тенанта
	if config.TenantManager != nil {
		if err := config.TenantManager.CheckTenantLimits(tenantID); err != nil {
			return nil, err
		}
	}

	enhancedConn := &EnhancedConnection{
		tenantID:     tenantID,
		startTime:    time.Now(),
		config:       config,
	}

	// Инициализируем lastActivity
	enhancedConn.lastActivity.Store(time.Now())

	// Инициализируем компоненты
	if config.EnableDDoSProtection {
		enhancedConn.ddosProtection = NewDDoSProtection()
	}

	if config.EnableGeographicRouting {
		enhancedConn.geoRouter = NewGeographicRouter()
	}

	// Получаем информацию о тенанте
	if config.TenantManager != nil {
		if tenant, exists := config.TenantManager.GetTenant(tenantID); exists {
			enhancedConn.tenantInfo = tenant
			enhancedConn.BandwidthLimit = tenant.BandwidthLimit
			atomic.AddInt32(&tenant.ActiveConnections, 1)
		}
	}

	return enhancedConn, nil
}

// BeforeRead вызывается перед чтением данных
func (ec *EnhancedConnection) BeforeRead(remote string, nHint int) error {
	// Проверяем DDoS защиту
	if ec.ddosProtection != nil {
		if err := ec.ddosProtection.CheckRateLimit(remote); err != nil {
			return err
		}
	}
	return nil
}

// AfterRead вызывается после успешного чтения данных
func (ec *EnhancedConnection) AfterRead(n int) {
	atomic.AddInt64(&ec.bytesReceived, int64(n))
	ec.lastActivity.Store(time.Now())

	// Обновляем статистику тенанта
	if ec.tenantInfo != nil {
		atomic.AddInt64(&ec.tenantInfo.BytesReceived, int64(n))
		// TODO: Добавить thread-safe обновление LastActivity в TenantManager
	}

	// Интеграция с ACK Frequency manager
	if ec.config != nil && ec.config.ACKFreqManager != nil {
		// Регистрируем соединение при первом чтении
		connID := fmt.Sprintf("%s-%p", ec.tenantID, ec)
		// ACK manager будет автоматически вызываться при получении пакетов
		shouldACK := ec.config.ACKFreqManager.ShouldSendACK(connID, n)

		// Экспортируем метрики в Prometheus
		if ec.config.ExperimentalMetrics != nil {
			ackMetrics := ec.config.ACKFreqManager.GetMetrics()
			ec.config.ExperimentalMetrics.UpdateACKFrequencyMetrics(ackMetrics)

			// Записываем событие ACK
			if shouldACK {
				// Получаем текущий ACK delay из ACKFreqManager
				connMetrics := ec.config.ACKFreqManager.GetConnectionMetrics(connID)
				if connMetrics != nil {
					delay := connMetrics.AverageDelay.Seconds()
					ec.config.ExperimentalMetrics.RecordACK(true, true, delay)
				}
			}

			// Записываем per-connection метрики
			ec.config.ExperimentalMetrics.RecordACKFrequencyConnection(connID, ec.tenantID, 1, int64(n))
		}
	}
}

// BeforeWrite вызывается перед записью данных
func (ec *EnhancedConnection) BeforeWrite(remote string, n int) error {
	// Проверяем лимит пропускной способности
	if ec.BandwidthLimit > 0 {
		if atomic.LoadInt64(&ec.bytesSent) >= ec.BandwidthLimit {
			return fmt.Errorf("bandwidth limit exceeded")
		}
	}
	return nil
}

// AfterWrite вызывается после успешной записи данных
func (ec *EnhancedConnection) AfterWrite(n int) {
	atomic.AddInt64(&ec.bytesSent, int64(n))
	ec.lastActivity.Store(time.Now())

	// Обновляем статистику тенанта
	if ec.tenantInfo != nil {
		atomic.AddInt64(&ec.tenantInfo.BytesSent, int64(n))
		// TODO: Добавить thread-safe обновление LastActivity в TenantManager
	}
}

// Close закрывает соединение
func (ec *EnhancedConnection) Close() error {
	// Уменьшаем счетчик активных соединений тенанта
	if ec.tenantInfo != nil {
		// Безопасный декремент
		for {
			cur := atomic.LoadInt32(&ec.tenantInfo.ActiveConnections)
			if cur == 0 {
				break
			}
			if atomic.CompareAndSwapInt32(&ec.tenantInfo.ActiveConnections, cur, cur-1) {
				break
			}
		}
	}

	// Отключаем ACK Frequency manager
	if ec.config != nil && ec.config.ACKFreqManager != nil {
		connID := fmt.Sprintf("%s-%p", ec.tenantID, ec)
		ec.config.ACKFreqManager.UnregisterConnection(connID)
	}

	return nil
}

// OnPacketAck вызывается при получении ACK для пакета
func (ec *EnhancedConnection) OnPacketAck(bytesAcked int, rtt time.Duration) {
	if ec.config == nil || ec.config.CCManager == nil {
		return
	}

	// Обновляем BBRv2 congestion control
	now := time.Now()
	cwnd, pacingRate := ec.config.CCManager.OnAck(now, bytesAcked, rtt)

	// Экспортируем метрики в Prometheus
	if ec.config.ExperimentalMetrics != nil {
		connID := fmt.Sprintf("%s-%p", ec.tenantID, ec)
		metrics := ec.config.CCManager.GetMetrics()
		ec.config.ExperimentalMetrics.UpdateBBRv2Metrics(connID, ec.tenantID, metrics)
	}

	// Логируем изменения окна перегрузки
	if ec.config.logger != nil {
		ec.config.logger.Debug("BBRv2 ACK received",
			zap.Int("bytes_acked", bytesAcked),
			zap.Duration("rtt", rtt),
			zap.Int("new_cwnd", cwnd),
			zap.Int64("pacing_rate_bps", pacingRate))
	}
}

// OnPacketLoss вызывается при потере пакета
func (ec *EnhancedConnection) OnPacketLoss(bytesLost int) {
	if ec.config == nil || ec.config.CCManager == nil {
		return
	}

	// Обновляем BBRv2 на потерю пакета
	cwnd, pacingRate := ec.config.CCManager.OnLoss()

	// Логируем потерю пакета
	if ec.config.logger != nil {
		ec.config.logger.Warn("Packet loss detected",
			zap.Int("bytes_lost", bytesLost),
			zap.Int("new_cwnd", cwnd),
			zap.Int64("pacing_rate_bps", pacingRate))
	}
}

// ApplyFEC применяет Forward Error Correction к пакетам
func (ec *EnhancedConnection) ApplyFEC(packets [][]byte) ([][]byte, error) {
	if ec.config == nil || ec.config.FECManager == nil {
		return packets, nil // FEC не включен
	}

	// Создаем FEC группу
	startTime := time.Now()
	group, err := ec.config.FECManager.CreateGroup(packets)
	latency := time.Since(startTime).Seconds()

	if err != nil {
		return packets, fmt.Errorf("failed to create FEC group: %w", err)
	}

	// Возвращаем данные + избыточные пакеты
	result := make([][]byte, 0, len(group.DataPackets)+len(group.RedundantPackets))
	result = append(result, group.DataPackets...)
	result = append(result, group.RedundantPackets...)

	// Экспортируем метрики в Prometheus
	if ec.config.ExperimentalMetrics != nil {
		// Подсчитываем размер избыточности
		redundancyBytes := 0
		for _, pkt := range group.RedundantPackets {
			redundancyBytes += len(pkt)
		}

		// Записываем создание FEC группы
		ec.config.ExperimentalMetrics.RecordFECGroup(
			"xor", // algorithm - можно получить из config
			ec.tenantID,
			len(group.DataPackets),
			len(group.RedundantPackets),
			redundancyBytes,
		)

		// Обновляем общие FEC метрики
		fecMetrics := ec.config.FECManager.GetMetrics()
		ec.config.ExperimentalMetrics.UpdateFECMetrics(fecMetrics)

		// Записываем latency FEC операции
		ec.config.ExperimentalMetrics.RecordFECRecovery(true, 0, latency)
	}

	if ec.config.logger != nil {
		ec.config.logger.Debug("FEC applied",
			zap.Int("data_packets", len(group.DataPackets)),
			zap.Int("redundant_packets", len(group.RedundantPackets)))
	}

	return result, nil
}

// GreaseConnection применяет QUIC Bit Greasing к соединению
func (ec *EnhancedConnection) GreaseConnection(versions []uint32, params map[uint64][]byte) ([]uint32, map[uint64][]byte, error) {
	if ec.config == nil || ec.config.GreasingManager == nil {
		return versions, params, nil // Greasing не включен
	}

	// Применяем greasing к версиям
	greasedVersions, err := ec.config.GreasingManager.GreaseVersionNegotiation(versions)
	if err != nil {
		return versions, params, fmt.Errorf("failed to grease versions: %w", err)
	}

	// Экспортируем метрику version greasing
	if ec.config.ExperimentalMetrics != nil && len(greasedVersions) > len(versions) {
		ec.config.ExperimentalMetrics.RecordGreasingApplication("version")
	}

	// Применяем greasing к параметрам транспорта
	greasedParams, err := ec.config.GreasingManager.GreaseTransportParameters(params)
	if err != nil {
		return greasedVersions, params, fmt.Errorf("failed to grease parameters: %w", err)
	}

	// Экспортируем метрику transport param greasing
	if ec.config.ExperimentalMetrics != nil && len(greasedParams) > len(params) {
		ec.config.ExperimentalMetrics.RecordGreasingApplication("transport_param")
	}

	// Обновляем общие greasing метрики
	if ec.config.ExperimentalMetrics != nil {
		greasingMetrics := ec.config.GreasingManager.GetMetrics()
		ec.config.ExperimentalMetrics.UpdateGreasingMetrics(greasingMetrics)
	}

	if ec.config.logger != nil {
		ec.config.logger.Debug("QUIC Bit Greasing applied",
			zap.Int("versions", len(greasedVersions)),
			zap.Int("params", len(greasedParams)))
	}

	return greasedVersions, greasedParams, nil
}

// GetStats возвращает статистику соединения
func (ec *EnhancedConnection) GetStats() map[string]interface{} {
	lastActivity := ec.lastActivity.Load()
	var lastActivityTime time.Time
	if lastActivity != nil {
		lastActivityTime = lastActivity.(time.Time)
	}
	
	return map[string]interface{}{
		"tenant_id":      ec.tenantID,
		"bytes_received": atomic.LoadInt64(&ec.bytesReceived),
		"bytes_sent":     atomic.LoadInt64(&ec.bytesSent),
		"start_time":     ec.startTime.UTC().Format(time.RFC3339Nano),
		"last_activity":  lastActivityTime.UTC().Format(time.RFC3339Nano),
		"duration":       time.Since(ec.startTime).String(),
	}
}

// IsHealthy проверяет, что enhanced соединение здорово
func (ec *EnhancedConnection) IsHealthy() bool {
	// Проверяем, что соединение не закрыто
	lastActivity := ec.lastActivity.Load()
	if lastActivity == nil {
		return false
	}
	
	lastActivityTime := lastActivity.(time.Time)
	if time.Since(lastActivityTime) > 5*time.Minute {
		return false
	}
	
	// Проверяем лимиты пропускной способности
	if ec.BandwidthLimit > 0 && atomic.LoadInt64(&ec.bytesSent) >= ec.BandwidthLimit {
		return false
	}
	
	return true
}

// NewDDoSProtection создает новый DDoS защитник
func NewDDoSProtection() *DDoSProtection {
	return &DDoSProtection{
		rateLimiters:    make(map[string]*RateLimiter),
		blacklist:       make(map[string]time.Time),
		anomalyDetector: NewAnomalyDetector(),
	}
}

// CheckRateLimit проверяет лимит скорости
func (ddp *DDoSProtection) CheckRateLimit(clientIP string) error {
	ddp.mu.Lock()
	defer ddp.mu.Unlock()

	// Проверяем черный список
	if blacklistTime, exists := ddp.blacklist[clientIP]; exists {
		if time.Since(blacklistTime) < 5*time.Minute {
			return fmt.Errorf("IP %s is blacklisted", clientIP)
		}
		delete(ddp.blacklist, clientIP)
	}

	// Получаем или создаем rate limiter
	limiter, exists := ddp.rateLimiters[clientIP]
	if !exists {
		limiter = &RateLimiter{
			requests:  100,
			window:    time.Minute,
			lastReset: time.Now(),
		}
		ddp.rateLimiters[clientIP] = limiter
	}

	// Проверяем лимит
	if !limiter.Allow() {
		ddp.blacklist[clientIP] = time.Now()
		return fmt.Errorf("rate limit exceeded for IP %s", clientIP)
	}

	return nil
}

// Allow проверяет, разрешен ли запрос
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	if now.Sub(rl.lastReset) >= rl.window {
		rl.requests = 100
		rl.lastReset = now
	}

	if rl.requests > 0 {
		rl.requests--
		return true
	}

	return false
}

// NewAnomalyDetector создает новый детектор аномалий
func NewAnomalyDetector() *AnomalyDetector {
	return &AnomalyDetector{
		thresholds: map[string]float64{
			"connection_rate": 1000.0,
			"bandwidth_usage": 1000000.0, // 1MB/s
		},
	}
}

// NewGeographicRouter создает новый географический роутер
func NewGeographicRouter() *GeographicRouter {
	return &GeographicRouter{
		routes: map[string]string{
			"US": "us-east.relay.example.com",
			"EU": "eu-west.relay.example.com",
			"AS": "asia.relay.example.com",
		},
	}
}

// RouteByLocation маршрутизирует по географии
func (gr *GeographicRouter) RouteByLocation(clientIP string) string {
	// Простая логика определения региона по IP
	// В реальной реализации здесь будет GeoIP lookup
	if gr.isUSIP(clientIP) {
		return gr.routes["US"]
	} else if gr.isEUIP(clientIP) {
		return gr.routes["EU"]
	} else {
		return gr.routes["AS"]
	}
}

// isUSIP проверяет, является ли IP американским
func (gr *GeographicRouter) isUSIP(ip string) bool {
	// Упрощенная логика для примера
	return len(ip) > 0 && ip[0] == '1'
}

// isEUIP проверяет, является ли IP европейским
func (gr *GeographicRouter) isEUIP(ip string) bool {
	// Упрощенная логика для примера
	return len(ip) > 0 && ip[0] == '2'
}
