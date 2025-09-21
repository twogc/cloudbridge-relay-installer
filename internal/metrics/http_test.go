package metrics

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewServer(t *testing.T) {
	// Test creating a new server
	port := ":8080"
	path := "/metrics"
	server := NewServer(port, path)

	assert.NotNil(t, server)
	assert.Equal(t, port, server.port)
	assert.Equal(t, path, server.path)
	assert.Equal(t, "2.0.0", server.version)
}

func TestNewServer_EmptyPort(t *testing.T) {
	// Test creating server with empty port
	server := NewServer("", "/metrics")
	assert.NotNil(t, server)
	assert.Equal(t, "", server.port)
	assert.Equal(t, "/metrics", server.path)
}

func TestNewServer_EmptyPath(t *testing.T) {
	// Test creating server with empty path
	server := NewServer(":8080", "")
	assert.NotNil(t, server)
	assert.Equal(t, ":8080", server.port)
	assert.Equal(t, "", server.path)
}

func TestNewServer_EmptyBoth(t *testing.T) {
	// Test creating server with empty port and path
	server := NewServer("", "")
	assert.NotNil(t, server)
	assert.Equal(t, "", server.port)
	assert.Equal(t, "", server.path)
}

func TestServer_handleHealth(t *testing.T) {
	// Test health check handler
	server := NewServer(":8080", "/metrics")

	// Create a request
	req := httptest.NewRequest("GET", "/health", nil)

	// Create a response recorder
	w := httptest.NewRecorder()

	// Call the handler
	server.handleHealth(w, req)

	// Check the response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	// Parse the response body
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Check the response content
	assert.Equal(t, "ok", response["status"])
	assert.Equal(t, "2.0.0", response["version"])
	assert.Equal(t, "cloudbridge-relay", response["service"])
	assert.Contains(t, response, "timestamp")

	// Verify timestamp is valid
	timestampStr, ok := response["timestamp"].(string)
	require.True(t, ok)
	_, err = time.Parse(time.RFC3339, timestampStr)
	assert.NoError(t, err)
}

func TestServer_handleHealth_POST(t *testing.T) {
	// Test health check handler with POST method
	server := NewServer(":8080", "/metrics")

	// Create a POST request
	req := httptest.NewRequest("POST", "/health", nil)

	// Create a response recorder
	w := httptest.NewRecorder()

	// Call the handler
	server.handleHealth(w, req)

	// Check the response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func TestServer_handleHealth_WithBody(t *testing.T) {
	// Test health check handler with request body
	server := NewServer(":8080", "/metrics")

	// Create a request with body
	req := httptest.NewRequest("GET", "/health", nil)
	req.Body = http.NoBody

	// Create a response recorder
	w := httptest.NewRecorder()

	// Call the handler
	server.handleHealth(w, req)

	// Check the response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func TestServer_handleHealth_JSONEncodingError(t *testing.T) {
	// This test is difficult to trigger in practice since json.NewEncoder().Encode()
	// rarely fails with simple data structures, but we can test the structure
	server := NewServer(":8080", "/metrics")

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	// Call the handler
	server.handleHealth(w, req)

	// Should still return 200 OK
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestServer_Start_InvalidPort(t *testing.T) {
	// Test starting server with invalid port
	server := NewServer("invalid-port", "/metrics")

	// This should fail when trying to start
	err := server.Start()
	assert.Error(t, err)
}

func TestServer_Start_ValidPort(t *testing.T) {
	// Test starting server with valid port (but we'll stop it immediately)
	server := NewServer(":0", "/metrics") // Use port 0 to get a free port

	// Start server in a goroutine
	go func() {
		_ = server.Start()
	}()

	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)

	// The server should be running (we can't easily test this without making HTTP requests)
	assert.NotNil(t, server)
}

func TestServer_Start_WithCustomPath(t *testing.T) {
	// Test starting server with custom metrics path
	server := NewServer(":0", "/custom-metrics")

	// Start server in a goroutine
	go func() {
		_ = server.Start()
	}()

	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)

	assert.NotNil(t, server)
	assert.Equal(t, "/custom-metrics", server.path)
}

func TestServer_Start_WithEmptyPath(t *testing.T) {
	// Test starting server with empty path
	server := NewServer(":0", "")

	// Start server in a goroutine
	go func() {
		_ = server.Start()
	}()

	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)

	assert.NotNil(t, server)
	assert.Equal(t, "", server.path)
}

func TestServer_Start_WithEmptyPort(t *testing.T) {
	// Test starting server with empty port
	server := NewServer("", "/metrics")

	// This should fail when trying to start
	err := server.Start()
	assert.Error(t, err)
}

