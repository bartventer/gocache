package ramcache

import (
	"context"
	"net/url"
	"testing"

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
