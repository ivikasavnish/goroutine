package goroutine

import (
	"fmt"
	"runtime"
	"testing"
)

func TestNewSuperSlice(t *testing.T) {
	data := []int{1, 2, 3, 4, 5}
	ss := NewSuperSlice(data)

	if ss == nil {
		t.Fatal("NewSuperSlice returned nil")
	}
	if len(ss.GetData()) != 5 {
		t.Errorf("Expected length 5, got %d", len(ss.GetData()))
	}
}

func TestDefaultSuperSliceConfig(t *testing.T) {
	config := DefaultSuperSliceConfig()

	if config.Threshold != 1000 {
		t.Errorf("Expected threshold 1000, got %d", config.Threshold)
	}
	if config.NumWorkers != runtime.NumCPU() {
		t.Errorf("Expected workers %d, got %d", runtime.NumCPU(), config.NumWorkers)
	}
	if config.UseIterable {
		t.Error("Expected UseIterable to be false")
	}
	if config.InPlace {
		t.Error("Expected InPlace to be false")
	}
}

func TestSuperSliceProcessSequential(t *testing.T) {
	data := []int{1, 2, 3, 4, 5}
	ss := NewSuperSlice(data)

	result := ss.Process(func(index int, item int) int {
		return item * 2
	})

	expected := []int{2, 4, 6, 8, 10}
	if len(result) != len(expected) {
		t.Fatalf("Expected length %d, got %d", len(expected), len(result))
	}
	for i := range result {
		if result[i] != expected[i] {
			t.Errorf("At index %d: expected %d, got %d", i, expected[i], result[i])
		}
	}
}

func TestSuperSliceProcessParallel(t *testing.T) {
	// Create a slice above default threshold
	size := 1500
	data := make([]int, size)
	for i := 0; i < size; i++ {
		data[i] = i
	}

	ss := NewSuperSlice(data)
	result := ss.Process(func(index int, item int) int {
		return item * 2
	})

	if len(result) != size {
		t.Fatalf("Expected length %d, got %d", size, len(result))
	}
	for i := range result {
		if result[i] != data[i]*2 {
			t.Errorf("At index %d: expected %d, got %d", i, data[i]*2, result[i])
		}
	}
}

func TestSuperSliceWithThreshold(t *testing.T) {
	data := []int{1, 2, 3, 4, 5}
	ss := NewSuperSlice(data).WithThreshold(3)

	if ss.config.Threshold != 3 {
		t.Errorf("Expected threshold 3, got %d", ss.config.Threshold)
	}

	result := ss.Process(func(index int, item int) int {
		return item + 10
	})

	expected := []int{11, 12, 13, 14, 15}
	for i := range result {
		if result[i] != expected[i] {
			t.Errorf("At index %d: expected %d, got %d", i, expected[i], result[i])
		}
	}
}

func TestSuperSliceWithWorkers(t *testing.T) {
	data := make([]int, 100)
	ss := NewSuperSlice(data).WithWorkers(4)

	if ss.config.NumWorkers != 4 {
		t.Errorf("Expected 4 workers, got %d", ss.config.NumWorkers)
	}
}

func TestSuperSliceInPlace(t *testing.T) {
	data := []int{1, 2, 3, 4, 5}
	originalPtr := &data[0] // Store pointer to first element
	ss := NewSuperSlice(data).WithInPlace()

	result := ss.Process(func(index int, item int) int {
		return item * 10
	})

	// Verify result is same reference as original data
	resultPtr := &result[0]
	if originalPtr != resultPtr {
		t.Error("Expected in-place update to modify original slice, got different slice")
	}

	// Verify values are correct
	for i := range data {
		if data[i] != result[i] {
			t.Errorf("At index %d: data=%d, result=%d", i, data[i], result[i])
		}
		if data[i] != (i+1)*10 {
			t.Errorf("At index %d: expected %d, got %d", i, (i+1)*10, data[i])
		}
	}
}

func TestSuperSliceWithIterable(t *testing.T) {
	size := 1200
	data := make([]int, size)
	for i := 0; i < size; i++ {
		data[i] = i
	}

	ss := NewSuperSlice(data).WithIterable()

	if !ss.config.UseIterable {
		t.Error("Expected UseIterable to be true")
	}

	result := ss.Process(func(index int, item int) int {
		return item + 100
	})

	if len(result) != size {
		t.Fatalf("Expected length %d, got %d", size, len(result))
	}
	for i := range result {
		if result[i] != data[i]+100 {
			t.Errorf("At index %d: expected %d, got %d", i, data[i]+100, result[i])
		}
	}
}

func TestSuperSliceProcessWithError(t *testing.T) {
	data := []int{1, 2, -5, 4, 5}
	ss := NewSuperSlice(data)

	result, err := ss.ProcessWithError(func(index int, item int) (int, error) {
		if item < 0 {
			return 0, fmt.Errorf("negative number: %d", item)
		}
		return item * 2, nil
	})

	if err == nil {
		t.Error("Expected error for negative number, got nil")
	}
	if result != nil {
		t.Error("Expected nil result on error")
	}
}

