package redis

import (
	"context"
	"net/url"
	"testing"
	"time"

	cache "github.com/bartventer/gocache"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const defaultAddr = "localhost:6380"

// newRedisCache creates a new Redis cache with a test client.
func newRedisCache(t *testing.T) *redisCache {
	t.Helper()
	client := redis.NewClient(&redis.Options{
		Addr: defaultAddr,
	})
	err := client.Ping(context.Background()).Err()
	if err != nil {
		t.FailNow()
	}
	c := &redisCache{client: client, config: &cache.Config{CountLimit: 100}}
	t.Cleanup(func() {
		client.Close()
	})
	return c
}

func TestRedisCache_OpenCacheURL(t *testing.T) {
	r := &redisCache{}

	u, err := url.Parse("redis://" + defaultAddr + "?maxretries=5&minretrybackoff=1000ms")
	require.NoError(t, err)

	_, err = r.OpenCacheURL(context.Background(), u, &cache.Options{})
	require.NoError(t, err)
	assert.NotNil(t, r.client)
}

func TestRedisCache_New(t *testing.T) {
	ctx := context.Background()
	config := cache.Config{}
	options := redis.Options{
		Addr: defaultAddr,
	}

	r := New(ctx, &config, options)
	require.NotNil(t, r)
	assert.NotNil(t, r.client)
}

func TestRedisCache_Count(t *testing.T) {
	c := newRedisCache(t)

	key := "testKey1"
	value := "testValue1"

	if err := c.Set(context.Background(), key, value); err != nil {
		t.Fatalf("Failed to set key: %v", err)
	}
	t.Cleanup(func() {
		c.client.Del(context.Background(), key)
	})

	count, err := c.Count(context.Background(), "*")
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestRedisCache_Exists(t *testing.T) {
	c := newRedisCache(t)
	key := "testKey"
	value := "testValue"

	if err := c.Set(context.Background(), key, value); err != nil {
		t.Fatalf("Failed to set key: %v", err)
	}
	t.Cleanup(func() {
		c.client.Del(context.Background(), key)
	})

	exists, err := c.Exists(context.Background(), key)
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestRedisCache_Del(t *testing.T) {
	c := newRedisCache(t)

	key := "testKey"
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
}

func TestRedisCache_DelKeys(t *testing.T) {
	c := newRedisCache(t)

	keys := []string{"testKey1", "testKey2", "testKey3"}
	for _, key := range keys {
		if err := c.Set(context.Background(), key, "testValue"); err != nil {
			t.Fatalf("Failed to set key: %v", err)
		}
	}

	err := c.DelKeys(context.Background(), "testKey*")
	require.NoError(t, err)

	for _, key := range keys {
		n, errExist := c.client.Exists(context.Background(), key).Result()
		require.NoError(t, errExist)
		assert.Equal(t, int64(0), n)
	}

	// Non-existent key
	err = c.DelKeys(context.Background(), "nonExistentKey*")
	require.NoError(t, err)
}

func TestRedisCache_Clear(t *testing.T) {
	c := newRedisCache(t)

	key := "testKey"
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

func TestRedisCache_Get(t *testing.T) {
	c := newRedisCache(t)

	key := "testKey"
	value := "testValue"

	if err := c.client.Set(context.Background(), key, value, 0).Err(); err != nil {
		t.Fatalf("Failed to set key: %v", err)
	}
	t.Cleanup(func() {
		c.client.Del(context.Background(), key)
	})

	got, err := c.Get(context.Background(), key)
	require.NoError(t, err)
	assert.Equal(t, value, string(got))

	// Non-existent key
	_, err = c.Get(context.Background(), "nonExistentKey")
	require.Error(t, err)
	assert.Contains(t, err.Error(), cache.ErrKeyNotFound.Error())
}

func TestRedisCache_Set(t *testing.T) {
	c := newRedisCache(t)

	key := "testKey"
	value := "testValue"

	err := c.Set(context.Background(), key, value)
	require.NoError(t, err)
	t.Cleanup(func() {
		c.client.Del(context.Background(), key)
	})

	got, err := c.Get(context.Background(), key)
	require.NoError(t, err)
	assert.Equal(t, value, string(got))
}

func TestRedisCache_SetWithExpiry(t *testing.T) {
	c := newRedisCache(t)

	key := "testKey"
	value := "testValue"
	expiry := 1 * time.Second

	err := c.SetWithExpiry(context.Background(), key, value, expiry)
	require.NoError(t, err)
	t.Cleanup(func() {
		c.client.Del(context.Background(), key)
	})

	got, err := c.Get(context.Background(), key)
	require.NoError(t, err)
	assert.Equal(t, value, string(got))

	// Wait for the key to expire
	time.Sleep(expiry)

	_, err = c.Get(context.Background(), key)
	require.Error(t, err)
	assert.Contains(t, err.Error(), cache.ErrKeyNotFound.Error())
}

func TestRedisCache_Ping(t *testing.T) {
	c := newRedisCache(t)

	err := c.Ping(context.Background())
	require.NoError(t, err)
}

func TestRedisCache_Close(t *testing.T) {
	c := newRedisCache(t)

	err := c.Close()
	require.NoError(t, err)

	// After closing, pinging should result in an error
	err = c.Ping(context.Background())
	require.Error(t, err)
}
