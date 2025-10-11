package sync

import (
	"context"
	"testing"
	"time"

	"github.com/twogc/cloudbridge-relay-installer/internal/tenant"
	"go.uber.org/zap"
)

// MockZitadelClient для тестирования
type MockZitadelClient struct {
	tenants []*MockTenant
}

type MockTenant struct {
	ID          string
	Name        string
	Quotas      *MockQuotas
}

type MockQuotas struct {
	MaxTunnels     int
	MaxConnections int
	BandwidthLimit string
}

type MockTenantManager struct {
	tenants map[string]*MockTenant
}

func (m *MockTenantManager) ListTenants(ctx context.Context) ([]*MockTenant, error) {
	var result []*MockTenant
	for _, tenant := range m.tenants {
		result = append(result, tenant)
	}
	return result, nil
}

func TestNewTenantSync(t *testing.T) {
	logger := zap.NewNop()
	
	// Создаем mock клиенты
	mockZitadelClient := &MockZitadelClient{}
	mockRelayManager := &tenant.Manager{}
	
	sync := NewTenantSync(mockZitadelClient, mockRelayManager, logger)
	
	if sync == nil {
		t.Fatal("TenantSync is nil")
	}
	
	if sync.logger != logger {
		t.Error("Logger not set correctly")
	}
	
	if sync.zitadelClient != mockZitadelClient {
		t.Error("ZitadelClient not set correctly")
	}
	
	if sync.relayManager != mockRelayManager {
		t.Error("RelayManager not set correctly")
	}
}

func TestTenantSyncStartStop(t *testing.T) {
	logger := zap.NewNop()
	
	mockZitadelClient := &MockZitadelClient{}
	mockRelayManager := &tenant.Manager{}
	
	sync := NewTenantSync(mockZitadelClient, mockRelayManager, logger)
	
	// Тест что изначально не запущен
	if sync.IsRunning() {
		t.Error("Expected sync to not be running initially")
	}
	
	// Запускаем синхронизацию
	sync.Start(1 * time.Second)
	
	// Ждем немного
	time.Sleep(100 * time.Millisecond)
	
	// Останавливаем
	sync.Stop()
	
	// Проверяем что остановлен
	if sync.IsRunning() {
		t.Error("Expected sync to be stopped")
	}
}

func TestParseBandwidthLimit(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"1GB", 1000000000.0},
		{"500MB", 500000000.0},
		{"100MB", 100000000.0},
		{"", 1000000.0},
		{"unknown", 1000000.0},
	}
	
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseBandwidthLimit(tt.input)
			if result != tt.expected {
				t.Errorf("parseBandwidthLimit(%s) = %f, want %f", tt.input, result, tt.expected)
			}
		})
	}
}

func TestTenantSyncStats(t *testing.T) {
	logger := zap.NewNop()
	
	mockZitadelClient := &MockZitadelClient{}
	mockRelayManager := &tenant.Manager{}
	
	sync := NewTenantSync(mockZitadelClient, mockRelayManager, logger)
	
	// Получаем статистику
	stats := sync.GetStats()
	
	if stats == nil {
		t.Fatal("Stats should not be nil")
	}
	
	// Проверяем что есть нужные поля
	expectedFields := []string{"running", "last_sync_time", "sync_errors", "tenants_processed", "limits_changed"}
	for _, field := range expectedFields {
		if _, exists := stats[field]; !exists {
			t.Errorf("Expected field %s in stats", field)
		}
	}
}


