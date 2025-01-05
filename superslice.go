// Package streamify provides a fluent, chainable interface for sequence operations
// with support for lazy evaluation and parallel processing.
package goroutine

// Iterator represents a sequence of values that can be traversed.
type Iterator[T any] interface {
	// Next advances the iterator to the next element.
	// Returns false if there are no more elements.
	Next() bool
	// Value returns the current element.
	// Should only be called after a successful Next().
	Value() T
}

// Stream represents a sequence of values with chainable operations.
type Stream[T any] struct {
	iterator Iterator[T]
}

// sliceIterator implements Iterator for slice types.
type sliceIterator[T any] struct {
	slice []T
	index int
}

func (si *sliceIterator[T]) Next() bool {
	si.index++
	return si.index < len(si.slice)
}

func (si *sliceIterator[T]) Value() T {
	return si.slice[si.index]
}

// mapIterator transforms elements using a mapping function.
type mapIterator[T, U any] struct {
	source Iterator[T]
	mapper func(T) U
}

func (mi *mapIterator[T, U]) Next() bool {
	return mi.source.Next()
}

func (mi *mapIterator[T, U]) Value() U {
	return mi.mapper(mi.source.Value())
}

// filterIterator filters elements using a predicate function.
type filterIterator[T any] struct {
	source    Iterator[T]
	predicate func(T) bool
	current   T
	valid     bool
}

func (fi *filterIterator[T]) Next() bool {
	for fi.source.Next() {
		value := fi.source.Value()
		if fi.predicate(value) {
			fi.current = value
			fi.valid = true
			return true
		}
	}
	fi.valid = false
	return false
}

func (fi *filterIterator[T]) Value() T {
	return fi.current
}

// FromSlice creates a new Stream from a slice.
func FromSlice[T any](slice []T) *Stream[T] {
	return &Stream[T]{
		iterator: &sliceIterator[T]{slice: slice, index: -1},
	}
}

// Map applies a transformation function to each element in the Stream.
func (s *Stream[T]) Map[U any](mapper func(T) U) *Stream[U] {
	return &Stream[U]{
		iterator: &mapIterator[T, U]{
			source: s.iterator,
			mapper: mapper,
		},
	}
}

// Filter retains only elements that satisfy the given predicate.
func (s *Stream[T]) Filter(predicate func(T) bool) *Stream[T] {
	return &Stream[T]{
		iterator: &filterIterator[T]{
			source:    s.iterator,
			predicate: predicate,
		},
	}
}

// Collect accumulates all elements into a slice.
func (s *Stream[T]) Collect() []T {
	var result []T
	for s.iterator.Next() {
		result = append(result, s.iterator.Value())
	}
	return result
}

// ForEach applies an action to each element in the Stream.
func (s *Stream[T]) ForEach(action func(T)) {
	for s.iterator.Next() {
		action(s.iterator.Value())
	}
}

// Reduce combines elements using a reducer function and initial value.
func (s *Stream[T]) Reduce(initial T, reducer func(T, T) T) T {
	result := initial
	for s.iterator.Next() {
		result = reducer(result, s.iterator.Value())
	}
	return result
}