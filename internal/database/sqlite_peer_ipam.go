package database

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3" // SQLite driver for optional peer IPAM
	"go.uber.org/zap"
)

// peerIPAMSchemaSQLite is the peer IP allocation schema (matches isolated tests).
const peerIPAMSchemaSQLite = `
CREATE TABLE IF NOT EXISTS peer_ip_allocations (
	id            TEXT PRIMARY KEY,
	tenant_id     TEXT NOT NULL,
	peer_id       TEXT NOT NULL,
	allocated_ip  TEXT NOT NULL,
	allocated_at  DATETIME NOT NULL,
	released_at   DATETIME,
	is_active     BOOLEAN NOT NULL DEFAULT 1,
	created_at    DATETIME NOT NULL,
	updated_at    DATETIME NOT NULL,
	version       INTEGER NOT NULL DEFAULT 1
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_peer_ip_active
ON peer_ip_allocations(tenant_id, allocated_ip)
WHERE is_active = 1;

CREATE UNIQUE INDEX IF NOT EXISTS uq_peer_active_allocation
ON peer_ip_allocations(peer_id)
WHERE is_active = 1;
`

// NormalizeSQLiteDSN returns a DSN suitable for multi-connection in-memory use.
// Empty or ":memory:" → shared in-memory DB so sqlx pool shares one database.
func NormalizeSQLiteDSN(path string) string {
	if path == "" || path == ":memory:" {
		return "file:score_peer_ipam?mode=memory&cache=shared"
	}
	return path
}

// OpenSQLitePeerIPAM opens SQLite, applies peer IPAM schema, and returns a Repository
// usable by RelayManagerHybrid (GetDB → PeerIPAMRepository).
// path may be a file path, ":memory:", or empty (shared memory).
func OpenSQLitePeerIPAM(path string, logger *zap.Logger) (*Repository, error) {
	if logger == nil {
		logger = zap.NewNop()
	}
	dsn := NormalizeSQLiteDSN(path)

	db, err := sqlx.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	// Single connection for :memory:/shared simplicity and serializable-ish behaviour.
	db.SetMaxOpenConns(1)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}

	if _, err := db.Exec(peerIPAMSchemaSQLite); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("apply peer IPAM schema: %w", err)
	}

	logger.Info("SQLite peer IPAM repository ready",
		zap.String("dsn", dsn),
	)
	return NewRepository(db, logger), nil
}

// OpenSQLitePeerIPAMDB is like OpenSQLitePeerIPAM but returns raw *sql.DB for tests.
func OpenSQLitePeerIPAMDB(path string) (*sql.DB, error) {
	dsn := NormalizeSQLiteDSN(path)
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(peerIPAMSchemaSQLite); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}
