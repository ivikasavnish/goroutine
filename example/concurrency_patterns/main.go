package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ivikasavnish/goroutine"
)

func main() {
	fmt.Println("=== Go Concurrency Patterns Examples ===\n")

	// Run all examples
	pipelineExample()
	fanOutFanInExample()
	rateLimiterExample()
	semaphoreExample()
	generatorExample()
	smartWrapperBasicExample()
	smartWrapperRetryExample()
	smartWrapperCircuitBreakerExample()
	combinedPatternsExample()
}

// Example 1: Pipeline Pattern
func pipelineExample() {
	fmt.Println("--- Example 1: Pipeline Pattern ---")
	
	// Create a pipeline that processes numbers
	pipeline := goroutine.NewPipeline[int]().
		AddStage(func(n int) int { return n * 2 }).         // Double
		AddStage(func(n int) int { return n + 10 }).        // Add 10
		AddStage(func(n int) int { return n * n })          // Square

	// Process a single item
	result := pipeline.Execute(5)
	fmt.Printf("Pipeline result for 5: %d\n", result) // ((5*2)+10)^2 = 400

	// Process multiple items asynchronously
	items := []int{1, 2, 3, 4, 5}
	results := pipeline.ExecuteAsync(context.Background(), items)
	fmt.Printf("Pipeline results for %v: %v\n\n", items, results)
}

// Example 2: Fan-Out/Fan-In Pattern
func fanOutFanInExample() {
	fmt.Println("--- Example 2: Fan-Out/Fan-In Pattern ---")
	
	// Create fan-out with 3 workers
	fanOut := goroutine.NewFanOut[int, int](3)
	
	// Process items with multiple workers
	items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	ctx := context.Background()
	
	results := fanOut.ProcessWithIndex(ctx, items, func(idx int, n int) int {
		// Simulate work
		time.Sleep(100 * time.Millisecond)
		return n * n
	})
	
	fmt.Printf("Fan-out results: %v\n", results)
	fmt.Printf("Processed %d items with 3 workers\n\n", len(items))
}

// Example 3: Rate Limiter Pattern
func rateLimiterExample() {
	fmt.Println("--- Example 3: Rate Limiter Pattern ---")
	
	// Create rate limiter: 5 operations per second
	limiter := goroutine.NewRateLimiter(5)
	defer limiter.Stop()
	
	ctx := context.Background()
	start := time.Now()
	
	// Execute 10 operations with rate limiting
	for i := 0; i < 10; i++ {
		if err := limiter.Wait(ctx); err != nil {
			fmt.Printf("Error: %v\n", err)
			break
		}
		fmt.Printf("Operation %d executed at %.2fs\n", i+1, time.Since(start).Seconds())
	}
	
	fmt.Printf("Completed in %.2fs (expected ~2s for 10 ops at 5 ops/s)\n\n", time.Since(start).Seconds())
}

// Example 4: Semaphore Pattern
func semaphoreExample() {
	fmt.Println("--- Example 4: Semaphore Pattern ---")
	
	// Create semaphore with 3 permits (max 3 concurrent operations)
	sem := goroutine.NewSemaphore(3)
	ctx := context.Background()
	
	var wg sync.WaitGroup
	start := time.Now()
	
	// Launch 10 goroutines but only 3 can run concurrently
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			// Acquire permit
			if err := sem.Acquire(ctx); err != nil {
				fmt.Printf("Worker %d: failed to acquire: %v\n", id, err)
				return
			}
			defer sem.Release()
			
			// Simulate work
			fmt.Printf("Worker %d: started at %.2fs\n", id, time.Since(start).Seconds())
			time.Sleep(300 * time.Millisecond)
			fmt.Printf("Worker %d: finished at %.2fs\n", id, time.Since(start).Seconds())
		}(i)
	}
	
	wg.Wait()
	fmt.Printf("All workers completed in %.2fs\n\n", time.Since(start).Seconds())
}

// Example 5: Generator Pattern
func generatorExample() {
	fmt.Println("--- Example 5: Generator Pattern ---")
	
	// Create a generator that produces Fibonacci numbers
	ctx := context.Background()
	gen := goroutine.NewGenerator[int](ctx, 10, func(ctx context.Context, ch chan<- int) {
		a, b := 0, 1
		for i := 0; i < 10; i++ {
			select {
			case <-ctx.Done():
				return
			case ch <- a:
				a, b = b, a+b
			}
		}
	})
	defer gen.Close()
	
	// Collect all values
	fibs := gen.Collect()
	fmt.Printf("Generated Fibonacci numbers: %v\n\n", fibs)
}

