package goroutine

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"
)

var (
	ErrChannelClosed   = errors.New("channel is closed")
	ErrTimeout         = errors.New("operation timed out")
	ErrBackendFailed   = errors.New("backend operation failed")
	ErrInvalidBackend  = errors.New("invalid backend")
	ErrNotImplemented  = errors.New("not implemented")
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

// DistributedBackend interface for various message backends
type DistributedBackend interface {
	Send(ctx context.Context, topic string, message []byte) error
	Receive(ctx context.Context, topic string) ([]byte, error)
	Subscribe(ctx context.Context, topic string, handler func([]byte) error) error
	Close() error
	Type() string
}

// DistributedSafeChannel wraps SafeChannel with distributed backend support
type DistributedSafeChannel[T any] struct {
	safeChannel *SafeChannel[T]
	backend     DistributedBackend
	topic       string
	closed      bool
	mu          sync.RWMutex
}

// NewDistributedSafeChannel creates a distributed safe channel with specified backend
func NewDistributedSafeChannel[T any](backend DistributedBackend, topic string, bufferSize int, timeout time.Duration) *DistributedSafeChannel[T] {
	return &DistributedSafeChannel[T]{
		safeChannel: NewSafeChannel[T](bufferSize, timeout),
		backend:     backend,
		topic:       topic,
	}
}

// Send serializes and sends value through the backend with fallback to local channel
func (dsc *DistributedSafeChannel[T]) Send(ctx context.Context, value T) error {
	dsc.mu.RLock()
	if dsc.closed {
		dsc.mu.RUnlock()
		return ErrChannelClosed
	}
	dsc.mu.RUnlock()

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	// Try distributed backend first
	if dsc.backend != nil {
		if err := dsc.backend.Send(ctx, dsc.topic, data); err == nil {
			return nil
		}
		// Fall back to local channel on backend failure
	}

	// Use local channel as fallback
	return dsc.safeChannel.Send(ctx, value)
}

// Receive deserializes and receives value from backend with fallback to local channel
func (dsc *DistributedSafeChannel[T]) Receive(ctx context.Context) (T, error) {
	var zero T

	dsc.mu.RLock()
	if dsc.closed {
		dsc.mu.RUnlock()
		return zero, ErrChannelClosed
	}
	dsc.mu.RUnlock()

	// Try distributed backend first
	if dsc.backend != nil {
		data, err := dsc.backend.Receive(ctx, dsc.topic)
		if err == nil {
			var value T
			if err := json.Unmarshal(data, &value); err == nil {
				return value, nil
			}
		}
		// Fall back to local channel on backend failure
	}

	// Use local channel as fallback
	return dsc.safeChannel.Receive(ctx)
}

// Subscribe starts listening to backend messages and forwards them to local channel
func (dsc *DistributedSafeChannel[T]) Subscribe(ctx context.Context) error {
	dsc.mu.RLock()
	if dsc.closed || dsc.backend == nil {
		dsc.mu.RUnlock()
		return ErrChannelClosed
	}
	dsc.mu.RUnlock()

	return dsc.backend.Subscribe(ctx, dsc.topic, func(data []byte) error {
		var value T
		if err := json.Unmarshal(data, &value); err != nil {
			return err
		}
		return dsc.safeChannel.Send(ctx, value)
	})
}

// Close closes both backend and local channel
func (dsc *DistributedSafeChannel[T]) Close() error {
	dsc.mu.Lock()
	defer dsc.mu.Unlock()

	if dsc.closed {
		return ErrChannelClosed
	}

	dsc.closed = true
	dsc.safeChannel.Close()

	if dsc.backend != nil {
		return dsc.backend.Close()
	}

	return nil
}

// IsClosed checks if the distributed channel is closed
func (dsc *DistributedSafeChannel[T]) IsClosed() bool {
	dsc.mu.RLock()
	defer dsc.mu.RUnlock()
	return dsc.closed
}

// BackendType returns the type of backend in use
func (dsc *DistributedSafeChannel[T]) BackendType() string {
	if dsc.backend == nil {
		return "local"
	}
	return dsc.backend.Type()
}
