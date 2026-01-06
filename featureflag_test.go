package goroutine

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestFeatureFlag_IsEnabledForEnv(t *testing.T) {
	tests := []struct {
		name     string
		flag     *FeatureFlag
		env      Environment
		expected bool
	}{
		{
			name: "Global enabled, no env override",
			flag: &FeatureFlag{
				Name:         "test-flag",
				Enabled:      true,
				Environments: map[Environment]bool{},
			},
			env:      EnvProduction,
			expected: true,
		},
		{
			name: "Global disabled, no env override",
			flag: &FeatureFlag{
				Name:         "test-flag",
				Enabled:      false,
				Environments: map[Environment]bool{},
			},
			env:      EnvProduction,
			expected: false,
		},
		{
			name: "Global enabled, env override disabled",
			flag: &FeatureFlag{
				Name:    "test-flag",
				Enabled: true,
				Environments: map[Environment]bool{
					EnvProduction: false,
				},
			},
			env:      EnvProduction,
			expected: false,
		},
		{
			name: "Global disabled, env override enabled",
			flag: &FeatureFlag{
				Name:    "test-flag",
				Enabled: false,
				Environments: map[Environment]bool{
					EnvDevelopment: true,
				},
			},
			env:      EnvDevelopment,
			expected: true,
		},
		{
			name: "Global enabled, check different env without override",
			flag: &FeatureFlag{
				Name:    "test-flag",
				Enabled: true,
				Environments: map[Environment]bool{
					EnvProduction: false,
				},
			},
			env:      EnvStaging,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.flag.IsEnabledForEnv(tt.env)
			if result != tt.expected {
				t.Errorf("IsEnabledForEnv() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFeatureFlagSet_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This test requires Redis to be running
	// Skip if Redis is not available
	ctx := context.Background()
	config := &FeatureFlagSetConfig{
		RedisAddr:   "localhost:6379",
		KeyPrefix:   "test:featureflag:",
		CacheTTL:    5 * time.Second,
		Environment: EnvDevelopment,
	}

	ffs, err := NewFeatureFlagSet(config)
	if err != nil {
		t.Skipf("Skipping test: Redis not available: %v", err)
		return
	}
	defer ffs.Close()

	// Clean up test flags
	defer func() {
		flags, _ := ffs.ListFlags(ctx)
		for _, flag := range flags {
			ffs.DeleteFlag(ctx, flag.Name)
		}
	}()

	t.Run("CreateAndGetFlag", func(t *testing.T) {
		flagName := "test-create-flag"
		err := ffs.CreateFlag(ctx, flagName, true, "Test flag for creation")
		if err != nil {
			t.Fatalf("CreateFlag() error = %v", err)
		}

		flag, err := ffs.GetFlag(ctx, flagName)
		if err != nil {
			t.Fatalf("GetFlag() error = %v", err)
		}

		if flag.Name != flagName {
			t.Errorf("Flag name = %v, want %v", flag.Name, flagName)
		}
		if !flag.Enabled {
			t.Errorf("Flag enabled = %v, want %v", flag.Enabled, true)
		}
		if flag.Description != "Test flag for creation" {
			t.Errorf("Flag description = %v, want %v", flag.Description, "Test flag for creation")
		}

		// Clean up
		ffs.DeleteFlag(ctx, flagName)
	})

	t.Run("IsEnabled", func(t *testing.T) {
		flagName := "test-enabled-flag"
		err := ffs.CreateFlag(ctx, flagName, true, "Test enabled flag")
		if err != nil {
			t.Fatalf("CreateFlag() error = %v", err)
		}
		defer ffs.DeleteFlag(ctx, flagName)

		enabled, err := ffs.IsEnabled(ctx, flagName)
		if err != nil {
			t.Fatalf("IsEnabled() error = %v", err)
		}
		if !enabled {
			t.Errorf("IsEnabled() = %v, want %v", enabled, true)
		}
	})

	t.Run("UpdateFlag", func(t *testing.T) {
		flagName := "test-update-flag"
		err := ffs.CreateFlag(ctx, flagName, true, "Test update flag")
		if err != nil {
			t.Fatalf("CreateFlag() error = %v", err)
		}
		defer ffs.DeleteFlag(ctx, flagName)

		// Update to disabled
		err = ffs.UpdateFlag(ctx, flagName, false)
		if err != nil {
			t.Fatalf("UpdateFlag() error = %v", err)
		}

		enabled, err := ffs.IsEnabled(ctx, flagName)
		if err != nil {
			t.Fatalf("IsEnabled() error = %v", err)
		}
		if enabled {
			t.Errorf("IsEnabled() = %v, want %v", enabled, false)
		}
	})

	t.Run("SetFlagForEnv", func(t *testing.T) {
		flagName := "test-env-flag"
		err := ffs.CreateFlag(ctx, flagName, true, "Test environment flag")
		if err != nil {
			t.Fatalf("CreateFlag() error = %v", err)
		}
		defer ffs.DeleteFlag(ctx, flagName)

		// Set production to false
		err = ffs.SetFlagForEnv(ctx, flagName, EnvProduction, false)
		if err != nil {
			t.Fatalf("SetFlagForEnv() error = %v", err)
		}

		// Check production (should be false)
		enabled, err := ffs.IsEnabledForEnv(ctx, flagName, EnvProduction)
		if err != nil {
			t.Fatalf("IsEnabledForEnv() error = %v", err)
		}
		if enabled {
			t.Errorf("IsEnabledForEnv(prod) = %v, want %v", enabled, false)
		}

		// Check dev (should fall back to global true)
		enabled, err = ffs.IsEnabledForEnv(ctx, flagName, EnvDevelopment)
		if err != nil {
			t.Fatalf("IsEnabledForEnv() error = %v", err)
		}
		if !enabled {
			t.Errorf("IsEnabledForEnv(dev) = %v, want %v", enabled, true)
		}
	})

	t.Run("DeleteFlag", func(t *testing.T) {
		flagName := "test-delete-flag"
		err := ffs.CreateFlag(ctx, flagName, true, "Test delete flag")
		if err != nil {
			t.Fatalf("CreateFlag() error = %v", err)
		}

		err = ffs.DeleteFlag(ctx, flagName)
		if err != nil {
			t.Fatalf("DeleteFlag() error = %v", err)
		}

		_, err = ffs.GetFlag(ctx, flagName)
		if err == nil {
			t.Errorf("GetFlag() should return error for deleted flag")
		}
	})

	t.Run("ListFlags", func(t *testing.T) {
		// Create multiple flags
		flagNames := []string{"test-list-flag-1", "test-list-flag-2", "test-list-flag-3"}
		for _, name := range flagNames {
			err := ffs.CreateFlag(ctx, name, true, "Test list flag")
			if err != nil {
				t.Fatalf("CreateFlag() error = %v", err)
			}
			defer ffs.DeleteFlag(ctx, name)
		}

		flags, err := ffs.ListFlags(ctx)
		if err != nil {
			t.Fatalf("ListFlags() error = %v", err)
		}

		if len(flags) < len(flagNames) {
			t.Errorf("ListFlags() returned %d flags, want at least %d", len(flags), len(flagNames))
		}
	})

	t.Run("CacheWorks", func(t *testing.T) {
		flagName := "test-cache-flag"
		err := ffs.CreateFlag(ctx, flagName, true, "Test cache flag")
		if err != nil {
			t.Fatalf("CreateFlag() error = %v", err)
		}
		defer ffs.DeleteFlag(ctx, flagName)

		// First call - should fetch from Redis and cache
		flag1, err := ffs.GetFlag(ctx, flagName)
		if err != nil {
			t.Fatalf("GetFlag() error = %v", err)
		}

		// Second call - should use cache
		flag2, err := ffs.GetFlag(ctx, flagName)
		if err != nil {
			t.Fatalf("GetFlag() error = %v", err)
		}

		if flag1.Name != flag2.Name {
			t.Errorf("Cache not working correctly")
		}
	})

	t.Run("NonExistentFlag", func(t *testing.T) {
		enabled, err := ffs.IsEnabled(ctx, "non-existent-flag")
		if err != nil {
			t.Fatalf("IsEnabled() error = %v", err)
		}
		if enabled {
			t.Errorf("IsEnabled() for non-existent flag = %v, want %v", enabled, false)
		}
	})
}

func TestDefaultFeatureFlagSetConfig(t *testing.T) {
	config := DefaultFeatureFlagSetConfig()

	if config.RedisAddr != "localhost:6379" {
		t.Errorf("RedisAddr = %v, want %v", config.RedisAddr, "localhost:6379")
	}
	if config.RedisDB != 0 {
		t.Errorf("RedisDB = %v, want %v", config.RedisDB, 0)
	}
	if config.KeyPrefix != "featureflag:" {
		t.Errorf("KeyPrefix = %v, want %v", config.KeyPrefix, "featureflag:")
	}
	if config.CacheTTL != 30*time.Second {
		t.Errorf("CacheTTL = %v, want %v", config.CacheTTL, 30*time.Second)
	}
	if config.Environment != EnvDevelopment {
		t.Errorf("Environment = %v, want %v", config.Environment, EnvDevelopment)
	}
}

func TestEnvironmentConstants(t *testing.T) {
	if EnvProduction != "prod" {
		t.Errorf("EnvProduction = %v, want %v", EnvProduction, "prod")
	}
	if EnvStaging != "stage" {
		t.Errorf("EnvStaging = %v, want %v", EnvStaging, "stage")
	}
	if EnvDevelopment != "dev" {
		t.Errorf("EnvDevelopment = %v, want %v", EnvDevelopment, "dev")
	}
}

func TestRolloutPolicyConstants(t *testing.T) {
	if RolloutAllAtOnce != "all_at_once" {
		t.Errorf("RolloutAllAtOnce = %v, want %v", RolloutAllAtOnce, "all_at_once")
	}
	if RolloutGradual != "gradual" {
		t.Errorf("RolloutGradual = %v, want %v", RolloutGradual, "gradual")
	}
	if RolloutCanary != "canary" {
		t.Errorf("RolloutCanary = %v, want %v", RolloutCanary, "canary")
	}
	if RolloutTargeted != "targeted" {
		t.Errorf("RolloutTargeted = %v, want %v", RolloutTargeted, "targeted")
	}
}

func TestFeatureFlag_IsEnabledForUser_AllAtOnce(t *testing.T) {
	flag := &FeatureFlag{
		Name:    "test-flag",
		Enabled: true,
		Rollout: &RolloutConfig{
			Policy: RolloutAllAtOnce,
		},
	}

	// All users should be enabled
	users := []string{"user1", "user2", "user3", "user4", "user5"}
	for _, userID := range users {
		if !flag.IsEnabledForUser(EnvProduction, userID, nil) {
			t.Errorf("Expected flag to be enabled for user %s with all-at-once rollout", userID)
		}
	}
}

func TestFeatureFlag_IsEnabledForUser_Gradual(t *testing.T) {
	tests := []struct {
		name       string
		percentage int
		minEnabled int // Minimum number of enabled users out of 100
		maxEnabled int // Maximum number of enabled users out of 100
	}{
		{"0 percent", 0, 0, 0},
		{"25 percent", 25, 15, 35},
		{"50 percent", 50, 40, 60},
		{"75 percent", 75, 65, 85},
		{"100 percent", 100, 100, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := &FeatureFlag{
				Name:    "test-flag",
				Enabled: true,
				Rollout: &RolloutConfig{
					Policy:     RolloutGradual,
					Percentage: tt.percentage,
				},
			}

			enabledCount := 0
			for i := 0; i < 100; i++ {
				userID := fmt.Sprintf("user%d", i)
				if flag.IsEnabledForUser(EnvProduction, userID, nil) {
					enabledCount++
				}
			}

			if enabledCount < tt.minEnabled || enabledCount > tt.maxEnabled {
				t.Errorf("Expected between %d and %d enabled users, got %d", tt.minEnabled, tt.maxEnabled, enabledCount)
			}
		})
	}
}

func TestFeatureFlag_IsEnabledForUser_GradualConsistency(t *testing.T) {
	flag := &FeatureFlag{
		Name:    "test-flag",
		Enabled: true,
		Rollout: &RolloutConfig{
			Policy:     RolloutGradual,
			Percentage: 50,
		},
	}

	// Test consistency - same user should always get same result
	userID := "consistent-user"
	firstResult := flag.IsEnabledForUser(EnvProduction, userID, nil)

	for i := 0; i < 10; i++ {
		result := flag.IsEnabledForUser(EnvProduction, userID, nil)
		if result != firstResult {
			t.Errorf("Inconsistent result for user %s on attempt %d", userID, i)
		}
	}
}

func TestFeatureFlag_IsEnabledForUser_Canary(t *testing.T) {
	flag := &FeatureFlag{
		Name:    "test-flag",
		Enabled: true,
		Rollout: &RolloutConfig{
			Policy:         RolloutCanary,
			CanarySegments: []string{"beta", "internal"},
		},
	}

	tests := []struct {
		name         string
		userSegments []string
		expected     bool
	}{
		{"User in beta segment", []string{"beta"}, true},
		{"User in internal segment", []string{"internal"}, true},
		{"User in both segments", []string{"beta", "internal"}, true},
		{"User in beta and other", []string{"beta", "premium"}, true},
		{"User not in any canary segment", []string{"premium"}, false},
		{"User with no segments", []string{}, false},
		{"User with nil segments", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := flag.IsEnabledForUser(EnvProduction, "user123", tt.userSegments)
			if result != tt.expected {
				t.Errorf("Expected %v for segments %v, got %v", tt.expected, tt.userSegments, result)
			}
		})
	}
}

func TestFeatureFlag_IsEnabledForUser_Targeted(t *testing.T) {
	flag := &FeatureFlag{
		Name:    "test-flag",
		Enabled: true,
		Rollout: &RolloutConfig{
			Policy:        RolloutTargeted,
			TargetUserIDs: []string{"user1", "user5", "user10"},
		},
	}

	tests := []struct {
		name     string
		userID   string
		expected bool
	}{
		{"Targeted user 1", "user1", true},
		{"Targeted user 5", "user5", true},
		{"Targeted user 10", "user10", true},
		{"Non-targeted user 2", "user2", false},
		{"Non-targeted user 100", "user100", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := flag.IsEnabledForUser(EnvProduction, tt.userID, nil)
			if result != tt.expected {
				t.Errorf("Expected %v for user %s, got %v", tt.expected, tt.userID, result)
			}
		})
	}
}

func TestFeatureFlag_IsEnabledForUser_DisabledEnvironment(t *testing.T) {
	flag := &FeatureFlag{
		Name:    "test-flag",
		Enabled: false, // Disabled globally
		Rollout: &RolloutConfig{
			Policy:     RolloutGradual,
			Percentage: 100, // Even 100% rollout
		},
	}

	// Should return false even with rollout policy
	result := flag.IsEnabledForUser(EnvProduction, "user123", nil)
	if result {
		t.Errorf("Expected flag to be disabled when environment is disabled, regardless of rollout policy")
	}
}

func TestFeatureFlag_IsEnabledForUser_NoRolloutPolicy(t *testing.T) {
	flag := &FeatureFlag{
		Name:    "test-flag",
		Enabled: true,
		Rollout: nil, // No rollout policy
	}

	// Should return true for all users when no rollout policy
	users := []string{"user1", "user2", "user3"}
	for _, userID := range users {
		if !flag.IsEnabledForUser(EnvProduction, userID, nil) {
			t.Errorf("Expected flag to be enabled for user %s with no rollout policy", userID)
		}
	}
}

func TestFeatureFlagSet_RolloutPolicies_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	config := &FeatureFlagSetConfig{
		RedisAddr:   "localhost:6379",
		KeyPrefix:   "test:rollout:",
		CacheTTL:    5 * time.Second,
		Environment: EnvProduction,
	}

	ffs, err := NewFeatureFlagSet(config)
	if err != nil {
		t.Skipf("Skipping test: Redis not available: %v", err)
		return
	}
	defer ffs.Close()

	// Clean up test flags
	defer func() {
		flags, _ := ffs.ListFlags(ctx)
		for _, flag := range flags {
			ffs.DeleteFlag(ctx, flag.Name)
		}
	}()

	t.Run("GradualRollout", func(t *testing.T) {
		flagName := "test-gradual-rollout"
		err := ffs.CreateFlag(ctx, flagName, true, "Test gradual rollout")
		if err != nil {
			t.Fatalf("CreateFlag() error = %v", err)
		}
		defer ffs.DeleteFlag(ctx, flagName)

		// Set 50% rollout
		err = ffs.SetGradualRollout(ctx, flagName, 50)
		if err != nil {
			t.Fatalf("SetGradualRollout() error = %v", err)
		}

		// Check that some users are enabled and some are not
		enabledCount := 0
		for i := 0; i < 100; i++ {
			userID := fmt.Sprintf("user%d", i)
			enabled, _ := ffs.IsEnabledForUser(ctx, flagName, userID, nil)
			if enabled {
				enabledCount++
			}
		}

		// Should be roughly 50%, allow 20-80 range
		if enabledCount < 20 || enabledCount > 80 {
			t.Errorf("Expected roughly 50%% enabled, got %d%%", enabledCount)
		}
	})

	t.Run("CanaryRollout", func(t *testing.T) {
		flagName := "test-canary-rollout"
		err := ffs.CreateFlag(ctx, flagName, true, "Test canary rollout")
		if err != nil {
			t.Fatalf("CreateFlag() error = %v", err)
		}
		defer ffs.DeleteFlag(ctx, flagName)

		// Set canary rollout
		err = ffs.SetCanaryRollout(ctx, flagName, []string{"beta", "internal"})
		if err != nil {
			t.Fatalf("SetCanaryRollout() error = %v", err)
		}

		// Test with beta user
		enabled, _ := ffs.IsEnabledForUser(ctx, flagName, "user1", []string{"beta"})
		if !enabled {
			t.Errorf("Expected flag to be enabled for beta user")
		}

		// Test with non-beta user
		enabled, _ = ffs.IsEnabledForUser(ctx, flagName, "user2", []string{"premium"})
		if enabled {
			t.Errorf("Expected flag to be disabled for non-beta user")
		}
	})

	t.Run("TargetedRollout", func(t *testing.T) {
		flagName := "test-targeted-rollout"
		err := ffs.CreateFlag(ctx, flagName, true, "Test targeted rollout")
		if err != nil {
			t.Fatalf("CreateFlag() error = %v", err)
		}
		defer ffs.DeleteFlag(ctx, flagName)

		// Set targeted rollout
		targetUsers := []string{"alice", "bob", "charlie"}
		err = ffs.SetTargetedRollout(ctx, flagName, targetUsers)
		if err != nil {
			t.Fatalf("SetTargetedRollout() error = %v", err)
		}

		// Test with targeted user
		enabled, _ := ffs.IsEnabledForUser(ctx, flagName, "alice", nil)
		if !enabled {
			t.Errorf("Expected flag to be enabled for targeted user alice")
		}

		// Test with non-targeted user
		enabled, _ = ffs.IsEnabledForUser(ctx, flagName, "david", nil)
		if enabled {
			t.Errorf("Expected flag to be disabled for non-targeted user david")
		}
	})

	t.Run("AllAtOnceRollout", func(t *testing.T) {
		flagName := "test-all-at-once-rollout"
		err := ffs.CreateFlag(ctx, flagName, true, "Test all-at-once rollout")
		if err != nil {
			t.Fatalf("CreateFlag() error = %v", err)
		}
		defer ffs.DeleteFlag(ctx, flagName)

		// Set all-at-once rollout
		err = ffs.SetAllAtOnceRollout(ctx, flagName)
		if err != nil {
			t.Fatalf("SetAllAtOnceRollout() error = %v", err)
		}

		// All users should be enabled
		for i := 0; i < 10; i++ {
			userID := fmt.Sprintf("user%d", i)
			enabled, _ := ffs.IsEnabledForUser(ctx, flagName, userID, nil)
			if !enabled {
				t.Errorf("Expected flag to be enabled for all users with all-at-once rollout")
			}
		}
	})
}
