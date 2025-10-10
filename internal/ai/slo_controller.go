package ai

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"go.uber.org/zap"
)

// SLOController управляет SLO (Service Level Objectives) с гистерезисом
type SLOController struct {
	logger *zap.Logger
	config *SLOConfig

	// Метрики и состояние
	metrics    map[string]*SLOMetrics
	metricsMux sync.RWMutex

	// Гистерезис состояние
	hysteresis    map[string]*HysteresisState
	hysteresisMux sync.RWMutex

	// Hold-down timers
	holdDownTimers map[string]*HoldDownTimer
	timersMux      sync.RWMutex
}

// SLOConfig конфигурация SLO контроллера
type SLOConfig struct {
	// SLO thresholds
	LatencyThreshold      float64 `json:"latency_threshold"`      // ms
	ErrorRateThreshold    float64 `json:"error_rate_threshold"`   // %
	AvailabilityThreshold float64 `json:"availability_threshold"` // %

	// Hysteresis settings
	HysteresisEnabled   bool          `json:"hysteresis_enabled"`
	HysteresisThreshold float64       `json:"hysteresis_threshold"` // %
	StabilizationTime   time.Duration `json:"stabilization_time"`

	// Hold-down settings
	HoldDownEnabled bool          `json:"hold_down_enabled"`
	HoldDownTime    time.Duration `json:"hold_down_time"`
	MinHoldDownTime time.Duration `json:"min_hold_down_time"`
	MaxHoldDownTime time.Duration `json:"max_hold_down_time"`

	// Dampening settings
	DampeningEnabled bool    `json:"dampening_enabled"`
	DampeningFactor  float64 `json:"dampening_factor"` // 0.0-1.0
	EmaAlpha         float64 `json:"ema_alpha"`        // EMA smoothing factor

	// Cost function
	CostFunctionEnabled bool    `json:"cost_function_enabled"`
	SwitchPenalty       float64 `json:"switch_penalty"`   // Penalty for switching
	StabilityReward     float64 `json:"stability_reward"` // Reward for stability
}

// DefaultSLOConfig возвращает конфигурацию по умолчанию
func DefaultSLOConfig() *SLOConfig {
	return &SLOConfig{
		LatencyThreshold:      100.0, // 100ms
		ErrorRateThreshold:    1.0,   // 1%
		AvailabilityThreshold: 99.9,  // 99.9%

		HysteresisEnabled:   true,
		HysteresisThreshold: 10.0, // 10%
		StabilizationTime:   2 * time.Minute,

		HoldDownEnabled: true,
		HoldDownTime:    30 * time.Second,
		MinHoldDownTime: 10 * time.Second,
		MaxHoldDownTime: 5 * time.Minute,

		DampeningEnabled: true,
		DampeningFactor:  0.7, // 70% dampening
		EmaAlpha:         0.3, // 30% new data, 70% old data

		CostFunctionEnabled: true,
		SwitchPenalty:       0.1,  // 10% penalty for switching
		StabilityReward:     0.05, // 5% reward for stability
	}
}

// SLOMetrics представляет метрики SLO
type SLOMetrics struct {
	ServiceID string    `json:"service_id"`
	Timestamp time.Time `json:"timestamp"`

	// Основные метрики
	Latency      float64 `json:"latency"`      // ms
	ErrorRate    float64 `json:"error_rate"`   // %
	Availability float64 `json:"availability"` // %
	Throughput   float64 `json:"throughput"`   // req/s

	// Дополнительные метрики
	CPUUsage       float64 `json:"cpu_usage"`       // %
	MemoryUsage    float64 `json:"memory_usage"`    // %
	NetworkLatency float64 `json:"network_latency"` // ms

	// SLO статус
	SLOStatus      string `json:"slo_status"` // good, degraded, critical
	ViolationCount int    `json:"violation_count"`
}

