/*
Package ramcache implements the [driver.Cache] interface using an in-memory map.

It's useful for testing, development, and caching small data sets. It's not recommended
for production due to lack of data persistence across restarts.

# URL Format

The URL should have the following format:

	ramcache://[?query]

The optional query part can be used to configure the in-memory cache options through
query parameters. The keys of the query parameters should match the case-insensitive
field names of the [Options] structure.

# Value Types

Values being set in the cache should be of type [][byte], [string], or implement one
of the following interfaces:
  - [encoding.BinaryMarshaler]
  - [encoding.TextMarshaler]
  - [json.Marshaler]
  - [fmt.Stringer]
  - [io.Reader]

# Usage

	import (
	    "context"
	    "log"

	    "github.com/bartventer/gocache"
	    _ "github.com/bartventer/gocache/ramcache"
	)

	func main() {
	    ctx := context.Background()
		urlStr := "ramcache://?cleanupinterval=1m"
	    c, err := cache.OpenCache(ctx, urlStr)
	    if err != nil {
	        log.Fatalf("Failed to initialize cache: %v", err)
	    }
	    // ... use c with the cache.Cache interface
	}

You can create a RAM cache with [New]:

	import (
	    "context"

	    "github.com/bartventer/gocache/ramcache"
	)

	func main() {
	    ctx := context.Background()
	    c := ramcache.New[string](ctx, &ramcache.Options{
			CleanupInterval: 1 * time.Minute,
		})
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
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"sync"
	"time"

	cache "github.com/bartventer/gocache"
	"github.com/bartventer/gocache/internal/gcerrors"
	"github.com/bartventer/gocache/pkg/driver"
	"github.com/bartventer/gocache/pkg/keymod"
)

// Scheme is the cache scheme for the in-memory cache.
const Scheme = "ramcache"

func init() { //nolint:gochecknoinits // This is the entry point of the package.
	cache.RegisterCache(Scheme, &ramcache[string]{})
	cache.RegisterCache(Scheme, &ramcache[keymod.Key]{})
}

var _ driver.Cache[string] = new(ramcache[string])
var _ driver.Cache[keymod.Key] = new(ramcache[keymod.Key])

// ramcache is an in-memory implementation of the cache.Cache interface.
type ramcache[K driver.String] struct {
	once   sync.Once     // once ensures that the cache is initialized only once.
	store  *store        // store is the in-memory store.
	opts   *Options      // options is the cache options.
	stopCh chan struct{} // stopCh is the stop channel.
}

// New returns a new in-memory cache implementation.
func New[K driver.String](ctx context.Context, opts *Options) *ramcache[K] {
	r := &ramcache[K]{}
	r.init(ctx, opts)
	return r
}

// OpenCacheURL implements cache.URLOpener.
func (r *ramcache[K]) OpenCacheURL(ctx context.Context, u *url.URL) (*cache.GenericCache[K], error) {
	opts, err := optionsFromURL(u)
	if err != nil {
		return nil, gcerrors.NewWithScheme(Scheme, fmt.Errorf("failed to parse URL: %w", err))
	}
	r.init(ctx, &opts)
	return cache.NewCache(r), nil
}

func (r *ramcache[K]) init(_ context.Context, opts *Options) {
	r.once.Do(func() {
		r.store = newStore()
		if opts == nil {
			opts = &Options{}
		}
		opts.revise()
		r.opts = opts
		r.stopCh = make(chan struct{})
		go r.cleanupExpiredItems()
	})
}

// cleanupExpiredItems periodically removes expired items from the store.
func (r *ramcache[K]) cleanupExpiredItems() {
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
func (r *ramcache[K]) removeExpiredItems() {
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

// Count implements cache.Cache.
func (r *ramcache[K]) Count(ctx context.Context, pattern K) (int64, error) {
	return 0, gcerrors.NewWithScheme(Scheme, errors.Join(cache.ErrPatternMatchingNotSupported, fmt.Errorf("Count operation not supported")))
}

// Exists implements cache.Cache.
func (r *ramcache[K]) Exists(ctx context.Context, key K) (bool, error) {
	item, exists := r.store.Get(string(key))
	if exists && item.IsExpired() {
		r.store.Delete(string(key))
		exists = false
	}
	return exists, nil
}

// Del implements cache.Cache.
func (r *ramcache[K]) Del(ctx context.Context, key K) error {
	_, exists := r.store.Get(string(key))
	if !exists {
		return gcerrors.NewWithScheme(Scheme, errors.Join(cache.ErrKeyNotFound, fmt.Errorf("key %s not found", key)))
	}
	r.store.Delete(string(key))
	return nil
}

// DelKeys implements cache.Cache.
func (r *ramcache[K]) DelKeys(ctx context.Context, pattern K) error {
	return gcerrors.NewWithScheme(Scheme, errors.Join(cache.ErrPatternMatchingNotSupported, fmt.Errorf("pattern %s not supported", pattern)))
}

// Clear implements cache.Cache.
func (r *ramcache[K]) Clear(ctx context.Context) error {
	r.store.Clear()
	return nil
}

// Get implements cache.Cache.
func (r *ramcache[K]) Get(ctx context.Context, key K) ([]byte, error) {
	item, exists := r.store.Get(string(key))
	if !exists || item.IsExpired() {
		r.store.Delete(string(key))
		return nil, gcerrors.NewWithScheme(Scheme, errors.Join(cache.ErrKeyNotFound, fmt.Errorf("key %s not found", key)))
	}
	return item.Value, nil
}

// Set implements cache.Cache.
func (r *ramcache[K]) Set(ctx context.Context, key K, value interface{}) error {
	return r.set(key, value, 0)
}

// SetWithTTL implements cache.Cache.
func (r *ramcache[K]) SetWithTTL(ctx context.Context, key K, value interface{}, ttl time.Duration) error {
	if err := cache.ValidateTTL(ttl); err != nil {
		return gcerrors.NewWithScheme(Scheme, fmt.Errorf("invalid expiry duration %q: %w", ttl, err))
	}
	return r.set(key, value, ttl)
}

func (r *ramcache[K]) set(key K, value interface{}, expiry time.Duration) error {
	var data []byte
	switch v := value.(type) {
	case string:
		data = []byte(v)
	case []byte:
		data = v
	case encoding.BinaryMarshaler:
		var err error
		data, err = v.MarshalBinary()
		if err != nil {
			return gcerrors.NewWithScheme(Scheme, fmt.Errorf("failed to marshal value: %w", err))
		}
	case encoding.TextMarshaler:
		var err error
		data, err = v.MarshalText()
		if err != nil {
			return gcerrors.NewWithScheme(Scheme, fmt.Errorf("failed to marshal value: %w", err))
		}
	case json.Marshaler:
		var err error
		data, err = v.MarshalJSON()
		if err != nil {
			return gcerrors.NewWithScheme(Scheme, fmt.Errorf("failed to marshal value: %w", err))
		}
	case fmt.Stringer:
		data = []byte(v.String())
	case io.Reader:
		var err error
		data, err = io.ReadAll(v)
		if err != nil {
			return gcerrors.NewWithScheme(Scheme, fmt.Errorf("failed to read value: %w", err))
		}
	default:
		return gcerrors.NewWithScheme(Scheme, fmt.Errorf("unsupported value type: %T", v))
	}

	var expiryTime time.Time
	if expiry != 0 {
		expiryTime = time.Now().Add(expiry)
	}

	r.store.Set(string(key), item{Value: data, Expiry: expiryTime})
	return nil
}

// Close implements cache.Cache.
func (r *ramcache[K]) Close() error {
	close(r.stopCh)
	return nil
}

// Ping implements cache.Cache.
func (r *ramcache[K]) Ping(_ context.Context) error {
	return nil
}
