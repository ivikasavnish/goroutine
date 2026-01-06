package goroutine

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// Environment represents the deployment environment
type Environment string

const (
	// EnvProduction represents production environment
	EnvProduction Environment = "prod"
	// EnvStaging represents staging environment
	EnvStaging Environment = "stage"
	// EnvDevelopment represents development environment
	EnvDevelopment Environment = "dev"
)

// RolloutPolicy defines how a feature is rolled out
type RolloutPolicy string

const (
	// RolloutAllAtOnce enables the feature for all users immediately
	RolloutAllAtOnce RolloutPolicy = "all_at_once"
	// RolloutGradual enables the feature gradually using percentage
	RolloutGradual RolloutPolicy = "gradual"
	// RolloutCanary enables the feature for specific user segments first
	RolloutCanary RolloutPolicy = "canary"
	// RolloutTargeted enables the feature for specific user IDs only
	RolloutTargeted RolloutPolicy = "targeted"
)

// RolloutConfig contains rollout policy configuration
type RolloutConfig struct {
	// Policy is the rollout strategy to use
	Policy RolloutPolicy `json:"policy"`
	// Percentage is the percentage of users to enable (0-100) for gradual rollout
	Percentage int `json:"percentage,omitempty"`
	// TargetUserIDs is the list of specific user IDs for targeted rollout
	TargetUserIDs []string `json:"target_user_ids,omitempty"`
	// CanarySegments is the list of segments for canary rollout (e.g., "beta_users", "internal")
	CanarySegments []string `json:"canary_segments,omitempty"`
}

// FeatureFlag represents a single feature flag with environment-specific settings
type FeatureFlag struct {
	// Name is the unique identifier for the feature flag
	Name string `json:"name"`
	// Description describes what this flag controls
	Description string `json:"description,omitempty"`
	// Enabled is the global on/off switch for all environments
	Enabled bool `json:"enabled"`
	// Environments contains environment-specific overrides
	Environments map[Environment]bool `json:"environments,omitempty"`
	// Rollout contains rollout policy configuration
	Rollout *RolloutConfig `json:"rollout,omitempty"`
	// CreatedAt is when the flag was created
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt is when the flag was last updated
	UpdatedAt time.Time `json:"updated_at"`
}

// IsEnabledForEnv checks if the flag is enabled for a specific environment
func (f *FeatureFlag) IsEnabledForEnv(env Environment) bool {
	// Check environment-specific override first
	if enabled, exists := f.Environments[env]; exists {
		return enabled
	}
	// Fall back to global enabled status
	return f.Enabled
}

// IsEnabledForUser checks if the flag is enabled for a specific user with rollout policy
func (f *FeatureFlag) IsEnabledForUser(env Environment, userID string, userSegments []string) bool {
	// First check if flag is enabled for the environment
	if !f.IsEnabledForEnv(env) {
		return false
	}

	// If no rollout policy, return the environment setting
	if f.Rollout == nil {
		return true
	}

	// Apply rollout policy
	switch f.Rollout.Policy {
	case RolloutAllAtOnce:
		return true

	case RolloutGradual:
		// Use consistent hashing for percentage-based rollout
		return f.isUserInPercentage(userID, f.Rollout.Percentage)

	case RolloutCanary:
		// Check if user is in any of the canary segments
		return f.isUserInSegments(userSegments, f.Rollout.CanarySegments)

	case RolloutTargeted:
		// Check if user ID is in the target list
		return f.isUserTargeted(userID, f.Rollout.TargetUserIDs)

	default:
		return true
	}
}

// isUserInPercentage checks if a user falls within the rollout percentage using consistent hashing
func (f *FeatureFlag) isUserInPercentage(userID string, percentage int) bool {
	if percentage <= 0 {
		return false
	}
	if percentage >= 100 {
		return true
	}

	// Simple hash function for consistent percentage-based rollout
	hash := 0
	key := f.Name + ":" + userID
	for i := 0; i < len(key); i++ {
		hash = (hash*31 + int(key[i])) % 100
	}
	if hash < 0 {
		hash = -hash
	}
	return hash < percentage
}

