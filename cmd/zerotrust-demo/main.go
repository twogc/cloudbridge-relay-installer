package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/twogc/cloudbridge-relay-installer/internal/circuitbreaker"
	"github.com/twogc/cloudbridge-relay-installer/internal/health"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	fmt.Println("=== Zero Trust Architecture Demo ===")
	fmt.Println()

	// Demo 1: Circuit Breaker
	fmt.Println("1. Circuit Breaker Demo")
	demoCircuitBreaker(logger)
	fmt.Println()

	// Demo 2: Health Checker
	fmt.Println("2. Health Checker Demo")
	demoHealthChecker(logger)
	fmt.Println()

	fmt.Println("=== Demo Complete ===")
}

func demoCircuitBreaker(logger *zap.Logger) {
	// Create circuit breaker with aggressive settings for demo
	cb := circuitbreaker.NewCircuitBreaker(&circuitbreaker.Config{
		Name:             "demo-circuit-breaker",
		FailureThreshold: 3,
		SuccessThreshold: 2,
		Timeout:          5 * time.Second,
		WindowSize:       10 * time.Second,
		Logger:           logger,
	})

	fmt.Printf("Circuit Breaker created: %s\n", cb.GetName())
	fmt.Printf("Initial state: %s\n", cb.GetState().String())

	// Simulate successful operations
	fmt.Println("\n--- Simulating successful operations ---")
	for i := 0; i < 5; i++ {
		err := cb.Execute(context.Background(), func() error {
			fmt.Printf("  Operation %d: SUCCESS\n", i+1)
			return nil
		})
		if err != nil {
			fmt.Printf("  Operation %d: FAILED - %v\n", i+1, err)
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Simulate failures to trigger circuit breaker
	fmt.Println("\n--- Simulating failures to trigger circuit breaker ---")
	for i := 0; i < 5; i++ {
		err := cb.Execute(context.Background(), func() error {
			fmt.Printf("  Operation %d: FAILURE\n", i+1)
			return fmt.Errorf("simulated failure %d", i+1)
		})
		if err != nil {
			fmt.Printf("  Operation %d: BLOCKED - %v\n", i+1, err)
		}
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Printf("Circuit breaker state after failures: %s\n", cb.GetState().String())

	// Wait for timeout and try again
	fmt.Println("\n--- Waiting for timeout and trying again ---")
	time.Sleep(6 * time.Second)

	// Try a successful operation to close the circuit
	err := cb.Execute(context.Background(), func() error {
		fmt.Println("  Recovery operation: SUCCESS")
		return nil
	})
	if err != nil {
		fmt.Printf("  Recovery operation: BLOCKED - %v\n", err)
	}

	fmt.Printf("Circuit breaker state after recovery: %s\n", cb.GetState().String())

	// Show stats
	stats := cb.GetStats()
	fmt.Printf("Circuit breaker stats: %+v\n", stats)
}

func demoHealthChecker(logger *zap.Logger) {
	// Create health checker
	hc := health.NewHealthChecker(&health.Config{
		Interval: 2 * time.Second,
		Timeout:  1 * time.Second,
		Logger:   logger,
	})

	// Add health checks
	hc.AddCheck("http_check", health.HTTPHealthCheck("http_check", "https://httpbin.org/status/200"))
	hc.AddCheck("ping_check", health.PingHealthCheck("ping_check", "google.com"))
	hc.AddCheck("custom_check", health.CustomHealthCheck(
		"custom_check",
		"Custom health check that always succeeds",
		func(ctx context.Context) error {
			fmt.Println("  Custom health check: SUCCESS")
			return nil
		},
	))

	fmt.Printf("Health Checker created with %d checks\n", len(hc.GetResults()))

	// Start health checker
	hc.Start()
	fmt.Println("Health checker started")

	// Monitor health for a few cycles
	fmt.Println("\n--- Monitoring health checks ---")
	for i := 0; i < 3; i++ {
		time.Sleep(3 * time.Second)
		
		status := hc.GetStatus()
		results := hc.GetResults()
		
		fmt.Printf("Cycle %d - Overall status: %s\n", i+1, status)
		for name, result := range results {
			fmt.Printf("  %s: %s (duration: %v)\n", name, result.Status, result.Duration)
			if result.LastError != nil {
				fmt.Printf("    Error: %v\n", result.LastError)
			}
		}
		fmt.Println()
	}

	// Stop health checker
	hc.Stop()
	fmt.Println("Health checker stopped")

	// Show final results
	fmt.Println("--- Final health check results ---")
	results := hc.GetResults()
	for name, result := range results {
		fmt.Printf("%s: %s\n", name, result.Status)
	}
} 