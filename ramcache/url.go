package ramcache

import (
	"net/url"
	"strings"

	"github.com/mitchellh/mapstructure"
)

// paramKeyBlacklist is a list of keys that should not be set on the Options.
var paramKeyBlacklist = map[string]bool{
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
func optionsFromURL(u *url.URL, paramOverrides map[string]string) (Options, error) {
	opts := Options{}

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
		DecodeHook:       mapstructure.StringToTimeDurationHookFunc(),
		MatchName: func(mapKey, fieldName string) bool {
			return strings.EqualFold(mapKey, fieldName) && !paramKeyBlacklist[mapKey]
		},
	})
	if err != nil {
		return Options{}, err
	}
	err = decoder.Decode(queryParams)
	if err != nil {
		return Options{}, err
	}

	return opts, nil
}
