/*
Package redis provides a Redis Client implementation of the [cache.Cache] interface.
It uses the go-redis library to interact with a Redis Client.

# URL Format:

The URL should have the following format:

	redis://<host>:<port>[?query]

The <host>:<port> pair corresponds to the Redis Client node. The [?query] part,
though optional, can be used for additional configuration through query parameters.

Query parameters can be used to configure the Redis Client options. The keys of the query
parameters should correspond to the case-insensitive field names of [redis.Options].
However, not all options can be set as query parameters. The following options are excluded:

  - [redis.Options.Addr]
  - Any option that is a function

# Usage

Example via generic cache interface:

	import (
	    "context"
	    "log"
	    "net/url"

	    cache "github.com/bartventer/gocache"
	    _ "github.com/bartventer/gocache/redis"
	)

	func main() {
	    ctx := context.Background()
	    urlStr := "redis://localhost:7000?maxretries=5&minretrybackoff=1000ms"
	    c, err := cache.OpenCache(ctx, urlStr)
	    if err != nil {
	        log.Fatalf("Failed to initialize cache: %v", err)
	    }
	    // ... use c with the cache.Cache interface
	}

Example via [redis.New] constructor:

	import (
	    "context"
	    "log"
	    "net/url"

	    "github.com/bartventer/gocache/redis"
	)

	func main() {
	    ctx := context.Background()
	    c := redis.New(ctx, &redis.Options{
	        RedisOptions: redis.Options{
				Addr: "localhost:6379",
				MaxRetries: 5,
				MinRetryBackoff: 1000 * time.Millisecond,
			},
	    })
	    // ... use c with the cache.Cache interface
	}
*/
package redis

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"time"

	cache "github.com/bartventer/gocache"
	"github.com/bartventer/gocache/internal/gcerrors"
	"github.com/bartventer/gocache/keymod"
	"github.com/redis/go-redis/v9"
)

// Scheme is the cache scheme for Redis.
const Scheme = "redis"

func init() { //nolint:gochecknoinits // This is the entry point of the package.
	cache.RegisterCache(Scheme, &redisCache{})
}

// redisCache is a Redis implementation of the cache.Cache interface.
type redisCache struct {
	once   sync.Once     // once ensures that the cache is initialized only once.
	client *redis.Client // client is the Redis client.
	config *Config       // config is the cache configuration.
}

// New returns a new Redis cache implementation.
func New(ctx context.Context, opts *Options) *redisCache {
	r := &redisCache{}
	if opts == nil {
		opts = &Options{}
	}
	r.init(ctx, opts.Config, &opts.RedisOptions)
	return r
}

// Ensure RedisCache implements the cache.Cache interface.
var _ cache.Cache = &redisCache{}

// OpenCacheURL implements cache.URLOpener.
func (r *redisCache) OpenCacheURL(ctx context.Context, u *url.URL) (cache.Cache, error) {
	opts, err := optionsFromURL(u)
	if err != nil {
		return nil, gcerrors.NewWithScheme(Scheme, fmt.Errorf("error parsing URL: %w", err))
	}
	r.init(ctx, opts.Config, &opts.RedisOptions)
	return r, nil
}

func (r *redisCache) init(_ context.Context, config *Config, options *redis.Options) {
	r.once.Do(func() {
		if config == nil {
			config = &Config{}
		}
		config.revise()
		r.config = config
		r.client = redis.NewClient(options)
	})
}

// Count implements cache.Cache.
func (r *redisCache) Count(ctx context.Context, pattern string, modifiers ...keymod.Mod) (int64, error) {
	pattern = keymod.Modify(pattern, modifiers...)
	var count int64
	iter := r.client.Scan(ctx, 0, pattern, r.config.CountLimit).Iterator()
	for iter.Next(ctx) {
		count++
	}
	if err := iter.Err(); err != nil {
		return 0, gcerrors.NewWithScheme(Scheme, fmt.Errorf("error counting keys: %w", err))
	}
	return count, nil
}

// Exists implements cache.Cache.
func (r *redisCache) Exists(ctx context.Context, key string, modifiers ...keymod.Mod) (bool, error) {
	key = keymod.Modify(key, modifiers...)
	n, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, gcerrors.NewWithScheme(Scheme, fmt.Errorf("error checking key %s: %w", key, err))
	}
	return n > 0, nil
}

// Del implements cache.Cache.
func (r *redisCache) Del(ctx context.Context, key string, modifiers ...keymod.Mod) error {
	key = keymod.Modify(key, modifiers...)
	delCount, err := r.client.Del(ctx, key).Result()
	if err != nil {
		return gcerrors.NewWithScheme(Scheme, fmt.Errorf("error deleting key %s: %w", key, err))
	}
	if delCount == 0 {
		return gcerrors.NewWithScheme(Scheme, fmt.Errorf("%s: %w", key, cache.ErrKeyNotFound))
	}
	return nil
}

// DelKeys implements cache.Cache.
func (r *redisCache) DelKeys(ctx context.Context, pattern string, modifiers ...keymod.Mod) error {
	pattern = keymod.Modify(pattern, modifiers...)
	iter := r.client.Scan(ctx, 0, pattern, r.config.CountLimit).Iterator()
	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return gcerrors.NewWithScheme(Scheme, fmt.Errorf("error scanning keys: %w", err))
	}
	if len(keys) > 0 {
		_, err := r.client.Del(ctx, keys...).Result()
		if err != nil {
			return gcerrors.NewWithScheme(Scheme, fmt.Errorf("error deleting keys: %w", err))
		}
	}
	return nil
}

// Clear implements cache.Cache.
func (r *redisCache) Clear(ctx context.Context) error {
	return r.client.FlushDB(ctx).Err()
}

// Get implements cache.Cache.
func (r *redisCache) Get(ctx context.Context, key string, modifiers ...keymod.Mod) ([]byte, error) {
	key = keymod.Modify(key, modifiers...)
	val, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, gcerrors.NewWithScheme(Scheme, fmt.Errorf("%s: %w, underlying error: %w", key, cache.ErrKeyNotFound, err))
		} else {
			return nil, gcerrors.NewWithScheme(Scheme, fmt.Errorf("error getting key %s: %w", key, err))
		}
	}
	return val, nil
}

// Set implements cache.Cache.
func (r *redisCache) Set(ctx context.Context, key string, value interface{}, modifiers ...keymod.Mod) error {
	key = keymod.Modify(key, modifiers...)
	return r.client.Set(ctx, key, value, 0).Err()
}

// SetWithExpiry implements cache.Cache.
func (r *redisCache) SetWithExpiry(ctx context.Context, key string, value interface{}, expiry time.Duration, modifiers ...keymod.Mod) error {
	key = keymod.Modify(key, modifiers...)
	return r.client.Set(ctx, key, value, expiry).Err()
}

// Close implements cache.Cache.
func (r *redisCache) Close() error {
	return r.client.Close()
}

// Ping implements cache.Cache.
func (r *redisCache) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}
