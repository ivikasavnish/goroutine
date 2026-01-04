package goroutine

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestNewSwissMap(t *testing.T) {
	sm := NewSwissMap[string, int]()
	if sm == nil {
		t.Fatal("NewSwissMap returned nil")
	}
	if len(sm.shards) != 32 {
		t.Errorf("Expected 32 shards, got %d", len(sm.shards))
	}
}

func TestNewSwissMapWithShards(t *testing.T) {
	tests := []struct {
		input    uint32
		expected uint32
	}{
		{0, 32},   // Default to 32
		{1, 1},    // Already power of 2
		{2, 2},    // Already power of 2
		{3, 4},    // Round up to 4
		{5, 8},    // Round up to 8
		{16, 16},  // Already power of 2
		{20, 32},  // Round up to 32
		{64, 64},  // Already power of 2
		{100, 128}, // Round up to 128
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("shards_%d", tt.input), func(t *testing.T) {
			sm := NewSwissMapWithShards[string, int](tt.input)
			if uint32(len(sm.shards)) != tt.expected {
				t.Errorf("Expected %d shards, got %d", tt.expected, len(sm.shards))
			}
		})
	}
}

func TestSwissMapSetAndGet(t *testing.T) {
	sm := NewSwissMap[string, int]()
	
	// Test setting and getting a value
	sm.Set("key1", 42)
	val, ok := sm.Get("key1")
	if !ok {
		t.Error("Expected key1 to exist")
	}
	if val != 42 {
		t.Errorf("Expected value 42, got %d", val)
	}
	
	// Test getting non-existent key
	val, ok = sm.Get("nonexistent")
	if ok {
		t.Error("Expected key to not exist")
	}
	if val != 0 {
		t.Errorf("Expected zero value 0, got %d", val)
	}
}

func TestSwissMapDelete(t *testing.T) {
	sm := NewSwissMap[string, int]()
	
	sm.Set("key1", 42)
	if !sm.Has("key1") {
		t.Error("Expected key1 to exist")
	}
	
	sm.Delete("key1")
	if sm.Has("key1") {
		t.Error("Expected key1 to be deleted")
	}
	
	// Delete non-existent key should not panic
	sm.Delete("nonexistent")
}

func TestSwissMapHas(t *testing.T) {
	sm := NewSwissMap[string, int]()
	
	if sm.Has("key1") {
		t.Error("Expected key1 to not exist")
	}
	
	sm.Set("key1", 42)
	if !sm.Has("key1") {
		t.Error("Expected key1 to exist")
	}
}

func TestSwissMapGetOrSet(t *testing.T) {
	sm := NewSwissMap[string, int]()
	
	// First call should set the value
	val, existed := sm.GetOrSet("key1", 42)
	if existed {
		t.Error("Expected key1 to not exist initially")
	}
	if val != 42 {
		t.Errorf("Expected value 42, got %d", val)
	}
	
	// Second call should return existing value
	val, existed = sm.GetOrSet("key1", 100)
	if !existed {
		t.Error("Expected key1 to exist")
	}
	if val != 42 {
		t.Errorf("Expected value 42, got %d", val)
	}
}

func TestSwissMapGetOrCompute(t *testing.T) {
	sm := NewSwissMap[string, int]()
	
	computeCalls := 0
	compute := func() int {
		computeCalls++
		return 42
	}
	
	// First call should compute the value
	val := sm.GetOrCompute("key1", compute)
	if val != 42 {
		t.Errorf("Expected value 42, got %d", val)
	}
	if computeCalls != 1 {
		t.Errorf("Expected compute to be called once, got %d", computeCalls)
	}
	
	// Second call should return cached value without computing
	val = sm.GetOrCompute("key1", compute)
	if val != 42 {
		t.Errorf("Expected value 42, got %d", val)
	}
	if computeCalls != 1 {
		t.Errorf("Expected compute to be called once, got %d", computeCalls)
	}
}

func TestSwissMapLen(t *testing.T) {
	sm := NewSwissMap[string, int]()
	
	if sm.Len() != 0 {
		t.Errorf("Expected length 0, got %d", sm.Len())
	}
	
	sm.Set("key1", 1)
	sm.Set("key2", 2)
	sm.Set("key3", 3)
	
	if sm.Len() != 3 {
		t.Errorf("Expected length 3, got %d", sm.Len())
	}
	
	sm.Delete("key2")
	if sm.Len() != 2 {
		t.Errorf("Expected length 2, got %d", sm.Len())
	}
}

func TestSwissMapClear(t *testing.T) {
	sm := NewSwissMap[string, int]()
	
	sm.Set("key1", 1)
	sm.Set("key2", 2)
	sm.Set("key3", 3)
	
	if sm.Len() != 3 {
		t.Errorf("Expected length 3, got %d", sm.Len())
	}
	
	sm.Clear()
	
	if sm.Len() != 0 {
		t.Errorf("Expected length 0 after clear, got %d", sm.Len())
	}
	
	if sm.Has("key1") {
		t.Error("Expected key1 to be cleared")
	}
}

func TestSwissMapRange(t *testing.T) {
	sm := NewSwissMap[string, int]()
	
	expected := map[string]int{
		"key1": 1,
		"key2": 2,
		"key3": 3,
	}
	
	for k, v := range expected {
		sm.Set(k, v)
	}
	
	found := make(map[string]int)
	sm.Range(func(key string, value int) bool {
		found[key] = value
		return true
	})
	
	if len(found) != len(expected) {
		t.Errorf("Expected %d entries, got %d", len(expected), len(found))
	}
	
	for k, v := range expected {
		if found[k] != v {
			t.Errorf("Expected %s=%d, got %d", k, v, found[k])
		}
	}
}

