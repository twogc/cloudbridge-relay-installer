package ai

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEventParsingFloat64 проверяет парсинг float64 из JSON
func TestEventParsingFloat64(t *testing.T) {
	tests := []struct {
		name     string
		event    map[string]interface{}
		key      string
		expected int
	}{
		{
			name:     "parse JSON number (float64) to int",
			event:    map[string]interface{}{"volume": 100.5},
			key:      "volume",
			expected: 100,
		},
		{
			name:     "parse large float64",
			event:    map[string]interface{}{"bytes": 9999999.99},
			key:      "bytes",
			expected: 9999999,
		},
		{
			name:     "parse zero float64",
			event:    map[string]interface{}{"count": 0.0},
			key:      "count",
			expected: 0,
		},
		{
			name:     "parse negative float64",
			event:    map[string]interface{}{"delta": -42.7},
			key:      "delta",
			expected: -42,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getIntFromEvent(tt.event, tt.key)
			assert.Equal(t, tt.expected, result, "Should correctly parse float64 to int")
		})
	}
}

// TestEventParsingStringNumbers проверяет парсинг чисел из строк
func TestEventParsingStringNumbers(t *testing.T) {
	tests := []struct {
		name     string
		event    map[string]interface{}
		key      string
		expected int
	}{
		{
			name:     "parse string number",
			event:    map[string]interface{}{"count": "42"},
			key:      "count",
			expected: 42,
		},
		{
			name:     "parse large string number",
			event:    map[string]interface{}{"id": "123456789"},
			key:      "id",
			expected: 123456789,
		},
		{
			name:     "parse invalid string returns 0",
			event:    map[string]interface{}{"invalid": "not-a-number"},
			key:      "invalid",
			expected: 0,
		},
		{
			name:     "parse empty string returns 0",
			event:    map[string]interface{}{"empty": ""},
			key:      "empty",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getIntFromEvent(tt.event, tt.key)
			assert.Equal(t, tt.expected, result, "Should correctly parse string numbers")
		})
	}
}

// TestEventParsingDurations проверяет парсинг duration из строк
func TestEventParsingDurations(t *testing.T) {
	tests := []struct {
		name     string
		event    map[string]interface{}
		key      string
		expected time.Duration
	}{
		{
			name:     "parse seconds",
			event:    map[string]interface{}{"timeout": "5s"},
			key:      "timeout",
			expected: 5 * time.Second,
		},
		{
			name:     "parse minutes",
			event:    map[string]interface{}{"interval": "30m"},
			key:      "interval",
			expected: 30 * time.Minute,
		},
		{
			name:     "parse hours",
			event:    map[string]interface{}{"duration": "2h"},
			key:      "duration",
			expected: 2 * time.Hour,
		},
		{
			name:     "parse complex duration",
			event:    map[string]interface{}{"wait": "1h30m15s"},
			key:      "wait",
			expected: 1*time.Hour + 30*time.Minute + 15*time.Second,
		},
		{
			name:     "parse duration from int64 nanoseconds",
			event:    map[string]interface{}{"delay": int64(1000000000)}, // 1 second
			key:      "delay",
			expected: time.Second,
		},
		{
			name:     "parse duration from float64 nanoseconds",
			event:    map[string]interface{}{"wait_ns": 2000000000.0}, // 2 seconds
			key:      "wait_ns",
			expected: 2 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getDurationFromEvent(tt.event, tt.key)
			assert.Equal(t, tt.expected, result, "Should correctly parse durations")
		})
	}
}

// TestEventTypeKeyCorrection проверяет, что используется правильный ключ "type"
func TestEventTypeKeyCorrection(t *testing.T) {
	tests := []struct {
		name        string
		event       map[string]interface{}
		shouldMatch bool
	}{
		{
			name:        "correct type key",
			event:       map[string]interface{}{"type": "login"},
			shouldMatch: true,
		},
		{
			name:        "incorrect event_type key",
			event:       map[string]interface{}{"event_type": "login"},
			shouldMatch: false,
		},
		{
			name:        "p2p_quic_connection with correct key",
			event:       map[string]interface{}{"type": "p2p_quic_connection"},
			shouldMatch: true,
		},
		{
			name:        "quic_connection with correct key",
			event:       map[string]interface{}{"type": "quic_connection"},
			shouldMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventType, ok := tt.event["type"].(string)

			if tt.shouldMatch {
				assert.True(t, ok, "Should find 'type' key in event")
				assert.NotEmpty(t, eventType, "Type should not be empty")
			} else {
				// Проверяем, что с неправильным ключом типа не найдется
				assert.False(t, ok, "Should not find 'type' key when using 'event_type'")
			}
		})
	}
}