// HysteresisState представляет состояние гистерезиса
type HysteresisState struct {
	ServiceID        string    `json:"service_id"`
	CurrentState     string    `json:"current_state"` // good, degraded, critical
	PreviousState    string    `json:"previous_state"`
	LastTransition   time.Time `json:"last_transition"`
	StabilizationEnd time.Time `json:"stabilization_end"`

	// EMA фильтры
	LatencyEMA      float64 `json:"latency_ema"`
	ErrorRateEMA    float64 `json:"error_rate_ema"`
	AvailabilityEMA float64 `json:"availability_ema"`

	// Hold-down состояние
	InHoldDown  bool      `json:"in_hold_down"`
	HoldDownEnd time.Time `json:"hold_down_end"`

	// Переключения
	SwitchCount int       `json:"switch_count"`
	LastSwitch  time.Time `json:"last_switch"`
}

// HoldDownTimer представляет таймер hold-down
type HoldDownTimer struct {
	ServiceID     string    `json:"service_id"`
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	IsActive      bool      `json:"is_active"`
	TriggerReason string    `json:"trigger_reason"`
}

// NewSLOController создает новый SLO контроллер
func NewSLOController(logger *zap.Logger, config *SLOConfig) *SLOController {
	if config == nil {
		config = DefaultSLOConfig()
	}

	return &SLOController{
		logger:         logger,
		config:         config,
		metrics:        make(map[string]*SLOMetrics),
		hysteresis:     make(map[string]*HysteresisState),
		holdDownTimers: make(map[string]*HoldDownTimer),
	}
}

// UpdateMetrics обновляет метрики SLO с гистерезисом
func (sc *SLOController) UpdateMetrics(serviceID string, metrics *SLOMetrics) (*SLODecision, error) {
	sc.metricsMux.Lock()
	sc.metrics[serviceID] = metrics
	sc.metricsMux.Unlock()

	// Получаем или создаем состояние гистерезиса
	sc.hysteresisMux.Lock()
	state, exists := sc.hysteresis[serviceID]
	if !exists {
		state = &HysteresisState{
			ServiceID:       serviceID,
			CurrentState:    "good",
			LatencyEMA:      metrics.Latency,
			ErrorRateEMA:    metrics.ErrorRate,
			AvailabilityEMA: metrics.Availability,
		}
		sc.hysteresis[serviceID] = state
	}
	sc.hysteresisMux.Unlock()

	// Обновляем EMA фильтры
	sc.updateEMA(state, metrics)

	// Определяем новое состояние с гистерезисом
	newState := sc.determineStateWithHysteresis(state, metrics)

	// Проверяем hold-down
	if sc.config.HoldDownEnabled {
		if sc.isInHoldDown(serviceID) {
			sc.logger.Debug("Service in hold-down period, skipping state change",
				zap.String("service_id", serviceID),
				zap.String("current_state", state.CurrentState))
			return &SLODecision{
				Action:     "hold",
				Reason:     "hold_down_active",
				State:      state.CurrentState,
				Confidence: 1.0,
			}, nil
		}
	}

	// Применяем dampening
	if sc.config.DampeningEnabled {
		newState = sc.applyDampening(state, newState, metrics)
	}

	// Вычисляем cost function
	cost := sc.calculateCost(state, newState, metrics)

	// Принимаем решение
	decision := sc.makeDecision(state, newState, cost, metrics)

	// Обновляем состояние
	if decision.Action != "hold" {
		sc.updateHysteresisState(state, newState)
	}

	sc.logger.Info("SLO metrics updated with hysteresis",
		zap.String("service_id", serviceID),
		zap.String("old_state", state.CurrentState),
		zap.String("new_state", newState),
		zap.String("action", decision.Action),
		zap.String("reason", decision.Reason),
		zap.Float64("cost", cost))

	return decision, nil
}