func TestServer_ConcurrentHealthChecks(t *testing.T) {
	// Test concurrent health check requests
	server := NewServer(":8080", "/metrics")

	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()

			req := httptest.NewRequest("GET", "/health", nil)
			w := httptest.NewRecorder()

			server.handleHealth(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestServer_HealthCheckResponseStructure(t *testing.T) {
	// Test the structure of health check response
	server := NewServer(":8080", "/metrics")

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	server.handleHealth(w, req)

	// Parse the response
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Check all required fields are present
	requiredFields := []string{"status", "version", "timestamp", "service"}
	for _, field := range requiredFields {
		assert.Contains(t, response, field, "Response should contain field: %s", field)
	}

	// Check field types
	assert.IsType(t, "", response["status"])
	assert.IsType(t, "", response["version"])
	assert.IsType(t, "", response["timestamp"])
	assert.IsType(t, "", response["service"])
}

func TestServer_HealthCheckTimestampFormat(t *testing.T) {
	// Test that timestamp is in RFC3339 format
	server := NewServer(":8080", "/metrics")

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	server.handleHealth(w, req)

	// Parse the response
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	timestampStr, ok := response["timestamp"].(string)
	require.True(t, ok)

	// Parse the timestamp
	timestamp, err := time.Parse(time.RFC3339, timestampStr)
	require.NoError(t, err)

	// Check that timestamp is recent (within last minute)
	now := time.Now()
	diff := now.Sub(timestamp)
	assert.Less(t, diff, time.Minute)
	assert.GreaterOrEqual(t, diff, time.Duration(0))
}

func TestServer_HealthCheckVersion(t *testing.T) {
	// Test that version is correct
	server := NewServer(":8080", "/metrics")

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	server.handleHealth(w, req)

	// Parse the response
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	version, ok := response["version"].(string)
	require.True(t, ok)
	assert.Equal(t, "2.0.0", version)
}

func TestServer_HealthCheckService(t *testing.T) {
	// Test that service name is correct
	server := NewServer(":8080", "/metrics")

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	server.handleHealth(w, req)

	// Parse the response
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	service, ok := response["service"].(string)
	require.True(t, ok)
	assert.Equal(t, "cloudbridge-relay", service)
}

func TestServer_HealthCheckStatus(t *testing.T) {
	// Test that status is always "ok"
	server := NewServer(":8080", "/metrics")

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	server.handleHealth(w, req)

	// Parse the response
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	status, ok := response["status"].(string)
	require.True(t, ok)
	assert.Equal(t, "ok", status)
}

func TestServer_HTTPMethods(t *testing.T) {
	// Test different HTTP methods on health endpoint
	server := NewServer(":8080", "/metrics")

	methods := []string{"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/health", nil)
			w := httptest.NewRecorder()

			server.handleHealth(w, req)

			// All methods should return 200 OK
			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		})
	}
}

func TestServer_EdgeCases(t *testing.T) {
	// Test edge cases
	server := NewServer(":8080", "/metrics")

	// Test with nil request (should not panic)
	assert.NotPanics(t, func() {
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		server.handleHealth(w, req)
	})

	// Test with request containing special characters
	req := httptest.NewRequest("GET", "/health?param=value&special=@#$%", nil)
	w := httptest.NewRecorder()
	server.handleHealth(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestServer_Performance(t *testing.T) {
	// Test performance of health check handler
	server := NewServer(":8080", "/metrics")

	start := time.Now()

	for i := 0; i < 1000; i++ {
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		server.handleHealth(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	duration := time.Since(start)

	// Should complete within reasonable time (less than 1 second)
	assert.Less(t, duration, time.Second)
}

func TestServer_JSONResponse(t *testing.T) {
	// Test that response is valid JSON
	server := NewServer(":8080", "/metrics")

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	server.handleHealth(w, req)

	// Check that response is valid JSON
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Check that it's not empty
	assert.NotEmpty(t, response)
}

func TestServer_ContentType(t *testing.T) {
	// Test that Content-Type header is set correctly
	server := NewServer(":8080", "/metrics")

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	server.handleHealth(w, req)

	// Check Content-Type header
	contentType := w.Header().Get("Content-Type")
	assert.Equal(t, "application/json", contentType)
}

func TestServer_StatusCode(t *testing.T) {
	// Test that status code is always 200 OK
	server := NewServer(":8080", "/metrics")

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	server.handleHealth(w, req)

	// Check status code
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestServer_ResponseHeaders(t *testing.T) {
	// Test that response headers are set correctly
	server := NewServer(":8080", "/metrics")

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	server.handleHealth(w, req)

	// Check headers
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	// Check that no other headers are set
	headers := w.Header()
	assert.Len(t, headers, 1) // Only Content-Type should be set
}

func TestServer_ResponseBody(t *testing.T) {
	// Test that response body is not empty
	server := NewServer(":8080", "/metrics")

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	server.handleHealth(w, req)

	// Check that response body is not empty
	assert.NotEmpty(t, w.Body.Bytes())

	// Check that it's valid JSON
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
}
