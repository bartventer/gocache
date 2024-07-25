package ramcache

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"testing"
	"time"

	cache "github.com/bartventer/gocache"
	"github.com/bartventer/gocache/pkg/driver"
	"github.com/bartventer/gocache/pkg/drivertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRamcacheCache_OpenCacheURL(t *testing.T) {
	r := &ramcache[string]{}
	u, err := url.Parse("ramcache://?defaultttl=1h")
	require.NoError(t, err)

	_, err = r.OpenCacheURL(context.Background(), u)
	require.NoError(t, err)
	assert.NotNil(t, r.store)
}

func TestRamcacheCache_New(t *testing.T) {
	ctx := context.Background()

	r := New[string](ctx, &Options{})
	require.NotNil(t, r)
	assert.NotNil(t, r.store)
}

func Test_ramcache_removeExpiredItems(t *testing.T) {
	ctx := context.Background()
	r := &ramcache[string]{}
	r.init(ctx, &Options{CleanupInterval: 5 * time.Minute})

	tests := []struct {
		name     string
		key      string
		value    []byte
		expiry   time.Time
		expected bool
	}{
		{
			name:     "Expired item",
			key:      "expired",
			value:    []byte("expired"),
			expiry:   time.Now().Add(-time.Hour), // 1 hour in the past
			expected: false,                      // Expected to be removed
		},
		{
			name:     "Non-expired item",
			key:      "nonExpired",
			value:    []byte("nonExpired"),
			expiry:   time.Now().Add(time.Hour), // 1 hour in the future
			expected: true,                      // Expected to not be removed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Add item
			r.store.Set(tt.key, item{
				Value:  tt.value,
				Expiry: tt.expiry,
			})

			// Call the method under test
			r.removeExpiredItems()

			// Check if the item was removed or not
			_, exists := r.store.Get(tt.key)
			if exists != tt.expected {
				t.Errorf("Expected existence of item to be %v, but got %v", tt.expected, exists)
			}
		})
	}
}

func TestSetWithTTL_InvalidExpiry(t *testing.T) {
	ctx := context.Background()
	r := New[string](ctx, &Options{})

	err := r.SetWithTTL(ctx, "key", "value", -1*time.Second)
	if !errors.Is(err, cache.ErrInvalidTTL) {
		t.Errorf("Expected error to be cache.ErrInvalidTTL, got %v", err)
	}
}

func Test_ramcache_set(t *testing.T) {
	ctx := context.Background()
	cache := New[string](ctx, &Options{})

	tests := []struct {
		name    string
		key     string
		value   interface{}
		wantErr bool
	}{
		{
			name:    "set string",
			key:     "key1",
			value:   "value1",
			wantErr: false,
		},
		{
			name:    "set bytes",
			key:     "key2",
			value:   []byte("value2"),
			wantErr: false,
		},
		{
			name:    "set binary marshaler",
			key:     "key3",
			value:   &BinaryMarshaler{},
			wantErr: false,
		},
		{
			name:    "set text marshaler",
			key:     "key4",
			value:   &TextMarshaler{},
			wantErr: false,
		},
		{
			name:    "set unsupported type",
			key:     "key5",
			value:   123,
			wantErr: true,
		},
		{
			name:    "set binary marshaler error",
			key:     "key6",
			value:   &BinaryMarshalerError{},
			wantErr: true,
		},
		{
			name:    "set text marshaler error",
			key:     "key7",
			value:   &TextMarshalerError{},
			wantErr: true,
		},
		{
			name:    "set json marshaler",
			key:     "key8",
			value:   &JSONMarshaler{},
			wantErr: false,
		},
		{
			name:    "set stringer",
			key:     "key9",
			value:   &Stringer{},
			wantErr: false,
		},
		{
			name:    "set reader",
			key:     "key11",
			value:   strings.NewReader("reader"),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cache.Set(ctx, tt.key, tt.value)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

type BinaryMarshaler struct{}

func (bm *BinaryMarshaler) MarshalBinary() ([]byte, error) {
	return []byte("binary marshaler"), nil
}

type BinaryMarshalerError struct{}

func (bm *BinaryMarshalerError) MarshalBinary() ([]byte, error) {
	return nil, assert.AnError
}

type TextMarshaler struct{}

func (tm *TextMarshaler) MarshalText() ([]byte, error) {
	return []byte("text marshaler"), nil
}

type TextMarshalerError struct{}

func (tm *TextMarshalerError) MarshalText() ([]byte, error) {
	return nil, assert.AnError
}

type JSONMarshaler struct{}

func (jm *JSONMarshaler) MarshalJSON() ([]byte, error) {
	return []byte(`{"json": "marshaler"}`), nil
}

type Stringer struct{}

func (s *Stringer) String() string {
	return "stringer"
}

func setupCache[K driver.String](t *testing.T) *ramcache[K] {
	t.Helper()
	r := New[K](context.Background(), &Options{})
	return r
}

type harness[K driver.String] struct {
	cache *ramcache[K]
}

func (h *harness[K]) MakeCache(ctx context.Context) (driver.Cache[K], error) {
	return h.cache, nil
}

func (h *harness[K]) Close() {}

func (h *harness[K]) Options() drivertest.Options {
	return drivertest.Options{
		PatternMatchingDisabled: true, // Ramcache does not support pattern matching
		CloseIsNoop:             true, // Cache can still be used after closing
	}
}

func newHarness[K driver.String](ctx context.Context, t *testing.T) (drivertest.Harness[K], error) {
	cache := setupCache[K](t)
	return &harness[K]{
		cache: cache,
	}, nil
}

func TestConformance(t *testing.T) {
	drivertest.RunConformanceTests(t, newHarness[string])
}
