# CloudBridge Relay Server - Примеры конфигурации

## 1. Основная конфигурация (config.json)

### Минимальная конфигурация
```json
{
  "server": {
    "port": 9090,
    "max_tunnels": 100
  },
  "django": {
    "enabled": true,
    "api_endpoint": "http://localhost:8000/api/validate-relay-token/"
  },
  "wireguard": {
    "enabled": true,
    "port": 51820
  }
}
```

### Полная конфигурация
```json
{
  "server": {
    "port": 9090,
    "max_tunnels": 1000,
    "timeout": 30,
    "read_timeout": 60,
    "write_timeout": 60
  },
  "django": {
    "enabled": true,
    "api_endpoint": "http://localhost:8000/api/validate-relay-token/",
    "timeout": 5,
    "retry_count": 3,
    "retry_delay": 1
  },
  "keycloak": {
    "enabled": false,
    "server_url": "https://keycloak.example.com",
    "realm": "cloudbridge",
    "client_id": "relay-server",
    "public_key_url": "/auth/realms/cloudbridge/protocol/openid-connect/certs",
    "timeout": 10
  },
  "security": {
    "min_tls_version": "1.2",
    "cipher_suites": [
      "TLS_AES_256_GCM_SHA384",
      "TLS_CHACHA20_POLY1305_SHA256",
      "TLS_AES_128_GCM_SHA256"
    ],
    "mtls_enabled": true,
    "cert_file": "/app/certs/server.crt",
    "key_file": "/app/certs/server.key",
    "ca_file": "/app/certs/ca.crt"
  },
  "metrics": {
    "enabled": true,
    "port": 8084,
    "path": "/metrics",
    "interval": 30
  },
  "api": {
    "enabled": true,
    "port": 9090,
    "version": "v1",
    "cors_enabled": true,
    "cors_origins": ["*"]
  },
  "protocol": {
    "preferred": "quic",
    "fallback": "http2",
    "auto_switch": true
  },
  "quic": {
    "enabled": true,
    "port": 9090,
    "max_streams": 1000,
    "idle_timeout": 30,
    "keep_alive": 15
  },
  "zero_trust": {
    "enabled": false,
    "strict_mode": false,
    "default_deny": true
  },
  "enhanced_quic": {
    "enabled": true,
    "port": 9090,
    "ddos_protection": true,
    "geographic_routing": true,
    "keycloak_integration": false
  },
  "wireguard": {
    "enabled": true,
    "port": 51820,
    "interface": "wg-mesh0",
    "peer_discovery": true,
    "mesh_enabled": true,
    "keep_alive": 25,
    "persistent_keepalive": 25
  },
  "tenants": {
    "default": {
      "bandwidth_limit": 10485760,
      "connection_limit": 10,
      "rate_limit": 100,
      "burst": 20
    },
    "premium": {
      "bandwidth_limit": 104857600,
      "connection_limit": 100,
      "rate_limit": 1000,
      "burst": 200
    },
    "relay-test-tenant": {
      "bandwidth_limit": 104857600,
      "connection_limit": 20,
      "rate_limit": 500,
      "burst": 100,
      "jwt": {
        "enabled": true,
        "required_claims": ["client_id", "tenant_id", "exp"]
      }
    }
  },
  "logging": {
    "level": "info",
    "format": "json",
    "output": "stdout",
    "file": {
      "enabled": false,
      "path": "/app/logs/relay.log",
      "max_size": 100,
      "max_backups": 3,
      "max_age": 28
    }
  },
  "health_check": {
    "enabled": true,
    "interval": 30,
    "timeout": 10,
    "url": "http://localhost:9090/health"
  }
}
```

## 2. Enhanced QUIC конфигурация (enhanced_quic.yaml)

```yaml
# Enhanced QUIC Configuration
enhanced_quic:
  enabled: true
  port: 9090
  ddos_protection: true
  geographic_routing: true
  keycloak_integration: false

# Tenant configurations
tenants:
  default:
    bandwidth_limit: 10485760    # 10 MB/s
    connection_limit: 10
    rate_limit: 100
    burst: 20
    
  premium:
    bandwidth_limit: 104857600   # 100 MB/s
    connection_limit: 100
    rate_limit: 1000
    burst: 200
    
  standard:
    bandwidth_limit: 52428800    # 50 MB/s
    connection_limit: 50
    rate_limit: 500
    burst: 100
    
  basic:
    bandwidth_limit: 20971520    # 20 MB/s
    connection_limit: 20
    rate_limit: 200
    burst: 50
    
  relay-test-tenant:
    bandwidth_limit: 104857600   # 100 MB/s
    connection_limit: 20
    rate_limit: 500
    burst: 100
    features:
      - "ddos_protection"
      - "geographic_routing"
      - "priority_queuing"
```

