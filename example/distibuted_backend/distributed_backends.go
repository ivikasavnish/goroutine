package main

import (
	"context"
	"os"
	"sync"

	"github.com/ivikasavnish/goroutine"
	"github.com/redis/go-redis/v9"
)

// RedisBackend implements DistributedBackend for Redis
type RedisBackend struct {
	client *redis.Client
	topic  string
	closed bool
	mu     sync.RWMutex
}

// NewRedisBackend creates a new Redis backend
func NewRedisBackend(addr string, topic string) *RedisBackend {
	return &RedisBackend{
		client: redis.NewClient(&redis.Options{
			Addr: addr,
		}),
		topic: topic,
	}
}

// Send publishes a message to Redis channel
func (rb *RedisBackend) Send(ctx context.Context, topic string, message []byte) error {
	rb.mu.RLock()
	if rb.closed {
		rb.mu.RUnlock()
		return goroutine.ErrChannelClosed
	}
	rb.mu.RUnlock()

	return rb.client.Publish(ctx, topic, message).Err()
}

// Receive receives a message from Redis using blocking list pop
func (rb *RedisBackend) Receive(ctx context.Context, topic string) ([]byte, error) {
	rb.mu.RLock()
	if rb.closed {
		rb.mu.RUnlock()
		return nil, goroutine.ErrChannelClosed
	}
	rb.mu.RUnlock()

	results, err := rb.client.BLPop(ctx, 0, topic).Result()
	if err != nil {
		return nil, err
	}

	if len(results) < 2 {
		return nil, goroutine.ErrBackendFailed
	}

	return []byte(results[1]), nil
}

// Subscribe subscribes to Redis channel and calls handler for each message
func (rb *RedisBackend) Subscribe(ctx context.Context, topic string, handler func([]byte) error) error {
	rb.mu.RLock()
	if rb.closed {
		rb.mu.RUnlock()
		return goroutine.ErrChannelClosed
	}
	rb.mu.RUnlock()

	pubsub := rb.client.Subscribe(ctx, topic)
	defer pubsub.Close()

	ch := pubsub.Channel()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg := <-ch:
			if msg == nil {
				return goroutine.ErrChannelClosed
			}
			if err := handler([]byte(msg.Payload)); err != nil {
				return err
			}
		}
	}
}

// Close closes the Redis backend
func (rb *RedisBackend) Close() error {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	if rb.closed {
		return goroutine.ErrChannelClosed
	}

	rb.closed = true
	return rb.client.Close()
}

// Type returns the backend type
func (rb *RedisBackend) Type() string {
	return "redis"
}

// PipeBackend implements DistributedBackend for OS named pipes
type PipeBackend struct {
	pipePath string
	file     *os.File
	closed   bool
	mu       sync.RWMutex
	writeMu  sync.Mutex
	readMu   sync.Mutex
}

// NewPipeBackend creates a new OS pipe backend
func NewPipeBackend(pipePath string) *PipeBackend {
	return &PipeBackend{
		pipePath: pipePath,
	}
}

// Send sends a message through the named pipe
func (pb *PipeBackend) Send(ctx context.Context, topic string, message []byte) error {
	pb.mu.RLock()
	if pb.closed {
		pb.mu.RUnlock()
		return goroutine.ErrChannelClosed
	}
	pb.mu.RUnlock()

	pb.writeMu.Lock()
	defer pb.writeMu.Unlock()

	file, err := os.OpenFile(pb.pipePath, os.O_WRONLY, os.ModeCharDevice)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write length prefix (4 bytes) followed by message
	length := make([]byte, 4)
	for i := 0; i < 4; i++ {
		length[i] = byte((len(message) >> (8 * (3 - i))) & 0xFF)
	}

	if _, err := file.Write(length); err != nil {
		return err
	}

	if _, err := file.Write(message); err != nil {
		return err
	}

	return nil
}

// Receive receives a message from the named pipe
func (pb *PipeBackend) Receive(ctx context.Context, topic string) ([]byte, error) {
	pb.mu.RLock()
	if pb.closed {
		pb.mu.RUnlock()
		return nil, goroutine.ErrChannelClosed
	}
	pb.mu.RUnlock()

	pb.readMu.Lock()
	defer pb.readMu.Unlock()

	if pb.file == nil {
		file, err := os.OpenFile(pb.pipePath, os.O_RDONLY, os.ModeCharDevice)
		if err != nil {
			return nil, err
		}
		pb.file = file
	}

	// Read length prefix (4 bytes)
	length := make([]byte, 4)
	if _, err := pb.file.Read(length); err != nil {
		return nil, err
	}

	msgLen := int(length[0])<<24 | int(length[1])<<16 | int(length[2])<<8 | int(length[3])

	// Read message
	message := make([]byte, msgLen)
	if _, err := pb.file.Read(message); err != nil {
		return nil, err
	}

	return message, nil
}

// Subscribe subscribes to named pipe and calls handler for each message
func (pb *PipeBackend) Subscribe(ctx context.Context, topic string, handler func([]byte) error) error {
	pb.mu.RLock()
	if pb.closed {
		pb.mu.RUnlock()
		return goroutine.ErrChannelClosed
	}
	pb.mu.RUnlock()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			data, err := pb.Receive(ctx, topic)
			if err != nil {
				return err
			}
			if err := handler(data); err != nil {
				return err
			}
		}
	}
}

// Close closes the pipe backend
func (pb *PipeBackend) Close() error {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	if pb.closed {
		return goroutine.ErrChannelClosed
	}

	pb.closed = true

	if pb.file != nil {
		return pb.file.Close()
	}

	return nil
}

// Type returns the backend type
func (pb *PipeBackend) Type() string {
	return "pipe"
}
