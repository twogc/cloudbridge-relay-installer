package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"

	"github.com/twogc/cloudbridge-relay-installer/internal/ai/analytics"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai/analytics/collectors"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai/analytics/notifications"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai/analytics/reports"
)

func main() {
	// Инициализация логгера
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Загрузка конфигурации
	config, err := loadConfig()
	if err != nil {
		logger.Fatal("Failed to load config", zap.Error(err))
	}

	// Создание AI Analytics Engine
	engine := analytics.NewAIAnalyticsEngine(config, logger)

	// Добавление сборщиков данных
	if config.DataSources.Prometheus.Enabled {
		prometheusCollector := collectors.NewPrometheusCollector(config.DataSources.Prometheus.URL)
		engine.AddCollector(prometheusCollector)
		logger.Info("Added Prometheus collector", zap.String("url", config.DataSources.Prometheus.URL))
	}

	if config.DataSources.AIService.Enabled {
		aiServiceCollector := collectors.NewAIServiceCollector(config.DataSources.AIService.URL)
		engine.AddCollector(aiServiceCollector)
		logger.Info("Added AI Service collector", zap.String("url", config.DataSources.AIService.URL))
	}

	// Настройка Cloudflare AI клиента
	llmClient := analytics.NewCloudflareAIClient(
		config.CloudflareAI.AccountID,
		config.CloudflareAI.APIToken,
		config.CloudflareAI.Model,
		config.CloudflareAI.MaxTokens,
	)
	engine.SetLLMClient(llmClient)
	logger.Info("Cloudflare AI client configured",
		zap.String("model", config.CloudflareAI.Model),
		zap.String("account_id", config.CloudflareAI.AccountID))

	// Настройка системы уведомлений
	notificationSystem := notifications.NewNotificationSystem()

	// Настройка Slack webhook
	if config.Notifications.Slack.Enabled {
		slackWebhook := os.Getenv("SLACK_WEBHOOK_URL")
		if slackWebhook != "" {
			notificationSystem.AddSlackWebhook(slackWebhook)
			logger.Info("Slack notifications enabled")
		} else {
			logger.Warn("Slack webhook URL not provided, Slack notifications disabled")
		}
	}

	// Инициализация Telegram из переменных окружения (секретов)
	if botToken := os.Getenv("TELEGRAM_BOT_TOKEN"); botToken != "" {
		chatID := os.Getenv("TELEGRAM_CHAT_ID")
		parseMode := os.Getenv("TELEGRAM_PARSE_MODE")
		if parseMode == "" {
			parseMode = "HTML"
		}
		disablePreview := os.Getenv("TELEGRAM_DISABLE_WEB_PAGE_PREVIEW") == "true"
		notificationSystem.AddTelegramConfig(&notifications.TelegramConfig{
			BotToken:              botToken,
			ChatID:                chatID,
			ParseMode:             parseMode,
			DisableWebPagePreview: disablePreview,
		})
		masked := ""
		if len(botToken) > 8 {
			masked = botToken[:4] + "***" + botToken[len(botToken)-4:]
		}
		logger.Info("Telegram notifications enabled",
			zap.String("chat_id", chatID),
			zap.String("bot_token_masked", masked),
			zap.String("parse_mode", parseMode),
			zap.Bool("disable_web_page_preview", disablePreview))
	} else {
		logger.Info("Telegram not configured (TELEGRAM_BOT_TOKEN is empty)")
	}

	engine.SetNotificationSystem(notificationSystem)

	// Настройка генератора отчетов
	reportGenerator := reports.NewReportGenerator()
	engine.SetReportGenerator(reportGenerator)

	// Запуск планировщика
	engine.StartScheduler()

	// Запуск HTTP сервера
	router := gin.Default()
	engine.SetupRoutes(router)

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy", "service": "ai-analytics"})
	})

	// Graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		logger.Info("Shutting down AI Analytics Service...")
		engine.Stop()
		os.Exit(0)
	}()

	port := os.Getenv("AI_ANALYTICS_PORT")
	if port == "" {
		port = "8084"
	}

	logger.Info("AI Analytics Service started", zap.String("port", port))
	if err := router.Run(":" + port); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}

// loadConfig загружает конфигурацию из файла
func loadConfig() (*analytics.AnalyticsConfig, error) {
	configPath := os.Getenv("AI_ANALYTICS_CONFIG_PATH")
	if configPath == "" {
		configPath = "config/ai-analytics.yaml"
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

    // Пытаемся распарсить два возможных формата:
    // 1) С корневым ключом ai_analytics
    // 2) Плоский (поля на верхнем уровне)
    var withRoot struct {
        AIAnalytics analytics.AnalyticsConfig `yaml:"ai_analytics"`
    }
    if err := yaml.Unmarshal(data, &withRoot); err == nil && (withRoot.AIAnalytics.CloudflareAI.AccountID != "" || withRoot.AIAnalytics.Enabled) {
        return &withRoot.AIAnalytics, nil
    }

    var flat analytics.AnalyticsConfig
    if err := yaml.Unmarshal(data, &flat); err != nil {
        return nil, err
    }
    return &flat, nil
}
