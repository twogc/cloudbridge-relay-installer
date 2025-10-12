package compatibility

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ClientCompatibilityLayer provides backward compatibility for CloudBridge Client API paths
// Client expects: /api/v1/tenants/{tenant_id}/peers/*
// Server provides: /api/v1/peer/*
type ClientCompatibilityLayer struct {
	logger *zap.Logger
}

// NewClientCompatibilityLayer creates a new compatibility layer
func NewClientCompatibilityLayer(logger *zap.Logger) *ClientCompatibilityLayer {
	return &ClientCompatibilityLayer{
		logger: logger,
	}
}

// NativeHandlers stores references to native handlers for forwarding
type NativeHandlers struct {
	PeerRegister      gin.HandlerFunc
	PeerDiscovery     gin.HandlerFunc
	PeerHeartbeat     gin.HandlerFunc
	PeerStatus        gin.HandlerFunc
	GetPeer           gin.HandlerFunc
	DeletePeer        gin.HandlerFunc
	ICECredentials    gin.HandlerFunc
	IPAMAllocate      gin.HandlerFunc
	WireGuardConfig   gin.HandlerFunc
}

// RegisterRoutesWithHandlers registers client compatibility routes with handler references
// This method should be used when you have references to native handlers
func (ccl *ClientCompatibilityLayer) RegisterRoutesWithHandlers(router *gin.Engine, handlers *NativeHandlers) {
	ccl.logger.Info("Registering CloudBridge Client compatibility routes with handler references")

	// Legacy client routes group
	legacy := router.Group("/api/v1/tenants/:tenant_id")
	{
		// Peer registration
		if handlers.PeerRegister != nil {
			legacy.POST("/peers/register", ccl.wrapHandler(handlers.PeerRegister))
		}

		// Peer discovery
		if handlers.PeerDiscovery != nil {
			legacy.GET("/peers/discover", ccl.wrapHandlerWithPeerID(handlers.PeerDiscovery))
		}

		// Peer-specific routes
		peers := legacy.Group("/peers/:peer_id")
		{
			if handlers.PeerHeartbeat != nil {
				peers.POST("/heartbeat", ccl.wrapHandler(handlers.PeerHeartbeat))
			}
			if handlers.PeerStatus != nil {
				peers.POST("/status", ccl.wrapHandler(handlers.PeerStatus))
			}
			if handlers.GetPeer != nil {
				peers.GET("", ccl.wrapHandler(handlers.GetPeer))
			}
			if handlers.DeletePeer != nil {
				peers.DELETE("", ccl.wrapHandler(handlers.DeletePeer))
			}
		}

		// ICE credentials
		if handlers.ICECredentials != nil {
			legacy.POST("/ice/credentials", ccl.wrapHandler(handlers.ICECredentials))
		}

		// IPAM allocation
		if handlers.IPAMAllocate != nil {
			legacy.POST("/ipam/allocate", ccl.wrapHandler(handlers.IPAMAllocate))
		}

		// WireGuard config
		if handlers.WireGuardConfig != nil {
			legacy.GET("/wireguard/config", ccl.wrapHandler(handlers.WireGuardConfig))
		}
	}

	ccl.logger.Info("Client compatibility routes registered successfully",
		zap.Int("routes_count", 9))
}

// RegisterRoutes registers client compatibility routes (simplified version for testing)
// In production, use RegisterRoutesWithHandlers instead
func (ccl *ClientCompatibilityLayer) RegisterRoutes(router *gin.Engine) {
	ccl.logger.Info("Registering CloudBridge Client compatibility routes (simple mode)")

	// This method is for testing or when native handlers are registered separately
	// The routes will be registered but forwarding logic needs native handlers to be present

	legacy := router.Group("/api/v1/tenants/:tenant_id")
	{
		legacy.POST("/peers/register", ccl.compatibilityWrapper())
		legacy.GET("/peers/discover", ccl.compatibilityWrapper())

		peers := legacy.Group("/peers/:peer_id")
		{
			peers.POST("/heartbeat", ccl.compatibilityWrapper())
			peers.POST("/status", ccl.compatibilityWrapper())
			peers.GET("", ccl.compatibilityWrapper())
			peers.DELETE("", ccl.compatibilityWrapper())
		}

		legacy.POST("/ice/credentials", ccl.compatibilityWrapper())
		legacy.POST("/ipam/allocate", ccl.compatibilityWrapper())
		legacy.GET("/wireguard/config", ccl.compatibilityWrapper())
	}

	ccl.logger.Info("Client compatibility routes registered successfully",
		zap.Int("routes_count", 10))
}

