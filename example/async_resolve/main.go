package main

import (
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/ivikasavnish/goroutine"
)

// DataFetcher simulates fetching data from different sources
type DataFetcher struct {
	name     string
	delay    time.Duration
	success  bool
}

// Fetch simulates a network request
func (df *DataFetcher) Fetch() string {
	time.Sleep(df.delay)
	if df.success {
		return fmt.Sprintf("Data from %s", df.name)
	}
	return fmt.Sprintf("Error from %s", df.name)
}

func main() {
	fmt.Println("=== Async Resolve Examples ===\n")

	demo1BasicAsyncResolve()
	fmt.Println("\n" + strings.Repeat("=", 50) + "\n")

	demo2ParallelDataFetching()
	fmt.Println("\n" + strings.Repeat("=", 50) + "\n")

	demo3ConcurrentAPIRequests()
	fmt.Println("\n" + strings.Repeat("=", 50) + "\n")

	demo4ComputationalTasks()
	fmt.Println("\n" + strings.Repeat("=", 50) + "\n")

	demo5TimeoutHandling()
	fmt.Println("\n" + strings.Repeat("=", 50) + "\n")

	demo6MapReducePattern()
	fmt.Println("\n" + strings.Repeat("=", 50) + "\n")

	demo7DependentAsyncOperations()
	fmt.Println("\n" + strings.Repeat("=", 50) + "\n")

	demo8ErrorHandling()
}

// Example 1: Basic async resolve with simple tasks
func demo1BasicAsyncResolve() {
	fmt.Println("Demo 1: Basic Async Resolve")
	fmt.Println("Running 3 concurrent tasks and waiting for all to complete\n")

	group := goroutine.NewGroup()

	var task1Result, task2Result, task3Result any

	// Task 1: Simple string operation
	group.Assign(&task1Result, func() any {
		time.Sleep(500 * time.Millisecond)
		return "Task 1 completed"
	})

	// Task 2: Integer computation
	group.Assign(&task2Result, func() any {
		time.Sleep(300 * time.Millisecond)
		return 42
	})

	// Task 3: Boolean operation
	group.Assign(&task3Result, func() any {
		time.Sleep(200 * time.Millisecond)
		return true
	})

	fmt.Println("Starting tasks...")
	startTime := time.Now()

	// Wait for all tasks to complete
	group.Resolve()

	elapsed := time.Since(startTime)

	fmt.Printf("Task 1 Result: %v\n", task1Result)
	fmt.Printf("Task 2 Result: %v\n", task2Result)
	fmt.Printf("Task 3 Result: %v\n", task3Result)
	fmt.Printf("Total time: %v (max task time is ~500ms, all run in parallel)\n", elapsed)
}

// Example 2: Parallel data fetching from multiple sources
func demo2ParallelDataFetching() {
	fmt.Println("Demo 2: Parallel Data Fetching")
	fmt.Println("Fetching data from 5 different sources in parallel\n")

	group := goroutine.NewGroup()

	sources := []DataFetcher{
		{name: "Database", delay: 800 * time.Millisecond, success: true},
		{name: "Cache", delay: 100 * time.Millisecond, success: true},
		{name: "API-1", delay: 500 * time.Millisecond, success: true},
		{name: "API-2", delay: 600 * time.Millisecond, success: true},
		{name: "FileSystem", delay: 300 * time.Millisecond, success: true},
	}

	results := make([]any, len(sources))

	fmt.Println("Starting fetches...")
	startTime := time.Now()

	for i, source := range sources {
		idx := i
		fetcher := source
		group.Assign(&results[idx], func() any {
			fmt.Printf("  Fetching from %s...\n", fetcher.name)
			data := fetcher.Fetch()
			fmt.Printf("  ✓ %s completed\n", fetcher.name)
			return data
		})
	}

	group.Resolve()
	elapsed := time.Since(startTime)

	fmt.Println("\nResults:")
	for i, result := range results {
		fmt.Printf("  [%d] %v\n", i+1, result)
	}
	fmt.Printf("\nTotal time: %v (fastest would be ~800ms, all run in parallel)\n", elapsed)
}

