package memcache

import (
	"context"
	"testing"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemcacheCache_Exists(t *testing.T) {
	client := memcache.New("localhost:11211")

	c := &memcacheCache{client: client}
	key := "testKey"
	value := "testValue"

	if err := c.Set(context.Background(), key, value); err != nil {
		t.Fatalf("Failed to set key: %v", err)
	}
	t.Cleanup(func() {
		client.Delete(key)
	})

	exists, err := c.Exists(context.Background(), key)
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestMemcacheCache_Del(t *testing.T) {
	client := memcache.New("localhost:11211")

	c := &memcacheCache{client: client}

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
	require.NoError(t, err)
}

func TestMemcacheCache_Clear(t *testing.T) {
	client := memcache.New("localhost:11211")

	c := &memcacheCache{client: client}

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

func TestMemcacheCache_Get(t *testing.T) {
	client := memcache.New("localhost:11211")

	c := &memcacheCache{client: client}

	key := "testKey"
	value := "testValue"

	if err := c.Set(context.Background(), key, value); err != nil {
		t.Fatalf("Failed to set key: %v", err)
	}
	t.Cleanup(func() {
		client.Delete(key)
	})

	got, err := c.Get(context.Background(), key)
	require.NoError(t, err)
	assert.Equal(t, []byte(value), got)

	// Non-existent key
	_, err = c.Get(context.Background(), "nonExistentKey")
	require.Error(t, err)
	require.EqualError(t, memcache.ErrCacheMiss, err.Error())
}

func TestMemcacheCache_Set(t *testing.T) {
	client := memcache.New("localhost:11211")

	c := &memcacheCache{client: client}

	key := "testKey"
	value := "testValue"

	err := c.Set(context.Background(), key, value)
	require.NoError(t, err)
	t.Cleanup(func() {
		client.Delete(key)
	})

	got, err := c.Get(context.Background(), key)
	require.NoError(t, err)
	assert.Equal(t, []byte(value), got)
}

func TestMemcacheCache_SetWithExpiry(t *testing.T) {
	client := memcache.New("localhost:11211")

	c := &memcacheCache{client: client}

	key := "testKey"
	value := "testValue"
	expiry := 1 * time.Second

	err := c.SetWithExpiry(context.Background(), key, value, expiry)
	require.NoError(t, err)
	t.Cleanup(func() {
		client.Delete(key)
	})

	got, err := c.Get(context.Background(), key)
	require.NoError(t, err)
	assert.Equal(t, []byte(value), got)

	// Wait for the key to expire
	time.Sleep(expiry + 1*time.Second)

	_, err = c.Get(context.Background(), key)
	require.Error(t, err)
	require.EqualError(t, memcache.ErrCacheMiss, err.Error())
}
