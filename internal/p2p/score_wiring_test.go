package p2p

import (
	"testing"

	"github.com/twogc/cloudbridge-relay-installer/internal/database"
	"github.com/twogc/cloudbridge-relay-installer/internal/network"
	"go.uber.org/zap"
)

func TestSCOREWiringFromEnv(t *testing.T) {
	t.Setenv("SCORE_BASE_URL", "http://score.example:8080")
	t.Setenv("SCORE_API_KEY", "secret")
	cfg := SCOREWiringFromEnv()
	if !cfg.Enabled() {
		t.Fatal("expected Enabled with base URL")
	}
	if cfg.BaseURL != "http://score.example:8080" || cfg.APIKey != "secret" {
		t.Fatalf("unexpected cfg: %+v", cfg)
	}
}

func TestSCOREWiringDisabledWithoutURL(t *testing.T) {
	t.Setenv("SCORE_BASE_URL", "")
	t.Setenv("SCORE_API_KEY", "x")
	if SCOREWiringFromEnv().Enabled() {
		t.Fatal("expected disabled without base URL")
	}
}

func TestNewRelayManagerHybridFromSCORE_RequiresURL(t *testing.T) {
	_, err := NewRelayManagerHybridFromSCORE(zap.NewNop(), SCOREWiringConfig{}, nil)
	if err == nil {
		t.Fatal("expected error when SCORE disabled")
	}
}

func TestNewRelayManagerHybridFromSCORE_WiresAllocator(t *testing.T) {
	// Prefer SQLite fallback (no Postgres env).
	t.Setenv("SCORE_DATABASE_URL", "")
	t.Setenv("DATABASE_URL", "")
	t.Setenv("SCORE_PG_HOST", "")
	t.Setenv("SCORE_SQLITE_PATH", ":memory:")

	// Constructor builds clients only; SCORE HTTP may fail during ListActive.
	rm, err := NewRelayManagerHybridFromSCORE(zap.NewNop(), SCOREWiringConfig{
		BaseURL: "http://127.0.0.1:1",
		APIKey:  "test",
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rm == nil {
		t.Fatal("expected non-nil hybrid manager")
	}
	if rm.networkAllocator == nil {
		t.Fatal("expected production SCORE allocator wired")
	}
	if rm.networkStore == nil {
		t.Fatal("expected SCORE network store wired")
	}
	// SQLite peer IPAM should attach when no Postgres.
	if rm.peerIPAMRepo == nil {
		t.Fatal("expected peer IPAM from SQLite fallback")
	}
	_ = rm.Shutdown()
}

func TestResolvePeerIPAM_PrefersInjected(t *testing.T) {
	// Empty env so backends would open if used.
	t.Setenv("SCORE_DATABASE_URL", "")
	t.Setenv("SCORE_PG_HOST", "")
	t.Setenv("SCORE_SQLITE_PATH", ":memory:")
	// Injected non-nil: function should return same pointer without opening SQLite path for nil check —
	// we only assert non-nil passthrough with a real sqlite repo.
	repo, err := database.OpenSQLitePeerIPAM(":memory:", zap.NewNop())
	if err != nil {
		t.Fatal(err)
	}
	got := resolvePeerIPAMRepository(zap.NewNop(), repo)
	if got != repo {
		t.Fatal("expected injected repository returned as-is")
	}
}

func TestNewRelayManagerHybrid_NilRepository(t *testing.T) {
	store := network.NewSCORETenantNetworkStoreFixed("http://127.0.0.1:1", "k", zap.NewNop())
	rm := NewRelayManagerHybrid(zap.NewNop(), nil, store, nil)
	if rm == nil {
		t.Fatal("nil manager")
	}
	if rm.peerIPAMRepo != nil {
		t.Fatal("expected nil peer IPAM without repository")
	}
	_ = rm.Shutdown()
}
