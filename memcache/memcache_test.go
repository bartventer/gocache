package memcache

import (
	"context"
	"net/url"
	"strings"
	"testing"
	"time"

	cache "github.com/bartventer/gocache"
	"github.com/bartventer/gocache/internal/testutil"
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Defines the default Memcached network address.
const (
	defaultPort = "11211"
	defaultAddr = "localhost:" + defaultPort
)

// malformedKey is a key that is too long which will trigger the [memcache.ErrMalformedKey] error.
var malformedKey = strings.Repeat("a", 251)

// setupMemcached creates a new Memcached container.
func setupMemcached(t *testing.T) *memcacheCache {
	t.Helper()
	// Create a new Memcached container
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "memcached:alpine",
		ExposedPorts: []string{defaultPort},
		ConfigModifier: func(c *container.Config) {
			c.Healthcheck = &container.HealthConfig{
				Test:          []string{"CMD", "nc", "-vn", "-w", "1", "localhost", defaultPort},
				Interval:      30 * time.Second,
				Timeout:       60 * time.Second,
				Retries:       5,
				StartPeriod:   20 * time.Second,
				StartInterval: 5 * time.Second,
			}
		},
		WaitingFor: wait.ForHealthCheck(),
	}
	memcachedC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("Failed to start Memcached container: %v", err)
	}
	t.Cleanup(func() {
		if cleanupErr := memcachedC.Terminate(ctx); cleanupErr != nil {
			t.Fatalf("Failed to terminate Memcached container: %v", cleanupErr)
		}
	})
	// Get the Memcached container endpoint
	endpoint, err := memcachedC.Endpoint(ctx, "")
	if err != nil {
		t.Fatalf("Failed to get Memcached container endpoint: %v", err)
	}
	// Create a new Memcached client
	client := memcache.New(endpoint)
	t.Cleanup(func() {
		client.Close()
	})
	err = client.Ping()
	if err != nil {
		t.Fatalf("Failed to ping Memcached container: %v", err)
	}
	return &memcacheCache{client: client, config: &cache.Config{CountLimit: 100}}
}

func TestMemcacheCache_OpenCacheURL(t *testing.T) {
	t.Parallel()
	m := &memcacheCache{}

	u, err := url.Parse("memcache://" + defaultAddr)
	require.NoError(t, err)

	_, err = m.OpenCacheURL(context.Background(), u, &cache.Options{})
	require.NoError(t, err)
	assert.NotNil(t, m.client)
}

func TestMemcacheCache_New(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	config := cache.Config{}

	m := New(ctx, &config, defaultAddr)
	require.NotNil(t, m)
	assert.NotNil(t, m.client)
}

func TestMemcacheCache_Exists(t *testing.T) {
	t.Parallel()
	c := setupMemcached(t)
	key := testutil.UniqueKey(t)
	value := "testValue"

	if err := c.Set(context.Background(), key, value); err != nil {
		t.Fatalf("Failed to set key: %v", err)
	}
	t.Cleanup(func() {
		c.client.Delete(key)
	})

	exists, err := c.Exists(context.Background(), key)
	require.NoError(t, err)
	assert.True(t, exists)

	// Non-existent key
	exists, err = c.Exists(context.Background(), "nonExistentKey")
	require.NoError(t, err)
	assert.False(t, exists)

	// Malformed key
	_, err = c.Exists(context.Background(), malformedKey)
	require.Error(t, err)
	assert.Contains(t, err.Error(), memcache.ErrMalformedKey.Error())
}

func TestMemcacheCache_Del(t *testing.T) {
	t.Parallel()
	c := setupMemcached(t)
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

	// Malformed key
	err = c.Del(context.Background(), malformedKey)
	require.Error(t, err)
	assert.Contains(t, err.Error(), memcache.ErrMalformedKey.Error())
}

func TestMemcacheCache_Clear(t *testing.T) {
	t.Parallel()
	c := setupMemcached(t)
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

func TestMemcacheCache_Get(t *testing.T) {
	t.Parallel()
	c := setupMemcached(t)
	key := testutil.UniqueKey(t)
	value := "testValue"

	if err := c.Set(context.Background(), key, value); err != nil {
		t.Fatalf("Failed to set key: %v", err)
	}
	t.Cleanup(func() {
		c.client.Delete(key)
	})

	got, err := c.Get(context.Background(), key)
	require.NoError(t, err)
	assert.Equal(t, []byte(value), got)

	// Non-existent key
	_, err = c.Get(context.Background(), "nonExistentKey")
	require.Error(t, err)
	assert.Contains(t, err.Error(), cache.ErrKeyNotFound.Error())

	// Malformed key
	_, err = c.Get(context.Background(), malformedKey)
	require.Error(t, err)
	assert.Contains(t, err.Error(), memcache.ErrMalformedKey.Error())
}

func TestMemcacheCache_Set(t *testing.T) {
	t.Parallel()
	c := setupMemcached(t)
	key := testutil.UniqueKey(t)
	value := "testValue"

	err := c.Set(context.Background(), key, value)
	require.NoError(t, err)
	t.Cleanup(func() {
		c.client.Delete(key)
	})

	got, err := c.Get(context.Background(), key)
	require.NoError(t, err)
	assert.Equal(t, []byte(value), got)

	// Malformed key
	err = c.Set(context.Background(), malformedKey, value)
	require.Error(t, err)
	assert.Contains(t, err.Error(), memcache.ErrMalformedKey.Error())
}

func TestMemcacheCache_SetWithExpiry(t *testing.T) {
	t.Parallel()
	c := setupMemcached(t)
	key := testutil.UniqueKey(t)
	value := "testValue"
	expiry := 1 * time.Second

	err := c.SetWithExpiry(context.Background(), key, value, expiry)
	require.NoError(t, err)
	t.Cleanup(func() {
		c.client.Delete(key)
	})

	got, err := c.Get(context.Background(), key)
	require.NoError(t, err)
	assert.Equal(t, []byte(value), got)

	// Wait for the key to expire
	time.Sleep(expiry + 1*time.Second)

	_, err = c.Get(context.Background(), key)
	require.Error(t, err)
	assert.Contains(t, err.Error(), cache.ErrKeyNotFound.Error())

	// Malformed key
	err = c.SetWithExpiry(context.Background(), malformedKey, value, expiry)
	require.Error(t, err)
	assert.Contains(t, err.Error(), memcache.ErrMalformedKey.Error())
}

// Pattern matching operations not supported by Memcache

func TestMemcacheCache_Count(t *testing.T) {
	t.Parallel()
	c := setupMemcached(t)
	_, err := c.Count(context.Background(), "*")
	require.Error(t, err)
	assert.Contains(t, err.Error(), cache.ErrPatternMatchingNotSupported.Error())
}

func TestMemcacheCache_DelKeys(t *testing.T) {
	t.Parallel()
	c := setupMemcached(t)
	err := c.DelKeys(context.Background(), "*")
	require.Error(t, err)
	assert.Contains(t, err.Error(), cache.ErrPatternMatchingNotSupported.Error())
}

func TestMemcacheCache_Ping(t *testing.T) {
	t.Parallel()
	c := setupMemcached(t)

	err := c.Ping(context.Background())
	require.NoError(t, err)
}

func TestMemcacheCache_Close(t *testing.T) {
	t.Parallel()
	c := setupMemcached(t)

	err := c.Close()
	require.NoError(t, err)

	// After closing, pinging should still succeed because Close is a no-op
	err = c.Ping(context.Background())
	require.NoError(t, err)
}
