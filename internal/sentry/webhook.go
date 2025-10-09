package sentry

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// WebhookConfig конфигурация webhook для Sentry
type WebhookConfig struct {
	URL     string `json:"url"`
	Secret  string `json:"secret"`
	Enabled bool   `json:"enabled"`
}

// WebhookEvent событие webhook от Sentry
type WebhookEvent struct {
	Action     string                 `json:"action"`
	Installation *WebhookInstallation `json:"installation,omitempty"`
	Data       map[string]interface{} `json:"data"`
	CreatedAt  time.Time             `json:"created_at"`
}

// WebhookInstallation информация об установке
type WebhookInstallation struct {
	UUID string `json:"uuid"`
}

// WebhookHandler обработчик webhook событий
type WebhookHandler struct {
	config *WebhookConfig
	client *http.Client
}

// NewWebhookHandler создает новый обработчик webhook
func NewWebhookHandler(config *WebhookConfig) *WebhookHandler {
	return &WebhookHandler{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// VerifySignature проверяет подпись webhook
func (h *WebhookHandler) VerifySignature(payload []byte, signature string) bool {
	if h.config.Secret == "" {
		return true // Если секрет не настроен, пропускаем проверку
	}

	expectedMAC := h.calculateHMAC(payload)
	expectedSignature := fmt.Sprintf("sha256=%s", expectedMAC)
	
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// calculateHMAC вычисляет HMAC для подписи
func (h *WebhookHandler) calculateHMAC(payload []byte) string {
	mac := hmac.New(sha256.New, []byte(h.config.Secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

// HandleWebhook обрабатывает webhook событие
func (h *WebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	// Читаем тело запроса
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read payload", http.StatusBadRequest)
		return
	}

	// Проверяем подпись
	signature := r.Header.Get("X-Sentry-Hook-Signature")
	if !h.VerifySignature(payload, signature) {
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	// Парсим событие
	var event WebhookEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		http.Error(w, "Failed to parse event", http.StatusBadRequest)
		return
	}

	// Обрабатываем событие
	if err := h.processEvent(&event); err != nil {
		http.Error(w, "Failed to process event", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// processEvent обрабатывает событие webhook
func (h *WebhookHandler) processEvent(event *WebhookEvent) error {
	switch event.Action {
	case "created":
		return h.handleIssueCreated(event)
	case "resolved":
		return h.handleIssueResolved(event)
	case "assigned":
		return h.handleIssueAssigned(event)
	case "ignored":
		return h.handleIssueIgnored(event)
	default:
		return fmt.Errorf("unknown action: %s", event.Action)
	}
}

// handleIssueCreated обрабатывает создание новой проблемы
func (h *WebhookHandler) handleIssueCreated(event *WebhookEvent) error {
	// Логируем создание новой проблемы
	fmt.Printf("🚨 New issue created: %+v\n", event.Data)
	
	// Здесь можно добавить логику уведомлений
	// Например, отправка в Slack, email, etc.
	
	return nil
}

// handleIssueResolved обрабатывает решение проблемы
func (h *WebhookHandler) handleIssueResolved(event *WebhookEvent) error {
	// Логируем решение проблемы
	fmt.Printf("✅ Issue resolved: %+v\n", event.Data)
	
	return nil
}

// handleIssueAssigned обрабатывает назначение проблемы
func (h *WebhookHandler) handleIssueAssigned(event *WebhookEvent) error {
	// Логируем назначение проблемы
	fmt.Printf("👤 Issue assigned: %+v\n", event.Data)
	
	return nil
}

// handleIssueIgnored обрабатывает игнорирование проблемы
func (h *WebhookHandler) handleIssueIgnored(event *WebhookEvent) error {
	// Логируем игнорирование проблемы
	fmt.Printf("🙈 Issue ignored: %+v\n", event.Data)
	
	return nil
}

// DefaultWebhookConfig возвращает конфигурацию webhook по умолчанию
func DefaultWebhookConfig() *WebhookConfig {
	return &WebhookConfig{
		URL:     getEnvOrDefault("SENTRY_WEBHOOK_URL", ""),
		Secret:  getEnvOrDefault("SENTRY_WEBHOOK_SECRET", ""),
		Enabled: getEnvOrDefault("SENTRY_WEBHOOK_ENABLED", "false") == "true",
	}
}



