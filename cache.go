package goroutine

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// CacheControl represents cache control directives
type CacheControl struct {
	// NoCache forces fetch from source, bypassing cache
	NoCache bool
	// MaxAge defines how long cached data is valid
	MaxAge time.Duration
	// StaleWhileRevalidate allows stale data while fetching fresh data
	StaleWhileRevalidate bool
}

// DefaultCacheControl returns a cache control with sensible defaults
func DefaultCacheControl() *CacheControl {
	return &CacheControl{
		NoCache:              false,
		MaxAge:               5 * time.Minute,
		StaleWhileRevalidate: false,
	}
}

// CacheEntry represents a cached value with metadata
type CacheEntry[T any] struct {
	Value      T
	CachedAt   time.Time
	ExpiresAt  time.Time
	Valid      bool
}

// IsExpired checks if the cache entry has expired
func (ce *CacheEntry[T]) IsExpired() bool {
	return time.Now().After(ce.ExpiresAt)
}

// IsStale checks if the cache entry is stale but still usable
func (ce *CacheEntry[T]) IsStale() bool {
	return ce.IsExpired() && ce.Valid
}

// Cache provides a simple in-memory cache with TTL support
// Now uses SwissMap for improved concurrent performance
type Cache[K comparable, V any] struct {
	data *SwissMap[K, *CacheEntry[V]]
}

// NewCache creates a new cache instance
func NewCache[K comparable, V any]() *Cache[K, V] {
	return &Cache[K, V]{
		data: NewSwissMap[K, *CacheEntry[V]](),
	}
}

// Get retrieves a value from the cache
func (c *Cache[K, V]) Get(key K) (*CacheEntry[V], bool) {
	entry, exists := c.data.Get(key)
	if !exists || !entry.Valid {
		return nil, false
	}
	return entry, true
}

// Set stores a value in the cache with TTL
func (c *Cache[K, V]) Set(key K, value V, ttl time.Duration) {
	now := time.Now()
	c.data.Set(key, &CacheEntry[V]{
		Value:     value,
		CachedAt:  now,
		ExpiresAt: now.Add(ttl),
		Valid:     true,
	})
}

// Delete removes a value from the cache
func (c *Cache[K, V]) Delete(key K) {
	c.data.Delete(key)
}

// Clear removes all entries from the cache
func (c *Cache[K, V]) Clear() {
	c.data.Clear()
}

// Cleanup removes expired entries from the cache
func (c *Cache[K, V]) Cleanup() {
	now := time.Now()
	// Pre-allocate with reasonable capacity to reduce allocations
	keysToDelete := make([]K, 0, 16)
	
	c.data.Range(func(key K, entry *CacheEntry[V]) bool {
		if now.After(entry.ExpiresAt) {
			keysToDelete = append(keysToDelete, key)
		}
		return true
	})
	
	// Delete all expired keys in a batch
	for _, key := range keysToDelete {
		c.data.Delete(key)
	}
}

// CachedGroup wraps Group with caching support for preflight checks
type CachedGroup struct {
	group  *Group
	cache  *Cache[string, any]
	mu     sync.Mutex
}

// NewCachedGroup creates a new CachedGroup with integrated caching
func NewCachedGroup() *CachedGroup {
	return &CachedGroup{
		group: NewGroup(),
		cache: NewCache[string, any](),
	}
}

// AssignWithCache assigns a task with cache-first preflight check
// If the cache contains a valid entry for the key, it uses that value
// Otherwise, it fetches from the provided function and caches the result
func (cg *CachedGroup) AssignWithCache(
	key string,
	result *any,
	fn func() any,
	control *CacheControl,
) {
	if control == nil {
		control = DefaultCacheControl()
	}

	// Preflight: Check cache first unless NoCache is set
	if !control.NoCache {
		if entry, found := cg.cache.Get(key); found {
			if !entry.IsExpired() {
				// Cache hit with valid data
				*result = entry.Value
				return
			} else if control.StaleWhileRevalidate {
				// Return stale data immediately, revalidate in background
				*result = entry.Value
				// Use a separate variable for background revalidation to avoid race condition
				// The background result updates the cache but doesn't need to be used directly
				var bgResult any
				cg.group.Assign(&bgResult, func() any {
					val := fn()
					cg.cache.Set(key, val, control.MaxAge)
					return val
				})
				return
			}
		}
	}

	// Cache miss or NoCache: fetch from source
	cg.group.Assign(result, func() any {
		val := fn()
		cg.cache.Set(key, val, control.MaxAge)
		return val
	})
}

