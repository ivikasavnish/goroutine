# Portless with ngrok Integration Example

This example demonstrates how to expose your local services to the internet using portless with ngrok integration.

## Prerequisites

1. Install portless
2. Install ngrok from [ngrok.com/download](https://ngrok.com/download)
3. (Optional) Create a free ngrok account for custom features

## Setup

### Step 1: Start the Proxy

```bash
portless proxy start
```

The proxy starts on port 1355 and includes a dashboard.

### Step 2: Start Your Services

In separate terminals, start your services:

```bash
# Terminal 1: API service
portless api go run main.go

# Terminal 2: Frontend service  
portless frontend go run main.go

# Terminal 3: Backend service
portless backend go run main.go
```

### Step 3: View Dashboard (Optional)

```bash
portless dashboard
```

Opens a beautiful Tailwind CSS dashboard showing:
- All registered services
- Service ports and local URLs
- Proxy and ngrok status

### Step 4: Expose to Internet

```bash
portless ngrok start
```

This creates a single ngrok tunnel that exposes ALL your services!

**Example output:**
```
Starting ngrok tunnel on port 1355...

✓ ngrok tunnel started successfully!
✓ Public URL: https://abc123.ngrok.io

Your services are now accessible from the internet:
  - api: https://api.abc123.ngrok.io
  - frontend: https://frontend.abc123.ngrok.io
  - backend: https://backend.abc123.ngrok.io
```

## Advanced Usage

### Custom Region

```bash
portless ngrok start --region eu
```

Available regions: us, eu, ap, au, sa, jp, in

### Custom Subdomain (requires ngrok account)

```bash
portless ngrok start --subdomain mycompany
```

Creates: `https://mycompany.ngrok.io`
Services: `https://api.mycompany.ngrok.io`, etc.

### Check Status

```bash
portless ngrok status
```

Shows:
- Public ngrok URL
- Uptime
- Region
- List of all public service URLs

### Stop Tunnel

```bash
portless ngrok stop
```

## How It Works

1. **Single Tunnel**: Portless starts one ngrok tunnel pointing to the proxy (port 1355)
2. **Named Routing**: The proxy routes requests based on subdomain:
   - `https://api.ngrok-url.com` → `http://localhost:45679` (API service)
   - `https://frontend.ngrok-url.com` → `http://localhost:38291` (Frontend)
3. **Automatic Discovery**: New services automatically get public URLs

## Benefits

### Cost Effective
- Uses single ngrok connection (free tier supports 1 tunnel)
- All services share the same tunnel

### Developer Friendly
- Services keep their names in public URLs
- Easy to share specific services with teammates
- No need to remember port numbers

### Flexible
- Add/remove services without restarting ngrok
- Dashboard shows all URLs (local + public)
- Works with ngrok's free tier

## Security Notes

⚠️ **Warning**: Exposing local services to the internet has security implications:

1. Only expose services during development/testing
2. Don't expose services with sensitive data
3. Use ngrok's authentication features for added security
4. Stop the tunnel when not in use

## Troubleshooting

### "ngrok is not installed"
Install ngrok: https://ngrok.com/download

### "Proxy is not running"
Start the proxy first:
```bash
portless proxy start
```

### Can't access service via ngrok
1. Verify the service is running locally:
   ```bash
   curl http://api.localhost:1355
   ```
2. Check ngrok status:
   ```bash
   portless ngrok status
   ```
3. Check ngrok logs in `~/.portless/ngrok.log`

## Complete Example

```bash
# Terminal 1: Start proxy
portless proxy start

# Terminal 2: Start API
portless api go run api/main.go

# Terminal 3: Start Frontend  
portless frontend npm start

# Terminal 4: Expose via ngrok
portless ngrok start

# Terminal 5: View dashboard
portless dashboard

# Access services:
# Local: http://api.localhost:1355
# Public: https://api.abc123.ngrok.io
```

## Dashboard Screenshot

The dashboard displays:
- ✅ Proxy Status (Running/Stopped)
- ✅ ngrok Status (Active/Inactive)
- ✅ Service count
- ✅ Public ngrok URL (if active)
- ✅ Table of all services with local and public URLs
- ✅ Quick command reference

## Learn More

- [Portless Documentation](../../PORTLESS.md)
- [ngrok Documentation](https://ngrok.com/docs)