## 3. Docker Compose конфигурация

### Базовая конфигурация
```yaml
version: '3.8'

services:
  cloudbridge-relay:
    build: .
    container_name: cloudbridge-relay
    restart: unless-stopped
    network_mode: host
    volumes:
      - ./configs:/app/configs:ro
      - ./logs:/app/logs
      - /etc/wireguard:/etc/wireguard:ro
      - /var/run/wireguard:/var/run/wireguard
    environment:
      - CONFIG_PATH=/app/configs/config.json
      - LOG_LEVEL=info
      - WIREGUARD_ENABLED=true
      - TZ=UTC
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:9090/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"

  prometheus:
    image: prom/prometheus:latest
    container_name: cloudbridge-prometheus
    restart: unless-stopped
    ports:
      - "9091:9090"
    volumes:
      - ./prometheus/prometheus.yml:/etc/prometheus/prometheus.yml:ro
      - prometheus-data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/etc/prometheus/console_libraries'
      - '--web.console.templates=/etc/prometheus/consoles'
      - '--storage.tsdb.retention.time=200h'
      - '--web.enable-lifecycle'
    networks:
      - cloudbridge-network

  grafana:
    image: grafana/grafana:latest
    container_name: cloudbridge-grafana
    restart: unless-stopped
    ports:
      - "3000:3000"
    volumes:
      - grafana-data:/var/lib/grafana
      - ./grafana/dashboards:/etc/grafana/provisioning/dashboards:ro
      - ./grafana/datasources:/etc/grafana/provisioning/datasources:ro
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
      - GF_USERS_ALLOW_SIGN_UP=false
    networks:
      - cloudbridge-network

volumes:
  prometheus-data:
  grafana-data:

networks:
  cloudbridge-network:
    driver: bridge
```

### Конфигурация с внешней сетью
```yaml
version: '3.8'

services:
  cloudbridge-relay:
    build: .
    container_name: cloudbridge-relay
    restart: unless-stopped
    ports:
      - "9090:9090"
      - "9090:9090/udp"
      - "51820:51820/udp"
      - "8084:8084"
    volumes:
      - ./configs:/app/configs:ro
      - ./logs:/app/logs
      - ./certs:/app/certs:ro
    environment:
      - CONFIG_PATH=/app/configs/config.json
      - LOG_LEVEL=info
      - WIREGUARD_ENABLED=true
      - TZ=UTC
    networks:
      - cloudbridge-network
    depends_on:
      - prometheus
      - grafana

  prometheus:
    image: prom/prometheus:latest
    container_name: cloudbridge-prometheus
    restart: unless-stopped
    ports:
      - "9091:9090"
    volumes:
      - ./prometheus/prometheus.yml:/etc/prometheus/prometheus.yml:ro
      - prometheus-data:/prometheus
    networks:
      - cloudbridge-network

  grafana:
    image: grafana/grafana:latest
    container_name: cloudbridge-grafana
    restart: unless-stopped
    ports:
      - "3000:3000"
    volumes:
      - grafana-data:/var/lib/grafana
      - ./grafana/dashboards:/etc/grafana/provisioning/dashboards:ro
      - ./grafana/datasources:/etc/grafana/provisioning/datasources:ro
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
      - GF_USERS_ALLOW_SIGN_UP=false
    networks:
      - cloudbridge-network

volumes:
  prometheus-data:
  grafana-data:

networks:
  cloudbridge-network:
    driver: bridge
    ipam:
      config:
        - subnet: 172.20.0.0/16
```

## 4. Kubernetes конфигурация

### Deployment
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cloudbridge-relay
  namespace: cloudbridge
  labels:
    app: cloudbridge-relay
