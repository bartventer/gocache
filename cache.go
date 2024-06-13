/*
Package cache offers a library for managing caches in a unified manner in Go.

This package enables developers to transition between various cache implementations
(such as Redis, Redis Cluster, etc.) by altering the URL scheme. This is achieved by
offering consistent, idiomatic interfaces for common operations.

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
	    c, err := cache.OpenCache(ctx, urlStr, cache.Options{})
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
	"crypto/tls"
	"errors"
	"sync"
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
	Set(ctx context.Context, key string, value interface{}, modifiers ...keymod.KeyModifier) error

	// SetWithExpiry sets a key to a value in the cache with an expiry time.
	SetWithExpiry(ctx context.Context, key string, value interface{}, expiry time.Duration, modifiers ...keymod.KeyModifier) error

	// Exists checks if a key exists in the cache.
	Exists(ctx context.Context, key string, modifiers ...keymod.KeyModifier) (bool, error)

	// Count returns the number of keys in the cache matching a pattern.
	Count(ctx context.Context, pattern string, modifiers ...keymod.KeyModifier) (int64, error)

	// Get gets the value of a key from the cache.
	Get(ctx context.Context, key string, modifiers ...keymod.KeyModifier) ([]byte, error)

	// Del deletes a key from the cache.
	Del(ctx context.Context, key string, modifiers ...keymod.KeyModifier) error

	// DelKeys deletes all keys matching a pattern from the cache.
	DelKeys(ctx context.Context, pattern string, modifiers ...keymod.KeyModifier) error

	// Clear clears all keys from the cache.
	Clear(ctx context.Context) error

	// Ping checks if the cache is available.
	Ping(ctx context.Context) error

	// Close closes the cache connection.
	Close() error
}

const (
	// DefaultCountLimit is the default value for the [Config.CountLimit] option.
	DefaultCountLimit = 10
)

// Config is a struct that holds configuration options for the cache package.
//
// # Compatibility
//
// These options are recognized by all cache drivers.
type Config struct {
	once sync.Once // once ensures the configuration is revised only once.
	// CountLimit is the hint to the SCAN command about the amount of work to be done at each call.
	// It does not guarantee the exact number of elements returned at every iteration, but the server
	// will usually return this count or a few more elements per call. For small sets or hashes, all
	// elements may be returned in the first SCAN call regardless of the CountLimit value. The CountLimit
	// value can be changed between iterations. The default value is 10.
	CountLimit int64
}

// Revise revises the configuration options to ensure they contain sensible values.
func (c *Config) Revise() {
	c.once.Do(c.revise)
}

func (c *Config) revise() {
	if c.CountLimit <= 0 {
		c.CountLimit = DefaultCountLimit
	}
}

// Options is a struct that holds provider specific configuration options.
//
// # Compatibility
//
// These options are only recognized by the following drivers:
//   - [redis]
//   - [rediscluster]
//
// Other drivers will simply ignore these options.
//
// [redis]: https://pkg.go.dev/github.com/bartventer/gocache/redis
// [rediscluster]: https://pkg.go.dev/github.com/bartventer/gocache/rediscluster
type Options struct {
	Config
	// TLSConfig is the TLS configuration for the cache connection.
	TLSConfig *tls.Config
	// CredentialsProvider is a function that returns the username and password for the cache connection.
	CredentialsProvider func(ctx context.Context) (username string, password string, err error)
	// Metadata is a map of provider specific configuration options.
	// It offers an alternative method for configuring the provider.
	// These values will override the URL values.
	// Note: Network address (host:port) and function values will be ignored if provided.
	// Refer to the driver documentation for available options.
	// The map keys are case insensitive.
	//
	// Example usage for a Redis cache:
	//
	//  map[string]string{
	// 	 	"Network": "tcp",
	// 	 	"MaxRetries": "3",
	// 	 	"MinRetryBackoff": "8ms",
	// 	 	"PoolFIFO": "true",
	// 	}
	//
	// This is equivalent to providing query parameters in the URL:
	//
	// 	redis://localhost:6379?network=tcp&maxretries=3&minretrybackoff=8ms&poolfifo=true
	Metadata map[string]string
}