// Example 3: Concurrent API requests with processing
func demo3ConcurrentAPIRequests() {
	fmt.Println("Demo 3: Concurrent API Requests")
	fmt.Println("Simulating 4 concurrent HTTP API calls\n")

	type APIResponse struct {
		Endpoint string
		Status   int
		Data     string
		Time     time.Duration
	}

	group := goroutine.NewGroup()

	endpoints := []string{
		"https://api.example.com/users",
		"https://api.example.com/posts",
		"https://api.example.com/comments",
		"https://api.example.com/products",
	}

	responses := make([]any, len(endpoints))

	fmt.Println("Starting API calls...")
	startTime := time.Now()

	for i, endpoint := range endpoints {
		idx := i
		url := endpoint
		group.Assign(&responses[idx], func() any {
			callStart := time.Now()
			delay := time.Duration(rand.Intn(800-200) + 200) * time.Millisecond
			time.Sleep(delay)
			callTime := time.Since(callStart)

			// Simulate API response codes
			statusCode := 200
			if rand.Float32() < 0.1 {
				statusCode = 500
			}

			fmt.Printf("  API Call: %s - Status: %d (took %v)\n", url, statusCode, callTime)
			return APIResponse{
				Endpoint: url,
				Status:   statusCode,
				Data:     fmt.Sprintf("Response from %s", url),
				Time:     callTime,
			}
		})
	}

	group.Resolve()
	elapsed := time.Since(startTime)

	fmt.Println("\nAPI Responses Processed:")
	for i, resp := range responses {
		if apiResp, ok := resp.(APIResponse); ok {
			fmt.Printf("  [%d] %s -> Status %d (Time: %v)\n", i+1, apiResp.Endpoint, apiResp.Status, apiResp.Time)
		}
	}
	fmt.Printf("\nTotal time: %v\n", elapsed)
}

// Example 4: Computational tasks in parallel
func demo4ComputationalTasks() {
	fmt.Println("Demo 4: Computational Tasks")
	fmt.Println("Running CPU-intensive calculations in parallel\n")

	// Function to calculate Fibonacci number
	fib := func(n int) int {
		if n <= 1 {
			return n
		}
		a, b := 0, 1
		for i := 2; i <= n; i++ {
			a, b = b, a+b
		}
		return b
	}

	// Function to calculate prime factors
	primeFactors := func(n int) []int {
		factors := []int{}
		for d := 2; d*d <= n; d++ {
			for n%d == 0 {
				factors = append(factors, d)
				n /= d
			}
		}
		if n > 1 {
			factors = append(factors, n)
		}
		return factors
	}

	// Function to calculate factorial
	factorial := func(n int) uint64 {
		result := uint64(1)
		for i := 2; i <= n; i++ {
			result *= uint64(i)
		}
		return result
	}

	group := goroutine.NewGroup()
	var fibResult, factorsResult, factorialResult any

	fmt.Println("Starting computational tasks...")
	startTime := time.Now()

	// Task 1: Fibonacci
	group.Assign(&fibResult, func() any {
		fmt.Println("  Computing Fibonacci(40)...")
		result := fib(40)
		fmt.Println("  ✓ Fibonacci completed")
		return result
	})

	// Task 2: Prime factors
	group.Assign(&factorsResult, func() any {
		fmt.Println("  Computing prime factors of 123456789...")
		result := primeFactors(123456789)
		fmt.Println("  ✓ Prime factors completed")
		return result
	})

	// Task 3: Factorial
	group.Assign(&factorialResult, func() any {
		fmt.Println("  Computing 20!...")
		result := factorial(20)
		fmt.Println("  ✓ Factorial completed")
		return result
	})

	group.Resolve()
	elapsed := time.Since(startTime)

	fmt.Println("\nResults:")
	fmt.Printf("  Fibonacci(40) = %v\n", fibResult)
	fmt.Printf("  Prime factors of 123456789 = %v\n", factorsResult)
	fmt.Printf("  20! = %v\n", factorialResult)
	fmt.Printf("\nTotal time: %v\n", elapsed)
}

