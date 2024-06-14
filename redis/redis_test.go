package redis

import (
	"context"
	"net/url"
	"testing"
	"time"

	cache "github.com/bartventer/gocache"
	"github.com/bartventer/gocache/internal/testutil"
	"github.com/bartventer/gocache/keymod"
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

// setupRedis creates a new Redis cache with a test container.
func setupRedis(t *testing.T) *redisCache {
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
		Addr: endpoint,
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

func TestRedisCache_Count(t *testing.T) {
	t.Parallel()
	c := setupRedis(t)

	key := testutil.UniqueKey(t)
	value := "testValue1"

	if err := c.Set(context.Background(), key, value); err != nil {
		t.Fatalf("Failed to set key: %v", err)
	}
	t.Cleanup(func() {
		c.client.Del(context.Background(), key)
	})

	count, err := c.Count(context.Background(), "*")
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestRedisCache_Exists(t *testing.T) {
	t.Parallel()
	c := setupRedis(t)
	key := testutil.UniqueKey(t)
	value := "testValue"

	if err := c.Set(context.Background(), key, value); err != nil {
		t.Fatalf("Failed to set key: %v", err)
	}
	t.Cleanup(func() {
		c.client.Del(context.Background(), key)
	})

	exists, err := c.Exists(context.Background(), key)
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestRedisCache_Del(t *testing.T) {
	t.Parallel()
	c := setupRedis(t)

	key := testutil.UniqueKey(t)
	value := "testValue"

	if err := c.Set(context.Background(), key, value); err != nil {
		t.Fatalf("Failed to set key: %v", err)
	}

	err := c.Del(context.Background(), key)
	require.NoError(t, err)

	exists, err := c.Exists(context.Background(), key)
	require.NoError(t, err)
	assert.False(t, exists)

	// Non-existent key
	err = c.Del(context.Background(), "nonExistentKey")
	require.Error(t, err)
	assert.Contains(t, err.Error(), cache.ErrKeyNotFound.Error())
}

func TestRedisCache_DelKeys(t *testing.T) {
	t.Parallel()
	c := setupRedis(t)

	keys := []string{"testKey1", "testKey2", "testKey3"}
	hashTag := testutil.UniqueKey(t)
	for _, key := range keys {
		if err := c.Set(context.Background(), key, "testValue", keymod.HashTagModifier(hashTag)); err != nil {
			t.Fatalf("Failed to set key: %v", err)
		}
	}

	err := c.DelKeys(context.Background(), "testKey*", keymod.HashTagModifier(hashTag))
	require.NoError(t, err)

	for _, key := range keys {
		exists, existsErr := c.Exists(context.Background(), key, keymod.HashTagModifier(hashTag))
		require.NoError(t, existsErr)
		assert.False(t, exists)
	}

	// Non-existent key
	err = c.DelKeys(context.Background(), "nonExistentKey*", keymod.HashTagModifier(hashTag))
	require.NoError(t, err)
}

func TestRedisCache_Clear(t *testing.T) {
	t.Parallel()
	c := setupRedis(t)

	key := testutil.UniqueKey(t)
	value := "testValue"

	if err := c.Set(context.Background(), key, value); err != nil {
		t.Fatalf("Failed to set key: %v", err)
	}

	err := c.Clear(context.Background())
	require.NoError(t, err)

	exists, err := c.Exists(context.Background(), key)
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestRedisCache_Get(t *testing.T) {
	t.Parallel()
	c := setupRedis(t)

	key := testutil.UniqueKey(t)
	value := "testValue"

	if err := c.client.Set(context.Background(), key, value, 0).Err(); err != nil {
		t.Fatalf("Failed to set key: %v", err)
	}
	t.Cleanup(func() {
		c.client.Del(context.Background(), key)
	})

	got, err := c.Get(context.Background(), key)
	require.NoError(t, err)
	assert.Equal(t, value, string(got))

	// Non-existent key
	_, err = c.Get(context.Background(), "nonExistentKey")
	require.Error(t, err)
	assert.Contains(t, err.Error(), cache.ErrKeyNotFound.Error())
}

func TestRedisCache_Set(t *testing.T) {
	t.Parallel()
	c := setupRedis(t)

	key := testutil.UniqueKey(t)
	value := "testValue"

	err := c.Set(context.Background(), key, value)
	require.NoError(t, err)
	t.Cleanup(func() {
		c.client.Del(context.Background(), key)
	})

	got, err := c.Get(context.Background(), key)
	require.NoError(t, err)
	assert.Equal(t, value, string(got))
}

func TestRedisCache_SetWithExpiry(t *testing.T) {
	t.Parallel()
	c := setupRedis(t)

	key := testutil.UniqueKey(t)
	value := "testValue"
	expiry := 1 * time.Second

	err := c.SetWithExpiry(context.Background(), key, value, expiry)
	require.NoError(t, err)
	t.Cleanup(func() {
		c.client.Del(context.Background(), key)
	})

	got, err := c.Get(context.Background(), key)
	require.NoError(t, err)
	assert.Equal(t, value, string(got))

	// Wait for the key to expire
	time.Sleep(expiry)

	_, err = c.Get(context.Background(), key)
	require.Error(t, err)
	assert.Contains(t, err.Error(), cache.ErrKeyNotFound.Error())
}

func TestRedisCache_Ping(t *testing.T) {
	t.Parallel()
	c := setupRedis(t)

	err := c.Ping(context.Background())
	require.NoError(t, err)
}

func TestRedisCache_Close(t *testing.T) {
	t.Parallel()
	c := setupRedis(t)

	err := c.Close()
	require.NoError(t, err)

	// After closing, pinging should result in an error
	err = c.Ping(context.Background())
	require.Error(t, err)
}
