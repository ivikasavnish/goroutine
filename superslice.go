package goroutine

import (
	"runtime"
	"sync"
	"sync/atomic"
)

// Iterator represents a sequence of values that can be traversed.
type Iterator[T any] interface {
	Next() bool
	Value() T
}

// Stream represents a sequence of values with chainable operations.
// We need to add U to the type parameter list of Stream itself
type Stream[T any] struct {
	iterator Iterator[T]
}

type sliceIterator[T any] struct {
	slice []T
	pos   int
}

func (si *sliceIterator[T]) Next() bool {
	if si.pos+1 < len(si.slice) {
		si.pos++
		return true
	}
	return false
}

func (si *sliceIterator[T]) Value() T {
	return si.slice[si.pos]
}

type mapIterator[T, U any] struct {
	source  Iterator[T]
	mapper  func(T) U
	current U
}

func (mi *mapIterator[T, U]) Next() bool {
	if mi.source.Next() {
		sourceVal := mi.source.Value()
		mi.current = mi.mapper(sourceVal)
		return true
	}
	return false
}

func (mi *mapIterator[T, U]) Value() U {
	return mi.current
}

type filterIterator[T any] struct {
	source    Iterator[T]
	predicate func(T) bool
	current   T
	hasNext   bool
}

func (fi *filterIterator[T]) Next() bool {
	for fi.source.Next() {
		value := fi.source.Value()
		if fi.predicate(value) {
			fi.current = value
			fi.hasNext = true
			return true
		}
	}
	fi.hasNext = false
	return false
}

func (fi *filterIterator[T]) Value() T {
	if !fi.hasNext {
		panic("Value called without successful Next()")
	}
	return fi.current
}

// FromSlice creates a new Stream from a slice
func FromSlice[T any](slice []T) *Stream[T] {
	return &Stream[T]{
		iterator: &sliceIterator[T]{
			slice: slice,
			pos:   -1,
		},
	}
}

// Map converts a Stream[T] to a Stream[U]
// Instead of being a method with a type parameter, this is now a function
func Map[T, U any](s *Stream[T], mapper func(T) U) *Stream[U] {
	return &Stream[U]{
		iterator: &mapIterator[T, U]{
			source: s.iterator,
			mapper: mapper,
		},
	}
}

// Filter retains only elements that satisfy the given predicate
func (s *Stream[T]) Filter(predicate func(T) bool) *Stream[T] {
	return &Stream[T]{
		iterator: &filterIterator[T]{
			source:    s.iterator,
			predicate: predicate,
			hasNext:   false,
		},
	}
}

// Collect accumulates all elements into a slice
func (s *Stream[T]) Collect() []T {
	var result []T
	for s.iterator.Next() {
		result = append(result, s.iterator.Value())
	}
	return result
}

// ForEach applies an action to each element
func (s *Stream[T]) ForEach(action func(T)) {
	for s.iterator.Next() {
		action(s.iterator.Value())
	}
}

// Reduce combines elements using a reducer function and initial value
func (s *Stream[T]) Reduce(initial T, reducer func(T, T) T) T {
	result := initial
	for s.iterator.Next() {
		result = reducer(result, s.iterator.Value())
	}
	return result
}

// SuperSliceConfig holds configuration for SuperSlice processing
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

// DefaultSuperSliceConfig returns a configuration with sensible defaults
func DefaultSuperSliceConfig() *SuperSliceConfig {
	return &SuperSliceConfig{
		Threshold:   1000,
		NumWorkers:  runtime.NumCPU(),
		UseIterable: false,
		InPlace:     false,
	}
}

// SuperSlice provides efficient slice processing with configurable parallelization
type SuperSlice[T any] struct {
	data   []T
	config *SuperSliceConfig
}

// NewSuperSlice creates a new SuperSlice with default configuration
func NewSuperSlice[T any](data []T) *SuperSlice[T] {
	return &SuperSlice[T]{
		data:   data,
		config: DefaultSuperSliceConfig(),
	}
}

// NewSuperSliceWithConfig creates a new SuperSlice with custom configuration
func NewSuperSliceWithConfig[T any](data []T, config *SuperSliceConfig) *SuperSlice[T] {
	if config.NumWorkers <= 0 {
		config.NumWorkers = runtime.NumCPU()
	}
	if config.Threshold <= 0 {
		config.Threshold = 1000
	}
	return &SuperSlice[T]{
		data:   data,
		config: config,
	}
}

// WithConfig sets a custom configuration
func (ss *SuperSlice[T]) WithConfig(config *SuperSliceConfig) *SuperSlice[T] {
	if config.NumWorkers <= 0 {
		config.NumWorkers = runtime.NumCPU()
	}
	if config.Threshold <= 0 {
		config.Threshold = 1000
	}
	ss.config = config
	return ss
}

