package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SecurityConfig конфигурация для security middleware
type SecurityConfig struct {
	// Environment: development, staging, production
	Environment string `json:"environment"`
	
	// HTTPS settings
	ForceHTTPS           bool     `json:"force_https"`
	HSTSMaxAge           int      `json:"hsts_max_age"`           // в секундах
	HSTSIncludeSubdomain bool     `json:"hsts_include_subdomain"`
	HSTSPreload          bool     `json:"hsts_preload"`
	
	// CORS settings
	AllowedOrigins   []string `json:"allowed_origins"`
	AllowedMethods   []string `json:"allowed_methods"`
	AllowedHeaders   []string `json:"allowed_headers"`
	AllowCredentials bool     `json:"allow_credentials"`
	MaxAge           int      `json:"max_age"` // Preflight cache time
	
	// Security headers
	ContentTypeOptions    string `json:"content_type_options"`    // nosniff
	FrameOptions          string `json:"frame_options"`           // DENY, SAMEORIGIN
	XSSProtection         string `json:"xss_protection"`          // 1; mode=block
	ContentSecurityPolicy string `json:"content_security_policy"`
	ReferrerPolicy        string `json:"referrer_policy"`
	
	// Additional security
	RemoveServerHeader bool `json:"remove_server_header"`
	RemovePoweredBy    bool `json:"remove_powered_by"`
}

// DefaultSecurityConfig возвращает конфигурацию по умолчанию
func DefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		Environment:           "development",
		ForceHTTPS:           false,
		HSTSMaxAge:           31536000, // 1 год
		HSTSIncludeSubdomain: true,
		HSTSPreload:          true,
		AllowedOrigins: []string{
			"http://localhost:3000",
			"http://localhost:8080",
			"http://127.0.0.1:3000",
			"http://127.0.0.1:8080",
		},
		AllowedMethods: []string{
			"GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD",
		},
		AllowedHeaders: []string{
			"Origin", "Content-Type", "Accept", "Authorization",
			"X-Requested-With", "X-CSRF-Token",
		},
		AllowCredentials:      true,
		MaxAge:               86400, // 24 часа
		ContentTypeOptions:    "nosniff",
		FrameOptions:          "DENY",
		XSSProtection:         "1; mode=block",
		ContentSecurityPolicy: "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'",
		ReferrerPolicy:        "strict-origin-when-cross-origin",
		RemoveServerHeader:    true,
		RemovePoweredBy:       true,
	}
}

// ProductionSecurityConfig возвращает конфигурацию для продакшена
func ProductionSecurityConfig(allowedOrigins []string) *SecurityConfig {
	config := DefaultSecurityConfig()
	config.Environment = "production"
	config.ForceHTTPS = true
	
	if len(allowedOrigins) > 0 {
		config.AllowedOrigins = allowedOrigins
	} else {
		// В продакшене не должно быть wildcard origins
		config.AllowedOrigins = []string{
			"https://cloudbridge.example.com",
			"https://api.cloudbridge.example.com",
		}
	}
	
	// Более строгая CSP для продакшена
	config.ContentSecurityPolicy = "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data:; font-src 'self'; connect-src 'self'; frame-ancestors 'none'"
	
	return config
}

// SecurityMiddleware middleware для безопасности
type SecurityMiddleware struct {
	config *SecurityConfig
	logger *zap.Logger
}

// NewSecurityMiddleware создает новый security middleware
func NewSecurityMiddleware(config *SecurityConfig, logger *zap.Logger) *SecurityMiddleware {
	if config == nil {
		config = DefaultSecurityConfig()
	}
	
	return &SecurityMiddleware{
		config: config,
		logger: logger,
	}
}

// SecurityHeaders добавляет security headers
func (sm *SecurityMiddleware) SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// X-Content-Type-Options
		if sm.config.ContentTypeOptions != "" {
			c.Header("X-Content-Type-Options", sm.config.ContentTypeOptions)
		}
		
		// X-Frame-Options
		if sm.config.FrameOptions != "" {
			c.Header("X-Frame-Options", sm.config.FrameOptions)
		}
		
		// X-XSS-Protection
		if sm.config.XSSProtection != "" {
			c.Header("X-XSS-Protection", sm.config.XSSProtection)
		}
		
		// Content-Security-Policy
		if sm.config.ContentSecurityPolicy != "" {
			c.Header("Content-Security-Policy", sm.config.ContentSecurityPolicy)
		}
		
		// Referrer-Policy
		if sm.config.ReferrerPolicy != "" {
			c.Header("Referrer-Policy", sm.config.ReferrerPolicy)
		}
		
		// Strict-Transport-Security (только для HTTPS)
		if sm.config.ForceHTTPS && c.Request.TLS != nil {
			hstsValue := fmt.Sprintf("max-age=%d", sm.config.HSTSMaxAge)
			if sm.config.HSTSIncludeSubdomain {
				hstsValue += "; includeSubDomains"
			}
			if sm.config.HSTSPreload {
				hstsValue += "; preload"
			}
			c.Header("Strict-Transport-Security", hstsValue)
		}
		
		// Remove server identification headers
		if sm.config.RemoveServerHeader {
			c.Header("Server", "")
		}
		if sm.config.RemovePoweredBy {
			c.Header("X-Powered-By", "")
		}
		
		c.Next()
	}
}

