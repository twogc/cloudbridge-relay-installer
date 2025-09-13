package relay

import (
	"github.com/twogc/cloudbridge-relay-installer/internal/ai"
)

// AIServiceInterface определяет интерфейс для AI сервиса
type AIServiceInterface interface {
	Initialize() error
	Start() error
	Stop() error
	GetStatus() *ai.ServiceStatus
}

// Ensure MockAIService implements AIServiceInterface
var _ AIServiceInterface = (*MockAIService)(nil)
