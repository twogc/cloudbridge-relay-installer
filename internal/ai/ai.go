package ai

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// BehaviorAnalyzer анализирует поведение пользователей и аномалии
type BehaviorAnalyzer struct {
	Enabled       bool
	ModelPath     string
	Threshold     float64
	LearningRate  float64
	UserProfiles  map[string]*UserProfile
	AnomalyScores map[string]float64
	mutex         sync.RWMutex
}

// UserProfile профиль пользователя для анализа поведения
type UserProfile struct {
	UserID          string
	LoginPatterns   []LoginPattern
	AccessPatterns  []AccessPattern
	NetworkPatterns []NetworkPattern
	RiskScore       float64
	LastUpdated     time.Time
}

// LoginPattern паттерн входа пользователя
type LoginPattern struct {
	TimeOfDay  time.Time
	IPAddress  string
	DeviceType string
	Location   string
	Success    bool
}

// AccessPattern паттерн доступа к ресурсам
type AccessPattern struct {
	Resource   string
	TimeOfDay  time.Time
	Duration   time.Duration
	Frequency  int
	DataVolume int64
}

// NetworkPattern сетевой паттерн
type NetworkPattern struct {
	Destination string
	Protocol    string
	Port        int
	DataVolume  int64
	Frequency   int
}

// PredictiveSecurity предиктивная безопасность
type PredictiveSecurity struct {
	Enabled      bool
	ThreatModels map[string]*ThreatModel
	Predictions  map[string]*ThreatPrediction
	Confidence   float64
	mutex        sync.RWMutex
}

// ThreatModel модель угрозы
type ThreatModel struct {
	ThreatID   string
	Type       ThreatType
	Indicators []ThreatIndicator
	RiskLevel  RiskLevel
	Mitigation []MitigationStrategy
}

// ThreatType тип угрозы
type ThreatType int

const (
	ThreatTypeDDoS ThreatType = iota
	ThreatTypeBruteForce
	ThreatTypeDataExfiltration
	ThreatTypePrivilegeEscalation
	ThreatTypeMalware
)

// RiskLevel уровень риска
type RiskLevel int

const (
	RiskLevelLow RiskLevel = iota
	RiskLevelMedium
	RiskLevelHigh
	RiskLevelCritical
)

// ThreatIndicator индикатор угрозы
type ThreatIndicator struct {
	Name      string
	Value     float64
	Weight    float64
	Threshold float64
}

// MitigationStrategy стратегия митигации
type MitigationStrategy struct {
	Name          string
	Action        string
	Priority      int
	Effectiveness float64
}

// ThreatPrediction предикция угрозы
type ThreatPrediction struct {
	ThreatID    string
	Probability float64
	TimeWindow  time.Duration
	Confidence  float64
	Mitigation  []MitigationStrategy
}

// AutomatedResponse автоматическое реагирование
type AutomatedResponse struct {
	Enabled       bool
	ResponseRules map[string]*ResponseRule
	Actions       map[string]*ResponseAction
	History       []ResponseEvent
	mutex         sync.RWMutex
}

// ResponseRule правило автоматического реагирования
type ResponseRule struct {
	RuleID     string
	Name       string
	Conditions []ResponseCondition
	Actions    []string
	Priority   int
	Enabled    bool
}

// ResponseCondition условие для автоматического реагирования
type ResponseCondition struct {
	Type     ConditionType
	Field    string
	Operator string
	Value    interface{}
	Duration time.Duration
}

// ConditionType тип условия
type ConditionType int

const (
	ConditionTypeThreshold ConditionType = iota
	ConditionTypeAnomaly
	ConditionTypePattern
	ConditionTypeThreat
)

// ResponseAction действие автоматического реагирования
type ResponseAction struct {
	ActionID   string
	Name       string
	Type       ActionType
	Parameters map[string]interface{}
	Timeout    time.Duration
}

// ActionType тип действия
type ActionType int

const (
	ActionTypeBlock ActionType = iota
	ActionTypeAlert
	ActionTypeQuarantine
	ActionTypeRateLimit
	ActionTypeLog
)

// ResponseEvent событие автоматического реагирования
type ResponseEvent struct {
	EventID   string
	RuleID    string
	ActionID  string
	Timestamp time.Time
	Success   bool
	Details   map[string]interface{}
}

// AIMonitoring ИИ-мониторинг
type AIMonitoring struct {
	Enabled   bool
	Metrics   map[string]*AIMetric
	Alerts    []AIAlert
	Dashboard *AIDashboard
	mutex     sync.RWMutex
}

// AIMetric ИИ-метрика
type AIMetric struct {
	Name        string
	Value       float64
	Unit        string
	Trend       TrendDirection
	Threshold   float64
	LastUpdated time.Time
}

// TrendDirection направление тренда
type TrendDirection int

const (
	TrendStable TrendDirection = iota
	TrendUp
	TrendDown
	TrendSpike
)

// AIAlert ИИ-алерт
type AIAlert struct {
	AlertID      string
	Type         AlertType
	Severity     AlertSeverity
	Message      string
	Timestamp    time.Time
	Acknowledged bool
	Resolved     bool
}

// AlertType тип алерта
type AlertType int

const (
	AlertTypeAnomaly AlertType = iota
	AlertTypeThreat
	AlertTypePerformance
	AlertTypeSecurity
)

// AlertSeverity серьезность алерта
type AlertSeverity int

const (
	AlertSeverityInfo AlertSeverity = iota
	AlertSeverityWarning
	AlertSeverityError
	AlertSeverityCritical
)

// AIDashboard ИИ-дашборд
type AIDashboard struct {
	Title       string
	Widgets     []DashboardWidget
	RefreshRate time.Duration
}

