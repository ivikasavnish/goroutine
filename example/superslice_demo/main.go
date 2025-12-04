package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/ivikasavnish/goroutine"
)

func main() {
	fmt.Println("=== SuperSlice Examples ===\n")

	// Example 1: Basic usage with small slice (sequential processing)
	example1BasicSmallSlice()

	// Example 2: Large slice with parallel processing
	example2LargeSliceParallel()

	// Example 3: In-place updates
	example3InPlaceUpdate()

	// Example 4: Custom threshold configuration
	example4CustomThreshold()

	// Example 5: Custom worker count
	example5CustomWorkers()

	// Example 6: Using iterable mode
	example6IterableMode()

	// Example 7: Error handling
	example7ErrorHandling()

	// Example 8: ForEach operation
	example8ForEach()

	// Example 9: MapTo different type
	example9MapTo()

	// Example 10: Filter operation
	example10Filter()

	// Example 11: Performance comparison
	example11PerformanceComparison()

	// Advanced examples
	fmt.Println("=== Advanced Examples ===\n")
	exampleAdvanced1ComplexProcessing()
	exampleAdvanced2StringProcessing()
	exampleAdvanced3ChainedOperations()
	exampleAdvanced4DataNormalization()
	exampleAdvanced5BatchProcessing()
	exampleAdvanced6ConditionalProcessing()
	exampleAdvanced7TypeConversion()
}

func example1BasicSmallSlice() {
	fmt.Println("Example 1: Basic usage with small slice (sequential)")
	numbers := []int{1, 2, 3, 4, 5}
	ss := goroutine.NewSuperSlice(numbers)

	result := ss.Process(func(index int, item int) int {
		return item * 2
	})

	fmt.Printf("Original: %v\n", numbers)
	fmt.Printf("Doubled: %v\n\n", result)
}

func example2LargeSliceParallel() {
	fmt.Println("Example 2: Large slice with parallel processing")
	// Create a large slice (above threshold)
	size := 5000
	numbers := make([]int, size)
	for i := 0; i < size; i++ {
		numbers[i] = i + 1
	}

	ss := goroutine.NewSuperSlice(numbers)
	start := time.Now()

	result := ss.Process(func(index int, item int) int {
		// Simulate some computation
		return item * item
	})

	elapsed := time.Since(start)
	fmt.Printf("Processed %d items in %v\n", len(result), elapsed)
	fmt.Printf("First 10 results: %v\n\n", result[:10])
}

func example3InPlaceUpdate() {
	fmt.Println("Example 3: In-place updates")
	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	ss := goroutine.NewSuperSlice(numbers).WithInPlace()

	fmt.Printf("Before: %v\n", numbers)
	ss.Process(func(index int, item int) int {
		return item * 10
	})
	fmt.Printf("After in-place: %v\n\n", numbers)
}

func example4CustomThreshold() {
	fmt.Println("Example 4: Custom threshold configuration")
	numbers := make([]int, 100)
	for i := 0; i < 100; i++ {
		numbers[i] = i
	}

	// Set threshold to 50, so 100 items will use parallel processing
	ss := goroutine.NewSuperSlice(numbers).WithThreshold(50)

	result := ss.Process(func(index int, item int) int {
		return item + 100
	})

	fmt.Printf("Processed with custom threshold (50)\n")
	fmt.Printf("First 10 results: %v\n\n", result[:10])
}

func example5CustomWorkers() {
	fmt.Println("Example 5: Custom worker count")
	size := 2000
	numbers := make([]int, size)
	for i := 0; i < size; i++ {
		numbers[i] = i
	}

	// Use 4 workers instead of default NumCPU
	ss := goroutine.NewSuperSlice(numbers).WithWorkers(4)

	start := time.Now()
	result := ss.Process(func(index int, item int) int {
		return item * 3
	})
	elapsed := time.Since(start)

	fmt.Printf("Processed %d items with 4 workers in %v\n", len(result), elapsed)
	fmt.Printf("Sample results: %v\n\n", result[990:1000])
}

func example6IterableMode() {
	fmt.Println("Example 6: Using iterable mode")
	size := 1500
	numbers := make([]int, size)
	for i := 0; i < size; i++ {
		numbers[i] = i
	}

	ss := goroutine.NewSuperSlice(numbers).WithIterable()

	result := ss.Process(func(index int, item int) int {
		return item + 1000
	})

	fmt.Printf("Processed %d items in iterable mode\n", len(result))
	fmt.Printf("Sample results: %v\n\n", result[0:5])
}

