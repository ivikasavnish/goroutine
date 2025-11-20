# Quick Start: Async Resolve

## What is Async Resolve?

A way to run multiple tasks in parallel and wait for all to complete - like `Promise.all()` in JavaScript.

```go
group := goroutine.NewGroup()

var result1, result2 any

// Launch task 1
group.Assign(&result1, func() any {
    return doWork1()
})

// Launch task 2
group.Assign(&result2, func() any {
    return doWork2()
})

// Wait for both to finish
group.Resolve()

// Use results
fmt.Println(result1, result2)
```

## Build & Run

```bash
go build -o async-resolve-demo
./async-resolve-demo
```

## 8 Practical Examples

| Demo | Use Case | Key Feature |
|------|----------|------------|
| 1 | Basic async | Simple parallel tasks |
| 2 | Data fetching | Multiple sources in parallel |
| 3 | API calls | Concurrent HTTP requests |
| 4 | Computation | CPU-intensive parallel work |
| 5 | Timeouts | Time-bounded operations |
| 6 | Map-reduce | Distributed processing |
| 7 | Pipelines | Multi-stage workflows |
| 8 | Errors | Error handling pattern |

## Common Patterns

### Simple Parallel Tasks

```go
group := goroutine.NewGroup()
var res1, res2, res3 any

group.Assign(&res1, func() any { return doTask1() })
group.Assign(&res2, func() any { return doTask2() })
group.Assign(&res3, func() any { return doTask3() })

group.Resolve()  // Wait for all
```

### Parallel Slice Processing

```go
results := make([]any, len(items))

for i, item := range items {
    idx := i
    itm := item
    group.Assign(&results[idx], func() any {
        return process(itm)
    })
}

group.Resolve()
```

### With Timeout

```go
if group.ResolveWithTimeout(5 * time.Second) {
    // All done in time
} else {
    // Timeout - use fallback
}
```

### Error Handling

```go
type Result struct {
    Value any
    Error error
}

group.Assign(&result, func() any {
    val, err := riskyOp()
    return Result{Value: val, Error: err}
})

group.Resolve()

// Check error later
if res, ok := result.(Result); ok && res.Error != nil {
    log.Printf("Error: %v", res.Error)
}
```

## Speed Comparison

3 tasks, 100ms each:

```
Sequential:  300ms  (100 + 100 + 100)
Parallel:    100ms  (max task duration)
Speedup:     3x
```

## Key Points

✓ All tasks run in parallel (goroutines)
✓ Wait for completion: `Resolve()` or `ResolveWithTimeout()`
✓ Results collected in any type
✓ Type-safe with type assertion: `val.(int)`
✓ Goroutines continue after timeout

## Typical Output

```
Demo 2: Parallel Data Fetching
Fetching data from 5 different sources in parallel

Starting fetches...
  Fetching from Database...
  Fetching from Cache...
  ✓ Cache completed
  ✓ FileSystem completed
  ✓ API-1 completed
  ✓ API-2 completed
  ✓ Database completed

Total time: 814ms (all ran in parallel, max is ~800ms)
```

## When to Use

✓ Multiple independent tasks
✓ Concurrent API calls
✓ Parallel data processing
✓ Performance-critical code

## When NOT to Use

✗ Single async task (use goroutine directly)
✗ Streaming data (use channels)
✗ Dependent sequential tasks (use sync or callbacks)

## API Quick Reference

```go
group := goroutine.NewGroup()

// Launch task
group.Assign(&result, func() any {
    return value
})

// Wait forever
group.Resolve()

// Wait with timeout (returns true/false)
success := group.ResolveWithTimeout(5 * time.Second)
```

## Real-World Example

```go
type UserData struct {
    User     any
    Posts    any
    Comments any
}

group := goroutine.NewGroup()
var user, posts, comments any

// Fetch from 3 different APIs in parallel
group.Assign(&user, func() any {
    return fetchUser(userID)
})

group.Assign(&posts, func() any {
    return fetchPosts(userID)
})

group.Assign(&comments, func() any {
    return fetchComments(userID)
})

// Single wait for all to complete
group.Resolve()

// All data now available
userData := UserData{
    User:     user,
    Posts:    posts,
    Comments: comments,
}
```

## Troubleshooting

**Q: Results are nil?**
A: Pass pointer to result: `group.Assign(&result, fn)` ✓ not `group.Assign(result, fn)` ✗

**Q: Type assertion panic?**
A: Use safe assertion: `if val, ok := result.(int); ok { ... }`

**Q: Tasks not done after ResolveWithTimeout returns?**
A: That's normal! Tasks continue running. Use fallback value instead.

## Go Deeper

See `README.md` for:
- 8 detailed examples with explanations
- API reference
- Design patterns
- Advanced usage
- Performance analysis
