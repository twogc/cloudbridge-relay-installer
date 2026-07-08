#!/bin/bash

# AI Service Security Setup Script
# Sets up NetworkPolicy, RBAC, and security configurations

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
NAMESPACE="cloudbridge"
SERVICE_NAME="ai-service"
MONITORING_NAMESPACE="cloudbridge-monitoring"

echo -e "${BLUE}AI Service Security Setup Script${NC}"
echo "================================================"

# Function to check if namespace exists
check_namespace() {
    if kubectl get namespace $NAMESPACE >/dev/null 2>&1; then
        echo -e "${GREEN}Namespace $NAMESPACE exists${NC}"
        return 0
    else
        echo -e "${RED}Namespace $NAMESPACE not found${NC}"
        return 1
    fi
}

# Function to create NetworkPolicy
create_network_policy() {
    echo -e "${BLUE}Creating NetworkPolicy...${NC}"
    
    NETWORK_POLICY_YAML="apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: $SERVICE_NAME-network-policy
  namespace: $NAMESPACE
spec:
  podSelector:
    matchLabels:
      app.kubernetes.io/name: $SERVICE_NAME
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: $NAMESPACE
    - namespaceSelector:
        matchLabels:
          name: $MONITORING_NAMESPACE
    ports:
    - protocol: TCP
      port: 8083
    - protocol: TCP
      port: 8084
  egress:
  - to:
    - namespaceSelector:
        matchLabels:
          name: $NAMESPACE
    - namespaceSelector:
        matchLabels:
          name: $MONITORING_NAMESPACE
    ports:
    - protocol: TCP
      port: 9090
    - protocol: TCP
      port: 6379
    - protocol: TCP
      port: 7933
  - to: []
    ports:
    - protocol: TCP
      port: 53
    - protocol: UDP
      port: 53
    - protocol: TCP
      port: 443
    - protocol: TCP
      port: 80"
    
    echo -n "Creating NetworkPolicy... "
    echo "$NETWORK_POLICY_YAML" | kubectl apply -f - >/dev/null 2>&1
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}Created${NC}"
        return 0
    else
        echo -e "${RED}Failed${NC}"
        return 1
    fi
}

# Function to create RBAC
create_rbac() {
    echo -e "${BLUE}Creating RBAC...${NC}"
    
    # ServiceAccount
    SERVICE_ACCOUNT_YAML="apiVersion: v1
kind: ServiceAccount
metadata:
  name: $SERVICE_NAME
  namespace: $NAMESPACE
  labels:
    app.kubernetes.io/name: $SERVICE_NAME
    app.kubernetes.io/component: ai"
    
    echo -n "Creating ServiceAccount... "
    echo "$SERVICE_ACCOUNT_YAML" | kubectl apply -f - >/dev/null 2>&1
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}Created${NC}"
    else
        echo -e "${RED}Failed${NC}"
        return 1
    fi
    
    # Role
    ROLE_YAML="apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  namespace: $NAMESPACE
  name: $SERVICE_NAME-role
