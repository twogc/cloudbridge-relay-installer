#!/bin/bash

# Fix import issues
sed -i '/"time" imported and not used/d' /home/ubuntu/cloudbridge-relay-installer/internal/api/relay_handlers_test.go
sed -i '/"strings" imported and not used/d' /home/ubuntu/cloudbridge-relay-installer/internal/api/websocket_server_simple_test.go
sed -i '/"github.com\/gorilla\/websocket" imported and not used/d' /home/ubuntu/cloudbridge-relay-installer/internal/api/websocket_server_simple_test.go
sed -i '/"github.com\/gin-gonic\/gin" imported and not used/d' /home/ubuntu/cloudbridge-relay-installer/internal/api/protocol_handlers_test.go
sed -i '/"bytes" imported and not used/d' /home/ubuntu/cloudbridge-relay-installer/internal/api/heartbeat_handlers_test.go

# Fix type issues by using interface{} instead of specific types
sed -i 's/\*pipeline\.DataPipeline/\*MockDataPipeline/g' /home/ubuntu/cloudbridge-relay-installer/internal/api/data_pipeline_endpoints_test.go
sed -i 's/\*Hub/\*MockHub/g' /home/ubuntu/cloudbridge-relay-installer/internal/api/relay_handlers_test.go
sed -i 's/\*p2p\.P2PConnectionManager/\*MockP2PConnectionManager/g' /home/ubuntu/cloudbridge-relay-installer/internal/api/p2p_handlers_test.go
sed -i 's/\*p2p\.HeartbeatManager/\*MockHeartbeatManager/g' /home/ubuntu/cloudbridge-relay-installer/internal/api/heartbeat_handlers_test.go
sed -i 's/\*p2p\.RelayManager/\*MockRelayManager/g' /home/ubuntu/cloudbridge-relay-installer/internal/api/heartbeat_handlers_test.go
sed -i 's/\*keycloak\.KeycloakValidator/\*MockKeycloakValidator/g' /home/ubuntu/cloudbridge-relay-installer/internal/api/protocol_handlers_test.go
sed -i 's/\*AIMiddleware/\*MockAIMiddleware/g' /home/ubuntu/cloudbridge-relay-installer/internal/api/ai_metrics_exporter_test.go
sed -i 's/\*ai\.Client/\*MockAIClient/g' /home/ubuntu/cloudbridge-relay-installer/internal/api/ai_middleware_test.go

echo "Fixed basic type issues"
