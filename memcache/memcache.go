/*
Package memcache provides a Memcache Client implementation of the cache.Cache interface.
It uses the memcache library to interact with a Memcache Client.

# URL Format:

The URL should have the following format:

	memcache://<host>:<port>

The <host>:<port> pair corresponds to the Memcache Client node.

# Usage

Example via generic cache interface:

	import (
	    "context"
	    "log"
	    "net/url"

	    "github.com/bartventer/gocache"
	    _ "github.com/bartventer/gocache/memcache"
	)

	func main() {
	    ctx := context.Background()
	    urlStr := "memcache://localhost:11211"
	    c, err := cache.OpenCache(ctx, urlStr, cache.Options{})
	    if err != nil {
	        log.Fatalf("Failed to initialize cache: %v", err)
	    }
	    // ... use c with the cache.Cache interface
	}

Example via [memcache.New] constructor:

	import (
	    "context"
	    "log"
	    "net/url"

	    "github.com/bartventer/gocache"
	    "github.com/bartventer/gocache/memcache"
	    "github.com/bradfitz/gomemcache/memcache"
	)

	func main() {
	    ctx := context.Background()
	    c := memcache.New(ctx, cache.Config{}, "localhost:11211")
	    // ... use c with the cache.Cache interface
	}
*/
package memcache

import (
	"context"
	"net/url"
	"strings"
	"sync"
	"time"

	cache "github.com/bartventer/gocache"
	"github.com/bradfitz/gomemcache/memcache"
)

// Scheme is the cache scheme for Memcache.
const Scheme = "memcache"

func init() { //nolint:gochecknoinits // This is the entry point of the package.
	cache.RegisterCache(Scheme, &memcacheCache{})
}

// memcacheCache is a Memcache implementation of the cache.Cache interface.
type memcacheCache struct {
	once   sync.Once        // once ensures that the cache is initialized only once.
	client *memcache.Client // client is the Memcache client.
	config *cache.Config    // config is the cache configuration.
}

// New returns a new Memcache cache implementation.
func New(ctx context.Context, config cache.Config, server ...string) *memcacheCache {
	m := &memcacheCache{}
	m.init(ctx, config, server...)
	return m
}

// Ensure MemcacheCache implements the cache.Cache interface.
var _ cache.Cache = &memcacheCache{}

// OpenCacheURL opens a new Memcache cache using the given URL and options.
// It implements the cache.CacheURLOpener interface.
func (m *memcacheCache) OpenCacheURL(ctx context.Context, u *url.URL, options cache.Options) (cache.Cache, error) {
	addrs := strings.Split(u.Host, ",")
	// Initialize the Memcache client
	m.init(ctx, options.Config, addrs...)
	return m, nil
}

// init initializes the Memcache client with the given options.
// It implements the cache.Cache interface.
func (m *memcacheCache) init(_ context.Context, config cache.Config, server ...string) {
	m.once.Do(func() {
		m.config = &config
		m.client = memcache.New(server...)
	})
}

// Count implements cache.Cache.
func (m *memcacheCache) Count(ctx context.Context, pattern string) (int64, error) {
	// Memcache does not support key pattern matching
	return 0, nil
}

// Exists implements cache.Cache.
func (m *memcacheCache) Exists(ctx context.Context, key string) (bool, error) {
	_, err := m.client.Get(key)
	if err == memcache.ErrCacheMiss {
		return false, nil
	}
	return err == nil, err
}

// Del deletes a key from the cache.
// It implements the cache.Cache interface.
func (m *memcacheCache) Del(ctx context.Context, key string) error {
	return m.client.Delete(key)
}

// DelKeys deletes all keys matching a pattern from the cache.
// It implements the cache.Cache interface.
func (m *memcacheCache) DelKeys(ctx context.Context, pattern string) error {
	// Memcache does not support key pattern matching
	return nil
}

// Clear deletes all keys from the cache.
// It implements the cache.Cache interface.
func (m *memcacheCache) Clear(ctx context.Context) error {
	return m.client.DeleteAll()
}

// Get gets the value of a key from the cache.
// It implements the cache.Cache interface.
func (m *memcacheCache) Get(ctx context.Context, key string) ([]byte, error) {
	item, err := m.client.Get(key)
	if err != nil {
		return nil, err
	}
	return item.Value, nil
}

// Set sets a key to a value in the cache.
// It implements the cache.Cache interface.
func (m *memcacheCache) Set(ctx context.Context, key string, value interface{}) error {
	item := &memcache.Item{
		Key:   key,
		Value: []byte(value.(string)),
	}
	return m.client.Set(item)
}

// SetWithExpiry sets a key to a value in the cache with an expiry time.
// It implements the cache.Cache interface.
func (m *memcacheCache) SetWithExpiry(ctx context.Context, key string, value interface{}, expiry time.Duration) error {
	item := &memcache.Item{
		Key:        key,
		Value:      []byte(value.(string)),
		Expiration: int32(expiry.Seconds()),
	}
	return m.client.Set(item)
}
