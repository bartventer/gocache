package ramcache

import (
	"net/url"

	"github.com/bartventer/gocache/internal/urlparser"
)

// paramKeyBlacklist is a list of keys that should not be set on the Options.
var paramKeyBlacklist = map[string]struct{}{
	// placeholder for future options
}

// optionsFromURL parses a [url.URL] into [Options].
//
// The URL should have the following format:
//
//	ramcache://?defaultttl=5m
//
// All ramcache client options can be set as query parameters, except for the following:
//   - DefaultTTL
//
// Example:
//
//	ramcache://?defaultttl=5m
//
// This will return a Options with the DefaultTTL set to 5 minutes.
func optionsFromURL(u *url.URL) (Options, error) {
	var opts Options

	// Parse the query parameters into a map
	parser := urlparser.New()
	if err := parser.OptionsFromURL(u, &opts, paramKeyBlacklist); err != nil {
		return Options{}, err
	}

	return opts, nil
}
