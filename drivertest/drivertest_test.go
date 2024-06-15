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

type MockCache struct {
	data map[string]interface{}
	mu   sync.RWMutex
}

func NewMockCache() *MockCache {
	return &MockCache{
		data: make(map[string]interface{}),
	}
}

func (m *MockCache) Set(ctx context.Context, key string, value interface{}, modifiers ...keymod.KeyModifier) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
	return nil
}

func (m *MockCache) SetWithExpiry(ctx context.Context, key string, value interface{}, expiry time.Duration, modifiers ...keymod.KeyModifier) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
	go func() {
		<-time.After(expiry)
		m.mu.Lock()
		delete(m.data, key)
		m.mu.Unlock()
	}()
	return nil
}

func (m *MockCache) Exists(ctx context.Context, key string, modifiers ...keymod.KeyModifier) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.data[key]
	return ok, nil
}

func (m *MockCache) Count(ctx context.Context, pattern string, modifiers ...keymod.KeyModifier) (int64, error) {
	return 0, cache.ErrPatternMatchingNotSupported
}

func (m *MockCache) Get(ctx context.Context, key string, modifiers ...keymod.KeyModifier) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	value, ok := m.data[key]
	if !ok {
		return nil, cache.ErrKeyNotFound
	}
	switch v := value.(type) {
	case []byte:
		return v, nil
	case string:
		return []byte(v), nil
	default:
		return nil, fmt.Errorf("unsupported value type: %T", v)
	}
}

func (m *MockCache) Del(ctx context.Context, key string, modifiers ...keymod.KeyModifier) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.data[key]
	if !ok {
		return cache.ErrKeyNotFound
	}
	delete(m.data, key)
	return nil
}

func (m *MockCache) DelKeys(ctx context.Context, pattern string, modifiers ...keymod.KeyModifier) error {
	return cache.ErrPatternMatchingNotSupported
}

func (m *MockCache) Clear(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data = make(map[string]interface{})
	return nil
}

func (m *MockCache) Ping(ctx context.Context) error {
	return nil
}

func (m *MockCache) Close() error {
	return nil
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