// Example 5: Timeout handling
func demo5TimeoutHandling() {
	fmt.Println("Demo 5: Timeout Handling")
	fmt.Println("Using ResolveWithTimeout for time-bounded operations\n")

	fmt.Println("Test 1: Operations complete within timeout")
	group1 := goroutine.NewGroup()
	var res1, res2 any

	group1.Assign(&res1, func() any {
		time.Sleep(200 * time.Millisecond)
		return "Task 1 done"
	})
	group1.Assign(&res2, func() any {
		time.Sleep(300 * time.Millisecond)
		return "Task 2 done"
	})

	fmt.Println("  Waiting with 1 second timeout...")
	startTime := time.Now()
	if group1.ResolveWithTimeout(1 * time.Second) {
		elapsed := time.Since(startTime)
		fmt.Printf("  ✓ All tasks completed in %v\n", elapsed)
		fmt.Printf("    Results: %v, %v\n", res1, res2)
	} else {
		fmt.Println("  ✗ Timeout occurred")
	}

	fmt.Println("\nTest 2: Operations timeout")
	group2 := goroutine.NewGroup()
	var res3, res4 any

	group2.Assign(&res3, func() any {
		time.Sleep(2 * time.Second)
		return "This will timeout"
	})
	group2.Assign(&res4, func() any {
		time.Sleep(1 * time.Second)
		return "This also times out"
	})

	fmt.Println("  Waiting with 500ms timeout...")
	startTime = time.Now()
	if group2.ResolveWithTimeout(500 * time.Millisecond) {
		elapsed := time.Since(startTime)
		fmt.Printf("  ✓ All tasks completed in %v\n", elapsed)
	} else {
		elapsed := time.Since(startTime)
		fmt.Printf("  ✗ Timeout occurred after %v\n", elapsed)
		fmt.Println("    (Tasks continue running in background)")
	}
}

// Example 6: Map-reduce pattern
func demo6MapReducePattern() {
	fmt.Println("Demo 6: Map-Reduce Pattern")
	fmt.Println("Distributing work across multiple goroutines and reducing results\n")

	type WorkChunk struct {
		id    int
		data  []int
		sum   int
	}

	// Sample data split into chunks
	chunks := [][]int{
		{1, 2, 3, 4, 5},
		{6, 7, 8, 9, 10},
		{11, 12, 13, 14, 15},
		{16, 17, 18, 19, 20},
	}

	group := goroutine.NewGroup()
	results := make([]any, len(chunks))

	fmt.Println("Mapping phase: Processing chunks in parallel...")

	for i, chunk := range chunks {
		idx := i
		data := chunk
		group.Assign(&results[idx], func() any {
			sum := 0
			for _, val := range data {
				sum += val
			}
			fmt.Printf("  Chunk %d: sum = %d\n", idx, sum)
			return sum
		})
	}

	group.Resolve()

	fmt.Println("\nReducing phase: Combining results...")
	totalSum := 0
	for i, result := range results {
		if sum, ok := result.(int); ok {
			fmt.Printf("  Adding chunk %d sum: %d\n", i, sum)
			totalSum += sum
		}
	}

	fmt.Printf("\nTotal sum: %d (expected: %d)\n", totalSum, 20*21/2)
}

