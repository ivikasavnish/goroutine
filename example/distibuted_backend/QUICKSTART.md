# Quick Start Guide

## Built Successfully ✓

The demo has been built and is ready to run!

```bash
./safechannel-demo [options]
```

## Running Different Demos

### 1. Redis Backend (Default)

**Simple producer-consumer:**
```bash
./safechannel-demo -backend redis
```

**Multiple producers sending concurrently:**
```bash
./safechannel-demo -backend redis -demo multiple-producers
```

**Pub/Sub pattern:**
```bash
./safechannel-demo -backend redis -demo subscribe
```

**Custom Redis address:**
```bash
./safechannel-demo -backend redis -redis-addr localhost:6380
```

### 2. OS Named Pipes Backend (Local IPC)

**Simple producer-consumer:**
```bash
./safechannel-demo -backend pipe
```

**Multiple producers (sequential):**
```bash
./safechannel-demo -backend pipe -demo multiple-producers
```

**Streaming pattern:**
```bash
./safechannel-demo -backend pipe -demo subscribe
```

**Custom pipe path:**
```bash
./safechannel-demo -backend pipe -pipe-path /tmp/custom_pipe
```

## Prerequisites

### For Redis Backend
- Redis server running on localhost:6379 (or custom address)

Start Redis:
```bash
redis-server
```

Or with Docker:
```bash
docker run -d -p 6379:6379 redis:latest
```

### For Named Pipes Backend
- Unix-like system (macOS, Linux)
- No additional services needed!

Create the pipe (optional, auto-created):
```bash
mkfifo /tmp/safechannel.pipe
```

## Example Output

### Redis Producer-Consumer
```
=== Redis Distributed SafeChannel Demo ===
Connecting to Redis at localhost:6379...
✓ Connected to Redis successfully

Backend Type: redis
Demo Type: producer-consumer

Starting Producer-Consumer Demo...
Producer: Sending 5 messages
Consumer: Receiving and processing

[Producer] Sent: redis-msg-1 - Hello from Redis #1
[Consumer] Received: redis-msg-1 - Hello from Redis #1 (from RedisProducer)
[Producer] Sent: redis-msg-2 - Hello from Redis #2
[Consumer] Received: redis-msg-2 - Hello from Redis #2 (from RedisProducer)
...
Demo completed successfully!
```

### Pipe Multiple Producers
```
=== OS Named Pipe Distributed SafeChannel Demo ===
Using pipe: /tmp/safechannel.pipe
Backend Type: pipe
Demo Type: multiple-producers

Starting Multiple Producers Demo (Sequential)...
[Producer-1] Sent: pipe-p1-msg1
[Producer-1] Sent: pipe-p1-msg2
[Producer-2] Sent: pipe-p2-msg1
[Producer-2] Sent: pipe-p2-msg2
[Consumer] [1] Received: pipe-p1-msg1 from PipeProducer-1
[Consumer] [2] Received: pipe-p1-msg2 from PipeProducer-1
[Consumer] [3] Received: pipe-p2-msg1 from PipeProducer-2
[Consumer] [4] Received: pipe-p2-msg2 from PipeProducer-2
[Consumer] All messages received

Demo completed successfully!
```

## Troubleshooting

### "Failed to connect to Redis"
```bash
redis-server  # Start Redis first
```

### "No such file or directory" (pipes)
```bash
mkfifo /tmp/safechannel.pipe  # Create the pipe first
```

### "Address already in use"
Use a different Redis port:
```bash
./safechannel-demo -backend redis -redis-addr localhost:6380
```

## File Structure

```
example/distibuted_backend/
├── main.go                    # Complete runnable demo with all examples
├── distributed_backends.go    # Redis and Pipe backend implementations
├── go.mod                     # Module definition
├── go.sum                     # Dependency checksums
├── safechannel-demo           # Compiled binary
├── README.md                  # Comprehensive documentation
└── QUICKSTART.md             # This file
```

## Key Features Demonstrated

✅ **Thread-safe message passing**
✅ **Distributed backends (Redis & Pipes)**
✅ **Graceful context handling**
✅ **Automatic fallback to local channel**
✅ **JSON serialization/deserialization**
✅ **Producer-consumer patterns**
✅ **Concurrent producers**
✅ **Pub/Sub patterns**

## Project Structure

The demo is part of the larger SafeChannel project:

```
goroutine/
├── safechannel.go                    # Core SafeChannel implementation
├── distributed_backends.go           # Distributed backend interface
├── examples_test.go                  # 15 comprehensive examples
└── example/
    └── distibuted_backend/
        ├── main.go                   # This demo
        ├── distributed_backends.go   # Backend implementations
        └── README.md                 # Full documentation
```

## Next Steps

1. **Run a demo**: Start with `./safechannel-demo`
2. **Check the code**: Read `main.go` to see how backends are used
3. **Explore patterns**: Try different `-demo` flags
4. **Switch backends**: Use `-backend` flag to compare
5. **Read the docs**: See `README.md` for detailed explanations

## Support

For issues or questions:
- Check `README.md` for detailed documentation
- Review `main.go` source code for examples
- See parent `examples_test.go` for 15+ more examples