spec:
  replicas: 3
  selector:
    matchLabels:
      app: cloudbridge-relay
  template:
    metadata:
      labels:
        app: cloudbridge-relay
    spec:
      containers:
      - name: cloudbridge-relay
        image: cloudbridge-relay:latest
        imagePullPolicy: Always
        ports:
        - name: tcp-api
          containerPort: 9090
          protocol: TCP
        - name: udp-wireguard
          containerPort: 51820
          protocol: UDP
        - name: metrics
          containerPort: 8084
          protocol: TCP
        env:
        - name: CONFIG_PATH
          value: "/app/configs/config.json"
        - name: LOG_LEVEL
          value: "info"
        - name: WIREGUARD_ENABLED
          value: "true"
        - name: TZ
          value: "UTC"
        volumeMounts:
        - name: config
          mountPath: /app/configs
          readOnly: true
        - name: logs
          mountPath: /app/logs
        - name: certs
          mountPath: /app/certs
          readOnly: true
        resources:
          requests:
            memory: "512Mi"
            cpu: "250m"
          limits:
            memory: "2Gi"
            cpu: "2000m"
        livenessProbe:
          httpGet:
            path: /health
            port: 9090
          initialDelaySeconds: 40
          periodSeconds: 30
          timeoutSeconds: 10
          failureThreshold: 3
        readinessProbe:
          httpGet:
            path: /health
            port: 9090
          initialDelaySeconds: 10
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 3
      volumes:
      - name: config
        configMap:
          name: cloudbridge-config
      - name: logs
        emptyDir: {}
      - name: certs
        secret:
          secretName: cloudbridge-certs
      nodeSelector:
        kubernetes.io/os: linux
      tolerations:
      - key: "cloudbridge-relay"
        operator: "Equal"
        value: "true"
        effect: "NoSchedule"
```

### Service
```yaml
apiVersion: v1
kind: Service
metadata:
  name: cloudbridge-relay-service
  namespace: cloudbridge
  labels:
    app: cloudbridge-relay
spec:
  selector:
    app: cloudbridge-relay
  ports:
  - name: tcp-api
    port: 9090
    targetPort: 9090
    protocol: TCP
  - name: udp-wireguard
    port: 51820
    targetPort: 51820
    protocol: UDP
  - name: metrics
    port: 8084
    targetPort: 8084
    protocol: TCP
  type: LoadBalancer
  loadBalancerIP: 10.0.0.100
```

### ConfigMap
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: cloudbridge-config
  namespace: cloudbridge
data:
  config.json: |
    {
      "server": {
        "port": 9090,
        "max_tunnels": 1000,
        "timeout": 30
      },
      "django": {
        "enabled": true,
        "api_endpoint": "http://django-api:8000/api/validate-relay-token/",
        "timeout": 5,
        "retry_count": 3
      },
      "wireguard": {
        "enabled": true,
        "port": 51820,
        "interface": "wg-mesh0",
        "peer_discovery": true,
        "mesh_enabled": true
      },
      "tenants": {
        "default": {
          "bandwidth_limit": 10485760,
          "connection_limit": 10,
          "rate_limit": 100,
          "burst": 20
        }
      }
    }
  enhanced_quic.yaml: |
    enhanced_quic:
      enabled: true
      port: 9090
      ddos_protection: true
      geographic_routing: true
    
    tenants:
      default:
        bandwidth_limit: 10485760
        connection_limit: 10
        rate_limit: 100
        burst: 20
```

### Secret
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: cloudbridge-certs
  namespace: cloudbridge
type: Opaque
data:
  server.crt: LS0tLS1CRUdJTi... # base64 encoded
  server.key: LS0tLS1CRUdJTi... # base64 encoded
  ca.crt: LS0tLS1CRUdJTi...     # base64 encoded
```

## 5. Prometheus конфигурация

### prometheus.yml
```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

rule_files:
  - "relay_rules.yml"

alerting:
  alertmanagers:
    - static_configs:
        - targets:
          - alertmanager:9093

scrape_configs:
  - job_name: 'cloudbridge-relay'
    static_configs:
      - targets: ['cloudbridge-relay:8084']
    scrape_interval: 30s
    metrics_path: /metrics
    
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']

  - job_name: 'grafana'
    static_configs:
      - targets: ['grafana:3000']
```

### relay_rules.yml
```yaml
groups:
- name: cloudbridge-relay
  rules:
  - alert: HighErrorRate
    expr: rate(cloudbridge_errors_total[5m]) > 0.1
    for: 2m
    labels:
      severity: warning
    annotations:
      summary: "High error rate detected"
      description: "Error rate is {{ $value }} errors per second"

  - alert: HighLatency
    expr: histogram_quantile(0.95, rate(cloudbridge_request_duration_seconds_bucket[5m])) > 1
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "High latency detected"
      description: "95th percentile latency is {{ $value }} seconds"

  - alert: HighMemoryUsage
    expr: cloudbridge_memory_usage_bytes / 1024 / 1024 / 1024 > 1.5
    for: 5m
    labels:
      severity: critical
    annotations:
      summary: "High memory usage"
      description: "Memory usage is {{ $value }} GB"

  - alert: TooManyConnections
    expr: cloudbridge_active_connections > 800
    for: 2m
    labels:
      severity: warning
    annotations:
      summary: "Too many active connections"
      description: "Active connections: {{ $value }}"
