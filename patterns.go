package goroutine

import (
	"context"
	"sync"
	"time"
)

// Pipeline represents a stage in a processing pipeline
type Pipeline[T any] struct {
	stages []func(T) T
}

// NewPipeline creates a new pipeline
func NewPipeline[T any]() *Pipeline[T] {
	return &Pipeline[T]{
		stages: make([]func(T) T, 0),
	}
}

// AddStage adds a processing stage to the pipeline
func (p *Pipeline[T]) AddStage(stage func(T) T) *Pipeline[T] {
	p.stages = append(p.stages, stage)
	return p
}

// Execute runs the pipeline on a single item
func (p *Pipeline[T]) Execute(item T) T {
	result := item
	for _, stage := range p.stages {
		result = stage(result)
	}
	return result
}

// ExecuteAsync processes items through the pipeline concurrently
func (p *Pipeline[T]) ExecuteAsync(ctx context.Context, items []T) []T {
	if len(items) == 0 {
		return []T{}
	}

	results := make([]T, len(items))
	var wg sync.WaitGroup

	for i, item := range items {
		wg.Add(1)
		go func(idx int, val T) {
			defer wg.Done()
			select {
			case <-ctx.Done():
				return
			default:
				results[idx] = p.Execute(val)
			}
		}(i, item)
	}

	wg.Wait()
	return results
}

// FanOut distributes work across multiple workers and fans in results
type FanOut[T, R any] struct {
	numWorkers int
}

// NewFanOut creates a new fan-out pattern with specified number of workers
func NewFanOut[T, R any](numWorkers int) *FanOut[T, R] {
	if numWorkers <= 0 {
		numWorkers = 1
	}
	return &FanOut[T, R]{
		numWorkers: numWorkers,
	}
}

// Process fans out work to multiple workers and fans in results
func (f *FanOut[T, R]) Process(ctx context.Context, items []T, processor func(T) R) []R {
	if len(items) == 0 {
		return []R{}
	}

	type job struct {
		index int
		item  T
	}

	jobs := make(chan job, len(items))
	results := make(chan struct {
		index  int
		result R
	}, len(items))

	// Start workers
	var wg sync.WaitGroup
	for w := 0; w < f.numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				select {
				case <-ctx.Done():
					return
				default:
					results <- struct {
						index  int
						result R
					}{j.index, processor(j.item)}
				}
			}
		}()
	}

	// Send jobs
	for i, item := range items {
		select {
		case <-ctx.Done():
			close(jobs)
			return []R{}
		case jobs <- job{i, item}:
		}
	}
	close(jobs)

	// Wait for workers to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results in order
	output := make([]R, len(items))
	for result := range results {
		output[result.index] = result.result
	}

	return output
}

// ProcessWithIndex is a more efficient version that includes index
func (f *FanOut[T, R]) ProcessWithIndex(ctx context.Context, items []T, processor func(int, T) R) []R {
	if len(items) == 0 {
		return []R{}
	}

	type job struct {
		index int
		item  T
	}

	jobs := make(chan job, len(items))
	results := make(chan struct {
		index  int
		result R
	}, len(items))

	// Start workers
	var wg sync.WaitGroup
	for w := 0; w < f.numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				select {
				case <-ctx.Done():
					return
				default:
					results <- struct {
						index  int
						result R
					}{j.index, processor(j.index, j.item)}
				}
			}
		}()
	}

	// Send jobs
	for i, item := range items {
		select {
		case <-ctx.Done():
			close(jobs)
			return []R{}
		case jobs <- job{i, item}:
		}
	}
	close(jobs)

	// Wait for workers to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results in order
	output := make([]R, len(items))
	for result := range results {
		output[result.index] = result.result
	}

	return output
}

// RateLimiter controls the rate of operations
type RateLimiter struct {
	ticker   *time.Ticker
	tokens   chan struct{}
	capacity int
	done     chan struct{}
}