// isUserInSegments checks if user has any of the target segments
func (f *FeatureFlag) isUserInSegments(userSegments []string, targetSegments []string) bool {
	if len(targetSegments) == 0 {
		return false
	}

	segmentMap := make(map[string]bool)
	for _, seg := range userSegments {
		segmentMap[seg] = true
	}

	for _, target := range targetSegments {
		if segmentMap[target] {
			return true
		}
	}
	return false
}

// isUserTargeted checks if user ID is in the target list
func (f *FeatureFlag) isUserTargeted(userID string, targetUserIDs []string) bool {
	if len(targetUserIDs) == 0 {
		return false
	}

	for _, targetID := range targetUserIDs {
		if userID == targetID {
			return true
		}
	}
	return false
}

// FeatureFlagSet manages a collection of feature flags with Redis backend
type FeatureFlagSet struct {
	client      *redis.Client
	keyPrefix   string
	localCache  *Cache[string, FeatureFlag]
	cacheTTL    time.Duration
	mu          sync.RWMutex
	environment Environment
}

// FeatureFlagSetConfig configures the feature flag set
type FeatureFlagSetConfig struct {
	// RedisAddr is the Redis server address (default: "localhost:6379")
	RedisAddr string
	// RedisPassword is the Redis password (optional)
	RedisPassword string
	// RedisDB is the Redis database number (default: 0)
	RedisDB int
	// KeyPrefix is the prefix for all Redis keys (default: "featureflag:")
	KeyPrefix string
	// CacheTTL is the local cache TTL (default: 30 seconds)
	CacheTTL time.Duration
	// Environment is the current environment (default: dev)
	Environment Environment
}

// DefaultFeatureFlagSetConfig returns default configuration
func DefaultFeatureFlagSetConfig() *FeatureFlagSetConfig {
	return &FeatureFlagSetConfig{
		RedisAddr:   "localhost:6379",
		RedisDB:     0,
		KeyPrefix:   "featureflag:",
		CacheTTL:    30 * time.Second,
		Environment: EnvDevelopment,
	}
}

// NewFeatureFlagSet creates a new feature flag set with Redis backend
func NewFeatureFlagSet(config *FeatureFlagSetConfig) (*FeatureFlagSet, error) {
	if config == nil {
		config = DefaultFeatureFlagSetConfig()
	}

	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Password: config.RedisPassword,
		DB:       config.RedisDB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &FeatureFlagSet{
		client:      client,
		keyPrefix:   config.KeyPrefix,
		localCache:  NewCache[string, FeatureFlag](),
		cacheTTL:    config.CacheTTL,
		environment: config.Environment,
	}, nil
}

// NewFeatureFlagSetSimple creates a feature flag set with simple parameters
func NewFeatureFlagSetSimple(redisAddr string, env Environment) (*FeatureFlagSet, error) {
	config := &FeatureFlagSetConfig{
		RedisAddr:   redisAddr,
		KeyPrefix:   "featureflag:",
		CacheTTL:    30 * time.Second,
		Environment: env,
	}
	return NewFeatureFlagSet(config)
}

// getRedisKey returns the Redis key for a flag name
func (ffs *FeatureFlagSet) getRedisKey(flagName string) string {
	return ffs.keyPrefix + flagName
}

// IsEnabled checks if a feature flag is enabled for the current environment
func (ffs *FeatureFlagSet) IsEnabled(ctx context.Context, flagName string) (bool, error) {
	return ffs.IsEnabledForEnv(ctx, flagName, ffs.environment)
}

// IsEnabledForEnv checks if a feature flag is enabled for a specific environment
func (ffs *FeatureFlagSet) IsEnabledForEnv(ctx context.Context, flagName string, env Environment) (bool, error) {
	flag, err := ffs.GetFlag(ctx, flagName)
	if err != nil {
		// Check if it's a "not found" error - return false for missing flags (safe default)
		// For other errors, we also return false but could log for debugging
		return false, nil
	}
	return flag.IsEnabledForEnv(env), nil
}

