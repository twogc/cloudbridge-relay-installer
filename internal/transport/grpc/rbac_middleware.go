package grpc

import (
	"context"
	"fmt"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"go.uber.org/zap"
	
	"github.com/twogc/cloudbridge-relay-installer/internal/auth"
)

// RBACMiddleware provides Role-Based Access Control for gRPC services
type RBACMiddleware struct {
	authAdapter *auth.Adapter
	logger      *zap.Logger
	permissions map[string][]string // service.method -> required roles
	rateLimits  map[string]int      // tenant_id -> requests per minute
}

// NewRBACMiddleware creates new RBAC middleware
func NewRBACMiddleware(authAdapter *auth.Adapter, logger *zap.Logger) *RBACMiddleware {
	permissions := make(map[string][]string)
	
	// Configure method permissions according to CloudBridge security model
	permissions["/relay.v1.ControlService/Hello"] = []string{"peer", "client", "admin", "user"}  // ИСПРАВЛЕНО: добавлен Hello
	permissions["/relay.v1.ControlService/Authenticate"] = []string{"peer", "client", "admin"}
	permissions["/relay.v1.ControlService/GetStatus"] = []string{"peer", "client", "admin", "monitor"}
	permissions["/relay.v1.TunnelService/CreateTunnel"] = []string{"peer", "client"}
	permissions["/relay.v1.TunnelService/DestroyTunnel"] = []string{"peer", "client", "admin"}
	permissions["/relay.v1.TunnelService/ListTunnels"] = []string{"peer", "client", "admin", "monitor"}
	permissions["/relay.v1.HeartbeatService/SendHeartbeat"] = []string{"peer", "client"}
	permissions["/relay.v1.HeartbeatService/GetHeartbeatStats"] = []string{"admin", "monitor"}
	
	// Admin-only operations
	permissions["/relay.v1.AdminService/ManageTenant"] = []string{"admin"}
	permissions["/relay.v1.AdminService/GetMetrics"] = []string{"admin", "monitor"}
	
	return &RBACMiddleware{
		authAdapter: authAdapter,
		logger:      logger,
		permissions: permissions,
		rateLimits:  make(map[string]int),
	}
}

// UnaryInterceptor provides RBAC for unary gRPC calls
func (rbac *RBACMiddleware) UnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		
		// Skip auth for health checks
		if info.FullMethod == "/grpc.health.v1.Health/Check" {
			return handler(ctx, req)
		}
		
		// Extract and validate token
		tokenInfo, err := rbac.extractAndValidateToken(ctx)
		if err != nil {
			rbac.logSecurityEvent("AUTH_FAILED", info.FullMethod, "", err.Error())
			return nil, status.Errorf(codes.Unauthenticated, "authentication failed: %v", err)
		}
		
		// Check permissions
		if !rbac.hasPermission(info.FullMethod, tokenInfo.Roles) {
			rbac.logSecurityEvent("AUTHZ_FAILED", info.FullMethod, tokenInfo.Subject, "insufficient permissions")
			return nil, status.Errorf(codes.PermissionDenied, "insufficient permissions for %s", info.FullMethod)
		}
		
		// Rate limiting per tenant
		if err := rbac.checkRateLimit(tokenInfo.TenantID); err != nil {
			rbac.logSecurityEvent("RATE_LIMITED", info.FullMethod, tokenInfo.Subject, err.Error())
			return nil, status.Errorf(codes.ResourceExhausted, "rate limit exceeded: %v", err)
		}
		
		// Add enhanced context
		ctx = rbac.enhanceContext(ctx, tokenInfo)
		
		// Execute request
		resp, err := handler(ctx, req)
		
		// Log audit event
		duration := time.Since(start)
		rbac.logAuditEvent(info.FullMethod, tokenInfo.Subject, tokenInfo.TenantID, duration, err)
		
		return resp, err
	}
}

