package network

import (
	"fmt"
	"net/http"
)

// HTTPError представляет HTTP ошибку с дополнительной информацией
type HTTPError struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
	RetryAfter string `json:"retry_after,omitempty"`
	Body       string `json:"body,omitempty"`
}

// NewHTTPError создает новую HTTP ошибку
func NewHTTPError(statusCode int, message string) *HTTPError {
	return &HTTPError{
		StatusCode: statusCode,
		Message:    message,
	}
}

// NewHTTPErrorWithRetryAfter создает HTTP ошибку с Retry-After
func NewHTTPErrorWithRetryAfter(statusCode int, message, retryAfter string) *HTTPError {
	return &HTTPError{
		StatusCode: statusCode,
		Message:    message,
		RetryAfter: retryAfter,
	}
}

// Error реализует интерфейс error
func (e *HTTPError) Error() string {
	if e.Message == "" {
		return fmt.Sprintf("HTTP %d", e.StatusCode)
	}
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Message)
}

// IsRetryable проверяет, можно ли повторить запрос при данной HTTP ошибке
func (e *HTTPError) IsRetryable() bool {
	// 4xx ошибки обычно не ретраим, кроме 408, 429
	if e.StatusCode >= 400 && e.StatusCode < 500 {
		return e.StatusCode == http.StatusRequestTimeout || e.StatusCode == http.StatusTooManyRequests
	}

	// 5xx ошибки обычно ретраим
	return e.StatusCode >= 500
}

// GetRetryAfter возвращает Retry-After заголовок
func (e *HTTPError) GetRetryAfter() string {
	return e.RetryAfter
}

// IsTooManyRequests проверяет, является ли ошибка "Too Many Requests"
func (e *HTTPError) IsTooManyRequests() bool {
	return e.StatusCode == http.StatusTooManyRequests
}

// IsServiceUnavailable проверяет, является ли ошибка "Service Unavailable"
func (e *HTTPError) IsServiceUnavailable() bool {
	return e.StatusCode == http.StatusServiceUnavailable
}

// IsConflict проверяет, является ли ошибка "Conflict"
func (e *HTTPError) IsConflict() bool {
	return e.StatusCode == http.StatusConflict
}

// IsNotFound проверяет, является ли ошибка "Not Found"
func (e *HTTPError) IsNotFound() bool {
	return e.StatusCode == http.StatusNotFound
}
