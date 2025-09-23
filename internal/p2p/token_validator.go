package p2p

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
)

// TokenValidator представляет валидатор токенов для P2P
type TokenValidator struct {
	jwtSecret []byte
	logger    *zap.Logger
}

// TokenClaims представляет claims в JWT токене
type TokenClaims struct {
	PeerID   string `json:"peer_id"`
	TenantID string `json:"tenant_id"`
	Exp      int64  `json:"exp"`
	Iat      int64  `json:"iat"`
	Iss      string `json:"iss"`
	Sub      string `json:"sub"`
}

// ValidationResult представляет результат валидации токена
type ValidationResult struct {
	Valid    bool
	Claims   *TokenClaims
	Error    string
	PeerID   string
	TenantID string
}

// NewTokenValidator создает новый Token Validator
func NewTokenValidator(jwtSecret []byte, logger *zap.Logger) *TokenValidator {
	return &TokenValidator{
		jwtSecret: jwtSecret,
		logger:    logger,
	}
}

// ValidateToken валидирует JWT токен
func (tv *TokenValidator) ValidateToken(tokenString string) (*ValidationResult, error) {
	tv.logger.Debug("Validating P2P token", zap.Int("token_length", len(tokenString)))

	// Parse JWT token
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return &ValidationResult{
			Valid: false,
			Error: "invalid JWT format",
		}, fmt.Errorf("invalid JWT format")
	}

	// Decode header (basic validation)
	_, err := tv.decodeBase64(parts[0])
	if err != nil {
		return &ValidationResult{
			Valid: false,
			Error: "invalid header encoding",
		}, fmt.Errorf("invalid header encoding: %w", err)
	}

	// Decode payload
	payload, err := tv.decodeBase64(parts[1])
	if err != nil {
		return &ValidationResult{
			Valid: false,
			Error: "invalid payload encoding",
		}, fmt.Errorf("invalid payload encoding: %w", err)
	}

	// Verify signature
	signature, err := tv.decodeBase64URL(parts[2])
	if err != nil {
		return &ValidationResult{
			Valid: false,
			Error: "invalid signature encoding",
		}, fmt.Errorf("invalid signature encoding: %w", err)
	}

	// Check signature
	expectedSignature := tv.generateSignature(parts[0] + "." + parts[1])
	if !hmac.Equal(signature, expectedSignature) {
		return &ValidationResult{
			Valid: false,
			Error: "invalid signature",
		}, fmt.Errorf("invalid signature")
	}

	// Parse claims
	var claims TokenClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return &ValidationResult{
			Valid: false,
			Error: "invalid claims format",
		}, fmt.Errorf("invalid claims format: %w", err)
	}

	// Check expiration
	now := time.Now().Unix()
	if claims.Exp > 0 && now > claims.Exp {
		return &ValidationResult{
			Valid:    false,
			Error:    "token expired",
			Claims:   &claims,
			PeerID:   claims.PeerID,
			TenantID: claims.TenantID,
		}, fmt.Errorf("token expired")
	}

	// Check issued at time (not too far in the future)
	if claims.Iat > 0 && now < claims.Iat-300 { // Allow 5 minutes clock skew
		return &ValidationResult{
			Valid: false,
			Error: "token issued in the future",
		}, fmt.Errorf("token issued in the future")
	}

	tv.logger.Debug("Token validation successful",
		zap.String("peer_id", claims.PeerID),
		zap.String("tenant_id", claims.TenantID))

	return &ValidationResult{
		Valid:    true,
		Claims:   &claims,
		PeerID:   claims.PeerID,
		TenantID: claims.TenantID,
	}, nil
}

// ValidateP2PToken валидирует токен для P2P соединения
func (tv *TokenValidator) ValidateP2PToken(tokenString, expectedPeerID, expectedTenantID string) (*ValidationResult, error) {
	result, err := tv.ValidateToken(tokenString)
	if err != nil {
		return result, err
	}

	if !result.Valid {
		return result, fmt.Errorf("token validation failed: %s", result.Error)
	}

	// Проверяем соответствие peer ID
	if expectedPeerID != "" && result.Claims.PeerID != expectedPeerID {
		return &ValidationResult{
			Valid: false,
			Error: "peer ID mismatch",
		}, fmt.Errorf("peer ID mismatch: expected %s, got %s", expectedPeerID, result.Claims.PeerID)
	}

	// Проверяем соответствие tenant ID
	if expectedTenantID != "" && result.Claims.TenantID != expectedTenantID {
		return &ValidationResult{
			Valid: false,
			Error: "tenant ID mismatch",
		}, fmt.Errorf("tenant ID mismatch: expected %s, got %s", expectedTenantID, result.Claims.TenantID)
	}

	tv.logger.Info("P2P token validation successful",
		zap.String("peer_id", result.Claims.PeerID),
		zap.String("tenant_id", result.Claims.TenantID))

	return result, nil
}

// decodeBase64 декодирует base64 строку
func (tv *TokenValidator) decodeBase64(data string) ([]byte, error) {
	// Add padding if needed
	data = tv.addPadding(data)
	return base64.StdEncoding.DecodeString(data)
}

// decodeBase64URL декодирует base64url строку
func (tv *TokenValidator) decodeBase64URL(data string) ([]byte, error) {
	// Add padding if needed
	data = tv.addPadding(data)
	return base64.URLEncoding.DecodeString(data)
}

// addPadding добавляет padding для base64
func (tv *TokenValidator) addPadding(data string) string {
	switch len(data) % 4 {
	case 2:
		return data + "=="
	case 3:
		return data + "="
	default:
		return data
	}
}

// generateSignature генерирует HMAC подпись
func (tv *TokenValidator) generateSignature(data string) []byte {
	h := hmac.New(sha256.New, tv.jwtSecret)
	h.Write([]byte(data))
	return h.Sum(nil)
}

// CreateToken создает новый JWT токен для пира
func (tv *TokenValidator) CreateToken(peerID, tenantID string, duration time.Duration) (string, error) {
	now := time.Now()
	claims := TokenClaims{
		PeerID:   peerID,
		TenantID: tenantID,
		Exp:      now.Add(duration).Unix(),
		Iat:      now.Unix(),
		Iss:      "cloudbridge-relay",
		Sub:      peerID,
	}

	// Create header
	header := map[string]interface{}{
		"alg": "HS256",
		"typ": "JWT",
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("failed to marshal header: %w", err)
	}

	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("failed to marshal claims: %w", err)
	}

	// Encode header and payload
	headerEncoded := base64.URLEncoding.EncodeToString(headerJSON)
	claimsEncoded := base64.URLEncoding.EncodeToString(claimsJSON)

	// Create signature
	signingString := headerEncoded + "." + claimsEncoded
	signature := tv.generateSignature(signingString)
	signatureEncoded := base64.URLEncoding.EncodeToString(signature)

	token := signingString + "." + signatureEncoded

	tv.logger.Debug("Token created",
		zap.String("peer_id", peerID),
		zap.String("tenant_id", tenantID),
		zap.Duration("duration", duration))

	return token, nil
}