```

## 6. Grafana конфигурация

### datasources.yml
```yaml
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
    editable: true
```

### dashboards.yml
```yaml
apiVersion: 1

providers:
  - name: 'default'
    orgId: 1
    folder: ''
    type: file
    disableDeletion: false
    updateIntervalSeconds: 10
    allowUiUpdates: true
    options:
      path: /etc/grafana/provisioning/dashboards
```

## 7. Nginx конфигурация (Load Balancer)

```nginx
upstream cloudbridge_relay {
    least_conn;
    server 10.0.0.10:9090 max_fails=3 fail_timeout=30s;
    server 10.0.0.11:9090 max_fails=3 fail_timeout=30s;
    server 10.0.0.12:9090 max_fails=3 fail_timeout=30s;
}

server {
    listen 80;
    server_name relay.example.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name relay.example.com;

    ssl_certificate /etc/ssl/certs/relay.example.com.crt;
    ssl_certificate_key /etc/ssl/private/relay.example.com.key;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512:ECDHE-RSA-AES256-GCM-SHA384:DHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;

    location / {
        proxy_pass http://cloudbridge_relay;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        proxy_connect_timeout 30s;
        proxy_send_timeout 30s;
        proxy_read_timeout 30s;
        
        proxy_buffering off;
        proxy_request_buffering off;
    }

    location /health {
        proxy_pass http://cloudbridge_relay/health;
        access_log off;
    }
}
```

## 8. Systemd сервис

### cloudbridge-relay.service
```ini
[Unit]
Description=CloudBridge Relay Server
Documentation=https://github.com/2gc/cloudbridge-relay
After=network.target
Wants=network.target

[Service]
Type=simple
User=cloudbridge
Group=cloudbridge
WorkingDirectory=/opt/cloudbridge-relay
ExecStart=/opt/cloudbridge-relay/bin/cloudbridge-relay --config /opt/cloudbridge-relay/configs/config.json --log-level info
ExecReload=/bin/kill -HUP $MAINPID
KillMode=mixed
KillSignal=SIGINT
TimeoutStopSec=30
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal
SyslogIdentifier=cloudbridge-relay

# Security settings
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/opt/cloudbridge-relay/logs
ReadWritePaths=/var/run/wireguard
ReadOnlyPaths=/opt/cloudbridge-relay/configs
ReadOnlyPaths=/etc/wireguard

# Resource limits
LimitNOFILE=65536
LimitNPROC=32768

[Install]
WantedBy=multi-user.target
```

## 9. Environment переменные

### .env файл
```bash
# Основные настройки
CONFIG_PATH=/app/configs/config.json
LOG_LEVEL=info
WIREGUARD_ENABLED=true
TZ=UTC

# Сетевые настройки
SERVER_PORT=9090
WIREGUARD_PORT=51820
METRICS_PORT=8084

# Аутентификация
DJANGO_API_ENDPOINT=http://localhost:8000/api/validate-relay-token/
KEYCLOAK_ENABLED=false
KEYCLOAK_SERVER_URL=https://keycloak.example.com

# Безопасность
TLS_MIN_VERSION=1.2
MTLS_ENABLED=true
CERT_FILE=/app/certs/server.crt
KEY_FILE=/app/certs/server.key

# Производительность
MAX_TUNNELS=1000
MAX_CONNECTIONS=10000
READ_TIMEOUT=60
WRITE_TIMEOUT=60

# Мониторинг
METRICS_ENABLED=true
HEALTH_CHECK_INTERVAL=30s
PROMETHEUS_ENABLED=true

