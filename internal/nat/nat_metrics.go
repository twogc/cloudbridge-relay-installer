package nat

import (
	"sync"
	"time"
)

// NATMetrics представляет метрики NAT traversal
type NATMetrics struct {
	mu sync.RWMutex

	// STUN метрики
	STUNRequests            int           `json:"stun_requests"`
	STUNSuccesses           int           `json:"stun_successes"`
	STUNFailures            int           `json:"stun_failures"`
	AverageSTUNResponseTime time.Duration `json:"average_stun_response_time"`

	// ICE кандидаты
	HostCandidatesGathered            int `json:"host_candidates_gathered"`
	ServerReflexiveCandidatesGathered int `json:"server_reflexive_candidates_gathered"`
	RelayCandidatesGathered           int `json:"relay_candidates_gathered"`

	// NAT типы
	NATType           string `json:"nat_type"`
	OpenNATCount      int    `json:"open_nat_count"`
	SymmetricNATCount int    `json:"symmetric_nat_count"`
	ConeNATCount      int    `json:"cone_nat_count"`
	BlockedNATCount   int    `json:"blocked_nat_count"`

	// ICE connectivity checks
	ConnectivityChecks    int           `json:"connectivity_checks"`
	SuccessfulConnections int           `json:"successful_connections"`
	FailedConnections     int           `json:"failed_connections"`
	AverageConnectionTime time.Duration `json:"average_connection_time"`

	// TURN метрики
	TURNAllocations         int           `json:"turn_allocations"`
	TURNAllocationFailures  int           `json:"turn_allocation_failures"`
	ActiveTURNAllocations   int           `json:"active_turn_allocations"`
	TURNDataTransferred     int64         `json:"turn_data_transferred"`
	AverageTURNResponseTime time.Duration `json:"average_turn_response_time"`

	// Bandwidth метрики
	TotalBytesReceived int64 `json:"total_bytes_received"`
	TotalBytesSent     int64 `json:"total_bytes_sent"`
	PacketsReceived    int64 `json:"packets_received"`
	PacketsSent        int64 `json:"packets_sent"`
	PacketsLost        int64 `json:"packets_lost"`

	// Performance метрики
	RTT            time.Duration `json:"rtt"`
	Jitter         time.Duration `json:"jitter"`
	PacketLossRate float64       `json:"packet_loss_rate"`

	// Timing метрики
	LastUpdated        time.Time     `json:"last_updated"`
	GatheringStartTime time.Time     `json:"gathering_start_time"`
	GatheringDuration  time.Duration `json:"gathering_duration"`
}

// NewNATMetrics создает новые метрики NAT
func NewNATMetrics() *NATMetrics {
	return &NATMetrics{
		LastUpdated:        time.Now(),
		GatheringStartTime: time.Now(),
	}
}

// UpdateSTUNMetrics обновляет STUN метрики
func (nm *NATMetrics) UpdateSTUNMetrics(success bool, responseTime time.Duration) {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	nm.STUNRequests++
	if success {
		nm.STUNSuccesses++
		// Update average response time
		if nm.AverageSTUNResponseTime == 0 {
			nm.AverageSTUNResponseTime = responseTime
		} else {
			nm.AverageSTUNResponseTime = (nm.AverageSTUNResponseTime + responseTime) / 2
		}
	} else {
		nm.STUNFailures++
	}
	nm.LastUpdated = time.Now()
}

// UpdateConnectivityMetrics обновляет метрики connectivity
func (nm *NATMetrics) UpdateConnectivityMetrics(success bool, connectionTime time.Duration) {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	nm.ConnectivityChecks++
	if success {
		nm.SuccessfulConnections++
		// Update average connection time
		if nm.AverageConnectionTime == 0 {
			nm.AverageConnectionTime = connectionTime
		} else {
			nm.AverageConnectionTime = (nm.AverageConnectionTime + connectionTime) / 2
		}
	} else {
		nm.FailedConnections++
	}
	nm.LastUpdated = time.Now()
}

// UpdateTURNMetrics обновляет TURN метрики
func (nm *NATMetrics) UpdateTURNMetrics(allocated bool, responseTime time.Duration) {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	if allocated {
		nm.TURNAllocations++
		nm.ActiveTURNAllocations++
		// Update average TURN response time
		if nm.AverageTURNResponseTime == 0 {
			nm.AverageTURNResponseTime = responseTime
		} else {
			nm.AverageTURNResponseTime = (nm.AverageTURNResponseTime + responseTime) / 2
		}
	} else {
		nm.TURNAllocationFailures++
	}
	nm.LastUpdated = time.Now()
}

