# CloudBridge P2P API Reference

## Overview

The CloudBridge P2P API provides endpoints for peer discovery, real-time communication, and fallback routing in a peer-to-peer mesh network. This API enables secure communication between peers with JWT-based authentication and tenant isolation.

## Base URL

```
https://relay.2gc.ru/api/v1
```

## Authentication

All API endpoints require JWT authentication via the `Authorization` header:

```
Authorization: Bearer <jwt_token>
```

### JWT Token Structure

The JWT token must contain the following claims:

```json
{
  "sub": "peer-id",
  "tenant_id": "tenant-123",
  "org_id": "org-456",
  "device_id": "device-789",
  "connection_type": "p2p-mesh",
  "permissions": ["read", "write"],
  "wireguard_config": {
    "public_key": "peer-public-key",
    "endpoint": "192.168.1.1:51820",
    "listen_port": 51820,
    "allowed_ips": ["10.0.1.0/24"]
  },
  "mesh_config": {
    "auto_discovery": true,
    "routing": "hybrid",
    "encryption": "aes256"
  },
  "peer_whitelist": {
    "allowed_peers": ["peer-1", "peer-2"],
    "auto_approve": false,
    "max_peers": 10
  },
  "network_config": {
    "subnet": "10.0.1.0/24",
    "dns": ["8.8.8.8", "8.8.4.4"],
    "mtu": 1500
  }
}
```

## Peer Management

### Register Peer

Register a new peer in the P2P network.

**Endpoint:** `POST /peers/register`

**Request Body:**
```json
{
  "peer_id": "peer-123",
  "tenant_id": "tenant-456",
  "org_id": "org-789",
  "device_id": "device-abc",
  "connection_type": "p2p-mesh",
  "wireguard_config": {
    "public_key": "peer-public-key",
    "endpoint": "192.168.1.1:51820",
    "listen_port": 51820,
    "allowed_ips": ["10.0.1.0/24"]
  },
  "mesh_config": {
    "auto_discovery": true,
    "routing": "hybrid",
    "encryption": "aes256"
  },
  "peer_whitelist": {
    "allowed_peers": ["peer-1", "peer-2"],
    "auto_approve": false,
    "max_peers": 10
  },
  "network_config": {
    "subnet": "10.0.1.0/24",
    "dns": ["8.8.8.8", "8.8.4.4"],
    "mtu": 1500
  }
}
```

**Response:**
```json
{
  "success": true,
  "peer_id": "peer-123",
  "registered_at": "2023-01-01T12:00:00Z",
  "discovery_endpoint": "wss://relay.2gc.ru/ws/peers"
}
```

**Status Codes:**
- `200 OK` - Peer registered successfully
- `400 Bad Request` - Invalid request data
- `401 Unauthorized` - Invalid or missing JWT token
- `403 Forbidden` - Tenant mismatch
- `409 Conflict` - Peer already exists

### Discover Peers

Discover peers in the same tenant.

**Endpoint:** `GET /peers/discover?tenant_id={tenant_id}`

**Query Parameters:**
- `tenant_id` (required) - Tenant ID to discover peers for

**Response:**
```json
{
  "success": true,
  "tenant_id": "tenant-456",
  "peers": [
    {
      "id": "peer-uuid",
      "peer_id": "peer-123",
      "tenant_id": "tenant-456",
      "org_id": "org-789",
      "device_id": "device-abc",
      "connection_type": "p2p-mesh",
      "wireguard_config": {
        "public_key": "peer-public-key",
        "endpoint": "192.168.1.1:51820",
        "listen_port": 51820,
        "allowed_ips": ["10.0.1.0/24"]
      },
      "is_online": true,
      "last_seen": "2023-01-01T12:00:00Z",
      "registered_at": "2023-01-01T11:00:00Z",
      "active_connections": 2,
      "mesh_status": "connected",
      "latency_ms": 50
    }
  ]
}
```