// WithThreshold sets a custom threshold
func (ss *SuperSlice[T]) WithThreshold(threshold int) *SuperSlice[T] {
	ss.config.Threshold = threshold
	return ss
}

// WithWorkers sets the number of workers
func (ss *SuperSlice[T]) WithWorkers(numWorkers int) *SuperSlice[T] {
	if numWorkers <= 0 {
		numWorkers = runtime.NumCPU()
	}
	ss.config.NumWorkers = numWorkers
	return ss
}

// WithInPlace enables in-place updates
func (ss *SuperSlice[T]) WithInPlace() *SuperSlice[T] {
	ss.config.InPlace = true
	return ss
}

// WithIterable enables iterable processing
func (ss *SuperSlice[T]) WithIterable() *SuperSlice[T] {
	ss.config.UseIterable = true
	return ss
}

// Process applies a callback function to each element
// Returns a new slice with the results (or modifies in place if configured)
func (ss *SuperSlice[T]) Process(callback func(index int, item T) T) []T {
	size := len(ss.data)

	// Use normal processing for small slices
	if size < ss.config.Threshold {
		return ss.processSequential(callback)
	}

	// Use parallel processing for large slices
	if ss.config.UseIterable {
		return ss.processIterableParallel(callback)
	}
	return ss.processParallel(callback)
}

// ProcessWithError applies a callback that can return an error
func (ss *SuperSlice[T]) ProcessWithError(callback func(index int, item T) (T, error)) ([]T, error) {
	size := len(ss.data)

	// Use normal processing for small slices
	if size < ss.config.Threshold {
		return ss.processSequentialWithError(callback)
	}

	// Use parallel processing for large slices
	return ss.processParallelWithError(callback)
}

// ForEach applies a callback function to each element without returning results
func (ss *SuperSlice[T]) ForEach(callback func(index int, item T)) {
	size := len(ss.data)

	// Use normal processing for small slices
	if size < ss.config.Threshold {
		ss.forEachSequential(callback)
		return
	}

	// Use parallel processing for large slices
	ss.forEachParallel(callback)
}

// MapTo transforms slice elements to a different type
func MapTo[T, U any](ss *SuperSlice[T], mapper func(index int, item T) U) []U {
	size := len(ss.data)
	result := make([]U, size)

	// Use normal processing for small slices
	if size < ss.config.Threshold {
		for i, item := range ss.data {
			result[i] = mapper(i, item)
		}
		return result
	}

	// Use parallel processing for large slices
	var wg sync.WaitGroup
	chunkSize := (size + ss.config.NumWorkers - 1) / ss.config.NumWorkers

	for w := 0; w < ss.config.NumWorkers; w++ {
		wg.Add(1)
		start := w * chunkSize
		end := start + chunkSize
		if end > size {
			end = size
		}

		go func(start, end int) {
			defer wg.Done()
			for i := start; i < end; i++ {
				result[i] = mapper(i, ss.data[i])
			}
		}(start, end)
	}

	wg.Wait()
	return result
}

// Filter returns a new slice with elements that satisfy the predicate
func (ss *SuperSlice[T]) FilterSlice(predicate func(index int, item T) bool) []T {
	size := len(ss.data)

	// Use normal processing for small slices
	if size < ss.config.Threshold {
		result := make([]T, 0)
		for i, item := range ss.data {
			if predicate(i, item) {
				result = append(result, item)
			}
		}
		return result
	}

	// Use parallel processing for large slices
	// Create a slice to track which elements match
	matches := make([]bool, size)
	var wg sync.WaitGroup
	chunkSize := (size + ss.config.NumWorkers - 1) / ss.config.NumWorkers

	for w := 0; w < ss.config.NumWorkers; w++ {
		wg.Add(1)
		start := w * chunkSize
		end := start + chunkSize
		if end > size {
			end = size
		}

		go func(start, end int) {
			defer wg.Done()
			for i := start; i < end; i++ {
				if predicate(i, ss.data[i]) {
					matches[i] = true
				}
			}
		}(start, end)
	}

	wg.Wait()

	// Collect results in order
	// Pre-count matches for efficient allocation
	matchCount := 0
	for i := 0; i < size; i++ {
		if matches[i] {
			matchCount++
		}
	}
	
	result := make([]T, 0, matchCount)
	for i := 0; i < size; i++ {
		if matches[i] {
			result = append(result, ss.data[i])
		}
	}

	return result
}

