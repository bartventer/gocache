/*
Package memcache provides a Memcache Client implementation of the [cache.Cache] interface.
It uses the memcache library to interact with a Memcache Client.

# URL Format:

The URL should have the following format:

	memcache://<host1>:<port1>,<host2>:<port2>,...,<hostN>:<portN>

Each <host>:<port> pair corresponds to the Memcache Client node.

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
	    c, err := cache.OpenCache(ctx, urlStr)
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
	)

	func main() {
	    ctx := context.Background()
	    c := memcache.New(ctx, &memcache.Options{
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
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	cache "github.com/bartventer/gocache"
	"github.com/bartventer/gocache/internal/gcerrors"
	"github.com/bartventer/gocache/keymod"
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
}

// New returns a new Memcache cache implementation.
func New(ctx context.Context, opts *Options) *memcacheCache {
	m := &memcacheCache{}
	m.init(ctx, opts)
	return m
}

// Ensure MemcacheCache implements the cache.Cache interface.
var _ cache.Cache = &memcacheCache{}

// OpenCacheURL implements cache.URLOpener.
func (m *memcacheCache) OpenCacheURL(ctx context.Context, u *url.URL) (cache.Cache, error) {
	addrs := strings.Split(u.Host, ",")
	m.init(ctx, &Options{Addrs: addrs})
	return m, nil
}

func (m *memcacheCache) init(_ context.Context, opts *Options) {
	m.once.Do(func() {
		if opts == nil {
			opts = &Options{}
		}
		m.client = memcache.New(opts.Addrs...)
	})
}

// Count implements cache.Cache.
func (m *memcacheCache) Count(_ context.Context, pattern string, modifiers ...keymod.Mod) (int64, error) {
	return 0, gcerrors.NewWithScheme(Scheme, fmt.Errorf("Count not supported: %w", cache.ErrPatternMatchingNotSupported))
}

// Exists implements cache.Cache.
func (m *memcacheCache) Exists(_ context.Context, key string, modifiers ...keymod.Mod) (bool, error) {
	key = keymod.Modify(key, modifiers...)
	_, err := m.client.Get(key)
	if err != nil {
		if err == memcache.ErrCacheMiss {
			return false, nil
		} else {
			return false, gcerrors.NewWithScheme(Scheme, fmt.Errorf("error checking key %s, underlying error: %w", key, err))
		}
	}
	return true, nil
}

// Del implements cache.Cache.
func (m *memcacheCache) Del(_ context.Context, key string, modifiers ...keymod.Mod) error {
	key = keymod.Modify(key, modifiers...)
	err := m.client.Delete(key)
	if err != nil {
		if err == memcache.ErrCacheMiss {
			return gcerrors.NewWithScheme(Scheme, fmt.Errorf("%s: %w, underlying error: %w", key, cache.ErrKeyNotFound, err))
		} else {
			return gcerrors.NewWithScheme(Scheme, fmt.Errorf("error deleting key %s, underlying error: %w", key, err))
		}
	}
	return nil
}

// DelKeys implements cache.Cache.
func (m *memcacheCache) DelKeys(_ context.Context, pattern string, modifiers ...keymod.Mod) error {
	return gcerrors.NewWithScheme(Scheme, fmt.Errorf("DelKeys not supported: %w", cache.ErrPatternMatchingNotSupported))
}

// Clear implements cache.Cache.
func (m *memcacheCache) Clear(_ context.Context) error {
	return m.client.DeleteAll()
}

// Get implements cache.Cache.
func (m *memcacheCache) Get(_ context.Context, key string, modifiers ...keymod.Mod) ([]byte, error) {
	key = keymod.Modify(key, modifiers...)
	item, err := m.client.Get(key)
	if err != nil {
		if err == memcache.ErrCacheMiss {
			return nil, gcerrors.NewWithScheme(Scheme, fmt.Errorf("%s: %w, underlying error: %w", key, cache.ErrKeyNotFound, err))
		} else {
			return nil, gcerrors.NewWithScheme(Scheme, fmt.Errorf("error getting key %s, underlying error: %w", key, err))
		}
	}
	return item.Value, nil
}

// Set implements cache.Cache.
func (m *memcacheCache) Set(_ context.Context, key string, value interface{}, modifiers ...keymod.Mod) error {
	key = keymod.Modify(key, modifiers...)
	item := &memcache.Item{
		Key:   key,
		Value: []byte(value.(string)),
	}
	err := m.client.Set(item)
	if err != nil {
		return gcerrors.NewWithScheme(Scheme, fmt.Errorf("error setting key %s: %w", key, err))
	}
	return nil
}

// SetWithExpiry implements cache.Cache.
func (m *memcacheCache) SetWithExpiry(_ context.Context, key string, value interface{}, expiry time.Duration, modifiers ...keymod.Mod) error {
	key = keymod.Modify(key, modifiers...)
	item := &memcache.Item{
		Key:        key,
		Value:      []byte(value.(string)),
		Expiration: int32(expiry.Seconds()),
	}
	err := m.client.Set(item)
	if err != nil {
		return gcerrors.NewWithScheme(Scheme, fmt.Errorf("error setting key %s with expiry: %w", key, err))
	}
	return nil
}

// Ping implements cache.Cache.
func (m *memcacheCache) Ping(ctx context.Context) error {
	return m.client.Ping()
}

// Close implements cache.Cache.
func (m *memcacheCache) Close() error {
	return m.client.Close()
}
