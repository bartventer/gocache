package redis

import (
	"context"
	"net/url"
	"testing"
	"time"

	cache "github.com/bartventer/gocache"
	"github.com/bartventer/gocache/drivertest"
	"github.com/docker/docker/api/types/container"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Defines the default Redis network address.
const (
	defaultPort = "6379"
	defaultAddr = "localhost:" + defaultPort
)

func TestRedisCache_OpenCacheURL(t *testing.T) {
	t.Parallel()
	r := &redisCache{}

	u, err := url.Parse("redis://" + defaultAddr + "?maxretries=5&minretrybackoff=1000ms")
	require.NoError(t, err)

	_, err = r.OpenCacheURL(context.Background(), u, &cache.Options{})
	require.NoError(t, err)
	assert.NotNil(t, r.client)
}

func TestRedisCache_New(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	config := cache.Config{}
	options := redis.Options{
		Addr: defaultAddr,
	}

	r := New(ctx, &config, options)
	require.NotNil(t, r)
	assert.NotNil(t, r.client)
}

// setupCache creates a new Redis cache with a test container.
func setupCache(t *testing.T) *redisCache {
	t.Helper()
	// Create a new Redis container
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "redis:alpine",
		ExposedPorts: []string{defaultPort},
		ConfigModifier: func(c *container.Config) {
			c.Healthcheck = &container.HealthConfig{
				Test:          []string{"CMD", "redis-cli", "ping"},
				Interval:      30 * time.Second,
				Timeout:       60 * time.Second,
				Retries:       5,
				StartPeriod:   20 * time.Second,
				StartInterval: 5 * time.Second,
			}
		},
		WaitingFor: wait.ForHealthCheck(),
	}
	redisC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("Failed to start Redis container: %v", err)
	}
	t.Cleanup(func() {
		if cleanupErr := redisC.Terminate(ctx); cleanupErr != nil {
			t.Fatalf("Failed to terminate Redis container: %v", cleanupErr)
		}
	})
	// Get the Redis container endpoint
	endpoint, err := redisC.Endpoint(ctx, "")
	if err != nil {
		t.Fatalf("Failed to get Redis container endpoint: %v", err)
	}
	// Create a new Redis cache
	client := redis.NewClient(&redis.Options{
		Addr:            endpoint,
		MaxRetries:      5,
		MinRetryBackoff: 1000 * time.Millisecond,
	})
	t.Cleanup(func() {
		client.Close()
	})
	err = client.Ping(context.Background()).Err()
	if err != nil {
		t.Fatalf("Failed to ping Redis container: %v", err)
	}
	return &redisCache{client: client, config: &cache.Config{CountLimit: 100}}
}

type harness struct {
	cache *redisCache
}

func (h *harness) MakeCache(ctx context.Context) (cache.Cache, error) {
	return h.cache, nil
}

func (h *harness) Close() {
	// Cleanup is handled in setup function
}

func (h *harness) Options() drivertest.Options {
	return drivertest.Options{
		PatternMatchingDisabled: false,
		CloseIsNoop:             false,
	}
}

func newHarness(ctx context.Context, t *testing.T) (drivertest.Harness, error) {
	cache := setupCache(t)
	return &harness{
		cache: cache,
	}, nil
}

func TestConformance(t *testing.T) {
	drivertest.RunConformanceTests(t, newHarness)
}
