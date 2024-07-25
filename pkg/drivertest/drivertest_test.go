package drivertest

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	cache "github.com/bartventer/gocache"
	"github.com/bartventer/gocache/pkg/driver"
	"github.com/bartventer/gocache/pkg/keymod"
)

// Item is a cache Item.
type Item struct {
	Value  []byte    // Value is the item value.
	Expiry time.Time // Expiry is the item expiry time. Default is 24 hours.
}

// MockCache is an in-memory implementation of the cache.Cache interface.
type MockCache[K driver.String] struct {
	once  sync.Once    // once ensures that the cache is initialized only once.
	mu    sync.RWMutex // mu guards the store.
	store map[K]Item   // store is the in-memory store.
}

func (r *MockCache[K]) init(_ context.Context) {
	r.once.Do(func() {
		r.store = make(map[K]Item)
	})
}

// Ensure MockCache implements the cache.Cache interface.
var _ driver.Cache[string] = new(MockCache[string])
var _ driver.Cache[keymod.Key] = new(MockCache[keymod.Key])

// Count implements cache.Cache.
func (r *MockCache[K]) Count(ctx context.Context, pattern K) (int64, error) {
	return 0, cache.ErrPatternMatchingNotSupported
}

// Exists implements cache.Cache.
func (r *MockCache[K]) Exists(ctx context.Context, key K) (bool, error) {
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
func (r *MockCache[K]) Del(ctx context.Context, key K) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.store[key]; !exists {
		return cache.ErrKeyNotFound
	}
	delete(r.store, key)
	return nil
}

// DelKeys implements cache.Cache.
func (r *MockCache[K]) DelKeys(ctx context.Context, pattern K) error {
	return cache.ErrPatternMatchingNotSupported
}

// Clear implements cache.Cache.
func (r *MockCache[K]) Clear(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.store = make(map[K]Item)
	return nil
}

// Get implements cache.Cache.
func (r *MockCache[K]) Get(ctx context.Context, key K) ([]byte, error) {
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
func (r *MockCache[K]) Set(ctx context.Context, key K, value interface{}) error {
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

// SetWithTTL implements cache.Cache.
func (r *MockCache[K]) SetWithTTL(ctx context.Context, key K, value interface{}, ttl time.Duration) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	switch v := value.(type) {
	case string:
		r.store[key] = Item{Value: []byte(v), Expiry: time.Now().Add(ttl)}
	case []byte:
		r.store[key] = Item{Value: v, Expiry: time.Now().Add(ttl)}
	default:
		return fmt.Errorf("unsupported value type: %T", v)
	}
	return nil
}

// Close implements cache.Cache.
func (r *MockCache[K]) Close() error {
	return nil
}

// Ping implements cache.Cache.
func (r *MockCache[K]) Ping(ctx context.Context) error {
	return nil
}

// NewMockCache returns a new MockCache.
func NewMockCache[K driver.String]() *MockCache[K] {
	return &MockCache[K]{store: make(map[K]Item)}
}

type MockHarness[K driver.String] struct{}

func (h *MockHarness[K]) MakeCache(ctx context.Context) (driver.Cache[K], error) {
	return NewMockCache[K](), nil
}

func (h *MockHarness[K]) Close() {}

func (h *MockHarness[K]) Options() Options {
	return Options{
		PatternMatchingDisabled: true,
		CloseIsNoop:             true,
	}
}

func TestRunConformanceTests(t *testing.T) {
	RunConformanceTests(t, func(ctx context.Context, t *testing.T) (Harness[string], error) {
		return &MockHarness[string]{}, nil
	})
}