// DashboardWidget виджет дашборда
type DashboardWidget struct {
	WidgetID string
	Type     WidgetType
	Title    string
	Data     interface{}
	Position WidgetPosition
}

// WidgetType тип виджета
type WidgetType int

const (
	WidgetTypeChart WidgetType = iota
	WidgetTypeTable
	WidgetTypeGauge
	WidgetTypeAlert
)

// WidgetPosition позиция виджета
type WidgetPosition struct {
	X      int
	Y      int
	Width  int
	Height int
}

// NewBehaviorAnalyzer создает новый анализатор поведения
func NewBehaviorAnalyzer() *BehaviorAnalyzer {
	return &BehaviorAnalyzer{
		Enabled:       true,
		ModelPath:     "/models/behavior",
		Threshold:     0.8,
		LearningRate:  0.01,
		UserProfiles:  make(map[string]*UserProfile),
		AnomalyScores: make(map[string]float64),
	}
}

// NewPredictiveSecurity создает новую предиктивную безопасность
func NewPredictiveSecurity() *PredictiveSecurity {
	return &PredictiveSecurity{
		Enabled:      true,
		ThreatModels: make(map[string]*ThreatModel),
		Predictions:  make(map[string]*ThreatPrediction),
		Confidence:   0.85,
	}
}

// NewAutomatedResponse создает новое автоматическое реагирование
func NewAutomatedResponse() *AutomatedResponse {
	return &AutomatedResponse{
		Enabled:       true,
		ResponseRules: make(map[string]*ResponseRule),
		Actions:       make(map[string]*ResponseAction),
		History:       make([]ResponseEvent, 0),
	}
}

// NewAIMonitoring создает новый ИИ-мониторинг
func NewAIMonitoring() *AIMonitoring {
	return &AIMonitoring{
		Enabled: true,
		Metrics: make(map[string]*AIMetric),
		Alerts:  make([]AIAlert, 0),
		Dashboard: &AIDashboard{
			Title:       "AI Security Dashboard",
			Widgets:     make([]DashboardWidget, 0),
			RefreshRate: 30 * time.Second,
		},
	}
}

// BehaviorAnalyzer методы
func (ba *BehaviorAnalyzer) Start() error {
	if !ba.Enabled {
		return fmt.Errorf("behavior analyzer is disabled")
	}

	// Инициализируем базовые профили пользователей
	ba.initializeDefaultProfiles()

	return nil
}

func (ba *BehaviorAnalyzer) Stop() error {
	// Сохраняем профили пользователей
	ba.saveUserProfiles()

	// Очищаем временные данные
	ba.AnomalyScores = make(map[string]float64)

	return nil
}

func (ba *BehaviorAnalyzer) AnalyzeUserBehavior(userID string, event interface{}) error {
	if userID == "" {
		return fmt.Errorf("user ID cannot be empty")
	}
	
	if event == nil {
		return fmt.Errorf("event cannot be nil")
	}
	
	ba.mutex.Lock()
	defer ba.mutex.Unlock()
	
	// Создаем профиль пользователя если его нет
	if ba.UserProfiles[userID] == nil {
		ba.UserProfiles[userID] = &UserProfile{
			UserID:          userID,
			LoginPatterns:   make([]LoginPattern, 0),
			AccessPatterns:  make([]AccessPattern, 0),
			NetworkPatterns: make([]NetworkPattern, 0),
			RiskScore:       0.0,
			LastUpdated:     time.Now(),
		}
	}

	// Анализируем событие и обновляем профиль
	profile := ba.UserProfiles[userID]
	profile.LastUpdated = time.Now()
	
	// Анализируем тип события
	if eventMap, ok := event.(map[string]interface{}); ok {
		ba.analyzeEvent(profile, eventMap)
	}
	
	return nil
}

func (ba *BehaviorAnalyzer) DetectAnomaly(userID string) (float64, error) {
	if userID == "" {
		return 0, fmt.Errorf("user ID cannot be empty")
	}
	
	ba.mutex.RLock()
	profile, exists := ba.UserProfiles[userID]
	ba.mutex.RUnlock()
	
	if !exists {
		return 0, fmt.Errorf("user profile not found for user: %s", userID)
	}
	
	// Вычисляем скор аномалии на основе паттернов поведения
	score := ba.calculateAnomalyScore(profile)
	
	// Применяем пороговое значение
	if score > ba.Threshold {
		score = 1.0 // Максимальный риск
	} else {
		score = score / ba.Threshold // Нормализуем к диапазону [0, 1]
	}

	ba.mutex.Lock()
	ba.AnomalyScores[userID] = score
	ba.mutex.Unlock()
	
	return score, nil
}

func (ba *BehaviorAnalyzer) UpdateUserProfile(userID string, event interface{}) error {
	if userID == "" {
		return fmt.Errorf("user ID cannot be empty")
	}
	
	ba.mutex.RLock()
	profile, exists := ba.UserProfiles[userID]
	ba.mutex.RUnlock()
	
	if !exists {
		return fmt.Errorf("user profile not found for user: %s", userID)
	}
	
	// Обновляем временную метку
		profile.LastUpdated = time.Now()
	
	// Обновляем риск на основе события
	if event != nil {
		ba.updateRiskScore(profile, event)
	}
	
	return nil
}

// initializeDefaultProfiles инициализирует базовые профили пользователей
func (ba *BehaviorAnalyzer) initializeDefaultProfiles() {
	// Создаем базовые профили для системных пользователей
	systemUsers := []string{"system", "admin", "monitor"}

	for _, userID := range systemUsers {
		ba.UserProfiles[userID] = &UserProfile{
			UserID:          userID,
			LoginPatterns:   make([]LoginPattern, 0),
			AccessPatterns:  make([]AccessPattern, 0),
			NetworkPatterns: make([]NetworkPattern, 0),
			RiskScore:       0.0,
			LastUpdated:     time.Now(),
		}
	}
}

