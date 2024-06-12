package redis

import (
	"net/url"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/redis/go-redis/v9"
)

// paramKeyBlacklist is a list of keys that should not be set on the Redis Cluster options.
var paramKeyBlacklist = map[string]bool{
	"addr":                       true,
	"newclient":                  true,
	"clusterslots":               true,
	"dialer":                     true,
	"onconnect":                  true,
	"credentialsprovider":        true,
	"credentialsprovidercontext": true,
	"tlsconfig":                  true,
}

// optionsFromURL parses a [url.URL] into [redis.Options].
//
// The URL should have the following format:
//
//	redis://host:port
//
// All redis client options can be set as query parameters, except for the following:
//   - Addr
//   - Any option that is a function
//   - Any options defined in cache.Options
//
// Example:
//
//	redis://localhost:6379?maxretries=5&minretrybackoff=1000
//
// This will return a redis.Options with the Addr set to "localhost:6379",
// MaxRetries set to 5, and MinRetryBackoff set to 1000.
func optionsFromURL(u *url.URL, paramOverrides map[string]string) (redis.Options, error) {
	opts := redis.Options{}

	// Parse the query parameters into a map
	queryParams := make(map[string]string)
	for key, values := range u.Query() {
		if len(values) > 0 {
			queryParams[key] = values[0]
		}
	}
	// Merge the extra parameters
	for key, value := range paramOverrides {
		queryParams[key] = value
	}
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		WeaklyTypedInput: true,
		Result:           &opts,
		MatchName: func(mapKey, fieldName string) bool {
			return strings.EqualFold(mapKey, fieldName) && !paramKeyBlacklist[mapKey]
		},
	})
	if err != nil {
		return redis.Options{}, err
	}
	err = decoder.Decode(queryParams)
	if err != nil {
		return redis.Options{}, err
	}

	// Set the Addr from the URL
	opts.Addr = u.Host

	return opts, nil
}
