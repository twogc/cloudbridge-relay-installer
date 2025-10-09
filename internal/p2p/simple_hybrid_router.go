package p2p

import (
	"go.uber.org/zap"
)

// SimpleHybridRouter простая реализация HybridRouterInterface
type SimpleHybridRouter struct {
	logger *zap.Logger
}

// NewSimpleHybridRouter создает новый простой hybrid router
func NewSimpleHybridRouter(logger *zap.Logger) *SimpleHybridRouter {
	return &SimpleHybridRouter{
		logger: logger,
	}
}

// GetRoutingStats возвращает статистику маршрутизации
func (shr *SimpleHybridRouter) GetRoutingStats() map[string]interface{} {
	return map[string]interface{}{
		"router_type": "simple",
		"status":      "active",
		"routes":      0,
	}
}
