package notifications

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/twogc/cloudbridge-relay-installer/internal/ai/analytics"
)

// NotificationSystem система уведомлений
type NotificationSystem struct {
	slackWebhook   string
	emailConfig    *EmailConfig
	telegramConfig *TelegramConfig
	webhooks       []string
}

// EmailConfig конфигурация Email
type EmailConfig struct {
	SMTPHost   string
	SMTPPort   int
	Username   string
	Password   string
	Recipients []string
}

// TelegramConfig конфигурация Telegram
type TelegramConfig struct {
	BotToken              string `json:"bot_token"`
	ChatID                string `json:"chat_id"`
	ParseMode             string `json:"parse_mode"` // HTML, Markdown
	DisableWebPagePreview bool   `json:"disable_web_page_preview"`
}

// NotificationConfig конфигурация уведомлений
type NotificationConfig struct {
	SlackWebhook   string          `json:"slack_webhook"`
	EmailConfig    *EmailConfig    `json:"email_config"`
	TelegramConfig *TelegramConfig `json:"telegram_config"`
	Webhooks       []string        `json:"webhooks"`
}

// NewNotificationSystem создает новую систему уведомлений
func NewNotificationSystem() *NotificationSystem {
	return &NotificationSystem{
		webhooks: make([]string, 0),
	}
}

// AddSlackWebhook добавляет Slack webhook
func (n *NotificationSystem) AddSlackWebhook(webhook string) {
	n.slackWebhook = webhook
}

// AddTelegramConfig добавляет Telegram конфигурацию
func (n *NotificationSystem) AddTelegramConfig(config *TelegramConfig) {
	n.telegramConfig = config
}

// Configure настраивает систему уведомлений
func (n *NotificationSystem) Configure(config NotificationConfig) {
	n.slackWebhook = config.SlackWebhook
	n.emailConfig = config.EmailConfig
	n.telegramConfig = config.TelegramConfig
	n.webhooks = config.Webhooks
}

// ConfigureFromMap настраивает систему уведомлений из произвольной карты (для runtime API)
func (n *NotificationSystem) ConfigureFromMap(m map[string]interface{}) {
	cfg := NotificationConfig{}
	if v, ok := m["slack_webhook"].(string); ok {
		cfg.SlackWebhook = v
	}
	if v, ok := m["webhooks"].([]interface{}); ok {
		cfg.Webhooks = make([]string, 0, len(v))
		for _, it := range v {
			// поддержка как строк, так и объектов {url, enabled}
			if s, ok := it.(string); ok {
				cfg.Webhooks = append(cfg.Webhooks, s)
				continue
			}
			if obj, ok := it.(map[string]interface{}); ok {
				if u, ok := obj["url"].(string); ok {
					// если есть enabled=false — пропускаем
					if en, ok := obj["enabled"].(bool); ok && !en {
						continue
					}
					cfg.Webhooks = append(cfg.Webhooks, u)
				}
			}
		}
	}
	if v, ok := m["email_config"].(map[string]interface{}); ok {
		ec := &EmailConfig{}
		if s, ok := v["smtp_host"].(string); ok {
			ec.SMTPHost = s
		}
		if f, ok := v["smtp_port"].(float64); ok {
			ec.SMTPPort = int(f)
		}
		if s, ok := v["username"].(string); ok {
			ec.Username = s
		}
		if s, ok := v["password"].(string); ok {
			ec.Password = s
		}
		if arr, ok := v["recipients"].([]interface{}); ok {
			ec.Recipients = make([]string, 0, len(arr))
			for _, it := range arr {
				if s, ok := it.(string); ok {
					ec.Recipients = append(ec.Recipients, s)
				}
			}
		}
		cfg.EmailConfig = ec
	}
	if v, ok := m["telegram_config"].(map[string]interface{}); ok {
		tc := &TelegramConfig{}
		if s, ok := v["bot_token"].(string); ok {
			tc.BotToken = s
		}
		if s, ok := v["chat_id"].(string); ok {
			tc.ChatID = s
		}
		if s, ok := v["parse_mode"].(string); ok {
			tc.ParseMode = s
		}
		if b, ok := v["disable_web_page_preview"].(bool); ok {
			tc.DisableWebPagePreview = b
		}
		cfg.TelegramConfig = tc
	}
	n.Configure(cfg)
}

