/*
Package memcache provides a Memcache Client implementation of the [driver.Cache] interface.
It uses the memcache library to interact with a Memcache Client.

# URL Format:

The URL should have the following format:

	memcache://<host1>:<port1>,<host2>:<port2>,...,<hostN>:<portN>

Each <host>:<port> pair corresponds to the Memcache Client node.

# Usage

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
	    c, err := cache.OpenCache(ctx, urlStr)
	    if err != nil {
	        log.Fatalf("Failed to initialize cache: %v", err)
	    }
	    // ... use c with the cache.Cache interface
	}

You can create a Memcache cache with [New]:

	import (
	    "context"
	    "log"
	    "net/url"

	    "github.com/bartventer/gocache/memcache"
	)

	func main() {
	    ctx := context.Background()
	    c := memcache.New[string](ctx, &memcache.Options{
			Addrs: []string{"localhost:11211"},
		})
	    // ... use c with the cache.Cache interface
	}

# Limitations

Please note that due to the limitations of the Memcache protocol, pattern matching
operations are not supported. This includes the [cache.Cache] Count and DelKeys methods, which will return a
[cache.ErrPatternMatchingNotSupported] error if called.
*/
package memcache

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	cache "github.com/bartventer/gocache"
	"github.com/bartventer/gocache/internal/gcerrors"
	"github.com/bartventer/gocache/pkg/driver"
	"github.com/bartventer/gocache/pkg/keymod"
	"github.com/bradfitz/gomemcache/memcache"
)

// Scheme is the cache scheme for Memcache.
const Scheme = "memcache"

func init() { //nolint:gochecknoinits // This is the entry point of the package.
	cache.RegisterCache(Scheme, &memcacheCache[string]{})
	cache.RegisterCache(Scheme, &memcacheCache[keymod.Key]{})
}

// memcacheCache is a Memcache implementation of the cache.Cache interface.
type memcacheCache[K driver.String] struct {
	once   sync.Once        // once ensures that the cache is initialized only once.
	client *memcache.Client // client is the Memcache client.
}

// New returns a new Memcache cache implementation.
func New[K driver.String](ctx context.Context, opts *Options) *memcacheCache[K] {
	m := &memcacheCache[K]{}
	m.init(ctx, opts)
	return m
}

// Ensure MemcacheCache implements the cache.Cache interface.
var _ driver.Cache[string] = new(memcacheCache[string])
var _ driver.Cache[keymod.Key] = new(memcacheCache[keymod.Key])

// OpenCacheURL implements cache.URLOpener.
func (m *memcacheCache[K]) OpenCacheURL(ctx context.Context, u *url.URL) (*cache.GenericCache[K], error) {
	addrs := strings.Split(u.Host, ",")
	m.init(ctx, &Options{Addrs: addrs})
	return cache.NewCache(m), nil
}

func (m *memcacheCache[K]) init(_ context.Context, opts *Options) {
	m.once.Do(func() {
		if opts == nil {
			opts = &Options{}
		}
		m.client = memcache.New(opts.Addrs...)
	})
}

// Count implements cache.Cache.
func (m *memcacheCache[K]) Count(_ context.Context, pattern K) (int64, error) {
	return 0, gcerrors.NewWithScheme(Scheme, errors.Join(cache.ErrPatternMatchingNotSupported, fmt.Errorf("Count operation not supported")))
}

// Exists implements cache.Cache.
func (m *memcacheCache[K]) Exists(_ context.Context, key K) (bool, error) {
	_, err := m.client.Get(string(key))
	if err != nil {
		if err == memcache.ErrCacheMiss {
			return false, nil
		} else {
			return false, gcerrors.NewWithScheme(Scheme, fmt.Errorf("error checking key %s: %w", key, err))
		}
	}
	return true, nil
}

// Del implements cache.Cache.
func (m *memcacheCache[K]) Del(_ context.Context, key K) error {
	err := m.client.Delete(string(key))
	if err != nil {
		if err == memcache.ErrCacheMiss {
			return gcerrors.NewWithScheme(Scheme, errors.Join(cache.ErrKeyNotFound, fmt.Errorf("key %s not found: %w", key, err)))
		} else {
			return gcerrors.NewWithScheme(Scheme, fmt.Errorf("error deleting key %s: %w", key, err))
		}
	}
	return nil
}

// DelKeys implements cache.Cache.
func (m *memcacheCache[K]) DelKeys(_ context.Context, pattern K) error {
	return gcerrors.NewWithScheme(Scheme, errors.Join(cache.ErrPatternMatchingNotSupported, fmt.Errorf("DelKeys operation not supported")))
}

// Clear implements cache.Cache.
func (m *memcacheCache[K]) Clear(_ context.Context) error {
	return m.client.DeleteAll()
}

// Get implements cache.Cache.
func (m *memcacheCache[K]) Get(_ context.Context, key K) ([]byte, error) {
	item, err := m.client.Get(string(key))
	if err != nil {
		if err == memcache.ErrCacheMiss {
			return nil, gcerrors.NewWithScheme(Scheme, errors.Join(cache.ErrKeyNotFound, fmt.Errorf("key %s not found: %w", key, err)))
		} else {
			return nil, gcerrors.NewWithScheme(Scheme, fmt.Errorf("error getting key %s: %w", key, err))
		}
	}
	return item.Value, nil
}

// Set implements cache.Cache.
func (m *memcacheCache[K]) Set(_ context.Context, key K, value interface{}) error {
	item := &memcache.Item{
		Key:   string(key),
		Value: []byte(value.(string)),
	}
	err := m.client.Set(item)
	if err != nil {
		return gcerrors.NewWithScheme(Scheme, fmt.Errorf("error setting key %s: %w", key, err))
	}
	return nil
}

// SetWithTTL implements cache.Cache.
func (m *memcacheCache[K]) SetWithTTL(_ context.Context, key K, value interface{}, ttl time.Duration) error {
	item := &memcache.Item{
		Key:        string(key),
		Value:      []byte(value.(string)),
		Expiration: int32(ttl.Seconds()),
	}
	err := m.client.Set(item)
	if err != nil {
		return gcerrors.NewWithScheme(Scheme, fmt.Errorf("error setting key %s with expiry: %w", key, err))
	}
	return nil
}

// Ping implements cache.Cache.
func (m *memcacheCache[K]) Ping(_ context.Context) error {
	return m.client.Ping()
}

// Close implements cache.Cache.
func (m *memcacheCache[K]) Close() error {
	return m.client.Close()
}
