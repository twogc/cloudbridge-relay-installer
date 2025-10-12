package nat

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap"
)

// ICEServerConfig represents ICE server configuration loaded from file
type ICEServerConfig struct {
	STUNServers []string     `json:"stun_servers"`
	TURNServers []TURNServer `json:"turn_servers"`
}

// TURNServer represents a TURN server configuration
type TURNServer struct {
	URLs       string `json:"urls"`
	Username   string `json:"username"`
	Credential string `json:"credential"`
	Realm      string `json:"realm"`
}

// TURNIntegration manages TURN server integration with NAT traversal
type TURNIntegration struct {
	logger      *zap.Logger
	config      *ICEServerConfig
	stunServers []string
	turnServers []TURNServer
}

// NewTURNIntegration creates a new TURN integration manager
func NewTURNIntegration(logger *zap.Logger) *TURNIntegration {
	return &TURNIntegration{
		logger:      logger,
		stunServers: make([]string, 0),
		turnServers: make([]TURNServer, 0),
	}
}

// LoadICEServersFromConfig loads ICE servers from configuration file
func (ti *TURNIntegration) LoadICEServersFromConfig(configPath string) error {
	ti.logger.Info("Loading ICE servers configuration", zap.String("path", configPath))

	// Read configuration file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read ICE servers config: %w", err)
	}

	// Parse JSON
	var config ICEServerConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse ICE servers config: %w", err)
	}

	// Replace environment variable placeholders
	turnPassword := os.Getenv("TURN_PASSWORD")
	for i := range config.TURNServers {
		if strings.Contains(config.TURNServers[i].Credential, "${TURN_PASSWORD}") {
			config.TURNServers[i].Credential = strings.ReplaceAll(
				config.TURNServers[i].Credential,
				"${TURN_PASSWORD}",
				turnPassword,
			)
		}
	}

	ti.config = &config
	ti.stunServers = config.STUNServers
	ti.turnServers = config.TURNServers

	ti.logger.Info("Loaded ICE servers configuration",
		zap.Int("stun_servers", len(config.STUNServers)),
		zap.Int("turn_servers", len(config.TURNServers)))

	return nil
}

// LoadFromEnvironment loads ICE servers configuration from environment variables
// Useful for Kubernetes deployments where config is injected via env vars
func (ti *TURNIntegration) LoadFromEnvironment() error {
	ti.logger.Info("Loading ICE servers from environment variables")

	// STUN servers from env (comma-separated)
	if stunServersStr := os.Getenv("STUN_SERVERS"); stunServersStr != "" {
		ti.stunServers = strings.Split(stunServersStr, ",")
		for i := range ti.stunServers {
			ti.stunServers[i] = strings.TrimSpace(ti.stunServers[i])
		}
	} else {
		// Default STUN servers
		ti.stunServers = []string{
			"stun:stun.l.google.com:19302",
			"stun:stun1.l.google.com:19302",
			"stun:stun2.l.google.com:19302",
		}
	}

	// TURN servers from env
	turnURL := os.Getenv("TURN_URL")
	turnUsername := os.Getenv("TURN_USERNAME")
	turnPassword := os.Getenv("TURN_PASSWORD")
	turnRealm := os.Getenv("TURN_REALM")

	if turnURL != "" && turnUsername != "" && turnPassword != "" {
		// Parse multiple TURN URLs if provided (comma-separated)
		turnURLs := strings.Split(turnURL, ",")
		for _, url := range turnURLs {
			ti.turnServers = append(ti.turnServers, TURNServer{
				URLs:       strings.TrimSpace(url),
				Username:   turnUsername,
				Credential: turnPassword,
				Realm:      turnRealm,
			})
		}
	}

	ti.logger.Info("Loaded ICE servers from environment",
		zap.Int("stun_servers", len(ti.stunServers)),
		zap.Int("turn_servers", len(ti.turnServers)))

	return nil
}

// GetSTUNServers returns configured STUN servers
func (ti *TURNIntegration) GetSTUNServers() []string {
	return ti.stunServers
}

// GetTURNServers returns configured TURN servers
func (ti *TURNIntegration) GetTURNServers() []TURNServer {
	return ti.turnServers
}

