package redis

import (
	"net/url"
	"time"

	"github.com/bartventer/gocache/internal/urlparser"
	"github.com/mitchellh/mapstructure"
)

// paramKeyBlacklist is a list of keys that should not be set on the Redis options.
var paramKeyBlacklist = map[string]bool{
	"addr":                       true,
	"newclient":                  true,
	"dialer":                     true,
	"onconnect":                  true,
	"credentialsprovider":        true,
	"credentialsprovidercontext": true,
}

// optionsFromURL parses a [url.URL] into [redis.Options].
//
// The URL should have the following format:
//
//	redis://host:port
//
// All redis client options can be set as query parameters, except for the following:
//   - [redis.Options.Addr]
//   - Any option that is a function
//   - Any options defined in cache.Options
//
// Example:
//
//	redis://localhost:6379?maxretries=5&minretrybackoff=512ms
//
// This will return a redis.Options with the Addr set to "localhost:6379",
// MaxRetries set to 5, and MinRetryBackoff set to 512ms.
func optionsFromURL(u *url.URL) (Options, error) {
	var opts Options

	// Parse the query parameters into a map
	parser := urlparser.NewURLParser(
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.StringToSliceHookFunc(","),
		mapstructure.StringToTimeHookFunc(time.RFC3339),
		mapstructure.StringToIPNetHookFunc(),
		mapstructure.StringToIPHookFunc(),
		mapstructure.RecursiveStructToMapHookFunc(),
		urlparser.StringToTLSConfigHookFunc(),
		urlparser.StringToCertificateHookFunc(),
	)
	if err := parser.OptionsFromURL(u, &opts, paramKeyBlacklist); err != nil {
		return Options{}, err
	}

	// Set the Addr from the URL
	opts.Addr = u.Host

	return opts, nil
}
