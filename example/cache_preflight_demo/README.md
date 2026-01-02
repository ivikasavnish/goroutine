# Cache & Preflight Examples

This directory contains 5 comprehensive examples demonstrating sync preflight patterns and caching to reduce downstream load on databases and APIs.

## Overview

The preflight pattern checks cache before making expensive calls to databases or APIs. This dramatically reduces:
- Database query load (60-90% reduction typical)
- API request costs
- Response latency
- Infrastructure costs

## Running the Examples

```bash
cd example/cache_preflight_demo
go run main.go
```

## Examples

### Demo 1: CachedGroup with Automatic Preflight

Shows how `CachedGroup` automatically checks cache before executing async tasks.

**Key Points:**
- First call: Cache miss, fetches from DB (~500ms)
- Second call: Cache hit, instant response (<1µs)
- Several orders of magnitude faster with cache preflight (typically 100,000-500,000x in the demo)

**Code Pattern:**
```go
cg := goroutine.NewCachedGroup()
control := &goroutine.CacheControl{
    NoCache: false,
    MaxAge:  5 * time.Minute,
}

cg.AssignWithCache("user:123", &result, func() any {
    return SimulateDBFetch("user:123")
}, control)
cg.Resolve()
```

### Demo 2: PreflightFetcher for Reducing Downstream Load

Demonstrates the `PreflightFetcher` which implements cache-first pattern.

**Key Points:**
- Fetch same key 3 times
- Only 1 DB call made (saved 2 calls)
- 66% reduction in downstream load
- Second and third fetches are instant

**Code Pattern:**
```go
fetcher := goroutine.NewPreflightFetcher(fetchFunc, control)

// First: DB call
val1, _ := fetcher.Fetch(ctx, "user:1")  // ~500ms

// Second: Cache hit
val2, _ := fetcher.Fetch(ctx, "user:1")  // <1µs

// Third: Cache hit
val3, _ := fetcher.Fetch(ctx, "user:1")  // <1µs
```

### Demo 3: NoCache Mode for Fresh Data

Shows how to bypass cache when fresh data is required.

**Key Points:**
- `NoCache: true` forces fetch from source
- Useful for write-after-read scenarios
- Ensures data consistency when needed

**Code Pattern:**
```go
control := &goroutine.CacheControl{
    NoCache: true,  // Bypass cache preflight
    MaxAge:  1 * time.Minute,
}

fetcher := goroutine.NewPreflightFetcher(fetchFunc, control)

// Each fetch hits the source
val1, _ := fetcher.Fetch(ctx, "key1")  // Source call #1
val2, _ := fetcher.Fetch(ctx, "key1")  // Source call #2
val3, _ := fetcher.Fetch(ctx, "key1")  // Source call #3
```

### Demo 4: Stale-While-Revalidate Pattern

Demonstrates returning stale data immediately while revalidating in background.

**Key Points:**
- Best user experience: instant responses
- Cache revalidates asynchronously
- Perfect for non-critical data
- Used by CDNs and modern web frameworks

**Flow:**
```
Initial:  Fetch fresh data (500ms)
          Cache with short TTL
          
Expired:  Return stale immediately (<1µs)
          Revalidate in background (500ms)
          
Next:     Fresh data from cache (<1µs)
```

**Code Pattern:**
```go
control := &goroutine.CacheControl{
    StaleWhileRevalidate: true,
    MaxAge:               100 * time.Millisecond,
}

fetcher := goroutine.NewPreflightFetcher(fetchFunc, control)

// After cache expires
val, _ := fetcher.FetchStaleWhileRevalidate(ctx, "key1")
// Returns stale data instantly, revalidates in background
```

### Demo 5: Reduced Downstream Load Comparison

Side-by-side comparison of load with and without caching.

**Scenario:** Fetching user data 5 times

**Results:**
```
WITHOUT Caching:
  Request 1: DB call #1
  Request 2: DB call #2
  Request 3: DB call #3
  Request 4: DB call #4
  Request 5: DB call #5
  Total: 5 DB calls (2.5 seconds)

WITH Caching:
  Request 1: DB call #1 (cached: false)
  Request 2: DB call #1 (cached: true)
  Request 3: DB call #1 (cached: true)
  Request 4: DB call #1 (cached: true)
  Request 5: DB call #1 (cached: true)
  Total: 1 DB call (0.5 seconds)
  
Load reduction: 80%
Response time: 5x faster average
```

## Cache Control Options

```go
type CacheControl struct {
    // NoCache forces fetch from source, bypassing cache
    NoCache bool
    
    // MaxAge defines how long cached data is valid
    MaxAge time.Duration
    
    // StaleWhileRevalidate allows stale data while fetching fresh data
    StaleWhileRevalidate bool
}
```

## Performance Metrics

From the demo output:

| Pattern | First Call | Subsequent Calls | Load Reduction |
|---------|-----------|------------------|----------------|
| CachedGroup | ~500ms | <1µs | 99.9% |
| PreflightFetcher | ~500ms | <1µs | 66-80% |
| Stale-while-revalidate | ~500ms | <1µs (stale) | Instant UX |

## When to Use

### ✅ Use Cache Preflight When:
- Reading frequently accessed data (user profiles, configs)
- High read-to-write ratio (>10:1)
- Database/API calls are slow (>100ms)
- Same data requested multiple times
- Slight staleness is acceptable

### ❌ Avoid Cache Preflight When:
- Real-time data required (stock prices, live scores)
- Write-heavy workloads
- Data changes constantly
- Memory is severely constrained

## Real-World Applications

1. **User Profile API**: Cache user profiles for 5 minutes
2. **Configuration Service**: Cache configs for 1 hour
3. **Product Catalog**: Cache product details for 15 minutes
4. **Rate Limiting**: Cache rate limit counters for 1 second
5. **Authentication Tokens**: Cache validation results for token lifetime

## Integration Tips

1. **Start Conservative**: Begin with short TTLs (1-5 minutes)
2. **Monitor Hit Rate**: Track cache hits vs misses
3. **Adjust Based on Traffic**: Increase TTL if hit rate is good
4. **Use NoCache for Writes**: Force fresh reads after writes
5. **Implement Cache Warming**: Pre-populate cache for cold starts

## See Also

- [Async Resolve Examples](../async_resolve/) - Parallel async operations
- [SuperSlice Examples](../superslice_demo/) - Parallel slice processing
- Main README - Full API documentation