// UpdateBandwidthMetrics обновляет метрики bandwidth
func (nm *NATMetrics) UpdateBandwidthMetrics(bytesReceived, bytesSent int64, packetsReceived, packetsSent, packetsLost int64) {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	nm.TotalBytesReceived += bytesReceived
	nm.TotalBytesSent += bytesSent
	nm.PacketsReceived += packetsReceived
	nm.PacketsSent += packetsSent
	nm.PacketsLost += packetsLost

	// Calculate packet loss rate
	totalPackets := nm.PacketsReceived + nm.PacketsSent
	if totalPackets > 0 {
		nm.PacketLossRate = float64(nm.PacketsLost) / float64(totalPackets) * 100
	}

	nm.LastUpdated = time.Now()
}

// UpdatePerformanceMetrics обновляет метрики performance
func (nm *NATMetrics) UpdatePerformanceMetrics(rtt, jitter time.Duration) {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	nm.RTT = rtt
	nm.Jitter = jitter
	nm.LastUpdated = time.Now()
}

// FinishGathering завершает процесс сбора кандидатов
func (nm *NATMetrics) FinishGathering() {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	nm.GatheringDuration = time.Since(nm.GatheringStartTime)
	nm.LastUpdated = time.Now()
}

// GetSuccessRate возвращает процент успешных соединений
func (nm *NATMetrics) GetSuccessRate() float64 {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	total := nm.SuccessfulConnections + nm.FailedConnections
	if total == 0 {
		return 0
	}
	return float64(nm.SuccessfulConnections) / float64(total) * 100
}

// GetSTUNSuccessRate возвращает процент успешных STUN запросов
func (nm *NATMetrics) GetSTUNSuccessRate() float64 {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	if nm.STUNRequests == 0 {
		return 0
	}
	return float64(nm.STUNSuccesses) / float64(nm.STUNRequests) * 100
}

// GetTURNSuccessRate возвращает процент успешных TURN allocations
func (nm *NATMetrics) GetTURNSuccessRate() float64 {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	total := nm.TURNAllocations + nm.TURNAllocationFailures
	if total == 0 {
		return 0
	}
	return float64(nm.TURNAllocations) / float64(total) * 100
}

// GetTotalCandidates возвращает общее количество собранных кандидатов
func (nm *NATMetrics) GetTotalCandidates() int {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	return nm.HostCandidatesGathered +
		nm.ServerReflexiveCandidatesGathered +
		nm.RelayCandidatesGathered
}

// Clone создает копию метрик (для безопасного чтения)
func (nm *NATMetrics) Clone() *NATMetrics {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	// Создаем новую структуру и копируем поля вручную для избежания копирования мьютекса
	clone := &NATMetrics{
		STUNRequests:                      nm.STUNRequests,
		STUNSuccesses:                     nm.STUNSuccesses,
		STUNFailures:                      nm.STUNFailures,
		AverageSTUNResponseTime:           nm.AverageSTUNResponseTime,
		HostCandidatesGathered:            nm.HostCandidatesGathered,
		ServerReflexiveCandidatesGathered: nm.ServerReflexiveCandidatesGathered,
		RelayCandidatesGathered:           nm.RelayCandidatesGathered,
		NATType:                           nm.NATType,
		OpenNATCount:                      nm.OpenNATCount,
		SymmetricNATCount:                 nm.SymmetricNATCount,
		ConeNATCount:                      nm.ConeNATCount,
		BlockedNATCount:                   nm.BlockedNATCount,
		ConnectivityChecks:                nm.ConnectivityChecks,
		SuccessfulConnections:             nm.SuccessfulConnections,
		FailedConnections:                 nm.FailedConnections,
		AverageConnectionTime:             nm.AverageConnectionTime,
		TURNAllocations:                   nm.TURNAllocations,
		TURNAllocationFailures:            nm.TURNAllocationFailures,
		ActiveTURNAllocations:             nm.ActiveTURNAllocations,
		TURNDataTransferred:               nm.TURNDataTransferred,
		AverageTURNResponseTime:           nm.AverageTURNResponseTime,
		TotalBytesReceived:                nm.TotalBytesReceived,
		TotalBytesSent:                    nm.TotalBytesSent,
		PacketsReceived:                   nm.PacketsReceived,
		PacketsSent:                       nm.PacketsSent,
		PacketsLost:                       nm.PacketsLost,
		RTT:                               nm.RTT,
		Jitter:                            nm.Jitter,
		PacketLossRate:                    nm.PacketLossRate,
		LastUpdated:                       nm.LastUpdated,
		GatheringStartTime:                nm.GatheringStartTime,
		GatheringDuration:                 nm.GatheringDuration,
	}
	return clone
}

