package notifications

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

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

// SendReport отправляет отчет через все настроенные каналы
func (n *NotificationSystem) SendReport(report *analytics.Report) error {
	// Slack уведомление
	if n.slackWebhook != "" {
		if err := n.sendSlackNotification(report); err != nil {
			return fmt.Errorf("failed to send Slack notification: %w", err)
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
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API error: status %d", resp.StatusCode)
	}

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
