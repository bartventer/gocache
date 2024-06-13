package memcache

import (
	"context"
	"net/url"
	"strings"
	"testing"
	"time"

	cache "github.com/bartventer/gocache"
	"github.com/bartventer/gocache/internal/testutil"
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const defaultAddr = "localhost:11211"

// malformedKey is a key that is too long which will trigger the [memcache.ErrMalformedKey] error.
var malformedKey = strings.Repeat("a", 251)

func TestMemcacheCache_OpenCacheURL(t *testing.T) {
	m := &memcacheCache{}

	u, err := url.Parse("memcache://" + defaultAddr)
	require.NoError(t, err)

	_, err = m.OpenCacheURL(context.Background(), u, &cache.Options{})
	require.NoError(t, err)
	assert.NotNil(t, m.client)
}

// newMemcacheCache creates a new Memcache cache with a test client.
func newMemcacheCache(t *testing.T) *memcacheCache {
	t.Helper()
	client := memcache.New(defaultAddr)
	err := client.Ping()
	if err != nil {
		t.FailNow()
	}
	c := &memcacheCache{client: client, config: &cache.Config{CountLimit: 100}}
	t.Cleanup(func() {
		client.Close()
	})
	return c
}

func TestMemcacheCache_New(t *testing.T) {
	ctx := context.Background()
	config := cache.Config{}

	m := New(ctx, &config, defaultAddr)
	require.NotNil(t, m)
	assert.NotNil(t, m.client)
}

func TestMemcacheCache_Exists(t *testing.T) {
	c := newMemcacheCache(t)
	key := testutil.UniqueKey(t)
	value := "testValue"

	if err := c.Set(context.Background(), key, value); err != nil {
		t.Fatalf("Failed to set key: %v", err)
	}
	t.Cleanup(func() {
		c.client.Delete(key)
	})

	exists, err := c.Exists(context.Background(), key)
	require.NoError(t, err)
	assert.True(t, exists)

	// Non-existent key
	exists, err = c.Exists(context.Background(), "nonExistentKey")
	require.NoError(t, err)
	assert.False(t, exists)

	// Malformed key
	_, err = c.Exists(context.Background(), malformedKey)
	require.Error(t, err)
	assert.Contains(t, err.Error(), memcache.ErrMalformedKey.Error())
}

func TestMemcacheCache_Del(t *testing.T) {
	c := newMemcacheCache(t)
	key := testutil.UniqueKey(t)
	value := "testValue"

	if err := c.Set(context.Background(), key, value); err != nil {
		t.Fatalf("Failed to set key: %v", err)
	}

	err := c.Del(context.Background(), key)
	require.NoError(t, err)

	exists, err := c.Exists(context.Background(), key)
	require.NoError(t, err)
	assert.False(t, exists)

	// Non-existent key
	err = c.Del(context.Background(), "nonExistentKey")
	require.Error(t, err)
	assert.Contains(t, err.Error(), cache.ErrKeyNotFound.Error())

	// Malformed key
	err = c.Del(context.Background(), malformedKey)
	require.Error(t, err)
	assert.Contains(t, err.Error(), memcache.ErrMalformedKey.Error())
}

func TestMemcacheCache_Clear(t *testing.T) {
	c := newMemcacheCache(t)
	key := testutil.UniqueKey(t)
	value := "testValue"

	if err := c.Set(context.Background(), key, value); err != nil {
		t.Fatalf("Failed to set key: %v", err)
	}

	err := c.Clear(context.Background())
	require.NoError(t, err)

	exists, err := c.Exists(context.Background(), key)
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestMemcacheCache_Get(t *testing.T) {
	c := newMemcacheCache(t)
	key := testutil.UniqueKey(t)
	value := "testValue"

	if err := c.Set(context.Background(), key, value); err != nil {
		t.Fatalf("Failed to set key: %v", err)
	}
	t.Cleanup(func() {
		c.client.Delete(key)
	})

	got, err := c.Get(context.Background(), key)
	require.NoError(t, err)
	assert.Equal(t, []byte(value), got)

	// Non-existent key
	_, err = c.Get(context.Background(), "nonExistentKey")
	require.Error(t, err)
	assert.Contains(t, err.Error(), cache.ErrKeyNotFound.Error())

	// Malformed key
	_, err = c.Get(context.Background(), malformedKey)
	require.Error(t, err)
	assert.Contains(t, err.Error(), memcache.ErrMalformedKey.Error())
}

func TestMemcacheCache_Set(t *testing.T) {
	c := newMemcacheCache(t)
	key := testutil.UniqueKey(t)
	value := "testValue"

	err := c.Set(context.Background(), key, value)
	require.NoError(t, err)
	t.Cleanup(func() {
		c.client.Delete(key)
	})

	got, err := c.Get(context.Background(), key)
	require.NoError(t, err)
	assert.Equal(t, []byte(value), got)

	// Malformed key
	err = c.Set(context.Background(), malformedKey, value)
	require.Error(t, err)
	assert.Contains(t, err.Error(), memcache.ErrMalformedKey.Error())
}

func TestMemcacheCache_SetWithExpiry(t *testing.T) {
	c := newMemcacheCache(t)
	key := testutil.UniqueKey(t)
	value := "testValue"
	expiry := 1 * time.Second

	err := c.SetWithExpiry(context.Background(), key, value, expiry)
	require.NoError(t, err)
	t.Cleanup(func() {
		c.client.Delete(key)
	})

	got, err := c.Get(context.Background(), key)
	require.NoError(t, err)
	assert.Equal(t, []byte(value), got)

	// Wait for the key to expire
	time.Sleep(expiry + 1*time.Second)

	_, err = c.Get(context.Background(), key)
	require.Error(t, err)
	assert.Contains(t, err.Error(), cache.ErrKeyNotFound.Error())

	// Malformed key
	err = c.SetWithExpiry(context.Background(), malformedKey, value, expiry)
	require.Error(t, err)
	assert.Contains(t, err.Error(), memcache.ErrMalformedKey.Error())
}

// Pattern matching operations not supported by Memcache

func TestMemcacheCache_Count(t *testing.T) {
	c := newMemcacheCache(t)
	_, err := c.Count(context.Background(), "*")
	require.Error(t, err)
	assert.Contains(t, err.Error(), cache.ErrPatternMatchingNotSupported.Error())
}

func TestMemcacheCache_DelKeys(t *testing.T) {
	c := newMemcacheCache(t)
	err := c.DelKeys(context.Background(), "*")
	require.Error(t, err)
	assert.Contains(t, err.Error(), cache.ErrPatternMatchingNotSupported.Error())
}

func TestMemcacheCache_Ping(t *testing.T) {
	c := newMemcacheCache(t)

	err := c.Ping(context.Background())
	require.NoError(t, err)
}

func TestMemcacheCache_Close(t *testing.T) {
	c := newMemcacheCache(t)

	err := c.Close()
	require.NoError(t, err)

	// After closing, pinging should still succeed because Close is a no-op
	err = c.Ping(context.Background())
	require.NoError(t, err)
}
