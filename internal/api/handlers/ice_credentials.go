package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/twogc/cloudbridge-relay-installer/internal/nat"
)

// ICECredentialsHandler handles ICE credentials requests from clients
type ICECredentialsHandler struct {
	logger          *zap.Logger
	turnIntegration *nat.TURNIntegration
}

// NewICECredentialsHandler creates a new ICE credentials handler
func NewICECredentialsHandler(logger *zap.Logger, turnIntegration *nat.TURNIntegration) *ICECredentialsHandler {
	return &ICECredentialsHandler{
		logger:          logger,
		turnIntegration: turnIntegration,
	}
}

// ICECredentialsRequest represents the request for ICE credentials
type ICECredentialsRequest struct {
	PeerID string `json:"peer_id,omitempty"`
}

// ICECredentialsResponse represents the response with ICE credentials
type ICECredentialsResponse struct {
	Success     bool                     `json:"success"`
	TenantID    string                   `json:"tenant_id,omitempty"`
	PeerID      string                   `json:"peer_id,omitempty"`
	ICEServers  []ICEServer              `json:"ice_servers"`
	TTL         int                      `json:"ttl,omitempty"` // Time-to-live in seconds
	ExpiresAt   string                   `json:"expires_at,omitempty"`
	Credentials *TURNCredentials         `json:"turn_credentials,omitempty"`
	Policy      string                   `json:"ice_transport_policy,omitempty"`
}

// ICEServer represents an ICE server configuration
type ICEServer struct {
	URLs       interface{} `json:"urls"` // Can be string or []string
	Username   string      `json:"username,omitempty"`
	Credential string      `json:"credential,omitempty"`
	CredentialType string  `json:"credentialType,omitempty"`
}

// TURNCredentials represents TURN server credentials
type TURNCredentials struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	TTL        int    `json:"ttl"`
	URIs       []string `json:"uris"`
}

// Handle processes ICE credentials requests
func (h *ICECredentialsHandler) Handle(c *gin.Context) {
	// Extract tenant_id from context (set by compatibility layer)
	tenantID, _ := c.Get("tenant_id")
	tenantIDStr, _ := tenantID.(string)

	// Extract peer_id from context, request body, or query param
	peerID := h.extractPeerID(c)

	h.logger.Info("ICE credentials request",
		zap.String("tenant_id", tenantIDStr),
		zap.String("peer_id", peerID),
		zap.String("remote_addr", c.ClientIP()))

	// Get ICE servers configuration
	iceServersConfig := h.turnIntegration.GetICEServersForClient()

	// Convert to response format
	response := &ICECredentialsResponse{
		Success:  true,
		TenantID: tenantIDStr,
		PeerID:   peerID,
		Policy:   "all", // Use both STUN and TURN
		TTL:      3600,  // 1 hour
	}

	// Convert ICE servers
	if iceServers, ok := iceServersConfig["iceServers"].([]map[string]interface{}); ok {
		response.ICEServers = make([]ICEServer, 0, len(iceServers))
		for _, server := range iceServers {
			iceServer := ICEServer{
				URLs: server["urls"],
			}
			if username, ok := server["username"].(string); ok {
				iceServer.Username = username
			}
			if credential, ok := server["credential"].(string); ok {
				iceServer.Credential = credential
				iceServer.CredentialType = "password"
			}
			response.ICEServers = append(response.ICEServers, iceServer)
		}
	}

	// Generate peer-specific TURN credentials if peer_id is available
	if peerID != "" {
		username, password, err := h.turnIntegration.GetTURNCredentialsForPeer(peerID)
		if err != nil {
			h.logger.Warn("Failed to generate TURN credentials",
				zap.String("peer_id", peerID),
				zap.Error(err))
		} else {
			response.Credentials = &TURNCredentials{
				Username: username,
				Password: password,
				TTL:      3600,
				URIs:     h.getTURNURIs(),
			}
		}
	}

	h.logger.Debug("Returning ICE credentials",
		zap.Int("ice_servers_count", len(response.ICEServers)),
		zap.Bool("has_turn_credentials", response.Credentials != nil))

	c.JSON(http.StatusOK, response)
}

// extractPeerID extracts peer_id from various sources
func (h *ICECredentialsHandler) extractPeerID(c *gin.Context) string {
	// Try context first (set by auth middleware)
	if peerID, exists := c.Get("peer_id"); exists {
		if peerIDStr, ok := peerID.(string); ok {
			return peerIDStr
		}
	}

	// Try JWT claims
	if claims, exists := c.Get("claims"); exists {
		if claimsMap, ok := claims.(map[string]interface{}); ok {
			if peerID, ok := claimsMap["peer_id"].(string); ok {
				return peerID
			}
			if sub, ok := claimsMap["sub"].(string); ok {
				return sub
			}
		}
	}

	// Try request body
	var req ICECredentialsRequest
	if err := c.ShouldBindJSON(&req); err == nil && req.PeerID != "" {
		return req.PeerID
	}

	// Try query parameter
	if peerID := c.Query("peer_id"); peerID != "" {
		return peerID
	}

	// Try header
	if peerID := c.GetHeader("X-Peer-ID"); peerID != "" {
		return peerID
	}

	return ""
}

// getTURNURIs returns list of TURN server URIs
func (h *ICECredentialsHandler) getTURNURIs() []string {
	turnServers := h.turnIntegration.GetTURNServers()
	uris := make([]string, 0, len(turnServers))
	for _, server := range turnServers {
		uris = append(uris, server.URLs)
	}
	return uris
}

// HandleHealthCheck returns health status of ICE credentials service
func (h *ICECredentialsHandler) HandleHealthCheck(c *gin.Context) {
	health := h.turnIntegration.HealthCheck()

	status := http.StatusOK
	if health["status"] == "unhealthy" {
		status = http.StatusServiceUnavailable
	} else if health["status"] == "degraded" {
		status = http.StatusOK // Still return 200 but with degraded status
	}

	c.JSON(status, health)
}
