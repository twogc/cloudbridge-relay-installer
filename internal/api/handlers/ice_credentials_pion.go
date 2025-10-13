package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/twogc/cloudbridge-relay-installer/internal/nat"
)

// ICECredentialsHandlerPion handles ICE credentials using pion/turn standards (RFC 5766)
// This replaces the legacy ice_credentials.go with standards-compliant implementation
type ICECredentialsHandlerPion struct {
	logger                *zap.Logger
	turnIntegration       *nat.TURNIntegration       // For STUN servers
	turnCredentialService *nat.TURNCredentialService // RFC 5766 TURN credentials
	defaultTTL            time.Duration
}

// NewICECredentialsHandlerPion creates a new RFC 5766 compliant handler
func NewICECredentialsHandlerPion(
	logger *zap.Logger,
	turnIntegration *nat.TURNIntegration,
	turnCredentialService *nat.TURNCredentialService,
) *ICECredentialsHandlerPion {
	return &ICECredentialsHandlerPion{
		logger:                logger,
		turnIntegration:       turnIntegration,
		turnCredentialService: turnCredentialService,
		defaultTTL:            10 * time.Minute,
	}
}

// ICECredentialsResponsePion represents RFC 5766 compliant response
type ICECredentialsResponsePion struct {
	Success      bool          `json:"success"`
	TenantID     string        `json:"tenant_id,omitempty"`
	PeerID       string        `json:"peer_id,omitempty"`
	ICEServers   []ICEServer   `json:"iceServers"`              // WebRTC RTCIceServer format
	TTL          int           `json:"ttl"`                     // Seconds until credentials expire
	ExpiresAt    string        `json:"expiresAt"`               // ISO 8601 timestamp
	Policy       string        `json:"iceTransportPolicy"`      // "all" | "relay"
	QuotaUsed    int           `json:"quotaUsed,omitempty"`     // Current TURN allocations
	QuotaLimit   int           `json:"quotaLimit,omitempty"`    // Max TURN allocations
	Standard     string        `json:"standard"`                // "RFC 5766"
	Implementation string      `json:"implementation"`          // "pion/turn v3"
}

// Handle processes ICE credentials requests with RFC 5766 compliance
func (h *ICECredentialsHandlerPion) Handle(c *gin.Context) {
	// Extract tenant_id from context (set by compatibility layer)
	tenantID, _ := c.Get("tenant_id")
	tenantIDStr, _ := tenantID.(string)

	// Extract peer_id from context, request body, or query param
	peerID := h.extractPeerID(c)

	// Parse requested TTL from query param (optional)
	requestedTTL := h.parseTTL(c)

	h.logger.Info("ICE credentials request (RFC 5766)",
		zap.String("tenant_id", tenantIDStr),
		zap.String("peer_id", peerID),
		zap.Duration("requested_ttl", requestedTTL),
		zap.String("remote_addr", c.ClientIP()))

	// Generate short-lived TURN credentials using pion/turn standards
	turnCred, err := h.turnCredentialService.GenerateCredential(
		tenantIDStr,
		peerID,
		requestedTTL,
	)
	if err != nil {
		h.logger.Error("Failed to generate TURN credentials",
			zap.String("tenant_id", tenantIDStr),
			zap.String("peer_id", peerID),
			zap.Error(err))

		c.JSON(http.StatusTooManyRequests, gin.H{
			"success": false,
			"error":   "quota_exceeded",
			"message": err.Error(),
		})
		return
	}

	// Build ICE servers list
	iceServers := make([]ICEServer, 0)

	// Add STUN servers (from legacy config)
	stunServers := h.turnIntegration.GetSTUNServers()
	for _, stunURL := range stunServers {
		iceServers = append(iceServers, ICEServer{
			URLs: stunURL,
		})
	}

	// Add TURN servers with short-lived credentials
	for _, turnURI := range turnCred.URIs {
		iceServers = append(iceServers, ICEServer{
			URLs:           turnURI,
			Username:       turnCred.Username,
			Credential:     turnCred.Password,
			CredentialType: "password",
		})
	}

	// Build response
	response := &ICECredentialsResponsePion{
		Success:        true,
		TenantID:       tenantIDStr,
		PeerID:         peerID,
		ICEServers:     iceServers,
		TTL:            turnCred.TTL,
		ExpiresAt:      turnCred.ExpiresAt.Format(time.RFC3339),
		Policy:         "all", // Use both STUN and TURN
		QuotaUsed:      turnCred.QuotaUsed,
		QuotaLimit:     turnCred.QuotaLimit,
		Standard:       "RFC 5766 + draft-uberti-behave-turn-rest-00",
		Implementation: "pion/turn v3",
	}

	h.logger.Debug("Returning ICE credentials (RFC 5766)",
		zap.Int("stun_servers", len(stunServers)),
		zap.Int("turn_servers", len(turnCred.URIs)),
		zap.Int("ttl_seconds", turnCred.TTL),
		zap.Int("quota_used", turnCred.QuotaUsed))

	c.JSON(http.StatusOK, response)
}

// extractPeerID extracts peer_id from various sources
func (h *ICECredentialsHandlerPion) extractPeerID(c *gin.Context) string {
	// Try context first (set by auth middleware)
	if peerID, exists := c.Get("peer_id"); exists {
		if peerIDStr, ok := peerID.(string); ok && peerIDStr != "" {
			return peerIDStr
		}
	}

	// Try URL parameter
	if peerID := c.Param("peer_id"); peerID != "" {
		return peerID
	}

	// Try JWT claims
	if claims, exists := c.Get("claims"); exists {
		if claimsMap, ok := claims.(map[string]interface{}); ok {
			if peerID, ok := claimsMap["peer_id"].(string); ok && peerID != "" {
				return peerID
			}
			if sub, ok := claimsMap["sub"].(string); ok && sub != "" {
				return sub
			}
		}
	}

	// Try query parameter
	if peerID := c.Query("peer_id"); peerID != "" {
		return peerID
	}

	// Try header
	if peerID := c.GetHeader("X-Peer-ID"); peerID != "" {
		return peerID
	}

	// Try request body
	var req struct {
		PeerID string `json:"peer_id"`
	}
	if err := c.ShouldBindJSON(&req); err == nil && req.PeerID != "" {
		return req.PeerID
	}

	return ""
}

// parseTTL parses TTL from query parameter
func (h *ICECredentialsHandlerPion) parseTTL(c *gin.Context) time.Duration {
	ttlStr := c.Query("ttl")
	if ttlStr == "" {
		return h.defaultTTL
	}

	// Try parsing as duration string (e.g., "10m", "600s")
	if ttl, err := time.ParseDuration(ttlStr); err == nil {
		return ttl
	}

	// Default if parse fails
	return h.defaultTTL
}

// HandleHealthCheck returns health status of ICE credentials service
func (h *ICECredentialsHandlerPion) HandleHealthCheck(c *gin.Context) {
	health := h.turnCredentialService.HealthCheck()

	status := http.StatusOK
	if health["status"] == "unhealthy" {
		status = http.StatusServiceUnavailable
	}

	c.JSON(status, health)
}