rules:
- apiGroups: [\"\"]
  resources: [\"configmaps\", \"secrets\"]
  verbs: [\"get\", \"list\", \"watch\"]
- apiGroups: [\"apps\"]
  resources: [\"deployments\"]
  verbs: [\"get\", \"list\", \"watch\"]
- apiGroups: [\"\"]
  resources: [\"pods\"]
  verbs: [\"get\", \"list\", \"watch\"]"
    
    echo -n "Creating Role... "
    echo "$ROLE_YAML" | kubectl apply -f - >/dev/null 2>&1
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}Created${NC}"
    else
        echo -e "${RED}Failed${NC}"
        return 1
    fi
    
    # RoleBinding
    ROLE_BINDING_YAML="apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: $SERVICE_NAME-rolebinding
  namespace: $NAMESPACE
subjects:
- kind: ServiceAccount
  name: $SERVICE_NAME
  namespace: $NAMESPACE
roleRef:
  kind: Role
  name: $SERVICE_NAME-role
  apiGroup: rbac.authorization.k8s.io"
    
    echo -n "Creating RoleBinding... "
    echo "$ROLE_BINDING_YAML" | kubectl apply -f - >/dev/null 2>&1
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}Created${NC}"
        return 0
    else
        echo -e "${RED}Failed${NC}"
        return 1
    fi
}

# Function to create PodSecurityPolicy
create_pod_security_policy() {
    echo -e "${BLUE}Creating PodSecurityPolicy...${NC}"
    
    PSP_YAML="apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: $SERVICE_NAME-psp
  namespace: $NAMESPACE
spec:
  privileged: false
  allowPrivilegeEscalation: false
  requiredDropCapabilities:
    - ALL
  volumes:
    - 'configMap'
    - 'emptyDir'
    - 'projected'
    - 'secret'
    - 'downwardAPI'
    - 'persistentVolumeClaim'
  runAsUser:
    rule: 'MustRunAsNonRoot'
  seLinux:
    rule: 'RunAsAny'
  fsGroup:
    rule: 'RunAsAny'"
    
    echo -n "Creating PodSecurityPolicy... "
    echo "$PSP_YAML" | kubectl apply -f - >/dev/null 2>&1
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}Created${NC}"
        return 0
    else
        echo -e "${RED}Failed${NC}"
        return 1
    fi
}

# Function to create SecurityContextConstraints (for OpenShift)
create_scc() {
    echo -e "${BLUE}Creating SecurityContextConstraints...${NC}"
    
    SCC_YAML="apiVersion: security.openshift.io/v1
kind: SecurityContextConstraints
metadata:
  name: $SERVICE_NAME-scc
allowHostDirVolumePlugin: false
allowHostIPC: false
allowHostNetwork: false
allowHostPID: false
allowHostPorts: false
allowPrivilegeEscalation: false
allowPrivilegedContainer: false
allowedCapabilities: null
defaultAddCapabilities: null
fsGroup:
  type: MustRunAs
  ranges:
  - min: 1000
    max: 65535
readOnlyRootFilesystem: true
requiredDropCapabilities:
- ALL
runAsUser:
  type: MustRunAsNonRoot
  uid: 1000
seLinuxContext:
  type: MustRunAs
volumes:
- configMap
- downwardAPI
- emptyDir
- persistentVolumeClaim
- projected
- secret"
    
    echo -n "Creating SecurityContextConstraints... "
    echo "$SCC_YAML" | kubectl apply -f - >/dev/null 2>&1
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}Created${NC}"
        return 0
    else
        echo -e "${YELLOW}  SCC not supported (not OpenShift)${NC}"
        return 0
    fi
}

# Function to create Secret for API keys
create_secrets() {
    echo -e "${BLUE}Creating Secrets...${NC}"
    
    # Generate random API key
    API_KEY=$(openssl rand -hex 32)
    
    SECRET_YAML="apiVersion: v1
kind: Secret
metadata:
  name: $SERVICE_NAME-secrets
  namespace: $NAMESPACE
type: Opaque
data:
  api-key: $(echo -n $API_KEY | base64)
  prometheus-token: $(echo -n 'prometheus-token-123' | base64)
  redis-password: $(echo -n 'redis-password-123' | base64)"
    
    echo -n "Creating Secrets... "
    echo "$SECRET_YAML" | kubectl apply -f - >/dev/null 2>&1
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}Created${NC}"
        echo -e "${YELLOW}   API Key: $API_KEY${NC}"
        return 0
    else
        echo -e "${RED}Failed${NC}"
        return 1
    fi
}

# Function to create ConfigMap for security settings
create_security_config() {
    echo -e "${BLUE}Creating Security ConfigMap...${NC}"
    
    SECURITY_CONFIG_YAML="apiVersion: v1
kind: ConfigMap
metadata:
  name: $SERVICE_NAME-security-config
  namespace: $NAMESPACE
data:
  security.yaml: |
    security:
      enabled: true
      api_key_required: true
      rate_limiting:
        enabled: true
        requests_per_minute: 100
        burst_size: 20
      cors:
        enabled: true
        allowed_origins:
          - \"https://cloudbridge.example.com\"
          - \"https://grafana.cloudbridge.example.com\"
        allowed_methods:
          - GET
          - POST
          - PUT
          - DELETE
        allowed_headers:
          - Content-Type
          - Authorization
          - X-API-Key
      authentication:
        enabled: true
        jwt_secret: \"jwt-secret-key\"
        token_expiry: \"24h\"
      authorization:
        enabled: true
        roles:
          - admin
          - user
          - readonly
        permissions:
          admin:
            - \"*\"  # All permissions
          user:
            - \"ai:read\"
            - \"ai:write\"
            - \"ai:predict\"
          readonly:
            - \"ai:read\"
            - \"ai:status\""
    
    echo -n "Creating Security ConfigMap... "
    echo "$SECURITY_CONFIG_YAML" | kubectl apply -f - >/dev/null 2>&1
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}Created${NC}"
        return 0
    else
        echo -e "${RED}Failed${NC}"
        return 1
    fi
}

# Function to create PodSecurityPolicy
create_pod_security_standards() {
    echo -e "${BLUE}Creating Pod Security Standards...${NC}"
    
    # Add security labels to namespace
    kubectl label namespace $NAMESPACE pod-security.kubernetes.io/enforce=restricted --overwrite >/dev/null 2>&1
    kubectl label namespace $NAMESPACE pod-security.kubernetes.io/audit=restricted --overwrite >/dev/null 2>&1
    kubectl label namespace $NAMESPACE pod-security.kubernetes.io/warn=restricted --overwrite >/dev/null 2>&1
    
    echo -e "${GREEN}Pod Security Standards applied${NC}"
}

echo -e "${YELLOW}Setting up AI Service Security...${NC}"
echo

# Check prerequisites
echo -e "${BLUE}Checking prerequisites...${NC}"
check_namespace $NAMESPACE

echo

# Create security resources
create_network_policy
echo
create_rbac
echo
create_pod_security_policy
echo
create_scc
echo
create_secrets
echo
create_security_config
echo
create_pod_security_standards

echo

# Summary
echo "================================================"
echo -e "${GREEN}AI Service Security Setup Complete!${NC}"
echo

echo -e "${BLUE}Security Features Enabled:${NC}"
echo "  - NetworkPolicy: Restricted network access"
echo "  - RBAC: Role-based access control"
echo "  - PodSecurityPolicy: Container security policies"
echo "  - Secrets: API keys and tokens"
echo "  - Security ConfigMap: Rate limiting and CORS"
echo "  - Pod Security Standards: Restricted security profile"
echo

echo -e "${BLUE}Security Commands:${NC}"
echo "  - NetworkPolicy: kubectl get networkpolicy -n $NAMESPACE"
echo "  - RBAC: kubectl get role,rolebinding,serviceaccount -n $NAMESPACE"
echo "  - Secrets: kubectl get secrets -n $NAMESPACE"
echo "  - ConfigMaps: kubectl get configmaps -n $NAMESPACE"
echo

echo -e "${YELLOW}📝 Next Steps:${NC}"
echo "  1. Update AI service deployment to use the new ServiceAccount"
echo "  2. Configure API key authentication in the application"
echo "  3. Test network policies and RBAC permissions"
echo "  4. Set up monitoring for security events"
echo "  5. Regular security audits and updates"


