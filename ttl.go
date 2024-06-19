package cache

import "time"

// ValidateTTL validates the TTL and returns an error if it's invalid.
// A TTL is invalid if it's negative.
func ValidateTTL(ttl time.Duration) error {
	if ttl < 0 {
		return ErrInvalidTTL
	}
	return nil
}
