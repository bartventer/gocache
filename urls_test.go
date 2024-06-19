package cache

import (
	"context"
	"errors"
	"net/url"
	"testing"
)

type mockURLOpener struct{}

func (m *mockURLOpener) OpenCacheURL(ctx context.Context, u *url.URL) (Cache, error) {
	if u.Scheme == "err" {
		return nil, errors.New("forced error")
	}
	return nil, nil
}

func TestCache(t *testing.T) {
	ctx := context.Background()
	mux := new(urlMux)

	fake := &mockURLOpener{}
	mux.RegisterCache("foo", fake)
	mux.RegisterCache("err", fake)

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
			_, gotErr := mux.OpenCache(ctx, tc.url)
			if (gotErr != nil) != tc.wantErr {
				t.Fatalf("got err %v, want error %v", gotErr, tc.wantErr)
			}
		})
	}
}

func TestRegisterCache(t *testing.T) {
	fake := &mockURLOpener{}

	// Test registering a new scheme.
	RegisterCache("new", fake)

	// Test registering an existing scheme, should panic.
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	RegisterCache("new", fake)
}

func TestOpenCache(t *testing.T) {
	ctx := context.Background()
	fake := &mockURLOpener{}
	RegisterCache("foo", fake)

	// Test opening a registered scheme.
	_, err := OpenCache(ctx, "foo://mycache")
	if err != nil {
		t.Errorf("OpenCache() error = %v, want nil", err)
	}

	// Test opening an unregistered scheme, should return an error.
	_, err = OpenCache(ctx, "bar://mycache")
	if err == nil {
		t.Errorf("OpenCache() error = nil, want non-nil")
	}
}
