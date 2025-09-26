package streams

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/twogc/cloudbridge-relay-installer/internal/kube/cluster_registry"
	"github.com/twogc/cloudbridge-relay-installer/internal/kube/metrics"
)

// Handler handles WebSocket upgrade requests for exec/attach/logs/port-forward.
// It performs transparent WebSocket proxying to the upstream Kubernetes API server,
// preserving subprotocols and handling proper connection management.
type Handler struct {
	registry cluster_registry.Registry
	metrics  *metrics.KubeProxyMetrics
	upgrader websocket.Upgrader
}

// NewHandler creates a new WebSocket streams handler.
func NewHandler(registry cluster_registry.Registry, metrics *metrics.KubeProxyMetrics) *Handler {
	return &Handler{
		registry: registry,
		metrics:  metrics,
		upgrader: websocket.Upgrader{
			HandshakeTimeout: 10 * time.Second,
			CheckOrigin:      func(r *http.Request) bool { return true }, // TODO: implement proper origin checking
			Subprotocols:     []string{"v4.channel.k8s.io", "v3.channel.k8s.io", "v2.channel.k8s.io", "channel.k8s.io"},
		},
	}
}

// ServeWebSocketProxy handles WebSocket upgrade and proxying to upstream API server.
func (h *Handler) ServeWebSocketProxy(w http.ResponseWriter, r *http.Request, clusterID string) {
	// Validate cluster exists
	cluster, ok := h.registry.Get(clusterID)
	if !ok {
		h.metrics.RecordWSError(clusterID, "cluster_not_found")
		http.Error(w, "cluster not found", http.StatusNotFound)
		return
	}

	// Extract subprotocol from request headers
	requestedProtocols := websocket.Subprotocols(r)
	selectedProtocol := selectSubprotocol(requestedProtocols)

	// Build upstream URL
	upstreamURL, err := buildUpstreamWebSocketURL(cluster, r)
	if err != nil {
		h.metrics.RecordWSError(clusterID, "invalid_upstream_url")
		http.Error(w, "invalid upstream URL", http.StatusBadRequest)
		return
	}

	// Create upstream WebSocket connection
	upstreamHeaders := make(http.Header)
	copyWebSocketHeaders(r.Header, upstreamHeaders)
	if selectedProtocol != "" {
		upstreamHeaders.Set("Sec-WebSocket-Protocol", selectedProtocol)
	}

	dialer := &websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
		Subprotocols:     requestedProtocols,
		TLSClientConfig:  buildTLSConfig(cluster),
	}

	upstreamConn, resp, err := dialer.Dial(upstreamURL.String(), upstreamHeaders)
	if err != nil {
		h.metrics.RecordWSHandshake(clusterID, getStatusCode(resp), selectedProtocol)
		h.metrics.RecordWSError(clusterID, "upstream_dial_failed")
		http.Error(w, "failed to connect to upstream", http.StatusBadGateway)
		return
	}
	defer upstreamConn.Close()

	// Upgrade client connection
	clientConn, err := h.upgrader.Upgrade(w, r, http.Header{
		"Sec-WebSocket-Protocol": []string{selectedProtocol},
	})
	if err != nil {
		h.metrics.RecordWSHandshake(clusterID, "400", selectedProtocol)
		h.metrics.RecordWSError(clusterID, "client_upgrade_failed")
		return
	}
	defer clientConn.Close()

	// Record successful handshake and upgrade
	h.metrics.RecordWSHandshake(clusterID, "101", selectedProtocol)
	h.metrics.RecordWSUpgrade(clusterID, selectedProtocol)
	h.metrics.IncrementWSStreams()
	defer h.metrics.DecrementWSStreams()

	// Start bidirectional proxying
	done := make(chan struct{}, 2)

	// Client -> Upstream
	go func() {
		defer func() { done <- struct{}{} }()
		h.proxy(clientConn, upstreamConn, clusterID, "client_to_upstream")
	}()

	// Upstream -> Client
	go func() {
		defer func() { done <- struct{}{} }()
		h.proxy(upstreamConn, clientConn, clusterID, "upstream_to_client")
	}()

	// Wait for either direction to close
	<-done
}

