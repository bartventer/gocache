/*
Package redis provides a Redis Client implementation of the cache.Cache interface.
It uses the go-redis library to interact with a Redis Client.

# URL Format:

The URL should have the following format:

	redis://<host>:<port>[?query]

The <host>:<port> pair corresponds to the Redis Client node. The [?query] part,
though optional, can be used for additional configuration through query parameters.

Query parameters can be used to configure the Redis Client options. The keys of the query
parameters should correspond to the case-insensitive field names of [redis.Options].
However, not all options can be set as query parameters. The following options are excluded:

  - [redis.Options].Addr
  - Any option that is a function
  - Any options defined in [cache.Options]

# Usage

Example via generic cache interface:

	import (
	    "context"
	    "log"
	    "net/url"

	    "github.com/bartventer/gocache"
	    _ "github.com/bartventer/gocache/redis"
	)

	func main() {
	    ctx := context.Background()
	    urlStr := "redis://localhost:7000?maxretries=5&minretrybackoff=1000"
	    c, err := cache.OpenCache(ctx, urlStr, cache.Options{})
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

	    "github.com/bartventer/gocache"
	    "github.com/bartventer/gocache/redis"
	    "github.com/redis/go-redis/v9"
	)

	func main() {
	    ctx := context.Background()
	    c := redis.New(ctx, cache.Config{}, redis.Options{
	        Addr: "localhost:7000",
	    })
	    // ... use c with the cache.Cache interface
	}
*/
package redis

import (
	"context"
	"net/url"
	"sync"
	"time"

	cache "github.com/bartventer/gocache"
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
	config *cache.Config // config is the cache configuration.
}

// New returns a new Redis cache implementation.
func New(ctx context.Context, config cache.Config, options redis.Options) *redisCache {
	r := &redisCache{}
	r.init(ctx, config, options)
	return r
}

// Ensure RedisCache implements the cache.Cache interface.
var _ cache.Cache = &redisCache{}

// OpenCacheURL opens a new Redis cache using the given URL and options.
// It implements the cache.CacheURLOpener interface.
func (r *redisCache) OpenCacheURL(ctx context.Context, u *url.URL, options cache.Options) (cache.Cache, error) {
	// Parse the URL into Redis options
	clusterOpts, err := optionsFromURL(u, options.ExtraParams)
	if err != nil {
		return nil, err
	}
	// Set configured options
	clusterOpts.TLSConfig = options.TLSConfig
	clusterOpts.CredentialsProviderContext = options.CredentialsProvider

	// Initialize the Redis client
	r.init(ctx, options.Config, clusterOpts)
	return r, nil
}

// init initializes the Redis client with the given options.
// It implements the cache.Cache interface.
func (r *redisCache) init(_ context.Context, config cache.Config, options redis.Options) {
	r.once.Do(func() {
		r.config = &config
		r.client = redis.NewClient(&options)
	})
}

// Count implements cache.Cache.
func (r *redisCache) Count(ctx context.Context, pattern string) (int64, error) {
	keys, _, err := r.client.Scan(ctx, 0, pattern, r.config.CountLimit).Result()
	if err != nil {
		return 0, err
	}
	return int64(len(keys)), nil
}

// Exists implements cache.Cache.
func (r *redisCache) Exists(ctx context.Context, key string) (bool, error) {
	n, err := r.client.Exists(ctx, key).Result()
	return n > 0, err
}

// Del deletes a key from the cache.
// It implements the cache.Cache interface.
func (r *redisCache) Del(ctx context.Context, key string) error {
	result := r.client.Del(ctx, key)
	if result.Err() != nil {
		return result.Err()
	}
	if result.Val() == 0 {
		return cache.ErrKeyNotFound
	}
	return nil
}

// DelKeys deletes all keys matching a pattern from the cache.
// It implements the cache.Cache interface.
func (r *redisCache) DelKeys(ctx context.Context, pattern string) error {
	var cursor uint64
	pipeline := r.client.Pipeline()

	for {
		var keys []string
		var err error

		keys, cursor, err = r.client.Scan(ctx, cursor, pattern, 10).Result()
		if err != nil {
			return err
		}

		for _, key := range keys {
			pipeline.Del(ctx, key)
		}

		if cursor == 0 {
			break
		}
	}

	_, err := pipeline.Exec(ctx)
	return err
}

// Clear deletes all keys from the cache.
// It implements the cache.Cache interface.
func (r *redisCache) Clear(ctx context.Context) error {
	return r.client.FlushDB(ctx).Err()
}

// Get gets the value of a key from the cache.
// It implements the cache.Cache interface.
func (r *redisCache) Get(ctx context.Context, key string) ([]byte, error) {
	return r.client.Get(ctx, key).Bytes()
}

// Set sets a key to a value in the cache.
// It implements the cache.Cache interface.
func (r *redisCache) Set(ctx context.Context, key string, value interface{}) error {
	return r.client.Set(ctx, key, value, 0).Err()
}

// SetWithExpiry sets a key to a value in the cache with an expiry time.
// It implements the cache.Cache interface.
func (r *redisCache) SetWithExpiry(ctx context.Context, key string, value interface{}, expiry time.Duration) error {
	return r.client.Set(ctx, key, value, expiry).Err()
}
