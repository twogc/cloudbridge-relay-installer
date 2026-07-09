package database

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestOpenSQLitePeerIPAM_Allocates(t *testing.T) {
	repo, err := OpenSQLitePeerIPAM(":memory:", zap.NewNop())
	require.NoError(t, err)
	require.NotNil(t, repo)
	require.NotNil(t, repo.GetDB())

	peerRepo := NewPeerIPAMRepository(repo.GetDB(), zap.NewNop())
	ip, err := peerRepo.AllocateIPForPeer("tenant-a", "peer-1", "10.100.77.0/24")
	require.NoError(t, err)
	require.NotEmpty(t, ip)

	ip2, err := peerRepo.AllocateIPForPeer("tenant-a", "peer-1", "10.100.77.0/24")
	require.NoError(t, err)
	require.Equal(t, ip, ip2, "same peer reuses allocation")
}

func TestNormalizeSQLiteDSN(t *testing.T) {
	if NormalizeSQLiteDSN("") == "" {
		t.Fatal("empty should normalize")
	}
	if NormalizeSQLiteDSN(":memory:") == ":memory:" {
		t.Fatal("memory should become shared cache DSN")
	}
	if NormalizeSQLiteDSN("/tmp/x.db") != "/tmp/x.db" {
		t.Fatal("file path should pass through")
	}
}
