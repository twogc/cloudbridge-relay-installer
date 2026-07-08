// CloudBridge Enterprise Networking - VXLAN Overlay Manager  
// =========================================================

package networking

import (
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"sync"
	"time"

	"go.uber.org/zap"
)

// VXLANOverlayManager управляет VXLAN overlay сетями для tenant'ов
type VXLANOverlayManager struct {
	overlays      map[string]*VXLANOverlay  // tenantID -> overlay
	vniCounter    int32                     // VNI (VXLAN Network Identifier) counter
	bridgeCounter int32                     // Bridge counter
	mutex         sync.RWMutex
	logger        *zap.Logger
	underlayNet   *net.IPNet                // Underlay сеть (например, 192.168.1.0/24)
}

// VXLANOverlay представляет VXLAN overlay для tenant'а
type VXLANOverlay struct {
	TenantID      string
	VNI           int32           // VXLAN Network ID (уникальный для tenant'а)
	BridgeName    string          // br-tenant-001, br-tenant-002
	VXLANDevice   string          // vxlan-001, vxlan-002  
	OverlayNet    *net.IPNet      // Overlay сеть (10.100.77.0/24)
	UnderlayPeers map[string]string // peerID -> underlay IP
	CreatedAt     time.Time
	
	// VTEP (VXLAN Tunnel Endpoint) информация
	LocalVTEP     string          // IP адрес локального VTEP
	RemoteVTEPs   []string        // IP адреса удаленных VTEP'ов
}

// NewVXLANOverlayManager создает новый overlay менеджер
func NewVXLANOverlayManager(underlayNet *net.IPNet, logger *zap.Logger) *VXLANOverlayManager {
	return &VXLANOverlayManager{
		overlays:      make(map[string]*VXLANOverlay),
		vniCounter:    1000,  // Начинаем с VNI 1001
		bridgeCounter: 100,   // Начинаем с br-100
		underlayNet:   underlayNet,
		logger:        logger,
	}
}

// CreateTenantOverlay создает VXLAN overlay для tenant'а
func (vm *VXLANOverlayManager) CreateTenantOverlay(tenantID string, overlayNet *net.IPNet) (*VXLANOverlay, error) {
	vm.mutex.Lock()
	defer vm.mutex.Unlock()

	// Проверяем существующий overlay
	if existing, exists := vm.overlays[tenantID]; exists {
		return existing, nil
	}

	// Выделяем VNI и bridge ID
	vm.vniCounter++
	vm.bridgeCounter++
	
	vni := vm.vniCounter
	bridgeName := fmt.Sprintf("br-tenant-%d", vm.bridgeCounter)
	vxlanDevice := fmt.Sprintf("vxlan-%d", vni)

	// Получаем локальный VTEP IP (первый доступный в underlay)
	localVTEP := vm.getNextUnderlayIP()

	overlay := &VXLANOverlay{
		TenantID:      tenantID,
		VNI:           vni,
		BridgeName:    bridgeName,
		VXLANDevice:   vxlanDevice,
		OverlayNet:    overlayNet,
		UnderlayPeers: make(map[string]string),
		LocalVTEP:     localVTEP,
		RemoteVTEPs:   make([]string, 0),
		CreatedAt:     time.Now(),
	}

	// Создаем VXLAN инфраструктуру
	err := vm.createVXLANInfrastructure(overlay)
	if err != nil {
		return nil, fmt.Errorf("failed to create VXLAN infrastructure: %w", err)
	}

	vm.overlays[tenantID] = overlay

	vm.logger.Info("Created tenant VXLAN overlay",
		zap.String("tenant_id", tenantID),
		zap.Int32("vni", vni),
		zap.String("bridge", bridgeName),
		zap.String("overlay_net", overlayNet.String()),
		zap.String("local_vtep", localVTEP))

	return overlay, nil
}

// createVXLANInfrastructure создает VXLAN инфраструктуру в системе
func (vm *VXLANOverlayManager) createVXLANInfrastructure(overlay *VXLANOverlay) error {
	
	// 1. Создаем VXLAN device
	cmd := exec.Command("ip", "link", "add", overlay.VXLANDevice, "type", "vxlan",
		"id", strconv.Itoa(int(overlay.VNI)),
		"local", overlay.LocalVTEP,
		"dstport", "4789")  // Стандартный VXLAN порт
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create VXLAN device: %w", err)
	}

	// 2. Создаем bridge для overlay сети
	cmd = exec.Command("ip", "link", "add", "name", overlay.BridgeName, "type", "bridge")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create bridge: %w", err)
	}

	// 3. Подключаем VXLAN device к bridge
	cmd = exec.Command("ip", "link", "set", overlay.VXLANDevice, "master", overlay.BridgeName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to attach VXLAN to bridge: %w", err)
	}

	// 4. Назначаем IP адрес bridge (gateway для overlay сети)
	bridgeIP := overlay.OverlayNet.IP
	bridgeIP[len(bridgeIP)-1] = 1 // .1 как gateway
	
	cmd = exec.Command("ip", "addr", "add",
		fmt.Sprintf("%s/%d", bridgeIP.String(), 24),
		"dev", overlay.BridgeName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to assign bridge IP: %w", err)
	}

	// 5. Поднимаем интерфейсы
	for _, device := range []string{overlay.VXLANDevice, overlay.BridgeName} {
		cmd = exec.Command("ip", "link", "set", "up", "dev", device)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to bring up %s: %w", device, err)
		}
	}

	// 6. Настраиваем forwarding и bridge параметры
	cmd = exec.Command("sysctl", "-w", "net.ipv4.ip_forward=1")
	if err := cmd.Run(); err != nil {
		vm.logger.Warn("Failed to enable IP forwarding", zap.Error(err))
	}

	// Отключаем STP на bridge для лучшей производительности
	cmd = exec.Command("ip", "link", "set", overlay.BridgeName, "type", "bridge", "stp_state", "0")
	if err := cmd.Run(); err != nil {
		vm.logger.Warn("Failed to disable STP", zap.Error(err))
	}

	return nil
}

