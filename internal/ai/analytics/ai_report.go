package analytics

import (
	"time"
)

// AIReport представляет профессиональный отчет системы CloudBridge AI
type AIReport struct {
	Timestamp    time.Time `json:"timestamp"`
	SystemStatus string    `json:"system_status"` // "excellent", "good", "attention", "critical"
	HealthScore  int       `json:"health_score"`  // 0-100

	Analysis struct {
		OverallAssessment string   `json:"overall_assessment"`
		Performance      string   `json:"performance"`
		Issues           []string `json:"issues"`
		Recommendations  []string `json:"recommendations"`
		Forecast         string   `json:"forecast"`
	} `json:"analysis"`

	Metrics struct {
		AnomaliesDetected    int     `json:"anomalies_detected"`
		PredictionsMade      int     `json:"predictions_made"`
		RequestsProcessed    int     `json:"requests_processed"`
		RoutesOptimized      int     `json:"routes_optimized"`
		PredictionAccuracy   float64 `json:"prediction_accuracy"`
		CPUUsage            float64 `json:"cpu_usage"`
		MemoryUsage         float64 `json:"memory_usage"`
		DiskUsage           float64 `json:"disk_usage"`
		ActiveAlerts        int     `json:"active_alerts"`
		ProcessingTime      float64 `json:"processing_time"`
		ServiceUptime       float64 `json:"service_uptime"`
	} `json:"metrics"`
}

// AIMetricsData представляет данные метрик для анализа AI отчетов
type AIMetricsData struct {
	AnomaliesDetected    int     `json:"anomalies_detected"`
	PredictionsMade      int     `json:"predictions_made"`
	RequestsProcessed    int     `json:"requests_processed"`
	RoutesOptimized      int     `json:"routes_optimized"`
	PredictionAccuracy   float64 `json:"prediction_accuracy"`
	CPUUsage            float64 `json:"cpu_usage"`
	MemoryUsage         float64 `json:"memory_usage"`
	DiskUsage           float64 `json:"disk_usage"`
	ActiveAlerts        int     `json:"active_alerts"`
	ProcessingTime      float64 `json:"processing_time"`
	ServiceUptime       float64 `json:"service_uptime"`
}

// PROFESSIONAL_PROMPT - профессиональный промпт для Cloudflare AI
const PROFESSIONAL_PROMPT = `
Ты - эксперт по анализу систем мониторинга CloudBridge AI. 
Проанализируй следующие метрики и создай профессиональный отчет на русском языке.

МЕТРИКИ СИСТЕМЫ:
- Обнаружено аномалий: %d
- Сделано предсказаний: %d  
- Обработано запросов: %d
- Оптимизировано маршрутов: %d
- Точность предсказаний: %.2f
- Использование CPU: %.1f%%
- Использование памяти: %.1f MB
- Использование диска: %.1f GB
- Активные алерты: %d
- Время обработки: %.1f секунд
- Время работы сервиса: %.1f часов

ТРЕБОВАНИЯ К ОТЧЕТУ:
1. ОЦЕНКА СОСТОЯНИЯ: Общая оценка здоровья системы (отлично/хорошо/требует внимания/критично)
2. АНАЛИЗ ПРОИЗВОДИТЕЛЬНОСТИ: Детальный анализ работы AI компонентов
3. ВЫЯВЛЕНИЕ ПРОБЛЕМ: Потенциальные проблемы и их приоритет
4. РЕКОМЕНДАЦИИ: Конкретные действия для оптимизации
5. ПРОГНОЗ: Ожидаемые тренды и рекомендации на будущее

ФОРМАТ ОТЧЕТА:
- Профессиональный стиль без эмодзи
- Структурированный анализ
- Конкретные цифры и проценты
- Практические рекомендации
- Максимум 2000 символов

Создай детальный анализ системы CloudBridge AI.`

// GenerateReportRequest представляет запрос на генерацию отчета
type GenerateReportRequest struct {
	ReportType         string `json:"report_type"`
	Format             string `json:"format"`
	Language           string `json:"language"`
	UseCloudflareAI    bool   `json:"use_cloudflare_ai"`
	IncludeMetrics     bool   `json:"include_metrics"`
	IncludeAnalysis    bool   `json:"include_analysis"`
	IncludeRecommendations bool `json:"include_recommendations"`
	IncludeForecast    bool   `json:"include_forecast"`
	Prompt             string `json:"prompt"`
}