func example7ErrorHandling() {
	fmt.Println("Example 7: Error handling")
	numbers := []int{1, 2, -5, 4, 5}
	ss := goroutine.NewSuperSlice(numbers)

	result, err := ss.ProcessWithError(func(index int, item int) (int, error) {
		if item < 0 {
			return 0, fmt.Errorf("negative number at index %d: %d", index, item)
		}
		return item * 2, nil
	})

	if err != nil {
		fmt.Printf("Error occurred: %v\n\n", err)
	} else {
		fmt.Printf("Result: %v\n\n", result)
	}
}

func example8ForEach() {
	fmt.Println("Example 8: ForEach operation")
	numbers := []int{10, 20, 30, 40, 50}
	ss := goroutine.NewSuperSlice(numbers)

	fmt.Println("Printing each item:")
	ss.ForEach(func(index int, item int) {
		fmt.Printf("  Index %d: %d\n", index, item)
	})
	fmt.Println()
}

func example9MapTo() {
	fmt.Println("Example 9: MapTo different type")
	numbers := []int{1, 2, 3, 4, 5}
	ss := goroutine.NewSuperSlice(numbers)

	// Map integers to strings
	strings := goroutine.MapTo(ss, func(index int, item int) string {
		return fmt.Sprintf("Number_%d", item)
	})

	fmt.Printf("Original: %v\n", numbers)
	fmt.Printf("Mapped to strings: %v\n\n", strings)
}

func example10Filter() {
	fmt.Println("Example 10: Filter operation")
	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	ss := goroutine.NewSuperSlice(numbers)

	// Filter even numbers
	evens := ss.FilterSlice(func(index int, item int) bool {
		return item%2 == 0
	})

	fmt.Printf("Original: %v\n", numbers)
	fmt.Printf("Even numbers: %v\n\n", evens)
}

func example11PerformanceComparison() {
	fmt.Println("Example 11: Performance comparison")
	size := 10000
	numbers := make([]float64, size)
	for i := 0; i < size; i++ {
		numbers[i] = float64(i)
	}

	// Sequential (below threshold)
	ss1 := goroutine.NewSuperSlice(numbers).WithThreshold(20000)
	start1 := time.Now()
	result1 := ss1.Process(func(index int, item float64) float64 {
		return math.Sqrt(item) * math.Sin(item)
	})
	elapsed1 := time.Since(start1)

	// Parallel (above threshold)
	ss2 := goroutine.NewSuperSlice(numbers).WithThreshold(100)
	start2 := time.Now()
	result2 := ss2.Process(func(index int, item float64) float64 {
		return math.Sqrt(item) * math.Sin(item)
	})
	elapsed2 := time.Since(start2)

	fmt.Printf("Sequential processing: %v\n", elapsed1)
	fmt.Printf("Parallel processing: %v\n", elapsed2)
	fmt.Printf("Speedup: %.2fx\n", float64(elapsed1)/float64(elapsed2))
	fmt.Printf("Results match: %v\n\n", math.Abs(result1[100]-result2[100]) < 0.0001)
}

// Additional advanced examples

func exampleAdvanced1ComplexProcessing() {
	fmt.Println("Advanced Example 1: Complex data processing")

	type Record struct {
		ID    int
		Value string
		Score float64
	}

	// Create sample records
	records := make([]Record, 2000)
	for i := 0; i < 2000; i++ {
		records[i] = Record{
			ID:    i,
			Value: fmt.Sprintf("Record_%d", i),
			Score: float64(i) * 0.5,
		}
	}

	ss := goroutine.NewSuperSlice(records).WithWorkers(8)

	// Transform records
	result := ss.Process(func(index int, item Record) Record {
		return Record{
			ID:    item.ID,
			Value: strings.ToUpper(item.Value),
			Score: item.Score * 2,
		}
	})

	fmt.Printf("Processed %d records\n", len(result))
	fmt.Printf("Sample: ID=%d, Value=%s, Score=%.2f\n\n",
		result[10].ID, result[10].Value, result[10].Score)
}

