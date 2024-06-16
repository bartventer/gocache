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
	// If not set, the default is 24 hours.
	DefaultTTL time.Duration
	// CleanupInterval is the interval at which checks for expired items are performed.
	// If not set, the default is 5 minutes.
	CleanupInterval time.Duration
}

// Revise revises the options, ensuring sensible defaults are set.
func (r *Options) Revise() {
	if r.DefaultTTL <= 0 {
		r.DefaultTTL = 24 * time.Hour
	}
	if r.CleanupInterval <= 0 {
		r.CleanupInterval = 5 * time.Minute
	}
}

// ramcache is an in-memory implementation of the cache.Cache interface.
type ramcache struct {
	once   sync.Once     // once ensures that the cache is initialized only once.
	store  *store        // store is the in-memory store.
	config *cache.Config // config is the cache configuration.
	opts   *Options      // options is the cache options.
	stopCh chan struct{} // stopCh is the stop channel.
}

// OpenCacheURL implements cache.URLOpener.
func (r *ramcache) OpenCacheURL(ctx context.Context, u *url.URL, options *cache.Options) (cache.Cache, error) {
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
		r.store = newStore()
		options.Revise()
		r.opts = &options
		r.stopCh = make(chan struct{})
		go r.cleanupExpiredItems()
	})
}

// cleanupExpiredItems periodically removes expired items from the store.
func (r *ramcache) cleanupExpiredItems() {
	ticker := time.NewTicker(r.opts.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			r.removeExpiredItems()
		case <-r.stopCh:
			return
		}
	}
}

// removeExpiredItems removes expired items from the store.
func (r *ramcache) removeExpiredItems() {
	keyItems := r.store.KeyItemsSortedByExpiry()
	for _, ki := range keyItems {
		if ki.Item.IsExpired() {
			r.store.Delete(ki.Key)
		} else {
			// Items are sorted by expiry time, so we can break early
			break
		}
	}
}

// New returns a new in-memory cache implementation.
func New(ctx context.Context, config *cache.Config, options Options) *ramcache {
	config.Revise()
	r := &ramcache{}
	r.init(ctx, config, options)
	return r
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
	item, exists := r.store.Get(key)
	if exists && item.IsExpired() {
		r.store.Delete(key)
		exists = false
	}
	return exists, nil
}

// Del implements cache.Cache.
func (r *ramcache) Del(ctx context.Context, key string, modifiers ...keymod.KeyModifier) error {
	key = keymod.ModifyKey(key, modifiers...)
	_, exists := r.store.Get(key)
	if !exists {
		return gcerrors.NewWithScheme(Scheme, fmt.Errorf("%s: %w", key, cache.ErrKeyNotFound))
	}
	r.store.Delete(key)
	return nil
}

// DelKeys implements cache.Cache.
func (r *ramcache) DelKeys(ctx context.Context, pattern string, modifiers ...keymod.KeyModifier) error {
	return gcerrors.NewWithScheme(Scheme, fmt.Errorf("DelKeys not supported: %w", cache.ErrPatternMatchingNotSupported))
}

// Clear implements cache.Cache.
func (r *ramcache) Clear(ctx context.Context) error {
	r.store.Clear()
	return nil
}

// Get implements cache.Cache.
func (r *ramcache) Get(ctx context.Context, key string, modifiers ...keymod.KeyModifier) ([]byte, error) {
	key = keymod.ModifyKey(key, modifiers...)
	item, exists := r.store.Get(key)
	if !exists || item.IsExpired() {
		r.store.Delete(key)
		return nil, gcerrors.NewWithScheme(Scheme, cache.ErrKeyNotFound)
	}
	return item.Value, nil
}

// Set implements cache.Cache.
func (r *ramcache) Set(ctx context.Context, key string, value interface{}, modifiers ...keymod.KeyModifier) error {
	key = keymod.ModifyKey(key, modifiers...)
	switch v := value.(type) {
	case string:
		r.store.Set(key, Item{Value: []byte(v), Expiry: time.Now().Add(r.opts.DefaultTTL)})
	case []byte:
		r.store.Set(key, Item{Value: v, Expiry: time.Now().Add(r.opts.DefaultTTL)})
	default:
		return gcerrors.NewWithScheme(Scheme, fmt.Errorf("unsupported value type: %T", v))
	}
	return nil
}

// SetWithExpiry implements cache.Cache.
func (r *ramcache) SetWithExpiry(ctx context.Context, key string, value interface{}, expiry time.Duration, modifiers ...keymod.KeyModifier) error {
	key = keymod.ModifyKey(key, modifiers...)
	switch v := value.(type) {
	case string:
		r.store.Set(key, Item{Value: []byte(v), Expiry: time.Now().Add(expiry)})
	case []byte:
		r.store.Set(key, Item{Value: v, Expiry: time.Now().Add(expiry)})
	default:
		return gcerrors.NewWithScheme(Scheme, fmt.Errorf("unsupported value type: %T", v))
	}
	return nil
}

// Close implements cache.Cache.
func (r *ramcache) Close() error {
	close(r.stopCh)
	return nil
}

// Ping implements cache.Cache.
func (r *ramcache) Ping(ctx context.Context) error {
	return nil
}