// Example 7: Dependent async operations
func demo7DependentAsyncOperations() {
	fmt.Println("Demo 7: Dependent Async Operations")
	fmt.Println("First resolve a batch, then use results for another batch\n")

	// Stage 1: Fetch user IDs
	fmt.Println("Stage 1: Fetching user IDs...")
	group1 := goroutine.NewGroup()
	var userIDs1, userIDs2 any

	group1.Assign(&userIDs1, func() any {
		time.Sleep(200 * time.Millisecond)
		fmt.Println("  ✓ User IDs batch 1 fetched: [1, 2, 3]")
		return []int{1, 2, 3}
	})

	group1.Assign(&userIDs2, func() any {
		time.Sleep(200 * time.Millisecond)
		fmt.Println("  ✓ User IDs batch 2 fetched: [4, 5, 6]")
		return []int{4, 5, 6}
	})

	startTime := time.Now()
	group1.Resolve()
	stage1Time := time.Since(startTime)
	fmt.Printf("Stage 1 completed in %v\n\n", stage1Time)

	// Stage 2: Fetch user details using IDs from Stage 1
	fmt.Println("Stage 2: Fetching user details...")
	group2 := goroutine.NewGroup()

	allIDs := []int{}
	if ids1, ok := userIDs1.([]int); ok {
		allIDs = append(allIDs, ids1...)
	}
	if ids2, ok := userIDs2.([]int); ok {
		allIDs = append(allIDs, ids2...)
	}

	results := make([]any, len(allIDs))

	for i, userID := range allIDs {
		idx := i
		id := userID
		group2.Assign(&results[idx], func() any {
			time.Sleep(time.Duration(rand.Intn(200-100)+100) * time.Millisecond)
			name := fmt.Sprintf("User-%d", id)
			fmt.Printf("  ✓ User detail fetched: %s\n", name)
			return name
		})
	}

	startTime = time.Now()
	group2.Resolve()
	stage2Time := time.Since(startTime)
	fmt.Printf("Stage 2 completed in %v\n\n", stage2Time)

	fmt.Println("Results:")
	for i, result := range results {
		fmt.Printf("  [%d] %v\n", i+1, result)
	}
	fmt.Printf("Total time: %v\n", stage1Time+stage2Time)
}

// Example 8: Error handling pattern
func demo8ErrorHandling() {
	fmt.Println("Demo 8: Error Handling Pattern")
	fmt.Println("Using async resolve with error handling\n")

	type Result struct {
		Success bool
		Value   interface{}
		Error   string
	}

	performTask := func(name string, shouldFail bool) Result {
		time.Sleep(time.Duration(rand.Intn(300-100)+100) * time.Millisecond)

		if shouldFail {
			fmt.Printf("  ✗ %s failed\n", name)
			return Result{
				Success: false,
				Error:   fmt.Sprintf("%s encountered an error", name),
			}
		}

		fmt.Printf("  ✓ %s succeeded\n", name)
		return Result{
			Success: true,
			Value:   fmt.Sprintf("Data from %s", name),
		}
	}

	group := goroutine.NewGroup()
	results := make([]any, 5)

	fmt.Println("Running tasks with error handling...")

	tasks := []struct {
		name      string
		shouldFail bool
	}{
		{"Task-1", false},
		{"Task-2", true},
		{"Task-3", false},
		{"Task-4", true},
		{"Task-5", false},
	}

	for i, task := range tasks {
		idx := i
		t := task
		group.Assign(&results[idx], func() any {
			return performTask(t.name, t.shouldFail)
		})
	}

	startTime := time.Now()
	group.Resolve()
	elapsed := time.Since(startTime)

	fmt.Println("\nResults Summary:")
	successCount := 0
	failureCount := 0

	for i, result := range results {
		if res, ok := result.(Result); ok {
			status := "✓"
			if !res.Success {
				status = "✗"
				failureCount++
			} else {
				successCount++
			}
			fmt.Printf("  [%d] %s Status: %v\n", i+1, status, res.Success)
			if res.Success {
				fmt.Printf("       Value: %v\n", res.Value)
			} else {
				fmt.Printf("       Error: %s\n", res.Error)
			}
		}
	}

	fmt.Printf("\nSummary: %d succeeded, %d failed\n", successCount, failureCount)
	fmt.Printf("Total time: %v\n", elapsed)
}

// Helper function: Parallel sqrt calculations
func parallelSqrt(numbers []float64) []float64 {
	group := goroutine.NewGroup()
	results := make([]any, len(numbers))

	for i, num := range numbers {
		idx := i
		val := num
		group.Assign(&results[idx], func() any {
			return math.Sqrt(val)
		})
	}

	group.Resolve()

	sqrts := make([]float64, len(results))
	for i, result := range results {
		if val, ok := result.(float64); ok {
			sqrts[i] = val
		}
	}

	return sqrts
}

// Helper function: Parallel HTTP client
func getURL(url string) (int, error) {
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	return resp.StatusCode, nil
}
