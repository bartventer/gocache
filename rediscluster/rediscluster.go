/*
Package rediscluster provides a Redis Cluster implementation of the [driver.Cache] interface.
It uses the go-redis library to interact with a Redis Cluster.

# URL Format:

The URL should have the following format:

	rediscluster://<host1>:<port1>,<host2>:<port2>,...,<hostN>:<portN>[?query]

Each <host>:<port> pair corresponds to a Redis Cluster node. You can specify any number
of nodes, each separated by a comma.

The optional query part can be used to configure the Redis Cluster options through
query parameters. The keys of the query parameters should match the case-insensitive
field names of the [Options] structure (excluding [redis.ClusterOptions.Addrs]).

# Usage

	import (
		"context"
		"log"
		"net/url"

		"github.com/bartventer/gocache"
		_ "github.com/bartventer/gocache/rediscluster"
	)

	func main() {
		ctx := context.Background()
		urlStr := "rediscluster://localhost:7000,localhost:7001,localhost:7002?maxretries=5&minretrybackoff=1000ms"
		c, err := cache.OpenCache(ctx, urlStr)
		if err != nil {
			log.Fatalf("Failed to initialize cache: %v", err)
		}
		// ... use c with the cache.Cache interface
	}

You can create a Redis Cluster cache with [New]:

	import (
		"context"
		"log"
		"net/url"

		"github.com/bartventer/gocache/rediscluster"
	)

	func main() {
		ctx := context.Background()
		c := rediscluster.New[string](ctx, &rediscluster.Options{
			ClusterOptions: rediscluster.ClusterOptions{
				Addrs: []string{"localhost:7000", "localhost:7001", "localhost:7002"},
				MaxRetries: 5,
				MinRetryBackoff: 1000 * time.Millisecond,
			},
		})
		// ... use c with the cache.Cache interface
	}
*/
package rediscluster

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

// Scheme is the cache scheme for Redis Cluster.
const Scheme = "rediscluster"

func init() { //nolint:gochecknoinits // This is the entry point of the package.
	cache.RegisterCache(Scheme, &redisClusterCache[string]{})
	cache.RegisterCache(Scheme, &redisClusterCache[keymod.Key]{})
}

// redisClusterCache is a Redis Cluster implementation of the cache.Cache interface.
type redisClusterCache[K driver.String] struct {
	once   sync.Once            // once ensures that the cache is initialized only once.
	client *redis.ClusterClient // client is the Redis Cluster client.
	config *Config              // config is the cache configuration.
}

// New returns a new Redis Cluster cache implementation.
func New[K driver.String](ctx context.Context, opts *Options) *redisClusterCache[K] {
	r := &redisClusterCache[K]{}
	if opts == nil {
		opts = &Options{}
	}
	r.init(ctx, opts.Config, &opts.ClusterOptions)
	return r
}

// Ensure RedisClusterCache implements the cache.Cache interface.
var _ driver.Cache[string] = new(redisClusterCache[string])
var _ driver.Cache[keymod.Key] = new(redisClusterCache[keymod.Key])

// OptionsFromURL implements cache.URLOpener.
func (r *redisClusterCache[K]) OpenCacheURL(ctx context.Context, u *url.URL) (*cache.GenericCache[K], error) {
	opts, err := optionsFromURL(u)
	if err != nil {
		return nil, gcerrors.NewWithScheme(Scheme, fmt.Errorf("error parsing URL: %w", err))
	}
	r.init(ctx, opts.Config, &opts.ClusterOptions)
	return cache.NewCache(r), nil
}

func (r *redisClusterCache[K]) init(_ context.Context, config *Config, options *redis.ClusterOptions) {
	r.once.Do(func() {
		if config == nil {
			config = &Config{}
		}
		config.revise()
		r.config = config
		r.client = redis.NewClusterClient(options)
	})
}

// Count implements cache.Cache.
func (r *redisClusterCache[K]) Count(ctx context.Context, pattern K) (int64, error) {
	var count int64
	err := r.client.ForEachMaster(ctx, func(ctx context.Context, client *redis.Client) error {
		iter := client.Scan(ctx, 0, string(pattern), r.config.CountLimit).Iterator()
		for iter.Next(ctx) {
			count++
		}
		if err := iter.Err(); err != nil {
			return gcerrors.NewWithScheme(Scheme, fmt.Errorf("error counting keys: %w", err))
		}
		return nil
	})

	if err != nil {
		return 0, gcerrors.NewWithScheme(Scheme, fmt.Errorf("error counting keys: %w", err))
	}

	return count, nil
}

// Exists implements cache.Cache.
func (r *redisClusterCache[K]) Exists(ctx context.Context, key K) (bool, error) {
	n, err := r.client.Exists(ctx, string(key)).Result()
	if err != nil {
		return false, gcerrors.NewWithScheme(Scheme, fmt.Errorf("error checking key %s: %w", key, err))
	}
	return n > 0, nil
}

// Del implements cache.Cache.
func (r *redisClusterCache[K]) Del(ctx context.Context, key K) error {
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
func (r *redisClusterCache[K]) DelKeys(ctx context.Context, pattern K) error {
	return r.client.ForEachMaster(ctx, func(ctx context.Context, client *redis.Client) error {
		iter := client.Scan(ctx, 0, string(pattern), r.config.CountLimit).Iterator()
		var keys []string
		for iter.Next(ctx) {
			keys = append(keys, iter.Val())
		}
		if err := iter.Err(); err != nil {
			return gcerrors.NewWithScheme(Scheme, fmt.Errorf("error scanning keys: %w", err))
		}
		if len(keys) > 0 {
			_, err := client.Del(ctx, keys...).Result()
			if err != nil {
				return gcerrors.NewWithScheme(Scheme, fmt.Errorf("error deleting keys: %w", err))
			}
		}
		return nil
	})
}

// Clear implements cache.Cache.
func (r *redisClusterCache[K]) Clear(ctx context.Context) error {
	return r.client.ForEachMaster(ctx, func(ctx context.Context, client *redis.Client) error {
		return client.FlushAll(ctx).Err()
	})
}

// Get implements cache.Cache.
func (r *redisClusterCache[K]) Get(ctx context.Context, key K) ([]byte, error) {
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
func (r *redisClusterCache[K]) Set(ctx context.Context, key K, value interface{}) error {
	return r.client.Set(ctx, string(key), value, 0).Err()
}

// SetWithTTL implements cache.Cache.
func (r *redisClusterCache[K]) SetWithTTL(ctx context.Context, key K, value interface{}, ttl time.Duration) error {
	return r.client.Set(ctx, string(key), value, ttl).Err()
}

// Ping implements cache.Cache.
func (r *redisClusterCache[K]) Ping(ctx context.Context) error {
	return r.client.ForEachShard(ctx, func(ctx context.Context, client *redis.Client) error {
		return client.Ping(ctx).Err()
	})
}

// Close implements cache.Cache.
func (r *redisClusterCache[K]) Close() error {
	return r.client.Close()
}
