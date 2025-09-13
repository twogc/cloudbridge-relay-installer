package cache

import (
	"sync"
	"time"

	"go.uber.org/zap"
)

// OptimizedCache представляет оптимизированный кэш с LRU и TTL
type OptimizedCache struct {
	mu       sync.RWMutex
	items    map[string]*CacheItem
	access   map[string]time.Time
	capacity int
	ttl      time.Duration
	logger   *zap.Logger
}

// CacheItem представляет элемент кэша
type CacheItem struct {
	Value     interface{}
	ExpiresAt time.Time
	AccessAt  time.Time
	Size      int64
}

// NewOptimizedCache создает новый оптимизированный кэш
func NewOptimizedCache(capacity int, ttl time.Duration, logger *zap.Logger) *OptimizedCache {
	cache := &OptimizedCache{
		items:    make(map[string]*CacheItem),
		access:   make(map[string]time.Time),
		capacity: capacity,
		ttl:      ttl,
		logger:   logger,
	}

	// Запускаем горутину для очистки устаревших элементов
	go cache.cleanupWorker()

	return cache
}

// Get получает значение из кэша
func (c *OptimizedCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return nil, false
	}

	// Проверяем TTL
	if time.Now().After(item.ExpiresAt) {
		delete(c.items, key)
		delete(c.access, key)
		return nil, false
	}

	// Обновляем время доступа
	item.AccessAt = time.Now()
	c.access[key] = time.Now()

	return item.Value, true
}

// Set устанавливает значение в кэш
func (c *OptimizedCache) Set(key string, value interface{}, size int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Если кэш переполнен, удаляем наименее используемые элементы
	if len(c.items) >= c.capacity {
		c.evictLRU()
	}

	// Создаем новый элемент
	item := &CacheItem{
		Value:     value,
		ExpiresAt: time.Now().Add(c.ttl),
		AccessAt:  time.Now(),
		Size:      size,
	}

	c.items[key] = item
	c.access[key] = time.Now()
}

// Delete удаляет элемент из кэша
func (c *OptimizedCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
	delete(c.access, key)
}

// Clear очищает весь кэш
func (c *OptimizedCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*CacheItem)
	c.access = make(map[string]time.Time)
}

// Size возвращает количество элементов в кэше
func (c *OptimizedCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.items)
}

// Stats возвращает статистику кэша
func (c *OptimizedCache) Stats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var totalSize int64
	for _, item := range c.items {
		totalSize += item.Size
	}

	return map[string]interface{}{
		"items":      len(c.items),
		"capacity":   c.capacity,
		"total_size": totalSize,
		"hit_rate":   c.calculateHitRate(),
	}
}

// evictLRU удаляет наименее используемые элементы
func (c *OptimizedCache) evictLRU() {
	if len(c.items) == 0 {
		return
	}

	// Находим элемент с наименьшим временем доступа
	var oldestKey string
	var oldestTime time.Time

	for key, accessTime := range c.access {
		if oldestKey == "" || accessTime.Before(oldestTime) {
			oldestKey = key
			oldestTime = accessTime
		}
	}

	if oldestKey != "" {
		delete(c.items, oldestKey)
		delete(c.access, oldestKey)
	}
}

// cleanupWorker периодически очищает устаревшие элементы
func (c *OptimizedCache) cleanupWorker() {
	ticker := time.NewTicker(c.ttl / 2)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanup()
		}
	}
}

// cleanup удаляет устаревшие элементы
func (c *OptimizedCache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, item := range c.items {
		if now.After(item.ExpiresAt) {
			delete(c.items, key)
			delete(c.access, key)
		}
	}
}

// calculateHitRate вычисляет коэффициент попаданий (заглушка)
func (c *OptimizedCache) calculateHitRate() float64 {
	// В реальной реализации здесь была бы логика подсчета попаданий/промахов
	return 0.85 // Заглушка для демонстрации
}

// BatchGet получает несколько значений за один раз
func (c *OptimizedCache) BatchGet(keys []string) map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]interface{})
	now := time.Now()

	for _, key := range keys {
		item, exists := c.items[key]
		if exists && now.Before(item.ExpiresAt) {
			result[key] = item.Value
			// Обновляем время доступа
			item.AccessAt = now
			c.access[key] = now
		}
	}

	return result
}

// BatchSet устанавливает несколько значений за один раз
func (c *OptimizedCache) BatchSet(items map[string]interface{}, sizes map[string]int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, value := range items {
		// Если кэш переполнен, удаляем наименее используемые элементы
		if len(c.items) >= c.capacity {
			c.evictLRU()
		}

		size := int64(0)
		if s, exists := sizes[key]; exists {
			size = s
		}

		item := &CacheItem{
			Value:     value,
			ExpiresAt: now.Add(c.ttl),
			AccessAt:  now,
			Size:      size,
		}

		c.items[key] = item
		c.access[key] = now
	}
}

// GetWithTTL получает значение с возможностью установки кастомного TTL
func (c *OptimizedCache) GetWithTTL(key string, customTTL time.Duration) (interface{}, bool) {
	value, exists := c.Get(key)
	if exists && customTTL > 0 {
		// Обновляем TTL для элемента
		c.mu.Lock()
		if item, ok := c.items[key]; ok {
			item.ExpiresAt = time.Now().Add(customTTL)
		}
		c.mu.Unlock()
	}
	return value, exists
}
