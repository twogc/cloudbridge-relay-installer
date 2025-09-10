package routing

import (
	"math"
	"sync"
)

// MultiArmedBandit реализует алгоритм Multi-Armed Bandit для выбора протокола
type MultiArmedBandit struct {
	arms            map[string]*Arm
	explorationRate float64
	mu              sync.RWMutex
}

// Arm представляет "руку" в multi-armed bandit
type Arm struct {
	Count    int     `json:"count"`
	Reward   float64 `json:"reward"`
	Protocol string  `json:"protocol"`
}

// NewMultiArmedBandit создает новый multi-armed bandit
func NewMultiArmedBandit(explorationRate float64) *MultiArmedBandit {
	return &MultiArmedBandit{
		arms:            make(map[string]*Arm),
		explorationRate: explorationRate,
	}
}

// Initialize инициализирует multi-armed bandit
func (mab *MultiArmedBandit) Initialize() {
	mab.mu.Lock()
	defer mab.mu.Unlock()

	// Инициализируем доступные протоколы
	protocols := []string{"QUIC", "WireGuard", "TCP", "UDP"}

	for _, protocol := range protocols {
		mab.arms[protocol] = &Arm{
			Count:    0,
			Reward:   0.0,
			Protocol: protocol,
		}
	}
}

// Select выбирает протокол на основе UCB1 алгоритма
func (mab *MultiArmedBandit) Select(source, destination string) string {
	mab.mu.Lock()
	defer mab.mu.Unlock()

	// Если есть неиспользованные протоколы, выбираем их
	for protocol, arm := range mab.arms {
		if arm.Count == 0 {
			return protocol
		}
	}

	// UCB1 алгоритм
	bestProtocol := ""
	bestValue := -1.0
	totalCount := mab.getTotalCount()

	for protocol, arm := range mab.arms {
		// UCB1 формула
		exploitation := arm.Reward
		exploration := math.Sqrt(2 * math.Log(float64(totalCount)) / float64(arm.Count))
		value := exploitation + exploration

		if value > bestValue {
			bestValue = value
			bestProtocol = protocol
		}
	}

	return bestProtocol
}

// UpdateReward обновляет награду для выбранного протокола
func (mab *MultiArmedBandit) UpdateReward(protocol string, reward float64) {
	mab.mu.Lock()
	defer mab.mu.Unlock()

	arm, exists := mab.arms[protocol]
	if !exists {
		return
	}

	// Обновляем счетчик и награду
	arm.Count++

	// Экспоненциальное скользящее среднее
	arm.Reward = arm.Reward + 0.1*(reward-arm.Reward)
}

// GetUsageCount возвращает количество использований протокола
func (mab *MultiArmedBandit) GetUsageCount(protocol string) int {
	mab.mu.RLock()
	defer mab.mu.RUnlock()

	arm, exists := mab.arms[protocol]
	if !exists {
		return 0
	}

	return arm.Count
}

// GetReward возвращает текущую награду протокола
func (mab *MultiArmedBandit) GetReward(protocol string) float64 {
	mab.mu.RLock()
	defer mab.mu.RUnlock()

	arm, exists := mab.arms[protocol]
	if !exists {
		return 0.0
	}

	return arm.Reward
}

// getTotalCount возвращает общее количество использований
func (mab *MultiArmedBandit) getTotalCount() int {
	total := 0
	for _, arm := range mab.arms {
		total += arm.Count
	}
	return total
}

// GetStats возвращает статистику по всем протоколам
func (mab *MultiArmedBandit) GetStats() map[string]interface{} {
	mab.mu.RLock()
	defer mab.mu.RUnlock()

	stats := make(map[string]interface{})

	for protocol, arm := range mab.arms {
		stats[protocol] = map[string]interface{}{
			"count":  arm.Count,
			"reward": arm.Reward,
		}
	}

	return stats
}

