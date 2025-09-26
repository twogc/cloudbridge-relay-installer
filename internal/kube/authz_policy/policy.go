package authz_policy

import (
	"fmt"
	"strings"
	"time"
)

// Verb represents a Kubernetes verb like get, list, watch, create, delete etc.
type Verb string

const (
	VerbGet         Verb = "get"
	VerbList        Verb = "list"
	VerbWatch       Verb = "watch"
	VerbCreate      Verb = "create"
	VerbUpdate      Verb = "update"
	VerbPatch       Verb = "patch"
	VerbDelete      Verb = "delete"
	VerbExec        Verb = "exec"        // for pods/exec
	VerbAttach      Verb = "attach"      // for pods/attach
	VerbLogs        Verb = "logs"        // for pods/log
	VerbProxy       Verb = "proxy"       // for nodes/proxy, services/proxy
	VerbPortForward Verb = "portforward" // for pods/portforward
)

// Rule defines allow-list rule for a given combination of verbs/resources/namespaces.
type Rule struct {
	Name       string   `json:"name"`       // human-readable rule name
	Verbs      []Verb   `json:"verbs"`      // allowed verbs
	APIGroups  []string `json:"apiGroups"`  // allowed API groups ("" for core group)
	Resources  []string `json:"resources"`  // allowed resources
	Namespaces []string `json:"namespaces"` // allowed namespaces (empty = all)
	Deny       bool     `json:"deny"`       // if true, this is a deny rule
}

// TenantPolicy holds policies for a specific tenant/user/role
type TenantPolicy struct {
	TenantID string   `json:"tenantId"`
	UserID   string   `json:"userId,omitempty"`
	Groups   []string `json:"groups,omitempty"`
	Rules    []Rule   `json:"rules"`
}

// PolicyManager manages authorization policies for multiple tenants
type PolicyManager struct {
	policies      map[string]*TenantPolicy // tenantId -> policy
	defaultPolicy *TenantPolicy            // fallback policy
}

// NewPolicyManager creates a new policy manager
func NewPolicyManager(policies []*TenantPolicy) *PolicyManager {
	policyMap := make(map[string]*TenantPolicy)
	for _, p := range policies {
		if p != nil && p.TenantID != "" {
			policyMap[p.TenantID] = p
		}
	}

	// Create restrictive default policy
	defaultPolicy := &TenantPolicy{
		TenantID: "default",
		Rules: []Rule{
			{
				Name:       "default-read-only",
				Verbs:      []Verb{VerbGet, VerbList},
				APIGroups:  []string{""},
				Resources:  []string{"pods", "services", "configmaps"},
				Namespaces: []string{"default"},
			},
		},
	}

	return &PolicyManager{
		policies:      policyMap,
		defaultPolicy: defaultPolicy,
	}
}

// AuthRequest represents an authorization request
type AuthRequest struct {
	TenantID   string   `json:"tenantId"`
	UserID     string   `json:"userId"`
	Groups     []string `json:"groups"`
	Verb       Verb     `json:"verb"`
	APIGroup   string   `json:"apiGroup"`
	Resource   string   `json:"resource"`
	Namespace  string   `json:"namespace"`
	Name       string   `json:"name,omitempty"` // resource name
	Path       string   `json:"path,omitempty"` // full request path
	RemoteAddr string   `json:"remoteAddr,omitempty"`
}

// AuthResponse represents an authorization response
type AuthResponse struct {
	Allowed     bool   `json:"allowed"`
	Reason      string `json:"reason,omitempty"`
	MatchedRule string `json:"matchedRule,omitempty"`
}

// AuditEvent represents an authorization audit event
type AuditEvent struct {
	Timestamp time.Time     `json:"timestamp"`
	RequestID string        `json:"requestId"`
	ClusterID string        `json:"clusterId"`
	Request   *AuthRequest  `json:"request"`
	Response  *AuthResponse `json:"response"`
	Duration  time.Duration `json:"duration"`
}

// Authorizer evaluates requests against policies and generates audit events
type Authorizer struct {
	manager *PolicyManager
	auditCh chan *AuditEvent // async audit logging
}

// NewAuthorizer creates a new authorizer with policy manager
func NewAuthorizer(manager *PolicyManager) *Authorizer {
	auditCh := make(chan *AuditEvent, 1000) // buffered channel
	return &Authorizer{
		manager: manager,
		auditCh: auditCh,
	}
}

// Authorize evaluates an authorization request
func (a *Authorizer) Authorize(req *AuthRequest) *AuthResponse {
	startTime := time.Now()

	// Get policy for tenant
	policy := a.manager.GetPolicy(req.TenantID)

	// Evaluate rules
	response := a.evaluateRules(policy, req)

	// Create audit event
	auditEvent := &AuditEvent{
		Timestamp: startTime,
		Request:   req,
		Response:  response,
		Duration:  time.Since(startTime),
	}

	// Send to audit channel (non-blocking)
	select {
	case a.auditCh <- auditEvent:
	default:
		// Channel full, drop event (could log warning)
	}

	return response
}

