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

## Examples

Comprehensive examples are available in the `example/` directory:

- **[Async Resolve Examples](example/async_resolve/)** - 8 examples demonstrating async task patterns
- **[SuperSlice Examples](example/superslice_demo/)** - 18 examples showing parallel slice processing
- **[Distributed Backend Examples](example/distibuted_backend/)** - SafeChannel with multiple backends
- **[Recasting Examples](example/recasting_demo/)** - Type conversion utilities

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
