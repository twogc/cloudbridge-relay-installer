# CloudBridge Client API Compatibility Layer

## Overview

This package provides backward compatibility for CloudBridge Client API paths. The client expects legacy paths (`/api/v1/tenants/{tenant_id}/peers/*`), while the server uses native paths (`/api/v1/peer/*`).

## Purpose

- **Zero breaking changes**: Existing clients continue to work without modifications
- **Gradual migration**: New clients can use native paths
- **Transparent forwarding**: Legacy requests are automatically translated
- **Observability**: All translations are logged and metriced

## Path Mappings

| Client Legacy Path | Server Native Path | Description |
|--------------------|--------------------|-------------|
| `POST /api/v1/tenants/{tid}/peers/register` | `POST /api/v1/peer/register` | Peer registration |
| `GET /api/v1/tenants/{tid}/peers/discover` | `GET /api/v1/peer/{pid}/peers` | Peer discovery |
| `POST /api/v1/tenants/{tid}/peers/{pid}/heartbeat` | `POST /api/v1/peer/{pid}/heartbeat` | Heartbeat |
| `POST /api/v1/tenants/{tid}/peers/{pid}/status` | `POST /api/v1/peer/{pid}/status` | Status update |
| `GET /api/v1/tenants/{tid}/peers/{pid}` | `GET /api/v1/peer/{pid}` | Get peer info |
| `DELETE /api/v1/tenants/{tid}/peers/{pid}` | `DELETE /api/v1/peer/{pid}` | Delete peer |
| `POST /api/v1/tenants/{tid}/ice/credentials` | `POST /api/v1/ice/credentials` | ICE credentials |
| `POST /api/v1/tenants/{tid}/ipam/allocate` | `POST /api/v1/ipam/allocate` | IPAM allocation |
| `GET /api/v1/tenants/{tid}/wireguard/config` | `GET /api/v1/wireguard/config` | WireGuard config |

## Usage

### Basic Setup

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/twogc/cloudbridge-relay-installer/internal/api/compatibility"
    "go.uber.org/zap"
)

func main() {
    router := gin.Default()
    logger, _ := zap.NewProduction()

    // Create compatibility layer
    ccl := compatibility.NewClientCompatibilityLayer(logger)

    // Register compatibility routes
    ccl.RegisterRoutes(router)

    // Register native routes
    registerNativeRoutes(router)

    router.Run(":8080")
}
```

### With Middleware

```go
router := gin.Default()

// Add logging middleware
router.Use(compatibility.LogCompatibilityMiddleware(logger))

// Add metrics middleware
metrics := &compatibility.CompatibilityMetrics{
    PathTranslations: make(map[string]int64),
}
router.Use(compatibility.MetricsMiddleware(metrics))

// Add tenant ID middleware
router.Use(compatibility.TenantIDMiddleware())

// Register routes
ccl := compatibility.NewClientCompatibilityLayer(logger)
ccl.RegisterRoutes(router)
```

## Metrics

The compatibility layer tracks:
- Total requests through compatibility layer
- Success/error counts
- Per-path translation counts

```go
type CompatibilityMetrics struct {
    RequestsTotal    int64
    SuccessCount     int64
    ErrorCount       int64
    PathTranslations map[string]int64
}
```

## Observability

### Logging

All path translations are logged at DEBUG level:

```json
{
  "level": "debug",
  "msg": "API path translation",
  "legacy_path": "/api/v1/tenants/test-tenant/peers/register",
  "native_path": "/api/v1/peer/register",
  "tenant_id": "test-tenant",
  "peer_id": ""
}
```

Legacy requests are logged at INFO level:

```json
{
  "level": "info",
  "msg": "Client compatibility request",
  "method": "POST",
  "path": "/api/v1/tenants/test-tenant/peers/register",
  "remote_addr": "192.168.1.1",
  "user_agent": "CloudBridge-Client/1.4.20"
}
```

### Health Check

```go
health := ccl.HealthCheck()
// Returns:
// {
//   "name": "client_compatibility_layer",
//   "status": "healthy",
//   "version": "1.0.0",
//   "routes_registered": [...]
// }
```

## Testing

### Run Tests

```bash
# Unit tests
go test ./internal/api/compatibility/... -v

# With coverage
go test ./internal/api/compatibility/... -cover -coverprofile=coverage.out

# View coverage
go tool cover -html=coverage.out
```

### Benchmark

```bash
go test ./internal/api/compatibility/... -bench=. -benchmem
```

Expected performance:
- Peer Register: ~50-100 µs/op
- Peer Discovery: ~50-100 µs/op
- Minimal overhead: < 10 µs for path translation

## 🔧 Configuration

### Extracting Peer ID

The compatibility layer extracts `peer_id` from multiple sources (in order of priority):

1. Request context (set by auth middleware)
2. JWT claims (`peer_id` field)
3. JWT claims (`sub` field as fallback)
4. Query parameter (`?peer_id=...`)
5. HTTP header (`X-Peer-ID`)

### Custom Extraction

```go
// Override extractPeerID for custom logic
func customExtractPeerID(c *gin.Context) string {
    // Your custom logic
    return peerID
}
```

## 📝 Examples

### Client Request (Legacy Path)

```bash
curl -X POST https://relay.2gc.ru/api/v1/tenants/my-tenant/peers/register \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "public_key": "...",
    "allowed_ips": ["10.0.0.1/32"]
  }'
```

### Server Receives (Native Path)

The request is automatically translated to:

```bash
POST /api/v1/peer/register
X-Tenant-ID: my-tenant
Authorization: Bearer $JWT_TOKEN
{
  "public_key": "...",
  "allowed_ips": ["10.0.0.1/32"]
}
```

## Important Notes

1. **tenant_id Propagation**: The `tenant_id` is extracted from the legacy path and added to the request context as `c.Set("tenant_id", tenantID)`. Native handlers can access it via `c.Get("tenant_id")`.

2. **Peer Discovery Special Case**: The peer discovery endpoint requires extracting `peer_id` since the client doesn't include it in the path. The compatibility layer handles this automatically.

3. **Backward Compatibility**: This layer maintains 100% backward compatibility with existing clients. No client changes are required.

4. **Performance**: The translation overhead is minimal (<10 µs per request). Use benchmarks to verify in your environment.

5. **Migration Path**: Clients can gradually migrate to native paths. Both legacy and native paths work simultaneously.

## 🔜 Future Enhancements

- [ ] Prometheus metrics export
- [ ] Request/response transformation
- [ ] Rate limiting per tenant
- [ ] Deprecation warnings with headers
- [ ] Automatic client version detection

## 📚 References

- [CloudBridge Client Repository](https://github.com/2gc-dev/cloudbridge-client)
- [Integration Analysis](../../../docs/implementation/CLIENT_SERVER_INTEGRATION_ANALYSIS.md)
- [Sync Plan](../../../docs/implementation/CLIENT_SERVER_SYNC_PLAN.md)

## 🤝 Contributing

When adding new compatibility routes:

1. Add the route mapping in `RegisterRoutes()`
2. Update path mappings table in this README
3. Add unit tests in `client_routes_test.go`
4. Update health check routes list

---

**Version**: 1.0.0
**Last Updated**: 2025-10-13
**Status**: Production Ready
