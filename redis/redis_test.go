package redis

import (
	"context"
	"testing"
	"time"

	"github.com/bartventer/gocache/cache"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedisCache_Count(t *testing.T) {
	// Mock the Redis client
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	c := &redisCache{client: client}

	key := "testKey1"
	value := "testValue1"

	if err := c.Set(context.Background(), key, value); err != nil {
		t.Fatalf("Failed to set key: %v", err)
	}
	t.Cleanup(func() {
		client.Del(context.Background(), key)
	})

	count, err := c.Count(context.Background(), "*")
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestRedisCache_Exists(t *testing.T) {
	// Mock the Redis client
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	c := &redisCache{client: client}
	key := "testKey"
	value := "testValue"

	if err := c.Set(context.Background(), key, value); err != nil {
		t.Fatalf("Failed to set key: %v", err)
	}
	t.Cleanup(func() {
		client.Del(context.Background(), key)
	})

	exists, err := c.Exists(context.Background(), key)
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestRedisCache_Del(t *testing.T) {
	// Mock the Redis client
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	c := &redisCache{client: client}

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
	assert.EqualError(t, cache.ErrKeyNotFound, err.Error())
}

func TestRedisCache_DelKeys(t *testing.T) {
	// Mock the Redis client
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	c := &redisCache{client: client}

	keys := []string{"testKey1", "testKey2", "testKey3"}
	for _, key := range keys {
		if err := c.Set(context.Background(), key, "testValue"); err != nil {
			t.Fatalf("Failed to set key: %v", err)
		}
	}

	err := c.DelKeys(context.Background(), "testKey*")
	require.NoError(t, err)

	for _, key := range keys {
		exists, errExist := c.Exists(context.Background(), key)
		require.NoError(t, errExist)
		assert.False(t, exists)
	}

	// Non-existent key
	err = c.DelKeys(context.Background(), "nonExistentKey*")
	require.NoError(t, err)
}

func TestRedisCache_Clear(t *testing.T) {
	// Mock the Redis client
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	c := &redisCache{client: client}

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
	// Mock the Redis client
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	c := &redisCache{client: client}

	key := "testKey"
	value := "testValue"

	if err := c.Set(context.Background(), key, value); err != nil {
		t.Fatalf("Failed to set key: %v", err)
	}
	t.Cleanup(func() {
		client.Del(context.Background(), key)
	})

	got, err := c.Get(context.Background(), key)
	require.NoError(t, err)
	assert.Equal(t, value, got)

	// Non-existent key
	_, err = c.Get(context.Background(), "nonExistentKey")
	require.Error(t, err)
	require.EqualError(t, cache.ErrKeyNotFound, err.Error())
}

func TestRedisCache_Set(t *testing.T) {
	// Mock the Redis client
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	c := &redisCache{client: client}

	key := "testKey"
	value := "testValue"

	err := c.Set(context.Background(), key, value)
	require.NoError(t, err)
	t.Cleanup(func() {
		client.Del(context.Background(), key)
	})

	got, err := c.Get(context.Background(), key)
	require.NoError(t, err)
	assert.Equal(t, value, got)
}

func TestRedisCache_SetWithExpiry(t *testing.T) {
	// Mock the Redis client
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	c := &redisCache{client: client}

	key := "testKey"
	value := "testValue"
	expiry := 1 * time.Second

	err := c.SetWithExpiry(context.Background(), key, value, expiry)
	require.NoError(t, err)
	t.Cleanup(func() {
		client.Del(context.Background(), key)
	})

	got, err := c.Get(context.Background(), key)
	require.NoError(t, err)
	assert.Equal(t, value, got)

	// Wait for the key to expire
	time.Sleep(expiry)

	_, err = c.Get(context.Background(), key)
	require.Error(t, err)
	require.EqualError(t, cache.ErrKeyNotFound, err.Error())
}