// evaluateRules checks request against policy rules
func (a *Authorizer) evaluateRules(policy *TenantPolicy, req *AuthRequest) *AuthResponse {
	// First check deny rules
	for _, rule := range policy.Rules {
		if rule.Deny && a.ruleMatches(&rule, req) {
			return &AuthResponse{
				Allowed:     false,
				Reason:      "denied by explicit deny rule",
				MatchedRule: rule.Name,
			}
		}
	}

	// Then check allow rules
	for _, rule := range policy.Rules {
		if !rule.Deny && a.ruleMatches(&rule, req) {
			return &AuthResponse{
				Allowed:     true,
				MatchedRule: rule.Name,
			}
		}
	}

	// No matching allow rule found
	return &AuthResponse{
		Allowed: false,
		Reason:  "no matching allow rule",
	}
}

// ruleMatches checks if a rule matches the request
func (a *Authorizer) ruleMatches(rule *Rule, req *AuthRequest) bool {
	// Check verb
	if !containsVerb(rule.Verbs, req.Verb) {
		return false
	}

	// Check API group
	if !matches(req.APIGroup, rule.APIGroups) {
		return false
	}

	// Check resource
	if !matches(req.Resource, rule.Resources) {
		return false
	}

	// Check namespace
	if !matches(req.Namespace, rule.Namespaces) {
		return false
	}

	return true
}

// GetPolicy returns policy for a tenant, falls back to default
func (pm *PolicyManager) GetPolicy(tenantID string) *TenantPolicy {
	if policy, exists := pm.policies[tenantID]; exists {
		return policy
	}
	return pm.defaultPolicy
}

// UpdatePolicy updates or adds a tenant policy
func (pm *PolicyManager) UpdatePolicy(policy *TenantPolicy) {
	if policy != nil && policy.TenantID != "" {
		pm.policies[policy.TenantID] = policy
	}
}

// RemovePolicy removes a tenant policy
func (pm *PolicyManager) RemovePolicy(tenantID string) {
	delete(pm.policies, tenantID)
}

// GetAuditChannel returns the audit events channel
func (a *Authorizer) GetAuditChannel() <-chan *AuditEvent {
	return a.auditCh
}

// Helper functions
func containsVerb(list []Verb, v Verb) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

func matches(value string, allowed []string) bool {
	if len(allowed) == 0 {
		return true // empty list means wildcard/all allowed
	}
	for _, a := range allowed {
		if a == "*" || strings.EqualFold(a, value) {
			return true
		}
	}
	return false
}

// ParseVerbFromMethod converts HTTP method to Kubernetes verb
func ParseVerbFromMethod(method string) Verb {
	switch strings.ToUpper(method) {
	case "GET":
		return VerbGet
	case "POST":
		return VerbCreate
	case "PUT":
		return VerbUpdate
	case "PATCH":
		return VerbPatch
	case "DELETE":
		return VerbDelete
	default:
		return Verb(strings.ToLower(method))
	}
}

// ParseResourceFromPath extracts resource and namespace from Kubernetes API path
func ParseResourceFromPath(path string) (apiGroup, resource, namespace, name string, err error) {
	// Remove leading slash and split path
	path = strings.TrimPrefix(path, "/")
	parts := strings.Split(path, "/")

	if len(parts) < 2 {
		return "", "", "", "", fmt.Errorf("invalid API path: %s", path)
	}

	// Handle different API path patterns:
	// /api/v1/pods
	// /api/v1/namespaces/{namespace}/pods
	// /api/v1/namespaces/{namespace}/pods/{name}
	// /apis/{group}/{version}/...

	if parts[0] == "api" && len(parts) >= 3 {
		// Core API group (/api/v1/...)
		apiGroup = ""
		if parts[2] == "namespaces" && len(parts) >= 5 {
			// Namespaced resource: /api/v1/namespaces/{ns}/pods
			namespace = parts[3]
			resource = parts[4]
			if len(parts) > 5 {
				name = parts[5]
			}
		} else {
			// Cluster-scoped or direct resource: /api/v1/pods
			resource = parts[2]
			if len(parts) > 3 {
				name = parts[3]
			}
		}
	} else if parts[0] == "apis" && len(parts) >= 4 {
		// Named API group (/apis/{group}/{version}/...)
		apiGroup = parts[1]
		if parts[3] == "namespaces" && len(parts) >= 6 {
			// Namespaced resource: /apis/{group}/{version}/namespaces/{ns}/resources
			namespace = parts[4]
			resource = parts[5]
			if len(parts) > 6 {
				name = parts[6]
			}
		} else {
			// Cluster-scoped resource: /apis/{group}/{version}/resources
			resource = parts[3]
			if len(parts) > 4 {
				name = parts[4]
			}
		}
	} else {
		return "", "", "", "", fmt.Errorf("unsupported API path pattern: %s", path)
	}

	return apiGroup, resource, namespace, name, nil
}
