package drivertest

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	cache "github.com/bartventer/gocache"
	"github.com/bartventer/gocache/keymod"
)

// Item is a cache Item.
type Item struct {
	Value  []byte    // Value is the item value.
	Expiry time.Time // Expiry is the item expiry time. Default is 24 hours.
}

// MockCache is an in-memory implementation of the cache.Cache interface.
type MockCache struct {
	once   sync.Once       // once ensures that the cache is initialized only once.
	mu     sync.RWMutex    // mu guards the store.
	store  map[string]Item // store is the in-memory store.
	config *cache.Config   // config is the cache configuration.
	opts   *Options        // options is the cache options.
}

func (r *MockCache) init(_ context.Context, config *cache.Config, options Options) {
	r.once.Do(func() {
		r.config = config
		r.store = make(map[string]Item)
		r.opts = &options
	})
}

// Ensure MockCache implements the cache.Cache interface.
var _ cache.Cache = &MockCache{}

// Count implements cache.Cache.
func (r *MockCache) Count(ctx context.Context, pattern string, modifiers ...keymod.KeyModifier) (int64, error) {
	return 0, cache.ErrPatternMatchingNotSupported
}

// Exists implements cache.Cache.
func (r *MockCache) Exists(ctx context.Context, key string, modifiers ...keymod.KeyModifier) (bool, error) {
	key = keymod.Modify(key, modifiers...)
	r.mu.RLock()
	item, exists := r.store[key]
	r.mu.RUnlock()
	if exists && time.Now().After(item.Expiry) {
		r.mu.Lock()
		delete(r.store, key)
		r.mu.Unlock()
		exists = false
	}
	return exists, nil
}

// Del implements cache.Cache.
func (r *MockCache) Del(ctx context.Context, key string, modifiers ...keymod.KeyModifier) error {
	key = keymod.Modify(key, modifiers...)
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.store[key]; !exists {
		return cache.ErrKeyNotFound
	}
	delete(r.store, key)
	return nil
}

// DelKeys implements cache.Cache.
func (r *MockCache) DelKeys(ctx context.Context, pattern string, modifiers ...keymod.KeyModifier) error {
	return cache.ErrPatternMatchingNotSupported
}

// Clear implements cache.Cache.
func (r *MockCache) Clear(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.store = make(map[string]Item)
	return nil
}

// Get implements cache.Cache.
func (r *MockCache) Get(ctx context.Context, key string, modifiers ...keymod.KeyModifier) ([]byte, error) {
	key = keymod.Modify(key, modifiers...)
	r.mu.Lock()
	defer r.mu.Unlock()
	it, exists := r.store[key]
	if !exists || time.Now().After(it.Expiry) {
		delete(r.store, key)
		return nil, cache.ErrKeyNotFound
	}
	return it.Value, nil
}

// Set implements cache.Cache.
func (r *MockCache) Set(ctx context.Context, key string, value interface{}, modifiers ...keymod.KeyModifier) error {
	key = keymod.Modify(key, modifiers...)
	r.mu.Lock()
	defer r.mu.Unlock()
	switch v := value.(type) {
	case string:
		r.store[key] = Item{Value: []byte(v), Expiry: time.Now().Add(1 * time.Hour)}
	case []byte:
		r.store[key] = Item{Value: v, Expiry: time.Now().Add(1 * time.Hour)}
	default:
		return fmt.Errorf("unsupported value type: %T", v)
	}
	return nil
}

// SetWithExpiry implements cache.Cache.
func (r *MockCache) SetWithExpiry(ctx context.Context, key string, value interface{}, expiry time.Duration, modifiers ...keymod.KeyModifier) error {
	key = keymod.Modify(key, modifiers...)
	r.mu.Lock()
	defer r.mu.Unlock()
	switch v := value.(type) {
	case string:
		r.store[key] = Item{Value: []byte(v), Expiry: time.Now().Add(expiry)}
	case []byte:
		r.store[key] = Item{Value: v, Expiry: time.Now().Add(expiry)}
	default:
		return fmt.Errorf("unsupported value type: %T", v)
	}
	return nil
}

// Close implements cache.Cache.
func (r *MockCache) Close() error {
	return nil
}

// Ping implements cache.Cache.
func (r *MockCache) Ping(ctx context.Context) error {
	return nil
}

// NewMockCache returns a new MockCache.
func NewMockCache() *MockCache {
	return &MockCache{
		store: make(map[string]Item),
		config: &cache.Config{
			CountLimit: 1000,
		},
		opts: &Options{
			PatternMatchingDisabled: true,
			CloseIsNoop:             true,
		},
	}
}

type MockHarness struct{}

func (h *MockHarness) MakeCache(ctx context.Context) (cache.Cache, error) {
	return NewMockCache(), nil
}

func (h *MockHarness) Close() {}

func (h *MockHarness) Options() Options {
	return Options{
		PatternMatchingDisabled: true,
		CloseIsNoop:             true,
	}
}

func TestRunConformanceTests(t *testing.T) {
	type args struct {
		t          *testing.T
		newHarness HarnessMaker
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "RunConformanceTests",
			args: args{
				t:          t,
				newHarness: func(ctx context.Context, t *testing.T) (Harness, error) { return &MockHarness{}, nil },
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RunConformanceTests(tt.args.t, tt.args.newHarness)
		})
	}
}
