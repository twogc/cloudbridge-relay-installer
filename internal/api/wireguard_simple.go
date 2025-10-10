package api

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/twogc/cloudbridge-relay-installer/internal/p2p"
	"github.com/twogc/cloudbridge-relay-installer/internal/wireguard"
	"go.uber.org/zap"
)

// WireGuardConfigResponse ответ с WireGuard конфигурацией
type WireGuardConfigResponse struct {
	Success         bool   `json:"success"`
	TenantCIDR      string `json:"tenant_cidr"`
	PeerIP          string `json:"peer_ip"`
	RelayEndpoint   string `json:"relay_endpoint"`
	WGConfig        string `json:"wg_config"`
	InterfaceName   string `json:"interface_name"`
	PrivateKey      string `json:"private_key"`
	PublicKey       string `json:"public_key"`
	AllowedIPs      string `json:"allowed_ips"`
	ServerPublicKey string `json:"server_public_key"`
}

// WireGuardSimpleEndpoints упрощенные WireGuard эндпоинты
type WireGuardSimpleEndpoints struct {
	peerManager        *PeerManager
	relayManager       *p2p.RelayManager
	relayManagerHybrid *p2p.RelayManagerHybrid
	wireguardManager   *wireguard.WireGuardTenantManager
	configManager      *wireguard.ConfigManager
	logger             *zap.Logger
}

// NewWireGuardSimpleEndpoints создает упрощенные WireGuard эндпоинты
func NewWireGuardSimpleEndpoints(
	peerManager *PeerManager,
	relayManager *p2p.RelayManager,
	relayManagerHybrid *p2p.RelayManagerHybrid,
	wireguardManager *wireguard.WireGuardTenantManager,
	configManager *wireguard.ConfigManager,
	logger *zap.Logger,
) *WireGuardSimpleEndpoints {
	return &WireGuardSimpleEndpoints{
		peerManager:        peerManager,
		relayManager:       relayManager,
		relayManagerHybrid: relayManagerHybrid,
		wireguardManager:   wireguardManager,
		configManager:      configManager,
		logger:             logger,
	}
}

