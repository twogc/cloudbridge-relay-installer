package adminapi

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
	"go.uber.org/zap"
)

// AuthMiddleware проверяет JWT и роль admin-ui (заглушка)
func AuthMiddleware(logger *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Отладочная информация
		logger.Info("AuthMiddleware called", zap.String("path", c.Request.URL.Path))

		// Простая заглушка для аутентификации
		// TODO: Реализовать реальную аутентификацию через Zitadel
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			logger.Warn("Missing or invalid Authorization header", zap.String("header", header))
			c.AbortWithStatusJSON(401, gin.H{"error": "missing or invalid Authorization header"})
			return
		}

		// Простая проверка токена (заглушка)
		token := strings.TrimPrefix(header, "Bearer ")
		if token == "" {
			logger.Warn("Empty token")
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid token"})
			return
		}

		logger.Info("Token accepted (stub validation)", zap.String("token_prefix", token[:min(10, len(token))]+"..."))
		
		// Заглушка для пользователя
		userInfo := map[string]interface{}{
			"username": "admin",
			"roles":   []string{"admin-ui"},
		}
		c.Set("user", userInfo)
		c.Next()
	}
}

// min возвращает минимальное значение
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// RunServer запускает Admin API
func RunServer(addr string, logger *logger.Logger) error {
	r := gin.Default()

	// CORS: разрешаем только UI-админку (замените origin на ваш реальный)
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "https://auth.2gc.ru:5173"}, // Добавляем production origin
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * 60 * 60,
	}))

	r.Use(AuthMiddleware(logger)) // Упрощенная аутентификация

	r.GET("/api/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "Admin API is running",
		})
	})

	r.GET("/api/audit/logs", func(c *gin.Context) {
		const maxLines = 100
		file, err := os.Open("/var/log/relay/audit.log")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot open audit log", "details": err.Error()})
			if logger != nil {
				logger.Error("Cannot open audit log", zap.Error(err))
			}
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

	// Проверяем наличие TLS сертификатов
	certFile := "/etc/relay/certs/fullchain.pem"
	keyFile := "/etc/relay/certs/privkey.pem"

	if _, err := os.Stat(certFile); err == nil {
		if _, err := os.Stat(keyFile); err == nil {
			logger.Info("Starting Admin API with HTTPS", zap.String("addr", addr))
			return r.RunTLS(addr, certFile, keyFile)
		}
	}

	logger.Info("Starting Admin API with HTTP", zap.String("addr", addr))
	return r.Run(addr)
}
