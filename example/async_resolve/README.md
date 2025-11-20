# Async Resolve Examples

Comprehensive examples demonstrating the `Group` async resolve functionality for managing concurrent operations in Go.

## Overview

The async resolve pattern allows you to:
- Launch multiple async operations simultaneously
- Wait for all operations to complete
- Optionally enforce time limits with timeouts
- Collect results from all operations

This is similar to `Promise.all()` in JavaScript or `CompletableFuture.allOf()` in Java, but implemented using Go's native goroutines and sync.WaitGroup.

## Quick Start

### Build

```bash
go build -o async-resolve-demo
```

### Run All Examples

```bash
./async-resolve-demo
```

## 8 Runnable Examples

### 1. Basic Async Resolve
**File**: `main.go` - `demo1BasicAsyncResolve()`

Demonstrates the fundamental async resolve pattern with simple tasks.

```go
group := goroutine.NewGroup()

var result1, result2, result3 any

group.Assign(&result1, func() any {
    time.Sleep(500 * time.Millisecond)
    return "Task 1 completed"
})

group.Assign(&result2, func() any {
    time.Sleep(300 * time.Millisecond)
    return 42
})

group.Assign(&result3, func() any {
    time.Sleep(200 * time.Millisecond)
    return true
})

// Wait for all tasks to complete
group.Resolve()

fmt.Printf("Result 1: %v\n", result1) // "Task 1 completed"
fmt.Printf("Result 2: %v\n", result2) // 42
fmt.Printf("Result 3: %v\n", result3) // true
```

**Key Points:**
- All tasks run in parallel
- Total time is max task duration (~500ms), not sum
- Results available after Resolve()
- Simple type-agnostic interface

### 2. Parallel Data Fetching
**File**: `main.go` - `demo2ParallelDataFetching()`

Real-world scenario: Fetching data from multiple sources simultaneously.

```go
group := goroutine.NewGroup()
results := make([]any, len(sources))

for i, source := range sources {
    idx := i
    fetcher := source
    group.Assign(&results[idx], func() any {
        return fetcher.Fetch()
    })
}

group.Resolve()

// All results available simultaneously
for i, result := range results {
    fmt.Printf("Source %d: %v\n", i, result)
}
```

**Use Case:**
- Fetching from database, cache, and APIs
- No blocking - faster than sequential calls
- Great for aggregating data from multiple sources

**Expected Output:**
```
Starting fetches...
  Fetching from Database...
  Fetching from Cache...
  Fetching from API-1...
  ✓ Cache completed
  ✓ API-1 completed
  ✓ API-2 completed
  ✓ FileSystem completed
  ✓ Database completed

Total time: ~800ms (max fetch time)
```

### 3. Concurrent API Requests
**File**: `main.go` - `demo3ConcurrentAPIRequests()`

Simulates concurrent HTTP API calls with error codes.

```go
group := goroutine.NewGroup()
responses := make([]any, len(endpoints))

for i, endpoint := range endpoints {
    idx := i
    url := endpoint
    group.Assign(&responses[idx], func() any {
        resp, _ := http.Get(url)
        return APIResponse{
            Endpoint: url,
            Status:   resp.StatusCode,
            Time:     elapsed,
        }
    })
}

group.Resolve()
```

**Benefits:**
- Non-blocking concurrent API calls
- Better performance than sequential requests
- Collect all responses together
- Handle failures per-response

### 4. Computational Tasks
**File**: `main.go` - `demo4ComputationalTasks()`

Running CPU-intensive calculations in parallel.

```go
group := goroutine.NewGroup()
var fibResult, factorsResult, factorialResult any

group.Assign(&fibResult, func() any {
    return fibonacci(40)
})

group.Assign(&factorsResult, func() any {
    return primeFactors(123456789)
})

group.Assign(&factorialResult, func() any {
    return factorial(20)
})

group.Resolve()
```

**Demonstrations:**
- Fibonacci(40) calculation
- Prime factor decomposition
- Factorial computation
- Parallel execution gains multi-core efficiency

**Example Output:**
```
Starting computational tasks...
  Computing Fibonacci(40)...
  Computing prime factors of 123456789...
  Computing 20!...
  ✓ Fibonacci completed
  ✓ Prime factors completed
  ✓ Factorial completed

Results:
  Fibonacci(40) = 102334155
  Prime factors of 123456789 = [3 3 3607 3803]
  20! = 2432902008176640000

Total time: ~2s (all compute in parallel)
```

### 5. Timeout Handling
**File**: `main.go` - `demo5TimeoutHandling()`

Using `ResolveWithTimeout` for time-bounded operations.

```go
group := goroutine.NewGroup()
var result1, result2 any

group.Assign(&result1, func() any {
    time.Sleep(200 * time.Millisecond)
    return "Done"
})

group.Assign(&result2, func() any {
    time.Sleep(300 * time.Millisecond)
    return "Also done"
})

// Wait up to 1 second
if group.ResolveWithTimeout(1 * time.Second) {
    fmt.Println("All tasks completed in time")
    fmt.Printf("Results: %v, %v\n", result1, result2)
} else {
    fmt.Println("Timeout! Not all tasks completed")
    // Tasks continue running in background
}
```

