package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ivikasavnish/goroutine"
)

func main() {
	fmt.Println("=== SwissMap Examples ===")
	fmt.Println()

	// Example 1: Basic Usage
	example1BasicUsage()

	// Example 2: GetOrSet and GetOrCompute
	example2GetOrSetAndCompute()

	// Example 3: Concurrent Access
	example3ConcurrentAccess()

	// Example 4: Range and Iteration
	example4RangeAndIteration()

	// Example 5: Different Key Types
	example5DifferentKeyTypes()

	// Example 6: High-Performance Concurrent Workload
	example6HighPerformance()
}

func example1BasicUsage() {
	fmt.Println("Example 1: Basic Usage")
	fmt.Println("----------------------")

	// Create a new SwissMap
	sm := goroutine.NewSwissMap[string, int]()

	// Set values
	sm.Set("apple", 5)
	sm.Set("banana", 3)
	sm.Set("orange", 7)

	// Get values
	if val, ok := sm.Get("apple"); ok {
		fmt.Printf("apple: %d\n", val)
	}

	// Check existence
	if sm.Has("banana") {
		fmt.Println("banana exists in the map")
	}

	// Get map length
	fmt.Printf("Map size: %d\n", sm.Len())

	// Delete a key
	sm.Delete("banana")
	fmt.Printf("After deleting banana, size: %d\n", sm.Len())

	fmt.Println()
}

func example2GetOrSetAndCompute() {
	fmt.Println("Example 2: GetOrSet and GetOrCompute")
	fmt.Println("-------------------------------------")

	sm := goroutine.NewSwissMap[string, int]()

	// GetOrSet - set value if not present
	val1, existed := sm.GetOrSet("counter", 0)
	fmt.Printf("First GetOrSet: value=%d, existed=%v\n", val1, existed)

	val2, existed := sm.GetOrSet("counter", 100)
	fmt.Printf("Second GetOrSet: value=%d, existed=%v\n", val2, existed)

	// GetOrCompute - compute value only if not present
	computeCount := 0
	compute := func() int {
		computeCount++
		fmt.Printf("  Computing value (call #%d)...\n", computeCount)
		return 42
	}

	result1 := sm.GetOrCompute("lazy", compute)
	fmt.Printf("First GetOrCompute: %d\n", result1)

	result2 := sm.GetOrCompute("lazy", compute)
	fmt.Printf("Second GetOrCompute: %d (computed only once)\n", result2)

	fmt.Println()
}

func example3ConcurrentAccess() {
	fmt.Println("Example 3: Concurrent Access")
	fmt.Println("-----------------------------")

	sm := goroutine.NewSwissMap[int, string]()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	const numWorkers = 10
	const operationsPerWorker = 100

	// Multiple goroutines writing concurrently
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < operationsPerWorker; j++ {
				select {
				case <-ctx.Done():
					return
				default:
					key := workerID*operationsPerWorker + j
					sm.Set(key, fmt.Sprintf("worker-%d-value-%d", workerID, j))
				}
			}
		}(i)
	}

	// Multiple goroutines reading concurrently
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < operationsPerWorker; j++ {
				select {
				case <-ctx.Done():
					return
				default:
					key := workerID*operationsPerWorker + j
					sm.Get(key)
				}
			}
		}(i)
	}

	wg.Wait()
	fmt.Printf("Concurrent operations completed. Final map size: %d\n", sm.Len())
	fmt.Println()
}

func example4RangeAndIteration() {
	fmt.Println("Example 4: Range and Iteration")
	fmt.Println("-------------------------------")

	sm := goroutine.NewSwissMap[string, int]()

	// Populate the map
	fruits := map[string]int{
		"apple":      5,
		"banana":     3,
		"orange":     7,
		"grape":      10,
		"watermelon": 2,
	}

	for k, v := range fruits {
		sm.Set(k, v)
	}

	// Range over all entries
	fmt.Println("All fruits:")
	sm.Range(func(key string, value int) bool {
		fmt.Printf("  %s: %d\n", key, value)
		return true
	})

	// Early stop in Range
	fmt.Println("\nFirst 3 fruits (early stop):")
	count := 0
	sm.Range(func(key string, value int) bool {
		fmt.Printf("  %s: %d\n", key, value)
		count++
		return count < 3
	})

	// Get all keys
	keys := sm.Keys()
	fmt.Printf("\nAll keys: %v\n", keys)

	// Get all values
	values := sm.Values()
	fmt.Printf("All values: %v\n", values)

	// Convert to regular map
	regularMap := sm.ToMap()
	fmt.Printf("Converted to map: %d entries\n", len(regularMap))

	fmt.Println()
}

func example5DifferentKeyTypes() {
	fmt.Println("Example 5: Different Key Types")
	fmt.Println("-------------------------------")

	// String keys
	stringMap := goroutine.NewSwissMap[string, string]()
	stringMap.Set("name", "John Doe")
	stringMap.Set("city", "New York")
	fmt.Printf("String map: name=%s, city=%s\n", 
		mustGet(stringMap, "name"), 
		mustGet(stringMap, "city"))

	// Int keys
	intMap := goroutine.NewSwissMap[int, string]()
	intMap.Set(1, "first")
	intMap.Set(2, "second")
	fmt.Printf("Int map: 1=%s, 2=%s\n", 
		mustGet(intMap, 1), 
		mustGet(intMap, 2))

	// Int64 keys
	int64Map := goroutine.NewSwissMap[int64, float64]()
	int64Map.Set(1000000000000, 3.14159)
	fmt.Printf("Int64 map: 1000000000000=%f\n", 
		mustGet(int64Map, 1000000000000))

	// Struct values
	type User struct {
		Name  string
		Email string
		Age   int
	}

	userMap := goroutine.NewSwissMap[string, User]()
	userMap.Set("user123", User{"Alice", "alice@example.com", 30})
	userMap.Set("user456", User{"Bob", "bob@example.com", 25})

	if user, ok := userMap.Get("user123"); ok {
		fmt.Printf("User: %s, Email: %s, Age: %d\n", user.Name, user.Email, user.Age)
	}

	fmt.Println()
}

func example6HighPerformance() {
	fmt.Println("Example 6: High-Performance Concurrent Workload")
	fmt.Println("------------------------------------------------")

	// Create SwissMap with custom shard count for even better performance
	sm := goroutine.NewSwissMapWithShards[int, int](64)

	start := time.Now()
	const numGoroutines = 100
	const numOperations = 10000

	var wg sync.WaitGroup

	// Simulate high-concurrency workload
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := id*numOperations + j
				// Mix of operations
				switch j % 3 {
				case 0:
					sm.Set(key, key*2)
				case 1:
					sm.Get(key)
				case 2:
					sm.GetOrCompute(key, func() int { return key * 3 })
				}
			}
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(start)

	totalOps := numGoroutines * numOperations
	opsPerSecond := float64(totalOps) / elapsed.Seconds()

	fmt.Printf("Completed %d operations in %v\n", totalOps, elapsed)
	fmt.Printf("Operations per second: %.0f\n", opsPerSecond)
	fmt.Printf("Final map size: %d entries\n", sm.Len())

	fmt.Println()
}

// Helper function to get value or panic (for example purposes)
func mustGet[K comparable, V any](sm *goroutine.SwissMap[K, V], key K) V {
	val, ok := sm.Get(key)
	if !ok {
		panic(fmt.Sprintf("key %v not found", key))
	}
	return val
}