// updateEMA обновляет EMA фильтры
func (sc *SLOController) updateEMA(state *HysteresisState, metrics *SLOMetrics) {
	alpha := sc.config.EmaAlpha

	state.LatencyEMA = alpha*metrics.Latency + (1-alpha)*state.LatencyEMA
	state.ErrorRateEMA = alpha*metrics.ErrorRate + (1-alpha)*state.ErrorRateEMA
	state.AvailabilityEMA = alpha*metrics.Availability + (1-alpha)*state.AvailabilityEMA
}

// determineStateWithHysteresis определяет состояние с учетом гистерезиса
func (sc *SLOController) determineStateWithHysteresis(state *HysteresisState, metrics *SLOMetrics) string {
	// Определяем базовое состояние
	baseState := sc.determineBaseState(metrics)

	// Применяем гистерезис
	if sc.config.HysteresisEnabled {
		return sc.applyHysteresis(state, baseState, metrics)
	}

	return baseState
}

// determineBaseState определяет базовое состояние без гистерезиса
func (sc *SLOController) determineBaseState(metrics *SLOMetrics) string {
	if metrics.Latency > sc.config.LatencyThreshold ||
		metrics.ErrorRate > sc.config.ErrorRateThreshold ||
		metrics.Availability < sc.config.AvailabilityThreshold {
		return "critical"
	}

	// Проверяем degraded состояние (50% от threshold)
	if metrics.Latency > sc.config.LatencyThreshold*0.5 ||
		metrics.ErrorRate > sc.config.ErrorRateThreshold*0.5 ||
		metrics.Availability < sc.config.AvailabilityThreshold+0.1 {
		return "degraded"
	}

	return "good"
}

// applyHysteresis применяет гистерезис
func (sc *SLOController) applyHysteresis(state *HysteresisState, newState string, metrics *SLOMetrics) string {
	// Если состояние не изменилось, возвращаем текущее
	if state.CurrentState == newState {
		return state.CurrentState
	}

	// Вычисляем улучшение/ухудшение
	improvement := sc.calculateImprovement(state, newState, metrics)

	// Применяем гистерезис
	threshold := sc.config.HysteresisThreshold / 100.0

	if improvement > threshold {
		// Улучшение достаточно значительное
		return newState
	} else if improvement < -threshold {
		// Ухудшение достаточно значительное
		return newState
	} else {
		// Изменение в пределах гистерезиса - сохраняем текущее состояние
		sc.logger.Debug("Change within hysteresis threshold, keeping current state",
			zap.String("service_id", state.ServiceID),
			zap.String("current_state", state.CurrentState),
			zap.String("proposed_state", newState),
			zap.Float64("improvement", improvement),
			zap.Float64("threshold", threshold))
		return state.CurrentState
	}
}

// calculateImprovement вычисляет улучшение/ухудшение
func (sc *SLOController) calculateImprovement(state *HysteresisState, newState string, metrics *SLOMetrics) float64 {
	// Простая метрика улучшения на основе латентности
	currentLatency := state.LatencyEMA
	newLatency := metrics.Latency

	if currentLatency == 0 {
		return 0
	}

	// Положительное значение = улучшение, отрицательное = ухудшение
	return (currentLatency - newLatency) / currentLatency
}

// applyDampening применяет dampening для предотвращения флаппинга
func (sc *SLOController) applyDampening(state *HysteresisState, newState string, metrics *SLOMetrics) string {
	// Если состояние не изменилось, возвращаем
	if state.CurrentState == newState {
		return newState
	}

	// Проверяем, не слишком ли часто переключаемся
	if time.Since(state.LastSwitch) < sc.config.StabilizationTime {
		sc.logger.Debug("Too frequent switching, applying dampening",
			zap.String("service_id", state.ServiceID),
			zap.Duration("time_since_last_switch", time.Since(state.LastSwitch)),
			zap.Duration("stabilization_time", sc.config.StabilizationTime))
		return state.CurrentState
	}

	// Применяем dampening factor
	if sc.config.DampeningEnabled {
		// С вероятностью dampening factor сохраняем текущее состояние
		if rand.Float64() < sc.config.DampeningFactor {
			sc.logger.Debug("Dampening applied, keeping current state",
				zap.String("service_id", state.ServiceID),
				zap.Float64("dampening_factor", sc.config.DampeningFactor))
			return state.CurrentState
		}
	}

	return newState
}

