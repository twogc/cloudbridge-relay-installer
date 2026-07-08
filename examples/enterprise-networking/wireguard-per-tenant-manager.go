// CloudBridge Enterprise Networking - WireGuard Per-Tenant Manager
// ================================================================

package networking

import (
	"fmt"
	"net"
	"os/exec"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// WireGuardTenantManager управляет отдельными WG интерфейсами для каждого tenant'а
type WireGuardTenantManager struct {
	interfaces    map[string]*WGTenantInterface  // tenantID -> interface
	portCounter   int32                          // Счетчик портов начиная с 51821
	mutex         sync.RWMutex
	logger        *zap.Logger
}

// WGTenantInterface представляет WireGuard интерфейс для tenant'а
type WGTenantInterface struct {
	TenantID      string
	InterfaceName string    // wg-tenant-001, wg-tenant-002
	ListenPort    int       // 51821, 51822, 51823...
	PrivateKey    string
	PublicKey     string
	Network       *net.IPNet // 10.100.77.0/24, 10.100.78.0/24
	Peers         map[string]*WGPeer
	CreatedAt     time.Time
}

// WGPeer представляет peer в рамках tenant'а
type WGPeer struct {
	PeerID      string
	PublicKey   string
	AllowedIPs  []string
	Endpoint    string
	LastSeen    time.Time
}

// NewWireGuardTenantManager создает новый менеджер
func NewWireGuardTenantManager(logger *zap.Logger) *WireGuardTenantManager {
	return &WireGuardTenantManager{
		interfaces:  make(map[string]*WGTenantInterface),
		portCounter: 51820, // Начинаем с 51821
		logger:      logger,
	}
}

// CreateTenantInterface создает изолированный WG интерфейс для tenant'а
func (wm *WireGuardTenantManager) CreateTenantInterface(tenantID string, network *net.IPNet) (*WGTenantInterface, error) {
	wm.mutex.Lock()
	defer wm.mutex.Unlock()

	// Проверяем что интерфейс не существует
	if existing, exists := wm.interfaces[tenantID]; exists {
		return existing, nil
	}

	// Генерируем ключи
	privateKey, publicKey, err := wm.generateKeyPair()
	if err != nil {
		return nil, fmt.Errorf("failed to generate keys: %w", err)
	}

	// Выделяем порт
	wm.portCounter++
	port := wm.portCounter

	// Создаем интерфейс
	interfaceName := fmt.Sprintf("wg-tenant-%s", tenantID[len(tenantID)-3:]) // Последние 3 символа
	
	wgInterface := &WGTenantInterface{
		TenantID:      tenantID,
		InterfaceName: interfaceName,
		ListenPort:    int(port),
		PrivateKey:    privateKey,
		PublicKey:     publicKey,
		Network:       network,
		Peers:         make(map[string]*WGPeer),
		CreatedAt:     time.Now(),
	}

	// Создаем системный интерфейс
	err = wm.createSystemInterface(wgInterface)
	if err != nil {
		return nil, fmt.Errorf("failed to create system interface: %w", err)
	}

	wm.interfaces[tenantID] = wgInterface

	wm.logger.Info("Created tenant WireGuard interface",
		zap.String("tenant_id", tenantID),
		zap.String("interface", interfaceName),
		zap.Int("port", int(port)),
		zap.String("network", network.String()))

	return wgInterface, nil
}

// createSystemInterface создает реальный WG интерфейс в системе
func (wm *WireGuardTenantManager) createSystemInterface(wgInterface *WGTenantInterface) error {
	// 1. Создаем интерфейс
	cmd := exec.Command("ip", "link", "add", "dev", wgInterface.InterfaceName, "type", "wireguard")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create interface: %w", err)
	}

	// 2. Назначаем IP адрес (например, .1 в подсети)
	interfaceIP := wgInterface.Network.IP
	interfaceIP[len(interfaceIP)-1] = 1 // .1
	
	cmd = exec.Command("ip", "addr", "add", 
		fmt.Sprintf("%s/%d", interfaceIP.String(), 24), 
		"dev", wgInterface.InterfaceName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to assign IP: %w", err)
	}

	// 3. Конфигурируем WireGuard
	cmd = exec.Command("wg", "set", wgInterface.InterfaceName,
		"private-key", "/dev/stdin",
		"listen-port", fmt.Sprintf("%d", wgInterface.ListenPort))
	cmd.Stdin = strings.NewReader(wgInterface.PrivateKey)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to configure wireguard: %w", err)
	}

	// 4. Поднимаем интерфейс
	cmd = exec.Command("ip", "link", "set", "up", "dev", wgInterface.InterfaceName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to bring up interface: %w", err)
	}

	return nil
}

// AddPeer добавляет peer в tenant интерфейс
func (wm *WireGuardTenantManager) AddPeer(tenantID, peerID, publicKey string, allowedIPs []string) error {
	wm.mutex.Lock()
	defer wm.mutex.Unlock()

	wgInterface, exists := wm.interfaces[tenantID]
	if !exists {
		return fmt.Errorf("tenant interface not found: %s", tenantID)
	}

	// Добавляем peer в WireGuard
	cmd := exec.Command("wg", "set", wgInterface.InterfaceName,
		"peer", publicKey,
		"allowed-ips", strings.Join(allowedIPs, ","))
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add peer: %w", err)
	}

	// Сохраняем информацию о peer'е
	wgInterface.Peers[peerID] = &WGPeer{
		PeerID:     peerID,
		PublicKey:  publicKey,
		AllowedIPs: allowedIPs,
		LastSeen:   time.Now(),
	}

	wm.logger.Info("Added peer to tenant interface",
		zap.String("tenant_id", tenantID),
		zap.String("peer_id", peerID),
		zap.String("interface", wgInterface.InterfaceName))

	return nil
}

// generateKeyPair генерирует пару ключей WireGuard
func (wm *WireGuardTenantManager) generateKeyPair() (privateKey, publicKey string, err error) {
	// Генерируем private key
	cmd := exec.Command("wg", "genkey")
	privateKeyBytes, err := cmd.Output()
	if err != nil {
		return "", "", err
	}
	privateKey = strings.TrimSpace(string(privateKeyBytes))

	// Генерируем public key
	cmd = exec.Command("wg", "pubkey")
	cmd.Stdin = strings.NewReader(privateKey)
	publicKeyBytes, err := cmd.Output()
	if err != nil {
		return "", "", err
	}
	publicKey = strings.TrimSpace(string(publicKeyBytes))

	return privateKey, publicKey, nil
}

// GetTenantInterface возвращает интерфейс для tenant'а
func (wm *WireGuardTenantManager) GetTenantInterface(tenantID string) (*WGTenantInterface, bool) {
	wm.mutex.RLock()
	defer wm.mutex.RUnlock()
	
	wgInterface, exists := wm.interfaces[tenantID]
	return wgInterface, exists
}

// ПРЕИМУЩЕСТВА ЭТОГО ПОДХОДА:
// Полная изоляция на сетевом уровне
// Независимые ключи для каждого tenant'а  
// Отдельные порты и интерфейсы
// Простая операционная модель
// Высокая безопасность

// НЕДОСТАТКИ:
// Увеличение количества сетевых интерфейсов
// Больше портов для открытия в firewall
// Сложность управления большим количеством интерфейсов
