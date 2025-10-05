package api

import (
	"go.uber.org/zap"
)

// APIRouter основной роутер API
type APIRouter struct {
	logger *zap.Logger
}

// NewAPIRouter создает новый APIRouter
func NewAPIRouter(logger *zap.Logger) *APIRouter {
	return &APIRouter{
		logger: logger,
	}
}
