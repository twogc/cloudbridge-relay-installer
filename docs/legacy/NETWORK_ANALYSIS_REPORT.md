# Network Analysis Report - LoadBalancer Connectivity Issue

**Date:** September 28, 2025  
**Issue:** VK Cloud LoadBalancer cannot reach Kubernetes backend  
**Status:** Root Cause Identified  

## Network Architecture Analysis

### Current Network Configuration

```
┌─────────────────────────────────────────────────────────────┐
│                    VK Cloud Infrastructure                  │
├─────────────────────────────────────────────────────────────┤
│  LoadBalancer (109.120.180.160)                            │
│  ├── External IP: 109.120.180.160 (public)                 │
│  └── Internal IP: 10.0.0.242 (VK Cloud network)            │
│                                                             │
│  Subnet: 10.0.0.0/24 (subnet_4226)                         │
│  ├── Gateway: 10.0.0.1                                     │
│  ├── DHCP Pool: 10.0.0.2-10.0.0.254                        │
│  └── Current Server: 10.0.0.249 (relay server)             │
└─────────────────────────────────────────────────────────────┘
                              │
                              │  NETWORK ISOLATION
                              │
┌─────────────────────────────────────────────────────────────┐
│                    Local Infrastructure                     │
├─────────────────────────────────────────────────────────────┤
│  Kubernetes (Minikube)                                      │
│  ├── Cluster IP: 192.168.49.2                              │
│  ├── NodePort: 32500                                       │
│  └── Service: cloudbridge-relay-external                   │
│                                                             │
│  Subnet: 192.168.49.0/24 (Minikube network)                │
│  ├── Gateway: 192.168.49.1                                 │
│  ├── Bridge: br-f9faf2d8fa80                               │
│  └── Current Host: 10.0.0.249 (has route to 192.168.49.x) │
└─────────────────────────────────────────────────────────────┘
```

### Network Routing Analysis

**Current routing table on relay server (10.0.0.249):**
```bash
10.0.0.0/24 dev ens3          # VK Cloud network
192.168.49.0/24 dev br-f9faf2d8fa80  # Minikube network
```

**Problem:** LoadBalancer (10.0.0.242) has no route to 192.168.49.0/24

## Root Cause Confirmation

### You are CORRECT!

**Да, вы правильно поняли проблему:**

1. **LoadBalancer находится в сети:** `10.0.0.0/24`
2. **Kubernetes находится в сети:** `192.168.49.0/24`
3. **Нет маршрута между сетями:** LoadBalancer не может достичь 192.168.49.x

### Current LoadBalancer Member Status
```
Pool: c9c66e2b-8f2a-42a1-937f-a1c4084965c2 (HTTP)
└── Member: 192.168.49.2:32500 [ERROR]
    ├── Address: 192.168.49.2 (unreachable from 10.0.0.242)
    ├── Protocol Port: 32500
    └── Status: ERROR (cannot connect)
```

## Solution Options

### Option 1: Network Bridge (RECOMMENDED for Production)

**Add route to VK Cloud network to reach Minikube:**

```bash
# On VK Cloud router (if possible)
openstack router set --route destination=192.168.49.0/24,gateway=10.0.0.249 router_6328

# Or configure relay server as gateway
# LoadBalancer -> 10.0.0.249 -> 192.168.49.2:32500
```

**Steps:**
1. **Configure relay server (10.0.0.249) as gateway**
2. **Add static route on LoadBalancer** pointing to relay server
3. **Update LoadBalancer member** to use relay server as proxy

### Option 2: Move Kubernetes to VK Cloud Network

**Deploy Kubernetes on VK Cloud VM with 10.0.0.x IP:**

```bash
# Create VM in VK Cloud network
openstack server create \
  --flavor m1.large \
  --image ubuntu-20.04 \
  --network network_1740 \
  --key-name cloudbridge-key \
  cloudbridge-k8s

# VM will get IP like 10.0.0.100
# Deploy Kubernetes on this VM
# Update LoadBalancer member to 10.0.0.100:8080
```

