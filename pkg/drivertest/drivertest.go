// Package drivertest provides conformance tests for cache implementations.
package drivertest

import (
	"context"
	"fmt"
	"testing"
	"time"

	cache "github.com/bartventer/gocache"
	"github.com/bartventer/gocache/pkg/driver"
	"github.com/bartventer/gocache/pkg/keymod"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Options describes the set of options that a cache supports.
type Options struct {
	// PatternMatchingDisabled is true if the cache does not support pattern matching.
	// If true, the cache does not support the following methods:
	//  - Count
	//  - DelKeys
	PatternMatchingDisabled bool

	// CloseIsNoop is true if the Close method is a no-op for the cache.
	// If true, the cache should still be usable after Close is called.
	CloseIsNoop bool
}

// Harness descibes the functionality test harnesses must provide to run
// conformance tests.
type Harness[K driver.String] interface {
	// MakeCache makes a [driver.Cache] for testing.
	MakeCache(context.Context) (driver.Cache[K], error)

	// Close closes resources used by the harness.
	Close()

	// Options returns the set of options that the cache supports.
	Options() Options
}

// HarnessMaker describes functions that construct a harness for running tests.
// It is called exactly once per test; Harness.Close() will be called when the test is complete.
type HarnessMaker[K driver.String, TB testing.TB] func(ctx context.Context, t TB) (Harness[K], error)

// RunConformanceTests runs conformance tests for driver implementations of the [cache.Cache[K]] interface.
func RunConformanceTests[K driver.String](t *testing.T, newHarness HarnessMaker[K, *testing.T]) {
	t.Helper()

	t.Run("Set", func(t *testing.T) { withCache(t, newHarness, testSet) })
	t.Run("SetWithTTL", func(t *testing.T) { withCache(t, newHarness, testSetWithTTL) })
	t.Run("Exists", func(t *testing.T) { withCache(t, newHarness, testExists) })
	t.Run("Count", func(t *testing.T) { withCache(t, newHarness, testCount) })
	t.Run("Get", func(t *testing.T) { withCache(t, newHarness, testGet) })
	t.Run("Del", func(t *testing.T) { withCache(t, newHarness, testDel) })
	t.Run("DelKeys", func(t *testing.T) { withCache(t, newHarness, testDelKeys) })
	t.Run("Clear", func(t *testing.T) { withCache(t, newHarness, testClear) })
	t.Run("Ping", func(t *testing.T) { withCache(t, newHarness, testPing) })
	t.Run("Close", func(t *testing.T) { withCache(t, newHarness, testClose) })
}

// withCache creates a new cache and runs the test function.
func withCache[K driver.String](t *testing.T, newHarness HarnessMaker[K, *testing.T], f func(*testing.T, *cache.GenericCache[K], Options)) {
	t.Helper()
	t.Parallel()

	ctx := context.Background()
	h, err := newHarness(ctx, t)
	if err != nil {
		t.Fatal(err)
	}
	defer h.Close()

	c, err := h.MakeCache(ctx)
	require.NoError(t, err)

	f(t, cache.NewCache(c), h.Options())
}

// makeKey creates a unique key for the test.
func makeKey[K driver.String](t *testing.T) K {
	t.Helper()
	key := fmt.Sprintf("%s-%d", t.Name(), time.Now().UnixNano())
	return K(key)
}

// testSet tests the Set method of the cache.
func testSet[K driver.String](t *testing.T, c *cache.GenericCache[K], opts Options) {
	key := makeKey[K](t)
	value := "testValue"

	err := c.Set(context.Background(), key, value)
	require.NoError(t, err)
	t.Cleanup(func() {
		c.Del(context.Background(), key)
	})

	got, err := c.Get(context.Background(), key)
	require.NoError(t, err)
	assert.Equal(t, value, string(got))
}

// testSetWithTTL tests the SetWithTTL method of the cache.
func testSetWithTTL[K driver.String](t *testing.T, c *cache.GenericCache[K], opts Options) {
	key := makeKey[K](t)
	value := "testValue"
	ttl := 1 * time.Second

	err := c.SetWithTTL(context.Background(), key, value, ttl)
	require.NoError(t, err)
	t.Cleanup(func() {
		c.Del(context.Background(), key)
	})

	got, err := c.Get(context.Background(), key)
	require.NoError(t, err)
	assert.Equal(t, value, string(got))

	// Wait for the key to expire
	time.Sleep(ttl)

	_, err = c.Get(context.Background(), key)
	require.Error(t, err)
	assert.Contains(t, err.Error(), cache.ErrKeyNotFound.Error())
}

// testExists tests the Exists method of the cache.
func testExists[K driver.String](t *testing.T, c *cache.GenericCache[K], opts Options) {
	key := makeKey[K](t)
	value := "testValue"

	err := c.Set(context.Background(), key, value)
	require.NoError(t, err)
	t.Cleanup(func() {
		c.Del(context.Background(), key)
	})

	exists, err := c.Exists(context.Background(), key)
	require.NoError(t, err)
	assert.True(t, exists)
}

// testCount tests the Count method of the cache.
func testCount[K driver.String](t *testing.T, c *cache.GenericCache[K], opts Options) {
	key := makeKey[K](t)
	value := "testValue"

	if opts.PatternMatchingDisabled {
		_, err := c.Count(context.Background(), "*")
		require.Error(t, err)
		assert.Contains(t, err.Error(), cache.ErrPatternMatchingNotSupported.Error())
		return
	}

	err := c.Set(context.Background(), key, value)
	require.NoError(t, err)
	t.Cleanup(func() {
		c.Del(context.Background(), key)
	})

	count, err := c.Count(context.Background(), "*")
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

// testGet tests the Get method of the cache.
func testGet[K driver.String](t *testing.T, c *cache.GenericCache[K], opts Options) {
	key := makeKey[K](t)
	value := "testValue"

	err := c.Set(context.Background(), key, value)
	require.NoError(t, err)
	t.Cleanup(func() {
		c.Del(context.Background(), key)
	})

	got, err := c.Get(context.Background(), key)
	require.NoError(t, err)
	assert.Equal(t, value, string(got))
}

// testDel tests the Del method of the cache.
func testDel[K driver.String](t *testing.T, c *cache.GenericCache[K], opts Options) {
	key := makeKey[K](t)
	value := "testValue"

	err := c.Set(context.Background(), key, value)
	require.NoError(t, err)

	err = c.Del(context.Background(), key)
	require.NoError(t, err)

	exists, err := c.Exists(context.Background(), key)
	require.NoError(t, err)
	assert.False(t, exists)

	// Non-existent key
	err = c.Del(context.Background(), "nonExistentKey")
	require.Error(t, err)
	assert.Contains(t, err.Error(), cache.ErrKeyNotFound.Error())
}

// testDelKeys tests the DelKeys method of the cache.
func testDelKeys[K driver.String](t *testing.T, c *cache.GenericCache[K], opts Options) {
	keys := []string{"testKey1", "testKey2", "testKey3", "testKey4", "testKey5"}
	hashTag := makeKey[K](t)

	if opts.PatternMatchingDisabled {
		err := c.DelKeys(context.Background(), K(keymod.Key("testKey*").TagPrefix(string(hashTag))))
		require.Error(t, err)
		assert.Contains(t, err.Error(), cache.ErrPatternMatchingNotSupported.Error())
		return
	}

	for _, key := range keys {
		if err := c.Set(context.Background(), K(keymod.Key(key).TagPrefix(string(hashTag))), "testValue"); err != nil {
			t.Fatalf("Failed to set key: %v", err)
		}
	}

	count, err := c.Count(context.Background(), K(keymod.Key("testKey*").TagPrefix(string(hashTag))))
	require.NoError(t, err)
	if !assert.Equal(t, int64(5), count) {
		t.FailNow()
	}

	err = c.DelKeys(context.Background(), K(keymod.Key("testKey*").TagPrefix(string(hashTag))))
	require.NoError(t, err)

	res, err := c.Count(context.Background(), K(keymod.Key("testKey*").TagPrefix(string(hashTag))))
	require.NoError(t, err)
	assert.Equal(t, int64(0), res)

	// Non-existent key
	err = c.DelKeys(context.Background(), K(keymod.Key("nonExistentKey*").TagPrefix(string(hashTag))))
	require.NoError(t, err)
}

// testClear tests the Clear method of the cache.
func testClear[K driver.String](t *testing.T, c *cache.GenericCache[K], opts Options) {
	key := makeKey[K](t)
	value := "testValue"

	err := c.Set(context.Background(), key, value)
	require.NoError(t, err)

	err = c.Clear(context.Background())
	require.NoError(t, err)

	exists, err := c.Exists(context.Background(), key)
	require.NoError(t, err)
	assert.False(t, exists)
}

// testPing tests the Ping method of the cache.
func testPing[K driver.String](t *testing.T, c *cache.GenericCache[K], opts Options) {
	err := c.Ping(context.Background())
	require.NoError(t, err)
}

// testClose tests the Close method of the cache.
func testClose[K driver.String](t *testing.T, c *cache.GenericCache[K], opts Options) {
	err := c.Close()
	require.NoError(t, err)

	err = c.Ping(context.Background())
	if opts.CloseIsNoop {
		// After closing, pinging should still succeed because Close is a no-op
		require.NoError(t, err)
	} else {
		require.Error(t, err)
	}
}
