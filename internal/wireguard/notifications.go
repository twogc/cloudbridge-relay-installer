package wireguard

import (
	"sync"

	"go.uber.org/zap"
)

// P2PUnavailableCallback represents a callback function for P2P unavailability
type P2PUnavailableCallback func(reason string, fallback string)

// P2PNotificationManager manages P2P notifications
type P2PNotificationManager struct {
	callbacks []P2PUnavailableCallback
	mu        sync.RWMutex
	logger    *zap.Logger
}

// NewP2PNotificationManager creates a new P2P notification manager
func NewP2PNotificationManager(logger *zap.Logger) *P2PNotificationManager {
	return &P2PNotificationManager{
		callbacks: make([]P2PUnavailableCallback, 0),
		logger:    logger,
	}
}

// RegisterCallback registers a callback for P2P unavailability
func (nm *P2PNotificationManager) RegisterCallback(callback P2PUnavailableCallback) {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	nm.callbacks = append(nm.callbacks, callback)
	nm.logger.Info("P2P notification callback registered")
}

// NotifyP2PUnavailable notifies all registered callbacks about P2P unavailability
func (nm *P2PNotificationManager) NotifyP2PUnavailable(reason, fallback string) {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	nm.logger.Warn("P2P unavailable, notifying callbacks",
		zap.String("reason", reason),
		zap.String("fallback", fallback),
	)

	for _, callback := range nm.callbacks {
		go func(cb P2PUnavailableCallback) {
			defer func() {
				if r := recover(); r != nil {
					nm.logger.Error("P2P notification callback panicked",
						zap.Any("panic", r),
					)
				}
			}()
			cb(reason, fallback)
		}(callback)
	}
}

// GetCallbackCount returns the number of registered callbacks
func (nm *P2PNotificationManager) GetCallbackCount() int {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	return len(nm.callbacks)
}

// ClearCallbacks clears all registered callbacks
func (nm *P2PNotificationManager) ClearCallbacks() {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	nm.callbacks = make([]P2PUnavailableCallback, 0)
	nm.logger.Info("All P2P notification callbacks cleared")
}