// IsEnabledForUser checks if a feature flag is enabled for a specific user with rollout policy
func (ffs *FeatureFlagSet) IsEnabledForUser(ctx context.Context, flagName string, userID string, userSegments []string) (bool, error) {
	flag, err := ffs.GetFlag(ctx, flagName)
	if err != nil {
		// Return false for missing flags (safe default)
		return false, nil
	}
	return flag.IsEnabledForUser(ffs.environment, userID, userSegments), nil
}

// IsEnabledForUserInEnv checks if a feature flag is enabled for a user in a specific environment
func (ffs *FeatureFlagSet) IsEnabledForUserInEnv(ctx context.Context, flagName string, env Environment, userID string, userSegments []string) (bool, error) {
	flag, err := ffs.GetFlag(ctx, flagName)
	if err != nil {
		// Return false for missing flags (safe default)
		return false, nil
	}
	return flag.IsEnabledForUser(env, userID, userSegments), nil
}

// GetFlag retrieves a feature flag by name
func (ffs *FeatureFlagSet) GetFlag(ctx context.Context, flagName string) (*FeatureFlag, error) {
	// Check local cache first
	if entry, exists := ffs.localCache.Get(flagName); exists && !entry.IsExpired() {
		flag := entry.Value
		return &flag, nil
	}

	// Fetch from Redis
	key := ffs.getRedisKey(flagName)
	data, err := ffs.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("feature flag not found: %s", flagName)
		}
		return nil, fmt.Errorf("failed to get feature flag: %w", err)
	}

	var flag FeatureFlag
	if err := json.Unmarshal(data, &flag); err != nil {
		return nil, fmt.Errorf("failed to unmarshal feature flag: %w", err)
	}

	// Update local cache
	ffs.localCache.Set(flagName, flag, ffs.cacheTTL)

	return &flag, nil
}

// SetFlag creates or updates a feature flag
func (ffs *FeatureFlagSet) SetFlag(ctx context.Context, flag *FeatureFlag) error {
	ffs.mu.Lock()
	defer ffs.mu.Unlock()

	now := time.Now()
	if flag.CreatedAt.IsZero() {
		flag.CreatedAt = now
	}
	flag.UpdatedAt = now

	if flag.Environments == nil {
		flag.Environments = make(map[Environment]bool)
	}

	data, err := json.Marshal(flag)
	if err != nil {
		return fmt.Errorf("failed to marshal feature flag: %w", err)
	}

	key := ffs.getRedisKey(flag.Name)
	if err := ffs.client.Set(ctx, key, data, 0).Err(); err != nil {
		return fmt.Errorf("failed to set feature flag: %w", err)
	}

	// Update local cache
	ffs.localCache.Set(flag.Name, *flag, ffs.cacheTTL)

	return nil
}