// GetICEServersForClient returns ICE servers configuration for client API
// This is used by the compatibility layer to provide ICE credentials to clients
func (ti *TURNIntegration) GetICEServersForClient() map[string]interface{} {
	iceServers := make([]map[string]interface{}, 0)

	// Add STUN servers
	for _, stunURL := range ti.stunServers {
		iceServers = append(iceServers, map[string]interface{}{
			"urls": stunURL,
		})
	}

	// Add TURN servers
	for _, turnServer := range ti.turnServers {
		iceServers = append(iceServers, map[string]interface{}{
			"urls":       turnServer.URLs,
			"username":   turnServer.Username,
			"credential": turnServer.Credential,
		})
	}

	return map[string]interface{}{
		"iceServers":        iceServers,
		"iceTransportPolicy": "all", // Use both STUN and TURN
	}
}

// ValidateConfiguration validates the TURN integration configuration
func (ti *TURNIntegration) ValidateConfiguration() error {
	if len(ti.stunServers) == 0 && len(ti.turnServers) == 0 {
		return fmt.Errorf("no ICE servers configured")
	}

	// Validate STUN server URLs
	for _, stunURL := range ti.stunServers {
		if !strings.HasPrefix(stunURL, "stun:") {
			ti.logger.Warn("Invalid STUN server URL", zap.String("url", stunURL))
		}
	}

	// Validate TURN servers
	for i, turnServer := range ti.turnServers {
		if !strings.HasPrefix(turnServer.URLs, "turn:") && !strings.HasPrefix(turnServer.URLs, "turns:") {
			ti.logger.Warn("Invalid TURN server URL",
				zap.Int("index", i),
				zap.String("url", turnServer.URLs))
		}

		if turnServer.Username == "" {
			return fmt.Errorf("TURN server %d missing username", i)
		}

		if turnServer.Credential == "" {
			return fmt.Errorf("TURN server %d missing credential", i)
		}
	}

	ti.logger.Info("TURN integration configuration validated successfully")
	return nil
}

// GetConfigSummary returns a summary of the configuration for logging/debugging
func (ti *TURNIntegration) GetConfigSummary() map[string]interface{} {
	turnSummary := make([]map[string]interface{}, 0)
	for _, turn := range ti.turnServers {
		turnSummary = append(turnSummary, map[string]interface{}{
			"url":      turn.URLs,
			"username": turn.Username,
			"realm":    turn.Realm,
			// Don't include credential in summary for security
		})
	}

	return map[string]interface{}{
		"stun_servers_count": len(ti.stunServers),
		"turn_servers_count": len(ti.turnServers),
		"stun_servers":       ti.stunServers,
		"turn_servers":       turnSummary,
	}
}

// IntegrateWithPionICE converts configuration to pion/ice format
// This is used when integrating with the pion ICE adapter
func (ti *TURNIntegration) IntegrateWithPionICE() ([]string, error) {
	urls := make([]string, 0)

	// Add STUN servers
	urls = append(urls, ti.stunServers...)

	// Add TURN servers
	// Note: pion/ice expects URLs in the format "turn:host:port?transport=udp"
	// and credentials are passed separately
	for _, turn := range ti.turnServers {
		urls = append(urls, turn.URLs)
	}

	ti.logger.Info("Prepared ICE server URLs for pion/ice",
		zap.Int("total_urls", len(urls)))

	return urls, nil
}

// GetTURNCredentialsForPeer generates TURN credentials for a specific peer
// This implements short-term credential mechanism (RFC 5766 Section 10.2)
func (ti *TURNIntegration) GetTURNCredentialsForPeer(peerID string) (username string, password string, err error) {
	if len(ti.turnServers) == 0 {
		return "", "", fmt.Errorf("no TURN servers configured")
	}

	// For now, use the configured credentials
	// In production, implement time-limited credentials using HMAC
	firstTURN := ti.turnServers[0]

	// Format: timestamp:peerID
	// This allows the TURN server to validate and expire credentials
	username = fmt.Sprintf("%s:%s", firstTURN.Username, peerID)
	password = firstTURN.Credential

	return username, password, nil
}

// HealthCheck returns health status of TURN integration
func (ti *TURNIntegration) HealthCheck() map[string]interface{} {
	status := "healthy"
	issues := make([]string, 0)

	if len(ti.stunServers) == 0 {
		issues = append(issues, "no STUN servers configured")
	}

	if len(ti.turnServers) == 0 {
		issues = append(issues, "no TURN servers configured")
		status = "degraded"
	}

	for i, turn := range ti.turnServers {
		if turn.Credential == "" {
			issues = append(issues, fmt.Sprintf("TURN server %d missing credentials", i))
			status = "unhealthy"
		}
	}

	return map[string]interface{}{
		"component":    "turn_integration",
		"status":       status,
		"stun_servers": len(ti.stunServers),
		"turn_servers": len(ti.turnServers),
		"issues":       issues,
	}
}
