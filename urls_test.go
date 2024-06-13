package cache

import (
	"context"
	"errors"
	"net/url"
	"testing"
)

type mockCacheURLOpener struct{}

func (m *mockCacheURLOpener) OpenCacheURL(ctx context.Context, u *url.URL, options *Options) (Cache, error) {
	if u.Scheme == "err" {
		return nil, errors.New("forced error")
	}
	return nil, nil
}

func TestCache(t *testing.T) {
	ctx := context.Background()
	mux := new(urlMux)

	fake := &mockCacheURLOpener{}
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
			_, gotErr := mux.OpenCache(ctx, tc.url, &Options{})
			if (gotErr != nil) != tc.wantErr {
				t.Fatalf("got err %v, want error %v", gotErr, tc.wantErr)
			}
		})
	}
}
