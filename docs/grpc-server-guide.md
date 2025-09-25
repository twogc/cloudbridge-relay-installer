# gRPC Server Implementation Guide

## Overview

This guide describes the gRPC server implementation for the CloudBridge Relay server. The gRPC server provides an alternative transport protocol to the existing JSON/HTTP API, offering better performance, type safety, and streaming capabilities.

## Architecture

The gRPC server implementation consists of:

1. **Protocol Definitions**: Located in `internal/transport/grpc/proto/relay/v1/`
   - `control.proto`: Authentication and control operations
   - `tunnel.proto`: Tunnel management and data streaming 
   - `heartbeat.proto`: Health monitoring and heartbeats

2. **Service Implementations**:
   - `ControlService`: Handles Hello, Authentication, and Status operations
   - `TunnelService`: Manages tunnel lifecycle and data streaming
   - `HeartbeatService`: Provides health monitoring with bidirectional streaming

3. **Integration Layer**: `internal/transport/grpc/integration.go`
   - Manages server lifecycle
   - Integrates with existing auth, tunnel, and heartbeat managers

## Configuration

Add the following section to your `config.json`:

```json
{
  "grpc": {
    "enabled": true,
    "port": 8444,
    "tls_enabled": true,
    "cert_file": "cert/server.crt",
    "key_file": "cert/server.key",
    "mtls_enabled": false,
    "keepalive_ping": "15s",
    "keepalive_timeout": "10s",
    "max_recv_msg_size": 4194304,
    "max_send_msg_size": 4194304
  }
}
```

### Configuration Options

- `enabled`: Enable/disable gRPC server (default: false)
- `port`: TCP port for gRPC server (default: 8444)
- `tls_enabled`: Enable TLS encryption (default: false)
- `cert_file`: Path to TLS certificate file
- `key_file`: Path to TLS private key file  
- `mtls_enabled`: Enable mutual TLS authentication (default: false)
- `keepalive_ping`: Server keepalive ping interval (default: 15s)
- `keepalive_timeout`: Server keepalive timeout (default: 10s)
- `max_recv_msg_size`: Maximum receive message size in bytes (default: 4MB)
- `max_send_msg_size`: Maximum send message size in bytes (default: 4MB)

## Features

### 1. Enhanced Security
- **TLS 1.3**: Full encryption with modern cipher suites
- **mTLS Support**: Optional mutual authentication
- **JWT Authentication**: Integrated with existing auth system

### 2. Performance
- **HTTP/2**: Multiplexed connections and header compression
- **Streaming**: Bidirectional data streams for real-time communication
- **Connection Pooling**: Efficient connection reuse
- **Keepalive**: Configurable connection health monitoring

### 3. Observability
- **Prometheus Metrics**: Built-in gRPC metrics collection
- **Structured Logging**: Request/response logging with tracing
- **Health Checks**: Dedicated health monitoring service

### 4. Reliability
- **Graceful Shutdown**: Clean server termination
- **Error Handling**: Proper gRPC status codes and error messages
- **Rate Limiting**: Per-client request rate limiting
- **Recovery**: Panic recovery middleware

## API Reference

### ControlService

#### Hello
Initial handshake and capability negotiation.
```protobuf
rpc Hello(HelloRequest) returns (HelloResponse);
```

#### Authenticate  
JWT token authentication.
```protobuf
rpc Authenticate(AuthRequest) returns (AuthResponse);
```

#### GetStatus
Connection status information.
```protobuf
rpc GetStatus(StatusRequest) returns (StatusResponse);
```

### TunnelService

#### CreateTunnel
Creates a new tunnel connection.
```protobuf
rpc CreateTunnel(CreateTunnelRequest) returns (CreateTunnelResponse);
```

#### StreamData
Bidirectional data streaming for tunnel traffic.
```protobuf
rpc StreamData(stream DataPacket) returns (stream DataPacket);
```

#### ListTunnels
Lists active tunnels.
```protobuf
rpc ListTunnels(ListTunnelsRequest) returns (ListTunnelsResponse);
```

### HeartbeatService

#### StreamHeartbeat
Bidirectional heartbeat streaming.
```protobuf
rpc StreamHeartbeat(stream HeartbeatRequest) returns (stream HeartbeatResponse);
```

#### GetHealth
Health status endpoint.
```protobuf
rpc GetHealth(HealthRequest) returns (HealthResponse);
```

## Client Integration

The gRPC server is designed to work with the existing relay client. The client can use the `--transport grpc` flag to connect via gRPC instead of HTTP.

Example client usage:
```bash
./cloudbridge-client --transport grpc --server relay.example.com:8444
```

## Testing

Run the test suite:
```bash
go test ./internal/transport/grpc -v
```

## Monitoring

The gRPC server exports the following metrics:

- `grpc_server_handled_total`: Total number of RPCs completed
- `grpc_server_handling_seconds`: Histogram of response latencies
- `grpc_server_started_total`: Total number of RPCs started

Access metrics via the existing Prometheus endpoint at `/metrics`.

## Migration Strategy

1. **Phase 1**: Deploy gRPC server alongside existing HTTP API (both enabled)
2. **Phase 2**: Test gRPC with subset of clients using feature flags
3. **Phase 3**: Migrate clients to gRPC transport
4. **Phase 4**: Deprecate HTTP API after successful migration

## Troubleshooting

### Common Issues

1. **Certificate Errors**: Ensure `cert_file` and `key_file` paths are correct
2. **Port Conflicts**: Check that port 8444 is not in use by other services
3. **Firewall**: Ensure port 8444 is open for TCP traffic
4. **TLS Version**: Client must support TLS 1.2+ for encrypted connections

### Debug Logging

Enable debug logging to see detailed gRPC operations:
```json
{
  "logging": {
    "level": "debug"
  }
}
```

### Health Check

Test server health:
```bash
grpc_health_probe -addr=relay.example.com:8444
```

## Performance Tuning

### Server-side
- Adjust `max_recv_msg_size` and `max_send_msg_size` based on typical payload sizes
- Tune `keepalive_ping` and `keepalive_timeout` for your network conditions
- Consider enabling compression for high-latency networks

### Client-side  
- Use connection pooling for multiple concurrent operations
- Implement proper retry policies with exponential backoff
- Enable TLS session resumption for faster reconnections

## Future Enhancements

- **Load Balancing**: Client-side load balancing across multiple relay servers
- **Circuit Breaker**: Automatic failure detection and recovery
- **Compression**: gRPC-level message compression
- **Reflections**: gRPC server reflection for debugging
- **Admin API**: Management operations via dedicated gRPC service
