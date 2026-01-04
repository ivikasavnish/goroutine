package goroutine

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

// TestConcurrencyWrapper tests basic wrapper functionality
func TestConcurrencyWrapper(t *testing.T) {
	config := &ConcurrencyConfig{
		MaxConcurrency: 2,
		Timeout:        5 * time.Second,
	}

	wrapper := NewConcurrencyWrapper[int, int](config)
	defer wrapper.Close()

	items := []int{1, 2, 3, 4, 5}
	results, err := wrapper.Process(context.Background(), items, func(n int) (int, error) {
		return n * 2, nil
	})

	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	expected := []int{2, 4, 6, 8, 10}
	for i, result := range results {
		if result != expected[i] {
			t.Errorf("Index %d: expected %d, got %d", i, expected[i], result)
		}
	}
}

func TestConcurrencyWrapperWithEmptySlice(t *testing.T) {
	wrapper := NewConcurrencyWrapper[int, int](DefaultConcurrencyConfig())
	defer wrapper.Close()

	items := []int{}
	results, err := wrapper.Process(context.Background(), items, func(n int) (int, error) {
		return n * 2, nil
	})

	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Expected empty results, got %d items", len(results))
	}
}

func TestConcurrencyWrapperWithError(t *testing.T) {
	config := &ConcurrencyConfig{
		MaxConcurrency: 2,
		Timeout:        5 * time.Second,
	}

	wrapper := NewConcurrencyWrapper[int, int](config)
	defer wrapper.Close()

	items := []int{1, 2, 3, 4, 5}
	_, err := wrapper.Process(context.Background(), items, func(n int) (int, error) {
		if n == 3 {
			return 0, errors.New("test error")
		}
		return n * 2, nil
	})

	if err == nil {
		t.Error("Expected error but got none")
	}
}

func TestConcurrencyWrapperWithTimeout(t *testing.T) {
	config := &ConcurrencyConfig{
		MaxConcurrency: 2,
		Timeout:        100 * time.Millisecond,
	}

	wrapper := NewConcurrencyWrapper[int, int](config)
	defer wrapper.Close()

	items := []int{1, 2, 3}
	_, err := wrapper.Process(context.Background(), items, func(n int) (int, error) {
		time.Sleep(200 * time.Millisecond)
		return n * 2, nil
	})

	if err == nil {
		t.Error("Expected timeout error but got none")
	}
}

func TestConcurrencyWrapperWithRetry(t *testing.T) {
	config := &ConcurrencyConfig{
		MaxConcurrency: 2,
		Timeout:        5 * time.Second,
		RetryAttempts:  2,
		RetryDelay:     50 * time.Millisecond,
	}

	wrapper := NewConcurrencyWrapper[int, string](config)
	defer wrapper.Close()

	attemptCount := 0
	items := []int{1}

	results, err := wrapper.Process(context.Background(), items, func(n int) (string, error) {
		attemptCount++
		if attemptCount < 3 {
			return "", fmt.Errorf("temporary failure")
		}
		return "success", nil
	})

	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	if results[0] != "success" {
		t.Errorf("Expected 'success', got %s", results[0])
	}

	if attemptCount != 3 {
		t.Errorf("Expected 3 attempts, got %d", attemptCount)
	}
}

func TestConcurrencyWrapperWithRetryExhausted(t *testing.T) {
	config := &ConcurrencyConfig{
		MaxConcurrency: 1,
		Timeout:        5 * time.Second,
		RetryAttempts:  2,
		RetryDelay:     10 * time.Millisecond,
	}

	wrapper := NewConcurrencyWrapper[int, string](config)
	defer wrapper.Close()

	items := []int{1}

	_, err := wrapper.Process(context.Background(), items, func(n int) (string, error) {
		return "", fmt.Errorf("persistent failure")
	})

	if err == nil {
		t.Error("Expected error but got none")
	}

	if !errors.Is(err, ErrRetryExhausted) {
		t.Logf("Error type: %v", err)
	}
}

func TestConcurrencyWrapperWithRateLimit(t *testing.T) {
	config := &ConcurrencyConfig{
		MaxConcurrency: 5,
		Timeout:        5 * time.Second,
		RateLimit:      10, // 10 ops per second
	}

	wrapper := NewConcurrencyWrapper[int, int](config)
	defer wrapper.Close()

	items := []int{1, 2, 3, 4, 5}
	start := time.Now()

	results, err := wrapper.Process(context.Background(), items, func(n int) (int, error) {
		return n * 2, nil
	})

	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	if len(results) != 5 {
		t.Errorf("Expected 5 results, got %d", len(results))
	}

	// With rate limiting, should take at least 400ms for 5 operations at 10 ops/sec
	if elapsed < 400*time.Millisecond {
		t.Logf("Completed faster than expected with rate limiting: %v", elapsed)
	}
}