// CreateFlag creates a new feature flag with default settings
func (ffs *FeatureFlagSet) CreateFlag(ctx context.Context, name string, enabled bool, description string) error {
	flag := &FeatureFlag{
		Name:         name,
		Description:  description,
		Enabled:      enabled,
		Environments: make(map[Environment]bool),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	return ffs.SetFlag(ctx, flag)
}

// UpdateFlag updates an existing feature flag's enabled status
func (ffs *FeatureFlagSet) UpdateFlag(ctx context.Context, name string, enabled bool) error {
	flag, err := ffs.GetFlag(ctx, name)
	if err != nil {
		return err
	}
	flag.Enabled = enabled
	flag.UpdatedAt = time.Now()
	return ffs.SetFlag(ctx, flag)
}

// SetFlagForEnv sets the enabled status for a specific environment
func (ffs *FeatureFlagSet) SetFlagForEnv(ctx context.Context, name string, env Environment, enabled bool) error {
	flag, err := ffs.GetFlag(ctx, name)
	if err != nil {
		return err
	}
	if flag.Environments == nil {
		flag.Environments = make(map[Environment]bool)
	}
	flag.Environments[env] = enabled
	flag.UpdatedAt = time.Now()
	return ffs.SetFlag(ctx, flag)
}

// SetRolloutPolicy sets the rollout policy for a feature flag
func (ffs *FeatureFlagSet) SetRolloutPolicy(ctx context.Context, name string, rollout *RolloutConfig) error {
	flag, err := ffs.GetFlag(ctx, name)
	if err != nil {
		return err
	}
	flag.Rollout = rollout
	flag.UpdatedAt = time.Now()
	return ffs.SetFlag(ctx, flag)
}

// SetGradualRollout configures a gradual rollout with percentage
func (ffs *FeatureFlagSet) SetGradualRollout(ctx context.Context, name string, percentage int) error {
	if percentage < 0 || percentage > 100 {
		return fmt.Errorf("percentage must be between 0 and 100")
	}
	return ffs.SetRolloutPolicy(ctx, name, &RolloutConfig{
		Policy:     RolloutGradual,
		Percentage: percentage,
	})
}

// SetCanaryRollout configures a canary rollout with segments
func (ffs *FeatureFlagSet) SetCanaryRollout(ctx context.Context, name string, segments []string) error {
	return ffs.SetRolloutPolicy(ctx, name, &RolloutConfig{
		Policy:         RolloutCanary,
		CanarySegments: segments,
	})
}

// SetTargetedRollout configures a targeted rollout with specific user IDs
func (ffs *FeatureFlagSet) SetTargetedRollout(ctx context.Context, name string, userIDs []string) error {
	return ffs.SetRolloutPolicy(ctx, name, &RolloutConfig{
		Policy:        RolloutTargeted,
		TargetUserIDs: userIDs,
	})
}

// SetAllAtOnceRollout configures an all-at-once rollout
func (ffs *FeatureFlagSet) SetAllAtOnceRollout(ctx context.Context, name string) error {
	return ffs.SetRolloutPolicy(ctx, name, &RolloutConfig{
		Policy: RolloutAllAtOnce,
	})
}

// DeleteFlag removes a feature flag
func (ffs *FeatureFlagSet) DeleteFlag(ctx context.Context, flagName string) error {
	ffs.mu.Lock()
	defer ffs.mu.Unlock()

	key := ffs.getRedisKey(flagName)
	if err := ffs.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete feature flag: %w", err)
	}

	// Remove from local cache
	ffs.localCache.Delete(flagName)

	return nil
}

// ListFlags returns all feature flags
func (ffs *FeatureFlagSet) ListFlags(ctx context.Context) ([]*FeatureFlag, error) {
	pattern := ffs.keyPrefix + "*"
	keys, err := ffs.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to list feature flags: %w", err)
	}

	flags := make([]*FeatureFlag, 0, len(keys))
	for _, key := range keys {
		data, err := ffs.client.Get(ctx, key).Bytes()
		if err != nil {
			continue // Skip on error
		}

		var flag FeatureFlag
		if err := json.Unmarshal(data, &flag); err != nil {
			continue // Skip on error
		}

		flags = append(flags, &flag)
	}

	return flags, nil
}

// ClearCache clears the local cache
func (ffs *FeatureFlagSet) ClearCache() {
	ffs.localCache.Clear()
}

// Close closes the Redis connection
func (ffs *FeatureFlagSet) Close() error {
	return ffs.client.Close()
}

// GetEnvironment returns the current environment
func (ffs *FeatureFlagSet) GetEnvironment() Environment {
	return ffs.environment
}

// SetEnvironment updates the current environment
func (ffs *FeatureFlagSet) SetEnvironment(env Environment) {
	ffs.mu.Lock()
	defer ffs.mu.Unlock()
	ffs.environment = env
}
