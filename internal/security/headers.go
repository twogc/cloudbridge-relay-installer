package security

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// SecurityHeaders middleware добавляет security headers к HTTP ответам
func SecurityHeaders() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// X-Frame-Options: предотвращает clickjacking атаки
			w.Header().Set("X-Frame-Options", "DENY")

			// X-Content-Type-Options: предотвращает MIME type sniffing
			w.Header().Set("X-Content-Type-Options", "nosniff")

			// X-XSS-Protection: включает XSS фильтр браузера
			w.Header().Set("X-XSS-Protection", "1; mode=block")

			// Strict-Transport-Security: принуждает использовать HTTPS
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")

			// Referrer-Policy: контролирует передачу referrer информации
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

			// Content-Security-Policy: предотвращает XSS и injection атаки
			w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'; connect-src 'self'; frame-ancestors 'none';")

			// Permissions-Policy: контролирует доступ к браузерным API
			w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=(), payment=(), usb=(), magnetometer=(), gyroscope=(), speaker=(), vibrate=(), fullscreen=(self), sync-xhr=()")

			// X-Permitted-Cross-Domain-Policies: предотвращает cross-domain атаки
			w.Header().Set("X-Permitted-Cross-Domain-Policies", "none")

			// Cross-Origin headers для дополнительной защиты
			w.Header().Set("Cross-Origin-Embedder-Policy", "require-corp")
			w.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
			w.Header().Set("Cross-Origin-Resource-Policy", "same-origin")

			// Удаляем заголовки, которые могут раскрывать информацию о сервере
			w.Header().Del("Server")
			w.Header().Del("X-Powered-By")

			next.ServeHTTP(w, r)
		})
	}
}

// GinSecurityHeaders middleware для Gin framework
func GinSecurityHeaders() func(*gin.Context) {
	return func(c *gin.Context) {
		// X-Frame-Options: предотвращает clickjacking атаки
		c.Header("X-Frame-Options", "DENY")

		// X-Content-Type-Options: предотвращает MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// X-XSS-Protection: включает XSS фильтр браузера
		c.Header("X-XSS-Protection", "1; mode=block")

		// Strict-Transport-Security: принуждает использовать HTTPS
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")

		// Referrer-Policy: контролирует передачу referrer информации
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Content-Security-Policy: предотвращает XSS и injection атаки
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'; connect-src 'self'; frame-ancestors 'none';")

		// Permissions-Policy: контролирует доступ к браузерным API
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=(), payment=(), usb=(), magnetometer=(), gyroscope=(), speaker=(), vibrate=(), fullscreen=(self), sync-xhr=()")

		// X-Permitted-Cross-Domain-Policies: предотвращает cross-domain атаки
		c.Header("X-Permitted-Cross-Domain-Policies", "none")

		// Cross-Origin headers для дополнительной защиты
		c.Header("Cross-Origin-Embedder-Policy", "require-corp")
		c.Header("Cross-Origin-Opener-Policy", "same-origin")
		c.Header("Cross-Origin-Resource-Policy", "same-origin")

		// Удаляем заголовки, которые могут раскрывать информацию о сервере
		c.Header("Server", "")
		c.Header("X-Powered-By", "")

		c.Next()
	}
}
