# Feature Flag Demo

This demo showcases the easy-to-use feature flag system with Redis backend, environment-specific configuration, and rollout policies.

## Features Demonstrated

1. **Basic Feature Flags** - Simple on/off switches for features
2. **Environment-Specific Flags** - Different flag values for prod/stage/dev
3. **Feature Rollout Scenario** - Gradual rollout across environments
4. **Rollout Policies** - Gradual, canary, targeted, and all-at-once deployments
5. **Flag Management** - Create, update, delete, and list flags

## Prerequisites

- Go 1.22 or higher
- Redis server running (default: localhost:6379)

## Setup Redis

### Using Docker
```bash
docker run -d -p 6379:6379 redis:latest
```

### Using Homebrew (macOS)
```bash
brew install redis
brew services start redis
```

### Using apt (Ubuntu/Debian)
```bash
sudo apt install redis-server
sudo systemctl start redis-server
```

## Running the Demo

### With default settings (dev environment, localhost Redis)
```bash
cd example/featureflag_demo
go run main.go
```

### Specify environment
```bash
APP_ENV=prod go run main.go
APP_ENV=stage go run main.go
APP_ENV=dev go run main.go
```

### Specify Redis address
```bash
REDIS_ADDR=redis.example.com:6379 go run main.go
```

### Combined
```bash
APP_ENV=prod REDIS_ADDR=localhost:6379 go run main.go
```

## Demo Output

The demo will show:

1. **Basic Feature Flag**
   - Creating a flag
   - Checking if it's enabled
   - Toggling the flag
   - Using the flag in code

2. **Environment-Specific Flags**
   - Creating a flag with different values per environment
   - Checking status across all environments
   - Environment-aware flag evaluation

3. **Feature Rollout Scenario**
   - Simulating a gradual rollout
   - Dev → Staging → Production
   - Progressive feature enablement

4. **Rollout Policies**
   - Gradual rollout with percentage-based deployment
   - Canary rollout with user segment targeting
   - Targeted rollout to specific users
   - All-at-once rollout

5. **List All Flags**
   - Viewing all flags with their settings
   - Rollout policy configurations
   - Metadata including creation and update times

## Usage in Your Application

```go
package main

import (
    "context"
    "os"
    "github.com/ivikasavnish/goroutine"
)

func main() {
    // Get environment from config
    env := goroutine.EnvProduction
    
    // Create feature flag set
    ffs, err := goroutine.NewFeatureFlagSetSimple("localhost:6379", env)
    if err != nil {
        panic(err)
    }
    defer ffs.Close()
    
    ctx := context.Background()
    
    // Check if a feature is enabled
    enabled, _ := ffs.IsEnabled(ctx, "new-checkout-flow")
    
    if enabled {
        // Use new feature
        newCheckout()
    } else {
        // Use old feature
        oldCheckout()
    }
}
```

## Key Benefits

✅ **Simple API** - Easy to understand and use
✅ **Redis Backend** - Distributed, fast, and reliable
✅ **Environment Aware** - Different settings for prod/stage/dev
✅ **Local Caching** - Fast reads with automatic cache invalidation
✅ **Safe Defaults** - Unknown flags default to disabled
✅ **Flexible** - Global and environment-specific overrides

## API Overview

### Creating Flags
```go
// Simple creation
ffs.CreateFlag(ctx, "my-feature", true, "My feature description")

// With environment overrides
flag := &goroutine.FeatureFlag{
    Name:    "my-feature",
    Enabled: true,
    Environments: map[goroutine.Environment]bool{
        goroutine.EnvProduction: false, // Disabled in prod
        goroutine.EnvStaging:    true,  // Enabled in stage
    },
}
ffs.SetFlag(ctx, flag)
```

### Checking Flags
```go
// For current environment
enabled, _ := ffs.IsEnabled(ctx, "my-feature")

// For specific environment
enabled, _ := ffs.IsEnabledForEnv(ctx, "my-feature", goroutine.EnvProduction)

// For specific user with rollout policy
userID := "user123"
userSegments := []string{"beta", "premium"}
enabled, _ := ffs.IsEnabledForUser(ctx, "my-feature", userID, userSegments)
```

### Rollout Policies
```go
// Gradual rollout - 50% of users
ffs.SetGradualRollout(ctx, "my-feature", 50)

// Canary rollout - specific segments
ffs.SetCanaryRollout(ctx, "my-feature", []string{"beta", "internal"})

// Targeted rollout - specific users
ffs.SetTargetedRollout(ctx, "my-feature", []string{"alice", "bob"})

// All-at-once rollout
ffs.SetAllAtOnceRollout(ctx, "my-feature")
```

### Updating Flags
```go
// Update global enabled status
ffs.UpdateFlag(ctx, "my-feature", false)

// Update specific environment
ffs.SetFlagForEnv(ctx, "my-feature", goroutine.EnvProduction, true)
```

### Managing Flags
```go
// List all flags
flags, _ := ffs.ListFlags(ctx)

// Delete a flag
ffs.DeleteFlag(ctx, "my-feature")

// Clear local cache
ffs.ClearCache()
```

## Notes

- Flags are stored in Redis with JSON serialization
- Local cache reduces Redis load (default 30s TTL)
- Non-existent flags default to `false` (safe default)
- Environment overrides take precedence over global settings
- All operations are thread-safe
