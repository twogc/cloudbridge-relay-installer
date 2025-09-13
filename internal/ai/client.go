package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/twogc/cloudbridge-relay-installer/internal/ai/anomaly"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai/predictive"
	"github.com/twogc/cloudbridge-relay-installer/internal/ai/routing"
	"go.uber.org/zap"
)

// Client представляет клиент для взаимодействия с AI Service
type Client struct {
	baseURL    string
	httpClient *http.Client
	logger     *zap.Logger
	timeout    time.Duration
}

// NewClient создает новый AI клиент
func NewClient(baseURL string, logger *zap.Logger) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger:  logger,
		timeout: 30 * time.Second,
	}
}

// DetectAnomalyRequest запрос на обнаружение аномалий
type DetectAnomalyRequest struct {
	Metrics *anomaly.NetworkMetrics `json:"metrics"`
}

// DetectAnomalyResponse ответ на обнаружение аномалий
type DetectAnomalyResponse struct {
	Result *anomaly.AnomalyResult `json:"result"`
	Error  string                 `json:"error,omitempty"`
}

// DetectAnomaly обнаруживает аномалии в сетевых метриках
func (c *Client) DetectAnomaly(ctx context.Context, metrics *anomaly.NetworkMetrics) (*anomaly.AnomalyResult, error) {
	req := &DetectAnomalyRequest{
		Metrics: metrics,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/ai/anomaly/detect", c.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("AI service returned status %d: %s", resp.StatusCode, string(body))
	}

	var response DetectAnomalyResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if response.Error != "" {
		return nil, fmt.Errorf("AI service error: %s", response.Error)
	}

	return response.Result, nil
}

// PredictLoadRequest запрос на предсказание нагрузки
type PredictLoadRequest struct {
	Horizon int `json:"horizon"`
}

// PredictLoadResponse ответ на предсказание нагрузки
type PredictLoadResponse struct {
	Result *predictive.LoadPrediction `json:"result"`
	Error  string                     `json:"error,omitempty"`
}

// PredictLoad предсказывает нагрузку на заданный горизонт
func (c *Client) PredictLoad(ctx context.Context, horizon int) (*predictive.LoadPrediction, error) {
	req := &PredictLoadRequest{
		Horizon: horizon,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/ai/predict/load", c.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("AI service returned status %d: %s", resp.StatusCode, string(body))
	}

	var response PredictLoadResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if response.Error != "" {
		return nil, fmt.Errorf("AI service error: %s", response.Error)
	}

	return response.Result, nil
}

// OptimizeRouteRequest запрос на оптимизацию маршрута
type OptimizeRouteRequest struct {
	Source      string                 `json:"source"`
	Destination string                 `json:"destination"`
	Constraints map[string]interface{} `json:"constraints"`
}

// OptimizeRouteResponse ответ на оптимизацию маршрута
type OptimizeRouteResponse struct {
	Result *routing.OptimizedRoute `json:"result"`
	Error  string                  `json:"error,omitempty"`
}

// OptimizeRoute оптимизирует маршрут между источником и назначением
func (c *Client) OptimizeRoute(ctx context.Context, source, destination string, constraints map[string]interface{}) (*routing.OptimizedRoute, error) {
	req := &OptimizeRouteRequest{
		Source:      source,
		Destination: destination,
		Constraints: constraints,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/ai/route/optimize", c.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("AI service returned status %d: %s", resp.StatusCode, string(body))
	}

	var response OptimizeRouteResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if response.Error != "" {
		return nil, fmt.Errorf("AI service error: %s", response.Error)
	}

	return response.Result, nil
}

// GetStatus получает статус AI сервиса
func (c *Client) GetStatus(ctx context.Context) (*ClientServiceStatus, error) {
	url := fmt.Sprintf("%s/api/v1/ai/status", c.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("AI service returned status %d: %s", resp.StatusCode, string(body))
	}

	var status ClientServiceStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &status, nil
}

// HealthCheck проверяет здоровье AI сервиса
func (c *Client) HealthCheck(ctx context.Context) error {
	url := fmt.Sprintf("%s/health", c.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("AI service health check failed with status %d", resp.StatusCode)
	}

	return nil
}

// ClientServiceStatus представляет статус AI сервиса для клиента
type ClientServiceStatus struct {
	AnomalyDetector bool `json:"anomaly_detector"`
	DataPipeline    bool `json:"data_pipeline"`
	LoadPredictor   bool `json:"load_predictor"`
	Monitoring      bool `json:"monitoring"`
	RouteOptimizer  bool `json:"route_optimizer"`
}