// Resolve waits for all assigned tasks to complete
func (cg *CachedGroup) Resolve() {
	cg.group.Resolve()
}

// ResolveWithTimeout waits for tasks with a timeout
func (cg *CachedGroup) ResolveWithTimeout(timeout time.Duration) bool {
	return cg.group.ResolveWithTimeout(timeout)
}

// GetCache returns the underlying cache for direct access
func (cg *CachedGroup) GetCache() *Cache[string, any] {
	return cg.cache
}

// ClearCache clears all cached entries
func (cg *CachedGroup) ClearCache() {
	cg.cache.Clear()
}

// PreflightFetcher represents a data source with preflight cache check
type PreflightFetcher[T any] struct {
	cache      *Cache[string, T]
	fetchFunc  func(ctx context.Context, key string) (T, error)
	control    *CacheControl
	mu         sync.RWMutex
}

// NewPreflightFetcher creates a new preflight fetcher with caching
func NewPreflightFetcher[T any](
	fetchFunc func(ctx context.Context, key string) (T, error),
	control *CacheControl,
) *PreflightFetcher[T] {
	if control == nil {
		control = DefaultCacheControl()
	}
	return &PreflightFetcher[T]{
		cache:     NewCache[string, T](),
		fetchFunc: fetchFunc,
		control:   control,
	}
}

// Fetch retrieves data with preflight cache check
// This implements the cache-first pattern to reduce downstream load
func (pf *PreflightFetcher[T]) Fetch(ctx context.Context, key string) (T, error) {
	var zero T

	// Preflight: Check cache first unless NoCache is set
	if !pf.control.NoCache {
		if entry, found := pf.cache.Get(key); found {
			if !entry.IsExpired() {
				// Cache hit with valid data - reduces downstream load
				return entry.Value, nil
			}
		}
	}

	// Cache miss or NoCache: fetch from source
	value, err := pf.fetchFunc(ctx, key)
	if err != nil {
		return zero, err
	}

	// Store in cache for future preflight checks
	pf.cache.Set(key, value, pf.control.MaxAge)
	return value, nil
}

// FetchStaleWhileRevalidate returns stale data immediately while revalidating in background
func (pf *PreflightFetcher[T]) FetchStaleWhileRevalidate(ctx context.Context, key string) (T, error) {
	// Check if we have stale data
	if entry, found := pf.cache.Get(key); found {
		if entry.IsStale() && pf.control.StaleWhileRevalidate {
			// Return stale data immediately
			go func() {
				// Revalidate in background with a fresh context to avoid cancellation issues
				bgCtx := context.Background()
				if value, err := pf.fetchFunc(bgCtx, key); err == nil {
					pf.cache.Set(key, value, pf.control.MaxAge)
				}
			}()
			return entry.Value, nil
		}
		if !entry.IsExpired() {
			return entry.Value, nil
		}
	}

	// No stale data available, fetch synchronously
	return pf.Fetch(ctx, key)
}

// SetCacheControl updates the cache control directives
func (pf *PreflightFetcher[T]) SetCacheControl(control *CacheControl) {
	pf.mu.Lock()
	defer pf.mu.Unlock()
	pf.control = control
}

// ClearCache clears all cached entries
func (pf *PreflightFetcher[T]) ClearCache() {
	pf.cache.Clear()
}

// KeyFunc is a function that generates a cache key from parameters
type KeyFunc[P any] func(params P) string

