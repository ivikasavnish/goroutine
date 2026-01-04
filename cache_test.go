package goroutine

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestNewCache(t *testing.T) {
	cache := NewCache[string, int]()
	if cache == nil {
		t.Fatal("NewCache returned nil")
	}
}

func TestCacheGetSet(t *testing.T) {
	cache := NewCache[string, string]()
	
	// Test Set and Get
	cache.Set("key1", "value1", 1*time.Minute)
	
	entry, found := cache.Get("key1")
	if !found {
		t.Fatal("Expected to find key1 in cache")
	}
	if entry.Value != "value1" {
		t.Errorf("Expected value1, got %s", entry.Value)
	}
	if entry.IsExpired() {
		t.Error("Entry should not be expired")
	}
}

func TestCacheExpiration(t *testing.T) {
	cache := NewCache[string, string]()
	
	// Set with very short TTL
	cache.Set("key1", "value1", 50*time.Millisecond)
	
	// Should be valid immediately
	entry, found := cache.Get("key1")
	if !found || entry.IsExpired() {
		t.Error("Entry should be valid immediately after set")
	}
	
	// Wait for expiration
	time.Sleep(100 * time.Millisecond)
	
	// Entry should still exist but be expired
	entry, found = cache.Get("key1")
	if !found {
		t.Error("Entry should still exist in cache")
	}
	if !entry.IsExpired() {
		t.Error("Entry should be expired")
	}
}

func TestCacheDelete(t *testing.T) {
	cache := NewCache[string, string]()
	
	cache.Set("key1", "value1", 1*time.Minute)
	cache.Delete("key1")
	
	_, found := cache.Get("key1")
	if found {
		t.Error("Key should be deleted from cache")
	}
}

func TestCacheClear(t *testing.T) {
	cache := NewCache[string, string]()
	
	cache.Set("key1", "value1", 1*time.Minute)
	cache.Set("key2", "value2", 1*time.Minute)
	cache.Clear()
	
	_, found1 := cache.Get("key1")
	_, found2 := cache.Get("key2")
	
	if found1 || found2 {
		t.Error("Cache should be empty after Clear")
	}
}

func TestCacheCleanup(t *testing.T) {
	cache := NewCache[string, string]()
	
	// Add entries with different TTLs
	cache.Set("short", "value1", 50*time.Millisecond)
	cache.Set("long", "value2", 1*time.Minute)
	
	// Wait for short entry to expire
	time.Sleep(100 * time.Millisecond)
	
	// Run cleanup
	cache.Cleanup()
	
	// Short entry should be removed
	_, foundShort := cache.Get("short")
	if foundShort {
		t.Error("Expired entry should be removed by Cleanup")
	}
	
	// Long entry should still exist
	_, foundLong := cache.Get("long")
	if !foundLong {
		t.Error("Valid entry should not be removed by Cleanup")
	}
}

func TestDefaultCacheControl(t *testing.T) {
	control := DefaultCacheControl()
	
	if control.NoCache {
		t.Error("NoCache should be false by default")
	}
	if control.MaxAge != 5*time.Minute {
		t.Errorf("Expected MaxAge 5m, got %v", control.MaxAge)
	}
	if control.StaleWhileRevalidate {
		t.Error("StaleWhileRevalidate should be false by default")
	}
}

func TestNewCachedGroup(t *testing.T) {
	cg := NewCachedGroup()
	if cg == nil {
		t.Fatal("NewCachedGroup returned nil")
	}
	if cg.group == nil {
		t.Fatal("CachedGroup.group is nil")
	}
	if cg.cache == nil {
		t.Fatal("CachedGroup.cache is nil")
	}
}

func TestCachedGroupWithCache(t *testing.T) {
	cg := NewCachedGroup()
	
	callCount := 0
	var result any
	
	// First call should execute function
	control := &CacheControl{
		NoCache: false,
		MaxAge:  1 * time.Minute,
	}
	
	cg.AssignWithCache("key1", &result, func() any {
		callCount++
		return "value1"
	}, control)
	
	cg.Resolve()
	
	if callCount != 1 {
		t.Errorf("Expected 1 call, got %d", callCount)
	}
	if result != "value1" {
		t.Errorf("Expected value1, got %v", result)
	}
	
	// Second call should use cache
	var result2 any
	cg2 := NewCachedGroup()
	cg2.cache = cg.cache // Share cache
	
	cg2.AssignWithCache("key1", &result2, func() any {
		callCount++
		return "value2"
	}, control)
	
	cg2.Resolve()
	
	// Call count should still be 1 (cache hit)
	if callCount != 1 {
		t.Errorf("Expected 1 call (cache hit), got %d", callCount)
	}
	if result2 != "value1" {
		t.Errorf("Expected cached value1, got %v", result2)
	}
}