**Return Values:**
- `true`: All tasks completed within timeout
- `false`: Timeout occurred before completion

**Important:**
- Tasks continue running even after timeout
- Only the wait is interrupted
- Useful for fallbacks or timeouts

### 6. Map-Reduce Pattern
**File**: `main.go` - `demo6MapReducePattern()`

Distributing work across goroutines then reducing results.

```go
chunks := [][]int{
    {1, 2, 3, 4, 5},
    {6, 7, 8, 9, 10},
    {11, 12, 13, 14, 15},
}

group := goroutine.NewGroup()
results := make([]any, len(chunks))

// Map phase: Process chunks in parallel
for i, chunk := range chunks {
    idx := i
    data := chunk
    group.Assign(&results[idx], func() any {
        sum := 0
        for _, val := range data {
            sum += val
        }
        return sum
    })
}

// Wait for all map operations
group.Resolve()

// Reduce phase: Combine results
totalSum := 0
for _, result := range results {
    if sum, ok := result.(int); ok {
        totalSum += sum
    }
}
```

**Pattern:**
1. **Map Phase**: Distribute work to goroutines
2. **Resolve**: Wait for all to complete
3. **Reduce Phase**: Combine results

**Real-World Uses:**
- Distributed data processing
- Batch processing with parallelization
- Load balancing across workers

### 7. Dependent Async Operations
**File**: `main.go` - `demo7DependentAsyncOperations()`

Multi-stage async operations where later stages depend on earlier results.

```go
// Stage 1: Fetch user IDs
group1 := goroutine.NewGroup()
var userIDs1, userIDs2 any

group1.Assign(&userIDs1, func() any {
    return []int{1, 2, 3}
})

group1.Assign(&userIDs2, func() any {
    return []int{4, 5, 6}
})

group1.Resolve() // Wait for Stage 1

// Stage 2: Fetch user details using IDs from Stage 1
group2 := goroutine.NewGroup()
results := make([]any, 0)

allIDs := append(userIDs1.([]int), userIDs2.([]int)...)

for _, userID := range allIDs {
    id := userID
    group2.Assign(&results, func() any {
        return fetchUserDetails(id)
    })
}

group2.Resolve() // Wait for Stage 2
```

**Use Cases:**
- Multi-stage pipelines
- ETL processes
- Request chains with dependent calls

### 8. Error Handling Pattern
**File**: `main.go` - `demo8ErrorHandling()`

Using async resolve with structured error handling.

```go
type Result struct {
    Success bool
    Value   interface{}
    Error   string
}

group := goroutine.NewGroup()
results := make([]any, numTasks)

for i := 0; i < numTasks; i++ {
    idx := i
    group.Assign(&results[idx], func() any {
        result, err := performTask()
        if err != nil {
            return Result{
                Success: false,
                Error:   err.Error(),
            }
        }
        return Result{
            Success: true,
            Value:   result,
        }
    })
}

group.Resolve()

// Process results
for _, result := range results {
    if res, ok := result.(Result); ok {
        if res.Success {
            fmt.Printf("Success: %v\n", res.Value)
        } else {
            fmt.Printf("Error: %s\n", res.Error)
        }
    }
}
```

**Pattern:**
- Each task returns structured Result
- Errors stored in Result type
- No panics propagate to caller
- Easy error aggregation and reporting

## API Reference

### Group Type

```go
type Group struct {
    tasks []*Task[any]
    mu    sync.Mutex
    wg    sync.WaitGroup
}
```

### NewGroup()

Creates a new task group.

```go
group := goroutine.NewGroup()
```

### Assign(result *any, fn func() any)

Assigns a function to run asynchronously.

```go
var result any

group.Assign(&result, func() any {
    // Do async work
    return computedValue
})
```

**Parameters:**
- `result`: Pointer to store the result
- `fn`: Function returning any type

**Behavior:**
- Launches goroutine immediately
- Result updated when task completes
- Results available after Resolve()

### Resolve()

Waits for all assigned tasks to complete.

```go
group.Resolve()
// All tasks completed here
```

**Behavior:**
- Blocks until all goroutines finish
- Uses sync.WaitGroup internally
- No timeout (infinite wait)

### ResolveWithTimeout(timeout time.Duration) bool

Waits with timeout.

```go
if group.ResolveWithTimeout(5 * time.Second) {
    fmt.Println("All tasks completed")
} else {
    fmt.Println("Timeout occurred")
    // Tasks still running in background
}
```

**Return:**
- `true`: All tasks completed within timeout
- `false`: Timeout occurred

**Important:**
- Tasks continue running after timeout
- Only the wait is interrupted

## Patterns & Best Practices