// calculateCost вычисляет cost function для переключения
func (sc *SLOController) calculateCost(state *HysteresisState, newState string, metrics *SLOMetrics) float64 {
	if !sc.config.CostFunctionEnabled {
		return 0
	}

	cost := 0.0

	// Penalty за переключение
	if state.CurrentState != newState {
		cost += sc.config.SwitchPenalty
	}

	// Reward за стабильность
	if state.CurrentState == newState {
		cost -= sc.config.StabilityReward
	}

	// Дополнительная penalty за частые переключения
	if state.SwitchCount > 3 {
		cost += float64(state.SwitchCount) * 0.02 // 2% за каждое переключение
	}

	return cost
}

// makeDecision принимает решение на основе cost function
func (sc *SLOController) makeDecision(state *HysteresisState, newState string, cost float64, metrics *SLOMetrics) *SLODecision {
	decision := &SLODecision{
		ServiceID: state.ServiceID,
		Timestamp: time.Now(),
		State:     newState,
		Cost:      cost,
	}

	// Если cost слишком высокий, не переключаемся
	if cost > 0.2 { // 20% threshold
		decision.Action = "hold"
		decision.Reason = "high_cost"
		decision.Confidence = 1.0 - cost
		return decision
	}

	// Если состояние не изменилось
	if state.CurrentState == newState {
		decision.Action = "maintain"
		decision.Reason = "no_change"
		decision.Confidence = 1.0
		return decision
	}

	// Переключаемся
	decision.Action = "switch"
	decision.Reason = "slo_violation"
	decision.Confidence = 1.0 - cost

	// Устанавливаем hold-down timer
	if sc.config.HoldDownEnabled {
		sc.setHoldDownTimer(state.ServiceID, "state_change")
	}

	return decision
}

// isInHoldDown проверяет, находится ли сервис в hold-down
func (sc *SLOController) isInHoldDown(serviceID string) bool {
	sc.timersMux.RLock()
	timer, exists := sc.holdDownTimers[serviceID]
	sc.timersMux.RUnlock()

	if !exists {
		return false
	}

	return timer.IsActive && time.Now().Before(timer.EndTime)
}

// setHoldDownTimer устанавливает hold-down таймер
func (sc *SLOController) setHoldDownTimer(serviceID, reason string) {
	holdDownTime := sc.config.HoldDownTime

	// Адаптивный hold-down на основе частоты переключений
	sc.hysteresisMux.RLock()
	state, exists := sc.hysteresis[serviceID]
	sc.hysteresisMux.RUnlock()

	if exists && state.SwitchCount > 5 {
		// Увеличиваем hold-down для часто переключающихся сервисов
		holdDownTime = time.Duration(float64(holdDownTime) * 1.5)
		if holdDownTime > sc.config.MaxHoldDownTime {
			holdDownTime = sc.config.MaxHoldDownTime
		}
	}

	timer := &HoldDownTimer{
		ServiceID:     serviceID,
		StartTime:     time.Now(),
		EndTime:       time.Now().Add(holdDownTime),
		IsActive:      true,
		TriggerReason: reason,
	}

	sc.timersMux.Lock()
	sc.holdDownTimers[serviceID] = timer
	sc.timersMux.Unlock()

	sc.logger.Info("Hold-down timer set",
		zap.String("service_id", serviceID),
		zap.Duration("hold_down_time", holdDownTime),
		zap.String("reason", reason))
}