func TestCachedGroupNoCache(t *testing.T) {
	cg := NewCachedGroup()
	
	callCount := 0
	var result any
	
	control := &CacheControl{
		NoCache: true,
		MaxAge:  1 * time.Minute,
	}
	
	// First call
	cg.AssignWithCache("key1", &result, func() any {
		callCount++
		return "value1"
	}, control)
	cg.Resolve()
	
	// Second call with NoCache should execute function again
	cg2 := NewCachedGroup()
	cg2.cache = cg.cache // Share cache
	
	var result2 any
	cg2.AssignWithCache("key1", &result2, func() any {
		callCount++
		return "value2"
	}, control)
	cg2.Resolve()
	
	if callCount != 2 {
		t.Errorf("Expected 2 calls with NoCache, got %d", callCount)
	}
}

func TestPreflightFetcher(t *testing.T) {
	callCount := 0
	fetchFunc := func(ctx context.Context, key string) (string, error) {
		callCount++
		return "value-" + key, nil
	}
	
	control := &CacheControl{
		NoCache: false,
		MaxAge:  1 * time.Minute,
	}
	
	fetcher := NewPreflightFetcher(fetchFunc, control)
	ctx := context.Background()
	
	// First fetch should call fetchFunc
	val1, err := fetcher.Fetch(ctx, "key1")
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}
	if val1 != "value-key1" {
		t.Errorf("Expected value-key1, got %s", val1)
	}
	if callCount != 1 {
		t.Errorf("Expected 1 call, got %d", callCount)
	}
	
	// Second fetch should use cache (preflight check)
	val2, err := fetcher.Fetch(ctx, "key1")
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}
	if val2 != "value-key1" {
		t.Errorf("Expected cached value-key1, got %s", val2)
	}
	if callCount != 1 {
		t.Errorf("Expected 1 call (cache hit), got %d", callCount)
	}
}

func TestPreflightFetcherNoCache(t *testing.T) {
	callCount := 0
	fetchFunc := func(ctx context.Context, key string) (string, error) {
		callCount++
		return "value", nil
	}
	
	control := &CacheControl{
		NoCache: true,
		MaxAge:  1 * time.Minute,
	}
	
	fetcher := NewPreflightFetcher(fetchFunc, control)
	ctx := context.Background()
	
	// Both fetches should call fetchFunc with NoCache
	_, _ = fetcher.Fetch(ctx, "key1")
	_, _ = fetcher.Fetch(ctx, "key1")
	
	if callCount != 2 {
		t.Errorf("Expected 2 calls with NoCache, got %d", callCount)
	}
}

func TestPreflightFetcherError(t *testing.T) {
	expectedErr := errors.New("fetch error")
	fetchFunc := func(ctx context.Context, key string) (string, error) {
		return "", expectedErr
	}
	
	control := DefaultCacheControl()
	fetcher := NewPreflightFetcher(fetchFunc, control)
	ctx := context.Background()
	
	_, err := fetcher.Fetch(ctx, "key1")
	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
}

func TestPreflightFetcherStaleWhileRevalidate(t *testing.T) {
	callCount := 0
	fetchFunc := func(ctx context.Context, key string) (string, error) {
		callCount++
		time.Sleep(50 * time.Millisecond) // Simulate slow fetch
		return "fresh-value", nil
	}
	
	control := &CacheControl{
		NoCache:              false,
		MaxAge:               50 * time.Millisecond, // Short TTL
		StaleWhileRevalidate: true,
	}
	
	fetcher := NewPreflightFetcher(fetchFunc, control)
	ctx := context.Background()
	
	// Initial fetch
	val1, err := fetcher.Fetch(ctx, "key1")
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}
	if val1 != "fresh-value" {
		t.Errorf("Expected fresh-value, got %s", val1)
	}
	
	// Modify cache to have stale data
	fetcher.cache.Set("key1", "stale-value", -1*time.Second) // Expired
	
	// Fetch with stale-while-revalidate should return stale immediately
	val2, err := fetcher.FetchStaleWhileRevalidate(ctx, "key1")
	if err != nil {
		t.Fatalf("FetchStaleWhileRevalidate failed: %v", err)
	}
	if val2 != "stale-value" {
		t.Errorf("Expected stale-value immediately, got %s", val2)
	}
	
	// Wait for background revalidation
	time.Sleep(150 * time.Millisecond)
	
	// Should now have fresh value in cache
	entry, found := fetcher.cache.Get("key1")
	if !found {
		t.Fatal("Expected entry in cache after revalidation")
	}
	if entry.Value != "fresh-value" {
		t.Errorf("Expected fresh-value after revalidation, got %s", entry.Value)
	}
}