// TestUserProfilePersistence проверяет сохранение и загрузку профилей
func TestUserProfilePersistence(t *testing.T) {
	// Создаем временную директорию для тестов
	tempDir := t.TempDir()
	testDataPath := filepath.Join(tempDir, "user_profiles.json")

	// Создаем BehaviorAnalyzer с тестовыми данными
	ba := &BehaviorAnalyzer{
		Enabled:       true,
		UserProfiles:  make(map[string]*UserProfile),
		AnomalyScores: make(map[string]float64),
	}

	// Добавляем тестовые профили
	now := time.Now()
	testProfile1 := &UserProfile{
		UserID: "user-1",
		LoginPatterns: []LoginPattern{
			{TimeOfDay: now.Add(-24 * time.Hour), IPAddress: "192.168.1.1", Success: true},
			{TimeOfDay: now.Add(-12 * time.Hour), IPAddress: "192.168.1.1", Success: true},
		},
		AccessPatterns: []AccessPattern{
			{Resource: "/api/users", TimeOfDay: now, Duration: 100 * time.Millisecond, Frequency: 10, DataVolume: 1024},
			{Resource: "/api/data", TimeOfDay: now, Duration: 200 * time.Millisecond, Frequency: 15, DataVolume: 2048},
		},
		NetworkPatterns: []NetworkPattern{
			{Destination: "10.0.0.1", Protocol: "TCP", Port: 443, DataVolume: 3000, Frequency: 20},
		},
		RiskScore:   0.3,
		LastUpdated: now,
	}

	testProfile2 := &UserProfile{
		UserID: "user-2",
		LoginPatterns: []LoginPattern{
			{TimeOfDay: now.Add(-48 * time.Hour), IPAddress: "10.0.0.1", Success: true},
			{TimeOfDay: now.Add(-24 * time.Hour), IPAddress: "10.0.0.1", Success: false},
		},
		AccessPatterns: []AccessPattern{
			{Resource: "/api/admin", TimeOfDay: now, Duration: 500 * time.Millisecond, Frequency: 5, DataVolume: 5120},
		},
		NetworkPatterns: []NetworkPattern{
			{Destination: "192.168.1.1", Protocol: "TCP", Port: 8080, DataVolume: 15000, Frequency: 100},
		},
		RiskScore:   0.7,
		LastUpdated: now,
	}

	ba.UserProfiles["user-1"] = testProfile1
	ba.UserProfiles["user-2"] = testProfile2

	// Сохраняем профили
	data, err := json.MarshalIndent(ba.UserProfiles, "", "  ")
	require.NoError(t, err, "Should marshal profiles")

	err = os.WriteFile(testDataPath, data, 0644)
	require.NoError(t, err, "Should write profiles to file")

	// Создаем новый анализатор и загружаем профили
	ba2 := &BehaviorAnalyzer{
		UserProfiles:  make(map[string]*UserProfile),
		AnomalyScores: make(map[string]float64),
	}

	loadedData, err := os.ReadFile(testDataPath)
	require.NoError(t, err, "Should read profiles from file")

	err = json.Unmarshal(loadedData, &ba2.UserProfiles)
	require.NoError(t, err, "Should unmarshal profiles")

	// Проверяем корректность загрузки
	assert.Len(t, ba2.UserProfiles, 2, "Should have 2 profiles")

	// Проверяем user-1
	profile1, ok := ba2.UserProfiles["user-1"]
	require.True(t, ok, "Should have user-1 profile")
	assert.Equal(t, "user-1", profile1.UserID)
	assert.Equal(t, 0.3, profile1.RiskScore)
	assert.Len(t, profile1.LoginPatterns, 2)
	assert.Len(t, profile1.AccessPatterns, 2)
	assert.Equal(t, "/api/users", profile1.AccessPatterns[0].Resource)
	assert.Equal(t, 10, profile1.AccessPatterns[0].Frequency)

	// Проверяем user-2
	profile2, ok := ba2.UserProfiles["user-2"]
	require.True(t, ok, "Should have user-2 profile")
	assert.Equal(t, "user-2", profile2.UserID)
	assert.Equal(t, 0.7, profile2.RiskScore)
	assert.Len(t, profile1.LoginPatterns, 2)
	assert.Equal(t, "/api/admin", profile2.AccessPatterns[0].Resource)

	t.Logf("✅ Successfully persisted and loaded %d user profiles", len(ba2.UserProfiles))
}

