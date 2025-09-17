#!/usr/bin/env bash
set -euo pipefail

# WireGuard kernel integration test using network namespaces
# Requires: iproute2, wg, wgctrl (kernel module), iptables or nft, ping

if [[ $EUID -ne 0 ]]; then
  echo "Please run as root" >&2
  exit 1
fi

NS1=wgtest1
NS2=wgtest2
IF1=wg0
IF2=wg0
IP1=10.200.200.1/24
IP2=10.200.200.2/24
PORT1=51821
PORT2=51822

cleanup() {
  ip netns del $NS1 2>/dev/null || true
  ip netns del $NS2 2>/dev/null || true
}
trap cleanup EXIT

cleanup

ip netns add $NS1
ip netns add $NS2

# Create WireGuard interfaces
ip netns exec $NS1 ip link add dev $IF1 type wireguard
ip netns exec $NS2 ip link add dev $IF2 type wireguard

# Assign IPs
ip netns exec $NS1 ip address add $IP1 dev $IF1
ip netns exec $NS2 ip address add $IP2 dev $IF2

# Generate keys
PRIV1=$(wg genkey)
PUB1=$(echo "$PRIV1" | wg pubkey)
PRIV2=$(wg genkey)
PUB2=$(echo "$PRIV2" | wg pubkey)

# Configure devices
ip netns exec $NS1 wg set $IF1 listen-port $PORT1 private-key <(echo "$PRIV1")
ip netns exec $NS2 wg set $IF2 listen-port $PORT2 private-key <(echo "$PRIV2")

# Setup veth pair for connectivity between namespaces
ip link add veth1 type veth peer name veth2
ip link set veth1 netns $NS1
ip link set veth2 netns $NS2
ip netns exec $NS1 ip addr add 172.30.0.1/24 dev veth1
ip netns exec $NS2 ip addr add 172.30.0.2/24 dev veth2
ip netns exec $NS1 ip link set veth1 up
ip netns exec $NS2 ip link set veth2 up

# Bring up WireGuard interfaces
ip netns exec $NS1 ip link set $IF1 up
ip netns exec $NS2 ip link set $IF2 up

# Set endpoints using the veth IPs
ip netns exec $NS1 wg set $IF1 peer $PUB2 allowed-ips 10.200.200.2/32 endpoint 172.30.0.2:$PORT2 persistent-keepalive 5
ip netns exec $NS2 wg set $IF2 peer $PUB1 allowed-ips 10.200.200.1/32 endpoint 172.30.0.1:$PORT1 persistent-keepalive 5

# Test ping
ip netns exec $NS1 ping -c 2 -W 2 10.200.200.2
ip netns exec $NS2 ping -c 2 -W 2 10.200.200.1

echo "WG netns integration test: SUCCESS"


