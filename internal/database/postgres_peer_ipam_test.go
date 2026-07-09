package database

import (
	"testing"
)

func TestPostgresPeerIPAMFromEnv_URL(t *testing.T) {
	t.Setenv("SCORE_DATABASE_URL", "postgres://u:p@db:5432/cloudbridge_p2p?sslmode=disable")
	t.Setenv("DATABASE_URL", "")
	t.Setenv("SCORE_PG_HOST", "")
	cfg := PostgresPeerIPAMFromEnv()
	if !cfg.Enabled() {
		t.Fatal("expected enabled with SCORE_DATABASE_URL")
	}
	dsn, err := cfg.DSN()
	if err != nil {
		t.Fatal(err)
	}
	if dsn == "" {
		t.Fatal("empty dsn")
	}
}

func TestPostgresPeerIPAMFromEnv_Discrete(t *testing.T) {
	t.Setenv("SCORE_DATABASE_URL", "")
	t.Setenv("DATABASE_URL", "")
	t.Setenv("SCORE_PG_HOST", "pg.internal")
	t.Setenv("SCORE_PG_PORT", "5433")
	t.Setenv("SCORE_PG_USER", "cb")
	t.Setenv("SCORE_PG_PASSWORD", "secret")
	t.Setenv("SCORE_PG_DATABASE", "ipam")
	t.Setenv("SCORE_PG_SSLMODE", "require")
	cfg := PostgresPeerIPAMFromEnv()
	if !cfg.Enabled() {
		t.Fatal("expected enabled with SCORE_PG_HOST")
	}
	dsn, err := cfg.DSN()
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"host=pg.internal", "port=5433", "dbname=ipam", "user=cb", "sslmode=require"} {
		if !contains(dsn, want) {
			t.Fatalf("dsn %q missing %q", dsn, want)
		}
	}
}

func TestPostgresPeerIPAM_Disabled(t *testing.T) {
	t.Setenv("SCORE_DATABASE_URL", "")
	t.Setenv("DATABASE_URL", "")
	t.Setenv("SCORE_PG_HOST", "")
	cfg := PostgresPeerIPAMFromEnv()
	if cfg.Enabled() {
		t.Fatal("expected disabled")
	}
	if _, err := OpenPostgresPeerIPAM(cfg, nil); err == nil {
		t.Fatal("expected error when not configured")
	}
}

func TestPostgresPeerIPAM_ConnectFail(t *testing.T) {
	cfg := PostgresPeerIPAMConfig{
		Host:     "127.0.0.1",
		Port:     1,
		User:     "x",
		Password: "y",
		Database: "z",
		SSLMode:  "disable",
	}
	_, err := OpenPostgresPeerIPAM(cfg, nil)
	if err == nil {
		t.Fatal("expected connect error on closed port")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		(len(s) > 0 && (func() bool {
			for i := 0; i+len(sub) <= len(s); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		})()))
}
