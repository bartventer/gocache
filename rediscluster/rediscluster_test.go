package rediscluster

import (
	"context"
	"net/url"
	"testing"
	"time"

	cache "github.com/bartventer/gocache"
	"github.com/bartventer/gocache/internal/testutil"
	"github.com/bartventer/gocache/keymod"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var clusterNodes = []string{
	"localhost:7000",
	"localhost:7001",
	"localhost:7002",
	"localhost:7003",
	"localhost:7004",
	"localhost:7005",
}

// newRedisCache creates a new Redis cache with a test client.
func newRedisClusterCache(t *testing.T) *redisClusterCache {
	t.Helper()
	client := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: clusterNodes,
	})
	err := client.Ping(context.Background()).Err()
	if err != nil {
		t.FailNow()
	}
	c := &redisClusterCache{client: client, config: &cache.Config{CountLimit: 100}}
	t.Cleanup(func() {
		client.Close()
	})
	return c
}

func TestRedisCache_OpenCacheURL(t *testing.T) {
	r := &redisClusterCache{}

	u, err := url.Parse("rediscluster://localhost:7000,localhost:7001,localhost:7002,localhost:7003,localhost:7004,localhost:7005?maxretries=5&minretrybackoff=1000ms")
	require.NoError(t, err)

	_, err = r.OpenCacheURL(context.Background(), u, &cache.Options{})
	require.NoError(t, err)
	assert.NotNil(t, r.client)
}

func TestRedisCache_New(t *testing.T) {
	ctx := context.Background()
	config := cache.Config{}
	options := redis.ClusterOptions{
		Addrs: clusterNodes,
	}

	r := New(ctx, &config, options)
	require.NotNil(t, r)
	assert.NotNil(t, r.client)
}

func TestRedisClusterCache_Count(t *testing.T) {
	c := newRedisClusterCache(t)

	key := testutil.UniqueKey(t)
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

func TestRedisClusterCache_Exists(t *testing.T) {
	c := newRedisClusterCache(t)
	key := testutil.UniqueKey(t)
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

func TestRedisClusterCache_Del(t *testing.T) {
	c := newRedisClusterCache(t)

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
}

func TestRedisClusterCache_DelKeys(t *testing.T) {
	c := newRedisClusterCache(t)

	keys := []string{"testKey1", "testKey2", "testKey3", "testKey4", "testKey5"}
	hashTag := testutil.UniqueKey(t)
	for _, key := range keys {
		if err := c.Set(context.Background(), key, "testValue", keymod.HashTagModifier(hashTag)); err != nil {
			t.Fatalf("Failed to set key: %v", err)
		}
	}

	count, err := c.Count(context.Background(), "testKey*", keymod.HashTagModifier(hashTag))
	require.NoError(t, err)
	if !assert.Equal(t, int64(5), count) {
		t.FailNow()
	}

	err = c.DelKeys(context.Background(), "testKey*", keymod.HashTagModifier(hashTag))
	require.NoError(t, err)

	res, err := c.Count(context.Background(), "testKey*", keymod.HashTagModifier(hashTag))
	require.NoError(t, err)
	assert.Equal(t, int64(0), res)

	// Non-existent key
	err = c.DelKeys(context.Background(), "nonExistentKey*", keymod.HashTagModifier(hashTag))
	require.NoError(t, err)
}

func TestRedisClusterCache_Clear(t *testing.T) {
	c := newRedisClusterCache(t)

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

func TestRedisClusterCache_Get(t *testing.T) {
	c := newRedisClusterCache(t)

	key := testutil.UniqueKey(t)
	value := "testValue"

	if err := c.Set(context.Background(), key, value); err != nil {
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
}

func TestRedisClusterCache_Set(t *testing.T) {
	c := newRedisClusterCache(t)

	key := testutil.UniqueKey(t)
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

func TestRedisClusterCache_SetWithExpiry(t *testing.T) {
	c := newRedisClusterCache(t)

	key := testutil.UniqueKey(t) + time.Now().String()
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

func TestRedisClusterCache_Ping(t *testing.T) {
	c := newRedisClusterCache(t)

	err := c.Ping(context.Background())
	require.NoError(t, err)
}

func TestRedisClusterCache_Close(t *testing.T) {
	c := newRedisClusterCache(t)

	err := c.Close()
	require.NoError(t, err)

	// After closing, pinging should result in an error
	err = c.Ping(context.Background())
	require.Error(t, err)
}
