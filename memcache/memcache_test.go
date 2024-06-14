package memcache

import (
	"context"
	"net/url"
	"strings"
	"testing"
	"time"

	cache "github.com/bartventer/gocache"
	"github.com/bartventer/gocache/drivertest"
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

// setupCache creates a new Memcached container.
func setupCache(t *testing.T) *memcacheCache {
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

func TestMemcacheCache_MalformedKey(t *testing.T) {
	t.Parallel()
	c := setupCache(t)
	// malformedKey is a key that is too long which will trigger the [memcache.ErrMalformedKey] error.
	malformedKey := strings.Repeat("a", 251)
	value := "testValue"

	// Test Exists function with malformed key
	_, err := c.Exists(context.Background(), malformedKey)
	require.Error(t, err)
	assert.Contains(t, err.Error(), memcache.ErrMalformedKey.Error())

	// Test Del function with malformed key
	err = c.Del(context.Background(), malformedKey)
	require.Error(t, err)
	assert.Contains(t, err.Error(), memcache.ErrMalformedKey.Error())

	// Test Get function with malformed key
	_, err = c.Get(context.Background(), malformedKey)
	require.Error(t, err)
	assert.Contains(t, err.Error(), memcache.ErrMalformedKey.Error())

	// Test Set function with malformed key
	err = c.Set(context.Background(), malformedKey, value)
	require.Error(t, err)
	assert.Contains(t, err.Error(), memcache.ErrMalformedKey.Error())
}

type harness struct {
	cache *memcacheCache
}

func (h *harness) MakeCache(ctx context.Context) (cache.Cache, error) {
	return h.cache, nil
}

func (h *harness) Close() {
	// Cleanup is handled in setup function
}

func (h *harness) Options() drivertest.Options {
	return drivertest.Options{
		PatternMatchingDisabled: true, // Memcached does not support pattern matching
		CloseIsNoop:             true, // Cache can still be used after closing
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
