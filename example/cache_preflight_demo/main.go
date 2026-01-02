package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ivikasavnish/goroutine"
)

const (
	// DBLatencyMs simulates database latency
	DBLatencyMs = 500 * time.Millisecond
	// CacheLatencyMs simulates cache latency (unused but kept for reference)
	CacheLatencyMs = 50 * time.Millisecond
)

// SimulateDBFetch simulates a slow database fetch
func SimulateDBFetch(id string) string {
	time.Sleep(DBLatencyMs) // Simulate DB latency
	return fmt.Sprintf("DB Data for %s", id)
}

func main() {
	fmt.Println("=== Sync Preflight & Caching Examples ===\n")

	demo1CachedGroupBasic()
	fmt.Println("\n" + strings.Repeat("=", 50) + "\n")

	demo2PreflightFetcher()
	fmt.Println("\n" + strings.Repeat("=", 50) + "\n")

	demo3NoCacheMode()
	fmt.Println("\n" + strings.Repeat("=", 50) + "\n")

	demo4StaleWhileRevalidate()
	fmt.Println("\n" + strings.Repeat("=", 50) + "\n")

	demo5ReducedDownstreamLoad()
}

// Example 1: CachedGroup with automatic cache preflight
func demo1CachedGroupBasic() {
	fmt.Println("Demo 1: CachedGroup with Automatic Preflight")
	fmt.Println("First call fetches from DB, second call uses cache\n")

	cg := goroutine.NewCachedGroup()

	control := &goroutine.CacheControl{
		NoCache: false,
		MaxAge:  5 * time.Minute,
	}

	// First fetch - should call the function
	var result1 any
	fmt.Println("First fetch (cache miss)...")
	start := time.Now()

	cg.AssignWithCache("user:123", &result1, func() any {
		fmt.Println("  -> Fetching from DB...")
		return SimulateDBFetch("user:123")
	}, control)

	cg.Resolve()
	elapsed1 := time.Since(start)
	fmt.Printf("Result: %v (took %v)\n\n", result1, elapsed1)

	// Second fetch - should use cache (preflight hit)
	var result2 any
	fmt.Println("Second fetch (cache hit via preflight)...")
	start = time.Now()

	cg.AssignWithCache("user:123", &result2, func() any {
		fmt.Println("  -> Fetching from DB...")
		return SimulateDBFetch("user:123")
	}, control)

	cg.Resolve()
	elapsed2 := time.Since(start)
	fmt.Printf("Result: %v (took %v)\n", result2, elapsed2)
	fmt.Printf("\nSpeedup: %.1fx faster with cache preflight\n", float64(elapsed1)/float64(elapsed2))
}

// Example 2: PreflightFetcher for reducing downstream load
func demo2PreflightFetcher() {
	fmt.Println("Demo 2: PreflightFetcher Pattern")
	fmt.Println("Demonstrates cache-first pattern to reduce DB load\n")

	// Simulate DB calls counter
	dbCallCount := 0

	fetchFunc := func(ctx context.Context, key string) (string, error) {
		dbCallCount++
		fmt.Printf("  -> DB Call #%d for key: %s\n", dbCallCount, key)
		time.Sleep(DBLatencyMs)
		return fmt.Sprintf("Data-%s", key), nil
	}

	control := &goroutine.CacheControl{
		NoCache: false,
		MaxAge:  1 * time.Minute,
	}

	fetcher := goroutine.NewPreflightFetcher(fetchFunc, control)
	ctx := context.Background()

	// Fetch same key multiple times
	fmt.Println("Fetching 'user:1' three times:")
	for i := 1; i <= 3; i++ {
		start := time.Now()
		val, err := fetcher.Fetch(ctx, "user:1")
		elapsed := time.Since(start)

		if err != nil {
			fmt.Printf("  [%d] Error: %v\n", i, err)
		} else {
			fmt.Printf("  [%d] Result: %s (took %v)\n", i, val, elapsed)
		}
	}

	fmt.Printf("\nTotal DB calls: %d (saved 2 calls via preflight caching)\n", dbCallCount)
	fmt.Println("✓ Downstream load reduced by 66%")
}

