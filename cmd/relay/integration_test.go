// +build integration

package main

import (
	"context"
	"net/http"
	"testing"
	"time"
)

// TestRelayStartup проверяет что relay сервер может запуститься и отвечает на health check
func TestRelayStartup(t *testing.T) {
	// Этот тест требует running relay server
	// Запустите relay перед тестом: go run cmd/relay/main.go

	tests := []struct {
		name     string
		endpoint string
		want     int
	}{
		{
			name:     "Health check",
			endpoint: "http://localhost:8080/health",
			want:     http.StatusOK,
		},
		{
			name:     "Metrics endpoint",
			endpoint: "http://localhost:8080/metrics",
			want:     http.StatusOK,
		},
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			req, err := http.NewRequestWithContext(ctx, "GET", tt.endpoint, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			resp, err := client.Do(req)
			if err != nil {
				t.Skipf("Relay server not running: %v (run: go run cmd/relay/main.go)", err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.want {
				t.Errorf("Expected status %d, got %d", tt.want, resp.StatusCode)
			}
		})
	}
}

// TestP2PAPIEndpoints проверяет P2P API endpoints
func TestP2PAPIEndpoints(t *testing.T) {
	// Требует running relay server с P2P API

	tests := []struct {
		name     string
		endpoint string
		method   string
		want     int
	}{
		{
			name:     "GET /peers",
			endpoint: "http://localhost:8082/peers",
			method:   "GET",
			want:     http.StatusOK,
		},
		{
			name:     "GET /health",
			endpoint: "http://localhost:8082/health",
			method:   "GET",
			want:     http.StatusOK,
		},
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			req, err := http.NewRequestWithContext(ctx, tt.method, tt.endpoint, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			resp, err := client.Do(req)
			if err != nil {
				t.Skipf("P2P API not running: %v", err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.want {
				t.Errorf("Expected status %d, got %d", tt.want, resp.StatusCode)
			}
		})
	}
}