// wrapHandler wraps a native handler with compatibility layer logic
func (ccl *ClientCompatibilityLayer) wrapHandler(nativeHandler gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract tenant_id from legacy path
		tenantID := c.Param("tenant_id")
		peerID := c.Param("peer_id")

		// Add to context for downstream handlers
		c.Set("tenant_id", tenantID)
		if peerID != "" {
			c.Set("peer_id", peerID)
		}

		// Log the translation
		ccl.logger.Debug("API path translation",
			zap.String("legacy_path", c.Request.URL.Path),
			zap.String("tenant_id", tenantID),
			zap.String("peer_id", peerID))

		// Call the native handler
		nativeHandler(c)
	}
}

// wrapHandlerWithPeerID wraps a handler and extracts peer_id for discovery
func (ccl *ClientCompatibilityLayer) wrapHandlerWithPeerID(nativeHandler gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.Param("tenant_id")
		c.Set("tenant_id", tenantID)

		// Extract peer_id from various sources
		peerID := ccl.extractPeerID(c)
		if peerID != "" {
			c.Set("peer_id", peerID)
			// Update params for native handler
			c.Params = append(c.Params, gin.Param{Key: "peer_id", Value: peerID})
		}

		ccl.logger.Debug("Peer discovery path translation",
			zap.String("tenant_id", tenantID),
			zap.String("peer_id", peerID))

		// Call the native handler
		nativeHandler(c)
	}
}

// compatibilityWrapper is a simple wrapper that just sets context
func (ccl *ClientCompatibilityLayer) compatibilityWrapper() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.Param("tenant_id")
		peerID := c.Param("peer_id")

		c.Set("tenant_id", tenantID)
		if peerID != "" {
			c.Set("peer_id", peerID)
		}

		ccl.logger.Debug("Compatibility wrapper",
			zap.String("path", c.Request.URL.Path),
			zap.String("tenant_id", tenantID),
			zap.String("peer_id", peerID))

		c.Next()
	}
}

// aliasToNativeHandler creates a handler that forwards to native endpoint
func (ccl *ClientCompatibilityLayer) aliasToNativeHandler(nativePath string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract tenant_id from legacy path
		tenantID := c.Param("tenant_id")
		peerID := c.Param("peer_id")

		// Add tenant_id to context for downstream handlers
		c.Set("tenant_id", tenantID)

		// Log the translation
		ccl.logger.Debug("API path translation",
			zap.String("legacy_path", c.Request.URL.Path),
			zap.String("native_path", nativePath),
			zap.String("tenant_id", tenantID),
			zap.String("peer_id", peerID))

		// Replace :peer_id placeholder with actual peer_id if present
		actualPath := strings.ReplaceAll(nativePath, ":peer_id", peerID)

		// Forward to native endpoint by creating internal redirect
		// This ensures the request is properly routed through Gin's router
		c.Request.URL.Path = actualPath
		c.Params = gin.Params{
			{Key: "peer_id", Value: peerID},
		}

		// IMPORTANT: Don't call c.Next() here
		// Let Gin re-route the request with the new path
	}
}

// peerDiscoveryHandler handles peer discovery with special logic
// Client sends: GET /api/v1/tenants/{tenant_id}/peers/discover
// Server expects: GET /api/v1/peer/{peer_id}/peers
func (ccl *ClientCompatibilityLayer) peerDiscoveryHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.Param("tenant_id")
		c.Set("tenant_id", tenantID)

		// Get peer_id from JWT claims or request context
		peerID := ccl.extractPeerID(c)
		if peerID == "" {
			ccl.logger.Warn("Peer discovery without peer_id",
				zap.String("tenant_id", tenantID))
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "peer_id not found in request context",
				"code":    http.StatusBadRequest,
			})
			return
		}

		ccl.logger.Debug("Peer discovery path translation",
			zap.String("tenant_id", tenantID),
			zap.String("peer_id", peerID))

		// Forward to native endpoint
		c.Request.URL.Path = "/api/v1/peer/" + peerID + "/peers"
		c.Next()
	}
}

