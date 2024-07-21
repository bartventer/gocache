package cache

import (
	"context"
	"errors"
	"net/url"
	"sync"

	"github.com/bartventer/gocache/internal/gcerrors"
)

// URLOpener defines the interface for opening a cache using a URL.
type URLOpener interface {
	// OpenCacheURL opens a cache using a URL and options.
	OpenCacheURL(ctx context.Context, u *url.URL) (*Cache, error)
}

// urlMux is a multiplexer for cache schemes.
type urlMux struct {
	mu      sync.RWMutex         // mu is a mutex for synchronizing access to the schemes map.
	schemes map[string]URLOpener // schemes maps a cache scheme to a URLOpener.
}

// RegisterCache registers a URLOpener for a given scheme.
// If a URLOpener is already registered for the scheme, it panics.
func (m *urlMux) RegisterCache(scheme string, opener URLOpener) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.schemes == nil {
		m.schemes = make(map[string]URLOpener)
	}
	if _, exists := m.schemes[scheme]; exists {
		panic(gcerrors.New(errors.New("scheme already registered: " + scheme)))
	}
	m.schemes[scheme] = opener
}

// OpenCache opens a cache for the provided URL string.
// It returns an error if the URL cannot be parsed, or if no URLOpener is registered for the URL's scheme.
func (m *urlMux) OpenCache(ctx context.Context, urlstr string) (*Cache, error) {
	u, err := url.Parse(urlstr)
	if err != nil {
		return nil, err
	}
	m.mu.RLock()
	opener, ok := m.schemes[u.Scheme]
	m.mu.RUnlock()
	if !ok {
		return nil, gcerrors.New(errors.New("no registered opener for scheme: " + u.Scheme))
	}
	return opener.OpenCacheURL(ctx, u)
}

var defaultURLMux = new(urlMux)

// RegisterCache registers a [URLOpener] for a given scheme.
// If a [URLOpener] is already registered for the scheme, it panics.
func RegisterCache(scheme string, opener URLOpener) {
	defaultURLMux.RegisterCache(scheme, opener)
}

// OpenCache opens a [Cache] for the provided URL string.
// It returns an error if the URL cannot be parsed, or if no [URLOpener] is registered for the URL's scheme.
func OpenCache(ctx context.Context, urlstr string) (*Cache, error) {
	return defaultURLMux.OpenCache(ctx, urlstr)
}
