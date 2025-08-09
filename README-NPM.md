# @twogc/cloudbridge-relay

**Enterprise-grade high-performance relay server with WireGuard integration**

[![npm version](https://img.shields.io/npm/v/@twogc/cloudbridge-relay.svg)](https://www.npmjs.com/package/@twogc/cloudbridge-relay)
[![npm downloads](https://img.shields.io/npm/dm/@twogc/cloudbridge-relay.svg)](https://www.npmjs.com/package/@twogc/cloudbridge-relay)
[![License](https://img.shields.io/npm/l/@twogc/cloudbridge-relay.svg)](LICENSE)
[![Node.js](https://img.shields.io/node/v/@twogc/cloudbridge-relay.svg)](https://nodejs.org/)

## 🚀 Features

- **🔐 WireGuard Integration** - Real WireGuard P2P mesh network
- **⚡ Enhanced QUIC** - Optimized QUIC protocol with performance improvements
- **🛡️ Post-Quantum Cryptography** - Quantum-resistant algorithms
- **📊 Metrics & Monitoring** - Prometheus metrics and health checks
- **🔒 Security** - Rate limiting, DDoS protection, tenant isolation
- **🌐 Multi-Protocol** - QUIC, HTTP/2, WebSocket support
- **⚙️ Easy Setup** - Interactive CLI wizard and configuration

## 📦 Installation

```bash
# Install package
npm install @twogc/cloudbridge-relay

# Or with yarn
yarn add @twogc/cloudbridge-relay

# Or with pnpm
pnpm add @twogc/cloudbridge-relay
```

## 🚀 Quick Start

### Basic Usage

```javascript
const { CloudBridgeRelayQuickStart } = require('@twogc/cloudbridge-relay');

// Quick start with default configuration
const relay = new CloudBridgeRelayQuickStart();

// Start server
await relay.startServer();

// Start client
await relay.startClient();

// Check status
const status = relay.getStatus();
console.log(status); // { client: true, server: true }

// Stop everything
await relay.stop();
```

### Advanced Configuration

```javascript
const { CloudBridgeRelayQuickStart } = require('@twogc/cloudbridge-relay');

const config = {
  server: {
    port: 8080,
    host: '0.0.0.0',
    timeout: 30000
  },
  wireguard: {
    enabled: true,
    interface: 'wg-mesh0',
    port: 51820,
    privateKey: 'your-private-key',
    publicKey: 'your-public-key'
  },
  quic: {
    enabled: true,
    enhanced: true,
    maxStreams: 1000
  },
  metrics: {
    enabled: true,
    port: 9090,
    path: '/metrics'
  },
  security: {
    rateLimit: 1000,
    jwtSecret: 'your-secret-key'
  }
};

const relay = new CloudBridgeRelayQuickStart(config);
await relay.startServer();
```

## 🛠️ CLI Usage

### Interactive Setup

```bash
# Run interactive setup wizard
npx @twogc/cloudbridge-relay setup
```

### Server Mode

```bash
# Start server with default settings
npx @twogc/cloudbridge-relay server

# Start server with custom configuration
npx @twogc/cloudbridge-relay server --port 9000 --host 127.0.0.1

# Start server with specific features
npx @twogc/cloudbridge-relay server --wireguard --quic --metrics
```

### Client Mode

```bash
# Connect to server
npx @twogc/cloudbridge-relay client --server ws://localhost:8080

# Connect with specific features
npx @twogc/cloudbridge-relay client --wireguard --quic
```

### Status Check

```bash
# Check service status
npx @twogc/cloudbridge-relay status
```

### Configuration

```bash
# Generate configuration file
npx @twogc/cloudbridge-relay config --output my-config.yaml
```

## 📚 API Reference

### CloudBridgeRelayQuickStart

Main class for quick setup and management.

#### Constructor

```javascript
new CloudBridgeRelayQuickStart(config?: Partial<CloudBridgeConfig>)
```

#### Methods

- `startClient(): Promise<CloudBridgeRelay>` - Start client
- `startServer(): Promise<RelayServer>` - Start server
- `stop(): Promise<void>` - Stop all components
- `getStatus(): { client: boolean, server: boolean }` - Get status

### CloudBridgeRelay

Client class for connecting to relay servers.

```javascript
const { CloudBridgeRelay } = require('@twogc/cloudbridge-relay');

const client = new CloudBridgeRelay(config);
await client.connect();
await client.disconnect();
```

### RelayServer

Server class for running relay servers.

```javascript
const { RelayServer } = require('@twogc/cloudbridge-relay');

const server = new RelayServer(config);
await server.start();
await server.stop();
```

### WireGuardManager

WireGuard management and P2P mesh network.

```javascript
const { WireGuardManager } = require('@twogc/cloudbridge-relay');

const wg = new WireGuardManager(config);
await wg.start();
await wg.addPeer(peerConfig);
await wg.removePeer(publicKey);
```

### QUICManager

Enhanced QUIC protocol management.

```javascript
const { QUICManager } = require('@twogc/cloudbridge-relay');

const quic = new QUICManager(config);
await quic.start();
await quic.createStream();
```

## 🔧 Configuration

### Server Configuration

```javascript
{
  server: {
    port: 8080,           // Server port
    host: '0.0.0.0',      // Server host
    timeout: 30000        // Connection timeout (ms)
  }
}
```

### WireGuard Configuration

```javascript
{
  wireguard: {
    enabled: true,                    // Enable WireGuard
    interface: 'wg-mesh0',           // Interface name
    port: 51820,                     // WireGuard port
    privateKey: 'auto-generated',    // Private key
    publicKey: 'auto-generated'      // Public key
  }
}
```

### QUIC Configuration

```javascript
{
  quic: {
    enabled: true,        // Enable QUIC
    enhanced: true,       // Enable enhanced features
    maxStreams: 1000      // Maximum concurrent streams
  }
}
```

### Metrics Configuration

```javascript
{
  metrics: {
    enabled: true,        // Enable metrics
    port: 9090,          // Metrics port
    path: '/metrics'      // Metrics path
  }
}
```

### Security Configuration

```javascript
{
  security: {
    rateLimit: 1000,     // Requests per minute
    jwtSecret: 'secret'   // JWT secret key
  }
}
```

## 📊 Monitoring

### Health Check

```bash
# Check server health
curl http://localhost:8080/health

# Check metrics
curl http://localhost:9090/metrics
```

### Prometheus Metrics

The server exposes Prometheus metrics at `/metrics` endpoint:

- `cloudbridge_connections_total` - Total connections
- `cloudbridge_requests_total` - Total requests
- `cloudbridge_latency_seconds` - Connection latency
- `cloudbridge_errors_total` - Total errors
- `cloudbridge_wireguard_peers` - WireGuard peers count
- `cloudbridge_quic_streams` - QUIC streams count

### Grafana Dashboard

Import the provided Grafana dashboard for visualization:

```bash
# Dashboard JSON is available in the package
cat node_modules/@twogc/cloudbridge-relay/dashboards/cloudbridge-relay.json
```

## 🔒 Security

### Rate Limiting

```javascript
const config = {
  security: {
    rateLimit: 1000  // 1000 requests per minute per IP
  }
};
```

### Authentication

```javascript
const config = {
  security: {
    jwtSecret: 'your-secret-key',
    requireAuth: true
  }
};
```

### WireGuard Security

- Automatic key generation
- Quantum-resistant cryptography
- P2P mesh network security
- Peer authentication

## 🧪 Testing

### Unit Tests

```bash
# Run tests
npm test

# Run tests with coverage
npm run test:coverage

# Run tests in watch mode
npm run test:watch
```

### Integration Tests

```bash
# Run integration tests
npm run test:integration

# Run performance tests
npm run test:performance
```

## 📦 Docker Support

```bash
# Pull Docker image
docker pull twogc/cloudbridge-relay:latest

# Run with Docker
docker run -d \
  --name cloudbridge-relay \
  --privileged \
  -p 8080:8080 \
  -p 9090:9090 \
  -p 51820:51820/udp \
  twogc/cloudbridge-relay:latest
```

## 🔧 Development

### Building from Source

```bash
# Clone repository
git clone https://github.com/twogc/cloudbridge-relay-installer.git
cd cloudbridge-relay-installer

# Install dependencies
npm install

# Build package
npm run build

# Run tests
npm test

# Lint code
npm run lint
```

### Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Run tests and linting
6. Submit a pull request

## 📚 Examples

### Basic Server

```javascript
const { RelayServer } = require('@twogc/cloudbridge-relay');

const server = new RelayServer({
  server: { port: 8080 },
  wireguard: { enabled: true },
  quic: { enabled: true },
  metrics: { enabled: true }
});

server.start().then(() => {
  console.log('Server started on port 8080');
});
```

### Basic Client

```javascript
const { CloudBridgeRelay } = require('@twogc/cloudbridge-relay');

const client = new CloudBridgeRelay({
  server: { host: 'localhost', port: 8080 },
  wireguard: { enabled: true },
  quic: { enabled: true }
});

client.connect().then(() => {
  console.log('Connected to server');
});
```

### WireGuard P2P Mesh

```javascript
const { WireGuardManager } = require('@twogc/cloudbridge-relay');

const wg = new WireGuardManager({
  interface: 'wg-mesh0',
  port: 51820
});

await wg.start();

// Add peer to mesh network
await wg.addPeer({
  publicKey: 'peer-public-key',
  allowedIPs: ['10.100.1.2/32'],
  endpoint: '192.168.1.100:51820'
});
```

## 🆘 Support

### Documentation

- [API Reference](https://github.com/twogc/cloudbridge-relay-installer/docs/api.md)
- [Configuration Guide](https://github.com/twogc/cloudbridge-relay-installer/docs/configuration.md)
- [Security Guide](https://github.com/twogc/cloudbridge-relay-installer/docs/security.md)

### Community

- [GitHub Issues](https://github.com/twogc/cloudbridge-relay-installer/issues)
- [GitHub Discussions](https://github.com/twogc/cloudbridge-relay-installer/discussions)
- [Email Support](mailto:dev@2gc.ru)

### Commercial Support

For commercial support and consulting:

- [2GC Website](https://2gc.ru)
- [Enterprise Support](mailto:enterprise@2gc.ru)

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🏆 Status

![npm version](https://img.shields.io/npm/v/@twogc/cloudbridge-relay.svg)
![npm downloads](https://img.shields.io/npm/dm/@twogc/cloudbridge-relay.svg)
![GitHub stars](https://img.shields.io/github/stars/twogc/cloudbridge-relay-installer.svg)
![GitHub forks](https://img.shields.io/github/forks/twogc/cloudbridge-relay-installer.svg)

---

**Made with ❤️ by 2GC Team**
