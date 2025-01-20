package goroutine

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrChannelClosed = errors.New("channel is closed")
	ErrTimeout       = errors.New("operation timed out")
)

// SafeChannel provides a thread-safe wrapper around a channel with timeout capabilities
type SafeChannel[T any] struct {
	ch      chan T
	closed  bool
	mu      sync.RWMutex
	timeout time.Duration
}

// NewSafeChannel creates a new SafeChannel with the specified buffer size and default timeout
func NewSafeChannel[T any](bufferSize int, defaultTimeout time.Duration) *SafeChannel[T] {
	return &SafeChannel[T]{
		ch:      make(chan T, bufferSize),
		timeout: defaultTimeout,
	}
}

// Send attempts to send a value to the channel with timeout
func (sc *SafeChannel[T]) Send(ctx context.Context, value T) error {
	sc.mu.RLock()
	if sc.closed {
		sc.mu.RUnlock()
		return ErrChannelClosed
	}
	sc.mu.RUnlock()

	select {
	case sc.ch <- value:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(sc.timeout):
		return ErrTimeout
	}
}

// Receive attempts to receive a value from the channel with timeout
func (sc *SafeChannel[T]) Receive(ctx context.Context) (T, error) {
	sc.mu.RLock()
	if sc.closed && len(sc.ch) == 0 {
		sc.mu.RUnlock()
		var zero T
		return zero, ErrChannelClosed
	}
	sc.mu.RUnlock()

	select {
	case value, ok := <-sc.ch:
		if !ok {
			var zero T
			return zero, ErrChannelClosed
		}
		return value, nil
	case <-ctx.Done():
		var zero T
		return zero, ctx.Err()
	case <-time.After(sc.timeout):
		var zero T
		return zero, ErrTimeout
	}
}

// TrySend attempts to send a value without blocking
func (sc *SafeChannel[T]) TrySend(value T) error {
	sc.mu.RLock()
	if sc.closed {
		sc.mu.RUnlock()
		return ErrChannelClosed
	}
	sc.mu.RUnlock()

	select {
	case sc.ch <- value:
		return nil
	default:
		return ErrTimeout
	}
}

// TryReceive attempts to receive a value without blocking
func (sc *SafeChannel[T]) TryReceive() (T, error) {
	sc.mu.RLock()
	if sc.closed && len(sc.ch) == 0 {
		sc.mu.RUnlock()
		var zero T
		return zero, ErrChannelClosed
	}
	sc.mu.RUnlock()

	select {
	case value, ok := <-sc.ch:
		if !ok {
			var zero T
			return zero, ErrChannelClosed
		}
		return value, nil
	default:
		var zero T
		return zero, ErrTimeout
	}
}

// Close safely closes the channel
func (sc *SafeChannel[T]) Close() {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if !sc.closed {
		sc.closed = true
		close(sc.ch)
	}
}

// IsClosed checks if the channel is closed
func (sc *SafeChannel[T]) IsClosed() bool {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.closed
}

// Len returns the current number of items in the channel
func (sc *SafeChannel[T]) Len() int {
	return len(sc.ch)
}

// Cap returns the capacity of the channel
func (sc *SafeChannel[T]) Cap() int {
	return cap(sc.ch)
}
