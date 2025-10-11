package zitadel

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

// ZitadelValidator представляет валидатор токенов Zitadel
type ZitadelValidator struct {
	client *Client
	logger *zap.Logger
}

// NewZitadelValidator создает валидатор Zitadel
func NewZitadelValidator(client *Client, logger *zap.Logger) *ZitadelValidator {
	return &ZitadelValidator{
		client: client,
		logger: logger,
	}
}

// ValidateToken валидирует JWT токен через Zitadel клиент
func (zv *ZitadelValidator) ValidateToken(ctx context.Context, tokenString string) (*TokenClaims, error) {
	if zv.client == nil {
		return nil, fmt.Errorf("zitadel client not initialized")
	}

	// Используем новый метод валидации из client.go
	claims, err := zv.client.ValidateToken(ctx, tokenString)
	if err != nil {
		zv.logger.Debug("Token validation failed",
			zap.Error(err),
			zap.String("token_prefix", tokenString[:min(len(tokenString), 20)]))
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	zv.logger.Debug("Token validated successfully",
		zap.String("sub", claims.Subject),
		zap.String("org_id", claims.OrgID))

	return claims, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
