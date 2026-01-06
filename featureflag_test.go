package goroutine

import (
	"context"
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