**Status Codes:**
- `200 OK` - Peers discovered successfully
- `400 Bad Request` - Missing tenant_id parameter
- `401 Unauthorized` - Invalid or missing JWT token
- `403 Forbidden` - Tenant mismatch

### Update Peer Status

Update the status of a peer.

**Endpoint:** `PUT /peers/{peer_id}/status`

**Path Parameters:**
- `peer_id` (required) - ID of the peer to update

**Request Body:**
```json
{
  "is_online": true,
  "latency_ms": 50,
  "active_connections": 2,
  "mesh_status": "connected"
}
```

**Response:**
```json
{
  "success": true,
  "updated_at": "2023-01-01T12:00:00Z"
}
```

**Status Codes:**
- `200 OK` - Status updated successfully
- `400 Bad Request` - Invalid request data
- `401 Unauthorized` - Invalid or missing JWT token
- `403 Forbidden` - Tenant mismatch
- `404 Not Found` - Peer not found

### Get Peer Info

Get detailed information about a specific peer.

**Endpoint:** `GET /peers/{peer_id}`

**Path Parameters:**
- `peer_id` (required) - ID of the peer

**Response:**
```json
{
  "success": true,
  "peer": {
    "id": "peer-uuid",
    "peer_id": "peer-123",
    "tenant_id": "tenant-456",
    "org_id": "org-789",
    "device_id": "device-abc",
    "connection_type": "p2p-mesh",
    "wireguard_config": {
      "public_key": "peer-public-key",
      "endpoint": "192.168.1.1:51820",
      "listen_port": 51820,
      "allowed_ips": ["10.0.1.0/24"]
    },
    "mesh_config": {
      "auto_discovery": true,
      "routing": "hybrid",
      "encryption": "aes256"
    },
    "peer_whitelist": {
      "allowed_peers": ["peer-1", "peer-2"],
      "auto_approve": false,
      "max_peers": 10
    },
    "network_config": {
      "subnet": "10.0.1.0/24",
      "dns": ["8.8.8.8", "8.8.4.4"],
      "mtu": 1500
    },
    "is_online": true,
    "last_seen": "2023-01-01T12:00:00Z",
    "registered_at": "2023-01-01T11:00:00Z",
    "active_connections": 2,
    "mesh_status": "connected",
    "latency_ms": 50
  }
}
```

**Status Codes:**
- `200 OK` - Peer information retrieved successfully
- `401 Unauthorized` - Invalid or missing JWT token
- `403 Forbidden` - Tenant mismatch
- `404 Not Found` - Peer not found

## Relay Routing

### Route Through Relay

Route a message through the relay server when direct P2P connection is not available.

**Endpoint:** `POST /relay/route`

**Request Body:**
```json
{
  "from_peer": "peer-123",
  "to_peer": "peer-456",
  "data": "base64-encoded-data",
  "message_type": "chat-message"
}
```

**Response:**
```json
{
  "success": true,
  "routed_at": "2023-01-01T12:00:00Z",
  "latency_ms": 25
}
```

**Status Codes:**
- `200 OK` - Message routed successfully
- `400 Bad Request` - Invalid request data
- `401 Unauthorized` - Invalid or missing JWT token
- `403 Forbidden` - Peer not in whitelist or tenant mismatch
- `404 Not Found` - Source or destination peer not found
- `500 Internal Server Error` - Routing failed

### Get Relay Stats

Get statistics about relay usage for a tenant.

**Endpoint:** `GET /relay/stats?tenant_id={tenant_id}`

**Query Parameters:**
- `tenant_id` (required) - Tenant ID to get stats for

**Response:**
```json
{
  "success": true,
  "tenant_id": "tenant-456",
  "stats": {
    "total_peers": 10,
    "online_peers": 8,
    "direct_connections": 15,
    "relay_connections": 3,
    "total_traffic_bytes": 1048576,
    "average_latency_ms": 45
  }
}
```

**Status Codes:**
- `200 OK` - Stats retrieved successfully
- `400 Bad Request` - Missing tenant_id parameter
- `401 Unauthorized` - Invalid or missing JWT token
- `403 Forbidden` - Tenant mismatch

