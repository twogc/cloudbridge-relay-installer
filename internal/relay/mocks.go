package relay

import (
	"github.com/stretchr/testify/mock"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai"
)

// MockAIService представляет мок AI сервиса для тестирования
type MockAIService struct {
	mock.Mock
}

func (m *MockAIService) Initialize() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockAIService) Start() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockAIService) Stop() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockAIService) GetStatus() *ai.ServiceStatus {
	args := m.Called()
	if args.Get(0) == nil {
		return &ai.ServiceStatus{
			Components: map[string]bool{
				"anomaly_detector": false,
				"load_predictor":   false,
				"route_optimizer":  false,
			},
		}
	}
	return args.Get(0).(*ai.ServiceStatus)
}
