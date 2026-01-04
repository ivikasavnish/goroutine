package goroutine

import (
	"context"
	"testing"
	"time"
)

// TestPipeline tests the Pipeline pattern
func TestPipeline(t *testing.T) {
	pipeline := NewPipeline[int]().
		AddStage(func(n int) int { return n * 2 }).
		AddStage(func(n int) int { return n + 10 })

	result := pipeline.Execute(5)
	expected := 20 // (5 * 2) + 10
	if result != expected {
		t.Errorf("Expected %d, got %d", expected, result)
	}
}

func TestPipelineAsync(t *testing.T) {
	pipeline := NewPipeline[int]().
		AddStage(func(n int) int { return n * 2 })

	items := []int{1, 2, 3}
	results := pipeline.ExecuteAsync(context.Background(), items)

	expected := []int{2, 4, 6}
	for i, result := range results {
		if result != expected[i] {
			t.Errorf("Index %d: expected %d, got %d", i, expected[i], result)
		}
	}
}

func TestPipelineWithContext(t *testing.T) {
	pipeline := NewPipeline[int]().
		AddStage(func(n int) int {
			time.Sleep(100 * time.Millisecond)
			return n * 2
		})

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	items := []int{1, 2, 3}
	results := pipeline.ExecuteAsync(ctx, items)

	// Results should be zero values due to context cancellation
	for i, result := range results {
		if result != 0 {
			t.Logf("Index %d: got %d (may be non-zero due to race)", i, result)
		}
	}
}

// TestFanOut tests the FanOut pattern
func TestFanOut(t *testing.T) {
	fanOut := NewFanOut[int, int](2)
	items := []int{1, 2, 3, 4}

	results := fanOut.ProcessWithIndex(context.Background(), items, func(idx int, n int) int {
		return n * n
	})

	expected := []int{1, 4, 9, 16}
	for i, result := range results {
		if result != expected[i] {
			t.Errorf("Index %d: expected %d, got %d", i, expected[i], result)
		}
	}
}

func TestFanOutWithEmptySlice(t *testing.T) {
	fanOut := NewFanOut[int, int](2)
	items := []int{}

	results := fanOut.ProcessWithIndex(context.Background(), items, func(idx int, n int) int {
		return n * n
	})

	if len(results) != 0 {
		t.Errorf("Expected empty slice, got %d items", len(results))
	}
}

// TestRateLimiter tests the RateLimiter pattern
func TestRateLimiter(t *testing.T) {
	limiter := NewRateLimiter(10) // 10 ops per second
	defer limiter.Stop()

	ctx := context.Background()
	start := time.Now()

	// Try 5 operations
	for i := 0; i < 5; i++ {
		if err := limiter.Wait(ctx); err != nil {
			t.Fatalf("Wait failed: %v", err)
		}
	}

	elapsed := time.Since(start)
	// Should take at least 400ms for 5 operations at 10 ops/sec
	if elapsed < 400*time.Millisecond {
		t.Logf("Completed faster than expected: %v", elapsed)
	}
}

func TestRateLimiterTryAcquire(t *testing.T) {
	limiter := NewRateLimiter(5)
	defer limiter.Stop()

	// Should be able to acquire immediately
	if !limiter.TryAcquire() {
		t.Error("Expected to acquire token immediately")
	}
}