func exampleAdvanced2StringProcessing() {
	fmt.Println("Advanced Example 2: String processing")

	words := []string{"hello", "world", "parallel", "processing", "goroutines", "efficient"}
	ss := goroutine.NewSuperSlice(words)

	// Convert to uppercase and add prefix
	result := ss.Process(func(index int, item string) string {
		return fmt.Sprintf("[%d]_%s", index, strings.ToUpper(item))
	})

	fmt.Printf("Original: %v\n", words)
	fmt.Printf("Processed: %v\n\n", result)
}

func exampleAdvanced3ChainedOperations() {
	fmt.Println("Advanced Example 3: Chained operations")

	numbers := make([]int, 3000)
	for i := 0; i < 3000; i++ {
		numbers[i] = i
	}

	ss := goroutine.NewSuperSlice(numbers).
		WithThreshold(500).
		WithWorkers(4)

	// First transformation
	squared := ss.Process(func(index int, item int) int {
		return item * item
	})

	// Second transformation on result
	ss2 := goroutine.NewSuperSlice(squared)
	filtered := ss2.FilterSlice(func(index int, item int) bool {
		return item%2 == 0
	})

	fmt.Printf("Original count: %d\n", len(numbers))
	fmt.Printf("Squared count: %d\n", len(squared))
	fmt.Printf("Filtered even count: %d\n", len(filtered))
	fmt.Printf("Sample filtered: %v\n\n", filtered[:5])
}

func exampleAdvanced4DataNormalization() {
	fmt.Println("Advanced Example 4: Data normalization")

	data := []float64{10.5, 20.3, 15.7, 30.2, 25.8, 18.4}
	
	// Calculate mean
	sum := 0.0
	for _, v := range data {
		sum += v
	}
	mean := sum / float64(len(data))

	// Normalize using SuperSlice
	ss := goroutine.NewSuperSlice(data)
	normalized := ss.Process(func(index int, item float64) float64 {
		return (item - mean) / mean
	})

	fmt.Printf("Original: %v\n", data)
	fmt.Printf("Mean: %.2f\n", mean)
	fmt.Printf("Normalized: %v\n\n", normalized)
}

func exampleAdvanced5BatchProcessing() {
	fmt.Println("Advanced Example 5: Batch processing with custom config")

	config := &goroutine.SuperSliceConfig{
		Threshold:   100,
		NumWorkers:  6,
		UseIterable: true,
		InPlace:     false,
	}

	numbers := make([]int, 5000)
	for i := 0; i < 5000; i++ {
		numbers[i] = i
	}

	ss := goroutine.NewSuperSliceWithConfig(numbers, config)

	start := time.Now()
	result := ss.Process(func(index int, item int) int {
		// Simulate complex computation
		val := item
		for j := 0; j < 10; j++ {
			val = (val * 13) % 1000
		}
		return val
	})
	elapsed := time.Since(start)

	fmt.Printf("Batch processed %d items in %v\n", len(result), elapsed)
	fmt.Printf("Configuration: Threshold=%d, Workers=%d, Iterable=%v\n",
		config.Threshold, config.NumWorkers, config.UseIterable)
	fmt.Printf("Sample: %v\n\n", result[100:110])
}

func exampleAdvanced6ConditionalProcessing() {
	fmt.Println("Advanced Example 6: Conditional processing")

	numbers := make([]int, 2500)
	for i := 0; i < 2500; i++ {
		numbers[i] = i
	}

	ss := goroutine.NewSuperSlice(numbers).WithThreshold(1000)

	result := ss.Process(func(index int, item int) int {
		if item%2 == 0 {
			return item * 2
		}
		return item * 3
	})

	fmt.Printf("Processed %d items with conditional logic\n", len(result))
	fmt.Printf("Sample results: %v\n\n", result[10:20])
}

func exampleAdvanced7TypeConversion() {
	fmt.Println("Advanced Example 7: Type conversion with MapTo")

	numbers := make([]int, 1500)
	for i := 0; i < 1500; i++ {
		numbers[i] = i + 1
	}

	ss := goroutine.NewSuperSlice(numbers).WithThreshold(500)

	// Convert to strings with formatting
	strings := goroutine.MapTo(ss, func(index int, item int) string {
		return fmt.Sprintf("%04d: %s", item, strconv.Itoa(item*item))
	})

	fmt.Printf("Converted %d integers to formatted strings\n", len(strings))
	fmt.Printf("Samples:\n")
	for i := 0; i < 5; i++ {
		fmt.Printf("  %s\n", strings[i])
	}
	fmt.Println()
}
