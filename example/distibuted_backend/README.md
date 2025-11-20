# Distributed SafeChannel Demo

Runnable examples demonstrating distributed message passing using SafeChannel with two backends:
- **Redis**: Fast, networked pub/sub messaging
- **OS Named Pipes**: Same-host inter-process communication

## Quick Start

### Prerequisites

- Go 1.24.0 or higher
- Redis server (for Redis backend demos) - optional, can skip pipe demos
- Unix-like system (macOS, Linux) for pipe demos

### Running the Demo

```bash
# Build
go build -o safechannel-demo

# Run with Redis backend (default)
./safechannel-demo -backend redis

# Run with Redis backend - multiple producers demo
./safechannel-demo -backend redis -demo multiple-producers

# Run with Redis backend - subscribe demo
./safechannel-demo -backend redis -demo subscribe

# Run with named pipes backend
./safechannel-demo -backend pipe

# Specify custom Redis address
./safechannel-demo -backend redis -redis-addr localhost:6380
```

## Command-Line Flags

```
-backend string
    Backend type: redis or pipe (default "redis")

-redis-addr string
    Redis address (default "localhost:6379")

-pipe-path string
    Named pipe path (default "/tmp/safechannel.pipe")

-demo string
    Demo type: producer-consumer, multiple-producers, subscribe (default "producer-consumer")
```

## Demo Types

### 1. Producer-Consumer (`producer-consumer`)
Basic producer-consumer pattern with one producer sending messages and one consumer receiving them.

```bash
./safechannel-demo -backend redis -demo producer-consumer
```

**Output Example:**
```
=== Redis Distributed SafeChannel Demo ===
✓ Connected to Redis successfully

Backend Type: redis
Demo Type: producer-consumer

Starting Producer-Consumer Demo...
[Producer] Sent: redis-msg-1 - Hello from Redis #1
[Consumer] Received: redis-msg-1 - Hello from Redis #1 (from RedisProducer)
[Producer] Sent: redis-msg-2 - Hello from Redis #2
...
```

### 2. Multiple Producers (`multiple-producers`)
Three producers sending messages concurrently to a single consumer.

```bash
./safechannel-demo -backend redis -demo multiple-producers
```

**Output Example:**
```
Starting Multiple Producers Demo...
[Producer-1] Sent: redis-p1-msg1
[Producer-2] Sent: redis-p2-msg1
[Producer-3] Sent: redis-p3-msg1
[Consumer] [1] Received: redis-p1-msg1 from RedisProducer-1
[Consumer] [2] Received: redis-p2-msg1 from RedisProducer-2
...
```

### 3. Subscribe (`subscribe`)
Pub/sub pattern where a publisher sends messages to subscribers.

```bash
./safechannel-demo -backend redis -demo subscribe
```

**Output Example:**
```
[Subscriber] Received: redis-sub-msg-1 - Subscription message #1
[Publisher] Published: redis-sub-msg-1
[Subscriber] Received: redis-sub-msg-2 - Subscription message #2
...
```

## Backend Comparison

### Redis Backend
- **Use Case**: Distributed systems, cross-machine messaging
- **Features**:
  - Network-based communication
  - Persistent message storage
  - Pub/Sub and list-based patterns
  - Supports multiple clients
  - Automatic fallback to local channel on connection failure

```bash
./safechannel-demo -backend redis -redis-addr localhost:6379
```

### Named Pipes Backend
- **Use Case**: Same-host IPC, low-latency communication
- **Features**:
  - FIFO message ordering
  - No network overhead
  - Zero-copy semantics
  - Best for single-machine applications
  - Works with or without Redis

```bash
./safechannel-demo -backend pipe -pipe-path /tmp/safechannel.pipe
```

## Architecture

### Components

1. **SafeChannel** - Thread-safe channel wrapper with timeout capabilities
   - Located in `../../safechannel.go`
   - Provides Send, Receive, TrySend, TryReceive methods

2. **DistributedSafeChannel** - Extends SafeChannel with backend support
   - Automatic serialization/deserialization
   - Fallback to local channel on backend failure

3. **Backend Implementations**
   - `RedisBackend`: Uses Redis Pub/Sub and list operations
   - `PipeBackend`: Uses OS named pipes for IPC

### Message Flow

```
Producer
   ↓
DistributedSafeChannel.Send()
   ↓
Backend (Redis/Pipe) → Network/IPC
   ↓
DistributedSafeChannel.Receive()
   ↓
Consumer
```

## Code Examples

### Using Redis Backend

```go
import "github.com/ivikasavnish/goroutine"

// Create Redis backend
redisBackend := NewRedisBackend("localhost:6379", "messages")
defer redisBackend.Close()

// Create distributed channel
channel := goroutine.NewDistributedSafeChannel[Message](
    redisBackend,
    "demo-messages",
    100,
    5*time.Second,
)
defer channel.Close()

// Send message
msg := Message{ID: "1", Content: "Hello"}
err := channel.Send(context.Background(), msg)

// Receive message
received, err := channel.Receive(context.Background())
```

### Using Pipes Backend

```go
// Create pipe backend
pipeBackend := NewPipeBackend("/tmp/safechannel.pipe")
defer pipeBackend.Close()

// Create distributed channel
channel := goroutine.NewDistributedSafeChannel[Message](
    pipeBackend,
    "local-topic",
    100,
    5*time.Second,
)
defer channel.Close()

// Use same Send/Receive interface
msg := Message{ID: "1", Content: "Hello"}
err := channel.Send(context.Background(), msg)
```

## Error Handling

The demo includes graceful error handling:
- Connection timeouts for Redis
- Context cancellation for clean shutdown
- Automatic fallback to local channel on backend failure

Press `Ctrl+C` to gracefully shutdown the demo.

## Testing

Run the tests from the parent directory:

```bash
cd ../../
go test ./...
```

## Troubleshooting

### Redis Connection Error
```
Failed to connect to Redis: connection refused
```
**Solution**: Start Redis server
```bash
redis-server
```

### Pipe File Error
```
open /tmp/safechannel.pipe: no such file or directory
```
**Solution**: Create the named pipe first (on Linux/macOS)
```bash
mkfifo /tmp/safechannel.pipe
```

### Port Already in Use
```bash
# Use custom Redis port
./safechannel-demo -backend redis -redis-addr localhost:6380
```

## Performance Characteristics

| Backend | Latency | Throughput | Use Case |
|---------|---------|-----------|----------|
| Redis | ~1ms | High | Distributed systems |
| Pipes | <1ms | Medium | Same-host IPC |
| Local | <0.1ms | Very High | Single process |

## Advanced Usage

### Switching Backends at Runtime
Backends are pluggable via the `DistributedBackend` interface:

```go
var backend goroutine.DistributedBackend
if useRedis {
    backend = NewRedisBackend(addr, topic)
} else {
    backend = NewPipeBackend(pipePath)
}

channel := goroutine.NewDistributedSafeChannel[Message](backend, topic, 100, 5*time.Second)
```

### Custom Message Types
Works with any JSON-serializable type:

```go
type Order struct {
    ID    string `json:"id"`
    Items []Item `json:"items"`
    Total float64 `json:"total"`
}

channel := goroutine.NewDistributedSafeChannel[Order](backend, "orders", 100, 5*time.Second)
```

## Contributing

Examples demonstrate core functionality. See parent documentation for advanced features.
