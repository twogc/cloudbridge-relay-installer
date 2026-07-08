package cadence

import (
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestCreateWorkflowServiceClient_UnreachableIsStub(t *testing.T) {
	// Closed port: shared factory must return (nil, nil) stub mode, not hard error.
	client, err := CreateWorkflowServiceClient("127.0.0.1:1", zap.NewNop())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client != nil {
		t.Fatal("expected nil client in stub mode")
	}
}

func TestCreateWorkflowServiceClientWithOptions_UnreachableIsStub(t *testing.T) {
	client, err := CreateWorkflowServiceClientWithOptions("127.0.0.1:1", ServiceClientOptions{
		Timeout:     200 * time.Millisecond,
		ServiceName: "cadence-frontend",
	}, zap.NewNop())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client != nil {
		t.Fatal("expected nil client in stub mode")
	}
}

func TestCreateWorkflowServiceClientImproved_UnreachableIsStub(t *testing.T) {
	client, err := CreateWorkflowServiceClientImproved("127.0.0.1:1", zap.NewNop())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client != nil {
		t.Fatal("expected nil client in stub mode")
	}
}

func TestHealthCheck_NilClient(t *testing.T) {
	if err := HealthCheck(nil, zap.NewNop()); err == nil {
		t.Fatal("expected error for nil client")
	}
}
