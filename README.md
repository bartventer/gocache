# Cache

[![Go Reference](https://pkg.go.dev/badge/github.com/bartventer/gocache.svg)](https://pkg.go.dev/github.com/bartventer/gocache)
[![Release](https://img.shields.io/github/release/bartventer/gocache.svg)](https://github.com/bartventer/gocache/releases/latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/bartventer/gocache)](https://goreportcard.com/report/github.com/bartventer/gocache)
[![Coverage Status](https://coveralls.io/repos/github/bartventer/gocache/badge.svg?branch=master)](https://coveralls.io/github/bartventer/gocache?branch=master)
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
    c, err := cache.OpenCache(ctx, urlStr, cache.Options{})
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
    "log"

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
    c, err := cache.OpenCache(ctx, urlStr, cache.Options{})
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
    "log"

    "github.com/bartventer/gocache"
    "github.com/bartventer/gocache/rediscluster"
    "github.com/redis/go-redis/v9"
)

func main() {
    ctx := context.Background()
    c := rediscluster.New(ctx, cache.Config{}, redis.ClusterOptions{
        Addrs: []string{"localhost:7000", "localhost:7001", "localhost:7002"},
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
    c, err := cache.OpenCache(ctx, urlStr, cache.Options{})
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
    "log"

    "github.com/bartventer/gocache"
    "github.com/bartventer/gocache/memcache"
)

func main() {
    ctx := context.Background()
    c := memcache.New(ctx, cache.Config{}, "localhost:11211")
    // ... use c with the cache.Cache interface
}
```

## Contributing

All contributions are welcome! Open a pull request to request a feature or submit a bug report.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.