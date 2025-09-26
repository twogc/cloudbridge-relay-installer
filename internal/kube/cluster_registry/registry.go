package cluster_registry

// Cluster represents a single Kubernetes cluster configuration stored in the registry.
// TODO: extend with CA data, client certs/keys, service-account tokens and other metadata.
// Notice: keep fields minimal for now – they will be expanded in subsequent commits.

type Cluster struct {
	ID            string // unique identifier used in URL (e.g. clusterId)
	APIServerURL  string // full https://host:port of the API server
	CAPEM         []byte // CA bundle (PEM), optional when SkipTLSVerify is true
	ClientCertPEM []byte // optional client certificate for mTLS
	ClientKeyPEM  []byte // optional client private key for mTLS
	SkipTLSVerify bool   // allow insecure-skip-tls-verify (dev only)
}

// Registry provides read-only access to cluster metadata. In the first version the
// registry is immutable and loaded from configuration at startup. Later versions may
// support CRUD operations backed by a DB or admin API.

type Registry interface {
	// Get returns the cluster configuration by its ID. If the cluster is not found,
	// (nil, false) is returned.
	Get(id string) (*Cluster, bool)
}

// StaticRegistry is an in-memory implementation loaded from configuration. It is
// concurrency-safe and read-only after creation.

// TODO: add option to watch config file/Helm values for live reload.

type StaticRegistry struct {
	clusters map[string]*Cluster
}

// NewStaticRegistry creates a new registry from the given slice.
func NewStaticRegistry(clusters []*Cluster) *StaticRegistry {
	m := make(map[string]*Cluster, len(clusters))
	for _, c := range clusters {
		if c != nil {
			m[c.ID] = c
		}
	}
	return &StaticRegistry{clusters: m}
}

// Get implements Registry.
func (r *StaticRegistry) Get(id string) (*Cluster, bool) {
	c, ok := r.clusters[id]
	return c, ok
}
