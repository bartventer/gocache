package rediscluster

import (
	"net/url"
	"strings"
	"time"

	"github.com/bartventer/gocache/internal/urlparser"
	"github.com/mitchellh/mapstructure"
)

// paramKeyBlacklist is a list of keys that should not be set on the Redis Cluster options.
var paramKeyBlacklist = map[string]bool{
	"addrs":                      true,
	"newclient":                  true,
	"clusterslots":               true,
	"dialer":                     true,
	"onconnect":                  true,
	"credentialsprovider":        true,
	"credentialsprovidercontext": true,
}

// optionsFromURL parses a [url.URL] into [redis.ClusterOptions].
//
// The URL should have the following format:
//
//	rediscluster://<host1>:<port1>,<host2>:<port2>,...,<hostN>:<portN>[?query]
//
// All cluster options can be set as query parameters, except for the following:
//   - Addrs
//   - Any option that is a function
//   - Any options defined in cache.Options
//
// Example:
//
//	redis://localhost:6379,localhost:6380?maxretries=5&minretrybackoff=1000ms
//
// This will return a redis.ClusterOptions with the Addrs set to ["localhost:6379", "localhost:6380"],
// MaxRetries set to 5, and MinRetryBackoff set to 1000ms.
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

	// Set the Addrs from the URL
	opts.Addrs = strings.Split(u.Host, ",")

	return opts, nil
}
