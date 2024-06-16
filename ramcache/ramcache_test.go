package ramcache

import (
	"context"
	"net/url"
	"testing"
	"time"

	cache "github.com/bartventer/gocache"
	"github.com/bartventer/gocache/drivertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRamcacheCache_OpenCacheURL(t *testing.T) {
	t.Parallel()
	r := &ramcache{}
	u, err := url.Parse("ramcache://?defaultttl=1h")
	require.NoError(t, err)

	_, err = r.OpenCacheURL(context.Background(), u, &cache.Options{})
	require.NoError(t, err)
	assert.NotNil(t, r.store)
}

func TestRamcacheCache_New(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	config := cache.Config{}

	r := New(ctx, &config, Options{})
	require.NotNil(t, r)
	assert.NotNil(t, r.store)
}

func Test_ramcache_removeExpiredItems(t *testing.T) {
	ctx := context.Background()
	r := &ramcache{}
	r.init(ctx, &cache.Config{}, Options{DefaultTTL: 24 * time.Hour, CleanupInterval: 5 * time.Minute})

	// Add an expired item
	expiredKey := "expired"
	r.store.Set(expiredKey, item{
		Value:  []byte("expired"),
		Expiry: time.Now().Add(-time.Hour), // 1 hour in the past
	})

	// Add a non-expired item
	nonExpiredKey := "nonExpired"
	r.store.Set(nonExpiredKey, item{
		Value:  []byte("nonExpired"),
		Expiry: time.Now().Add(time.Hour), // 1 hour in the future
	})

	// Call the method under test
	r.removeExpiredItems()

	// Check that the expired item was removed
	if _, exists := r.store.Get(expiredKey); exists {
		t.Errorf("Expected expired item to be removed, but it was not")
	}

	// Check that the non-expired item was not removed
	if _, exists := r.store.Get(nonExpiredKey); !exists {
		t.Errorf("Expected non-expired item to not be removed, but it was")
	}
}

func setupCache(t *testing.T) *ramcache {
	t.Helper()
	config := cache.Config{}
	r := New(context.Background(), &config, Options{})
	return r
}

type harness struct {
	cache *ramcache
}

func (h *harness) MakeCache(ctx context.Context) (cache.Cache, error) {
	return h.cache, nil
}

func (h *harness) Close() {}

func (h *harness) Options() drivertest.Options {
	return drivertest.Options{
		PatternMatchingDisabled: true, // Ramcache does not support pattern matching
		CloseIsNoop:             true, // Cache can still be used after closing
	}
}

func newHarness(ctx context.Context, t *testing.T) (drivertest.Harness, error) {
	cache := setupCache(t)
	return &harness{
		cache: cache,
	}, nil
}

func TestConformance(t *testing.T) {
	drivertest.RunConformanceTests(t, newHarness)
}