func TestParametricFetcher(t *testing.T) {
	callCount := 0
	fetchFunc := func(ctx context.Context, userID int) (string, error) {
		callCount++
		return fmt.Sprintf("user-data-%d", userID), nil
	}
	
	// Use FormatKeyFunc for automatic key generation
	keyFunc := FormatKeyFunc[int]("user:%d")
	control := &CacheControl{
		NoCache: false,
		MaxAge:  1 * time.Minute,
	}
	
	fetcher := NewParametricFetcher(fetchFunc, keyFunc, control)
	ctx := context.Background()
	
	// First fetch should call fetchFunc
	val1, err := fetcher.Fetch(ctx, 123)
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}
	if val1 != "user-data-123" {
		t.Errorf("Expected user-data-123, got %s", val1)
	}
	if callCount != 1 {
		t.Errorf("Expected 1 call, got %d", callCount)
	}
	
	// Second fetch should use cache (no manual key construction needed!)
	val2, err := fetcher.Fetch(ctx, 123)
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}
	if val2 != "user-data-123" {
		t.Errorf("Expected cached user-data-123, got %s", val2)
	}
	if callCount != 1 {
		t.Errorf("Expected 1 call (cache hit), got %d", callCount)
	}
	
	// Different parameter should call fetchFunc again
	val3, err := fetcher.Fetch(ctx, 456)
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}
	if val3 != "user-data-456" {
		t.Errorf("Expected user-data-456, got %s", val3)
	}
	if callCount != 2 {
		t.Errorf("Expected 2 calls, got %d", callCount)
	}
}

func TestParametricFetcherDefaultKeyFunc(t *testing.T) {
	fetchFunc := func(ctx context.Context, key string) (string, error) {
		return "data-" + key, nil
	}
	
	// Use nil keyFunc for default behavior
	fetcher := NewParametricFetcher(fetchFunc, nil, nil)
	ctx := context.Background()
	
	val, err := fetcher.Fetch(ctx, "test")
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}
	if val != "data-test" {
		t.Errorf("Expected data-test, got %s", val)
	}
}

func TestParametricFetcherStaleWhileRevalidate(t *testing.T) {
	fetchFunc := func(ctx context.Context, id int) (string, error) {
		time.Sleep(50 * time.Millisecond)
		return fmt.Sprintf("fresh-%d", id), nil
	}
	
	keyFunc := FormatKeyFunc[int]("item:%d")
	control := &CacheControl{
		NoCache:              false,
		MaxAge:               50 * time.Millisecond,
		StaleWhileRevalidate: true,
	}
	
	fetcher := NewParametricFetcher(fetchFunc, keyFunc, control)
	ctx := context.Background()
	
	// Initial fetch
	val1, _ := fetcher.Fetch(ctx, 1)
	if val1 != "fresh-1" {
		t.Errorf("Expected fresh-1, got %s", val1)
	}
	
	// Wait for cache to expire
	time.Sleep(100 * time.Millisecond)
	
	// Manually set stale data in cache
	fetcher.cache.Set("item:1", "stale-1", -1*time.Second)
	
	// Fetch with stale-while-revalidate - should return stale immediately
	start := time.Now()
	val2, err := fetcher.FetchStaleWhileRevalidate(ctx, 1)
	elapsed := time.Since(start)
	
	if err != nil {
		t.Fatalf("FetchStaleWhileRevalidate failed: %v", err)
	}
	if val2 != "stale-1" {
		t.Errorf("Expected stale-1 immediately, got %s", val2)
	}
	// Should be instant (much less than the 50ms fetch time)
	if elapsed > 20*time.Millisecond {
		t.Errorf("Expected instant response, took %v", elapsed)
	}
}

type UserRequest struct {
	UserID int
	Locale string
}

func TestParametricFetcherWithStruct(t *testing.T) {
	callCount := 0
	fetchFunc := func(ctx context.Context, req UserRequest) (string, error) {
		callCount++
		return fmt.Sprintf("user-%d-%s", req.UserID, req.Locale), nil
	}
	
	// Custom key function for struct
	keyFunc := func(req UserRequest) string {
		return fmt.Sprintf("user:%d:%s", req.UserID, req.Locale)
	}
	
	fetcher := NewParametricFetcher(fetchFunc, keyFunc, nil)
	ctx := context.Background()
	
	req1 := UserRequest{UserID: 123, Locale: "en"}
	req2 := UserRequest{UserID: 123, Locale: "fr"}
	
	// First request
	val1, err := fetcher.Fetch(ctx, req1)
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}
	if val1 != "user-123-en" {
		t.Errorf("Expected user-123-en, got %s", val1)
	}
	
	// Same request should hit cache
	val2, err := fetcher.Fetch(ctx, req1)
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}
	if val2 != "user-123-en" {
		t.Errorf("Expected cached user-123-en, got %s", val2)
	}
	if callCount != 1 {
		t.Errorf("Expected 1 call (cache hit), got %d", callCount)
	}
	
	// Different locale should miss cache
	val3, err := fetcher.Fetch(ctx, req2)
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}
	if val3 != "user-123-fr" {
		t.Errorf("Expected user-123-fr, got %s", val3)
	}
	if callCount != 2 {
		t.Errorf("Expected 2 calls, got %d", callCount)
	}
}
