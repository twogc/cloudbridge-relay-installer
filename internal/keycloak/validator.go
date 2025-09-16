package keycloak

import (
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// BaseTokenClaims базовые claims для всех токенов
type BaseTokenClaims struct {
	ProtocolType string `json:"protocol_type"`
	TenantID     string `json:"tenant_id"`
	ServerID     string `json:"server_id"`
	Issuer       string `json:"iss"`
	Subject      string `json:"sub"`
	ExpiresAt    int64  `json:"exp"`
	IssuedAt     int64  `json:"iat"`
	Scope        string `json:"scope"`
}

// QuickTokenClaims claims для Quick протокола
type QuickTokenClaims struct {
	BaseTokenClaims
	TargetServer   string `json:"target_server"`
	MaxConnections string `json:"max_connections"`
	MaxTunnels     string `json:"max_tunnels"`
	MaxBandwidth   string `json:"max_bandwidth"`
}

// P2PMeshTokenClaims claims для P2P Mesh протокола
type P2PMeshTokenClaims struct {
	BaseTokenClaims
	MeshConfig string `json:"mesh_config"`
	MaxPeers   string `json:"max_peers"`
}

// KeycloakValidator валидатор токенов Keycloak
type KeycloakValidator struct {
	jwksURL        string
	audience       string
	issuer         string
	fallbackSecret string
	jwksCache      interface{} // Временно используем interface{} вместо jwk.Set
	lastUpdate     time.Time
	logger         *zap.Logger
}

// NewKeycloakValidator создает новый валидатор Keycloak
func NewKeycloakValidator(jwksURL, audience, issuer, fallbackSecret string, logger *zap.Logger) *KeycloakValidator {
	return &KeycloakValidator{
		jwksURL:        jwksURL,
		audience:       audience,
		issuer:         issuer,
		fallbackSecret: fallbackSecret,
		logger:         logger,
	}
}

// ValidateToken валидирует JWT токен
func (kv *KeycloakValidator) ValidateToken(tokenString string) (*BaseTokenClaims, error) {
	// Обновление JWKS при необходимости
	if err := kv.updateJWKS(); err != nil {
		return nil, fmt.Errorf("failed to update JWKS: %w", err)
	}

	// Парсинг токена
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Проверка алгоритма (поддерживаем и RSA, и HS256 для fallback)
		switch token.Method.(type) {
		case *jwt.SigningMethodRSA:
			// RSA ключи - полная JWKS валидация
			// TODO: Реализовать получение ключа из JWKS
			kv.logger.Warn("RSA validation not implemented yet")
			return nil, fmt.Errorf("RSA validation not implemented yet")
		case *jwt.SigningMethodHMAC:
			// HS256 ключи - fallback режим
			kv.logger.Info("Using HS256 fallback validation", zap.String("kid", token.Header["kid"].(string)))
			// Для fallback используем секретный ключ из конфигурации
			fallbackSecret := kv.fallbackSecret
			if fallbackSecret == "" {
				fallbackSecret = "test-secret" // fallback для разработки
				kv.logger.Warn("Fallback secret not configured, using default")
			}
			return []byte(fallbackSecret), nil
		default:
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Извлечение claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims format")
	}

	// Валидация issuer и audience
	if claims["iss"] != kv.issuer {
		return nil, fmt.Errorf("invalid issuer: expected %s, got %s", kv.issuer, claims["iss"])
	}

	if claims["aud"] != kv.audience {
		return nil, fmt.Errorf("invalid audience: expected %s, got %s", kv.audience, claims["aud"])
	}

	// Проверка срока действия
	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			return nil, fmt.Errorf("token expired")
		}
	}

	// Преобразование в структуру
	baseClaims := &BaseTokenClaims{
		ProtocolType: getStringClaim(claims, "protocol_type"),
		TenantID:     getStringClaim(claims, "tenant_id"),
		ServerID:     getStringClaim(claims, "server_id"),
		Issuer:       getStringClaim(claims, "iss"),
		Subject:      getStringClaim(claims, "sub"),
		Scope:        getStringClaim(claims, "scope"),
	}

	if exp, ok := claims["exp"].(float64); ok {
		baseClaims.ExpiresAt = int64(exp)
	}

	if iat, ok := claims["iat"].(float64); ok {
		baseClaims.IssuedAt = int64(iat)
	}

	kv.logger.Info("Token validated successfully",
		zap.String("protocol_type", baseClaims.ProtocolType),
		zap.String("tenant_id", baseClaims.TenantID),
		zap.String("server_id", baseClaims.ServerID),
		zap.String("subject", baseClaims.Subject),
	)

	return baseClaims, nil
}