// extractPeerID extracts peer_id from JWT claims or request context
func (ccl *ClientCompatibilityLayer) extractPeerID(c *gin.Context) string {
	// Try to get from context (set by auth middleware)
	if peerID, exists := c.Get("peer_id"); exists {
		if peerIDStr, ok := peerID.(string); ok {
			return peerIDStr
		}
	}

	// Try to get from JWT claims
	if claims, exists := c.Get("claims"); exists {
		if claimsMap, ok := claims.(map[string]interface{}); ok {
			if peerID, ok := claimsMap["peer_id"].(string); ok {
				return peerID
			}
			// Fallback to "sub" claim
			if sub, ok := claimsMap["sub"].(string); ok {
				return sub
			}
		}
	}

	// Try to get from query parameter
	if peerID := c.Query("peer_id"); peerID != "" {
		return peerID
	}

	// Try to get from header
	if peerID := c.GetHeader("X-Peer-ID"); peerID != "" {
		return peerID
	}

	return ""
}

// Middleware for compatibility layer
// This can be used to add compatibility logic to specific routes

// TenantIDMiddleware extracts tenant_id and adds it to context
func TenantIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract tenant_id from path if present
		if tenantID := c.Param("tenant_id"); tenantID != "" {
			c.Set("tenant_id", tenantID)
		}

		// Also try to get from JWT claims
		if claims, exists := c.Get("claims"); exists {
			if claimsMap, ok := claims.(map[string]interface{}); ok {
				if tenantID, ok := claimsMap["tenant_id"].(string); ok {
					c.Set("tenant_id", tenantID)
				}
			}
		}

		c.Next()
	}
}

// LogCompatibilityMiddleware logs all compatibility layer requests
func LogCompatibilityMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only log legacy paths
		if strings.HasPrefix(c.Request.URL.Path, "/api/v1/tenants/") {
			logger.Info("Client compatibility request",
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.String("remote_addr", c.ClientIP()),
				zap.String("user_agent", c.Request.UserAgent()))
		}
		c.Next()
	}
}

// CompatibilityMetrics tracks compatibility layer usage
type CompatibilityMetrics struct {
	RequestsTotal  int64
	SuccessCount   int64
	ErrorCount     int64
	PathTranslations map[string]int64
}

// MetricsMiddleware tracks compatibility layer metrics
func MetricsMiddleware(metrics *CompatibilityMetrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only track legacy paths
		if strings.HasPrefix(c.Request.URL.Path, "/api/v1/tenants/") {
			metrics.RequestsTotal++

			// Track path translation
			if metrics.PathTranslations == nil {
				metrics.PathTranslations = make(map[string]int64)
			}
			metrics.PathTranslations[c.Request.URL.Path]++

			c.Next()

			// Track success/error
			if c.Writer.Status() >= 200 && c.Writer.Status() < 300 {
				metrics.SuccessCount++
			} else {
				metrics.ErrorCount++
			}
		} else {
			c.Next()
		}
	}
}

// HealthCheck returns compatibility layer health status
func (ccl *ClientCompatibilityLayer) HealthCheck() map[string]interface{} {
	return map[string]interface{}{
		"name":    "client_compatibility_layer",
		"status":  "healthy",
		"version": "1.0.0",
		"routes_registered": []string{
			"/api/v1/tenants/:tenant_id/peers/register",
			"/api/v1/tenants/:tenant_id/peers/discover",
			"/api/v1/tenants/:tenant_id/peers/:peer_id/heartbeat",
			"/api/v1/tenants/:tenant_id/peers/:peer_id/status",
			"/api/v1/tenants/:tenant_id/ice/credentials",
			"/api/v1/tenants/:tenant_id/ipam/allocate",
			"/api/v1/tenants/:tenant_id/wireguard/config",
		},
	}
}