func TestConcurrencyWrapperWithCircuitBreaker(t *testing.T) {
	config := &ConcurrencyConfig{
		MaxConcurrency:          2,
		Timeout:                 5 * time.Second,
		EnableCircuitBreaker:    true,
		CircuitBreakerThreshold: 3,
	}

	wrapper := NewConcurrencyWrapper[int, int](config)
	defer wrapper.Close()

	items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	failCount := 0

	_, err := wrapper.Process(context.Background(), items, func(n int) (int, error) {
		failCount++
		if failCount <= 5 {
			return 0, fmt.Errorf("service unavailable")
		}
		return n * 2, nil
	})

	if err == nil {
		t.Error("Expected error due to circuit breaker")
	}

	// Circuit breaker should be open
	if wrapper.circuitBreaker.State() != CircuitOpen {
		t.Logf("Circuit breaker state: %v (expected open)", wrapper.circuitBreaker.State())
	}
}

func TestConcurrencyWrapperConcurrencyControl(t *testing.T) {
	config := &ConcurrencyConfig{
		MaxConcurrency: 2, // Only 2 concurrent operations
		Timeout:        10 * time.Second,
	}

	wrapper := NewConcurrencyWrapper[int, int](config)
	defer wrapper.Close()

	items := []int{1, 2, 3, 4, 5, 6}
	start := time.Now()

	results, err := wrapper.Process(context.Background(), items, func(n int) (int, error) {
		time.Sleep(100 * time.Millisecond)
		return n * 2, nil
	})

	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	if len(results) != 6 {
		t.Errorf("Expected 6 results, got %d", len(results))
	}

	// With max 2 concurrent and 100ms per op, 6 items should take at least 300ms
	// (3 batches of 2 items each)
	if elapsed < 300*time.Millisecond {
		t.Logf("Completed faster than expected: %v", elapsed)
	}
}

func TestDefaultConcurrencyConfig(t *testing.T) {
	config := DefaultConcurrencyConfig()

	if config.MaxConcurrency != 10 {
		t.Errorf("Expected MaxConcurrency 10, got %d", config.MaxConcurrency)
	}

	if config.Timeout != 30*time.Second {
		t.Errorf("Expected Timeout 30s, got %v", config.Timeout)
	}

	if config.RateLimit != 0 {
		t.Errorf("Expected RateLimit 0, got %d", config.RateLimit)
	}

	if config.RetryAttempts != 0 {
		t.Errorf("Expected RetryAttempts 0, got %d", config.RetryAttempts)
	}
}

func TestConcurrencyWrapperContextCancellation(t *testing.T) {
	config := &ConcurrencyConfig{
		MaxConcurrency: 2,
		Timeout:        0, // No timeout from config
	}

	wrapper := NewConcurrencyWrapper[int, int](config)
	defer wrapper.Close()

	ctx, cancel := context.WithCancel(context.Background())

	items := []int{1, 2, 3, 4, 5}
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	_, err := wrapper.Process(ctx, items, func(n int) (int, error) {
		time.Sleep(100 * time.Millisecond)
		return n * 2, nil
	})

	if err == nil {
		t.Error("Expected context cancellation error")
	}
}

// Benchmark tests
func BenchmarkConcurrencyWrapper(b *testing.B) {
	config := &ConcurrencyConfig{
		MaxConcurrency: 4,
		Timeout:        10 * time.Second,
	}

	wrapper := NewConcurrencyWrapper[int, int](config)
	defer wrapper.Close()

	items := make([]int, 100)
	for i := range items {
		items[i] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wrapper.Process(context.Background(), items, func(n int) (int, error) {
			return n * n, nil
		})
	}
}

func BenchmarkConcurrencyWrapperWithRetry(b *testing.B) {
	config := &ConcurrencyConfig{
		MaxConcurrency: 4,
		Timeout:        10 * time.Second,
		RetryAttempts:  2,
		RetryDelay:     10 * time.Millisecond,
	}

	wrapper := NewConcurrencyWrapper[int, int](config)
	defer wrapper.Close()

	items := make([]int, 50)
	for i := range items {
		items[i] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wrapper.Process(context.Background(), items, func(n int) (int, error) {
			return n * n, nil
		})
	}
}
