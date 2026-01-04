package goroutine

import (
	"fmt"
	"hash/fnv"
	"math"
	"strconv"
	"sync"
)

// SwissMap is a high-performance, thread-safe generic map using sharded architecture
// for optimal concurrent access. It divides the map into multiple shards, each protected
// by its own mutex, reducing lock contention in high-concurrency scenarios.
type SwissMap[K comparable, V any] struct {
	shards    []*swissMapShard[K, V]
	shardMask uint32
}

// swissMapShard represents a single shard of the SwissMap
type swissMapShard[K comparable, V any] struct {
	data map[K]V
	mu   sync.RWMutex
}

// NewSwissMap creates a new SwissMap with default shard count (32)
// This provides excellent performance for most concurrent workloads
func NewSwissMap[K comparable, V any]() *SwissMap[K, V] {
	return NewSwissMapWithShards[K, V](32)
}

// NewSwissMapWithShards creates a new SwissMap with specified shard count
// The shard count must be a power of 2 for optimal performance
// Higher shard counts reduce lock contention but increase memory overhead
func NewSwissMapWithShards[K comparable, V any](shardCount uint32) *SwissMap[K, V] {
	// Ensure shard count is power of 2
	if shardCount == 0 {
		shardCount = 32
	}
	if shardCount&(shardCount-1) != 0 {
		// Round up to next power of 2
		shardCount = nextPowerOf2(shardCount)
	}

	shards := make([]*swissMapShard[K, V], shardCount)
	for i := range shards {
		shards[i] = &swissMapShard[K, V]{
			data: make(map[K]V),
		}
	}

	return &SwissMap[K, V]{
		shards:    shards,
		shardMask: shardCount - 1,
	}
}

// getShard returns the shard for a given key using hash-based distribution
func (sm *SwissMap[K, V]) getShard(key K) *swissMapShard[K, V] {
	h := fnv.New32a()
	// Hash the key using its string representation
	h.Write([]byte(keyToBytes(key)))
	hash := h.Sum32()
	return sm.shards[hash&sm.shardMask]
}

// Set stores a key-value pair in the map
func (sm *SwissMap[K, V]) Set(key K, value V) {
	shard := sm.getShard(key)
	shard.mu.Lock()
	defer shard.mu.Unlock()
	shard.data[key] = value
}

// Get retrieves a value from the map
// Returns the value and true if found, zero value and false otherwise
func (sm *SwissMap[K, V]) Get(key K) (V, bool) {
	shard := sm.getShard(key)
	shard.mu.RLock()
	defer shard.mu.RUnlock()
	val, ok := shard.data[key]
	return val, ok
}

// Delete removes a key from the map
func (sm *SwissMap[K, V]) Delete(key K) {
	shard := sm.getShard(key)
	shard.mu.Lock()
	defer shard.mu.Unlock()
	delete(shard.data, key)
}

// Has checks if a key exists in the map
func (sm *SwissMap[K, V]) Has(key K) bool {
	shard := sm.getShard(key)
	shard.mu.RLock()
	defer shard.mu.RUnlock()
	_, ok := shard.data[key]
	return ok
}

// GetOrSet retrieves a value or sets it if not present
// Returns the value (existing or newly set) and true if it was already present
func (sm *SwissMap[K, V]) GetOrSet(key K, value V) (V, bool) {
	shard := sm.getShard(key)
	shard.mu.Lock()
	defer shard.mu.Unlock()
	
	if existing, ok := shard.data[key]; ok {
		return existing, true
	}
	shard.data[key] = value
	return value, false
}

// GetOrCompute retrieves a value or computes and sets it if not present
// The compute function is only called if the key doesn't exist
func (sm *SwissMap[K, V]) GetOrCompute(key K, compute func() V) V {
	shard := sm.getShard(key)
	shard.mu.Lock()
	defer shard.mu.Unlock()
	
	if existing, ok := shard.data[key]; ok {
		return existing
	}
	value := compute()
	shard.data[key] = value
	return value
}

// Len returns the total number of entries in the map
func (sm *SwissMap[K, V]) Len() int {
	count := 0
	for _, shard := range sm.shards {
		shard.mu.RLock()
		count += len(shard.data)
		shard.mu.RUnlock()
	}
	return count
}