## Health and Monitoring

### Health Check

Check the health status of the P2P API server.

**Endpoint:** `GET /health`

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2023-01-01T12:00:00Z",
  "services": {
    "database": "healthy",
    "redis": "healthy",
    "websocket": "healthy"
  },
  "metrics": {
    "total_peers": 10,
    "online_peers": 8,
    "active_connections": 15,
    "websocket_connections": 5,
    "uptime_seconds": 3600,
    "memory_usage_mb": 128,
    "goroutines": 25
  },
  "version": "v1.0.0",
  "uptime": "1h0m0s"
}
```

### System Metrics

Get detailed system metrics.

**Endpoint:** `GET /health/metrics`

**Response:**
```json
{
  "success": true,
  "metrics": {
    "memory": {
      "alloc_mb": 64,
      "total_alloc_mb": 1024,
      "sys_mb": 256,
      "num_gc": 10,
      "gc_cpu_fraction": 0.01
    },
    "runtime": {
      "goroutines": 25,
      "cpus": 4,
      "go_version": "go1.21.0"
    },
    "peers": {
      "total": 10,
      "online": 8
    },
    "connections": {
      "total": 15,
      "active": 12
    },
    "websocket": {
      "connections": 5
    },
    "uptime": 3600
  },
  "timestamp": "2023-01-01T12:00:00Z"
}
```

### Peer Metrics

Get metrics for peers.

**Endpoint:** `GET /health/peers?tenant_id={tenant_id}`

**Query Parameters:**
- `tenant_id` (optional) - Filter by tenant ID

**Response:**
```json
{
  "success": true,
  "tenant_id": "tenant-456",
  "metrics": {
    "total_peers": 10,
    "online_peers": 8,
    "offline_peers": 2,
    "avg_latency_ms": 45,
    "connection_types": {
      "p2p-mesh": 8,
      "relay": 2
    },
    "mesh_statuses": {
      "connected": 6,
      "connecting": 2,
      "disconnected": 2
    }
  },
  "timestamp": "2023-01-01T12:00:00Z"
}
```

### Connection Metrics

Get metrics for peer connections.

**Endpoint:** `GET /health/connections?tenant_id={tenant_id}`

**Query Parameters:**
- `tenant_id` (optional) - Filter by tenant ID

**Response:**
```json
{
  "success": true,
  "tenant_id": "tenant-456",
  "metrics": {
    "total_connections": 15,
    "active_connections": 12,
    "inactive_connections": 3,
    "total_bytes": 1048576,
    "avg_latency_ms": 45,
    "connection_types": {
      "direct": 10,
      "relay": 5
    }
  },
  "timestamp": "2023-01-01T12:00:00Z"
}
```

## WebSocket Real-time Updates

### WebSocket Connection

Connect to the WebSocket endpoint for real-time updates.

**Endpoint:** `GET /ws/peers`

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**WebSocket Messages:**

#### Peer Joined
```json
{
  "type": "peer_joined",
  "data": {
    "peer_id": "peer-123",
    "wireguard_config": {
      "public_key": "peer-public-key",
      "endpoint": "192.168.1.1:51820",
      "listen_port": 51820,
      "allowed_ips": ["10.0.1.0/24"]
    },
    "timestamp": "2023-01-01T12:00:00Z"
  },
  "timestamp": "2023-01-01T12:00:00Z"
}
```

#### Peer Left
```json
{
  "type": "peer_left",
  "data": {
    "peer_id": "peer-123",
    "timestamp": "2023-01-01T12:00:00Z"
  },
  "timestamp": "2023-01-01T12:00:00Z"
}
```

#### Peer Status Update
```json
{
  "type": "peer_status_update",
  "data": {
    "peer_id": "peer-123",
    "is_online": true,
    "latency_ms": 50,
    "mesh_status": "connected",
    "timestamp": "2023-01-01T12:00:00Z"
  },
  "timestamp": "2023-01-01T12:00:00Z"
}
```

#### Relay Message
```json
{
  "type": "relay_message",
  "data": {
    "type": "relay_message",
    "message_type": "chat-message",
    "data": "base64-encoded-data",
    "timestamp": "2023-01-01T12:00:00Z"
  },
  "timestamp": "2023-01-01T12:00:00Z"
}
```

## Error Responses

All error responses follow this format:

```json
{
  "error": "Error type",
  "code": 400,
  "message": "Detailed error message"
}
```

### Common Error Codes

- `400 Bad Request` - Invalid request data or missing required fields
- `401 Unauthorized` - Invalid or missing JWT token
- `403 Forbidden` - Insufficient permissions or tenant mismatch
- `404 Not Found` - Resource not found
- `409 Conflict` - Resource already exists
- `500 Internal Server Error` - Server error

## Rate Limiting

The API implements rate limiting to prevent abuse. Rate limits are applied per IP address and JWT token:

- **General API calls**: 100 requests per minute
- **WebSocket connections**: 10 connections per minute
- **Peer registration**: 5 registrations per minute

Rate limit headers are included in responses:

```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1640995200
```

## Examples

### Complete Peer Registration Flow

1. **Register Peer:**
```bash
curl -X POST https://relay.2gc.ru/api/v1/peers/register \
  -H "Authorization: Bearer <jwt_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "peer_id": "peer-123",
    "tenant_id": "tenant-456",
    "org_id": "org-789",
    "device_id": "device-abc",
    "connection_type": "p2p-mesh",
    "wireguard_config": {
      "public_key": "peer-public-key",
      "endpoint": "192.168.1.1:51820",
      "listen_port": 51820,
      "allowed_ips": ["10.0.1.0/24"]
    }
  }'
