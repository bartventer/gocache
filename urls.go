package cache

import (
	"context"
	"errors"
	"net/url"
	"sync"
)

// CacheURLOpener is an interface for opening caches using URLs.
type CacheURLOpener interface {
	// OpenCacheURL opens a cache using a URL and options.
	OpenCacheURL(ctx context.Context, u *url.URL, options Options) (Cache, error)
}

var (
	mu      sync.RWMutex                      // mu is a mutex for synchronizing access to the schemes map.
	schemes = make(map[string]CacheURLOpener) // schemes is a map of registered cache openers by scheme.
)

// RegisterCache registers a CacheURLOpener for a given scheme.
// If a CacheURLOpener is already registered for the scheme, it panics.
func RegisterCache(scheme string, opener CacheURLOpener) {
	mu.Lock()
	defer mu.Unlock()
	if _, exists := schemes[scheme]; exists {
		panic("scheme already registered: " + scheme)
	}
	schemes[scheme] = opener
}

// OpenCache opens a cache using a URL string and options.
// It returns an error if the URL cannot be parsed, or if no CacheURLOpener is registered for the URL's scheme.
func OpenCache(ctx context.Context, urlstr string, options Options) (Cache, error) {
	u, err := url.Parse(urlstr)
	if err != nil {
		return nil, err
	}
	mu.RLock()
	opener, ok := schemes[u.Scheme]
	mu.RUnlock()
	if !ok {
		return nil, errors.New("no registered opener for scheme: " + u.Scheme)
	}
	if options.CountLimit <= 0 {
		options.CountLimit = DefaultCountLimit
	}
	return opener.OpenCacheURL(ctx, u, options)
}
