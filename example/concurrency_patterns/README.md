# Concurrency Patterns Examples

This directory contains comprehensive examples demonstrating advanced Go concurrency patterns and the smart wrapper functionality.

## Examples Included

### 1. Pipeline Pattern
Demonstrates how to create processing pipelines with multiple stages that transform data sequentially.

```go
pipeline := goroutine.NewPipeline[int]().
    AddStage(func(n int) int { return n * 2 }).
    AddStage(func(n int) int { return n + 10 }).
    AddStage(func(n int) int { return n * n })
```

### 2. Fan-Out/Fan-In Pattern
Shows how to distribute work across multiple workers and collect results efficiently.

```go
fanOut := goroutine.NewFanOut[int, int](3)
results := fanOut.ProcessWithIndex(ctx, items, processor)
```

### 3. Rate Limiter Pattern
Controls the rate of operations to prevent overwhelming external services.

```go
limiter := goroutine.NewRateLimiter(5) // 5 operations per second
limiter.Wait(ctx)
```

### 4. Semaphore Pattern
Limits concurrent access to resources with configurable permits.

```go
sem := goroutine.NewSemaphore(3) // Max 3 concurrent operations
sem.Acquire(ctx)
defer sem.Release()
```

### 5. Generator Pattern
Produces values on demand using channels and context.

```go
gen := goroutine.NewGenerator[int](ctx, bufferSize, producerFunc)
values := gen.Collect()
```

### 6-8. Smart Wrapper Examples
Demonstrates the unified concurrency wrapper with:
- Basic concurrency control
- Automatic retry logic
- Circuit breaker pattern

```go
config := &goroutine.ConcurrencyConfig{
    MaxConcurrency: 3,
    Timeout: 5 * time.Second,
    RetryAttempts: 3,
    EnableCircuitBreaker: true,
}
wrapper := goroutine.NewConcurrencyWrapper[int, int](config)
results, err := wrapper.Process(ctx, items, processor)
```

### 9. Combined Patterns
Shows how to compose multiple patterns together for complex workflows.

## Running the Examples

```bash
cd example/concurrency_patterns
go run main.go
```

## Key Features Demonstrated

- **Pipeline**: Composable data transformation stages
- **Fan-Out/Fan-In**: Parallel processing with result aggregation
- **Rate Limiting**: Controlled operation execution rate
- **Semaphore**: Resource access control
- **Generator**: Lazy value production
- **Smart Wrapper**: Unified API with retry, timeout, rate limiting, and circuit breaker
- **Pattern Composition**: Combining multiple patterns for complex scenarios

## Use Cases

- **Pipeline**: Data ETL, multi-stage processing
- **Fan-Out/Fan-In**: Parallel API calls, batch processing
- **Rate Limiter**: API throttling, resource protection
- **Semaphore**: Database connection pooling, resource limiting
- **Generator**: Stream processing, infinite sequences
- **Smart Wrapper**: Resilient service calls with fault tolerance
