/*
Package cache offers a library for managing caches in a unified manner in Go.

This package enables developers to transition between various cache implementations
(such as Redis, Redis Cluster, Memcache, etc.) by altering the URL scheme. This is
achieved by offering consistent, idiomatic interfaces for common operations.

Central to this package are "portable types", built atop service-specific drivers for
supported cache services. For instance, the [GenericCache] type is a portable type that
implements the [driver.Cache] interface. This allows developers to use the same code
with different cache implementations.

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
	    c, err := cache.OpenCache[string](ctx, urlStr)
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

# Custom Key Types

The cache package supports any string-like type for keys. Custom key types can be used
by registering it alongside the cache implementation with [RegisterCache]. This allows the
cache package to automatically convert the custom key type to a string when interacting
with the cache.

The following key types are supported by default:
  - string
  - [keymod.Key]

For an example of how to register a custom key type, refer to the [redis] driver package.

[redis]: https://pkg.go.dev/github.com/bartventer/gocache/redis
*/
package cache

import (
	"context"
	"time"

	"github.com/bartventer/gocache/pkg/driver"
	"github.com/bartventer/gocache/pkg/keymod"
)

// Supports any string-like type for keys.
var _ driver.Cache[string] = new(GenericCache[string])
var _ driver.Cache[keymod.Key] = new(GenericCache[keymod.Key])

// KeyCache is a type alias for [GenericCache] with [keymod.Key] keys.
type KeyCache = GenericCache[keymod.Key]

// Cache is a type alias for [GenericCache] with string keys.
type Cache = GenericCache[string]

// GenericCache is a portable type that implements [driver.Cache].
type GenericCache[K driver.String] struct {
	driver driver.Cache[K]
}

// Clear implements [driver.Cache].
func (c *GenericCache[K]) Clear(ctx context.Context) error {
	return c.driver.Clear(ctx)
}

// Close implements [driver.Cache].
func (c *GenericCache[K]) Close() error {
	return c.driver.Close()
}

// Count implements [driver.Cache].
func (c *GenericCache[K]) Count(ctx context.Context, pattern K) (int64, error) {
	return c.driver.Count(ctx, pattern)
}

// Del implements [driver.Cache].
func (c *GenericCache[K]) Del(ctx context.Context, key K) error {
	return c.driver.Del(ctx, key)
}

// DelKeys implements [driver.Cache].
func (c *GenericCache[K]) DelKeys(ctx context.Context, pattern K) error {
	return c.driver.DelKeys(ctx, pattern)
}

// Exists implements [driver.Cache].
func (c *GenericCache[K]) Exists(ctx context.Context, key K) (bool, error) {
	return c.driver.Exists(ctx, key)
}

// Get implements [driver.Cache].
func (c *GenericCache[K]) Get(ctx context.Context, key K) ([]byte, error) {
	return c.driver.Get(ctx, key)
}

// Ping implements [driver.Cache].
func (c *GenericCache[K]) Ping(ctx context.Context) error {
	return c.driver.Ping(ctx)
}

// Set implements [driver.Cache].
func (c *GenericCache[K]) Set(ctx context.Context, key K, value interface{}) error {
	return c.driver.Set(ctx, key, value)
}

// SetWithTTL implements [driver.Cache].
func (c *GenericCache[K]) SetWithTTL(ctx context.Context, key K, value interface{}, ttl time.Duration) error {
	return c.driver.SetWithTTL(ctx, key, value, ttl)
}

// NewCache creates a new [GenericCache] using the provided driver. Not intended for direct application use.
func NewCache[K driver.String](driver driver.Cache[K]) *GenericCache[K] {
	return &GenericCache[K]{driver: driver}
}
