package redis

// Options for the Redis cache.

import (
	"github.com/redis/go-redis/v9"
)

const (
	// DefaultCountLimit is the default value for the [Config.CountLimit] option.
	DefaultCountLimit = 10
)

// Config is a struct that holds configuration options for the cache package.
//
// # Compatibility
//
// These options are recognized by all cache drivers.
type Config struct {
	// CountLimit is the hint to the SCAN command about the amount of work to be done at each call.
	// The default value is 10.
	//
	// Refer to [redis scan] for more information.
	//
	// [redis scan]: https://redis.io/docs/latest/commands/scan/
	CountLimit int64
}

// revise revises the configuration options to ensure they contain sensible values.
func (c *Config) revise() {
	if c.CountLimit <= 0 {
		c.CountLimit = DefaultCountLimit
	}
}

// Options is the configuration for the Redis cache.
type Options struct {
	*Config
	redis.Options
}
