# Portless Demo

This example demonstrates how to use the portless CLI tool with a simple Go HTTP server.

## Running the Demo

### Step 1: Build and Install Portless

```bash
# From the repository root
make install
```

Or install via curl:
```bash
curl -fsSL https://raw.githubusercontent.com/ivikasavnish/goroutine/main/install.sh | sh
```

### Step 2: Start the Proxy

In one terminal:
```bash
portless proxy start
```

### Step 3: Run the Demo

In another terminal:
```bash
cd example/portless_demo
portless demo go run main.go
```

### Step 4: Access the Service

Open your browser or use curl:
```bash
curl http://demo.localhost:1355
curl http://demo.localhost:1355/health
```

## What's Happening?

1. Portless finds a free port (e.g., 45123)
2. Sets `PORT=45123` environment variable
3. Starts your Go server with that port
4. Routes `http://demo.localhost:1355` to `http://localhost:45123`

## Run Multiple Services

You can run multiple services at once:

```bash
# Terminal 1
portless api go run main.go

# Terminal 2  
portless frontend go run main.go

# Terminal 3
portless backend go run main.go
```

Access them at:
- http://api.localhost:1355
- http://frontend.localhost:1355
- http://backend.localhost:1355

## Benefits

- No port conflicts
- No need to remember port numbers
- Stable URLs across restarts
- Easy to share with teammates
- Works great in monorepos
