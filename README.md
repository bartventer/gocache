# Cache

[![Go Reference](https://pkg.go.dev/badge/github.com/bartventer/gocache.svg)](https://pkg.go.dev/github.com/bartventer/gocache)
[![Release](https://img.shields.io/github/release/bartventer/gocache.svg)](https://github.com/bartventer/gocache/releases/latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/bartventer/gocache)](https://goreportcard.com/report/github.com/bartventer/gocache)
[![codecov](https://codecov.io/gh/bartventer/gocache/graph/badge.svg?token=rtp2vxaccX)](https://codecov.io/gh/bartventer/gocache)
[![Test](https://github.com/bartventer/gocache/actions/workflows/default.yml/badge.svg)](https://github.com/bartventer/gocache/actions/workflows/default.yml)
![GitHub issues](https://img.shields.io/github/issues/bartventer/gocache)
[![License](https://img.shields.io/github/license/bartventer/gocache.svg)](LICENSE)

The `Cache` package provides a unified interface for managing caches in Go. It allows developers to switch between various cache implementations (such as Redis, Memcache, etc.) by simply altering the URL scheme.

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

_Pull requests for additional cache implementations are welcome!_

## Usage

To use a cache implementation, import the relevant driver package and use the `OpenCache` function to create a new cache. The cache package will automatically use the correct cache implementation based on the URL scheme.

### Redis

The [redis](https://pkg.go.dev/github.com/bartventer/gocache/redis) package provides a [Redis](https://redis.io) cache implementation using the [go-redis/redis](https://github.com/go-redis/redis) client.

```go
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
```

#### Redis Constructor

```go
import (
    "context"

    "github.com/bartventer/gocache/redis"
)

func main() {
    ctx := context.Background()
    c := redis.New(ctx, &redis.Options{
        RedisOptions: &redis.Options{
            Addr: "localhost:7000",
            MaxRetries: 5,
            MinRetryBackoff: 1000 * time.Millisecond,
        },
    })
    // ... use c with the cache.Cache interface
}
```

### Redis Cluster

The [rediscluster](https://pkg.go.dev/github.com/bartventer/gocache/rediscluster) package provides a [Redis Cluster](https://redis.io/topics/cluster-spec) cache implementation using the [go-redis/redis](https://github.com/go-redis/redis) client.

```go
import (
    "context"
    "log"

    cache "github.com/bartventer/gocache"
    _ "github.com/bartventer/gocache/rediscluster"
)

func main() {
    ctx := context.Background()
    urlStr := "rediscluster://localhost:7000,localhost:7001,localhost:7002?maxretries=5&minretrybackoff=1000"
    c, err := cache.OpenCache(ctx, urlStr)
    if err != nil {
        log.Fatalf("Failed to initialize cache: %v", err)
    }
    // ... use c with the cache.Cache interface
}
```

#### Redis Cluster Constructor

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
            MinRetryBackoff: 1000 * time.Millisecond,
        },
    })
    // ... use c with the cache.Cache interface
}
```

### Memcache

The [memcache](https://pkg.go.dev/github.com/bartventer/gocache/memcache) package provides a [Memcache](https://memcached.org) cache implementation using the [bradfitz/gomemcache](https://github.com/bradfitz/gomemcache) client.

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

The [ramcache](https://pkg.go.dev/github.com/bartventer/gocache/ramcache) package provides an in-memory cache implementation using a map.

```go
import (
    "context"
    "log"

    cache "github.com/bartventer/gocache"
    _ "github.com/bartventer/gocache/ramcache"
)

func main() {
    ctx := context.Background()
    urlStr := "ramcache://?defaultttl=5m"
    c, err := cache.OpenCache(ctx, urlStr)
    if err != nil {
        log.Fatalf("Failed to initialize cache: %v", err)
    }
    // ... use c with the cache.Cache interface
}
```

#### RAM Cache Constructor

```go
import (
    "context"

    "github.com/bartventer/gocache/ramcache"
)

func main() {
    ctx := context.Background()
    c := ramcache.New(ctx, &ramcache.Options{
        DefaultTTL: 5 * time.Minute,
    })
    // ... use c with the cache.Cache interface
}
```

## Contributing

All contributions are welcome! Open a pull request to request a feature or submit a bug report.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Acknowledgements

This package has been inspired by and uses similar patterns to those used in [Google's Go Cloud Development Kit](https://github.com/google/go-cloud). Make sure to [check it out](https://gocloud.dev/)!