// processSequential handles sequential processing
func (ss *SuperSlice[T]) processSequential(callback func(index int, item T) T) []T {
	if ss.config.InPlace {
		for i := range ss.data {
			ss.data[i] = callback(i, ss.data[i])
		}
		return ss.data
	}

	result := make([]T, len(ss.data))
	for i, item := range ss.data {
		result[i] = callback(i, item)
	}
	return result
}

// processParallel handles parallel processing with worker pool
func (ss *SuperSlice[T]) processParallel(callback func(index int, item T) T) []T {
	size := len(ss.data)
	var result []T

	if ss.config.InPlace {
		result = ss.data
	} else {
		result = make([]T, size)
	}

	var wg sync.WaitGroup
	chunkSize := (size + ss.config.NumWorkers - 1) / ss.config.NumWorkers

	for w := 0; w < ss.config.NumWorkers; w++ {
		wg.Add(1)
		start := w * chunkSize
		end := start + chunkSize
		if end > size {
			end = size
		}

		go func(start, end int) {
			defer wg.Done()
			for i := start; i < end; i++ {
				result[i] = callback(i, ss.data[i])
			}
		}(start, end)
	}

	wg.Wait()
	return result
}

// processIterableParallel processes using iterable with parallel execution
func (ss *SuperSlice[T]) processIterableParallel(callback func(index int, item T) T) []T {
	size := len(ss.data)
	var result []T

	if ss.config.InPlace {
		result = ss.data
	} else {
		result = make([]T, size)
	}

	var wg sync.WaitGroup
	jobs := make(chan int, size)

	// Start workers
	for w := 0; w < ss.config.NumWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for index := range jobs {
				result[index] = callback(index, ss.data[index])
			}
		}()
	}

	// Send jobs
	for i := 0; i < size; i++ {
		jobs <- i
	}
	close(jobs)

	wg.Wait()
	return result
}

// processSequentialWithError handles sequential processing with error handling
func (ss *SuperSlice[T]) processSequentialWithError(callback func(index int, item T) (T, error)) ([]T, error) {
	if ss.config.InPlace {
		for i := range ss.data {
			val, err := callback(i, ss.data[i])
			if err != nil {
				return nil, err
			}
			ss.data[i] = val
		}
		return ss.data, nil
	}

	result := make([]T, len(ss.data))
	for i, item := range ss.data {
		val, err := callback(i, item)
		if err != nil {
			return nil, err
		}
		result[i] = val
	}
	return result, nil
}

// processParallelWithError handles parallel processing with error handling
func (ss *SuperSlice[T]) processParallelWithError(callback func(index int, item T) (T, error)) ([]T, error) {
	size := len(ss.data)
	var result []T

	if ss.config.InPlace {
		result = ss.data
	} else {
		result = make([]T, size)
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error
	var errOccurred atomic.Bool
	chunkSize := (size + ss.config.NumWorkers - 1) / ss.config.NumWorkers

	for w := 0; w < ss.config.NumWorkers; w++ {
		wg.Add(1)
		start := w * chunkSize
		end := start + chunkSize
		if end > size {
			end = size
		}

		go func(start, end int) {
			defer wg.Done()
			for i := start; i < end; i++ {
				// Check if an error already occurred using atomic operation
				if errOccurred.Load() {
					return
				}

				val, err := callback(i, ss.data[i])
				if err != nil {
					// CompareAndSwap is atomic - only one goroutine will succeed
					// in setting the flag from false to true
					if errOccurred.CompareAndSwap(false, true) {
						mu.Lock()
						firstErr = err
						mu.Unlock()
					}
					return
				}
				result[i] = val
			}
		}(start, end)
	}

	wg.Wait()
	if firstErr != nil {
		return nil, firstErr
	}
	return result, nil
}

// forEachSequential applies callback sequentially
func (ss *SuperSlice[T]) forEachSequential(callback func(index int, item T)) {
	for i, item := range ss.data {
		callback(i, item)
	}
}

// forEachParallel applies callback in parallel
func (ss *SuperSlice[T]) forEachParallel(callback func(index int, item T)) {
	size := len(ss.data)
	var wg sync.WaitGroup
	chunkSize := (size + ss.config.NumWorkers - 1) / ss.config.NumWorkers

	for w := 0; w < ss.config.NumWorkers; w++ {
		wg.Add(1)
		start := w * chunkSize
		end := start + chunkSize
		if end > size {
			end = size
		}

		go func(start, end int) {
			defer wg.Done()
			for i := start; i < end; i++ {
				callback(i, ss.data[i])
			}
		}(start, end)
	}

	wg.Wait()
}

// GetData returns the underlying slice
func (ss *SuperSlice[T]) GetData() []T {
	return ss.data
}

// Len returns the length of the slice
func (ss *SuperSlice[T]) Len() int {
	return len(ss.data)
}
