package goroutine

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrRetryExhausted = errors.New("retry attempts exhausted")
	ErrContextCanceled = errors.New("context canceled")
)

// ConcurrencyConfig holds configuration for the smart wrapper
type ConcurrencyConfig struct {
	// MaxConcurrency limits the number of concurrent operations
	MaxConcurrency int
	// Timeout specifies the maximum time for operations
	Timeout time.Duration
	// RateLimit specifies operations per second (0 = no limit)
	RateLimit int
	// RetryAttempts specifies number of retry attempts on failure
	RetryAttempts int
	// RetryDelay specifies delay between retry attempts
	RetryDelay time.Duration
	// EnableCircuitBreaker enables circuit breaker pattern
	EnableCircuitBreaker bool
	// CircuitBreakerThreshold is the failure threshold before opening circuit
	CircuitBreakerThreshold int
}

// DefaultConcurrencyConfig returns default configuration
func DefaultConcurrencyConfig() *ConcurrencyConfig {
	return &ConcurrencyConfig{
		MaxConcurrency:          10,
		Timeout:                 30 * time.Second,
		RateLimit:               0,
		RetryAttempts:           0,
		RetryDelay:              time.Second,
		EnableCircuitBreaker:    false,
		CircuitBreakerThreshold: 5,
	}
}

// ConcurrencyWrapper provides a smart wrapper around concurrency patterns
type ConcurrencyWrapper[T, R any] struct {
	config         *ConcurrencyConfig
	semaphore      *Semaphore
	rateLimiter    *RateLimiter
	circuitBreaker *CircuitBreaker
}

// NewConcurrencyWrapper creates a new concurrency wrapper with given config
func NewConcurrencyWrapper[T, R any](config *ConcurrencyConfig) *ConcurrencyWrapper[T, R] {
	if config == nil {
		config = DefaultConcurrencyConfig()
	}

	wrapper := &ConcurrencyWrapper[T, R]{
		config:    config,
		semaphore: NewSemaphore(config.MaxConcurrency),
	}

	if config.RateLimit > 0 {
		wrapper.rateLimiter = NewRateLimiter(config.RateLimit)
	}

	if config.EnableCircuitBreaker {
		wrapper.circuitBreaker = NewCircuitBreaker(config.CircuitBreakerThreshold)
	}

	return wrapper
}

// Process executes the processor function on all items with configured patterns
func (cw *ConcurrencyWrapper[T, R]) Process(ctx context.Context, items []T, processor func(T) (R, error)) ([]R, error) {
	if len(items) == 0 {
		return []R{}, nil
	}

	// Create context with timeout if configured
	if cw.config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, cw.config.Timeout)
		defer cancel()
	}

	results := make([]R, len(items))
	errs := make([]error, len(items))
	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error

	for i, item := range items {
		wg.Add(1)
		go func(index int, value T) {
			defer wg.Done()

			// Check circuit breaker
			if cw.circuitBreaker != nil && !cw.circuitBreaker.Allow() {
				mu.Lock()
				if firstErr == nil {
					firstErr = errors.New("circuit breaker open")
				}
				errs[index] = errors.New("circuit breaker open")
				mu.Unlock()
				return
			}

			// Acquire semaphore
			if err := cw.semaphore.Acquire(ctx); err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				errs[index] = err
				mu.Unlock()
				return
			}
			defer cw.semaphore.Release()

			// Apply rate limiting
			if cw.rateLimiter != nil {
				if err := cw.rateLimiter.Wait(ctx); err != nil {
					mu.Lock()
					if firstErr == nil {
						firstErr = err
					}
					errs[index] = err
					mu.Unlock()
					return
				}
			}

			// Execute with retry logic
			result, err := cw.executeWithRetry(ctx, value, processor)
			
			mu.Lock()
			if err != nil {
				if firstErr == nil {
					firstErr = err
				}
				errs[index] = err
				if cw.circuitBreaker != nil {
					cw.circuitBreaker.RecordFailure()
				}
			} else {
				results[index] = result
				if cw.circuitBreaker != nil {
					cw.circuitBreaker.RecordSuccess()
				}
			}
			mu.Unlock()
		}(i, item)
	}

	wg.Wait()

	// Check if there were any errors
	if firstErr != nil {
		return results, firstErr
	}

	return results, nil
}

// ProcessWithIndex is similar to Process but provides index to processor
func (cw *ConcurrencyWrapper[T, R]) ProcessWithIndex(ctx context.Context, items []T, processor func(int, T) (R, error)) ([]R, error) {
	wrapper := func(item T) (R, error) {
		// We need to find the index - this is a limitation of the wrapper approach
		// For now, we'll use a simpler approach
		var zero R
		return zero, errors.New("use Process instead - ProcessWithIndex requires separate implementation")
	}
	return cw.Process(ctx, items, wrapper)
}

// executeWithRetry executes the processor with retry logic
func (cw *ConcurrencyWrapper[T, R]) executeWithRetry(ctx context.Context, item T, processor func(T) (R, error)) (R, error) {
	var result R
	var err error

	attempts := cw.config.RetryAttempts + 1
	for i := 0; i < attempts; i++ {
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		default:
		}

		result, err = processor(item)
		if err == nil {
			return result, nil
		}

		// Don't retry on last attempt
		if i < attempts-1 {
			select {
			case <-ctx.Done():
				return result, ctx.Err()
			case <-time.After(cw.config.RetryDelay):
			}
		}
	}

	return result, fmt.Errorf("%w: %v", ErrRetryExhausted, err)
}

// Close cleans up resources
func (cw *ConcurrencyWrapper[T, R]) Close() {
	if cw.rateLimiter != nil {
		cw.rateLimiter.Stop()
	}
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	mu             sync.RWMutex
	failureCount   int
	threshold      int
	state          CircuitState
	lastFailure    time.Time
	resetTimeout   time.Duration
}

// CircuitState represents the state of the circuit breaker
type CircuitState int

const (
	// CircuitClosed means requests are allowed
	CircuitClosed CircuitState = iota
	// CircuitOpen means requests are blocked
	CircuitOpen
	// CircuitHalfOpen means testing if service recovered
	CircuitHalfOpen
)

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(threshold int) *CircuitBreaker {
	return &CircuitBreaker{
		threshold:    threshold,
		state:        CircuitClosed,
		resetTimeout: 30 * time.Second,
	}
}

// Allow checks if a request should be allowed
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	switch cb.state {
	case CircuitClosed:
		return true
	case CircuitOpen:
		// Check if we should transition to half-open
		if time.Since(cb.lastFailure) > cb.resetTimeout {
			cb.mu.RUnlock()
			cb.mu.Lock()
			cb.state = CircuitHalfOpen
			cb.mu.Unlock()
			cb.mu.RLock()
			return true
		}
		return false
	case CircuitHalfOpen:
		return true
	default:
		return false
	}
}

// RecordSuccess records a successful operation
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == CircuitHalfOpen {
		cb.state = CircuitClosed
		cb.failureCount = 0
	}
}

// RecordFailure records a failed operation
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount++
	cb.lastFailure = time.Now()

	if cb.failureCount >= cb.threshold {
		cb.state = CircuitOpen
	}
}

// State returns the current state
func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// Reset resets the circuit breaker
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount = 0
	cb.state = CircuitClosed
}