// saveUserProfiles сохраняет профили пользователей
func (ba *BehaviorAnalyzer) saveUserProfiles() {
	// В реальной реализации здесь было бы сохранение в базу данных
	// Пока что просто логируем количество профилей
	// log.Printf("Saving %d user profiles", len(ba.UserProfiles))
}

// analyzeEvent анализирует событие и обновляет профиль
func (ba *BehaviorAnalyzer) analyzeEvent(profile *UserProfile, event map[string]interface{}) {
	eventType, ok := event["type"].(string)
	if !ok {
		return
	}

	switch eventType {
	case "login":
		ba.analyzeLoginEvent(profile, event)
	case "access":
		ba.analyzeAccessEvent(profile, event)
	case "network":
		ba.analyzeNetworkEvent(profile, event)
	default:
		// Обновляем общий риск для неизвестных событий
		profile.RiskScore += 0.01
	}
}

// analyzeLoginEvent анализирует событие входа
func (ba *BehaviorAnalyzer) analyzeLoginEvent(profile *UserProfile, event map[string]interface{}) {
	pattern := LoginPattern{
		TimeOfDay:  time.Now(),
		IPAddress:  getStringFromEvent(event, "ip_address"),
		DeviceType: getStringFromEvent(event, "device_type"),
		Location:   getStringFromEvent(event, "location"),
		Success:    getBoolFromEvent(event, "success"),
	}

	profile.LoginPatterns = append(profile.LoginPatterns, pattern)

	// Ограничиваем количество паттернов
	if len(profile.LoginPatterns) > 100 {
		profile.LoginPatterns = profile.LoginPatterns[1:]
	}

	// Анализируем аномалии входа
	if !pattern.Success {
		profile.RiskScore += 0.1 // Неудачный вход увеличивает риск
	}
}

// analyzeAccessEvent анализирует событие доступа
func (ba *BehaviorAnalyzer) analyzeAccessEvent(profile *UserProfile, event map[string]interface{}) {
	pattern := AccessPattern{
		Resource:   getStringFromEvent(event, "resource"),
		TimeOfDay:  time.Now(),
		Duration:   getDurationFromEvent(event, "duration"),
		Frequency:  getIntFromEvent(event, "frequency"),
		DataVolume: getInt64FromEvent(event, "data_volume"),
	}

	profile.AccessPatterns = append(profile.AccessPatterns, pattern)

	// Ограничиваем количество паттернов
	if len(profile.AccessPatterns) > 100 {
		profile.AccessPatterns = profile.AccessPatterns[1:]
	}
}

// analyzeNetworkEvent анализирует сетевые события
func (ba *BehaviorAnalyzer) analyzeNetworkEvent(profile *UserProfile, event map[string]interface{}) {
	pattern := NetworkPattern{
		Destination: getStringFromEvent(event, "destination"),
		Protocol:    getStringFromEvent(event, "protocol"),
		Port:        getIntFromEvent(event, "port"),
		DataVolume:  getInt64FromEvent(event, "data_volume"),
		Frequency:   getIntFromEvent(event, "frequency"),
	}

	profile.NetworkPatterns = append(profile.NetworkPatterns, pattern)

	// Ограничиваем количество паттернов
	if len(profile.NetworkPatterns) > 100 {
		profile.NetworkPatterns = profile.NetworkPatterns[1:]
	}
}

// calculateAnomalyScore вычисляет скор аномалии для профиля
func (ba *BehaviorAnalyzer) calculateAnomalyScore(profile *UserProfile) float64 {
	score := profile.RiskScore

	// Анализируем паттерны входа
	if len(profile.LoginPatterns) > 0 {
		score += ba.analyzeLoginPatterns(profile.LoginPatterns)
	}

	// Анализируем паттерны доступа
	if len(profile.AccessPatterns) > 0 {
		score += ba.analyzeAccessPatterns(profile.AccessPatterns)
	}

	// Анализируем сетевые паттерны
	if len(profile.NetworkPatterns) > 0 {
		score += ba.analyzeNetworkPatterns(profile.NetworkPatterns)
	}

	return score
}

// analyzeLoginPatterns анализирует паттерны входа
func (ba *BehaviorAnalyzer) analyzeLoginPatterns(patterns []LoginPattern) float64 {
	if len(patterns) == 0 {
		return 0
	}

	// Подсчитываем неудачные входы
	failedLogins := 0
	for _, pattern := range patterns {
		if !pattern.Success {
			failedLogins++
		}
	}

	// Возвращаем долю неудачных входов
	return float64(failedLogins) / float64(len(patterns))
}

// analyzeAccessPatterns анализирует паттерны доступа
func (ba *BehaviorAnalyzer) analyzeAccessPatterns(patterns []AccessPattern) float64 {
	if len(patterns) == 0 {
		return 0
	}

	// Анализируем частоту доступа
	totalFrequency := 0
	for _, pattern := range patterns {
		totalFrequency += pattern.Frequency
	}

	// Высокая частота доступа может указывать на аномалию
	avgFrequency := float64(totalFrequency) / float64(len(patterns))
	if avgFrequency > 100 { // Пороговое значение
		return 0.2
	}

	return 0
}

// analyzeNetworkPatterns анализирует сетевые паттерны
func (ba *BehaviorAnalyzer) analyzeNetworkPatterns(patterns []NetworkPattern) float64 {
	if len(patterns) == 0 {
		return 0
	}

	// Анализируем объем передаваемых данных
	totalDataVolume := int64(0)
	for _, pattern := range patterns {
		totalDataVolume += pattern.DataVolume
	}

	// Большой объем данных может указывать на аномалию
	avgDataVolume := float64(totalDataVolume) / float64(len(patterns))
	if avgDataVolume > 1000000 { // 1MB пороговое значение
		return 0.3
	}

	return 0
}