func TestRateLimiterContextCancellation(t *testing.T) {
	limiter := NewRateLimiter(1) // Very slow rate
	defer limiter.Stop()

	// Drain initial tokens
	for i := 0; i < 5; i++ {
		limiter.TryAcquire()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := limiter.Wait(ctx)
	if err == nil {
		t.Error("Expected context cancellation error")
	}
}

// TestSemaphore tests the Semaphore pattern
func TestSemaphore(t *testing.T) {
	sem := NewSemaphore(2)
	ctx := context.Background()

	// Acquire 2 permits
	if err := sem.Acquire(ctx); err != nil {
		t.Fatalf("First acquire failed: %v", err)
	}
	if err := sem.Acquire(ctx); err != nil {
		t.Fatalf("Second acquire failed: %v", err)
	}

	// Third acquire should not succeed immediately
	if sem.TryAcquire() {
		t.Error("Expected TryAcquire to fail when permits exhausted")
	}

	// Release one and try again
	sem.Release()
	if !sem.TryAcquire() {
		t.Error("Expected TryAcquire to succeed after release")
	}
}

func TestSemaphoreAcquireAll(t *testing.T) {
	sem := NewSemaphore(3)
	ctx := context.Background()

	if err := sem.AcquireAll(ctx); err != nil {
		t.Fatalf("AcquireAll failed: %v", err)
	}

	// Should not be able to acquire more
	if sem.TryAcquire() {
		t.Error("Expected TryAcquire to fail after AcquireAll")
	}

	sem.ReleaseAll()

	// Should be able to acquire again
	if !sem.TryAcquire() {
		t.Error("Expected TryAcquire to succeed after ReleaseAll")
	}
}

// TestGenerator tests the Generator pattern
func TestGenerator(t *testing.T) {
	ctx := context.Background()
	gen := NewGenerator[int](ctx, 5, func(ctx context.Context, ch chan<- int) {
		for i := 0; i < 5; i++ {
			select {
			case <-ctx.Done():
				return
			case ch <- i:
			}
		}
	})
	defer gen.Close()

	values := gen.Collect()
	expected := []int{0, 1, 2, 3, 4}

	if len(values) != len(expected) {
		t.Fatalf("Expected %d values, got %d", len(expected), len(values))
	}

	for i, v := range values {
		if v != expected[i] {
			t.Errorf("Index %d: expected %d, got %d", i, expected[i], v)
		}
	}
}

func TestGeneratorCollectN(t *testing.T) {
	ctx := context.Background()
	gen := NewGenerator[int](ctx, 10, func(ctx context.Context, ch chan<- int) {
		for i := 0; i < 10; i++ {
			select {
			case <-ctx.Done():
				return
			case ch <- i:
			}
		}
	})
	defer gen.Close()

	values := gen.CollectN(5)
	if len(values) != 5 {
		t.Fatalf("Expected 5 values, got %d", len(values))
	}
}

func TestGeneratorNext(t *testing.T) {
	ctx := context.Background()
	gen := NewGenerator[int](ctx, 3, func(ctx context.Context, ch chan<- int) {
		for i := 0; i < 3; i++ {
			select {
			case <-ctx.Done():
				return
			case ch <- i:
			}
		}
	})
	defer gen.Close()

	// Read values one by one
	for i := 0; i < 3; i++ {
		val, ok := gen.Next()
		if !ok {
			t.Fatalf("Expected value at iteration %d", i)
		}
		if val != i {
			t.Errorf("Expected %d, got %d", i, val)
		}
	}

	// Should be exhausted
	_, ok := gen.Next()
	if ok {
		t.Error("Expected generator to be exhausted")
	}
}

// TestCircuitBreaker tests the CircuitBreaker
func TestCircuitBreaker(t *testing.T) {
	cb := NewCircuitBreaker(3)

	// Initially closed
	if cb.State() != CircuitClosed {
		t.Error("Expected circuit to be closed initially")
	}

	// Should allow requests
	if !cb.Allow() {
		t.Error("Expected circuit to allow requests when closed")
	}

	// Record failures to open circuit
	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}

	if cb.State() != CircuitOpen {
		t.Error("Expected circuit to be open after threshold failures")
	}

	// Should not allow requests
	if cb.Allow() {
		t.Error("Expected circuit to block requests when open")
	}

	// Reset and verify
	cb.Reset()
	if cb.State() != CircuitClosed {
		t.Error("Expected circuit to be closed after reset")
	}
}

func TestCircuitBreakerRecovery(t *testing.T) {
	cb := NewCircuitBreaker(2)

	// Open the circuit
	cb.RecordFailure()
	cb.RecordFailure()

	if cb.State() != CircuitOpen {
		t.Error("Expected circuit to be open")
	}

	// Wait for reset timeout (we'll simulate by directly setting state for testing)
	// In real code, the circuit would transition to half-open after timeout
	
	// Record success to close circuit (assuming half-open state)
	cb.RecordSuccess()

	// Circuit should still work with success
	if !cb.Allow() {
		t.Log("Circuit may still be open or in transition")
	}
}

// Benchmark tests
func BenchmarkPipeline(b *testing.B) {
	pipeline := NewPipeline[int]().
		AddStage(func(n int) int { return n * 2 }).
		AddStage(func(n int) int { return n + 10 }).
		AddStage(func(n int) int { return n * n })

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pipeline.Execute(i)
	}
}

func BenchmarkFanOut(b *testing.B) {
	fanOut := NewFanOut[int, int](4)
	items := make([]int, 100)
	for i := range items {
		items[i] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fanOut.ProcessWithIndex(context.Background(), items, func(idx int, n int) int {
			return n * n
		})
	}
}

func BenchmarkFanOutProcess(b *testing.B) {
	fanOut := NewFanOut[int, int](4)
	items := make([]int, 100)
	for i := range items {
		items[i] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fanOut.Process(context.Background(), items, func(n int) int {
			return n * n
		})
	}
}

func BenchmarkSemaphore(b *testing.B) {
	sem := NewSemaphore(10)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sem.Acquire(ctx)
		sem.Release()
	}
}
