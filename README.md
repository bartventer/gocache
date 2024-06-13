# Cache

[![Go Reference](https://pkg.go.dev/badge/github.com/bartventer/gocache.svg)](https://pkg.go.dev/github.com/bartventer/gocache)
[![Release](https://img.shields.io/github/release/bartventer/gocache.svg)](https://github.com/bartventer/gocache/releases/latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/bartventer/gocache)](https://goreportcard.com/report/github.com/bartventer/gocache)
[![Coverage Status](https://coveralls.io/repos/github/bartventer/gocache/badge.svg?branch=master)](https://coveralls.io/github/bartventer/gocache?branch=master)
[![Build](https://github.com/bartventer/gocache/actions/workflows/default.yml/badge.svg)](https://github.com/bartventer/gocache/actions/workflows/default.yml)
![GitHub issues](https://img.shields.io/github/issues/bartventer/gocache)
[![License](https://img.shields.io/github/license/bartventer/gocache.svg)](LICENSE)

The `Cache` package provides a unified interface for managing caches in Go. It allows developers to switch between various cache implementations (such as Redis, Memcache, etc.) by simply altering the URL scheme.

## Installation

```bash
go get -u github.com/bartventer/cache
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

```go
import (
    "context"
    "log"

    cache "github.com/bartventer/gocache"
    // Enable the Redis cache implementation
    _ "github.com/bartventer/gocache/redis"
)

func main() {
    ctx := context.Background()
    urlStr := "redis://localhost:7000?maxretries=5&minretrybackoff=1000"
    c, err := cache.OpenCache(ctx, urlStr, cache.Options{})
    if err != nil {
        log.Fatalf("Failed to initialize cache: %v", err)
    }

    // Now you can use c with the cache.Cache interface
    err = c.Set(ctx, "key", "value")
    if err != nil {
        log.Fatalf("Failed to set key: %v", err)
    }

    value, err := c.Get(ctx, "key")
    if err != nil {
        log.Fatalf("Failed to get key: %v", err)
    }

    log.Printf("Value: %s", value)
}
```

## Contributing

All contributions are welcome! Open a pull request to request a feature or submit a bug report.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.