```

2. **Discover Peers:**
```bash
curl -X GET "https://relay.2gc.ru/api/v1/peers/discover?tenant_id=tenant-456" \
  -H "Authorization: Bearer <jwt_token>"
```

3. **Connect to WebSocket:**
```javascript
const ws = new WebSocket('wss://relay.2gc.ru/ws/peers', [], {
  headers: {
    'Authorization': 'Bearer <jwt_token>'
  }
});

ws.onmessage = function(event) {
  const message = JSON.parse(event.data);
  console.log('Received:', message);
};
```

4. **Update Peer Status:**
```bash
curl -X PUT https://relay.2gc.ru/api/v1/peers/peer-123/status \
  -H "Authorization: Bearer <jwt_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "is_online": true,
    "latency_ms": 50,
    "active_connections": 2,
    "mesh_status": "connected"
  }'
```

5. **Route Message Through Relay:**
```bash
curl -X POST https://relay.2gc.ru/api/v1/relay/route \
  -H "Authorization: Bearer <jwt_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "from_peer": "peer-123",
    "to_peer": "peer-456",
    "data": "SGVsbG8gV29ybGQ=",
    "message_type": "chat-message"
  }'
```

## Security Considerations

1. **JWT Token Security**: Always use HTTPS and secure JWT tokens with proper expiration
2. **IP Validation**: WireGuard endpoints are validated against allowed IP ranges
3. **Peer Whitelist**: Peers can only communicate with whitelisted peers
4. **Tenant Isolation**: Peers can only discover and communicate with peers in the same tenant
5. **Rate Limiting**: API calls are rate-limited to prevent abuse
6. **Input Validation**: All input data is validated and sanitized

## Monitoring and Observability

The P2P API provides comprehensive monitoring through:

1. **Health Endpoints**: Real-time health status and metrics
2. **Prometheus Metrics**: Detailed metrics for monitoring systems
3. **Structured Logging**: JSON-formatted logs with correlation IDs
4. **WebSocket Events**: Real-time notifications for peer events
5. **Database Metrics**: Connection pool and query performance metrics

## Support

For technical support and questions:

- **Documentation**: [CloudBridge Documentation](https://docs.2gc.ru)
- **API Status**: [Status Page](https://status.2gc.ru)
- **Support**: [support@2gc.ru](mailto:support@2gc.ru)
