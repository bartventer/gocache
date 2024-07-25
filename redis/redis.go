/*
Package redis provides a Redis Client implementation of the [driver.Cache] interface.
It uses the go-redis library to interact with a Redis Client.

# URL Format:

The URL should have the following format:

	redis://<host>:<port>[?query]

The <host>:<port> pair corresponds to the Redis Client node.

The optional query part can be used to configure the Redis Client options through
query parameters. The keys of the query parameters should match the case-insensitive
field names of the [Options] structure (excluding [redis.Options.Addr]).

# Usage

	import (
	    "context"
	    "log"

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

You can create a Redis cache with [New]:

	import (
	    "context"

	    "github.com/bartventer/gocache/redis"
	)

	func main() {
	    ctx := context.Background()
	    c := redis.New[string](ctx, &redis.Options{
	        RedisOptions: redis.RedisOptions{
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
	"errors"
	"fmt"
	"net/url"
	"sync"
	"time"

	cache "github.com/bartventer/gocache"
	"github.com/bartventer/gocache/internal/gcerrors"
	"github.com/bartventer/gocache/pkg/driver"
	"github.com/bartventer/gocache/pkg/keymod"
	"github.com/redis/go-redis/v9"
)

// Scheme is the cache scheme for Redis.
const Scheme = "redis"

func init() { //nolint:gochecknoinits // This is the entry point of the package.
	cache.RegisterCache(Scheme, &redisCache[string]{})
	cache.RegisterCache(Scheme, &redisCache[keymod.Key]{})
}

// redisCache is a Redis implementation of the cache.Cache interface.
type redisCache[K driver.String] struct {
	once   sync.Once     // once ensures that the cache is initialized only once.
	client *redis.Client // client is the Redis client.
	config *Config       // config is the cache configuration.
}

// New returns a new Redis cache implementation.
func New[K driver.String](ctx context.Context, opts *Options) *redisCache[K] {
	r := &redisCache[K]{}
	if opts == nil {
		opts = &Options{}
	}
	r.init(ctx, opts.Config, &opts.RedisOptions)
	return r
}

// Ensure RedisCache implements the cache.Cache interface.
var _ driver.Cache[string] = new(redisCache[string])
var _ driver.Cache[keymod.Key] = new(redisCache[keymod.Key])

// OpenCacheURL implements [cache.URLOpener].
func (r *redisCache[K]) OpenCacheURL(ctx context.Context, u *url.URL) (*cache.GenericCache[K], error) {
	opts, err := optionsFromURL(u)
	if err != nil {
		return nil, gcerrors.NewWithScheme(Scheme, fmt.Errorf("error parsing URL: %w", err))
	}
	r.init(ctx, opts.Config, &opts.RedisOptions)
	return cache.NewCache(r), nil
}

func (r *redisCache[K]) init(_ context.Context, config *Config, options *redis.Options) {
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
func (r *redisCache[K]) Count(ctx context.Context, pattern K) (int64, error) {
	var count int64
	iter := r.client.Scan(ctx, 0, string(pattern), r.config.CountLimit).Iterator()
	for iter.Next(ctx) {
		count++
	}
	if err := iter.Err(); err != nil {
		return 0, gcerrors.NewWithScheme(Scheme, fmt.Errorf("error counting keys: %w", err))
	}
	return count, nil
}

// Exists implements cache.Cache.
func (r *redisCache[K]) Exists(ctx context.Context, key K) (bool, error) {
	n, err := r.client.Exists(ctx, string(key)).Result()
	if err != nil {
		return false, gcerrors.NewWithScheme(Scheme, fmt.Errorf("error checking key %s: %w", key, err))
	}
	return n > 0, nil
}

// Del implements cache.Cache.
func (r *redisCache[K]) Del(ctx context.Context, key K) error {
	delCount, err := r.client.Del(ctx, string(key)).Result()
	if err != nil {
		return gcerrors.NewWithScheme(Scheme, fmt.Errorf("error deleting key %s: %w", key, err))
	}
	if delCount == 0 {
		return gcerrors.NewWithScheme(Scheme, errors.Join(cache.ErrKeyNotFound, fmt.Errorf("key %s not found", key)))
	}
	return nil
}

// DelKeys implements cache.Cache.
func (r *redisCache[K]) DelKeys(ctx context.Context, pattern K) error {
	iter := r.client.Scan(ctx, 0, string(pattern), r.config.CountLimit).Iterator()
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
func (r *redisCache[K]) Clear(ctx context.Context) error {
	return r.client.FlushDB(ctx).Err()
}

// Get implements cache.Cache.
func (r *redisCache[K]) Get(ctx context.Context, key K) ([]byte, error) {
	val, err := r.client.Get(ctx, string(key)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, gcerrors.NewWithScheme(Scheme, errors.Join(cache.ErrKeyNotFound, fmt.Errorf("key %s not found: %w", key, err)))
		} else {
			return nil, gcerrors.NewWithScheme(Scheme, fmt.Errorf("error getting key %s: %w", key, err))
		}
	}
	return val, nil
}

// Set implements cache.Cache.
func (r *redisCache[K]) Set(ctx context.Context, key K, value interface{}) error {
	return r.client.Set(ctx, string(key), value, 0).Err()
}

// SetWithTTL implements cache.Cache.
func (r *redisCache[K]) SetWithTTL(ctx context.Context, key K, value interface{}, ttl time.Duration) error {
	return r.client.Set(ctx, string(key), value, ttl).Err()
}

// Close implements cache.Cache.
func (r *redisCache[K]) Close() error {
	return r.client.Close()
}

// Ping implements cache.Cache.
func (r *redisCache[K]) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}
