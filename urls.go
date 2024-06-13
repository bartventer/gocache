package cache

import (
	"context"
	"errors"
	"net/url"
	"sync"

	"github.com/bartventer/gocache/internal/gcerrors"
)

// CacheURLOpener is an interface for opening caches using URLs.
type CacheURLOpener interface {
	// OpenCacheURL opens a cache using a URL and options.
	OpenCacheURL(ctx context.Context, u *url.URL, options *Options) (Cache, error)
}

// urlMux is a multiplexer for cache schemes.
type urlMux struct {
	mu      sync.RWMutex              // mu is a mutex for synchronizing access to the schemes map.
	schemes map[string]CacheURLOpener // schemes maps a cache scheme to a CacheURLOpener.
}

// RegisterCache registers a CacheURLOpener for a given scheme.
// If a CacheURLOpener is already registered for the scheme, it panics.
func (m *urlMux) RegisterCache(scheme string, opener CacheURLOpener) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.schemes == nil {
		m.schemes = make(map[string]CacheURLOpener)
	}
	if _, exists := m.schemes[scheme]; exists {
		panic(gcerrors.New(errors.New("scheme already registered: " + scheme)))
	}
	m.schemes[scheme] = opener
}

// OpenCache opens a cache using a URL string and options.
// It returns an error if the URL cannot be parsed, or if no CacheURLOpener is registered for the URL's scheme.
func (m *urlMux) OpenCache(ctx context.Context, urlstr string, options *Options) (Cache, error) {
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
	options.Revise()
	return opener.OpenCacheURL(ctx, u, options)
}

var defaultRegistry = new(urlMux)

// RegisterCache registers a CacheURLOpener for a given scheme.
// If a CacheURLOpener is already registered for the scheme, it panics.
func RegisterCache(scheme string, opener CacheURLOpener) {
	defaultRegistry.RegisterCache(scheme, opener)
}

// OpenCache opens a cache using a URL string and options.
// It returns an error if the URL cannot be parsed, or if no CacheURLOpener is registered for the URL's scheme.
func OpenCache(ctx context.Context, urlstr string, options *Options) (Cache, error) {
	return defaultRegistry.OpenCache(ctx, urlstr, options)
}
