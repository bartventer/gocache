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
Refer to the documentation of each cache implementation for more information.

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
	"errors"
	"time"

	"github.com/bartventer/gocache/keymod"
)

var (
	// ErrNoCache is returned when no cache implementation is available.
	ErrNoCache = errors.New("gocache: no cache implementation available")

	// ErrKeyNotFound is returned when a key is not found in the cache.
	ErrKeyNotFound = errors.New("gocache: key not found")

	// ErrPatternMatchingNotSupported is returned when a pattern matching operation is not supported
	// by the cache implementation.
	ErrPatternMatchingNotSupported = errors.New("gocache: pattern matching not supported")
)

// Cache is an interface that represents a cache. It has methods for setting, getting and deleting keys.
// Each cache implementation should implement this interface.
type Cache interface {
	// Set sets a key to a value in the cache.
	Set(ctx context.Context, key string, value interface{}, modifiers ...keymod.Mod) error

	// SetWithExpiry sets a key to a value in the cache with an expiry time.
	SetWithExpiry(ctx context.Context, key string, value interface{}, expiry time.Duration, modifiers ...keymod.Mod) error

	// Exists checks if a key exists in the cache.
	Exists(ctx context.Context, key string, modifiers ...keymod.Mod) (bool, error)

	// Count returns the number of keys in the cache matching a pattern.
	Count(ctx context.Context, pattern string, modifiers ...keymod.Mod) (int64, error)

	// Get gets the value of a key from the cache.
	Get(ctx context.Context, key string, modifiers ...keymod.Mod) ([]byte, error)

	// Del deletes a key from the cache.
	Del(ctx context.Context, key string, modifiers ...keymod.Mod) error

	// DelKeys deletes all keys matching a pattern from the cache.
	DelKeys(ctx context.Context, pattern string, modifiers ...keymod.Mod) error

	// Clear clears all keys from the cache.
	Clear(ctx context.Context) error

	// Ping checks if the cache is available.
	Ping(ctx context.Context) error

	// Close closes the cache connection.
	Close() error
}
