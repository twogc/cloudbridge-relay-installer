#!/usr/bin/env python3
import jwt
import json
import time

# JWT секрет (должен совпадать с сервером)
JWT_SECRET = "85Sk/NfLq3gqzCXzmBKbJCpL+f5BssXz3G8dVi3sPiE="

# Создаем JWT токен с поддерживаемым протоколом
payload = {
    "protocol_type": "p2p-mesh",
    "scope": "p2p-mesh-claims",
    "org_id": "tenant-216420165",
    "tenant_id": "tenant-216420165",
    "server_id": "server-1758219195703",
    "connection_type": "wireguard",  # Используем поддерживаемый протокол
    "max_peers": "10",
    "permissions": ["mesh_join", "mesh_manage"],
    "network_config": {
        "subnet": "10.0.0.0/24",
        "gateway": "10.0.0.1",
        "dns": ["8.8.8.8", "1.1.1.1"],
        "mtu": 1420,
        "firewall_rules": ["allow_ssh", "allow_http"],
        "enable_ipv6": False
    },
    "wireguard_config": {
        "interface": "wg0",
        "listen_port": 51820,
        "private_key": "test-private-key",
        "public_key": "test-public-key-12345"
    },
    "mesh_config": {
        "network_id": "mesh-network-001",
        "subnet": "10.0.0.0/16",
        "registry_url": "https://edge.2gc.ru:8080",
        "heartbeat_interval": "30s",
        "max_peers": 10,
        "routing_strategy": "performance_optimal",
        "enable_auto_discovery": True,
        "trust_level": "basic"
    },
    "peer_whitelist": ["peer-001", "peer-002", "peer-003"],
    "relay_endpoints": {
        "http_api": "edge.2gc.ru:8080",
        "p2p_api": "edge.2gc.ru:8082",
        "wireguard_endpoint": "edge.2gc.ru:51820"
    },
    "iat": int(time.time()),
    "iss": "https://auth.2gc.ru/realms/cloudbridge",
    "aud": "account",
    "sub": "server-client-server-1758219195703",
    "jti": f"jwt_{int(time.time())}_wireguard_token"
}

# Создаем токен
token = jwt.encode(payload, JWT_SECRET, algorithm="HS256", headers={"kid": "fallback-key"})

print("JWT Token (WireGuard protocol):")
print(token)
print("\nDecoded payload:")
print(json.dumps(payload, indent=2))

