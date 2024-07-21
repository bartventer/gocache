/*
Package cache offers a library for managing caches in a unified manner in Go.

This package enables developers to transition between various cache implementations
(such as Redis, Redis Cluster, Memcache, etc.) by altering the URL scheme. This is
achieved by offering consistent, idiomatic interfaces for common operations.

Central to this package are "portable types", built atop service-specific drivers for
supported cache services. For instance, cache.Cache portable type instances can be
created using redis.OpenCache, memcache.OpenCache, or any other supported driver. This
allows the cache.Cache to be used across your application without concern for the
underlying implementation.

# URL Format

The cache package uses URLs to specify cache implementations. The URL scheme is used
to determine the cache implementation to use. The URL format is:

	scheme://<host>:<port>[?query]

The scheme is used to determine the cache implementation. The host and port are used
to connect to the cache service. The optional query parameters can be used to configure
the cache implementation. Each cache implementation supports different query parameters.

# Usage

To use a cache implementation, import the relevant driver package and use the
[OpenCache] function to create a new cache. The cache package will automatically
use the correct cache implementation based on the URL scheme.

	import (
	    "context"
	    "log"

	    "github.com/bartventer/gocache"
	    // Enable the Redis cache implementation
	    _ "github.com/bartventer/gocache/redis"
	)

	func main() {
	    ctx := context.Background()
	    urlStr := "redis://localhost:7000?maxretries=5&minretrybackoff=1000"
	    c, err := cache.OpenCache(ctx, urlStr)
	    if err != nil {
	        log.Fatalf("Failed to initialize cache: %v", err)
	    }

	    // Now you can use c with the cache.Cache interface
	    err = c.Set(ctx, "key", "value")
	    if err != nil {
	        log.Fatalf("Failed to set key: %v", err)
	    }

	    value, err := c.Get(ctx, "key")
	    if err != nil {
	        log.Fatalf("Failed to get key: %v", err)
	    }

	    log.Printf("Value: %s", value)
	}

# Drivers

For specific URL formats, query parameters, and examples, refer to the documentation
of each cache implementation.
*/
package cache

import (
	"context"
	"time"

	"github.com/bartventer/gocache/pkg/driver"
	"github.com/bartventer/gocache/pkg/keymod"
)

var _ driver.Cache = new(Cache)

// Cache is a portable type that implements [driver.Cache].
type Cache struct {
	driver driver.Cache
}

// Clear implements [driver.Cache].
func (c *Cache) Clear(ctx context.Context) error {
	return c.driver.Clear(ctx)
}

// Close implements [driver.Cache].
func (c *Cache) Close() error {
	return c.driver.Close()
}

// Count implements [driver.Cache].
func (c *Cache) Count(ctx context.Context, pattern string, modifiers ...keymod.Mod) (int64, error) {
	return c.driver.Count(ctx, pattern, modifiers...)
}

// Del implements [driver.Cache].
func (c *Cache) Del(ctx context.Context, key string, modifiers ...keymod.Mod) error {
	return c.driver.Del(ctx, key, modifiers...)
}

// DelKeys implements [driver.Cache].
func (c *Cache) DelKeys(ctx context.Context, pattern string, modifiers ...keymod.Mod) error {
	return c.driver.DelKeys(ctx, pattern, modifiers...)
}

// Exists implements [driver.Cache].
func (c *Cache) Exists(ctx context.Context, key string, modifiers ...keymod.Mod) (bool, error) {
	return c.driver.Exists(ctx, key, modifiers...)
}

// Get implements [driver.Cache].
func (c *Cache) Get(ctx context.Context, key string, modifiers ...keymod.Mod) ([]byte, error) {
	return c.driver.Get(ctx, key, modifiers...)
}

// Ping implements [driver.Cache].
func (c *Cache) Ping(ctx context.Context) error {
	return c.driver.Ping(ctx)
}

// Set implements [driver.Cache].
func (c *Cache) Set(ctx context.Context, key string, value interface{}, modifiers ...keymod.Mod) error {
	return c.driver.Set(ctx, key, value, modifiers...)
}

// SetWithTTL implements [driver.Cache].
func (c *Cache) SetWithTTL(ctx context.Context, key string, value interface{}, ttl time.Duration, modifiers ...keymod.Mod) error {
	return c.driver.SetWithTTL(ctx, key, value, ttl, modifiers...)
}

// NewCache creates a new [Cache] using the provided driver. Not intended for direct application use.
func NewCache(driver driver.Cache) *Cache {
	return &Cache{driver: driver}
}
