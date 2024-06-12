package cache

import (
	"context"
	"time"
)

type MockCache struct{}

func (m *MockCache) Set(ctx context.Context, key string, value interface{}) error {
	return nil
}

func (m *MockCache) SetWithExpiry(ctx context.Context, key string, value interface{}, expiry time.Duration) error {
	return nil
}

func (m *MockCache) Exists(ctx context.Context, key string) (bool, error) {
	return true, nil
}

func (m *MockCache) Count(ctx context.Context, pattern string) (int64, error) {
	return 1, nil
}

func (m *MockCache) Get(ctx context.Context, key string) ([]byte, error) {
	return []byte("value"), nil
}

func (m *MockCache) Del(ctx context.Context, key string) error {
	return nil
}

func (m *MockCache) DelKeys(ctx context.Context, pattern string) error {
	return nil
}

func (m *MockCache) Clear(ctx context.Context) error {
	return nil
}

func (m *MockCache) Init(ctx context.Context, options Options) {}

// func TestNew(t *testing.T) {
// 	SetCurrentCache(&MockCache{})

// 	ctx := context.Background()
// 	c, err := New(ctx, Options{
// 		Addr:     "localhost:6379",
// 		Password: "",
// 		Username: "",
// 	})

// 	require.NoError(t, err)
// 	assert.NotNil(t, c)
// }

// func TestNew_NoCache(t *testing.T) {
// 	SetCurrentCache(nil)

// 	ctx := context.Background()
// 	c, err := New(ctx, Options{
// 		Addr:     "localhost:6379",
// 		Password: "",
// 		Username: "",
// 	})

// 	require.Error(t, err)
// 	assert.Nil(t, c)
// 	assert.Equal(t, ErrNoCache, err)
// }