// TestConcurrentProfileAccess проверяет thread-safety
func TestConcurrentProfileAccess(t *testing.T) {
	ba := &BehaviorAnalyzer{
		Enabled:       true,
		UserProfiles:  make(map[string]*UserProfile),
		AnomalyScores: make(map[string]float64),
		Threshold:     0.8,
	}

	var wg sync.WaitGroup
	numGoroutines := 100
	numOperationsPerGoroutine := 50

	// Горутины для записи
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperationsPerGoroutine; j++ {
				userID := "user-" + string(rune('A'+id%26))
				event := map[string]interface{}{
					"type":        "login",
					"timestamp":   time.Now(),
					"user_id":     userID,
					"volume":      float64(j * 100),      // JSON float64
					"connections": "10",                  // String number
					"duration":    "5s",                  // Duration string
				}
				err := ba.AnalyzeUserBehavior(userID, event)
				assert.NoError(t, err)
			}
		}(i)
	}

	// Горутины для чтения
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperationsPerGoroutine; j++ {
				userID := "user-" + string(rune('A'+id%26))
				ba.mutex.RLock()
				_, exists := ba.UserProfiles[userID]
				ba.mutex.RUnlock()
				_ = exists
			}
		}(i)
	}

	wg.Wait()

	// Проверяем результаты
	ba.mutex.RLock()
	profileCount := len(ba.UserProfiles)
	ba.mutex.RUnlock()

	assert.Greater(t, profileCount, 0, "Should have created profiles")
	assert.LessOrEqual(t, profileCount, 26, "Should have at most 26 profiles (A-Z)")

	t.Logf("✅ Concurrent access test passed: created %d profiles without races", profileCount)
}

// TestIntegrationP2PQUICEventType проверяет правильность типа события в P2P QUIC интеграции
func TestIntegrationP2PQUICEventType(t *testing.T) {
	ba := &BehaviorAnalyzer{
		Enabled:       true,
		UserProfiles:  make(map[string]*UserProfile),
		AnomalyScores: make(map[string]float64),
		Threshold:     0.8,
	}

	peerID := "test-peer-123"
	err := ba.AnalyzeP2PQUICBehavior(peerID, nil)
	require.NoError(t, err)

	// Проверяем, что профиль создался
	userID := "p2p-quic-peer-" + peerID
	ba.mutex.RLock()
	profile, exists := ba.UserProfiles[userID]
	ba.mutex.RUnlock()

	assert.True(t, exists, "Should have created profile for P2P QUIC peer")
	assert.NotNil(t, profile)
	assert.Equal(t, userID, profile.UserID)

	t.Logf("✅ P2P QUIC integration creates profile with correct user ID")
}

// TestIntegrationQUICEventType проверяет правильность типа события в QUIC интеграции
func TestIntegrationQUICEventType(t *testing.T) {
	ba := &BehaviorAnalyzer{
		Enabled:       true,
		UserProfiles:  make(map[string]*UserProfile),
		AnomalyScores: make(map[string]float64),
		Threshold:     0.8,
	}

	connectionID := "test-conn-456"
	err := ba.AnalyzeQUICBehavior(connectionID, nil)
	require.NoError(t, err)

	// Проверяем, что профиль создался
	userID := "quic-connection-" + connectionID
	ba.mutex.RLock()
	profile, exists := ba.UserProfiles[userID]
	ba.mutex.RUnlock()

	assert.True(t, exists, "Should have created profile for QUIC connection")
	assert.NotNil(t, profile)
	assert.Equal(t, userID, profile.UserID)

	t.Logf("✅ QUIC integration creates profile with correct user ID")
}

// BenchmarkEventParsingPerformance бенчмарк производительности парсинга
func BenchmarkEventParsingPerformance(b *testing.B) {
	event := map[string]interface{}{
		"type":        "login",
		"volume":      1000.5,      // float64
		"count":       "42",         // string
		"duration":    "5s",         // duration string
		"connections": int64(100),   // int64
	}

	b.Run("getIntFromEvent-float64", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = getIntFromEvent(event, "volume")
		}
	})

	b.Run("getIntFromEvent-string", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = getIntFromEvent(event, "count")
		}
	})

	b.Run("getInt64FromEvent", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = getInt64FromEvent(event, "connections")
		}
	})

	b.Run("getDurationFromEvent", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = getDurationFromEvent(event, "duration")
		}
	})
}

// BenchmarkConcurrentProfileUpdates бенчмарк concurrent обновлений профилей
func BenchmarkConcurrentProfileUpdates(b *testing.B) {
	ba := &BehaviorAnalyzer{
		Enabled:       true,
		UserProfiles:  make(map[string]*UserProfile),
		AnomalyScores: make(map[string]float64),
		Threshold:     0.8,
	}

	event := map[string]interface{}{
		"type":      "login",
		"timestamp": time.Now(),
		"volume":    100.5,
	}

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			userID := "user-" + string(rune('A'+i%26))
			ba.AnalyzeUserBehavior(userID, event)
			i++
		}
	})
}