// ParametricFetcher provides cache-first fetching with automatic key generation from parameters
// This simplifies the API by eliminating manual key construction
type ParametricFetcher[P any, T any] struct {
	cache     *Cache[string, T]
	fetchFunc func(ctx context.Context, params P) (T, error)
	keyFunc   KeyFunc[P]
	control   *CacheControl
	mu        sync.RWMutex
}

// NewParametricFetcher creates a fetcher with automatic key generation from parameters
// The keyFunc generates cache keys from parameters automatically
// If keyFunc is nil, uses fmt.Sprintf("%v", params) as default
func NewParametricFetcher[P any, T any](
	fetchFunc func(ctx context.Context, params P) (T, error),
	keyFunc KeyFunc[P],
	control *CacheControl,
) *ParametricFetcher[P, T] {
	if control == nil {
		control = DefaultCacheControl()
	}
	if keyFunc == nil {
		// Default key function uses fmt.Sprintf
		keyFunc = func(params P) string {
			return fmt.Sprintf("%v", params)
		}
	}
	return &ParametricFetcher[P, T]{
		cache:     NewCache[string, T](),
		fetchFunc: fetchFunc,
		keyFunc:   keyFunc,
		control:   control,
	}
}

// Fetch retrieves data with automatic key generation from parameters
// The cache key is automatically generated using the keyFunc
func (pf *ParametricFetcher[P, T]) Fetch(ctx context.Context, params P) (T, error) {
	var zero T
	
	// Generate cache key from parameters
	key := pf.keyFunc(params)

	// Preflight: Check cache first unless NoCache is set
	if !pf.control.NoCache {
		if entry, found := pf.cache.Get(key); found {
			if !entry.IsExpired() {
				// Cache hit with valid data - reduces downstream load
				return entry.Value, nil
			}
		}
	}

	// Cache miss or NoCache: fetch from source
	value, err := pf.fetchFunc(ctx, params)
	if err != nil {
		return zero, err
	}

	// Store in cache for future preflight checks
	pf.cache.Set(key, value, pf.control.MaxAge)
	return value, nil
}

// FetchStaleWhileRevalidate returns stale data immediately while revalidating in background
func (pf *ParametricFetcher[P, T]) FetchStaleWhileRevalidate(ctx context.Context, params P) (T, error) {
	// Generate cache key from parameters
	key := pf.keyFunc(params)
	
	// Check if we have stale data
	if entry, found := pf.cache.Get(key); found {
		if entry.IsStale() && pf.control.StaleWhileRevalidate {
			// Return stale data immediately
			go func() {
				// Revalidate in background with a fresh context to avoid cancellation issues
				bgCtx := context.Background()
				if value, err := pf.fetchFunc(bgCtx, params); err == nil {
					pf.cache.Set(key, value, pf.control.MaxAge)
				}
			}()
			return entry.Value, nil
		}
		if !entry.IsExpired() {
			return entry.Value, nil
		}
	}

	// No stale data available, fetch synchronously
	return pf.Fetch(ctx, params)
}

// SetCacheControl updates the cache control directives
func (pf *ParametricFetcher[P, T]) SetCacheControl(control *CacheControl) {
	pf.mu.Lock()
	defer pf.mu.Unlock()
	pf.control = control
}

// ClearCache clears all cached entries
func (pf *ParametricFetcher[P, T]) ClearCache() {
	pf.cache.Clear()
}

// SimpleKeyFunc creates a KeyFunc from a simple function
// Useful for common key generation patterns
func SimpleKeyFunc[P any](fn func(P) string) KeyFunc[P] {
	return fn
}

// StringKeyFunc creates a KeyFunc that uses the parameter directly as a string key
// Useful when the parameter is already a string identifier
func StringKeyFunc() KeyFunc[string] {
	return func(s string) string { return s }
}

// FormatKeyFunc creates a KeyFunc using a format string
// Example: FormatKeyFunc[int]("user:%d") generates keys like "user:123"
func FormatKeyFunc[P any](format string) KeyFunc[P] {
	return func(params P) string {
		return fmt.Sprintf(format, params)
	}
}