// Example 3: NoCache mode for fresh data
func demo3NoCacheMode() {
	fmt.Println("Demo 3: NoCache Mode")
	fmt.Println("Force fetch from source, bypassing cache preflight\n")

	callCount := 0
	fetchFunc := func(ctx context.Context, key string) (string, error) {
		callCount++
		fmt.Printf("  -> Fetch #%d\n", callCount)
		time.Sleep(200 * time.Millisecond)
		return fmt.Sprintf("Fresh-%d", callCount), nil
	}

	// With NoCache enabled
	control := &goroutine.CacheControl{
		NoCache: true, // Bypass cache preflight
		MaxAge:  1 * time.Minute,
	}

	fetcher := goroutine.NewPreflightFetcher(fetchFunc, control)
	ctx := context.Background()

	fmt.Println("Fetching with NoCache=true (bypass preflight):")
	for i := 1; i <= 3; i++ {
		val, _ := fetcher.Fetch(ctx, "key1")
		fmt.Printf("  [%d] %s\n", i, val)
	}

	fmt.Printf("\nTotal fetches: %d (no preflight caching)\n", callCount)
}

// Example 4: Stale-While-Revalidate pattern
func demo4StaleWhileRevalidate() {
	fmt.Println("Demo 4: Stale-While-Revalidate")
	fmt.Println("Return stale data immediately, revalidate in background\n")

	fetchFunc := func(ctx context.Context, key string) (string, error) {
		fmt.Println("  -> Revalidating from source...")
		time.Sleep(DBLatencyMs)
		return fmt.Sprintf("Fresh-%d", time.Now().Unix()), nil
	}

	control := &goroutine.CacheControl{
		NoCache:              false,
		MaxAge:               100 * time.Millisecond, // Very short TTL
		StaleWhileRevalidate: true,
	}

	fetcher := goroutine.NewPreflightFetcher(fetchFunc, control)
	ctx := context.Background()

	// Initial fetch
	fmt.Println("Initial fetch:")
	val1, _ := fetcher.Fetch(ctx, "key1")
	fmt.Printf("  Result: %s\n\n", val1)

	// Wait for cache to expire
	time.Sleep(200 * time.Millisecond)

	// Fetch with stale data
	fmt.Println("Fetch after expiration (stale-while-revalidate):")
	start := time.Now()
	val2, _ := fetcher.FetchStaleWhileRevalidate(ctx, "key1")
	elapsed := time.Since(start)
	fmt.Printf("  Result: %s (took %v - instant!)\n", val2, elapsed)
	fmt.Println("  Background revalidation in progress...")

	// Wait for revalidation
	time.Sleep(600 * time.Millisecond)

	// Next fetch gets fresh data from cache
	fmt.Println("\nNext fetch (revalidated data):")
	val3, _ := fetcher.Fetch(ctx, "key1")
	fmt.Printf("  Result: %s\n", val3)
}

// Example 5: Demonstrating reduced downstream load
func demo5ReducedDownstreamLoad() {
	fmt.Println("Demo 5: Reduced Downstream Load")
	fmt.Println("Compare DB load with and without preflight caching\n")

	// Scenario: Fetching user data multiple times
	dbCallsWithoutCache := 0
	dbCallsWithCache := 0

	// WITHOUT preflight caching
	fmt.Println("WITHOUT Preflight Caching:")
	for i := 1; i <= 5; i++ {
		dbCallsWithoutCache++
		SimulateDBFetch("user:123")
		fmt.Printf("  Request %d: DB call #%d\n", i, dbCallsWithoutCache)
	}

	fmt.Println("\nWITH Preflight Caching:")
	fetchFunc := func(ctx context.Context, key string) (string, error) {
		dbCallsWithCache++
		return SimulateDBFetch(key), nil
	}

	control := &goroutine.CacheControl{
		NoCache: false,
		MaxAge:  1 * time.Minute,
	}

	fetcher := goroutine.NewPreflightFetcher(fetchFunc, control)
	ctx := context.Background()

	for i := 1; i <= 5; i++ {
		fetcher.Fetch(ctx, "user:123")
		fmt.Printf("  Request %d: DB call #%d (cached: %v)\n",
			i, dbCallsWithCache, i > 1)
	}

	fmt.Println("\n" + strings.Repeat("─", 40))
	fmt.Printf("DB calls without cache: %d\n", dbCallsWithoutCache)
	fmt.Printf("DB calls with cache:    %d\n", dbCallsWithCache)
	fmt.Printf("Load reduction:         %d%%\n",
		(dbCallsWithoutCache-dbCallsWithCache)*100/dbCallsWithoutCache)
	fmt.Println("✓ Downstream load reduced significantly!")
}