// updateHysteresisState обновляет состояние гистерезиса
func (sc *SLOController) updateHysteresisState(state *HysteresisState, newState string) {
	state.PreviousState = state.CurrentState
	state.CurrentState = newState
	state.LastTransition = time.Now()

	if state.CurrentState != state.PreviousState {
		state.SwitchCount++
		state.LastSwitch = time.Now()
	}
}

// SLODecision представляет решение SLO контроллера
type SLODecision struct {
	ServiceID  string    `json:"service_id"`
	Timestamp  time.Time `json:"timestamp"`
	Action     string    `json:"action"` // switch, maintain, hold
	Reason     string    `json:"reason"`
	State      string    `json:"state"`
	Cost       float64   `json:"cost"`
	Confidence float64   `json:"confidence"`
}

// GetSLOStatus возвращает текущий статус SLO
func (sc *SLOController) GetSLOStatus(serviceID string) (*SLOStatus, error) {
	sc.metricsMux.RLock()
	metrics, exists := sc.metrics[serviceID]
	sc.metricsMux.RUnlock()

	if !exists {
		return nil, fmt.Errorf("service not found")
	}

	sc.hysteresisMux.RLock()
	state, exists := sc.hysteresis[serviceID]
	sc.hysteresisMux.RUnlock()

	if !exists {
		return nil, fmt.Errorf("hysteresis state not found")
	}

	return &SLOStatus{
		ServiceID:      serviceID,
		CurrentState:   state.CurrentState,
		PreviousState:  state.PreviousState,
		LastTransition: state.LastTransition,
		SwitchCount:    state.SwitchCount,
		InHoldDown:     sc.isInHoldDown(serviceID),
		Metrics:        metrics,
		EMA: &EMAValues{
			Latency:      state.LatencyEMA,
			ErrorRate:    state.ErrorRateEMA,
			Availability: state.AvailabilityEMA,
		},
	}, nil
}

// SLOStatus представляет статус SLO
type SLOStatus struct {
	ServiceID      string      `json:"service_id"`
	CurrentState   string      `json:"current_state"`
	PreviousState  string      `json:"previous_state"`
	LastTransition time.Time   `json:"last_transition"`
	SwitchCount    int         `json:"switch_count"`
	InHoldDown     bool        `json:"in_hold_down"`
	Metrics        *SLOMetrics `json:"metrics"`
	EMA            *EMAValues  `json:"ema"`
}

// EMAValues представляет значения EMA фильтров
type EMAValues struct {
	Latency      float64 `json:"latency"`
	ErrorRate    float64 `json:"error_rate"`
	Availability float64 `json:"availability"`
}

// CleanupExpiredTimers очищает истекшие таймеры
func (sc *SLOController) CleanupExpiredTimers() {
	now := time.Now()

	sc.timersMux.Lock()
	for serviceID, timer := range sc.holdDownTimers {
		if now.After(timer.EndTime) {
			timer.IsActive = false
			delete(sc.holdDownTimers, serviceID)
		}
	}
	sc.timersMux.Unlock()
}

// GetSLOStats возвращает статистику SLO
func (sc *SLOController) GetSLOStats() map[string]interface{} {
	sc.metricsMux.RLock()
	sc.hysteresisMux.RLock()
	sc.timersMux.RLock()
	defer sc.metricsMux.RUnlock()
	defer sc.hysteresisMux.RUnlock()
	defer sc.timersMux.RUnlock()

	stats := map[string]interface{}{
		"total_services":     len(sc.metrics),
		"active_timers":      len(sc.holdDownTimers),
		"state_distribution": make(map[string]int),
		"switch_counts":      make(map[int]int),
	}

	stateDist := make(map[string]int)
	switchCounts := make(map[int]int)

	for _, state := range sc.hysteresis {
		stateDist[state.CurrentState]++
		switchCounts[state.SwitchCount]++
	}

	stats["state_distribution"] = stateDist
	stats["switch_counts"] = switchCounts

	return stats
}