### Option 3: Use Relay Server as Proxy (IMMEDIATE FIX)

**Configure relay server to proxy traffic:**

```bash
# Update LoadBalancer member to point to relay server
openstack loadbalancer member delete c9c66e2b-8f2a-42a1-937f-a1c4084965c2 e403e795-a21b-4d56-bd05-45795c1756ab

# Add relay server as member
openstack loadbalancer member create \
  --address 10.0.0.249 \
  --protocol-port 8080 \
  c9c66e2b-8f2a-42a1-937f-a1c4084965c2
```

**Configure Nginx on relay server:**
```nginx
# /etc/nginx/sites-available/cloudbridge-proxy
server {
    listen 8080;
    server_name _;
    
    location / {
        proxy_pass http://192.168.49.2:32500;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## Network Requirements for LoadBalancer

### Current VK Cloud Network
- **Network ID:** `79dd5769-5888-4bd9-bda7-78d6daa944d5`
- **Subnet ID:** `d3272fe8-fec0-4fbb-8e67-6c6716972f19`
- **CIDR:** `10.0.0.0/24`
- **Gateway:** `10.0.0.1`
- **Available IPs:** `10.0.0.2-10.0.0.254`

### Required Network Access
- **LoadBalancer needs access to:** `192.168.49.0/24`
- **Current gateway:** `10.0.0.249` (relay server)
- **Route needed:** `192.168.49.0/24 via 10.0.0.249`

## Immediate Action Plan

### Phase 1: Quick Fix (Today)
```bash
# 1. Remove current broken member
openstack loadbalancer member delete \
  c9c66e2b-8f2a-42a1-937f-a1c4084965c2 \
  e403e795-a21b-4d56-bd05-45795c1756ab

# 2. Add relay server as member
openstack loadbalancer member create \
  --address 10.0.0.249 \
  --protocol-port 8080 \
  c9c66e2b-8f2a-42a1-937f-a1c4084965c2

# 3. Configure Nginx proxy on relay server (if SSH access available)
```

### Phase 2: Network Configuration (Production)
```bash
# 1. Add static route to VK Cloud router
openstack router set \
  --route destination=192.168.49.0/24,gateway=10.0.0.249 \
  router_6328

# 2. Or deploy Kubernetes on VK Cloud VM
# 3. Update LoadBalancer member to new IP
```

## Testing Network Connectivity

### From LoadBalancer perspective (10.0.0.242):
```bash
# This should work (same network)
ping 10.0.0.249  # hould work

# This won't work (different network)
ping 192.168.49.2  #  No route to host
```

### From relay server (10.0.0.249):
```bash
# Both should work
ping 10.0.0.242    # oadBalancer
ping 192.168.49.2  # ubernetes
```

## Verification Commands

### Check current member status:
```bash
openstack loadbalancer member show \
  c9c66e2b-8f2a-42a1-937f-a1c4084965c2 \
  e403e795-a21b-4d56-bd05-45795c1756ab
```

### Check network routing:
```bash
# On relay server
ip route show | grep 192.168.49
traceroute 192.168.49.2
```

### Test connectivity:
```bash
# From relay server to Kubernetes
curl http://192.168.49.2:32500/health

# From LoadBalancer network (if possible)
curl http://10.0.0.249:8080/health  # After proxy setup
```

## Conclusion

**Вы абсолютно правы!** 

Проблема в том, что LoadBalancer нужен доступ к сети `192.168.49.0/24`, но у него есть доступ только к `10.0.0.0/24`.

**Решения по приоритету:**
1. **Немедленно:** Использовать relay server (10.0.0.249) как proxy
2.  **Краткосрочно:** Добавить маршрут через relay server
3. **Долгосрочно:** Развернуть Kubernetes в VK Cloud сети

**Для пилотного проекта:** Продолжайте использовать прямое подключение `192.168.49.2:32500` - оно работает отлично!
