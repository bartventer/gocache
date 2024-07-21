// Package driver defines an interface for cache implementations, providing a standardized
// way to interact with different caching mechanisms. It includes operations for setting,
// getting, deleting keys, and managing cache lifecycles.
package driver

import (
	"context"
	"time"

	"github.com/bartventer/gocache/pkg/keymod"
)

// Cache defines the interface for cache operations. Implementations of Cache should
// provide mechanisms for key-value storage, retrieval, deletion, and lifecycle management.
// It supports basic operations such as setting and getting values, checking existence,
// counting keys, and more advanced operations like setting values with TTL, deleting keys
// by pattern, and clearing the cache.
type Cache interface {
	// Set stores a key-value pair in the cache. It overwrites any existing value for the key.
	Set(ctx context.Context, key string, value interface{}, modifiers ...keymod.Mod) error

	// SetWithTTL stores a key-value pair in the cache with a specified time-to-live.
	// After the TTL expires, the key-value pair is automatically removed from the cache.
	SetWithTTL(ctx context.Context, key string, value interface{}, ttl time.Duration, modifiers ...keymod.Mod) error

	// Exists checks whether a key exists in the cache. It returns true if the key exists, false otherwise.
	Exists(ctx context.Context, key string, modifiers ...keymod.Mod) (bool, error)

	// Count returns the number of keys in the cache that match a given pattern.
	Count(ctx context.Context, pattern string, modifiers ...keymod.Mod) (int64, error)

	// Get retrieves the value associated with a key from the cache. If the key does not exist, an error is returned.
	Get(ctx context.Context, key string, modifiers ...keymod.Mod) ([]byte, error)

	// Del removes a key from the cache. If the key does not exist, it does nothing.
	Del(ctx context.Context, key string, modifiers ...keymod.Mod) error

	// DelKeys removes all keys from the cache that match a given pattern.
	DelKeys(ctx context.Context, pattern string, modifiers ...keymod.Mod) error

	// Clear removes all key-value pairs from the cache.
	Clear(ctx context.Context) error

	// Ping verifies that the cache is accessible and operational.
	Ping(ctx context.Context) error

	// Close terminates the connection to the cache, releasing any allocated resources.
	Close() error
}