// updateRiskScore обновляет скор риска на основе события
func (ba *BehaviorAnalyzer) updateRiskScore(profile *UserProfile, event interface{}) {
	// Базовое обновление риска
	profile.RiskScore += 0.01

	// Ограничиваем максимальный риск
	if profile.RiskScore > 1.0 {
		profile.RiskScore = 1.0
	}

	// Применяем learning rate для адаптации
	profile.RiskScore *= (1.0 - ba.LearningRate)
}

// Вспомогательные функции для извлечения данных из событий
func getStringFromEvent(event map[string]interface{}, key string) string {
	if val, ok := event[key].(string); ok {
		return val
	}
	return ""
}

func getBoolFromEvent(event map[string]interface{}, key string) bool {
	if val, ok := event[key].(bool); ok {
		return val
	}
	return false
}

func getIntFromEvent(event map[string]interface{}, key string) int {
	if val, ok := event[key].(int); ok {
		return val
	}
	return 0
}

func getInt64FromEvent(event map[string]interface{}, key string) int64 {
	if val, ok := event[key].(int64); ok {
		return val
	}
	return 0
}

func getDurationFromEvent(event map[string]interface{}, key string) time.Duration {
	if val, ok := event[key].(time.Duration); ok {
		return val
	}
	return 0
}

// PredictiveSecurity методы
func (ps *PredictiveSecurity) Start() error {
	if !ps.Enabled {
		return fmt.Errorf("predictive security is disabled")
	}

	// Инициализируем базовые модели угроз
	ps.initializeDefaultThreatModels()

	return nil
}

func (ps *PredictiveSecurity) Stop() error {
	// Сохраняем модели угроз
	ps.saveThreatModels()

	// Очищаем предсказания
	ps.Predictions = make(map[string]*ThreatPrediction)

	return nil
}

func (ps *PredictiveSecurity) PredictThreat(indicators []ThreatIndicator) (*ThreatPrediction, error) {
	if len(indicators) == 0 {
		return nil, fmt.Errorf("no indicators provided for threat prediction")
	}
	
	// Анализируем индикаторы и вычисляем вероятность угрозы
	probability := ps.calculateThreatProbability(indicators)
	
	// Определяем тип угрозы на основе индикаторов
	threatType := ps.determineThreatType(indicators)
	
	// Создаем предсказание
	threatID := fmt.Sprintf("predicted-threat-%d", time.Now().Unix())
	prediction := &ThreatPrediction{
		ThreatID:    threatID,
		Probability: probability,
		TimeWindow:  ps.calculateTimeWindow(probability),
		Confidence:  ps.calculateConfidence(indicators),
		Mitigation:  ps.getMitigationStrategies(threatType),
	}

	ps.mutex.Lock()
	ps.Predictions[threatID] = prediction
	ps.mutex.Unlock()
	
	return prediction, nil
}

func (ps *PredictiveSecurity) AddThreatModel(model *ThreatModel) error {
	if model == nil {
		return fmt.Errorf("threat model cannot be nil")
	}
	
	if model.ThreatID == "" {
		return fmt.Errorf("threat model ID cannot be empty")
	}
	
	// Валидируем модель
	if err := ps.validateThreatModel(model); err != nil {
		return fmt.Errorf("invalid threat model: %w", err)
	}
	
	ps.mutex.Lock()
	ps.ThreatModels[model.ThreatID] = model
	ps.mutex.Unlock()
	
	return nil
}

func (ps *PredictiveSecurity) GetThreatPredictions() map[string]*ThreatPrediction {
	// Возвращаем копию предсказаний для безопасности
	ps.mutex.RLock()
	predictions := make(map[string]*ThreatPrediction)
	for id, prediction := range ps.Predictions {
		predictions[id] = prediction
	}
	ps.mutex.RUnlock()
	return predictions
}

// initializeDefaultThreatModels инициализирует базовые модели угроз
func (ps *PredictiveSecurity) initializeDefaultThreatModels() {
	// DDoS модель
	ddosModel := &ThreatModel{
		ThreatID: "ddos",
		Type:     ThreatTypeDDoS,
		Indicators: []ThreatIndicator{
			{Name: "packet_rate", Value: 0, Weight: 0.4, Threshold: 1000},
			{Name: "connection_count", Value: 0, Weight: 0.3, Threshold: 500},
			{Name: "bandwidth_usage", Value: 0, Weight: 0.3, Threshold: 0.8},
		},
		RiskLevel: RiskLevelHigh,
		Mitigation: []MitigationStrategy{
			{Name: "rate_limiting", Action: "block", Priority: 1, Effectiveness: 0.8},
			{Name: "traffic_filtering", Action: "filter", Priority: 2, Effectiveness: 0.7},
		},
	}
	ps.ThreatModels["ddos"] = ddosModel

	// Brute Force модель
	bruteForceModel := &ThreatModel{
		ThreatID: "brute_force",
		Type:     ThreatTypeBruteForce,
		Indicators: []ThreatIndicator{
			{Name: "failed_logins", Value: 0, Weight: 0.5, Threshold: 10},
			{Name: "login_frequency", Value: 0, Weight: 0.3, Threshold: 5},
			{Name: "ip_reputation", Value: 0, Weight: 0.2, Threshold: 0.3},
		},
		RiskLevel: RiskLevelMedium,
		Mitigation: []MitigationStrategy{
			{Name: "account_lockout", Action: "lock", Priority: 1, Effectiveness: 0.9},
			{Name: "ip_blocking", Action: "block", Priority: 2, Effectiveness: 0.8},
		},
	}
	ps.ThreatModels["brute_force"] = bruteForceModel
}