// SendReport отправляет отчет через все настроенные каналы
func (n *NotificationSystem) SendReport(report *analytics.Report) error {
	if n.telegramConfig == nil && n.slackWebhook == "" && len(n.webhooks) == 0 {
		fmt.Println("NotificationSystem: no channels configured; skipping send")
	}
	// Slack уведомление
	if n.slackWebhook != "" {
		if err := n.sendSlackNotification(report); err != nil {
			// Не прерываем цепочку уведомлений, логируем и продолжаем (чтобы Telegram не блокировался Slack'ом)
			fmt.Printf("Slack notification failed: %v\n", err)
		}
	}

	// Telegram уведомление
	if n.telegramConfig != nil {
		if err := n.sendTelegramNotification(report); err != nil {
			return fmt.Errorf("failed to send Telegram notification: %w", err)
		}
	}

	// Webhook уведомления
	for _, webhook := range n.webhooks {
		if err := n.sendWebhookNotification(report, webhook); err != nil {
			// Логируем ошибку, но не прерываем процесс
			fmt.Printf("Failed to send webhook to %s: %v\n", webhook, err)
		}
	}

	return nil
}

// sendSlackNotification отправляет уведомление в Slack
func (n *NotificationSystem) sendSlackNotification(report *analytics.Report) error {
	payload := map[string]interface{}{
		"text": "🤖 CloudBridge AI Analytics Report",
		"blocks": []map[string]interface{}{
			{
				"type": "header",
				"text": map[string]interface{}{
					"type": "plain_text",
					"text": "🤖 CloudBridge AI Analytics Report",
				},
			},
			{
				"type": "section",
				"text": map[string]interface{}{
					"type": "mrkdwn",
					"text": fmt.Sprintf("*%s*\n\n%s",
						report.Timestamp.Format("2006-01-02 15:04:05"),
						report.Analysis.Summary),
				},
			},
			{
				"type": "divider",
			},
		},
	}

	// Добавление инсайтов
	if len(report.Analysis.Insights) > 0 {
		insightsText := "*🔍 Ключевые инсайты:*\n"
		for _, insight := range report.Analysis.Insights {
			insightsText += fmt.Sprintf("• %s\n", insight)
		}

		payload["blocks"] = append(payload["blocks"].([]map[string]interface{}), map[string]interface{}{
			"type": "section",
			"text": map[string]interface{}{
				"type": "mrkdwn",
				"text": insightsText,
			},
		})
	}

	// Добавление критических проблем
	if len(report.Analysis.CriticalIssues) > 0 {
		issuesText := "*🚨 Критические проблемы:*\n"
		for _, issue := range report.Analysis.CriticalIssues {
			issuesText += fmt.Sprintf("• %s\n", issue)
		}

		payload["blocks"] = append(payload["blocks"].([]map[string]interface{}), map[string]interface{}{
			"type": "section",
			"text": map[string]interface{}{
				"type": "mrkdwn",
				"text": issuesText,
			},
		})
	}

	// Добавление рекомендаций
	if len(report.Analysis.Recommendations) > 0 {
		recommendationsText := "*💡 Рекомендации:*\n"
		for _, rec := range report.Analysis.Recommendations {
			recommendationsText += fmt.Sprintf("• %s\n", rec)
		}

		payload["blocks"] = append(payload["blocks"].([]map[string]interface{}), map[string]interface{}{
			"type": "section",
			"text": map[string]interface{}{
				"type": "mrkdwn",
				"text": recommendationsText,
			},
		})
	}

	// Добавление информации о производительности
	if report.Analysis.Performance != "" {
		payload["blocks"] = append(payload["blocks"].([]map[string]interface{}), map[string]interface{}{
			"type": "section",
			"text": map[string]interface{}{
				"type": "mrkdwn",
				"text": fmt.Sprintf("*📊 Производительность:* %s", report.Analysis.Performance),
			},
		})
	}

	// Добавление информации о здоровье системы
	if report.Metrics.Health != nil {
		payload["blocks"] = append(payload["blocks"].([]map[string]interface{}), map[string]interface{}{
			"type": "section",
			"text": map[string]interface{}{
				"type": "mrkdwn",
				"text": fmt.Sprintf("*🏥 Статус здоровья:* %s (%.1f/100)",
					report.Metrics.Health.Overall, report.Metrics.Health.Score),
			},
		})
	}

	// Добавление footer
	payload["blocks"] = append(payload["blocks"].([]map[string]interface{}), map[string]interface{}{
		"type": "context",
		"elements": []map[string]interface{}{
			{
				"type": "mrkdwn",
				"text": fmt.Sprintf("Report ID: %s | Generated by CloudBridge AI Analytics", report.ID),
			},
		},
	})

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(n.slackWebhook, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// sendWebhookNotification отправляет уведомление через webhook
func (n *NotificationSystem) sendWebhookNotification(report *analytics.Report, webhookURL string) error {
	payload := map[string]interface{}{
		"timestamp":       report.Timestamp,
		"report_id":       report.ID,
		"analysis":        report.Analysis,
		"health":          report.Metrics.Health,
		"anomalies_count": len(report.Metrics.Anomalies),
		"source":          "cloudbridge_ai_analytics",
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// sendTelegramNotification отправляет уведомление в Telegram
func (n *NotificationSystem) sendTelegramNotification(report *analytics.Report) error {
	// Форматирование сообщения для Telegram
	message := n.formatTelegramMessage(report)

	// Подготовка запроса
	payload := map[string]string{
		"chat_id":    n.telegramConfig.ChatID,
		"text":       message,
		"parse_mode": n.telegramConfig.ParseMode,
	}

	if n.telegramConfig.DisableWebPagePreview {
		payload["disable_web_page_preview"] = "true"
	}

	// Отправка запроса
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", n.telegramConfig.BotToken)
	// Логи до запроса (маскируем токен)
	fmt.Printf("Telegram: sending message to chat_id=%s parse_mode=%s url=...%s\n",
		n.telegramConfig.ChatID,
		n.telegramConfig.ParseMode,
		maskTokenTail(url, 8))
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Telegram: non-200 response status=%d body=%s\n", resp.StatusCode, string(body))
		return fmt.Errorf("telegram API error: status %d", resp.StatusCode)
	}

	fmt.Println("Telegram: message sent successfully (HTTP 200)")

	return nil
}

// formatTelegramMessage форматирует сообщение для Telegram
func (n *NotificationSystem) formatTelegramMessage(report *analytics.Report) string {
	var message strings.Builder

	// Заголовок
	message.WriteString("<b>CloudBridge AI Analytics Report</b>\n")
	message.WriteString(fmt.Sprintf("Timestamp: %s\n\n", report.Timestamp.Format("2006-01-02 15:04:05")))

	// Краткое резюме
	if report.Analysis.Summary != "" {
		message.WriteString("<b>Summary:</b>\n")
		message.WriteString(report.Analysis.Summary + "\n\n")
	}

	// Инсайты
	if len(report.Analysis.Insights) > 0 {
		message.WriteString("<b>Key Insights:</b>\n")
		for _, insight := range report.Analysis.Insights {
			message.WriteString(fmt.Sprintf("• %s\n", insight))
		}
		message.WriteString("\n")
	}

	// Критические проблемы
	if len(report.Analysis.CriticalIssues) > 0 {
		message.WriteString("<b>Critical Issues:</b>\n")
		for _, issue := range report.Analysis.CriticalIssues {
			message.WriteString(fmt.Sprintf("• %s\n", issue))
		}
		message.WriteString("\n")
	}

	// Рекомендации
	if len(report.Analysis.Recommendations) > 0 {
		message.WriteString("<b>Recommendations:</b>\n")
		for _, rec := range report.Analysis.Recommendations {
			message.WriteString(fmt.Sprintf("• %s\n", rec))
		}
		message.WriteString("\n")
	}

	// Производительность
	if report.Analysis.Performance != "" {
		message.WriteString(fmt.Sprintf("<b>Performance:</b> %s\n", report.Analysis.Performance))
	}

	// Здоровье системы
	if report.Metrics.Health != nil {
		message.WriteString(fmt.Sprintf("<b>Health Status:</b> %s (%.1f/100)\n",
			report.Metrics.Health.Overall, report.Metrics.Health.Score))
	}

	// Footer
	message.WriteString(fmt.Sprintf("\nReport ID: %s | Generated by CloudBridge AI Analytics", report.ID))

	return message.String()
}

// SendResourceAlert отправляет алерт о ресурсах (диск, CPU, RAM)
func (n *NotificationSystem) SendResourceAlert(alertType string, usage float64, threshold float64, details map[string]interface{}) error {
	if n.telegramConfig == nil {
		return fmt.Errorf("telegram config not set")
	}

	message := n.formatResourceAlertMessage(alertType, usage, threshold, details)

	// Подготовка запроса
	payload := map[string]string{
		"chat_id":    n.telegramConfig.ChatID,
		"text":       message,
		"parse_mode": n.telegramConfig.ParseMode,
	}

	if n.telegramConfig.DisableWebPagePreview {
		payload["disable_web_page_preview"] = "true"
	}

	// Отправка запроса
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", n.telegramConfig.BotToken)
	fmt.Printf("Telegram: sending resource alert to chat_id=%s parse_mode=%s url=...%s\n",
		n.telegramConfig.ChatID,
		n.telegramConfig.ParseMode,
		maskTokenTail(url, 8))
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Telegram: non-200 response status=%d body=%s\n", resp.StatusCode, string(body))
		return fmt.Errorf("telegram API error: status %d", resp.StatusCode)
	}

	return nil
}

// formatResourceAlertMessage форматирует сообщение об алерте ресурсов
func (n *NotificationSystem) formatResourceAlertMessage(alertType string, usage float64, threshold float64, details map[string]interface{}) string {
	var message strings.Builder

	// Заголовок
	message.WriteString(fmt.Sprintf("<b>Resource Alert: %s</b>\n", alertType))
	message.WriteString(fmt.Sprintf("Timestamp: %s\n\n", time.Now().Format("2006-01-02 15:04:05")))

	// Основная информация
	message.WriteString(fmt.Sprintf("Resource usage exceeded threshold\n"))
	message.WriteString(fmt.Sprintf("• <b>Current Usage:</b> %.1f%%\n", usage))
	message.WriteString(fmt.Sprintf("• <b>Threshold:</b> %.1f%%\n", threshold))
	message.WriteString(fmt.Sprintf("• <b>Excess:</b> %.1f%%\n\n", usage-threshold))

	// Детали
	if len(details) > 0 {
		message.WriteString("<b>Details:</b>\n")
		for key, value := range details {
			message.WriteString(fmt.Sprintf("• <b>%s:</b> %v\n", key, value))
		}
		message.WriteString("\n")
	}

	// Рекомендации
	message.WriteString("<b>Recommendations:</b>\n")
	switch alertType {
	case "Disk Usage":
		message.WriteString("• Clean up temporary files\n")
		message.WriteString("• Archive old logs\n")
		message.WriteString("• Consider disk expansion\n")
		message.WriteString("• Check for large files\n")
	case "CPU Usage":
		message.WriteString("• Check running processes\n")
		message.WriteString("• Optimize algorithms\n")
		message.WriteString("• Consider horizontal scaling\n")
		message.WriteString("• Review resource limits\n")
	case "Memory Usage":
		message.WriteString("• Check memory leaks\n")
		message.WriteString("• Optimize memory usage\n")
		message.WriteString("• Increase memory limits\n")
		message.WriteString("• Review caching strategies\n")
	default:
		message.WriteString("• Monitor resource usage\n")
		message.WriteString("• Check system performance\n")
		message.WriteString("• Review resource allocation\n")
	}

	// Footer
	message.WriteString(fmt.Sprintf("\nAlert ID: %s_%d | Generated by CloudBridge AI Analytics",
		strings.ToLower(strings.ReplaceAll(alertType, " ", "_")), time.Now().Unix()))

	return message.String()
}

// SendTelegramText отправляет произвольный текст в Telegram (для теста)
func (n *NotificationSystem) SendTelegramText(text string) error {
	if n.telegramConfig == nil {
		return fmt.Errorf("telegram config not set")
	}

	payload := map[string]string{
		"chat_id":    n.telegramConfig.ChatID,
		"text":       text,
		"parse_mode": n.telegramConfig.ParseMode,
	}
	if n.telegramConfig.DisableWebPagePreview {
		payload["disable_web_page_preview"] = "true"
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", n.telegramConfig.BotToken)
	fmt.Printf("Telegram: sending TEST message to chat_id=%s url=...%s\n",
		n.telegramConfig.ChatID,
		maskTokenTail(url, 8))
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Telegram: non-200 response (test) status=%d body=%s\n", resp.StatusCode, string(body))
		return fmt.Errorf("telegram API error: status %d", resp.StatusCode)
	}

	fmt.Println("Telegram: test message sent successfully (HTTP 200)")
	return nil
}

// maskTokenTail маскирует секрет, оставляя хвост последних n символов
func maskTokenTail(s string, tail int) string {
	if len(s) <= tail {
		return "***"
	}
	return "***" + s[len(s)-tail:]
}
