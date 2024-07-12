# Cache

[![Go Reference](https://pkg.go.dev/badge/github.com/bartventer/gocache.svg)](https://pkg.go.dev/github.com/bartventer/gocache)
[![Release](https://img.shields.io/github/release/bartventer/gocache.svg)](https://github.com/bartventer/gocache/releases/latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/bartventer/gocache)](https://goreportcard.com/report/github.com/bartventer/gocache)
[![codecov](https://codecov.io/gh/bartventer/gocache/graph/badge.svg?token=rtp2vxaccX)](https://codecov.io/gh/bartventer/gocache)
[![Test](https://github.com/bartventer/gocache/actions/workflows/default.yml/badge.svg)](https://github.com/bartventer/gocache/actions/workflows/default.yml)
![GitHub issues](https://img.shields.io/github/issues/bartventer/gocache)
[![License](https://img.shields.io/github/license/bartventer/gocache.svg)](LICENSE)

The `Cache` package in Go provides a unified, portable API for managing caches, enabling developers to write cache-related code once and transition seamlessly between different cache drivers with minimal reconfiguration. This approach simplifies both local testing and deployment to different environments.

## Installation

```bash
go get -u github.com/bartventer/gocache
```

## Supported Cache Implementations

| Name | Author | Docs |
|------|--------|------|
| Redis | [go-redis/redis](https://github.com/go-redis/redis) | [![Go Reference](https://pkg.go.dev/badge/github.com/bartventer/gocache/redis.svg)](https://pkg.go.dev/github.com/bartventer/gocache/redis) |
| Redis Cluster | [go-redis/redis](https://github.com/go-redis/redis) | [![Go Reference](https://pkg.go.dev/badge/github.com/bartventer/gocache/rediscluster.svg)](https://pkg.go.dev/github.com/bartventer/gocache/rediscluster) |
| Memcache | [bradfitz/gomemcache](https://github.com/bradfitz/gomemcache) | [![Go Reference](https://pkg.go.dev/badge/github.com/bartventer/gocache/memcache.svg)](https://pkg.go.dev/github.com/bartventer/gocache/memcache) |
| RAM Cache (in-memory) | [bartventer/gocache](https://github.com/bartventer/gocache) | [![Go Reference](https://pkg.go.dev/badge/github.com/bartventer/gocache/ramcache.svg)](https://pkg.go.dev/github.com/bartventer/gocache/ramcache) |

_**Note**: More coming soon!_

_See the [Contributing](#contributing) section if you would like to add a new cache implementation._

## Usage

To use a cache implementation, import the relevant driver package and use the `OpenCache` function to create a new cache. The cache package will automatically use the correct cache driver based on the URL scheme. Each driver also provides a constructor function for manual initialization.

### Redis

The [redis](https://pkg.go.dev/github.com/bartventer/gocache/redis) package provides a [Redis](https://redis.io) cache driver using the [go-redis/redis](https://github.com/go-redis/redis) client.

```go
import (
    "context"
    "log"

    cache "github.com/bartventer/gocache"
    _ "github.com/bartventer/gocache/redis"
)

func main() {
    ctx := context.Background()
    urlStr := "redis://localhost:7000?maxretries=5&minretrybackoff=1s"
    c, err := cache.OpenCache(ctx, urlStr)
    if err != nil {
        log.Fatalf("Failed to initialize cache: %v", err)
    }
    // ... use c with the cache.Cache interface
}
```

#### Redis Constructor

You can create a Redis cache with [redis.New](https://pkg.go.dev/github.com/bartventer/gocache/redis#New):

```go
import (
    "context"

    "github.com/bartventer/gocache/redis"
)

func main() {
    ctx := context.Background()
    c := redis.New(ctx, &redis.Options{
        RedisOptions: &redis.RedisOptions{
            Addr: "localhost:7000",
            MaxRetries: 5,
            MinRetryBackoff: 1 * time.Second,
        },
    })
    // ... use c with the cache.Cache interface
}
```

### Redis Cluster

The [rediscluster](https://pkg.go.dev/github.com/bartventer/gocache/rediscluster) package provides a [Redis Cluster](https://redis.io/topics/cluster-spec) cache driver using the [go-redis/redis](https://github.com/go-redis/redis) client.

```go
import (
    "context"
    "log"

    cache "github.com/bartventer/gocache"
    _ "github.com/bartventer/gocache/rediscluster"
)

func main() {
    ctx := context.Background()
    urlStr := "rediscluster://localhost:7000,localhost:7001,localhost:7002?maxretries=5&minretrybackoff=1s"
    c, err := cache.OpenCache(ctx, urlStr)
    if err != nil {
        log.Fatalf("Failed to initialize cache: %v", err)
    }
    // ... use c with the cache.Cache interface
}
```

#### Redis Cluster Constructor

You can create a Redis Cluster cache with [rediscluster.New](https://pkg.go.dev/github.com/bartventer/gocache/rediscluster#New):

```go
import (
    "context"

    "github.com/bartventer/gocache/rediscluster"
)

func main() {
    ctx := context.Background()
    c := rediscluster.New(ctx, &rediscluster.Options{
        ClusterOptions: &rediscluster.ClusterOptions{
            Addrs: []string{"localhost:7000", "localhost:7001", "localhost:7002"},
            MaxRetries: 5,
            MinRetryBackoff: 1 * time.Second,
        },
    })
    // ... use c with the cache.Cache interface
}
```

### Memcache

The [memcache](https://pkg.go.dev/github.com/bartventer/gocache/memcache) package provides a [Memcache](https://memcached.org) cache driver using the [bradfitz/gomemcache](https://github.com/bradfitz/gomemcache) client.

```go
import (
    "context"
    "log"

    cache "github.com/bartventer/gocache"
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
```

#### Memcache Constructor

You can create a Memcache cache with [memcache.New](https://pkg.go.dev/github.com/bartventer/gocache/memcache#New):

```go
import (
    "context"

    "github.com/bartventer/gocache/memcache"
)

func main() {
    ctx := context.Background()
    c := memcache.New(ctx, &memcache.Options{
        Addrs: []string{"localhost:11211"},
    })
    // ... use c with the cache.Cache interface
}
```

### RAM Cache (in-memory)

The [ramcache](https://pkg.go.dev/github.com/bartventer/gocache/ramcache) package provides an in-memory cache driver using a map.

```go
import (
    "context"
    "log"

    cache "github.com/bartventer/gocache"
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
```

#### RAM Cache Constructor

You can create a RAM cache with [ramcache.New](https://pkg.go.dev/github.com/bartventer/gocache/ramcache#New):

```go
import (
    "context"

    "github.com/bartventer/gocache/ramcache"
)

func main() {
    ctx := context.Background()
    c := ramcache.New(ctx, &ramcache.Options{
        CleanupInterval: 1 * time.Minute,
    })
    // ... use c with the cache.Cache interface
}
```

## Contributing

All contributions are welcome! See the [Contributing Guide](CONTRIBUTING.md) for more details.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.