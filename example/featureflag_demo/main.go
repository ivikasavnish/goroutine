package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ivikasavnish/goroutine"
)

func main() {
	fmt.Println("=== Feature Flag Demo ===")
	fmt.Println()

	// Get environment from ENV variable or default to dev
	envStr := os.Getenv("APP_ENV")
	if envStr == "" {
		envStr = "dev"
	}

	var env goroutine.Environment
	switch envStr {
	case "prod":
		env = goroutine.EnvProduction
	case "stage":
		env = goroutine.EnvStaging
	case "dev":
		env = goroutine.EnvDevelopment
	default:
		env = goroutine.EnvDevelopment
	}

	fmt.Printf("Current Environment: %s\n", env)
	fmt.Println()

	// Redis address from ENV variable or default
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	// Create feature flag set
	config := &goroutine.FeatureFlagSetConfig{
		RedisAddr:   redisAddr,
		KeyPrefix:   "demo:featureflag:",
		CacheTTL:    30 * time.Second,
		Environment: env,
	}

	ffs, err := goroutine.NewFeatureFlagSet(config)
	if err != nil {
		log.Fatalf("Failed to create feature flag set: %v\n", err)
	}
	defer ffs.Close()

	fmt.Println("✓ Connected to Redis successfully")
	fmt.Println()

	ctx := context.Background()

	// Demo 1: Basic feature flag
	fmt.Println("--- Demo 1: Basic Feature Flag ---")
	demoBasicFlag(ctx, ffs)

	// Demo 2: Environment-specific flags
	fmt.Println("\n--- Demo 2: Environment-Specific Flags ---")
	demoEnvironmentFlags(ctx, ffs, env)

	// Demo 3: Feature rollout scenario
	fmt.Println("\n--- Demo 3: Feature Rollout Scenario ---")
	demoFeatureRollout(ctx, ffs, env)

	// Demo 4: Rollout Policies
	fmt.Println("\n--- Demo 4: Rollout Policies ---")
	demoRolloutPolicies(ctx, ffs)

	// Demo 5: List all flags
	fmt.Println("\n--- Demo 5: List All Flags ---")
	demoListFlags(ctx, ffs)

	// Clean up demo flags
	fmt.Println("\n--- Cleaning Up Demo Flags ---")
	cleanupFlags(ctx, ffs)

	fmt.Println("\n=== Demo Complete ===")
}

func demoBasicFlag(ctx context.Context, ffs *goroutine.FeatureFlagSet) {
	flagName := "demo-basic-feature"

	// Create a simple flag
	err := ffs.CreateFlag(ctx, flagName, true, "A simple demo feature flag")
	if err != nil {
		log.Printf("Error creating flag: %v\n", err)
		return
	}

	fmt.Printf("Created flag: %s\n", flagName)

	// Check if enabled
	enabled, err := ffs.IsEnabled(ctx, flagName)
	if err != nil {
		log.Printf("Error checking flag: %v\n", err)
		return
	}

	fmt.Printf("Flag '%s' is enabled: %v\n", flagName, enabled)

	// Use the flag
	if enabled {
		fmt.Println("✓ Executing feature code...")
	} else {
		fmt.Println("✗ Feature is disabled")
	}

	// Toggle the flag
	fmt.Println("\nToggling flag to disabled...")
	err = ffs.UpdateFlag(ctx, flagName, false)
	if err != nil {
		log.Printf("Error updating flag: %v\n", err)
		return
	}

	enabled, _ = ffs.IsEnabled(ctx, flagName)
	fmt.Printf("Flag '%s' is now enabled: %v\n", flagName, enabled)
}

func demoEnvironmentFlags(ctx context.Context, ffs *goroutine.FeatureFlagSet, currentEnv goroutine.Environment) {
	flagName := "demo-env-feature"

	// Create a flag enabled globally
	err := ffs.CreateFlag(ctx, flagName, true, "Environment-specific feature flag")
	if err != nil {
		log.Printf("Error creating flag: %v\n", err)
		return
	}

	fmt.Printf("Created flag: %s (globally enabled)\n", flagName)

	// Disable for production only
	err = ffs.SetFlagForEnv(ctx, flagName, goroutine.EnvProduction, false)
	if err != nil {
		log.Printf("Error setting env flag: %v\n", err)
		return
	}

	fmt.Println("Disabled for production environment")

	// Check status across environments
	envs := []goroutine.Environment{
		goroutine.EnvProduction,
		goroutine.EnvStaging,
		goroutine.EnvDevelopment,
	}

	fmt.Println("\nFlag status across environments:")
	for _, env := range envs {
		enabled, _ := ffs.IsEnabledForEnv(ctx, flagName, env)
		marker := "✗"
		if enabled {
			marker = "✓"
		}
		current := ""
		if env == currentEnv {
			current = " (current)"
		}
		fmt.Printf("  %s %s: %v%s\n", marker, env, enabled, current)
	}

	// Check in current environment
	enabled, _ := ffs.IsEnabled(ctx, flagName)
	fmt.Printf("\nIn current environment (%s): %v\n", currentEnv, enabled)
}

