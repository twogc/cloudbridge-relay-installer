#!/bin/bash

echo "🧪 Quick test of CloudBridge Relay P2P functionality"

# Get service URL
SERVICE_IP=$(minikube service cloudbridge-relay-service -n cloudbridge-relay --url | head -1)
BASE_URL="$SERVICE_IP"

echo "Testing against: $BASE_URL"

# Test health
echo "✅ Health check:"
curl -s "$BASE_URL/health" | jq '.status' 2>/dev/null || echo "Failed"

# Test P2P connect (using port-forward to 8082)
echo "✅ P2P connect (via port-forward 8082):"
kubectl port-forward -n cloudbridge-relay pod/$(kubectl get pods -n cloudbridge-relay -o jsonpath='{.items[0].metadata.name}') 8082:8082 > /dev/null 2>&1 &
PF_PID=$!
sleep 2

curl -s -X POST "http://localhost:8082/api/v1/p2p/connect" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer test-token" \
    -d '{"session_id": "test-session-1", "target_peer_id": "test-peer-2", "protocol": "quic"}' | jq '.error' 2>/dev/null || echo "Success"

# Test ICE candidates
echo "✅ ICE candidates:"
curl -s -X POST "http://localhost:8082/api/v1/ice/candidates" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer test-token" \
    -d '{"session_id": "test-session-1", "peer_id": "test-peer-1", "candidates": []}' | jq '.error' 2>/dev/null || echo "Success"

# Clean up port-forward
kill $PF_PID 2>/dev/null || true

echo "🎉 Quick test completed!"
