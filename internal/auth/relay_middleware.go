package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

// RelayMiddleware предоставляет middleware для аутентификации relay серверов
type RelayMiddleware struct {
	validator *ZitadelValidator
	logger    *zap.Logger
}

// NewRelayMiddleware создает новый middleware для relay аутентификации
func NewRelayMiddleware(validator *ZitadelValidator, logger *zap.Logger) *RelayMiddleware {
	return &RelayMiddleware{
		validator: validator,
		logger:    logger,
	}
}

// RequireRelayAuth возвращает middleware, который требует аутентификации relay сервера
func (m *RelayMiddleware) RequireRelayAuth(requiredPermissions []string, requiredRoles []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Извлекаем токен из заголовка Authorization
			token, err := m.extractToken(r)
			if err != nil {
				m.logger.Error("Failed to extract token", zap.Error(err))
				http.Error(w, "Unauthorized: Invalid token", http.StatusUnauthorized)
				return
			}

			// Валидируем токен
			claims, err := m.validator.ValidateRelayToken(r.Context(), token)
			if err != nil {
				m.logger.Error("Token validation failed", zap.Error(err))
				http.Error(w, "Unauthorized: Invalid token", http.StatusUnauthorized)
				return
			}

			// Проверяем разрешения
			if len(requiredPermissions) > 0 {
				if err := m.validator.ValidatePermissions(claims, requiredPermissions); err != nil {
					m.logger.Error("Permission validation failed", zap.Error(err))
					http.Error(w, "Forbidden: Insufficient permissions", http.StatusForbidden)
					return
				}
			}

			// Проверяем роли
			if len(requiredRoles) > 0 {
				if err := m.validator.ValidateRoles(claims, requiredRoles); err != nil {
					m.logger.Error("Role validation failed", zap.Error(err))
					http.Error(w, "Forbidden: Insufficient roles", http.StatusForbidden)
					return
				}
			}

			// Добавляем claims в контекст
			ctx := context.WithValue(r.Context(), "relay_claims", claims)
			ctx = context.WithValue(ctx, "relay_server_id", claims.GetRelayServerID())
			ctx = context.WithValue(ctx, "tenant_id", claims.GetTenantID())
			ctx = context.WithValue(ctx, "org_id", claims.GetOrgID())

			m.logger.Info("Relay authentication successful",
				zap.String("relay_server_id", claims.GetRelayServerID()),
				zap.String("tenant_id", claims.GetTenantID()),
				zap.String("org_id", claims.GetOrgID()),
				zap.String("remote_addr", r.RemoteAddr))

			// Передаем управление следующему обработчику
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRelayPermissions возвращает middleware, который требует определенные разрешения
func (m *RelayMiddleware) RequireRelayPermissions(permissions []string) func(http.Handler) http.Handler {
	return m.RequireRelayAuth(permissions, nil)
}

// RequireRelayRoles возвращает middleware, который требует определенные роли
func (m *RelayMiddleware) RequireRelayRoles(roles []string) func(http.Handler) http.Handler {
	return m.RequireRelayAuth(nil, roles)
}

// RequireRelayAuthOnly возвращает middleware, который требует только валидный relay токен
func (m *RelayMiddleware) RequireRelayAuthOnly() func(http.Handler) http.Handler {
	return m.RequireRelayAuth(nil, nil)
}

// extractToken извлекает JWT токен из заголовка Authorization
func (m *RelayMiddleware) extractToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("missing Authorization header")
	}

	// Проверяем формат "Bearer <token>"
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", fmt.Errorf("invalid Authorization header format")
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", fmt.Errorf("empty token")
	}

	return token, nil
}

// GetRelayClaimsFromContext извлекает RelayClaims из контекста
func GetRelayClaimsFromContext(ctx context.Context) (*RelayClaims, bool) {
	claims, ok := ctx.Value("relay_claims").(*RelayClaims)
	return claims, ok
}

// GetRelayServerIDFromContext извлекает ID relay сервера из контекста
func GetRelayServerIDFromContext(ctx context.Context) (string, bool) {
	serverID, ok := ctx.Value("relay_server_id").(string)
	return serverID, ok
}

// GetTenantIDFromContext извлекает ID тенанта из контекста
func GetTenantIDFromContext(ctx context.Context) (string, bool) {
	tenantID, ok := ctx.Value("tenant_id").(string)
	return tenantID, ok
}

// GetOrgIDFromContext извлекает ID организации из контекста
func GetOrgIDFromContext(ctx context.Context) (string, bool) {
	orgID, ok := ctx.Value("org_id").(string)
	return orgID, ok
}

// RequireTenantIsolation возвращает middleware, который проверяет изоляцию тенантов
func (m *RelayMiddleware) RequireTenantIsolation() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Получаем claims из контекста
			claims, ok := GetRelayClaimsFromContext(r.Context())
			if !ok {
				m.logger.Error("No relay claims found in context")
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			// Проверяем, что у relay сервера есть права на работу с сетью
			if !claims.HasPermission("network:allocate") || !claims.HasPermission("network:deallocate") {
				m.logger.Error("Insufficient network permissions for tenant isolation",
					zap.String("relay_server_id", claims.GetRelayServerID()),
					zap.String("tenant_id", claims.GetTenantID()))
				http.Error(w, "Forbidden: Insufficient network permissions", http.StatusForbidden)
				return
			}

			// Проверяем, что у relay сервера есть права на работу с тенантами
			if !claims.HasPermission("tenant:network:read") || !claims.HasPermission("tenant:network:write") {
				m.logger.Error("Insufficient tenant permissions for isolation",
					zap.String("relay_server_id", claims.GetRelayServerID()),
					zap.String("tenant_id", claims.GetTenantID()))
				http.Error(w, "Forbidden: Insufficient tenant permissions", http.StatusForbidden)
				return
			}

			m.logger.Info("Tenant isolation check passed",
				zap.String("relay_server_id", claims.GetRelayServerID()),
				zap.String("tenant_id", claims.GetTenantID()))

			next.ServeHTTP(w, r)
		})
	}
}