func demoFeatureRollout(ctx context.Context, ffs *goroutine.FeatureFlagSet, currentEnv goroutine.Environment) {
	flagName := "demo-new-ui"

	fmt.Println("Scenario: Rolling out a new UI feature")
	fmt.Println()

	// Step 1: Create disabled globally
	err := ffs.CreateFlag(ctx, flagName, false, "New UI redesign")
	if err != nil {
		log.Printf("Error creating flag: %v\n", err)
		return
	}
	fmt.Println("Step 1: Created flag (disabled globally)")

	// Step 2: Enable for dev
	err = ffs.SetFlagForEnv(ctx, flagName, goroutine.EnvDevelopment, true)
	if err != nil {
		log.Printf("Error setting env flag: %v\n", err)
		return
	}
	fmt.Println("Step 2: Enabled for development environment")

	// Step 3: Enable for staging
	time.Sleep(100 * time.Millisecond)
	err = ffs.SetFlagForEnv(ctx, flagName, goroutine.EnvStaging, true)
	if err != nil {
		log.Printf("Error setting env flag: %v\n", err)
		return
	}
	fmt.Println("Step 3: Enabled for staging environment")

	// Step 4: Enable globally (including production)
	time.Sleep(100 * time.Millisecond)
	err = ffs.UpdateFlag(ctx, flagName, true)
	if err != nil {
		log.Printf("Error updating flag: %v\n", err)
		return
	}
	fmt.Println("Step 4: Enabled globally (production rollout complete)")

	fmt.Println("\nFinal status:")
	for _, env := range []goroutine.Environment{goroutine.EnvDevelopment, goroutine.EnvStaging, goroutine.EnvProduction} {
		enabled, _ := ffs.IsEnabledForEnv(ctx, flagName, env)
		fmt.Printf("  %s: %v\n", env, enabled)
	}

	// Demonstrate usage
	enabled, _ := ffs.IsEnabled(ctx, flagName)
	fmt.Printf("\nIn current environment (%s):\n", currentEnv)
	if enabled {
		fmt.Println("  ✓ Showing new UI")
	} else {
		fmt.Println("  ✗ Showing old UI")
	}
}