// saveThreatModels сохраняет модели угроз
func (ps *PredictiveSecurity) saveThreatModels() {
	// В реальной реализации здесь было бы сохранение в базу данных
	// log.Printf("Saving %d threat models", len(ps.ThreatModels))
}

// calculateThreatProbability вычисляет вероятность угрозы
func (ps *PredictiveSecurity) calculateThreatProbability(indicators []ThreatIndicator) float64 {
	totalWeight := 0.0
	weightedScore := 0.0

	for _, indicator := range indicators {
		// Нормализуем значение индикатора относительно порога
		normalizedValue := indicator.Value / indicator.Threshold
		if normalizedValue > 1.0 {
			normalizedValue = 1.0
		}

		weightedScore += normalizedValue * indicator.Weight
		totalWeight += indicator.Weight
	}

	if totalWeight == 0 {
		return 0.0
	}

	probability := weightedScore / totalWeight
	if probability > 1.0 {
		probability = 1.0
	}

	return probability
}

// determineThreatType определяет тип угрозы
func (ps *PredictiveSecurity) determineThreatType(indicators []ThreatIndicator) ThreatType {
	// Анализируем индикаторы для определения типа угрозы
	for _, indicator := range indicators {
		switch indicator.Name {
		case "packet_rate", "connection_count", "bandwidth_usage":
			return ThreatTypeDDoS
		case "failed_logins", "login_frequency":
			return ThreatTypeBruteForce
		case "data_exfiltration", "unusual_data_access":
			return ThreatTypeDataExfiltration
		case "privilege_escalation", "admin_access":
			return ThreatTypePrivilegeEscalation
		case "malware_signature", "suspicious_behavior":
			return ThreatTypeMalware
		}
	}

	return ThreatTypeDDoS // По умолчанию
}

// calculateTimeWindow вычисляет временное окно для угрозы
func (ps *PredictiveSecurity) calculateTimeWindow(probability float64) time.Duration {
	// Чем выше вероятность, тем меньше временное окно
	if probability > 0.8 {
		return 15 * time.Minute
	} else if probability > 0.6 {
		return 30 * time.Minute
	} else if probability > 0.4 {
		return 1 * time.Hour
	} else {
		return 2 * time.Hour
	}
}

