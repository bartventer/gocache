package cache

import "errors"

var (
	// ErrNoCache is returned when no cache implementation is available.
	ErrNoCache = errors.New("gocache: no cache implementation available")

	// ErrKeyNotFound is returned when a key is not found in the cache.
	ErrKeyNotFound = errors.New("gocache: key not found")

	// ErrPatternMatchingNotSupported is returned when a pattern matching operation is not supported
	// by the cache implementation.
	ErrPatternMatchingNotSupported = errors.New("gocache: pattern matching not supported")

	// ErrInvalidTTL is returned when an invalid TTL is provided.
	ErrInvalidTTL = errors.New("gocache: invalid TTL")
)