func TestSuperSliceProcessWithErrorSuccess(t *testing.T) {
	data := []int{1, 2, 3, 4, 5}
	ss := NewSuperSlice(data)

	result, err := ss.ProcessWithError(func(index int, item int) (int, error) {
		if item < 0 {
			return 0, fmt.Errorf("negative number: %d", item)
		}
		return item * 2, nil
	})

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if len(result) != 5 {
		t.Errorf("Expected length 5, got %d", len(result))
	}
}

func TestSuperSliceForEach(t *testing.T) {
	data := []int{1, 2, 3, 4, 5}
	ss := NewSuperSlice(data)

	// Use a slice to collect results safely
	results := make([]int, len(data))
	ss.ForEach(func(index int, item int) {
		results[index] = item
	})

	// Verify all items were processed
	for i, expected := range data {
		if results[i] != expected {
			t.Errorf("At index %d: expected %d, got %d", i, expected, results[i])
		}
	}
}

func TestMapTo(t *testing.T) {
	data := []int{1, 2, 3, 4, 5}
	ss := NewSuperSlice(data)

	result := MapTo(ss, func(index int, item int) string {
		return fmt.Sprintf("Item_%d", item)
	})

	expected := []string{"Item_1", "Item_2", "Item_3", "Item_4", "Item_5"}
	if len(result) != len(expected) {
		t.Fatalf("Expected length %d, got %d", len(expected), len(result))
	}
	for i := range result {
		if result[i] != expected[i] {
			t.Errorf("At index %d: expected %s, got %s", i, expected[i], result[i])
		}
	}
}

func TestSuperSliceFilter(t *testing.T) {
	data := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	ss := NewSuperSlice(data)

	evens := ss.FilterSlice(func(index int, item int) bool {
		return item%2 == 0
	})

	expected := []int{2, 4, 6, 8, 10}
	if len(evens) != len(expected) {
		t.Fatalf("Expected length %d, got %d", len(expected), len(evens))
	}
	for i := range evens {
		if evens[i] != expected[i] {
			t.Errorf("At index %d: expected %d, got %d", i, expected[i], evens[i])
		}
	}
}

func TestSuperSliceFilterParallel(t *testing.T) {
	// Create large slice to trigger parallel processing
	size := 1500
	data := make([]int, size)
	for i := 0; i < size; i++ {
		data[i] = i
	}

	ss := NewSuperSlice(data)
	evens := ss.FilterSlice(func(index int, item int) bool {
		return item%2 == 0
	})

	if len(evens) != size/2 {
		t.Errorf("Expected %d even numbers, got %d", size/2, len(evens))
	}

	// Verify all are even
	for i, v := range evens {
		if v%2 != 0 {
			t.Errorf("At index %d: expected even number, got %d", i, v)
		}
	}
}

func TestSuperSliceLen(t *testing.T) {
	data := []int{1, 2, 3, 4, 5}
	ss := NewSuperSlice(data)

	if ss.Len() != 5 {
		t.Errorf("Expected length 5, got %d", ss.Len())
	}
}

func TestNewSuperSliceWithConfig(t *testing.T) {
	config := &SuperSliceConfig{
		Threshold:   500,
		NumWorkers:  8,
		UseIterable: true,
		InPlace:     true,
	}

	data := []int{1, 2, 3}
	ss := NewSuperSliceWithConfig(data, config)

	if ss.config.Threshold != 500 {
		t.Errorf("Expected threshold 500, got %d", ss.config.Threshold)
	}
	if ss.config.NumWorkers != 8 {
		t.Errorf("Expected 8 workers, got %d", ss.config.NumWorkers)
	}
	if !ss.config.UseIterable {
		t.Error("Expected UseIterable to be true")
	}
	if !ss.config.InPlace {
		t.Error("Expected InPlace to be true")
	}
}

func TestSuperSliceConfigValidation(t *testing.T) {
	// Test that invalid config values are corrected
	config := &SuperSliceConfig{
		Threshold:  -100,
		NumWorkers: 0,
	}

	data := []int{1, 2, 3}
	ss := NewSuperSliceWithConfig(data, config)

	if ss.config.Threshold <= 0 {
		t.Errorf("Expected positive threshold, got %d", ss.config.Threshold)
	}
	if ss.config.NumWorkers <= 0 {
		t.Errorf("Expected positive workers, got %d", ss.config.NumWorkers)
	}
}

// Benchmark tests
func BenchmarkSuperSliceSequential(b *testing.B) {
	data := make([]int, 500)
	for i := 0; i < 500; i++ {
		data[i] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ss := NewSuperSlice(data).WithThreshold(10000)
		ss.Process(func(index int, item int) int {
			return item * 2
		})
	}
}

func BenchmarkSuperSliceParallel(b *testing.B) {
	data := make([]int, 5000)
	for i := 0; i < 5000; i++ {
		data[i] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ss := NewSuperSlice(data)
		ss.Process(func(index int, item int) int {
			return item * 2
		})
	}
}

func BenchmarkSuperSliceIterable(b *testing.B) {
	data := make([]int, 5000)
	for i := 0; i < 5000; i++ {
		data[i] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ss := NewSuperSlice(data).WithIterable()
		ss.Process(func(index int, item int) int {
			return item * 2
		})
	}
}
