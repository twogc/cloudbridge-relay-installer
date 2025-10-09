package sentry

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// SentryAPI клиент для работы с Sentry API
type SentryAPI struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// SentryProject проект в Sentry
type SentryProject struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Platform    string `json:"platform"`
	DateCreated string `json:"dateCreated"`
}

// SentryIssue проблема в Sentry
type SentryIssue struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Level       string    `json:"level"`
	Status      string    `json:"status"`
	FirstSeen   time.Time `json:"firstSeen"`
	LastSeen    time.Time `json:"lastSeen"`
	Count       int       `json:"count"`
	UserCount   int       `json:"userCount"`
	Culprit     string    `json:"culprit"`
	Permalink   string    `json:"permalink"`
}

// SentryEvent событие в Sentry
type SentryEvent struct {
	ID        string                 `json:"id"`
	Title     string                 `json:"title"`
	Message   string                 `json:"message"`
	Level     string                 `json:"level"`
	Platform  string                 `json:"platform"`
	Timestamp time.Time              `json:"timestamp"`
	Tags      map[string]string      `json:"tags"`
	Context   map[string]interface{} `json:"context"`
	User      map[string]interface{} `json:"user"`
}

// NewSentryAPI создает новый клиент Sentry API
func NewSentryAPI(baseURL, token string) *SentryAPI {
	return &SentryAPI{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetProjects получает список проектов
func (api *SentryAPI) GetProjects(ctx context.Context) ([]SentryProject, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/0/projects/", api.baseURL), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", api.token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := api.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	var projects []SentryProject
	if err := json.NewDecoder(resp.Body).Decode(&projects); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return projects, nil
}

// GetIssues получает список проблем для проекта
func (api *SentryAPI) GetIssues(ctx context.Context, orgSlug, projectSlug string) ([]SentryIssue, error) {
	url := fmt.Sprintf("%s/api/0/projects/%s/%s/issues/", api.baseURL, orgSlug, projectSlug)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", api.token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := api.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	var issues []SentryIssue
	if err := json.NewDecoder(resp.Body).Decode(&issues); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return issues, nil
}

// GetIssueEvents получает события для конкретной проблемы
func (api *SentryAPI) GetIssueEvents(ctx context.Context, orgSlug, projectSlug, issueID string) ([]SentryEvent, error) {
	url := fmt.Sprintf("%s/api/0/projects/%s/%s/issues/%s/events/", api.baseURL, orgSlug, projectSlug, issueID)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", api.token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := api.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	var events []SentryEvent
	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return events, nil
}

// CreateIssueNote создает заметку к проблеме
func (api *SentryAPI) CreateIssueNote(ctx context.Context, orgSlug, projectSlug, issueID, text string) error {
	url := fmt.Sprintf("%s/api/0/projects/%s/%s/issues/%s/notes/", api.baseURL, orgSlug, projectSlug, issueID)
	
	data := map[string]string{
		"text": text,
	}
	
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", api.token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := api.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status: %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}

// UpdateIssueStatus обновляет статус проблемы
func (api *SentryAPI) UpdateIssueStatus(ctx context.Context, orgSlug, projectSlug, issueID, status string) error {
	url := fmt.Sprintf("%s/api/0/projects/%s/%s/issues/%s/", api.baseURL, orgSlug, projectSlug, issueID)
	
	data := map[string]string{
		"status": status,
	}
	
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", api.token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := api.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status: %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetProjectStats получает статистику проекта
func (api *SentryAPI) GetProjectStats(ctx context.Context, orgSlug, projectSlug string, since, until time.Time) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/api/0/projects/%s/%s/stats/", api.baseURL, orgSlug, projectSlug)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Добавляем параметры времени
	q := req.URL.Query()
	q.Set("since", since.Format(time.RFC3339))
	q.Set("until", until.Format(time.RFC3339))
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", api.token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := api.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	var stats map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return stats, nil
}



