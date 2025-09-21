#!/bin/bash

# Fix type mismatches in test files
echo "Fixing type mismatches..."

# Fix MockDataPipeline -> *pipeline.DataPipeline
find /home/ubuntu/cloudbridge-relay-installer/internal/api -name "*_test.go" -exec sed -i 's/\*MockDataPipeline/\*pipeline.DataPipeline/g' {} \;

# Fix MockHub -> *Hub
find /home/ubuntu/cloudbridge-relay-installer/internal/api -name "*_test.go" -exec sed -i 's/\*MockHub/\*Hub/g' {} \;

# Fix MockP2PConnectionManager -> *p2p.P2PConnectionManager
find /home/ubuntu/cloudbridge-relay-installer/internal/api -name "*_test.go" -exec sed -i 's/\*MockP2PConnectionManager/\*p2p.P2PConnectionManager/g' {} \;

# Fix MockHeartbeatManager -> *p2p.HeartbeatManager
find /home/ubuntu/cloudbridge-relay-installer/internal/api -name "*_test.go" -exec sed -i 's/\*MockHeartbeatManager/\*p2p.HeartbeatManager/g' {} \;

# Fix MockRelayManager -> *p2p.RelayManager
find /home/ubuntu/cloudbridge-relay-installer/internal/api -name "*_test.go" -exec sed -i 's/\*MockRelayManager/\*p2p.RelayManager/g' {} \;

# Fix MockKeycloakValidator -> *keycloak.KeycloakValidator
find /home/ubuntu/cloudbridge-relay-installer/internal/api -name "*_test.go" -exec sed -i 's/\*MockKeycloakValidator/\*keycloak.KeycloakValidator/g' {} \;

# Fix MockAIMiddleware -> *AIMiddleware
find /home/ubuntu/cloudbridge-relay-installer/internal/api -name "*_test.go" -exec sed -i 's/\*MockAIMiddleware/\*AIMiddleware/g' {} \;

# Fix MockAIClient -> *ai.Client
find /home/ubuntu/cloudbridge-relay-installer/internal/api -name "*_test.go" -exec sed -i 's/\*MockAIClient/\*ai.Client/g' {} \;

# Fix MockLogger -> *logger.Logger
find /home/ubuntu/cloudbridge-relay-installer/internal/api -name "*_test.go" -exec sed -i 's/\*MockLogger/\*logger.Logger/g' {} \;

# Fix MockMetrics -> *metrics.Metrics
find /home/ubuntu/cloudbridge-relay-installer/internal/api -name "*_test.go" -exec sed -i 's/\*MockMetrics/\*metrics.Metrics/g' {} \;

# Fix MockTunnelManager -> *tunnel.Manager
find /home/ubuntu/cloudbridge-relay-installer/internal/api -name "*_test.go" -exec sed -i 's/\*MockTunnelManager/\*tunnel.Manager/g' {} \;

echo "Type fixes completed!"