// NewRateLimiter creates a rate limiter with specified rate (operations per second)
func NewRateLimiter(rate int) *RateLimiter {
	if rate <= 0 {
		rate = 1
	}

	interval := time.Second / time.Duration(rate)
	ticker := time.NewTicker(interval)
	tokens := make(chan struct{}, rate)
	done := make(chan struct{})

	// Fill initial tokens
	for i := 0; i < rate; i++ {
		tokens <- struct{}{}
	}

	rl := &RateLimiter{
		ticker:   ticker,
		tokens:   tokens,
		capacity: rate,
		done:     done,
	}

	// Start token refill goroutine
	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				select {
				case tokens <- struct{}{}:
				default:
					// Channel full, skip
				}
			}
		}
	}()

	return rl
}

// Wait blocks until a token is available
func (rl *RateLimiter) Wait(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-rl.tokens:
		return nil
	}
}

// TryAcquire attempts to acquire a token without blocking
func (rl *RateLimiter) TryAcquire() bool {
	select {
	case <-rl.tokens:
		return true
	default:
		return false
	}
}

// Stop stops the rate limiter
func (rl *RateLimiter) Stop() {
	close(rl.done)
	rl.ticker.Stop()
}

// Semaphore controls concurrent access to a resource
type Semaphore struct {
	permits  chan struct{}
	acquired int
	mu       sync.Mutex
}

// NewSemaphore creates a semaphore with specified number of permits
func NewSemaphore(permits int) *Semaphore {
	if permits <= 0 {
		permits = 1
	}
	return &Semaphore{
		permits:  make(chan struct{}, permits),
		acquired: 0,
	}
}

// Acquire acquires a permit, blocking if none available
func (s *Semaphore) Acquire(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case s.permits <- struct{}{}:
		s.mu.Lock()
		s.acquired++
		s.mu.Unlock()
		return nil
	}
}

// TryAcquire attempts to acquire a permit without blocking
func (s *Semaphore) TryAcquire() bool {
	select {
	case s.permits <- struct{}{}:
		s.mu.Lock()
		s.acquired++
		s.mu.Unlock()
		return true
	default:
		return false
	}
}

// Release releases a permit
func (s *Semaphore) Release() {
	s.mu.Lock()
	if s.acquired > 0 {
		s.acquired--
		<-s.permits
	}
	s.mu.Unlock()
}

// AcquireAll acquires all permits
func (s *Semaphore) AcquireAll(ctx context.Context) error {
	capacity := cap(s.permits)
	for i := 0; i < capacity; i++ {
		if err := s.Acquire(ctx); err != nil {
			// Release already acquired permits
			for j := 0; j < i; j++ {
				s.Release()
			}
			return err
		}
	}
	return nil
}

// ReleaseAll releases all acquired permits
func (s *Semaphore) ReleaseAll() {
	s.mu.Lock()
	count := s.acquired
	s.mu.Unlock()
	
	for i := 0; i < count; i++ {
		s.Release()
	}
}

// Generator produces values on demand
type Generator[T any] struct {
	ch     chan T
	ctx    context.Context
	cancel context.CancelFunc
}

// NewGenerator creates a generator with a producer function
func NewGenerator[T any](ctx context.Context, bufferSize int, producer func(context.Context, chan<- T)) *Generator[T] {
	genCtx, cancel := context.WithCancel(ctx)
	ch := make(chan T, bufferSize)

	g := &Generator[T]{
		ch:     ch,
		ctx:    genCtx,
		cancel: cancel,
	}

	go func() {
		defer close(ch)
		producer(genCtx, ch)
	}()

	return g
}

// Next retrieves the next value from the generator
func (g *Generator[T]) Next() (T, bool) {
	val, ok := <-g.ch
	return val, ok
}

// Channel returns the underlying channel for range operations
func (g *Generator[T]) Channel() <-chan T {
	return g.ch
}

// Close stops the generator
func (g *Generator[T]) Close() {
	g.cancel()
}

// Collect collects all remaining values into a slice
func (g *Generator[T]) Collect() []T {
	var results []T
	for val := range g.ch {
		results = append(results, val)
	}
	return results
}

// CollectN collects up to n values
func (g *Generator[T]) CollectN(n int) []T {
	results := make([]T, 0, n)
	for i := 0; i < n; i++ {
		val, ok := <-g.ch
		if !ok {
			break
		}
		results = append(results, val)
	}
	return results
}
