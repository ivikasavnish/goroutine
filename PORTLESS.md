# Portless - Named Localhost URLs for Go

Portless is a command-line tool that replaces traditional port numbers with stable, named `.localhost` URLs for your development servers. Built in Go, inspired by [vercel-labs/portless](https://github.com/vercel-labs/portless).

## Features

- 🎯 **Named URLs**: Access services via `http://myapp.localhost:1355` instead of remembering port numbers
- 🔄 **Automatic Port Management**: No more port conflicts - each service gets a random free port
- 🚀 **Simple CLI**: Easy to use command-line interface
- 🔒 **Process Management**: Automatic service registration and cleanup
- 🌐 **ngrok Integration**: Expose all services to the internet via a single ngrok tunnel
- 📊 **Web Dashboard**: Beautiful Tailwind CSS dashboard to monitor services
- 📦 **Easy Installation**: Install via curl or build from source
- 💻 **Cross-Platform**: Works on Linux and macOS

## Installation

### Quick Install (Linux/macOS)

```bash
curl -fsSL https://raw.githubusercontent.com/ivikasavnish/goroutine/main/install.sh | sh
```

### Manual Installation

#### Prerequisites
- Go 1.22 or later

#### From Source

```bash
git clone https://github.com/ivikasavnish/goroutine.git
cd goroutine
make install
```

Or build without installing:

```bash
make build
# Binary will be in bin/portless
```

## Quick Start

### 1. Start the Proxy Server

In one terminal, start the portless proxy:

```bash
portless proxy start
```

This starts a reverse proxy on port 1355 that routes requests to your services.

### 2. Run Your Services

In separate terminals, run your services with portless:

```bash
# Example: Node.js app
portless myapp npm start

# Example: Go server
portless api go run main.go

# Example: Python Flask app
portless backend python app.py
```

### 3. Access Your Services

Open your browser and navigate to:

- `http://myapp.localhost:1355`
- `http://api.localhost:1355`
- `http://backend.localhost:1355`

### 4. View Dashboard (Optional)

Open the web dashboard to see all registered services:

```bash
portless dashboard
```

This opens `http://localhost:1355/__dashboard` in your browser, showing:
- All registered services with their ports and URLs
- Proxy and ngrok status
- Quick command reference

### 5. Expose to Internet (Optional)

Share your local services with the world using ngrok:

```bash
# Start ngrok tunnel
portless ngrok start

# Check public URL
portless ngrok status

# Stop ngrok
portless ngrok stop
```

All your services become accessible via the ngrok public URL with their service names as subdomains!

## Usage

### Proxy Management

```bash
# Start the proxy server
portless proxy start

# Check proxy status
portless proxy status

# Stop the proxy server
portless proxy stop
```

### Running Services

```bash
# Long form
portless run <service-name> <command> [args...]

# Short form
portless <service-name> <command> [args...]

# Examples
portless webapp npm run dev
portless api go run server.go
portless db docker-compose up
```

### Environment Variables

Portless automatically sets the `PORT` environment variable for your service:

```javascript
// Your app can read the port
const port = process.env.PORT || 3000;
app.listen(port);
```

```go
// Go example
port := os.Getenv("PORT")
if port == "" {
    port = "3000"
}
http.ListenAndServe(":"+port, handler)
```

## How It Works

1. **Proxy Server**: Portless runs a reverse proxy on port 1355
2. **Service Registration**: When you run a service, portless:
   - Finds a free random port
   - Registers the service name → port mapping
   - Sets the `PORT` environment variable
   - Starts your command
3. **Request Routing**: When you access `http://myapp.localhost:1355`:
   - The proxy looks up the registered port for "myapp"
   - Forwards the request to that port
   - Returns the response
4. **ngrok Tunneling** (optional): A single ngrok tunnel exposes ALL services:
   - ngrok connects to your proxy port (1355)
   - All services are accessible via `https://<service>.ngrok-url.com`
   - One tunnel, multiple services!

## ngrok Integration

### Prerequisites

Install ngrok from [ngrok.com/download](https://ngrok.com/download)

### Usage

```bash
# Start ngrok tunnel (requires proxy to be running)
portless ngrok start

# Start with specific region
portless ngrok start --region eu

# Start with custom subdomain (requires ngrok account)
portless ngrok start --subdomain mycompany

# Check status and get public URL
portless ngrok status

# Stop tunnel
portless ngrok stop
```

### How ngrok Works with Portless

When you start an ngrok tunnel, portless:

1. Starts ngrok pointing to the proxy port (1355)
2. Retrieves the public URL from ngrok's API
3. Updates the dashboard to show public URLs for all services

**Example:**
- Local: `http://api.localhost:1355`
- Public: `https://api.abc123.ngrok.io`

All your services automatically get public URLs with their names as subdomains!

### Benefits

- **Single Tunnel**: One ngrok tunnel exposes all services
- **Named URLs**: Services keep their names in public URLs
- **Free Tier**: Works with ngrok's free tier
- **Dynamic**: New services automatically get public URLs

## Web Dashboard

### Accessing the Dashboard

```bash
# Open dashboard (auto-opens browser)
portless dashboard

# Or manually visit
open http://localhost:1355/__dashboard
```

### Dashboard Features

The Tailwind CSS-styled dashboard shows:

- **Status Cards**: Proxy and ngrok tunnel status
- **Service List**: All registered services with:
  - Service name and assigned port
  - Local URLs (`http://<name>.localhost:1355`)
  - Public URLs (if ngrok is active)
- **Quick Commands**: Common portless commands
- **Auto-Refresh**: Updates every 5 seconds

### API Endpoint

Access dashboard data programmatically:

```bash
curl http://localhost:1355/__dashboard/api
```

Returns JSON with all services, ports, and URLs.

## Configuration

Portless stores its configuration in `~/.portless/`:

- `registry.json`: Service name to port mappings
- `proxy.pid`: Proxy server process ID

### Environment Variables

- `PORTLESS_PORT`: Set custom proxy port (default: 1355)
  ```bash
  PORTLESS_PORT=8080 portless proxy start
  ```
- `PORT`: Automatically set for your service by portless
- `SERVICE_NAME`: Automatically set to your service name by portless

## Comparison with Original Portless

| Feature | This Implementation | vercel-labs/portless |
|---------|-------------------|---------------------|
| Language | Go | Node.js/TypeScript |
| Installation | curl \| sh, binary | npm install -g |
| Proxy Port | 1355 | 1355 |
| Named URLs | ✓ | ✓ |
| Auto Port Assignment | ✓ | ✓ |
| Process Management | ✓ | ✓ |

## Examples

### Node.js (Express)

```bash
# In your package.json
{
  "scripts": {
    "dev": "portless myapp nodemon server.js"
  }
}

# Run it
npm run dev
# Access at http://myapp.localhost:1355
```

### Go (HTTP Server)

```bash
# Run your Go server
portless api go run main.go
# Access at http://api.localhost:1355
```

### Python (Flask)

```bash
# Flask automatically reads PORT
portless webapp flask run
# Access at http://webapp.localhost:1355
```

### Multiple Services (Monorepo)

```bash
# Terminal 1
portless frontend npm run dev

# Terminal 2
portless api go run cmd/api/main.go

# Terminal 3
portless auth python auth_service.py

# Access:
# http://frontend.localhost:1355
# http://api.localhost:1355
# http://auth.localhost:1355
```

## Troubleshooting

### "Proxy is not running"

Make sure to start the proxy first:
```bash
portless proxy start
```

### "Service not found"

Ensure your service is running via portless:
```bash
portless proxy status  # Check registered services
```

### Port Already in Use

If port 1355 (default proxy port) is in use, you can use a custom port:
```bash
PORTLESS_PORT=8080 portless proxy start
# Then access services at http://<name>.localhost:8080
```

Or stop the conflicting service to use the default port.

### Service Won't Start

Check that:
- Your command is correct
- Required dependencies are installed
- The service can bind to the assigned PORT

## Uninstallation

```bash
# Stop the proxy
portless proxy stop

# Uninstall the binary
make uninstall
# or
sudo rm /usr/local/bin/portless

# Remove configuration (optional)
rm -rf ~/.portless
```

## Development

### Building

```bash
make build
```

### Build for Multiple Platforms

```bash
make build-all
```

This creates binaries for:
- Linux (amd64, arm64)
- macOS (amd64, arm64)

### Testing

```bash
make test
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see the LICENSE file for details.

## Credits

Inspired by [vercel-labs/portless](https://github.com/vercel-labs/portless).
