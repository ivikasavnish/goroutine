# SuperSlice Examples

This directory contains comprehensive examples demonstrating the usage of the enhanced SuperSlice functionality in the goroutine package.

## Overview

SuperSlice provides efficient slice processing with automatic parallelization based on configurable thresholds. It intelligently switches between sequential and parallel processing to optimize performance.

## Key Features

1. **Automatic Parallelization**: Automatically uses goroutines and worker pools for large slices
2. **Configurable Threshold**: Set custom thresholds for when to switch to parallel processing (default: 1000)
3. **Worker Pool Management**: Configure the number of worker goroutines (default: NumCPU)
4. **In-Place Updates**: Option to modify slices in place to save memory
5. **Iterable Mode**: Process slices using an iterable pattern with job distribution
6. **Error Handling**: Built-in support for error handling in processing callbacks
7. **Type Transformations**: Map elements to different types with parallel processing
8. **Filter Operations**: Efficiently filter large slices in parallel

## Configuration Options

```go
type SuperSliceConfig struct {
    // Threshold determines when to switch to parallel processing (default: 1000)
    Threshold int
    
    // NumWorkers specifies the number of worker goroutines (default: NumCPU)
    NumWorkers int
    
    // UseIterable indicates whether to convert slice to iterable first (default: false)
    UseIterable bool
    
    // InPlace indicates whether to update the slice in place (default: false)
    InPlace bool
}
```

## Examples Included

### Basic Examples

1. **Basic Small Slice**: Sequential processing for small slices
2. **Large Slice Parallel**: Automatic parallel processing for large slices
3. **In-Place Updates**: Memory-efficient in-place modifications
4. **Custom Threshold**: Configure when to enable parallel processing
5. **Custom Workers**: Set the number of parallel workers
6. **Iterable Mode**: Use job-based distribution pattern
7. **Error Handling**: Process with error handling
8. **ForEach**: Apply operations without collecting results
9. **MapTo**: Transform to different types
10. **Filter**: Filter elements in parallel
11. **Performance Comparison**: Compare sequential vs parallel

### Advanced Examples

1. **Complex Data Processing**: Process struct-based records
2. **String Processing**: Efficient string transformations
3. **Chained Operations**: Combine multiple operations
4. **Data Normalization**: Statistical data processing
5. **Batch Processing**: Custom configuration for batch jobs
6. **Conditional Processing**: Apply conditional logic
7. **Type Conversion**: Advanced type transformations

## Running the Examples

```bash
cd example/superslice_demo
go run main.go
```

## Usage Patterns

### Simple Usage
```go
numbers := []int{1, 2, 3, 4, 5}
ss := goroutine.NewSuperSlice(numbers)
result := ss.Process(func(index int, item int) int {
    return item * 2
})
```

### With Configuration
```go
config := &goroutine.SuperSliceConfig{
    Threshold:   500,
    NumWorkers:  8,
    UseIterable: true,
    InPlace:     false,
}
ss := goroutine.NewSuperSliceWithConfig(data, config)
result := ss.Process(callback)
```

### Fluent API
```go
result := goroutine.NewSuperSlice(data).
    WithThreshold(500).
    WithWorkers(8).
    WithIterable().
    Process(callback)
```

### In-Place Updates
```go
goroutine.NewSuperSlice(data).
    WithInPlace().
    Process(func(index int, item int) int {
        return item * 10
    })
// data is now modified in place
```

### Type Transformation
```go
numbers := []int{1, 2, 3, 4, 5}
ss := goroutine.NewSuperSlice(numbers)
strings := goroutine.MapTo(ss, func(index int, item int) string {
    return fmt.Sprintf("Number_%d", item)
})
```

### Error Handling
```go
result, err := ss.ProcessWithError(func(index int, item int) (int, error) {
    if item < 0 {
        return 0, fmt.Errorf("invalid item: %d", item)
    }
    return item * 2, nil
})
```

## Performance Considerations

- **Small Slices (< threshold)**: Uses sequential processing to avoid goroutine overhead
- **Large Slices (>= threshold)**: Uses parallel processing with worker pools
- **Threshold Selection**: Balance between overhead and parallelism benefit
  - Lower threshold (e.g., 100): More aggressive parallelization
  - Higher threshold (e.g., 5000): Conservative, only for very large slices
- **Worker Count**: More workers can improve throughput but increase overhead
  - Default (NumCPU): Good balance for CPU-bound tasks
  - Custom: Tune based on workload (I/O vs CPU bound)

## When to Use SuperSlice

✅ **Good Use Cases:**
- Processing large datasets (thousands of elements)
- CPU-intensive operations per element
- Independent element processing
- Need for both sequential and parallel modes
- Memory-constrained environments (with in-place updates)

❌ **Not Ideal For:**
- Very small slices (< 100 elements) - overhead not worth it
- Operations requiring sequential ordering guarantees
- Highly interdependent element processing
- When simple for-loops are sufficient

## Integration with AsyncResolve

SuperSlice can work alongside the AsyncResolve functionality:

```go
group := goroutine.NewGroup()
ss := goroutine.NewSuperSlice(data)

// Process slice in parallel
result := ss.Process(callback)

// Use with async resolve for additional parallel operations
var finalResult any
group.Assign(&finalResult, func() any {
    return processResults(result)
})
group.Resolve()
```
