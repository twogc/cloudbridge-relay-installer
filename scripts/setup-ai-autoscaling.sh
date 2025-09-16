#!/bin/bash

# AI Service Autoscaling Setup Script
# Sets up HPA and VPA for AI service

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

echo -e "${BLUE}🚀 AI Service Autoscaling Setup Script${NC}"
echo "================================================"

# Function to check if namespace exists
check_namespace() {
    if kubectl get namespace $NAMESPACE >/dev/null 2>&1; then
        echo -e "${GREEN}✅ Namespace $NAMESPACE exists${NC}"
        return 0
    else
        echo -e "${RED}❌ Namespace $NAMESPACE not found${NC}"
        return 1
    fi
}

# Function to check if deployment exists
check_deployment() {
    if kubectl get deployment $SERVICE_NAME -n $NAMESPACE >/dev/null 2>&1; then
        echo -e "${GREEN}✅ Deployment $SERVICE_NAME exists${NC}"
        return 0
    else
        echo -e "${RED}❌ Deployment $SERVICE_NAME not found${NC}"
        return 1
    fi
}

# Function to create HPA
create_hpa() {
    echo -e "${BLUE}📊 Creating Horizontal Pod Autoscaler (HPA)...${NC}"
    
    HPA_YAML="apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: $SERVICE_NAME-hpa
  namespace: $NAMESPACE
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: $SERVICE_NAME
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
  behavior:
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
      - type: Percent
        value: 10
        periodSeconds: 60
    scaleUp:
      stabilizationWindowSeconds: 60
      policies:
      - type: Percent
        value: 50
        periodSeconds: 60
      - type: Pods
        value: 2
        periodSeconds: 60
      selectPolicy: Max"
    
    echo -n "Creating HPA... "
    echo "$HPA_YAML" | kubectl apply -f - >/dev/null 2>&1
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ Created${NC}"
        return 0
    else
        echo -e "${RED}❌ Failed${NC}"
        return 1
    fi
}

# Function to create VPA
create_vpa() {
    echo -e "${BLUE}📊 Creating Vertical Pod Autoscaler (VPA)...${NC}"
    
    # Check if VPA CRD exists
    if ! kubectl get crd verticalpodautoscalers.autoscaling.k8s.io >/dev/null 2>&1; then
        echo -e "${YELLOW}⚠️  VPA CRD not found. Installing VPA...${NC}"
        
        # Install VPA
        kubectl apply -f https://github.com/kubernetes/autoscaler/releases/download/vertical-pod-autoscaler-0.13.0/vpa-release.yaml >/dev/null 2>&1
        
        if [ $? -eq 0 ]; then
            echo -e "${GREEN}✅ VPA installed${NC}"
        else
            echo -e "${RED}❌ Failed to install VPA${NC}"
            return 1
        fi
        
        # Wait for VPA to be ready
        echo -n "Waiting for VPA to be ready... "
        kubectl wait --for=condition=available --timeout=300s deployment/vpa-admission-controller -n kube-system >/dev/null 2>&1
        if [ $? -eq 0 ]; then
            echo -e "${GREEN}✅ Ready${NC}"
        else
            echo -e "${YELLOW}⚠️  VPA may not be fully ready${NC}"
        fi
    fi
    
    VPA_YAML="apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: $SERVICE_NAME-vpa
  namespace: $NAMESPACE
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: $SERVICE_NAME
  updatePolicy:
    updateMode: 'Auto'
  resourcePolicy:
    containerPolicies:
    - containerName: ai-service
      minAllowed:
        cpu: 100m
        memory: 128Mi
      maxAllowed:
        cpu: 2000m
        memory: 4Gi
      controlledResources: ['cpu', 'memory']"
    
    echo -n "Creating VPA... "
    echo "$VPA_YAML" | kubectl apply -f - >/dev/null 2>&1
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ Created${NC}"
        return 0
    else
        echo -e "${RED}❌ Failed${NC}"
        return 1
    fi
}

# Function to create Pod Disruption Budget
create_pdb() {
    echo -e "${BLUE}📊 Creating Pod Disruption Budget (PDB)...${NC}"
    
    PDB_YAML="apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: $SERVICE_NAME-pdb
  namespace: $NAMESPACE
spec:
  minAvailable: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: $SERVICE_NAME"
    
    echo -n "Creating PDB... "
    echo "$PDB_YAML" | kubectl apply -f - >/dev/null 2>&1
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ Created${NC}"
        return 0
    else
        echo -e "${RED}❌ Failed${NC}"
        return 1
    fi
}

# Function to create custom metrics HPA
create_custom_metrics_hpa() {
    echo -e "${BLUE}📊 Creating Custom Metrics HPA...${NC}"
    
    # Check if metrics-server is installed
    if ! kubectl get deployment metrics-server -n kube-system >/dev/null 2>&1; then
        echo -e "${YELLOW}⚠️  Metrics server not found. Installing...${NC}"
        
        kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml >/dev/null 2>&1
        
        if [ $? -eq 0 ]; then
            echo -e "${GREEN}✅ Metrics server installed${NC}"
        else
            echo -e "${RED}❌ Failed to install metrics server${NC}"
            return 1
        fi
    fi
    
    CUSTOM_HPA_YAML="apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: $SERVICE_NAME-custom-hpa
  namespace: $NAMESPACE
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: $SERVICE_NAME
  minReplicas: 2
  maxReplicas: 15
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 60
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 70
  - type: Pods
    pods:
      metric:
        name: http_requests_per_second
      target:
        type: AverageValue
        averageValue: '100'"
    
    echo -n "Creating Custom Metrics HPA... "
    echo "$CUSTOM_HPA_YAML" | kubectl apply -f - >/dev/null 2>&1
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ Created${NC}"
        return 0
    else
        echo -e "${RED}❌ Failed${NC}"
        return 1
    fi
}

echo -e "${YELLOW}📋 Setting up AI Service Autoscaling...${NC}"
echo

# Check prerequisites
echo -e "${BLUE}🔍 Checking prerequisites...${NC}"
check_namespace $NAMESPACE
check_deployment $SERVICE_NAME

echo

# Create autoscaling resources
create_hpa
echo
create_vpa
echo
create_pdb
echo
create_custom_metrics_hpa

echo

# Summary
echo "================================================"
echo -e "${GREEN}🎉 AI Service Autoscaling Setup Complete!${NC}"
echo

echo -e "${BLUE}📊 Autoscaling Configuration:${NC}"
echo "  - HPA: 2-10 replicas based on CPU (70%) and Memory (80%)"
echo "  - VPA: Auto mode with CPU 100m-2000m, Memory 128Mi-4Gi"
echo "  - PDB: Minimum 1 pod available during disruptions"
echo "  - Custom HPA: 2-15 replicas with custom metrics"
echo

echo -e "${BLUE}🔍 Monitoring Commands:${NC}"
echo "  - HPA Status: kubectl get hpa -n $NAMESPACE"
echo "  - VPA Status: kubectl get vpa -n $NAMESPACE"
echo "  - PDB Status: kubectl get pdb -n $NAMESPACE"
echo "  - Pod Status: kubectl get pods -n $NAMESPACE -l app.kubernetes.io/name=$SERVICE_NAME"
echo

echo -e "${YELLOW}📝 Next Steps:${NC}"
echo "  1. Monitor autoscaling behavior with: kubectl get hpa -n $NAMESPACE -w"
echo "  2. Check VPA recommendations: kubectl describe vpa $SERVICE_NAME-vpa -n $NAMESPACE"
echo "  3. Test load to verify autoscaling works"
echo "  4. Adjust thresholds based on actual usage patterns"


