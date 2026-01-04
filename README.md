# Goroutine - Advanced Concurrent Processing for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/ivikasavnish/goroutine.svg)](https://pkg.go.dev/github.com/ivikasavnish/goroutine)
[![Go Report Card](https://goreportcard.com/badge/github.com/ivikasavnish/goroutine)](https://goreportcard.com/report/github.com/ivikasavnish/goroutine)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A powerful Go library providing advanced concurrent processing utilities, including async task resolution, safe channel operations, parallel slice processing, and flexible goroutine management.

## Features

### 🚀 **Async Resolve (Promise-like Pattern)**
- Launch multiple async operations simultaneously
- Wait for all operations to complete (similar to `Promise.all()`)
- Timeout support for time-bounded operations
- Type-agnostic result collection

### 🔒 **SafeChannel**
- Thread-safe channel wrapper with timeout capabilities
- Distributed backend support with multiple backend strategies
- Error handling for closed channels and timeouts
- Context-aware operations

### ⚡ **SuperSlice (Parallel Slice Processing)**
- Automatic parallelization based on configurable thresholds
- Worker pool management for efficient processing
- In-place updates to save memory
- Support for map, filter, forEach operations with type transformations
- Error handling in processing callbacks

### 🎯 **GoManager (Goroutine Management)**
- Named goroutine management with cancellation support
- Context-based lifecycle management
- Dynamic goroutine launching and cancellation

### 🔄 **Type Recasting**
- Flexible type conversion utilities
- Safe type transformations

### 🔀 **Concurrency Patterns (NEW)**
- **Pipeline**: Composable multi-stage data processing
- **Fan-Out/Fan-In**: Distribute work across workers and aggregate results
- **Rate Limiter**: Control operation execution rate
- **Semaphore**: Resource access control with permits
- **Generator**: Lazy value production with context support
- **Circuit Breaker**: Fault tolerance and failure handling

### 🎁 **Smart Concurrency Wrapper (NEW)**
- Unified API for concurrency control
- Automatic retry with configurable delays
- Timeout management
- Rate limiting integration
- Circuit breaker for fault tolerance
- Composable with other patterns

## Installation

```bash
go get github.com/ivikasavnish/goroutine
```

## Quick Start

### Async Resolve

Launch multiple async operations and wait for completion:

```go
package main

import (
    "fmt"
    "time"
    "github.com/ivikasavnish/goroutine"
)

func main() {
    group := goroutine.NewGroup()
    
    var result1, result2, result3 any
    
    // Launch async operations
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
}
```

### SuperSlice - Parallel Slice Processing

Process large slices efficiently with automatic parallelization:

```go
package main

import (
    "fmt"
    "github.com/ivikasavnish/goroutine"
)

func main() {
    // Create a large slice
    numbers := make([]int, 10000)
    for i := range numbers {
        numbers[i] = i + 1
    }
    
    // Process with automatic parallelization
    ss := goroutine.NewSuperSlice(numbers)
    result := ss.Process(func(index int, item int) int {
        return item * 2
    })
    
    fmt.Printf("Processed %d items\n", len(result))
}
```

### SuperSlice - Custom Configuration

Configure threshold and worker count:

```go
package main

import (
    "fmt"
    "github.com/ivikasavnish/goroutine"
)

func main() {
    data := []int{1, 2, 3, 4, 5}
    
    // Use fluent API for configuration
    result := goroutine.NewSuperSlice(data).
        WithThreshold(500).      // Switch to parallel at 500 items
        WithWorkers(8).          // Use 8 worker goroutines
        WithIterable().          // Use iterable processing mode
        Process(func(index int, item int) int {
            return item * 10
        })
    
    fmt.Println(result) // [10 20 30 40 50]
}
```

### SafeChannel

Thread-safe channel operations with timeout support:

```go
package main

import (
    "fmt"
    "time"
    "github.com/ivikasavnish/goroutine"
)

func main() {
    // Create a safe channel with buffer size 10 and 5-second timeout
    sc := goroutine.NewSafeChannel[int](10, 5*time.Second)
    
    // Send with timeout
    err := sc.SendWithTimeout(42, 1*time.Second)
    if err != nil {
        fmt.Printf("Send error: %v\n", err)
        return
    }
    
    // Receive with timeout
    value, err := sc.ReceiveWithTimeout(1 * time.Second)
    if err != nil {
        fmt.Printf("Receive error: %v\n", err)
        return
    }
    
    fmt.Printf("Received: %d\n", value) // 42
    
    // Close when done
    sc.Close()
}
```

### GoManager

Manage named goroutines with cancellation:

```go
package main

import (
    "fmt"
    "time"
    "github.com/ivikasavnish/goroutine"
)

func main() {
    manager := goroutine.NewGoManager()
    
    // Launch a named goroutine
    manager.GO("worker1", func() {
        for i := 0; i < 10; i++ {
            fmt.Printf("Worker 1: %d\n", i)
            time.Sleep(500 * time.Millisecond)
        }
    })
    
    // Let it run for 2 seconds
    time.Sleep(2 * time.Second)
    
    // Cancel the goroutine
    manager.Cancel("worker1")
    fmt.Println("Worker cancelled")
}
```

### Pipeline Pattern

Create composable data processing pipelines:

```go
package main

import (
    "context"
    "fmt"
    "github.com/ivikasavnish/goroutine"
)

func main() {
    // Create a pipeline with multiple stages
    pipeline := goroutine.NewPipeline[int]().
        AddStage(func(n int) int { return n * 2 }).
        AddStage(func(n int) int { return n + 10 }).
        AddStage(func(n int) int { return n * n })
    
    // Process single item
    result := pipeline.Execute(5) // ((5*2)+10)^2 = 400
    fmt.Printf("Result: %d\n", result)
    
    // Process multiple items asynchronously
    items := []int{1, 2, 3, 4, 5}
    results := pipeline.ExecuteAsync(context.Background(), items)
    fmt.Printf("Results: %v\n", results)
}
```

### Fan-Out/Fan-In Pattern

Distribute work across multiple workers:

```go
package main

import (
    "context"
    "fmt"
    "github.com/ivikasavnish/goroutine"
)

func main() {
    // Create fan-out with 3 workers
    fanOut := goroutine.NewFanOut[int, int](3)
    
    items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
    results := fanOut.ProcessWithIndex(context.Background(), items, 
        func(idx int, n int) int {
            return n * n
        })
    
    fmt.Printf("Results: %v\n", results)
}
```

### Smart Concurrency Wrapper

Unified API with retry, timeout, rate limiting, and circuit breaker:

```go
package main

import (
    "context"
    "fmt"
    "time"
    "github.com/ivikasavnish/goroutine"
)

func main() {
    // Configure wrapper
    config := &goroutine.ConcurrencyConfig{
        MaxConcurrency:          3,
        Timeout:                 5 * time.Second,
        RateLimit:               10,
        RetryAttempts:           2,
        RetryDelay:              100 * time.Millisecond,
        EnableCircuitBreaker:    true,
        CircuitBreakerThreshold: 5,
    }
    
    wrapper := goroutine.NewConcurrencyWrapper[int, int](config)
    defer wrapper.Close()
    
    items := []int{1, 2, 3, 4, 5}
    results, err := wrapper.Process(context.Background(), items, 
        func(n int) (int, error) {
            // Your processing logic here
            return n * 2, nil
        })
    
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    fmt.Printf("Results: %v\n", results)
}
```

## API Reference

### Async Resolve

#### `NewGroup() *Group`
Creates a new task group for managing async operations.

#### `(*Group) Assign(result *any, fn func() any)`
Assigns a function to run asynchronously. The result is stored in the provided pointer when the task completes.

#### `(*Group) Resolve()`
Blocks until all assigned tasks complete.

#### `(*Group) ResolveWithTimeout(timeout time.Duration) bool`
Waits for tasks to complete with a timeout. Returns `true` if all tasks completed, `false` if timeout occurred.

### SuperSlice

#### `NewSuperSlice[T any](slice []T) *SuperSlice[T]`
Creates a new SuperSlice from a slice.

#### `NewSuperSliceWithConfig[T any](slice []T, config *SuperSliceConfig) *SuperSlice[T]`
Creates a new SuperSlice with custom configuration.

#### Configuration Methods (Fluent API):
- `WithThreshold(threshold int) *SuperSlice[T]` - Set parallelization threshold (default: 1000)
- `WithWorkers(numWorkers int) *SuperSlice[T]` - Set number of worker goroutines (default: NumCPU)
- `WithIterable() *SuperSlice[T]` - Enable iterable processing mode
- `WithInPlace() *SuperSlice[T]` - Enable in-place updates

#### Processing Methods:
- `Process(callback func(int, T) T) []T` - Transform slice items
- `ProcessWithError(callback func(int, T) (T, error)) ([]T, error)` - Transform with error handling
- `ForEach(callback func(int, T))` - Iterate without collecting results
- `FilterSlice(predicate func(int, T) bool) []T` - Filter items in parallel

#### `MapTo[T, U any](ss *SuperSlice[T], mapper func(int, T) U) []U`
Transform slice items to a different type.

### SafeChannel

#### `NewSafeChannel[T any](bufferSize int, defaultTimeout time.Duration) *SafeChannel[T]`
Creates a new thread-safe channel wrapper.

#### `(*SafeChannel[T]) Send(value T) error`
Sends a value with default timeout.

#### `(*SafeChannel[T]) SendWithTimeout(value T, timeout time.Duration) error`
Sends a value with custom timeout.

#### `(*SafeChannel[T]) Receive() (T, error)`
Receives a value with default timeout.

#### `(*SafeChannel[T]) ReceiveWithTimeout(timeout time.Duration) (T, error)`
Receives a value with custom timeout.

#### `(*SafeChannel[T]) Close() error`
Safely closes the channel.

### GoManager

#### `NewGoManager() *GoManager`
Creates a new goroutine manager.

#### `(*GoManager) GO(name string, fn interface{}, argv ...interface{})`
Launches a named goroutine with cancellation support.

#### `(*GoManager) Cancel(name string)`
Cancels a named goroutine.

#### `(*GoManager) AddCancelFunc(name string, cancelFunc context.CancelFunc)`
Adds a cancel function for a named goroutine.

### Concurrency Patterns

#### `NewPipeline[T any]() *Pipeline[T]`
Creates a new pipeline for composable data processing.

#### `(*Pipeline[T]) AddStage(stage func(T) T) *Pipeline[T]`
Adds a processing stage to the pipeline. Returns the pipeline for chaining.

#### `(*Pipeline[T]) Execute(item T) T`
Executes the pipeline on a single item.

#### `(*Pipeline[T]) ExecuteAsync(ctx context.Context, items []T) []T`
Processes items through the pipeline concurrently.

#### `NewFanOut[T, R any](numWorkers int) *FanOut[T, R]`
Creates a fan-out pattern with specified number of workers.

#### `(*FanOut[T, R]) ProcessWithIndex(ctx context.Context, items []T, processor func(int, T) R) []R`
Distributes work across workers and collects results in order.

#### `NewRateLimiter(rate int) *RateLimiter`
Creates a rate limiter with specified operations per second.

#### `(*RateLimiter) Wait(ctx context.Context) error`
Blocks until a token is available.

#### `(*RateLimiter) TryAcquire() bool`
Attempts to acquire a token without blocking.

#### `(*RateLimiter) Stop()`
Stops the rate limiter and releases resources.

#### `NewSemaphore(permits int) *Semaphore`
Creates a semaphore with specified number of permits.

#### `(*Semaphore) Acquire(ctx context.Context) error`
Acquires a permit, blocking if none available.

#### `(*Semaphore) Release()`
Releases a permit.

#### `NewGenerator[T any](ctx context.Context, bufferSize int, producer func(context.Context, chan<- T)) *Generator[T]`
Creates a generator with a producer function.

#### `(*Generator[T]) Next() (T, bool)`
Retrieves the next value from the generator.

#### `(*Generator[T]) Collect() []T`
Collects all remaining values into a slice.

#### `(*Generator[T]) Close()`
Stops the generator.

### Smart Concurrency Wrapper

#### `NewConcurrencyWrapper[T, R any](config *ConcurrencyConfig) *ConcurrencyWrapper[T, R]`
Creates a new concurrency wrapper with the given configuration.

#### `DefaultConcurrencyConfig() *ConcurrencyConfig`
Returns a configuration with sensible defaults.

#### `(*ConcurrencyWrapper[T, R]) Process(ctx context.Context, items []T, processor func(T) (R, error)) ([]R, error)`
Processes items with automatic concurrency control, retry, rate limiting, and circuit breaker.

#### `(*ConcurrencyWrapper[T, R]) Close()`
Cleans up wrapper resources.

#### `NewCircuitBreaker(threshold int) *CircuitBreaker`
Creates a circuit breaker with specified failure threshold.

#### `(*CircuitBreaker) Allow() bool`
Checks if a request should be allowed.

#### `(*CircuitBreaker) RecordSuccess()`
Records a successful operation.

#### `(*CircuitBreaker) RecordFailure()`
Records a failed operation.

## Examples

Comprehensive examples are available in the `example/` directory:

- **[Async Resolve Examples](example/async_resolve/)** - 8 examples demonstrating async task patterns
- **[SuperSlice Examples](example/superslice_demo/)** - 18 examples showing parallel slice processing
- **[Distributed Backend Examples](example/distibuted_backend/)** - SafeChannel with multiple backends
- **[Recasting Examples](example/recasting_demo/)** - Type conversion utilities
- **[Concurrency Patterns Examples](example/concurrency_patterns/)** - 9 examples demonstrating pipeline, fan-out/fan-in, rate limiting, semaphore, generator, and smart wrapper patterns

## Performance Characteristics

### SuperSlice
- **Small slices (< 1000 items)**: Sequential processing to avoid overhead
- **Large slices (≥ 1000 items)**: Parallel processing with worker pools
- **Configurable threshold**: Balance between overhead and parallelism benefit
- **Worker pool**: Defaults to `runtime.NumCPU()`, configurable based on workload

### Async Resolve
- **Parallel execution**: All tasks run concurrently
- **Total time**: Max task duration (not sum of all durations)
- **Minimal overhead**: Uses efficient `sync.WaitGroup` internally

### Concurrency Patterns
- **Pipeline**: Minimal overhead, composable stages, sequential execution per item
- **Fan-Out/Fan-In**: Distributes work efficiently, maintains result order, scales with workers
- **Rate Limiter**: Token bucket algorithm, precise rate control, minimal memory
- **Semaphore**: Fast acquire/release, context-aware, no busy waiting
- **Generator**: Lazy evaluation, memory efficient, context-based cancellation
- **Smart Wrapper**: Combines patterns with minimal overhead, configurable policies

## Use Cases

### ✅ Good Use Cases

**SuperSlice:**
- Processing large datasets (thousands of elements)
- CPU-intensive operations per element
- Independent element processing
- Memory-constrained environments (with in-place updates)

**Async Resolve:**
- Fetching data from multiple sources simultaneously
- Concurrent API requests
- Parallel computational tasks
- Multi-stage pipelines with dependencies

**SafeChannel:**
- Distributed systems with multiple backends
- Timeout-sensitive operations
- Thread-safe channel operations

**Pipeline:**
- ETL (Extract, Transform, Load) operations
- Multi-stage data processing
- Sequential transformations
- Composable data workflows

**Fan-Out/Fan-In:**
- Batch API calls with aggregation
- Parallel data processing with order preservation
- Load distribution across workers
- High-throughput processing

**Rate Limiter:**
- API throttling and quota management
- Preventing service overload
- Controlled resource access
- Batch job rate control

**Semaphore:**
- Database connection pooling
- Limited resource access control
- Concurrency limiting
- Worker pool management

**Generator:**
- Stream processing
- Infinite sequences
- On-demand data production
- Memory-efficient iteration

**Smart Wrapper:**
- Resilient service calls
- Fault-tolerant distributed operations
- Automatic retry with backoff
- Circuit breaker for failing services

### ❌ Not Ideal For

**SuperSlice:**
- Very small slices (< 100 elements) - overhead not worth it
- Operations requiring sequential ordering guarantees
- Highly interdependent element processing

**Async Resolve:**
- Single async operation (use plain goroutine)
- Sequential dependencies between all tasks

## Testing

Run the test suite:

```bash
go test ./...
```

Run benchmarks:

```bash
go test -bench=. -benchmem
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Documentation

For detailed API documentation, visit [pkg.go.dev/github.com/ivikasavnish/goroutine](https://pkg.go.dev/github.com/ivikasavnish/goroutine).

## Author

Copyright (c) 2024 ivikasavnish

## Related Projects

- [Go Concurrency Patterns](https://go.dev/blog/pipelines)
- [errgroup](https://pkg.go.dev/golang.org/x/sync/errgroup) - Error group with context support
- [sync package](https://pkg.go.dev/sync) - Go's standard synchronization primitives