func TestSwissMapRangeEarlyStop(t *testing.T) {
	sm := NewSwissMap[string, int]()
	
	for i := 0; i < 100; i++ {
		sm.Set(fmt.Sprintf("key%d", i), i)
	}
	
	count := 0
	sm.Range(func(key string, value int) bool {
		count++
		return count < 5 // Stop after 5 iterations
	})
	
	if count != 5 {
		t.Errorf("Expected to iterate 5 times, got %d", count)
	}
}

func TestSwissMapKeys(t *testing.T) {
	sm := NewSwissMap[string, int]()
	
	sm.Set("key1", 1)
	sm.Set("key2", 2)
	sm.Set("key3", 3)
	
	keys := sm.Keys()
	if len(keys) != 3 {
		t.Errorf("Expected 3 keys, got %d", len(keys))
	}
	
	keySet := make(map[string]bool)
	for _, k := range keys {
		keySet[k] = true
	}
	
	for _, expected := range []string{"key1", "key2", "key3"} {
		if !keySet[expected] {
			t.Errorf("Expected key %s not found", expected)
		}
	}
}

func TestSwissMapValues(t *testing.T) {
	sm := NewSwissMap[string, int]()
	
	sm.Set("key1", 1)
	sm.Set("key2", 2)
	sm.Set("key3", 3)
	
	values := sm.Values()
	if len(values) != 3 {
		t.Errorf("Expected 3 values, got %d", len(values))
	}
	
	valueSet := make(map[int]bool)
	for _, v := range values {
		valueSet[v] = true
	}
	
	for _, expected := range []int{1, 2, 3} {
		if !valueSet[expected] {
			t.Errorf("Expected value %d not found", expected)
		}
	}
}

func TestSwissMapToMap(t *testing.T) {
	sm := NewSwissMap[string, int]()
	
	expected := map[string]int{
		"key1": 1,
		"key2": 2,
		"key3": 3,
	}
	
	for k, v := range expected {
		sm.Set(k, v)
	}
	
	result := sm.ToMap()
	
	if len(result) != len(expected) {
		t.Errorf("Expected %d entries, got %d", len(expected), len(result))
	}
	
	for k, v := range expected {
		if result[k] != v {
			t.Errorf("Expected %s=%d, got %d", k, v, result[k])
		}
	}
}

func TestSwissMapConcurrency(t *testing.T) {
	sm := NewSwissMap[int, int]()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	const numGoroutines = 100
	const numOperations = 1000
	
	var wg sync.WaitGroup
	
	// Writers
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				select {
				case <-ctx.Done():
					return
				default:
					key := id*numOperations + j
					sm.Set(key, key*2)
				}
			}
		}(i)
	}
	
	// Readers
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				select {
				case <-ctx.Done():
					return
				default:
					key := id*numOperations + j
					sm.Get(key)
				}
			}
		}(i)
	}
	
	// Deleters
	for i := 0; i < numGoroutines/10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				select {
				case <-ctx.Done():
					return
				default:
					key := id*numOperations + j
					sm.Delete(key)
				}
			}
		}(i)
	}
	
	wg.Wait()
	
	// Test completed without deadlock or race conditions
	t.Logf("Concurrent test completed successfully. Final map size: %d", sm.Len())
}

func TestSwissMapDifferentTypes(t *testing.T) {
	// Test with int keys
	intMap := NewSwissMap[int, string]()
	intMap.Set(1, "one")
	intMap.Set(2, "two")
	if val, ok := intMap.Get(1); !ok || val != "one" {
		t.Error("Int key map failed")
	}
	
	// Test with int64 keys
	int64Map := NewSwissMap[int64, string]()
	int64Map.Set(1000000000000, "big")
	if val, ok := int64Map.Get(1000000000000); !ok || val != "big" {
		t.Error("Int64 key map failed")
	}
	
	// Test with uint keys
	uintMap := NewSwissMap[uint, string]()
	uintMap.Set(42, "answer")
	if val, ok := uintMap.Get(42); !ok || val != "answer" {
		t.Error("Uint key map failed")
	}
	
	// Test with struct values
	type Person struct {
		Name string
		Age  int
	}
	personMap := NewSwissMap[string, Person]()
	personMap.Set("john", Person{"John Doe", 30})
	if val, ok := personMap.Get("john"); !ok || val.Name != "John Doe" {
		t.Error("Struct value map failed")
	}
}

func BenchmarkSwissMapSet(b *testing.B) {
	sm := NewSwissMap[int, int]()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sm.Set(i, i*2)
	}
}

func BenchmarkSwissMapGet(b *testing.B) {
	sm := NewSwissMap[int, int]()
	for i := 0; i < 10000; i++ {
		sm.Set(i, i*2)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sm.Get(i % 10000)
	}
}

func BenchmarkSwissMapConcurrentReadWrite(b *testing.B) {
	sm := NewSwissMap[int, int]()
	
	// Pre-populate
	for i := 0; i < 1000; i++ {
		sm.Set(i, i*2)
	}
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			if i%2 == 0 {
				sm.Set(i%1000, i)
			} else {
				sm.Get(i % 1000)
			}
			i++
		}
	})
}

func BenchmarkSwissMapGetOrCompute(b *testing.B) {
	sm := NewSwissMap[int, int]()
	compute := func() int { return 42 }
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sm.GetOrCompute(i%1000, compute)
	}
}
