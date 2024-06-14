/*
Package ramcache implements the [cache.Cache] interface using an in-memory map.

It's useful for testing, development, and caching small data sets. It's not recommended
for production due to lack of data persistence across restarts.

# URL Format

The URL should have the following format:

	ramcache://[?query]

The query part, though optional, can be used for additional configuration through query parameters.

Query parameters can be used to configure the in-memory cache options. The keys of the query
parameters should correspond to the case-insensitive field names of [Options].

# Usage

Example via generic cache interface:

	import (
	    "context"
	    "log"

	    "github.com/bartventer/gocache"
	    _ "github.com/bartventer/gocache/ramcache"
	)

	func main() {
	    ctx := context.Background()
	    c, err := cache.OpenCache(ctx, "ramcache://?defaultttl=5m", cache.Options{})
	    if err != nil {
	        log.Fatalf("Failed to initialize cache: %v", err)
	    }
	    // ... use c with the cache.Cache interface
	}

Example via [ramcache.New] constructor:

	import (
	    "context"

	    "github.com/bartventer/gocache"
	    "github.com/bartventer/gocache/ramcache"
	)

	func main() {
	    ctx := context.Background()
	    c := ramcache.New(ctx, &cache.Config{}, ramcache.Options{DefaultTTL: 5 * time.Minute})
	    // ... use c with the cache.Cache interface
	}

# Limitations

Please note that due to the limitations of the RAM Cache, pattern matching
operations are not supported. This includes the [cache.Cache] Count and DelKeys methods, which will return a
[cache.ErrPatternMatchingNotSupported] error if called.
*/
package ramcache

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sync"
	"time"

	cache "github.com/bartventer/gocache"
	"github.com/bartventer/gocache/internal/gcerrors"
	"github.com/bartventer/gocache/keymod"
)

// Scheme is the cache scheme for the in-memory cache.
const Scheme = "ramcache"

func init() { //nolint:gochecknoinits // This is the entry point of the package.
	cache.RegisterCache(Scheme, &ramcache{})
}

var _ cache.Cache = new(ramcache)
var _ cache.URLOpener = new(ramcache)

// Options are the cache options.
type Options struct {
	// DefaultTTL is the default time-to-live for cache items.
	DefaultTTL time.Duration
}

// Revise revises the options, ensuring sensible defaults are set.
func (r *Options) Revise() {
	if r.DefaultTTL <= 0 {
		r.DefaultTTL = 24 * time.Hour
	}
}

// Item is a cache Item.
type Item struct {
	Value  []byte    // Value is the item value.
	Expiry time.Time // Expiry is the item expiry time. Default is 24 hours.
}

// ramcache is an in-memory implementation of the cache.Cache interface.
type ramcache struct {
	once   sync.Once       // once ensures that the cache is initialized only once.
	mu     sync.RWMutex    // mu guards the store.
	store  map[string]Item // store is the in-memory store.
	config *cache.Config   // config is the cache configuration.
	opts   *Options        // options is the cache options.
}

// OpenCacheURL implements cache.URLOpener.
func (r *ramcache) OpenCacheURL(ctx context.Context, u *url.URL, options *cache.Options) (cache.Cache, error) {
	// Parse the URL into Redis options
	ramOpts, err := optionsFromURL(u, options.Metadata)
	if err != nil {
		return nil, err
	}
	r.init(ctx, &options.Config, ramOpts)
	return r, nil
}

func (r *ramcache) init(_ context.Context, config *cache.Config, options Options) {
	r.once.Do(func() {
		r.config = config
		r.store = make(map[string]Item)
		options.Revise()
		r.opts = &options
	})
}

// New returns a new in-memory cache implementation.
func New(ctx context.Context, config *cache.Config, options Options) *ramcache {
	config.Revise()
	r := &ramcache{}
	r.init(ctx, config, options)
	return r
}

// OpenCacheURL opens a cache using a URL.
func OpenCacheURL(ctx context.Context, url string) (cache.Cache, error) {
	// Implement URL parsing and opening here.
	return nil, errors.New("not implemented")
}

