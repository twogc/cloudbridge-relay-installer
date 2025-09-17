package wireguard

import (
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// WgctrlAdapter configures kernel WireGuard using wgctrl
type WgctrlAdapter struct {
	client        *wgctrl.Client
	interfaceName string
	logger        *zap.Logger
}

func NewWgctrlAdapter(interfaceName string, logger *zap.Logger) (*WgctrlAdapter, error) {
	c, err := wgctrl.New()
	if err != nil {
		return nil, err
	}
	return &WgctrlAdapter{
		client:        c,
		interfaceName: interfaceName,
		logger:        logger,
	}, nil
}

func (wa *WgctrlAdapter) Close() error {
	if wa.client != nil {
		return wa.client.Close()
	}
	return nil
}

// Start ensures interface exists, assigns IP/MTU, and applies base config
func (wa *WgctrlAdapter) Start(cfg *WireGuardConfig) error {
	// Ensure link exists
	if _, err := net.InterfaceByName(wa.interfaceName); err != nil {
		if !isValidInterfaceName(wa.interfaceName) {
			return fmt.Errorf("invalid interface name: %s", wa.interfaceName)
		}
		// ip link add dev <name> type wireguard
		cmd := exec.Command("ip", "link", "add", "dev", wa.interfaceName, "type", "wireguard") // #nosec G204
		if out, errRun := cmd.CombinedOutput(); errRun != nil {
			return fmt.Errorf("failed to create interface: %w, output: %s", errRun, string(out))
		}
	}

	// Assign address
	if cfg.Interface.Address != "" {
		cmd := exec.Command("ip", "address", "add", cfg.Interface.Address, "dev", wa.interfaceName) // #nosec G204
		_ = cmd.Run()                                                                               // ignore if already assigned
	}

	// Bring up
	cmd := exec.Command("ip", "link", "set", wa.interfaceName, "up") // #nosec G204
	if out, errRun := cmd.CombinedOutput(); errRun != nil {
		return fmt.Errorf("failed to bring up interface: %w, output: %s", errRun, string(out))
	}

	// Set MTU
	if cfg.Interface.MTU > 0 {
		cmd = exec.Command("ip", "link", "set", wa.interfaceName, "mtu", strconv.Itoa(cfg.Interface.MTU)) // #nosec G204
		if out, errRun := cmd.CombinedOutput(); errRun != nil {
			return fmt.Errorf("failed to set MTU: %w, output: %s", errRun, string(out))
		}
	}

	// Build wgtypes config
	devCfg := wgtypes.Config{}

	// Private key
	if strings.TrimSpace(cfg.Interface.PrivateKey) != "" {
		pk, err := wgtypes.ParseKey(cfg.Interface.PrivateKey)
		if err != nil {
			return fmt.Errorf("invalid private key: %w", err)
		}
		devCfg.PrivateKey = &pk
	}

	// Listen port
	if cfg.Interface.ListenPort > 0 {
		devCfg.ListenPort = &cfg.Interface.ListenPort
	}

	// Peers
	for _, p := range cfg.Peers {
		peerCfg := wgtypes.PeerConfig{}
		if strings.TrimSpace(p.PublicKey) != "" {
			pub, err := wgtypes.ParseKey(p.PublicKey)
			if err != nil {
				return fmt.Errorf("invalid public key: %w", err)
			}
			peerCfg.PublicKey = pub
		}
		if len(p.AllowedIPs) > 0 {
			var allowed []net.IPNet
			for _, cidr := range p.AllowedIPs {
				_, n, err := net.ParseCIDR(cidr)
				if err != nil {
					return fmt.Errorf("invalid allowed ip: %s", cidr)
				}
				allowed = append(allowed, *n)
			}
			peerCfg.AllowedIPs = allowed
		}
		if p.Endpoint != "" {
			ep, err := net.ResolveUDPAddr("udp", p.Endpoint)
			if err == nil {
				peerCfg.Endpoint = ep
			}
		}
		if p.PersistentKeepalive > 0 {
			v := p.PersistentKeepalive
			peerCfg.PersistentKeepaliveInterval = ptrDuration(time.Duration(v) * time.Second)
		}
		if strings.TrimSpace(p.PresharedKey) != "" {
			if psk, err := wgtypes.ParseKey(p.PresharedKey); err == nil {
				peerCfg.PresharedKey = &psk
			}
		}
		devCfg.Peers = append(devCfg.Peers, peerCfg)
	}

	if err := wa.client.ConfigureDevice(wa.interfaceName, devCfg); err != nil {
		return fmt.Errorf("wgctrl configure failed: %w", err)
	}

	wa.logger.Info("Kernel WireGuard configured via wgctrl",
		zap.String("interface", wa.interfaceName))
	return nil
}

func (wa *WgctrlAdapter) Stop() error {
	// Bring down and delete
	if !isValidInterfaceName(wa.interfaceName) {
		return fmt.Errorf("invalid interface name: %s", wa.interfaceName)
	}
	cmd := exec.Command("ip", "link", "set", wa.interfaceName, "down") // #nosec G204
	_ = cmd.Run()
	cmd = exec.Command("ip", "link", "del", wa.interfaceName) // #nosec G204
	if out, errRun := cmd.CombinedOutput(); errRun != nil {
		return fmt.Errorf("failed to delete interface: %w, output: %s", errRun, string(out))
	}
	return nil
}

func (wa *WgctrlAdapter) AddPeer(peer *WireGuardPeerConfig) error {
	pub, err := wgtypes.ParseKey(peer.PublicKey)
	if err != nil {
		return fmt.Errorf("invalid public key: %w", err)
	}
	pc := wgtypes.PeerConfig{PublicKey: pub, ReplaceAllowedIPs: true}
	if len(peer.AllowedIPs) > 0 {
		var allowed []net.IPNet
		for _, cidr := range peer.AllowedIPs {
			_, n, err := net.ParseCIDR(cidr)
			if err != nil {
				return fmt.Errorf("invalid allowed ip: %s", cidr)
			}
			allowed = append(allowed, *n)
		}
		pc.AllowedIPs = allowed
	}
	if peer.Endpoint != "" {
		ep, err := net.ResolveUDPAddr("udp", peer.Endpoint)
		if err == nil {
			pc.Endpoint = ep
		}
	}
	if peer.PersistentKeepalive > 0 {
		pc.PersistentKeepaliveInterval = ptrDuration(time.Duration(peer.PersistentKeepalive) * time.Second)
	}
	if strings.TrimSpace(peer.PresharedKey) != "" {
		if psk, err := wgtypes.ParseKey(peer.PresharedKey); err == nil {
			pc.PresharedKey = &psk
		}
	}

	cfg := wgtypes.Config{Peers: []wgtypes.PeerConfig{pc}}
	return wa.client.ConfigureDevice(wa.interfaceName, cfg)
}

func (wa *WgctrlAdapter) RemovePeer(publicKey string) error {
	pub, err := wgtypes.ParseKey(publicKey)
	if err != nil {
		return fmt.Errorf("invalid public key: %w", err)
	}
	cfg := wgtypes.Config{Peers: []wgtypes.PeerConfig{{PublicKey: pub, Remove: true}}}
	return wa.client.ConfigureDevice(wa.interfaceName, cfg)
}

func (wa *WgctrlAdapter) GetStatus() (map[string]interface{}, error) {
	dev, err := wa.client.Device(wa.interfaceName)
	if err != nil {
		return nil, err
	}
	peers := make([]map[string]interface{}, 0, len(dev.Peers))
	for _, p := range dev.Peers {
		allowed := make([]string, 0, len(p.AllowedIPs))
		for _, ipn := range p.AllowedIPs {
			allowed = append(allowed, ipn.String())
		}
		peers = append(peers, map[string]interface{}{
			"public_key": p.PublicKey.String(),
			"endpoint": func() string {
				if p.Endpoint != nil {
					return p.Endpoint.String()
				}
				return ""
			}(),
			"allowed_ips":    allowed,
			"last_handshake": p.LastHandshakeTime,
		})
	}
	return map[string]interface{}{
		"interface":   wa.interfaceName,
		"listen_port": dev.ListenPort,
		"peers":       peers,
		"status":      "running",
	}, nil
}

func ptrDuration(d time.Duration) *time.Duration { return &d }