# Логирование
LOG_FORMAT=json
LOG_OUTPUT=stdout
LOG_FILE_ENABLED=false
```

---

## 5. Cadence Workflow Engine конфигурация

### Базовая конфигурация Cadence
```json
{
  "cadence": {
    "enabled": true,
    "host_port": "localhost:7933",
    "domain": "cloudbridge",
    "task_list": "relay-task-list"
  }
}
```

### Расширенная конфигурация Cadence
```json
{
  "cadence": {
    "enabled": true,
    "host_port": "localhost:7933",
    "domain": "cloudbridge",
    "task_list": "relay-task-list",
    "client_options": {
      "timeout": "30s",
      "retry_policy": {
        "initial_interval": "1s",
        "maximum_interval": "10s",
        "backoff_coefficient": 2.0,
        "maximum_attempts": 3
      }
    },
    "worker_options": {
      "max_concurrent_activities": 100,
      "max_concurrent_workflows": 50,
      "identity": "cloudbridge-worker"
    }
  }
}
```

### Docker Compose с Cadence
```yaml
version: '3.8'
services:
  cloudbridge-relay:
    build: .
    ports:
      - "9090:9090"  # TCP Server
      - "9091:9091"  # HTTP API
      - "8084:8084"  # Metrics
    environment:
      - CADENCE_ENABLED=true
      - CADENCE_HOST_PORT=cadence:7933
      - CADENCE_DOMAIN=cloudbridge
    depends_on:
      - cadence
      - cassandra
    volumes:
      - ./configs:/app/configs
    restart: unless-stopped

  cadence:
    image: ubercadence/server:latest
    ports:
      - "7933:7933"
      - "7934:7934"
      - "7935:7935"
      - "7936:7936"
      - "7939:7939"
      - "8000:8000"
      - "8001:8001"
      - "8002:8002"
      - "8003:8003"
    environment:
      - CASSANDRA_SEEDS=cassandra
      - DB=cassandra
      - DB_PORT=9042
      - KEYSPACE=cadence
      - VISIBILITY_KEYSPACE=cadence_visibility
    depends_on:
      - cassandra
    restart: unless-stopped

  cassandra:
    image: cassandra:3.11
    ports:
      - "9042:9042"
    environment:
      - CASSANDRA_CLUSTER_NAME=cadence
      - CASSANDRA_DC=dc1
      - CASSANDRA_RACK=rack1
    volumes:
      - cassandra_data:/var/lib/cassandra
    restart: unless-stopped

  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9092:9090"  # Изменен порт для избежания конфликта
    volumes:
      - ./prometheus:/etc/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
    restart: unless-stopped

volumes:
  cassandra_data:
```

### Kubernetes ConfigMap для Cadence
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: cloudbridge-cadence-config
data:
  config.json: |
    {
      "cadence": {
        "enabled": true,
        "host_port": "cadence-service:7933",
        "domain": "cloudbridge",
        "task_list": "relay-task-list",
        "client_options": {
          "timeout": "30s",
          "retry_policy": {
            "initial_interval": "1s",
            "maximum_interval": "10s",
            "backoff_coefficient": 2.0,
            "maximum_attempts": 3
          }
        }
      }
    }
```

### Переменные окружения для Cadence
```bash
# Cadence настройки
CADENCE_ENABLED=true
CADENCE_HOST_PORT=localhost:7933
CADENCE_DOMAIN=cloudbridge
CADENCE_TASK_LIST=relay-task-list

# Cadence клиент
CADENCE_CLIENT_TIMEOUT=30s
CADENCE_CLIENT_RETRY_ATTEMPTS=3
CADENCE_CLIENT_RETRY_INTERVAL=1s

# Cadence worker
CADENCE_WORKER_MAX_ACTIVITIES=100
CADENCE_WORKER_MAX_WORKFLOWS=50
CADENCE_WORKER_IDENTITY=cloudbridge-worker

# Cadence домен
CADENCE_DOMAIN_RETENTION=7d
CADENCE_DOMAIN_GLOBAL=false
```

### Helm values для Cadence
```yaml
# helm/cloudbridge-relay/values.yaml
cadence:
  enabled: true
  hostPort: "cadence-service:7933"
  domain: "cloudbridge"
  taskList: "relay-task-list"
  
  client:
    timeout: "30s"
    retryAttempts: 3
    retryInterval: "1s"
    
  worker:
    maxActivities: 100
    maxWorkflows: 50
    identity: "cloudbridge-worker"
    
  domain:
    retention: "7d"
    global: false

# Порты
ports:
  tcp: 9090
  api: 9091
  metrics: 8084
  prometheus: 9092  # Изменен для избежания конфликта
```

### Тестовая конфигурация
```json
{
  "cadence": {
    "enabled": true,
    "host_port": "localhost:7933",
    "domain": "cloudbridge-test",
    "task_list": "relay-test-task-list",
    "client_options": {
      "timeout": "10s",
      "retry_policy": {
        "initial_interval": "500ms",
        "maximum_interval": "5s",
        "backoff_coefficient": 1.5,
        "maximum_attempts": 2
      }
    }
  }
}
```

---

*Примеры конфигурации для CloudBridge Relay Server v1.0.0*
*Обновлено: 2025-09-06*