// Reset сбрасывает все метрики
func (nm *NATMetrics) Reset() {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	*nm = NATMetrics{
		LastUpdated:        time.Now(),
		GatheringStartTime: time.Now(),
	}
}

// NATDiagnostics представляет диагностическую информацию NAT
type NATDiagnostics struct {
	NATType            string        `json:"nat_type"`
	PublicIP           string        `json:"public_ip"`
	LocalIP            string        `json:"local_ip"`
	STUNServers        []string      `json:"stun_servers"`
	WorkingSTUNServers []string      `json:"working_stun_servers"`
	TURNServers        []string      `json:"turn_servers"`
	WorkingTURNServers []string      `json:"working_turn_servers"`
	PortMapping        string        `json:"port_mapping"`
	SymmetricNAT       bool          `json:"symmetric_nat"`
	RestrictedCone     bool          `json:"restricted_cone"`
	UPnPAvailable      bool          `json:"upnp_available"`
	Recommendations    []string      `json:"recommendations"`
	ConnectivityScore  int           `json:"connectivity_score"` // 0-100
	LastDiagnosticTime time.Time     `json:"last_diagnostic_time"`
	DiagnosticDuration time.Duration `json:"diagnostic_duration"`
}

// GenerateDiagnostics создает диагностическую информацию на основе метрик
func (nm *NATMetrics) GenerateDiagnostics(publicIP, localIP string, stunServers, turnServers []string) *NATDiagnostics {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	diag := &NATDiagnostics{
		NATType:            nm.NATType,
		PublicIP:           publicIP,
		LocalIP:            localIP,
		STUNServers:        stunServers,
		WorkingSTUNServers: make([]string, 0),
		TURNServers:        turnServers,
		WorkingTURNServers: make([]string, 0),
		SymmetricNAT:       nm.NATType == "Symmetric",
		RestrictedCone:     nm.NATType == "Restricted Cone" || nm.NATType == "Port Restricted",
		LastDiagnosticTime: time.Now(),
		DiagnosticDuration: time.Since(nm.GatheringStartTime),
	}

	// Calculate connectivity score
	score := 0

	// STUN success contributes 30 points
	if nm.STUNRequests > 0 {
		score += int(nm.GetSTUNSuccessRate() * 0.3)
	}

	// Connection success contributes 40 points
	if nm.ConnectivityChecks > 0 {
		score += int(nm.GetSuccessRate() * 0.4)
	}

	// NAT type contributes 20 points
	switch nm.NATType {
	case "Open":
		score += 20
	case "Full Cone":
		score += 15
	case "Restricted Cone":
		score += 10
	case "Port Restricted":
		score += 8
	case "Symmetric":
		score += 5
	case "Blocked":
		score += 0
	}

	// TURN availability contributes 10 points
	if nm.TURNAllocations > 0 {
		score += int(nm.GetTURNSuccessRate() * 0.1)
	}

	diag.ConnectivityScore = score

	// Generate recommendations
	diag.Recommendations = nm.generateRecommendations()

	return diag
}

// generateRecommendations генерирует рекомендации на основе метрик
func (nm *NATMetrics) generateRecommendations() []string {
	var recommendations []string

	// STUN related recommendations
	if nm.GetSTUNSuccessRate() < 50 {
		recommendations = append(recommendations,
			"Consider using additional STUN servers for better reliability")
	}

	// NAT type recommendations
	switch nm.NATType {
	case "Symmetric":
		recommendations = append(recommendations,
			"Symmetric NAT detected - TURN relay may be required for connectivity")
		recommendations = append(recommendations,
			"Consider configuring port forwarding or UPnP if available")
	case "Blocked":
		recommendations = append(recommendations,
			"Network appears to be blocked - check firewall settings")
		recommendations = append(recommendations,
			"TURN relay will be required for connectivity")
	}

	// Connection success recommendations
	if nm.GetSuccessRate() < 70 {
		recommendations = append(recommendations,
			"Low connection success rate - consider additional TURN servers")
	}

	// Performance recommendations
	if nm.RTT > 200*time.Millisecond {
		recommendations = append(recommendations,
			"High RTT detected - consider using geographically closer servers")
	}

	if nm.PacketLossRate > 5 {
		recommendations = append(recommendations,
			"High packet loss detected - check network quality")
	}

	// TURN recommendations
	if nm.TURNAllocations == 0 && (nm.NATType == "Symmetric" || nm.NATType == "Blocked") {
		recommendations = append(recommendations,
			"Configure TURN servers for reliable connectivity with this NAT type")
	}

	return recommendations
}