### 1. Type Assertion Pattern

```go
var result any
group.Assign(&result, func() any {
    return 42
})
group.Resolve()

// Type assert before use
if num, ok := result.(int); ok {
    fmt.Printf("Got number: %d\n", num)
}
```

### 2. Slice Results Pattern

```go
results := make([]any, numTasks)
for i := 0; i < numTasks; i++ {
    idx := i
    group.Assign(&results[idx], func() any {
        return doWork(i)
    })
}
group.Resolve()

// All results in order
```

### 3. Error Handling Pattern

```go
type TaskResult struct {
    Value any
    Error error
}

group.Assign(&result, func() any {
    val, err := riskyOperation()
    return TaskResult{Value: val, Error: err}
})
```

### 4. Timeout with Fallback

```go
group := goroutine.NewGroup()
var cacheResult any

group.Assign(&cacheResult, func() any {
    return fetchFromDatabase()
})

if !group.ResolveWithTimeout(100 * time.Millisecond) {
    // Use fallback
    cacheResult = getDefaultValue()
}
```

### 5. Multi-Stage Pipeline

```go
// Stage 1
group1 := goroutine.NewGroup()
var stage1Results any
// ... assign tasks ...
group1.Resolve()

// Stage 2 using Stage 1 results
group2 := goroutine.NewGroup()
// ... use stage1Results for Stage 2 tasks ...
group2.Resolve()
```

## Performance Characteristics

| Scenario | Benefit |
|----------|---------|
| 3 tasks (100ms each) | ~100ms total (vs 300ms sequential) |
| 5 API calls (200ms each) | ~200ms total (vs 1000ms sequential) |
| 4 CPU-bound tasks | Near-linear speedup on 4-core CPU |
| Timeout handling | Quick failure detection |

## Comparison with Alternatives

### vs sync.WaitGroup

```go
// With Group (simpler)
group := goroutine.NewGroup()
group.Assign(&result, func() any { return value })
group.Resolve()

// With WaitGroup (more verbose)
var wg sync.WaitGroup
var result any
wg.Add(1)
go func() {
    defer wg.Done()
    result = value
}()
wg.Wait()
```

### vs sync/errgroup

```go
// Group: No error handling built-in, handle in Result struct
// errgroup: Returns first error automatically
// Choose based on your error handling needs
```

### vs channels

```go
// Group: Simpler for parallel same-type tasks
// Channels: Better for streaming, different types, pipelines
// Use Group when you need promise-all semantics
```

## Running the Examples

```bash
# Build
go build -o async-resolve-demo

# Run all examples
./async-resolve-demo

# Run and pipe to less (for detailed review)
./async-resolve-demo | less

# Benchmark (approximate)
time ./async-resolve-demo
```

## Expected Output

All 8 examples run sequentially, demonstrating:

1. Basic async execution
2. Parallel data fetching
3. Concurrent API calls
4. Computational parallelism
5. Timeout behavior (success and failure cases)
6. Map-reduce computation
7. Multi-stage pipelines
8. Error aggregation

## Troubleshooting

### Results are nil/empty

Ensure the result pointer is passed correctly:

```go
// ✓ Correct
var result any
group.Assign(&result, fn)

// ✗ Wrong
var result any
group.Assign(result, fn)  // Wrong! Pass pointer
```

### Type assertion panics

Always use safe type assertion:

```go
// ✓ Safe
if val, ok := result.(int); ok {
    // Use val
}

// ✗ Unsafe
val := result.(int)  // Panics if wrong type
```

### Tasks still running after ResolveWithTimeout returns false

This is expected behavior. Tasks continue executing in background:

```go
if !group.ResolveWithTimeout(timeout) {
    // Tasks still running!
    // Don't access results yet
    // Or use a fallback value
}
```

## Advanced Usage

### Custom Result Type

```go
type APIResult struct {
    StatusCode int
    Body       []byte
    Error      string
    Duration   time.Duration
}

var results []any
for i, url := range urls {
    idx := i
    group.Assign(&results[idx], func() any {
        return APIResult{ /* ... */ }
    })
}
```

### Concurrent Map Building

```go
results := make([]any, len(keys))
for i, key := range keys {
    idx := i
    k := key
    group.Assign(&results[idx], func() any {
        return map[string]any{
            "key": k,
            "value": fetchValue(k),
        }
    })
}
```

### Progress Tracking

```go
completed := 0
mu := sync.Mutex{}

for i := 0; i < numTasks; i++ {
    idx := i
    group.Assign(&results[idx], func() any {
        result := doWork(i)
        mu.Lock()
        completed++
        fmt.Printf("Progress: %d/%d\n", completed, numTasks)
        mu.Unlock()
        return result
    })
}
```

## See Also

- Parent documentation: `../../README.md`
- SafeChannel examples: `../distibuted_backend/`
- Core implementation: `../../asyncresolve.go`
- More examples: `../../examples_test.go`