// GetWireGuardConfig возвращает WireGuard конфигурацию для пира
func (wge *WireGuardSimpleEndpoints) GetWireGuardConfig(c *gin.Context) {
	peerID := c.Param("peer_id")
	tenantID := c.Param("tenant_id")

	// Валидация JWT токена
	claims, err := wge.validateToken(c)
	if err != nil {
		wge.logger.Error("Failed to validate token", zap.Error(err))
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Invalid token",
			Code:    http.StatusUnauthorized,
			Message: err.Error(),
		})
		return
	}

	// Проверка прав доступа
	if claims.TenantID != tenantID {
		c.JSON(http.StatusForbidden, ErrorResponse{
			Error:   "Tenant mismatch",
			Code:    http.StatusForbidden,
			Message: "Token tenant does not match request tenant",
		})
		return
	}

	// Получаем tenant CIDR от RelayManagerHybrid
	var tenantCIDR string
	if wge.relayManagerHybrid != nil {
		tenantCIDR = wge.relayManagerHybrid.GetTenantNetwork(tenantID)
	}

	// Если RelayManagerHybrid не доступен, используем заглушку
	if tenantCIDR == "" {
		tenantCIDR = "10.100.77.0/24" // Заглушка для тестирования
		wge.logger.Warn("Using fallback tenant CIDR",
			zap.String("tenant_id", tenantID),
			zap.String("fallback_cidr", tenantCIDR))
	}

	// Выделяем IP для пира
	var peerIP string
	if wge.relayManagerHybrid != nil {
		ip, err := wge.relayManagerHybrid.AllocatePeerIP(tenantID, peerID)
		if err != nil && !strings.Contains(err.Error(), "already") {
			// Попробуем получить существующий IP
			if wge.relayManagerHybrid.GetPeerIPAMRepo() != nil {
				if existing, e2 := wge.relayManagerHybrid.GetPeerIPAMRepo().GetPeerIP(peerID); e2 == nil {
					peerIP = existing
				} else {
					wge.logger.Error("Failed to allocate peer IP",
						zap.String("tenant_id", tenantID),
						zap.String("peer_id", peerID),
						zap.Error(err))
					c.JSON(http.StatusInternalServerError, ErrorResponse{
						Error:   "Failed to allocate peer IP",
						Code:    http.StatusInternalServerError,
						Message: err.Error(),
					})
					return
				}
			} else {
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Error:   "Failed to allocate peer IP",
					Code:    http.StatusInternalServerError,
					Message: err.Error(),
				})
				return
			}
		} else {
			peerIP = ip
		}
	} else {
		// Заглушка для тестирования
		peerIP = "10.100.77.10"
		wge.logger.Warn("Using fallback peer IP",
			zap.String("tenant_id", tenantID),
			zap.String("peer_id", peerID),
			zap.String("fallback_ip", peerIP))
	}

	// Получаем или создаем WireGuard туннель для тенанта
	var wgTenant *wireguard.WireGuardTenant

	if wge.wireguardManager != nil {
		var err error
		wgTenant, err = wge.wireguardManager.CreateTenantTunnel(tenantID)
		if err != nil {
			wge.logger.Error("Failed to create WireGuard tunnel for tenant",
				zap.String("tenant_id", tenantID),
				zap.Error(err))
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to create WireGuard tunnel",
				Code:    http.StatusInternalServerError,
				Message: err.Error(),
			})
			return
		}
	} else {
		// Создаем заглушку для тестирования
		wgTenant = &wireguard.WireGuardTenant{
			TenantID:            tenantID,
			PublicKey:           "WG_PUBLIC_KEY_PLACEHOLDER",
			PrivateKey:          "WG_PRIVATE_KEY_PLACEHOLDER",
			AllowedIPs:          []string{tenantCIDR},
			Endpoint:            "relay.example.com:51820",
			Port:                51820,
			PersistentKeepalive: 25,
		}
		wge.logger.Warn("Using fallback WireGuard tenant",
			zap.String("tenant_id", tenantID))
	}

	// Генерируем WireGuard конфигурацию
	wgConfig := wge.generatePeerWireGuardConfig(wgTenant, peerID, tenantCIDR, peerIP)

	wge.logger.Info("Generated WireGuard config for peer",
		zap.String("peer_id", peerID),
		zap.String("tenant_id", tenantID),
		zap.String("tenant_cidr", tenantCIDR),
		zap.String("peer_ip", peerIP))

	c.JSON(http.StatusOK, WireGuardConfigResponse{
		Success:       true,
		TenantCIDR:    tenantCIDR,
		PeerIP:        peerIP,
		RelayEndpoint: wgTenant.Endpoint,
		WGConfig:      wgConfig,
		InterfaceName: "wg-" + tenantID,
		PrivateKey:    wgTenant.PrivateKey,
		PublicKey:     wgTenant.PublicKey,
		AllowedIPs:    tenantCIDR, // Только tenant подсеть
	})
}

// generatePeerWireGuardConfig генерирует WireGuard конфигурацию для пира
func (wge *WireGuardSimpleEndpoints) generatePeerWireGuardConfig(wgTenant *wireguard.WireGuardTenant, peerID, tenantCIDR, peerIP string) string {

	// Создаем конфигурацию с реальными IP адресами
	config := fmt.Sprintf(`[Interface]
# CloudBridge P2P WireGuard Configuration
# Tenant: %s
# Peer ID: %s
# Peer IP: %s
# Tenant CIDR: %s

PrivateKey = %s
Address = %s/24
DNS = 8.8.8.8, 1.1.1.1

# PostUp and PostDown scripts for routing
PostUp = ip route add %s dev %%i
PostDown = ip route del %s dev %%i

[Peer]
# CloudBridge Relay Server
PublicKey = %s
Endpoint = %s:%d
AllowedIPs = %s
PersistentKeepalive = %d

# Connection settings
MTU = 1420
`,
		wge.getTenantIDFromContext(),
		peerID,
		peerIP,
		tenantCIDR,
		wgTenant.PrivateKey,
		peerIP,
		tenantCIDR,
		tenantCIDR,
		wge.getRelayPublicKey(),
		wgTenant.Endpoint,
		wgTenant.Port,
		tenantCIDR, // Только tenant подсеть для безопасности
		wgTenant.PersistentKeepalive)

	return config
}

// generateWGConfig создает минимальный WireGuard конфиг
func generateWGConfig(peerIP, tenantCIDR, relayPubKey, relayEndpoint string) string {
	return fmt.Sprintf(`[Interface]
# Сгенерируйте приватный ключ на клиенте: wg genkey
PrivateKey = <PUT_YOUR_PRIVATE_KEY_HERE>
Address = %s/24
DNS = 1.1.1.1

[Peer]
PublicKey = %s
Endpoint = %s
AllowedIPs = %s
PersistentKeepalive = 25
`, peerIP, relayPubKey, relayEndpoint, tenantCIDR)
}

