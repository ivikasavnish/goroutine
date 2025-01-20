package goroutine

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