// proxy handles bidirectional message forwarding between WebSocket connections
func (h *Handler) proxy(src, dst *websocket.Conn, clusterID, direction string) {
	for {
		messageType, data, err := src.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				dst.WriteMessage(websocket.CloseMessage, []byte{})
			} else {
				h.metrics.RecordWSError(clusterID, direction+"_read_error")
			}
			break
		}

		err = dst.WriteMessage(messageType, data)
		if err != nil {
			h.metrics.RecordWSError(clusterID, direction+"_write_error")
			break
		}
	}
}

// selectSubprotocol selects the best supported subprotocol from client request
func selectSubprotocol(requested []string) string {
	preferred := []string{"v4.channel.k8s.io", "v3.channel.k8s.io", "v2.channel.k8s.io", "channel.k8s.io"}
	for _, p := range preferred {
		for _, r := range requested {
			if r == p {
				return p
			}
		}
	}
	return ""
}

// buildUpstreamWebSocketURL converts HTTP request to upstream WebSocket URL
func buildUpstreamWebSocketURL(cluster *cluster_registry.Cluster, r *http.Request) (*url.URL, error) {
	upstream, err := url.Parse(cluster.APIServerURL)
	if err != nil {
		return nil, err
	}

	// Convert scheme to WebSocket
	scheme := "wss"
	if upstream.Scheme == "http" {
		scheme = "ws"
	}

	// Remove /kube/{clusterID} prefix from path
	path := r.URL.Path
	pathParts := strings.SplitN(strings.TrimPrefix(path, "/"), "/", 3)
	if len(pathParts) >= 3 && pathParts[0] == "kube" {
		path = "/" + pathParts[2] // Skip "kube" and clusterID
	}

	return &url.URL{
		Scheme:   scheme,
		Host:     upstream.Host,
		Path:     path,
		RawQuery: r.URL.RawQuery,
	}, nil
}

// copyWebSocketHeaders copies relevant headers for WebSocket upgrade
func copyWebSocketHeaders(src, dst http.Header) {
	headers := []string{
		"Authorization",
		"User-Agent",
		"X-Forwarded-For",
		"X-Forwarded-Proto",
		"X-Real-IP",
	}

	for _, header := range headers {
		if value := src.Get(header); value != "" {
			dst.Set(header, value)
		}
	}
}

// buildTLSConfig creates TLS config for upstream connection
func buildTLSConfig(cluster *cluster_registry.Cluster) *tls.Config {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: cluster.SkipTLSVerify,
	}

	// Load CA bundle if provided
	if len(cluster.CAPEM) > 0 {
		certPool := x509.NewCertPool()
		if certPool.AppendCertsFromPEM(cluster.CAPEM) {
			tlsConfig.RootCAs = certPool
		}
	}

	// Load client certificate and key for mTLS if provided
	if len(cluster.ClientCertPEM) > 0 && len(cluster.ClientKeyPEM) > 0 {
		cert, err := tls.X509KeyPair(cluster.ClientCertPEM, cluster.ClientKeyPEM)
		if err == nil {
			tlsConfig.Certificates = []tls.Certificate{cert}
		}
	}

	return tlsConfig
}

// getStatusCode extracts status code from HTTP response, returns "unknown" if nil
func getStatusCode(resp *http.Response) string {
	if resp == nil {
		return "unknown"
	}
	return http.StatusText(resp.StatusCode)
}

// IsWebSocketUpgrade checks if the request is a WebSocket upgrade request
func IsWebSocketUpgrade(r *http.Request) bool {
	return strings.ToLower(r.Header.Get("Connection")) == "upgrade" &&
		strings.ToLower(r.Header.Get("Upgrade")) == "websocket"
}