// calculateConfidence вычисляет уверенность в предсказании
func (ps *PredictiveSecurity) calculateConfidence(indicators []ThreatIndicator) float64 {
	// Базовая уверенность
	confidence := ps.Confidence

	// Корректируем на основе количества индикаторов
	indicatorCount := len(indicators)
	if indicatorCount < 2 {
		confidence *= 0.5 // Низкая уверенность при малом количестве индикаторов
	} else if indicatorCount > 5 {
		confidence *= 1.2 // Высокая уверенность при большом количестве индикаторов
	}

	// Ограничиваем максимальную уверенность
	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// getMitigationStrategies возвращает стратегии митигации для типа угрозы
func (ps *PredictiveSecurity) getMitigationStrategies(threatType ThreatType) []MitigationStrategy {
	// Ищем модель угрозы для данного типа
	for _, model := range ps.ThreatModels {
		if model.Type == threatType {
			return model.Mitigation
		}
	}

	// Возвращаем общие стратегии митигации
	return []MitigationStrategy{
		{Name: "general_monitoring", Action: "monitor", Priority: 1, Effectiveness: 0.5},
		{Name: "alert_notification", Action: "alert", Priority: 2, Effectiveness: 0.7},
	}
}

// validateThreatModel валидирует модель угрозы
func (ps *PredictiveSecurity) validateThreatModel(model *ThreatModel) error {
	if model.Type < ThreatTypeDDoS || model.Type > ThreatTypeMalware {
		return fmt.Errorf("invalid threat type: %d", model.Type)
	}

	if model.RiskLevel < RiskLevelLow || model.RiskLevel > RiskLevelCritical {
		return fmt.Errorf("invalid risk level: %d", model.RiskLevel)
	}

	if len(model.Indicators) == 0 {
		return fmt.Errorf("threat model must have at least one indicator")
	}

	// Валидируем индикаторы
	for i, indicator := range model.Indicators {
		if indicator.Name == "" {
			return fmt.Errorf("indicator %d: name cannot be empty", i)
		}
		if indicator.Weight < 0 || indicator.Weight > 1 {
			return fmt.Errorf("indicator %d: weight must be between 0 and 1", i)
		}
		if indicator.Threshold <= 0 {
			return fmt.Errorf("indicator %d: threshold must be positive", i)
		}
	}

	return nil
}

// AutomatedResponse методы
func (ar *AutomatedResponse) Start() error {
	if !ar.Enabled {
		return fmt.Errorf("automated response is disabled")
	}

	// Инициализируем базовые правила реагирования
	ar.initializeDefaultResponseRules()

	return nil
}

func (ar *AutomatedResponse) Stop() error {
	// Сохраняем историю реагирования
	ar.saveResponseHistory()

	// Очищаем активные правила
	ar.ResponseRules = make(map[string]*ResponseRule)

	return nil
}

func (ar *AutomatedResponse) AddResponseRule(rule *ResponseRule) error {
	if rule == nil {
		return fmt.Errorf("response rule cannot be nil")
	}
	
	if rule.RuleID == "" {
		return fmt.Errorf("response rule ID cannot be empty")
	}
	
	// Валидируем правило
	if err := ar.validateResponseRule(rule); err != nil {
		return fmt.Errorf("invalid response rule: %w", err)
	}
	
	ar.mutex.Lock()
	ar.ResponseRules[rule.RuleID] = rule
	ar.mutex.Unlock()
	
	return nil
}

func (ar *AutomatedResponse) TriggerResponse(event interface{}) error {
	if event == nil {
		return fmt.Errorf("event cannot be nil")
	}
	
	// Анализируем событие и находим подходящие правила
	triggeredRules := ar.findTriggeredRules(event)
	
	if len(triggeredRules) == 0 {
		return nil // Нет правил для срабатывания
	}
	
	// Выполняем действия для каждого сработавшего правила
	ar.mutex.Lock()
	for _, rule := range triggeredRules {
		responseEvent := ar.executeResponseRule(rule, event)
	ar.History = append(ar.History, responseEvent)
	}
	ar.mutex.Unlock()
	
	return nil
}

func (ar *AutomatedResponse) GetResponseHistory() []ResponseEvent {
	// Возвращаем копию истории для безопасности
	ar.mutex.RLock()
	history := make([]ResponseEvent, len(ar.History))
	copy(history, ar.History)
	ar.mutex.RUnlock()
	return history
}

// initializeDefaultResponseRules инициализирует базовые правила реагирования
func (ar *AutomatedResponse) initializeDefaultResponseRules() {
	// Правило для высокого риска
	highRiskRule := &ResponseRule{
		RuleID: "high_risk",
		Name:   "High Risk Response",
		Conditions: []ResponseCondition{
			{
				Type:     ConditionTypeThreshold,
				Field:    "risk_score",
				Operator: ">",
				Value:    0.8,
			},
		},
		Actions:  []string{"block_user", "send_alert"},
		Priority: 1,
		Enabled:  true,
	}
	ar.ResponseRules["high_risk"] = highRiskRule

	// Правило для аномалий
	anomalyRule := &ResponseRule{
		RuleID: "anomaly_detected",
		Name:   "Anomaly Response",
		Conditions: []ResponseCondition{
			{
				Type:     ConditionTypeAnomaly,
				Field:    "anomaly_score",
				Operator: ">",
				Value:    0.7,
			},
		},
		Actions:  []string{"rate_limit", "monitor"},
		Priority: 2,
		Enabled:  true,
	}
	ar.ResponseRules["anomaly_detected"] = anomalyRule
}

// saveResponseHistory сохраняет историю реагирования
func (ar *AutomatedResponse) saveResponseHistory() {
	// В реальной реализации здесь было бы сохранение в базу данных
	// log.Printf("Saving %d response events", len(ar.History))
}

// findTriggeredRules находит правила, которые должны сработать для события
func (ar *AutomatedResponse) findTriggeredRules(event interface{}) []*ResponseRule {
	var triggeredRules []*ResponseRule
	
	ar.mutex.RLock()
	for _, rule := range ar.ResponseRules {
		if !rule.Enabled {
			continue
		}
		
		if ar.evaluateRuleConditions(rule, event) {
			triggeredRules = append(triggeredRules, rule)
		}
	}
	ar.mutex.RUnlock()
	
	// Сортируем по приоритету (меньший номер = выше приоритет)
	for i := 0; i < len(triggeredRules)-1; i++ {
		for j := i + 1; j < len(triggeredRules); j++ {
			if triggeredRules[i].Priority > triggeredRules[j].Priority {
				triggeredRules[i], triggeredRules[j] = triggeredRules[j], triggeredRules[i]
			}
		}
	}
	
	return triggeredRules
}

// evaluateRuleConditions оценивает условия правила
func (ar *AutomatedResponse) evaluateRuleConditions(rule *ResponseRule, event interface{}) bool {
	if len(rule.Conditions) == 0 {
		return true // Правило без условий всегда срабатывает
	}

	// Все условия должны быть выполнены (AND логика)
	for _, condition := range rule.Conditions {
		if !ar.evaluateCondition(condition, event) {
			return false
		}
	}

	return true
}

// evaluateCondition оценивает одно условие
func (ar *AutomatedResponse) evaluateCondition(condition ResponseCondition, event interface{}) bool {
	eventMap, ok := event.(map[string]interface{})
	if !ok {
		return false
	}

	fieldValue, exists := eventMap[condition.Field]
	if !exists {
		return false
	}

	switch condition.Type {
	case ConditionTypeThreshold:
		return ar.evaluateThresholdCondition(fieldValue, condition.Operator, condition.Value)
	case ConditionTypeAnomaly:
		return ar.evaluateAnomalyCondition(fieldValue, condition.Operator, condition.Value)
	case ConditionTypePattern:
		return ar.evaluatePatternCondition(fieldValue, condition.Operator, condition.Value)
	case ConditionTypeThreat:
		return ar.evaluateThreatCondition(fieldValue, condition.Operator, condition.Value)
	default:
		return false
	}
}

// evaluateThresholdCondition оценивает пороговое условие
func (ar *AutomatedResponse) evaluateThresholdCondition(fieldValue interface{}, operator string, threshold interface{}) bool {
	fieldFloat, ok1 := fieldValue.(float64)
	thresholdFloat, ok2 := threshold.(float64)

	if !ok1 || !ok2 {
		return false
	}

	switch operator {
	case ">":
		return fieldFloat > thresholdFloat
	case ">=":
		return fieldFloat >= thresholdFloat
	case "<":
		return fieldFloat < thresholdFloat
	case "<=":
		return fieldFloat <= thresholdFloat
	case "==":
		return fieldFloat == thresholdFloat
	case "!=":
		return fieldFloat != thresholdFloat
	default:
		return false
	}
}

// evaluateAnomalyCondition оценивает условие аномалии
func (ar *AutomatedResponse) evaluateAnomalyCondition(fieldValue interface{}, operator string, threshold interface{}) bool {
	// Аномалии оцениваются как пороговые условия
	return ar.evaluateThresholdCondition(fieldValue, operator, threshold)
}

// evaluatePatternCondition оценивает условие паттерна
func (ar *AutomatedResponse) evaluatePatternCondition(fieldValue interface{}, operator string, pattern interface{}) bool {
	fieldStr, ok1 := fieldValue.(string)
	patternStr, ok2 := pattern.(string)

	if !ok1 || !ok2 {
		return false
	}

	switch operator {
	case "contains":
		return strings.Contains(fieldStr, patternStr)
	case "starts_with":
		return strings.HasPrefix(fieldStr, patternStr)
	case "ends_with":
		return strings.HasSuffix(fieldStr, patternStr)
	case "equals":
		return fieldStr == patternStr
	default:
		return false
	}
}

// evaluateThreatCondition оценивает условие угрозы
func (ar *AutomatedResponse) evaluateThreatCondition(fieldValue interface{}, operator string, threat interface{}) bool {
	// Угрозы оцениваются как пороговые условия
	return ar.evaluateThresholdCondition(fieldValue, operator, threat)
}

// executeResponseRule выполняет правило реагирования
func (ar *AutomatedResponse) executeResponseRule(rule *ResponseRule, event interface{}) ResponseEvent {
	eventID := fmt.Sprintf("response-%d", time.Now().Unix())

	// Выполняем действия
	success := true
	details := make(map[string]interface{})

	for _, action := range rule.Actions {
		actionSuccess := ar.executeAction(action, event)
		details[action] = actionSuccess
		if !actionSuccess {
			success = false
		}
	}

	return ResponseEvent{
		EventID:   eventID,
		RuleID:    rule.RuleID,
		ActionID:  strings.Join(rule.Actions, ","),
		Timestamp: time.Now(),
		Success:   success,
		Details:   details,
	}
}

// executeAction выполняет одно действие
func (ar *AutomatedResponse) executeAction(action string, event interface{}) bool {
	// В реальной реализации здесь были бы реальные действия
	// Пока что просто логируем и возвращаем успех
	switch action {
	case "block_user":
		// Блокируем пользователя
		return true
	case "send_alert":
		// Отправляем алерт
		return true
	case "rate_limit":
		// Ограничиваем скорость
		return true
	case "monitor":
		// Усиливаем мониторинг
		return true
	case "quarantine":
		// Помещаем в карантин
		return true
	default:
		// Неизвестное действие
		return false
	}
}

// validateResponseRule валидирует правило реагирования
func (ar *AutomatedResponse) validateResponseRule(rule *ResponseRule) error {
	if rule.Name == "" {
		return fmt.Errorf("rule name cannot be empty")
	}

	if len(rule.Conditions) == 0 {
		return fmt.Errorf("rule must have at least one condition")
	}

	if len(rule.Actions) == 0 {
		return fmt.Errorf("rule must have at least one action")
	}

	if rule.Priority < 1 {
		return fmt.Errorf("rule priority must be positive")
	}

	// Валидируем условия
	for i, condition := range rule.Conditions {
		if err := ar.validateCondition(condition); err != nil {
			return fmt.Errorf("condition %d: %w", i, err)
		}
	}

	return nil
}

// validateCondition валидирует условие
func (ar *AutomatedResponse) validateCondition(condition ResponseCondition) error {
	if condition.Field == "" {
		return fmt.Errorf("field cannot be empty")
	}

	if condition.Operator == "" {
		return fmt.Errorf("operator cannot be empty")
	}

	validOperators := []string{">", ">=", "<", "<=", "==", "!=", "contains", "starts_with", "ends_with", "equals"}
	valid := false
	for _, op := range validOperators {
		if condition.Operator == op {
			valid = true
			break
		}
	}

	if !valid {
		return fmt.Errorf("invalid operator: %s", condition.Operator)
	}

	return nil
}

// AIMonitoring методы
func (aim *AIMonitoring) Start() error {
	if !aim.Enabled {
		return fmt.Errorf("AI monitoring is disabled")
	}

	// Инициализируем базовые метрики
	aim.initializeDefaultMetrics()

	return nil
}

func (aim *AIMonitoring) Stop() error {
	// Сохраняем метрики
	aim.saveMetrics()

	// Очищаем алерты
	aim.Alerts = make([]AIAlert, 0)

	return nil
}

func (aim *AIMonitoring) AddMetric(metric *AIMetric) error {
	if metric == nil {
		return fmt.Errorf("metric cannot be nil")
	}
	
	if metric.Name == "" {
		return fmt.Errorf("metric name cannot be empty")
	}
	
	// Валидируем метрику
	if err := aim.validateMetric(metric); err != nil {
		return fmt.Errorf("invalid metric: %w", err)
	}
	
	aim.mutex.Lock()
	aim.Metrics[metric.Name] = metric
	aim.mutex.Unlock()
	
	return nil
}

func (aim *AIMonitoring) CreateAlert(alert *AIAlert) error {
	if alert == nil {
		return fmt.Errorf("alert cannot be nil")
	}
	
	if alert.AlertID == "" {
		return fmt.Errorf("alert ID cannot be empty")
	}
	
	// Валидируем алерт
	if err := aim.validateAlert(alert); err != nil {
		return fmt.Errorf("invalid alert: %w", err)
	}
	
	aim.mutex.Lock()
	// Проверяем, не существует ли уже алерт с таким ID
	for _, existingAlert := range aim.Alerts {
		if existingAlert.AlertID == alert.AlertID {
			aim.mutex.Unlock()
			return fmt.Errorf("alert with ID %s already exists", alert.AlertID)
		}
	}
	
	aim.Alerts = append(aim.Alerts, *alert)
	aim.mutex.Unlock()
	
	return nil
}

func (aim *AIMonitoring) GetDashboard() *AIDashboard {
	aim.mutex.Lock()
	if aim.Dashboard == nil {
		aim.Dashboard = &AIDashboard{
			Title:       "AI Security Dashboard",
			Widgets:     make([]DashboardWidget, 0),
			RefreshRate: 30 * time.Second,
		}
	}
	aim.mutex.Unlock()
	
	// Обновляем виджеты на основе текущих метрик и алертов
	aim.updateDashboardWidgets()
	
	return aim.Dashboard
}

// initializeDefaultMetrics инициализирует базовые метрики
func (aim *AIMonitoring) initializeDefaultMetrics() {
	// Метрика аномалий
	anomalyMetric := &AIMetric{
		Name:        "anomaly_score",
		Value:       0.0,
		Unit:        "score",
		Trend:       TrendStable,
		Threshold:   0.8,
		LastUpdated: time.Now(),
	}
	aim.Metrics["anomaly_score"] = anomalyMetric

	// Метрика риска
	riskMetric := &AIMetric{
		Name:        "risk_score",
		Value:       0.0,
		Unit:        "score",
		Trend:       TrendStable,
		Threshold:   0.7,
		LastUpdated: time.Now(),
	}
	aim.Metrics["risk_score"] = riskMetric

	// Метрика производительности
	performanceMetric := &AIMetric{
		Name:        "performance_score",
		Value:       1.0,
		Unit:        "score",
		Trend:       TrendStable,
		Threshold:   0.5,
		LastUpdated: time.Now(),
	}
	aim.Metrics["performance_score"] = performanceMetric
}

// saveMetrics сохраняет метрики
func (aim *AIMonitoring) saveMetrics() {
	// В реальной реализации здесь было бы сохранение в базу данных
	// log.Printf("Saving %d metrics", len(aim.Metrics))
}

// updateDashboardWidgets обновляет виджеты дашборда
func (aim *AIMonitoring) updateDashboardWidgets() {
	aim.mutex.Lock()
	aim.Dashboard.Widgets = make([]DashboardWidget, 0)
	
	// Добавляем виджеты для метрик
	for name, metric := range aim.Metrics {
		widget := DashboardWidget{
			WidgetID: fmt.Sprintf("metric_%s", name),
			Type:     WidgetTypeGauge,
			Title:    name,
			Data: map[string]interface{}{
				"value":     metric.Value,
				"threshold": metric.Threshold,
				"trend":     metric.Trend.String(),
			},
			Position: WidgetPosition{X: 0, Y: 0, Width: 200, Height: 150},
		}
		aim.Dashboard.Widgets = append(aim.Dashboard.Widgets, widget)
	}
	
	// Добавляем виджеты для алертов
	for _, alert := range aim.Alerts {
		if !alert.Resolved {
			widget := DashboardWidget{
				WidgetID: fmt.Sprintf("alert_%s", alert.AlertID),
				Type:     WidgetTypeAlert,
				Title:    alert.AlertID,
				Data: map[string]interface{}{
					"message":   alert.Message,
					"severity":  alert.Severity.String(),
					"timestamp": alert.Timestamp,
				},
				Position: WidgetPosition{X: 0, Y: 0, Width: 300, Height: 100},
			}
			aim.Dashboard.Widgets = append(aim.Dashboard.Widgets, widget)
		}
	}
	aim.mutex.Unlock()
}

// validateMetric валидирует метрику
func (aim *AIMonitoring) validateMetric(metric *AIMetric) error {
	if metric.Value < 0 {
		return fmt.Errorf("metric value cannot be negative")
	}
	
	if metric.Threshold < 0 {
		return fmt.Errorf("metric threshold cannot be negative")
	}
	
	if metric.Unit == "" {
		return fmt.Errorf("metric unit cannot be empty")
	}
	
	return nil
}

// validateAlert валидирует алерт
func (aim *AIMonitoring) validateAlert(alert *AIAlert) error {
	if alert.Message == "" {
		return fmt.Errorf("alert message cannot be empty")
	}

	if alert.Type < AlertTypeAnomaly || alert.Type > AlertTypeSecurity {
		return fmt.Errorf("invalid alert type: %d", alert.Type)
	}

	if alert.Severity < AlertSeverityInfo || alert.Severity > AlertSeverityCritical {
		return fmt.Errorf("invalid alert severity: %d", alert.Severity)
	}

	return nil
}

// String methods for enums

func (tt ThreatType) String() string {
	switch tt {
	case ThreatTypeDDoS:
		return "DDoS"
	case ThreatTypeBruteForce:
		return "BruteForce"
	case ThreatTypeDataExfiltration:
		return "DataExfiltration"
	case ThreatTypePrivilegeEscalation:
		return "PrivilegeEscalation"
	case ThreatTypeMalware:
		return "Malware"
	default:
		return "Unknown"
	}
}

func (rl RiskLevel) String() string {
	switch rl {
	case RiskLevelLow:
		return "Low"
	case RiskLevelMedium:
		return "Medium"
	case RiskLevelHigh:
		return "High"
	case RiskLevelCritical:
		return "Critical"
	default:
		return "Unknown"
	}
}

func (at ActionType) String() string {
	switch at {
	case ActionTypeBlock:
		return "Block"
	case ActionTypeAlert:
		return "Alert"
	case ActionTypeQuarantine:
		return "Quarantine"
	case ActionTypeRateLimit:
		return "RateLimit"
	case ActionTypeLog:
		return "Log"
	default:
		return "Unknown"
	}
}

func (as AlertSeverity) String() string {
	switch as {
	case AlertSeverityInfo:
		return "Info"
	case AlertSeverityWarning:
		return "Warning"
	case AlertSeverityError:
		return "Error"
	case AlertSeverityCritical:
		return "Critical"
	default:
		return "Unknown"
	}
}

func (td TrendDirection) String() string {
	switch td {
	case TrendStable:
		return "Stable"
	case TrendUp:
		return "Up"
	case TrendDown:
		return "Down"
	case TrendSpike:
		return "Spike"
	default:
		return "Unknown"
	}
}