// AddRemoteVTEP добавляет удаленный VTEP для tenant overlay
func (vm *VXLANOverlayManager) AddRemoteVTEP(tenantID, peerID, remoteVTEP string) error {
	vm.mutex.Lock()
	defer vm.mutex.Unlock()

	overlay, exists := vm.overlays[tenantID]
	if !exists {
		return fmt.Errorf("tenant overlay not found: %s", tenantID)
	}

	// Добавляем FDB (Forwarding Database) запись для multicast
	cmd := exec.Command("bridge", "fdb", "append", "00:00:00:00:00:00", 
		"dev", overlay.VXLANDevice, "dst", remoteVTEP)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add FDB entry: %w", err)
	}

	// Сохраняем peer информацию
	overlay.UnderlayPeers[peerID] = remoteVTEP
	overlay.RemoteVTEPs = append(overlay.RemoteVTEPs, remoteVTEP)

	vm.logger.Info("Added remote VTEP to overlay",
		zap.String("tenant_id", tenantID),
		zap.String("peer_id", peerID),
		zap.String("remote_vtep", remoteVTEP),
		zap.Int32("vni", overlay.VNI))

	return nil
}

// ConnectPeerToOverlay подключает peer к overlay сети
func (vm *VXLANOverlayManager) ConnectPeerToOverlay(tenantID, peerID string, peerIP net.IP) error {
	vm.mutex.RLock()
	overlay, exists := vm.overlays[tenantID]
	vm.mutex.RUnlock()
	
	if !exists {
		return fmt.Errorf("tenant overlay not found: %s", tenantID)
	}

	// Создаем veth pair для подключения peer'а
	vethPeer := fmt.Sprintf("veth-%s", peerID[:8])
	vethBridge := fmt.Sprintf("veth-br-%s", peerID[:8])

	// Создаем veth pair
	cmd := exec.Command("ip", "link", "add", vethPeer, "type", "veth", "peer", "name", vethBridge)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create veth pair: %w", err)
	}

	// Подключаем bridge-side к tenant bridge
	cmd = exec.Command("ip", "link", "set", vethBridge, "master", overlay.BridgeName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to connect to bridge: %w", err)
	}

	// Поднимаем интерфейсы
	for _, device := range []string{vethPeer, vethBridge} {
		cmd = exec.Command("ip", "link", "set", "up", "dev", device)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to bring up %s: %w", device, err)
		}
	}

	vm.logger.Info("Connected peer to overlay network",
		zap.String("tenant_id", tenantID),
		zap.String("peer_id", peerID),
		zap.String("peer_ip", peerIP.String()),
		zap.String("veth_peer", vethPeer))

	return nil
}

// getNextUnderlayIP возвращает следующий доступный IP в underlay сети
func (vm *VXLANOverlayManager) getNextUnderlayIP() string {
	// Простая реализация - в реальности нужна более сложная логика
	baseIP := vm.underlayNet.IP.To4()
	// Возвращаем следующий IP (нужно реализовать IPAM)
	nextIP := make(net.IP, 4)
	copy(nextIP, baseIP)
	nextIP[3] = byte(10 + len(vm.overlays)) // Простая логика
	return nextIP.String()
}

// GetTenantOverlay возвращает overlay для tenant'а
func (vm *VXLANOverlayManager) GetTenantOverlay(tenantID string) (*VXLANOverlay, bool) {
	vm.mutex.RLock()
	defer vm.mutex.RUnlock()
	
	overlay, exists := vm.overlays[tenantID]
	return overlay, exists
}

// ПРЕИМУЩЕСТВА VXLAN OVERLAY:
// Масштабируется до 16 миллионов VNI (2^24)
// Полная L2/L3 изоляция
// Работает поверх любой IP инфраструктуры  
// Стандартизированный протокол (RFC 7348)
// Поддержка multicast и unicast
// Совместимость с SDN контроллерами

// НЕДОСТАТКИ:
// Более сложная конфигурация
// Требует понимания сетевых технологий
// Дополнительный overhead (50 байт на пакет)
// Необходима overlay сеть администрирования
