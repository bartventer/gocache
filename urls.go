package cache

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sync"

	"github.com/bartventer/gocache/internal/gcerrors"
	"github.com/bartventer/gocache/pkg/driver"
	"github.com/bartventer/gocache/pkg/keymod"
)

// URLOpener defines the interface for opening a cache using a URL.
type URLOpener[K driver.String] interface {
	// OpenCacheURL opens a cache using a URL and options.
	OpenCacheURL(ctx context.Context, u *url.URL) (*GenericCache[K], error)
}

// urlMux is a multiplexer for cache schemes.
type urlMux struct {
	mu      sync.RWMutex                      // mu is a mutex for synchronizing access to the schemes map.
	schemes map[string]map[string]interface{} // schemes maps a cache scheme to a map of type strings to URLOpeners.
}

var defaultURLMux = new(urlMux)

// RegisterCache registers a [URLOpener] for a given scheme and type.
// If a [URLOpener] is already registered for the scheme and type, it panics.
func RegisterCache[K driver.String](scheme string, opener URLOpener[K]) {
	defaultURLMux.mu.Lock()
	defer defaultURLMux.mu.Unlock()
	if defaultURLMux.schemes == nil {
		defaultURLMux.schemes = make(map[string]map[string]interface{})
	}
	if _, exists := defaultURLMux.schemes[scheme]; !exists {
		defaultURLMux.schemes[scheme] = make(map[string]interface{})
	}
	typeKey := getTypeKey[K]()
	if _, exists := defaultURLMux.schemes[scheme][typeKey]; exists {
		panic(gcerrors.New(errors.New("scheme and type already registered: " + scheme + " - " + typeKey)))
	}
	defaultURLMux.schemes[scheme][typeKey] = opener
}

// getTypeKey returns a string representation of the type K.
func getTypeKey[K driver.String]() string {
	var k K
	return fmt.Sprintf("%T", k)
}

// OpenGenericCache opens a [GenericCache] for the provided URL string and type.
// It returns an error if the URL cannot be parsed, or if no [URLOpener] is registered for the URL's scheme and type.
func OpenGenericCache[K driver.String](ctx context.Context, urlstr string) (*GenericCache[K], error) {
	u, err := url.Parse(urlstr)
	if err != nil {
		return nil, err
	}
	defaultURLMux.mu.RLock()
	schemeOpeners, ok := defaultURLMux.schemes[u.Scheme]
	defaultURLMux.mu.RUnlock()
	if !ok {
		return nil, gcerrors.New(errors.New("no registered opener for scheme: " + u.Scheme))
	}
	typeKey := getTypeKey[K]()
	opener, ok := schemeOpeners[typeKey]
	if !ok {
		return nil, gcerrors.New(errors.New("no registered opener for type: " + typeKey))
	}
	return opener.(URLOpener[K]).OpenCacheURL(ctx, u)
}

// OpenCache opens a [Cache] for the provided URL string.
// It returns an error if the URL cannot be parsed, or if no [URLOpener] is registered for the URL's scheme.
func OpenCache(ctx context.Context, urlstr string) (*Cache, error) {
	return OpenGenericCache[string](ctx, urlstr)
}

// OpenKeyCache opens a [KeyCache] for the provided URL string.
// It returns an error if the URL cannot be parsed, or if no [URLOpener] is registered for the URL's scheme.
func OpenKeyCache(ctx context.Context, urlstr string) (*KeyCache, error) {
	return OpenGenericCache[keymod.Key](ctx, urlstr)
}