// Clear removes all entries from the map
func (sm *SwissMap[K, V]) Clear() {
	for _, shard := range sm.shards {
		shard.mu.Lock()
		shard.data = make(map[K]V)
		shard.mu.Unlock()
	}
}

// Range iterates over all key-value pairs in the map
// The function f is called for each entry. If f returns false, iteration stops
func (sm *SwissMap[K, V]) Range(f func(key K, value V) bool) {
	for _, shard := range sm.shards {
		shard.mu.RLock()
		for k, v := range shard.data {
			if !f(k, v) {
				shard.mu.RUnlock()
				return
			}
		}
		shard.mu.RUnlock()
	}
}

// Keys returns all keys in the map
func (sm *SwissMap[K, V]) Keys() []K {
	keys := make([]K, 0, sm.Len())
	sm.Range(func(key K, value V) bool {
		keys = append(keys, key)
		return true
	})
	return keys
}

// Values returns all values in the map
func (sm *SwissMap[K, V]) Values() []V {
	values := make([]V, 0, sm.Len())
	sm.Range(func(key K, value V) bool {
		values = append(values, value)
		return true
	})
	return values
}

// ToMap converts the SwissMap to a regular Go map
// This creates a snapshot of the current state
func (sm *SwissMap[K, V]) ToMap() map[K]V {
	result := make(map[K]V, sm.Len())
	sm.Range(func(key K, value V) bool {
		result[key] = value
		return true
	})
	return result
}

// nextPowerOf2 returns the next power of 2 greater than or equal to n
func nextPowerOf2(n uint32) uint32 {
	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n++
	return n
}

// keyToBytes converts a comparable key to bytes for hashing
// Uses fmt.Sprint for consistent conversion of all comparable types
func keyToBytes(key any) []byte {
	// For common types, use optimized byte conversion
	switch v := key.(type) {
	case string:
		return []byte(v)
	case int:
		// Handle both 32-bit and 64-bit architectures
		if strconv.IntSize == 64 {
			return []byte{byte(v), byte(v >> 8), byte(v >> 16), byte(v >> 24),
				byte(v >> 32), byte(v >> 40), byte(v >> 48), byte(v >> 56)}
		}
		return []byte{byte(v), byte(v >> 8), byte(v >> 16), byte(v >> 24)}
	case int32:
		return []byte{byte(v), byte(v >> 8), byte(v >> 16), byte(v >> 24)}
	case int64:
		return []byte{byte(v), byte(v >> 8), byte(v >> 16), byte(v >> 24),
			byte(v >> 32), byte(v >> 40), byte(v >> 48), byte(v >> 56)}
	case uint:
		// Handle both 32-bit and 64-bit architectures
		if strconv.IntSize == 64 {
			return []byte{byte(v), byte(v >> 8), byte(v >> 16), byte(v >> 24),
				byte(v >> 32), byte(v >> 40), byte(v >> 48), byte(v >> 56)}
		}
		return []byte{byte(v), byte(v >> 8), byte(v >> 16), byte(v >> 24)}
	case uint32:
		return []byte{byte(v), byte(v >> 8), byte(v >> 16), byte(v >> 24)}
	case uint64:
		return []byte{byte(v), byte(v >> 8), byte(v >> 16), byte(v >> 24),
			byte(v >> 32), byte(v >> 40), byte(v >> 48), byte(v >> 56)}
	case uint8:
		return []byte{byte(v)}
	case uint16:
		return []byte{byte(v), byte(v >> 8)}
	case int8:
		return []byte{byte(v)}
	case int16:
		return []byte{byte(v), byte(v >> 8)}
	case float32:
		bits := math.Float32bits(v)
		return []byte{byte(bits), byte(bits >> 8), byte(bits >> 16), byte(bits >> 24)}
	case float64:
		bits := math.Float64bits(v)
		return []byte{byte(bits), byte(bits >> 8), byte(bits >> 16), byte(bits >> 24),
			byte(bits >> 32), byte(bits >> 40), byte(bits >> 48), byte(bits >> 56)}
	case bool:
		if v {
			return []byte{1}
		}
		return []byte{0}
	default:
		// Fallback to fmt.Sprint for other comparable types
		// This ensures all types get properly hashed
		return []byte(fmt.Sprint(v))
	}
}
