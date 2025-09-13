package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AuthMiddleware middleware для аутентификации API endpoints
type AuthMiddleware struct {
	validator TokenValidator
	logger    *zap.Logger
}

// NewAuthMiddleware создает новый middleware для аутентификации
func NewAuthMiddleware(validator TokenValidator, logger *zap.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		validator: validator,
		logger:    logger,
	}
}

// RequireAuth middleware, который требует аутентификации
func (am *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Извлекаем токен из заголовка Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			am.logger.Warn("Missing Authorization header",
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
				zap.String("remote_addr", c.ClientIP()),
			)
			c.JSON(http.StatusUnauthorized, ErrorResponse{
				Error:   "Missing Authorization header",
				Code:    http.StatusUnauthorized,
				Message: "Authorization header is required",
			})
			c.Abort()
			return
		}

		// Извлекаем токен из "Bearer <token>"
		token := authHeader
		if strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		}
		if token == "" {
			am.logger.Warn("Empty token in Authorization header",
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
				zap.String("remote_addr", c.ClientIP()),
			)
			c.JSON(http.StatusUnauthorized, ErrorResponse{
				Error:   "Invalid authorization header",
				Code:    http.StatusUnauthorized,
				Message: "Authorization header must contain a valid token",
			})
			c.Abort()
			return
		}

		// Валидируем токен
		claims, err := am.validator.ValidateP2PToken(token)
		if err != nil {
			am.logger.Warn("Token validation failed",
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
				zap.String("remote_addr", c.ClientIP()),
				zap.Error(err),
			)
			c.JSON(http.StatusUnauthorized, ErrorResponse{
				Error:   "Invalid token",
				Code:    http.StatusUnauthorized,
				Message: err.Error(),
			})
			c.Abort()
			return
		}

		// Сохраняем claims в контексте для использования в handlers
		c.Set("claims", claims)
		c.Set("tenant_id", claims.TenantID)
		c.Set("user_id", claims.Subject)

		am.logger.Debug("Authentication successful",
			zap.String("path", c.Request.URL.Path),
			zap.String("method", c.Request.Method),
			zap.String("tenant_id", claims.TenantID),
			zap.String("user_id", claims.Subject),
			zap.String("remote_addr", c.ClientIP()),
		)

		c.Next()
	}
}

// RequireAdmin middleware, который требует административных прав
func (am *AuthMiddleware) RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Сначала проверяем аутентификацию
		am.RequireAuth()(c)
		if c.IsAborted() {
			return
		}

		// Получаем claims из контекста
		claims, exists := c.Get("claims")
		if !exists {
			am.logger.Error("Claims not found in context")
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Internal server error",
				Code:    http.StatusInternalServerError,
				Message: "Authentication context not found",
			})
			c.Abort()
			return
		}

		p2pClaims, ok := claims.(*P2PClaims)
		if !ok {
			am.logger.Error("Invalid claims type in context")
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Internal server error",
				Code:    http.StatusInternalServerError,
				Message: "Invalid authentication context",
			})
			c.Abort()
			return
		}

		// Проверяем административные права
		// В реальной реализации здесь должна быть проверка ролей/разрешений
		// Пока что проверяем, что пользователь имеет права администратора
		if !am.hasAdminRights(p2pClaims) {
			am.logger.Warn("Access denied - insufficient privileges",
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
				zap.String("tenant_id", p2pClaims.TenantID),
				zap.String("user_id", p2pClaims.Subject),
				zap.String("remote_addr", c.ClientIP()),
			)
			c.JSON(http.StatusForbidden, ErrorResponse{
				Error:   "Insufficient privileges",
				Code:    http.StatusForbidden,
				Message: "Administrator privileges required",
			})
			c.Abort()
			return
		}

		am.logger.Debug("Admin access granted",
			zap.String("path", c.Request.URL.Path),
			zap.String("method", c.Request.Method),
			zap.String("tenant_id", p2pClaims.TenantID),
			zap.String("user_id", p2pClaims.Subject),
		)

		c.Next()
	}
}

// hasAdminRights проверяет, имеет ли пользователь административные права
func (am *AuthMiddleware) hasAdminRights(claims *P2PClaims) bool {
	// В реальной реализации здесь должна быть проверка ролей из токена
	// или запрос к базе данных для проверки прав пользователя

	// Пока что используем простую проверку на основе tenant_id
	// Администраторы имеют доступ к определенным tenant'ам
	adminTenants := []string{
		"admin",
		"system",
		"cloudbridge-admin",
	}

	for _, adminTenant := range adminTenants {
		if claims.TenantID == adminTenant {
			return true
		}
	}

	// Также проверяем, есть ли в claims информация о роли
	// Это зависит от структуры JWT токена
	if claims.Permissions != nil {
		for _, permission := range claims.Permissions {
			if permission == "admin" || permission == "system" {
				return true
			}
		}
	}

	return false
}

// GetClaimsFromContext извлекает claims из контекста Gin
func GetClaimsFromContext(c *gin.Context) (*P2PClaims, bool) {
	claims, exists := c.Get("claims")
	if !exists {
		return nil, false
	}

	p2pClaims, ok := claims.(*P2PClaims)
	return p2pClaims, ok
}

// GetTenantIDFromContext извлекает tenant_id из контекста Gin
func GetTenantIDFromContext(c *gin.Context) (string, bool) {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		return "", false
	}

	tenantIDStr, ok := tenantID.(string)
	return tenantIDStr, ok
}

// GetUserIDFromContext извлекает user_id из контекста Gin
func GetUserIDFromContext(c *gin.Context) (string, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return "", false
	}

	userIDStr, ok := userID.(string)
	return userIDStr, ok
}
