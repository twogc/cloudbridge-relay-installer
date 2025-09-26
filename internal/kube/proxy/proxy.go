package proxy

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/twogc/cloudbridge-relay-installer/internal/kube/cluster_registry"
)

// UpstreamConfig holds parameters needed to build an http.Transport for a cluster.
type UpstreamConfig struct {
	TLSConfig *tls.Config
	Timeout   time.Duration
}

// Builder constructs http.ReverseProxy instances for specific clusters.
type Builder struct {
	registry cluster_registry.Registry
}

// NewBuilder returns a new proxy Builder.
func NewBuilder(reg cluster_registry.Registry) *Builder {
	return &Builder{registry: reg}
}

// Build returns a configured reverse proxy for the given cluster ID.
func (b *Builder) Build(clusterID string) (*httputil.ReverseProxy, error) {
	cluster, ok := b.registry.Get(clusterID)
	if !ok {
		return nil, ErrClusterNotFound
	}
	target, err := url.Parse(cluster.APIServerURL)
	if err != nil {
		return nil, err
	}

	// Configure TLS
	tlsConfig := &tls.Config{
		InsecureSkipVerify: cluster.SkipTLSVerify,
	}

	// Load CA bundle if provided
	if len(cluster.CAPEM) > 0 {
		certPool := x509.NewCertPool()
		if !certPool.AppendCertsFromPEM(cluster.CAPEM) {
			return nil, fmt.Errorf("failed to parse CA certificate for cluster %s", clusterID)
		}
		tlsConfig.RootCAs = certPool
	}

	// Load client certificate and key for mTLS if provided
	if len(cluster.ClientCertPEM) > 0 && len(cluster.ClientKeyPEM) > 0 {
		cert, err := tls.X509KeyPair(cluster.ClientCertPEM, cluster.ClientKeyPEM)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate for cluster %s: %v", clusterID, err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	transport := &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		TLSClientConfig:     tlsConfig,
		ForceAttemptHTTP2:   true,
		MaxIdleConns:        100,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	proxy := httputil.ReverseProxy{
		Director: func(req *http.Request) {
			// Rewrite URL to upstream. The incoming path already contains /kube/{clusterID}/…
			// Strip the prefix and append to upstream path.
			prefix := "/kube/" + clusterID
			req.URL.Path = singleJoiningSlash(target.Path, strings.TrimPrefix(req.URL.Path, prefix))
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
		},
		Transport: transport,
	}
	return &proxy, nil
}

// singleJoiningSlash is copied from net/http/httputil package (private func).
func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

// TODO: streaming upgrades (WebSocket/SPDY) handled in separate package.

// ErrClusterNotFound is returned when the cluster ID is unknown.
var ErrClusterNotFound = fmt.Errorf("cluster not found")