// CORS настраивает CORS с улучшенной безопасностью
func (sm *SecurityMiddleware) CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// Проверяем, разрешен ли origin
		if sm.isOriginAllowed(origin) {
			c.Header("Access-Control-Allow-Origin", origin)
		} else if sm.config.Environment == "development" && origin == "" {
			// В development режиме разрешаем запросы без Origin (например, Postman)
			c.Header("Access-Control-Allow-Origin", "*")
		}
		
		// Методы
		c.Header("Access-Control-Allow-Methods", strings.Join(sm.config.AllowedMethods, ", "))
		
		// Заголовки
		c.Header("Access-Control-Allow-Headers", strings.Join(sm.config.AllowedHeaders, ", "))
		
		// Credentials
		if sm.config.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}
		
		// Max age для preflight запросов
		c.Header("Access-Control-Max-Age", fmt.Sprintf("%d", sm.config.MaxAge))
		
		// Обрабатываем preflight запросы
		if c.Request.Method == "OPTIONS" {
			sm.logger.Debug("Handling CORS preflight request",
				zap.String("origin", origin),
				zap.String("method", c.Request.Header.Get("Access-Control-Request-Method")),
				zap.String("headers", c.Request.Header.Get("Access-Control-Request-Headers")),
			)
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		
		c.Next()
	}
}

// HTTPSRedirect принудительно перенаправляет на HTTPS
func (sm *SecurityMiddleware) HTTPSRedirect() gin.HandlerFunc {
	return func(c *gin.Context) {
		if sm.config.ForceHTTPS && c.Request.TLS == nil {
			// Проверяем заголовки от reverse proxy
			if c.Request.Header.Get("X-Forwarded-Proto") != "https" &&
			   c.Request.Header.Get("X-Forwarded-Ssl") != "on" {
				
				httpsURL := "https://" + c.Request.Host + c.Request.RequestURI
				
				sm.logger.Info("Redirecting HTTP to HTTPS",
					zap.String("original_url", c.Request.URL.String()),
					zap.String("https_url", httpsURL),
					zap.String("remote_addr", c.ClientIP()),
				)
				
				c.Redirect(http.StatusMovedPermanently, httpsURL)
				c.Abort()
				return
			}
		}
		
		c.Next()
	}
}

// isOriginAllowed проверяет, разрешен ли origin
func (sm *SecurityMiddleware) isOriginAllowed(origin string) bool {
	if origin == "" {
		return false
	}
	
	for _, allowedOrigin := range sm.config.AllowedOrigins {
		if allowedOrigin == "*" {
			return true
		}
		if allowedOrigin == origin {
			return true
		}
		// Поддержка wildcard поддоменов (например, *.example.com)
		if strings.HasPrefix(allowedOrigin, "*.") {
			domain := strings.TrimPrefix(allowedOrigin, "*.")
			if strings.HasSuffix(origin, "."+domain) || origin == "https://"+domain || origin == "http://"+domain {
				return true
			}
		}
	}
	
	return false
}

// LogSecurityEvent логирует security события
func (sm *SecurityMiddleware) LogSecurityEvent(c *gin.Context, event string, details map[string]interface{}) {
	fields := []zap.Field{
		zap.String("event", event),
		zap.String("path", c.Request.URL.Path),
		zap.String("method", c.Request.Method),
		zap.String("remote_addr", c.ClientIP()),
		zap.String("user_agent", c.Request.UserAgent()),
		zap.String("origin", c.Request.Header.Get("Origin")),
	}
	
	for key, value := range details {
		fields = append(fields, zap.Any(key, value))
	}
	
	sm.logger.Warn("Security event", fields...)
}

// SecurityAudit middleware для аудита безопасности
func (sm *SecurityMiddleware) SecurityAudit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Проверяем подозрительные паттерны в URL
		suspiciousPatterns := []string{
			"../", "..\\", ".env", "wp-admin", "phpmyadmin",
			"<script", "javascript:", "vbscript:", "onload=",
		}
		
		path := strings.ToLower(c.Request.URL.Path)
		query := strings.ToLower(c.Request.URL.RawQuery)
		
		for _, pattern := range suspiciousPatterns {
			if strings.Contains(path, pattern) || strings.Contains(query, pattern) {
				sm.LogSecurityEvent(c, "suspicious_request", map[string]interface{}{
					"pattern": pattern,
					"path":    c.Request.URL.Path,
					"query":   c.Request.URL.RawQuery,
				})
				break
			}
		}
		
		// Проверяем размер запроса
		if c.Request.ContentLength > 10*1024*1024 { // 10MB
			sm.LogSecurityEvent(c, "large_request", map[string]interface{}{
				"content_length": c.Request.ContentLength,
			})
		}
		
		// Проверяем User-Agent
		userAgent := c.Request.UserAgent()
		if userAgent == "" {
			sm.LogSecurityEvent(c, "missing_user_agent", nil)
		}
		
		c.Next()
	}
}

// GetConfig возвращает текущую конфигурацию
func (sm *SecurityMiddleware) GetConfig() *SecurityConfig {
	return sm.config
}

// UpdateConfig обновляет конфигурацию
func (sm *SecurityMiddleware) UpdateConfig(config *SecurityConfig) {
	sm.config = config
	sm.logger.Info("Security configuration updated",
		zap.String("environment", config.Environment),
		zap.Bool("force_https", config.ForceHTTPS),
		zap.Strings("allowed_origins", config.AllowedOrigins),
	)
}
