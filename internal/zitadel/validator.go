package zitadel

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// ZitadelValidator валидатор токенов Zitadel
type ZitadelValidator struct {
	domain         string
	projectID      string
	clientID       string
	clientSecret   string
	fallbackSecret string
	logger         *zap.Logger
}

// NewZitadelValidator создает валидатор Zitadel
func NewZitadelValidator(config *Config, logger *zap.Logger) (*ZitadelValidator, error) {
	return &ZitadelValidator{
		domain:         config.Domain,
		projectID:      config.ProjectID,
		clientID:       config.ClientID,
		clientSecret:   config.ClientSecret,
		fallbackSecret: config.FallbackSecret,
		logger:         logger,
	}, nil
}

// ValidateToken валидирует JWT токен (начинаем с fallback)
func (zv *ZitadelValidator) ValidateToken(tokenString string) (*BaseTokenClaims, error) {
	// Удаление Bearer префикса если есть
	if strings.HasPrefix(tokenString, "Bearer ") {
		tokenString = strings.TrimPrefix(tokenString, "Bearer ")
	}

	// Пока используем fallback валидацию
	// TODO: Добавить полноценную интеграцию с Zitadel API
	return zv.validateWithFallbackSecret(tokenString)
}

// ValidateP2PMeshToken валидирует токен для P2P Mesh протокола
func (zv *ZitadelValidator) ValidateP2PMeshToken(tokenString string) (*P2PMeshTokenClaims, error) {
	baseClaims, err := zv.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	// Проверка protocol_type
	if baseClaims.ProtocolType != "p2p-mesh" {
		return nil, fmt.Errorf("invalid protocol type for P2P mesh: %s", baseClaims.ProtocolType)
	}

	// Проверка scope
	if !strings.Contains(baseClaims.Scope, "p2p-mesh-claims") {
		return nil, fmt.Errorf("missing p2p-mesh-claims scope")
	}

	// Проверка обязательных claims
	if baseClaims.TenantID == "" {
		return nil, fmt.Errorf("missing tenant_id")
	}

	if baseClaims.ServerID == "" {
		return nil, fmt.Errorf("missing server_id")
	}

	p2pClaims := &P2PMeshTokenClaims{
		BaseTokenClaims: *baseClaims,
		MaxPeers:        "10", // default
	}

	zv.logger.Info("P2P Mesh token validated successfully",
		zap.String("tenant_id", p2pClaims.TenantID),
		zap.String("server_id", p2pClaims.ServerID),
		zap.String("max_peers", p2pClaims.MaxPeers),
	)

	return p2pClaims, nil
}

// ValidateQuickToken валидирует токен для Quick протокола
func (zv *ZitadelValidator) ValidateQuickToken(tokenString string) (*QuickTokenClaims, error) {
	baseClaims, err := zv.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	// Проверка protocol_type
	if baseClaims.ProtocolType != "quick" {
		return nil, fmt.Errorf("invalid protocol type for Quick: %s", baseClaims.ProtocolType)
	}

	quickClaims := &QuickTokenClaims{
		BaseTokenClaims: *baseClaims,
		MaxConnections:  "10",    // default
		MaxTunnels:      "5",     // default
		MaxBandwidth:    "100MB", // default
	}

	return quickClaims, nil
}

// validateWithFallbackSecret валидация через fallback секрет
func (zv *ZitadelValidator) validateWithFallbackSecret(tokenString string) (*BaseTokenClaims, error) {
	if zv.fallbackSecret == "" {
		return nil, fmt.Errorf("fallback secret not configured")
	}

	// Парсинг JWT токена с fallback секретом
	token, err := jwt.ParseWithClaims(tokenString, &jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(zv.fallbackSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(*jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims format")
	}

	// Извлечение базовых claims
	baseClaims := &BaseTokenClaims{}

	if iss, ok := (*claims)["iss"].(string); ok {
		baseClaims.Issuer = iss
	}

	if sub, ok := (*claims)["sub"].(string); ok {
		baseClaims.Subject = sub
	}

	if exp, ok := (*claims)["exp"].(float64); ok {
		baseClaims.ExpiresAt = int64(exp)
	}

	if iat, ok := (*claims)["iat"].(float64); ok {
		baseClaims.IssuedAt = int64(iat)
	}

	if protocolType, ok := (*claims)["protocol_type"].(string); ok {
		baseClaims.ProtocolType = protocolType
	}

	if tenantID, ok := (*claims)["tenant_id"].(string); ok {
		baseClaims.TenantID = tenantID
	}

	if serverID, ok := (*claims)["server_id"].(string); ok {
		baseClaims.ServerID = serverID
	}

	if scope, ok := (*claims)["scope"].(string); ok {
		baseClaims.Scope = scope
	}

	// Проверка времени жизни токена
	now := time.Now().Unix()
	if baseClaims.ExpiresAt > 0 && now > baseClaims.ExpiresAt {
		return nil, fmt.Errorf("token expired")
	}

	zv.logger.Debug("Token validated successfully",
		zap.String("tenant_id", baseClaims.TenantID),
		zap.String("server_id", baseClaims.ServerID),
		zap.String("protocol_type", baseClaims.ProtocolType),
		zap.String("issuer", baseClaims.Issuer),
	)

	return baseClaims, nil
}

// GetTenantInfo получает информацию о тенанте (заглушка)
func (zv *ZitadelValidator) GetTenantInfo(ctx context.Context, tenantID string) (map[string]interface{}, error) {
	// TODO: Реализовать получение информации о тенанте через Zitadel API
	return map[string]interface{}{
		"id":     tenantID,
		"name":   fmt.Sprintf("Tenant %s", tenantID),
		"domain": fmt.Sprintf("%s.%s", tenantID, zv.domain),
		"state":  "active",
	}, nil
}

// CreateTenant создает новый тенант (заглушка)
func (zv *ZitadelValidator) CreateTenant(ctx context.Context, name, domain string) (string, error) {
	// TODO: Реализовать создание тенанта через Zitadel API
	tenantID := fmt.Sprintf("tenant-%d", time.Now().Unix())

	zv.logger.Info("Would create tenant in Zitadel",
		zap.String("tenant_id", tenantID),
		zap.String("name", name),
		zap.String("domain", domain),
	)

	return tenantID, nil
}

// Close закрывает соединение
func (zv *ZitadelValidator) Close() error {
	return nil
}