// Ensure ramcache implements the cache.Cache interface.
var _ cache.Cache = &ramcache{}

// Count implements cache.Cache.
func (r *ramcache) Count(ctx context.Context, pattern string, modifiers ...keymod.KeyModifier) (int64, error) {
	return 0, gcerrors.NewWithScheme(Scheme, fmt.Errorf("Count not supported: %w", cache.ErrPatternMatchingNotSupported))
}

// Exists implements cache.Cache.
func (r *ramcache) Exists(ctx context.Context, key string, modifiers ...keymod.KeyModifier) (bool, error) {
	key = keymod.ModifyKey(key, modifiers...)
	r.mu.RLock()
	item, exists := r.store[key]
	r.mu.RUnlock()
	if exists && time.Now().After(item.Expiry) {
		r.mu.Lock()
		delete(r.store, key)
		r.mu.Unlock()
		exists = false
	}
	return exists, nil
}

// Del implements cache.Cache.
func (r *ramcache) Del(ctx context.Context, key string, modifiers ...keymod.KeyModifier) error {
	key = keymod.ModifyKey(key, modifiers...)
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.store[key]; !exists {
		return gcerrors.NewWithScheme(Scheme, fmt.Errorf("%s: %w", key, cache.ErrKeyNotFound))
	}
	delete(r.store, key)
	return nil
}

// DelKeys implements cache.Cache.
func (r *ramcache) DelKeys(ctx context.Context, pattern string, modifiers ...keymod.KeyModifier) error {
	return gcerrors.NewWithScheme(Scheme, fmt.Errorf("DelKeys not supported: %w", cache.ErrPatternMatchingNotSupported))
}

// Clear implements cache.Cache.
func (r *ramcache) Clear(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.store = make(map[string]Item)
	return nil
}

// Get implements cache.Cache.
func (r *ramcache) Get(ctx context.Context, key string, modifiers ...keymod.KeyModifier) ([]byte, error) {
	key = keymod.ModifyKey(key, modifiers...)
	r.mu.Lock()
	defer r.mu.Unlock()
	it, exists := r.store[key]
	if !exists || time.Now().After(it.Expiry) {
		delete(r.store, key)
		return nil, gcerrors.NewWithScheme(Scheme, cache.ErrKeyNotFound)
	}
	return it.Value, nil
}

// Set implements cache.Cache.
func (r *ramcache) Set(ctx context.Context, key string, value interface{}, modifiers ...keymod.KeyModifier) error {
	key = keymod.ModifyKey(key, modifiers...)
	r.mu.Lock()
	defer r.mu.Unlock()
	switch v := value.(type) {
	case string:
		r.store[key] = Item{Value: []byte(v), Expiry: time.Now().Add(r.opts.DefaultTTL)}
	case []byte:
		r.store[key] = Item{Value: v, Expiry: time.Now().Add(r.opts.DefaultTTL)}
	default:
		return gcerrors.NewWithScheme(Scheme, fmt.Errorf("unsupported value type: %T", v))
	}
	return nil
}

// SetWithExpiry implements cache.Cache.
func (r *ramcache) SetWithExpiry(ctx context.Context, key string, value interface{}, expiry time.Duration, modifiers ...keymod.KeyModifier) error {
	key = keymod.ModifyKey(key, modifiers...)
	r.mu.Lock()
	defer r.mu.Unlock()
	switch v := value.(type) {
	case string:
		r.store[key] = Item{Value: []byte(v), Expiry: time.Now().Add(expiry)}
	case []byte:
		r.store[key] = Item{Value: v, Expiry: time.Now().Add(expiry)}
	default:
		return gcerrors.NewWithScheme(Scheme, fmt.Errorf("unsupported value type: %T", v))
	}
	return nil
}

// Close implements cache.Cache.
func (r *ramcache) Close() error {
	return nil
}

// Ping implements cache.Cache.
func (r *ramcache) Ping(ctx context.Context) error {
	return nil
}