// StreamInterceptor provides RBAC for streaming gRPC calls
func (rbac *RBACMiddleware) StreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		start := time.Now()
		
		// Extract and validate token
		tokenInfo, err := rbac.extractAndValidateToken(ss.Context())
		if err != nil {
			rbac.logSecurityEvent("STREAM_AUTH_FAILED", info.FullMethod, "", err.Error())
			return status.Errorf(codes.Unauthenticated, "authentication failed: %v", err)
		}
		
		// Check permissions
		if !rbac.hasPermission(info.FullMethod, tokenInfo.Roles) {
			rbac.logSecurityEvent("STREAM_AUTHZ_FAILED", info.FullMethod, tokenInfo.Subject, "insufficient permissions")
			return status.Errorf(codes.PermissionDenied, "insufficient permissions for %s", info.FullMethod)
		}
		
		// Rate limiting
		if err := rbac.checkRateLimit(tokenInfo.TenantID); err != nil {
			rbac.logSecurityEvent("STREAM_RATE_LIMITED", info.FullMethod, tokenInfo.Subject, err.Error())
			return status.Errorf(codes.ResourceExhausted, "rate limit exceeded: %v", err)
		}
		
		// Execute stream
		err = handler(srv, &rbacServerStream{ServerStream: ss, tokenInfo: tokenInfo})
		
		// Log audit event
		duration := time.Since(start)
		rbac.logAuditEvent(info.FullMethod+" (stream)", tokenInfo.Subject, tokenInfo.TenantID, duration, err)
		
		return err
	}
}

// extractAndValidateToken extracts JWT token from metadata and validates it
func (rbac *RBACMiddleware) extractAndValidateToken(ctx context.Context) (*auth.TokenInfo, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("metadata not found")
	}
	
	authHeaders := md.Get("authorization")
	if len(authHeaders) == 0 {
		return nil, fmt.Errorf("authorization header not found")
	}
	
	token := strings.TrimPrefix(authHeaders[0], "Bearer ")
	
	// Validate token using auth adapter
	tokenInfo, err := rbac.authAdapter.ValidateToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}
	
	return tokenInfo, nil
}

// hasPermission checks if user has required role for the method
func (rbac *RBACMiddleware) hasPermission(method string, userRoles []string) bool {
	requiredRoles, exists := rbac.permissions[method]
	if !exists {
		// If no specific permission configured, require admin role
		requiredRoles = []string{"admin"}
	}
	
	// Check if user has any of the required roles
	for _, userRole := range userRoles {
		for _, requiredRole := range requiredRoles {
			if userRole == requiredRole {
				return true
			}
		}
	}
	
	return false
}

// checkRateLimit implements per-tenant rate limiting
func (rbac *RBACMiddleware) checkRateLimit(tenantID string) error {
	// Simple in-memory rate limiting
	// In production, use Redis or similar for distributed rate limiting
	
	currentLimit, exists := rbac.rateLimits[tenantID]
	if !exists {
		rbac.rateLimits[tenantID] = 1
		return nil
	}
	
	// Basic rate limiting: 100 requests per minute per tenant
	if currentLimit > 100 {
		return fmt.Errorf("tenant %s exceeded rate limit", tenantID)
	}
	
	rbac.rateLimits[tenantID] = currentLimit + 1
	return nil
}

// enhanceContext adds additional security context
func (rbac *RBACMiddleware) enhanceContext(ctx context.Context, tokenInfo *auth.TokenInfo) context.Context {
	ctx = context.WithValue(ctx, "tenant_id", tokenInfo.TenantID)
	ctx = context.WithValue(ctx, "user_roles", tokenInfo.Roles)
	ctx = context.WithValue(ctx, "user_subject", tokenInfo.Subject)
	ctx = context.WithValue(ctx, "auth_provider", tokenInfo.Provider)
	return ctx
}

// logSecurityEvent logs security-related events for monitoring
func (rbac *RBACMiddleware) logSecurityEvent(eventType, method, subject, details string) {
	rbac.logger.Warn("Security Event",
		zap.String("event_type", eventType),
		zap.String("method", method),
		zap.String("subject", subject),
		zap.String("details", details),
		zap.Time("timestamp", time.Now()),
	)
}

// logAuditEvent logs audit trail for compliance
func (rbac *RBACMiddleware) logAuditEvent(method, subject, tenantID string, duration time.Duration, err error) {
	status := "SUCCESS"
	if err != nil {
		status = "FAILED"
	}
	
	rbac.logger.Info("Audit Event",
		zap.String("method", method),
		zap.String("subject", subject),
		zap.String("tenant_id", tenantID),
		zap.String("status", status),
		zap.Duration("duration", duration),
		zap.Time("timestamp", time.Now()),
	)
}

// rbacServerStream wraps ServerStream to provide enhanced context
type rbacServerStream struct {
	grpc.ServerStream
	tokenInfo *auth.TokenInfo
}

// Context returns enhanced context with token information
func (s *rbacServerStream) Context() context.Context {
	ctx := s.ServerStream.Context()
	ctx = context.WithValue(ctx, "tenant_id", s.tokenInfo.TenantID)
	ctx = context.WithValue(ctx, "user_roles", s.tokenInfo.Roles)
	ctx = context.WithValue(ctx, "user_subject", s.tokenInfo.Subject)
	return ctx
}