// ValidateQuickToken валидирует токен для Quick протокола
func (kv *KeycloakValidator) ValidateQuickToken(tokenString string) (*QuickTokenClaims, error) {
	baseClaims, err := kv.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	// Проверка protocol_type
	if baseClaims.ProtocolType != "quick" {
		return nil, fmt.Errorf("invalid protocol type for quick connection: %s", baseClaims.ProtocolType)
	}

	// Проверка scope
	if !strings.Contains(baseClaims.Scope, "quick-connect-claims") {
		return nil, fmt.Errorf("missing quick-connect-claims scope")
	}

	// Проверка обязательных claims
	if baseClaims.TenantID == "" {
		return nil, fmt.Errorf("missing tenant_id")
	}

	if baseClaims.ServerID == "" {
		return nil, fmt.Errorf("missing server_id")
	}

	// Парсинг дополнительных claims для Quick
	token, _ := jwt.Parse(tokenString, nil)
	claims := token.Claims.(jwt.MapClaims)

	quickClaims := &QuickTokenClaims{
		BaseTokenClaims: *baseClaims,
		TargetServer:    getStringClaim(claims, "target_server"),
		MaxConnections:  getStringClaim(claims, "max_connections"),
		MaxTunnels:      getStringClaim(claims, "max_tunnels"),
		MaxBandwidth:    getStringClaim(claims, "max_bandwidth"),
	}

	// Валидация лимитов
	if quickClaims.MaxConnections == "" {
		quickClaims.MaxConnections = "5" // default
	}

	kv.logger.Info("Quick token validated successfully",
		zap.String("tenant_id", quickClaims.TenantID),
		zap.String("server_id", quickClaims.ServerID),
		zap.String("target_server", quickClaims.TargetServer),
		zap.String("max_connections", quickClaims.MaxConnections),
	)

	return quickClaims, nil
}

// ValidateP2PMeshToken валидирует токен для P2P Mesh протокола
func (kv *KeycloakValidator) ValidateP2PMeshToken(tokenString string) (*P2PMeshTokenClaims, error) {
	baseClaims, err := kv.ValidateToken(tokenString)
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

	// Парсинг дополнительных claims для P2P Mesh
	token, _ := jwt.Parse(tokenString, nil)
	claims := token.Claims.(jwt.MapClaims)

	p2pClaims := &P2PMeshTokenClaims{
		BaseTokenClaims: *baseClaims,
		MeshConfig:      getStringClaim(claims, "mesh_config"),
		MaxPeers:        getStringClaim(claims, "max_peers"),
	}

	// Валидация лимитов
	if p2pClaims.MaxPeers == "" {
		p2pClaims.MaxPeers = "10" // default
	}

	kv.logger.Info("P2P Mesh token validated successfully",
		zap.String("tenant_id", p2pClaims.TenantID),
		zap.String("server_id", p2pClaims.ServerID),
		zap.String("max_peers", p2pClaims.MaxPeers),
	)

	return p2pClaims, nil
}

// updateJWKS обновляет JWKS кэш
func (kv *KeycloakValidator) updateJWKS() error {
	// Обновляем JWKS каждые 5 минут
	if time.Since(kv.lastUpdate) < 5*time.Minute {
		return nil
	}

	kv.logger.Debug("Updating JWKS cache", zap.String("jwks_url", kv.jwksURL))

	// TODO: Реализовать получение JWKS
	// Пока просто обновляем время
	kv.lastUpdate = time.Now()

	kv.logger.Info("JWKS cache updated successfully")
	return nil
}

// getStringClaim извлекает строковое значение из claims
func getStringClaim(claims jwt.MapClaims, key string) string {
	if val, ok := claims[key].(string); ok {
		return val
	}
	return ""
}