// Example 6: Smart Wrapper - Basic Usage
func smartWrapperBasicExample() {
	fmt.Println("--- Example 6: Smart Wrapper - Basic Usage ---")
	
	// Configure concurrency wrapper
	config := &goroutine.ConcurrencyConfig{
		MaxConcurrency: 3,
		Timeout:        5 * time.Second,
		RateLimit:      0, // No rate limit
	}
	
	wrapper := goroutine.NewConcurrencyWrapper[int, int](config)
	defer wrapper.Close()
	
	// Process items with automatic concurrency control
	items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	ctx := context.Background()
	
	start := time.Now()
	results, err := wrapper.Process(ctx, items, func(n int) (int, error) {
		time.Sleep(200 * time.Millisecond) // Simulate work
		return n * n, nil
	})
	
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	
	fmt.Printf("Results: %v\n", results)
	fmt.Printf("Processed in %.2fs (max 3 concurrent)\n\n", time.Since(start).Seconds())
}

// Example 7: Smart Wrapper - With Retry
func smartWrapperRetryExample() {
	fmt.Println("--- Example 7: Smart Wrapper - With Retry ---")
	
	config := &goroutine.ConcurrencyConfig{
		MaxConcurrency: 5,
		Timeout:        10 * time.Second,
		RetryAttempts:  3,
		RetryDelay:     100 * time.Millisecond,
	}
	
	wrapper := goroutine.NewConcurrencyWrapper[int, string](config)
	defer wrapper.Close()
	
	ctx := context.Background()
	items := []int{1, 2, 3, 4, 5}
	
	// Simulate flaky operations
	attemptCount := make(map[int]int)
	var mu sync.Mutex
	
	results, err := wrapper.Process(ctx, items, func(n int) (string, error) {
		mu.Lock()
		attemptCount[n]++
		attempt := attemptCount[n]
		mu.Unlock()
		
		// Succeed on 3rd attempt
		if attempt < 3 {
			return "", fmt.Errorf("temporary failure")
		}
		return fmt.Sprintf("success-%d", n), nil
	})
	
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	
	fmt.Printf("Results: %v\n", results)
	fmt.Printf("Attempt counts: %v\n\n", attemptCount)
}

// Example 8: Smart Wrapper - Circuit Breaker
func smartWrapperCircuitBreakerExample() {
	fmt.Println("--- Example 8: Smart Wrapper - Circuit Breaker ---")
	
	config := &goroutine.ConcurrencyConfig{
		MaxConcurrency:          5,
		Timeout:                 10 * time.Second,
		EnableCircuitBreaker:    true,
		CircuitBreakerThreshold: 3,
	}
	
	wrapper := goroutine.NewConcurrencyWrapper[int, int](config)
	defer wrapper.Close()
	
	ctx := context.Background()
	items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	
	// Simulate failing service
	failCount := 0
	results, err := wrapper.Process(ctx, items, func(n int) (int, error) {
		failCount++
		// Fail first 5 attempts to trigger circuit breaker
		if failCount <= 5 {
			return 0, fmt.Errorf("service unavailable")
		}
		return n * 2, nil
	})
	
	fmt.Printf("Circuit breaker triggered after failures\n")
	fmt.Printf("Results: %v\n", results)
	fmt.Printf("Error: %v\n\n", err)
}

// Example 9: Combined Patterns
func combinedPatternsExample() {
	fmt.Println("--- Example 9: Combined Patterns ---")
	
	// Combine pipeline with smart wrapper
	pipeline := goroutine.NewPipeline[int]().
		AddStage(func(n int) int { return n * 2 }).
		AddStage(func(n int) int { return n + 10 })
	
	config := &goroutine.ConcurrencyConfig{
		MaxConcurrency: 3,
		Timeout:        5 * time.Second,
		RateLimit:      10,
	}
	
	wrapper := goroutine.NewConcurrencyWrapper[int, int](config)
	defer wrapper.Close()
	
	ctx := context.Background()
	items := []int{1, 2, 3, 4, 5}
	
	// First, process through pipeline
	pipelineResults := pipeline.ExecuteAsync(ctx, items)
	fmt.Printf("After pipeline: %v\n", pipelineResults)
	
	// Then process with wrapper (simulate additional processing)
	finalResults, err := wrapper.Process(ctx, pipelineResults, func(n int) (int, error) {
		time.Sleep(50 * time.Millisecond)
		return n * n, nil
	})
	
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	
	fmt.Printf("Final results: %v\n", finalResults)
}
