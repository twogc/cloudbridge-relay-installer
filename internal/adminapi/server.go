package adminapi

import (
	"net/http"

	"bufio"
	"encoding/json"
	"os"

	"context"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/twogc/cloudbridge-relay-installer/internal/keycloak"
	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
	"go.uber.org/zap"
)

// AuthMiddleware проверяет JWT и роль admin-ui
func AuthMiddleware(keycloakClient *keycloak.Client, logger *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Отладочная информация
		logger.Info("AuthMiddleware called", zap.String("path", c.Request.URL.Path))
		
		// Проверяем, что Keycloak клиент инициализирован
		if keycloakClient == nil {
			logger.Error("Keycloak client is nil - middleware disabled")
			c.AbortWithStatusJSON(500, gin.H{"error": "authentication service unavailable"})
			return
		}
		
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			logger.Warn("Missing or invalid Authorization header", zap.String("header", header))
			c.AbortWithStatusJSON(401, gin.H{"error": "missing or invalid Authorization header"})
			return
		}
		
		token := strings.TrimPrefix(header, "Bearer ")
		logger.Info("Validating token", zap.String("token_prefix", token[:10]+"..."))
		
		tokenInfo, err := keycloakClient.ValidateToken(context.Background(), token)
		if err != nil {
			logger.Warn("Token validation failed", zap.Error(err))
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid token"})
			return
		}
		
		logger.Info("Token validated successfully", zap.String("username", tokenInfo.Username), zap.Strings("roles", tokenInfo.Roles))
		
		isAdmin := false
		for _, role := range tokenInfo.Roles {
			if role == "admin-ui" {
				isAdmin = true
				break
			}
		}
		
		if !isAdmin {
			logger.Warn("Access denied: missing admin-ui role", zap.Strings("user_roles", tokenInfo.Roles))
			c.AbortWithStatusJSON(403, gin.H{"error": "forbidden: admin-ui role required"})
			return
		}
		
		logger.Info("Access granted", zap.String("username", tokenInfo.Username))
		// Можно добавить user info в context
		c.Set("user", tokenInfo)
		c.Next()
	}
}

// RunServer запускает Admin API с авторизацией через Keycloak
func RunServer(addr string, keycloakClient *keycloak.Client, logger *logger.Logger) error {
	r := gin.Default()

	// CORS: разрешаем только UI-админку (замените origin на ваш реальный)
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"}, // TODO: заменить на production origin
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * 60 * 60,
	}))

	r.Use(AuthMiddleware(keycloakClient, logger))

	r.GET("/api/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"message": "Admin API is running",
		})
	})

	r.GET("/api/audit/logs", func(c *gin.Context) {
		const maxLines = 100
		file, err := os.Open("logs/audit.log")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot open audit log"})
			return
		}
		defer file.Close()

		var lines []string
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot read audit log"})
			return
		}

		// Берём последние maxLines
		start := 0
		if len(lines) > maxLines {
			start = len(lines) - maxLines
		}
		lines = lines[start:]

		var events []map[string]interface{}
		for _, line := range lines {
			var evt map[string]interface{}
			if err := json.Unmarshal([]byte(line), &evt); err == nil {
				events = append(events, evt)
			}
		}

		c.JSON(http.StatusOK, gin.H{"events": events})
	})

	return r.Run(addr)
} 