func demoRolloutPolicies(ctx context.Context, ffs *goroutine.FeatureFlagSet) {
	fmt.Println("Demonstrating different rollout strategies")
	fmt.Println()

	// Demo 1: Gradual Rollout
	fmt.Println("1. Gradual Rollout (50%)")
	flagName := "demo-gradual-feature"
	err := ffs.CreateFlag(ctx, flagName, true, "Feature with gradual rollout")
	if err != nil {
		log.Printf("Error creating flag: %v\n", err)
		return
	}

	// Set 50% gradual rollout
	err = ffs.SetGradualRollout(ctx, flagName, 50)
	if err != nil {
		log.Printf("Error setting gradual rollout: %v\n", err)
		return
	}

	fmt.Println("   Testing with 10 users:")
	enabledCount := 0
	for i := 1; i <= 10; i++ {
		userID := fmt.Sprintf("user%d", i)
		enabled, _ := ffs.IsEnabledForUser(ctx, flagName, userID, nil)
		status := "✗"
		if enabled {
			status = "✓"
			enabledCount++
		}
		fmt.Printf("   %s User %s: %v\n", status, userID, enabled)
	}
	fmt.Printf("   Result: %d/10 users enabled (~50%%)\n", enabledCount)

	// Demo 2: Canary Rollout
	fmt.Println("\n2. Canary Rollout (beta users only)")
	flagName = "demo-canary-feature"
	err = ffs.CreateFlag(ctx, flagName, true, "Feature with canary rollout")
	if err != nil {
		log.Printf("Error creating flag: %v\n", err)
		return
	}

	// Set canary rollout for beta and internal users
	err = ffs.SetCanaryRollout(ctx, flagName, []string{"beta", "internal"})
	if err != nil {
		log.Printf("Error setting canary rollout: %v\n", err)
		return
	}

	fmt.Println("   Testing with different user segments:")
	testUsers := []struct {
		userID   string
		segments []string
	}{
		{"alice", []string{"beta"}},
		{"bob", []string{"internal"}},
		{"charlie", []string{"premium"}},
		{"david", []string{}},
	}

	for _, user := range testUsers {
		enabled, _ := ffs.IsEnabledForUser(ctx, flagName, user.userID, user.segments)
		status := "✗"
		if enabled {
			status = "✓"
		}
		fmt.Printf("   %s User %s (segments: %v): %v\n", status, user.userID, user.segments, enabled)
	}

	// Demo 3: Targeted Rollout
	fmt.Println("\n3. Targeted Rollout (specific users)")
	flagName = "demo-targeted-feature"
	err = ffs.CreateFlag(ctx, flagName, true, "Feature with targeted rollout")
	if err != nil {
		log.Printf("Error creating flag: %v\n", err)
		return
	}

	// Set targeted rollout for specific users
	targetUsers := []string{"alice", "charlie", "eve"}
	err = ffs.SetTargetedRollout(ctx, flagName, targetUsers)
	if err != nil {
		log.Printf("Error setting targeted rollout: %v\n", err)
		return
	}

	fmt.Printf("   Target users: %v\n", targetUsers)
	fmt.Println("   Testing with different users:")
	allUsers := []string{"alice", "bob", "charlie", "david", "eve"}
	for _, userID := range allUsers {
		enabled, _ := ffs.IsEnabledForUser(ctx, flagName, userID, nil)
		status := "✗"
		if enabled {
			status = "✓"
		}
		fmt.Printf("   %s User %s: %v\n", status, userID, enabled)
	}

	// Demo 4: All-at-Once Rollout
	fmt.Println("\n4. All-at-Once Rollout (everyone)")
	flagName = "demo-all-at-once-feature"
	err = ffs.CreateFlag(ctx, flagName, true, "Feature with all-at-once rollout")
	if err != nil {
		log.Printf("Error creating flag: %v\n", err)
		return
	}

	err = ffs.SetAllAtOnceRollout(ctx, flagName)
	if err != nil {
		log.Printf("Error setting all-at-once rollout: %v\n", err)
		return
	}

	fmt.Println("   Testing with 5 users:")
	for i := 1; i <= 5; i++ {
		userID := fmt.Sprintf("user%d", i)
		enabled, _ := ffs.IsEnabledForUser(ctx, flagName, userID, nil)
		status := "✓"
		if !enabled {
			status = "✗"
		}
		fmt.Printf("   %s User %s: %v\n", status, userID, enabled)
	}
	fmt.Println("   Result: All users enabled")
}

func demoListFlags(ctx context.Context, ffs *goroutine.FeatureFlagSet) {
	flags, err := ffs.ListFlags(ctx)
	if err != nil {
		log.Printf("Error listing flags: %v\n", err)
		return
	}

	fmt.Printf("Found %d feature flags:\n", len(flags))
	for i, flag := range flags {
		fmt.Printf("\n%d. %s\n", i+1, flag.Name)
		fmt.Printf("   Description: %s\n", flag.Description)
		fmt.Printf("   Globally Enabled: %v\n", flag.Enabled)
		if len(flag.Environments) > 0 {
			fmt.Println("   Environment Overrides:")
			for env, enabled := range flag.Environments {
				fmt.Printf("     - %s: %v\n", env, enabled)
			}
		}
		if flag.Rollout != nil {
			fmt.Println("   Rollout Policy:")
			fmt.Printf("     - Strategy: %s\n", flag.Rollout.Policy)
			if flag.Rollout.Percentage > 0 {
				fmt.Printf("     - Percentage: %d%%\n", flag.Rollout.Percentage)
			}
			if len(flag.Rollout.CanarySegments) > 0 {
				fmt.Printf("     - Canary Segments: %v\n", flag.Rollout.CanarySegments)
			}
			if len(flag.Rollout.TargetUserIDs) > 0 {
				fmt.Printf("     - Target Users: %v\n", flag.Rollout.TargetUserIDs)
			}
		}
		fmt.Printf("   Created: %s\n", flag.CreatedAt.Format(time.RFC3339))
		fmt.Printf("   Updated: %s\n", flag.UpdatedAt.Format(time.RFC3339))
	}
}

func cleanupFlags(ctx context.Context, ffs *goroutine.FeatureFlagSet) {
	flags, err := ffs.ListFlags(ctx)
	if err != nil {
		log.Printf("Error listing flags: %v\n", err)
		return
	}

	for _, flag := range flags {
		err := ffs.DeleteFlag(ctx, flag.Name)
		if err != nil {
			log.Printf("Error deleting flag %s: %v\n", flag.Name, err)
		} else {
			fmt.Printf("Deleted flag: %s\n", flag.Name)
		}
	}
}
