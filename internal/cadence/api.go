package cadence

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// APIHandler обрабатывает HTTP запросы для Cadence workflows
type APIHandler struct {
	client *Client
	logger *zap.Logger
}

// NewAPIHandler создает новый API handler для Cadence
func NewAPIHandler(cadenceClient *Client, logger *zap.Logger) *APIHandler {
	return &APIHandler{
		client: cadenceClient,
		logger: logger,
	}
}

// StartRelayDeploymentWorkflow запускает workflow развертывания relay
func (h *APIHandler) StartRelayDeploymentWorkflow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var input RelayDeploymentInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	workflowRun, err := h.client.ExecuteWorkflow(ctx, "RelayDeploymentWorkflow", input)
	if err != nil {
		h.logger.Error("Failed to start relay deployment workflow", zap.Error(err))
		http.Error(w, "Failed to start workflow", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"workflow_id": workflowRun.GetID(),
		"run_id":      workflowRun.GetRunID(),
		"status":      "started",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// StartCertificateRotationWorkflow запускает workflow ротации сертификатов
func (h *APIHandler) StartCertificateRotationWorkflow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var input CertificateRotationInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	workflowRun, err := h.client.ExecuteWorkflow(ctx, "CertificateRotationWorkflow", input)
	if err != nil {
		h.logger.Error("Failed to start certificate rotation workflow", zap.Error(err))
		http.Error(w, "Failed to start workflow", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"workflow_id": workflowRun.GetID(),
		"run_id":      workflowRun.GetRunID(),
		"status":      "started",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// StartNetworkMaintenanceWorkflow запускает workflow обслуживания сети
func (h *APIHandler) StartNetworkMaintenanceWorkflow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var input NetworkMaintenanceInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	workflowRun, err := h.client.ExecuteWorkflow(ctx, "NetworkMaintenanceWorkflow", input)
	if err != nil {
		h.logger.Error("Failed to start network maintenance workflow", zap.Error(err))
		http.Error(w, "Failed to start workflow", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"workflow_id": workflowRun.GetID(),
		"run_id":      workflowRun.GetRunID(),
		"status":      "started",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// StartSecurityAuditWorkflow запускает workflow аудита безопасности
func (h *APIHandler) StartSecurityAuditWorkflow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var input SecurityAuditInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	workflowRun, err := h.client.ExecuteWorkflow(ctx, "SecurityAuditWorkflow", input)
	if err != nil {
		h.logger.Error("Failed to start security audit workflow", zap.Error(err))
		http.Error(w, "Failed to start workflow", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"workflow_id": workflowRun.GetID(),
		"run_id":      workflowRun.GetRunID(),
		"status":      "started",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// StartBackupWorkflow запускает workflow резервного копирования
func (h *APIHandler) StartBackupWorkflow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var input BackupInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	workflowRun, err := h.client.ExecuteWorkflow(ctx, "BackupWorkflow", input)
	if err != nil {
		h.logger.Error("Failed to start backup workflow", zap.Error(err))
		http.Error(w, "Failed to start workflow", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"workflow_id": workflowRun.GetID(),
		"run_id":      workflowRun.GetRunID(),
		"status":      "started",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetWorkflowStatus получает статус workflow
func (h *APIHandler) GetWorkflowStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	workflowID := r.URL.Query().Get("workflow_id")
	if workflowID == "" {
		http.Error(w, "workflow_id parameter is required", http.StatusBadRequest)
		return
	}

	// Получаем информацию о workflow
	// В реальной реализации здесь должен быть вызов DescribeWorkflowExecution
	// Пока возвращаем базовую информацию
	response := map[string]interface{}{
		"workflow_id": workflowID,
		"run_id":      "",
		"status":      "running",
		"start_time":  time.Now(),
		"close_time":  nil,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HealthCheck проверяет состояние Cadence client
func (h *APIHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := h.client.HealthCheck(ctx)
	if err != nil {
		h.logger.Error("Cadence health check failed", zap.Error(err))
		http.Error(w, "Cadence is not healthy", http.StatusServiceUnavailable)
		return
	}

	response := map[string]interface{}{
		"status": "healthy",
		"timestamp": time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RegisterRoutes регистрирует маршруты для Cadence API
func (h *APIHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/cadence/workflows/relay-deployment", h.StartRelayDeploymentWorkflow)
	mux.HandleFunc("/api/v1/cadence/workflows/certificate-rotation", h.StartCertificateRotationWorkflow)
	mux.HandleFunc("/api/v1/cadence/workflows/network-maintenance", h.StartNetworkMaintenanceWorkflow)
	mux.HandleFunc("/api/v1/cadence/workflows/security-audit", h.StartSecurityAuditWorkflow)
	mux.HandleFunc("/api/v1/cadence/workflows/backup", h.StartBackupWorkflow)
	mux.HandleFunc("/api/v1/cadence/workflows/status", h.GetWorkflowStatus)
	mux.HandleFunc("/api/v1/cadence/health", h.HealthCheck)
}
