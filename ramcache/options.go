package ramcache

import "time"

// Options are the configuration options for the RAM cache.
type Options struct {
	// DefaultTTL is the default time-to-live for cache items.
	// If not set, the default is 24 hours.
	DefaultTTL time.Duration
	// CleanupInterval is the interval at which checks for expired items are performed.
	// If not set, the default is 5 minutes.
	CleanupInterval time.Duration
}

// revise revises the options, ensuring sensible defaults are set.
func (r *Options) revise() {
	if r.DefaultTTL <= 0 {
		r.DefaultTTL = 24 * time.Hour
	}
	if r.CleanupInterval <= 0 {
		r.CleanupInterval = 5 * time.Minute
	}
}