// getPeerIP возвращает IP адрес пира (заглушка)
func (wge *WireGuardSimpleEndpoints) getPeerIP(peerID, tenantCIDR string) string {
	// В реальной реализации это должно получаться из IPAM
	// Пока используем простую логику
	return "10.100.77.10" // Заглушка
}

// getTenantIDFromContext извлекает tenant ID из контекста
func (wge *WireGuardSimpleEndpoints) getTenantIDFromContext() string {
	// В реальной реализации это должно извлекаться из контекста запроса
	return "tenant-from-context"
}

// getRelayPublicKey возвращает публичный ключ relay сервера
func (wge *WireGuardSimpleEndpoints) getRelayPublicKey() string {
	// В реальной реализации это должно получаться из конфигурации relay сервера
	return "RELAY_PUBLIC_KEY_PLACEHOLDER"
}

// validateToken валидирует JWT токен
func (wge *WireGuardSimpleEndpoints) validateToken(c *gin.Context) (*P2PClaims, error) {
	// Проверяем, включена ли авторизация
	if os.Getenv("AUTH_ENABLED") == "false" {
		// Извлекаем server_id из токена даже при отключенной авторизации
		authHeader := c.GetHeader("Authorization")
		serverID := "anonymous"

		if authHeader != "" {
			token := strings.TrimPrefix(authHeader, "Bearer ")
			if token != authHeader {
				// Извлекаем server_id из токена формата "server-id:server-a"
				if strings.Contains(token, ":") {
					parts := strings.Split(token, ":")
					if len(parts) >= 2 {
						serverID = parts[1]
					}
				}
			}
		}

		tenantID := c.Param("tenant_id")
		if tenantID == "" {
			tenantID = "default"
		}

		return &P2PClaims{
			Subject:        "peer_" + serverID,
			TenantID:       tenantID,
			ServerID:       serverID,
			ConnectionType: "quic",
			ProtocolType:   "p2p-mesh",
			Permissions:    []string{"relay_connect", "relay_tunnel"},
		}, nil
	}

	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return nil, fmt.Errorf("missing authorization header")
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == authHeader {
		return nil, fmt.Errorf("invalid authorization format")
	}

	// Используем переданный валидатор (для тестов) или локальную JWT проверку
	if wge.peerManager.validator != nil {
		claims, err := wge.peerManager.validator.ValidateP2PToken(token)
		if err != nil {
			return nil, fmt.Errorf("invalid token: %v", err)
		}
		return claims, nil
	}

	// Fallback к локальной HS256/JWT проверке с fallback секретом
	claims, err := wge.validateP2PToken(token)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %v", err)
	}

	return claims, nil
}

// validateP2PToken валидирует P2P JWT токен
func (wge *WireGuardSimpleEndpoints) validateP2PToken(tokenString string) (*P2PClaims, error) {
	// Удаление "Bearer " префикса
	if len(tokenString) > 7 && strings.HasPrefix(tokenString, "Bearer ") {
		tokenString = tokenString[7:]
	}

	// Определяем секрет: сначала KEYCLOAK_FALLBACK_SECRET, затем JWT_SECRET
	secret := os.Getenv("KEYCLOAK_FALLBACK_SECRET")
	if secret == "" {
		secret = os.Getenv("JWT_SECRET")
	}
	if secret == "" {
		secret = "test-secret"
	}

	// Парсинг токена с HS256 секретом (raw, без base64 декодирования)
	token, err := jwt.ParseWithClaims(tokenString, &P2PClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*P2PClaims)
	if !ok || !token.Valid {
		return nil, err
	}

	// Проверка обязательных полей для P2P mesh
	if claims.ProtocolType != "p2p-mesh" {
		return nil, fmt.Errorf("unsupported protocol: %s", claims.ProtocolType)
	}

	if claims.ConnectionType != "quic" {
		return nil, fmt.Errorf("unsupported connection type: %s", claims.ConnectionType)
	}

	if strings.TrimSpace(claims.TenantID) == "" {
		return nil, fmt.Errorf("missing tenant_id")
	}

	if strings.TrimSpace(claims.ServerID) == "" {
		return nil, fmt.Errorf("missing server_id")
	}

	return claims, nil
}
