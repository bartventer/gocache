package cache

import (
	"context"
	"errors"
	"net/url"
	"testing"

	"github.com/bartventer/gocache/pkg/driver"
	"github.com/bartventer/gocache/pkg/keymod"
)

type mockURLOpener[K driver.String] struct{}

func (m *mockURLOpener[K]) OpenCacheURL(ctx context.Context, u *url.URL) (*GenericCache[K], error) {
	if u.Scheme == "err" {
		return nil, errors.New("forced error")
	}
	return nil, nil
}

func TestCache(t *testing.T) {
	ctx := context.Background()
	fake := &mockURLOpener[string]{}
	RegisterCache("foo", fake)
	RegisterCache("err", fake)

	for _, tc := range []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "empty URL",
			wantErr: true,
		},
		{
			name:    "invalid URL",
			url:     ":foo",
			wantErr: true,
		},
		{
			name:    "invalid URL no scheme",
			url:     "foo",
			wantErr: true,
		},
		{
			name:    "unregistered scheme",
			url:     "bar://mycache",
			wantErr: true,
		},
		{
			name:    "func returns error",
			url:     "err://mycache",
			wantErr: true,
		},
		{
			name: "no query options",
			url:  "foo://mycache",
		},
		{
			name: "empty query options",
			url:  "foo://mycache?",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, gotErr := OpenGenericCache[string](ctx, tc.url)
			if (gotErr != nil) != tc.wantErr {
				t.Fatalf("got err %v, want error %v", gotErr, tc.wantErr)
			}
		})
	}
}

func TestRegisterCache(t *testing.T) {
	fake := &mockURLOpener[string]{}

	// Test registering a new scheme.
	RegisterCache("new", fake)

	// Register same scheme but with different type.
	fake2 := &mockURLOpener[keymod.Key]{}
	RegisterCache("new", fake2)

	// Test registering an existing scheme and type. Should panic.
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	RegisterCache("new", fake)
}

func TestOpenCache(t *testing.T) {
	ctx := context.Background()
	fake := &mockURLOpener[string]{}

	// Test opening a cache with a valid scheme.
	RegisterCache("baz", fake)
	_, err := OpenGenericCache[string](ctx, "baz://mycache")
	if err != nil {
		t.Fatalf("Failed to open cache: %v", err)
	}

	// Test opening a cache with a valid scheme and invalid type.
	_, err = OpenGenericCache[keymod.Key](ctx, "baz://mycache")
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}
}
