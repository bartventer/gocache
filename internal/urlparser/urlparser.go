/*
Package urlparser provides utilities for parsing URL query parameters into
a given struct. It uses the mapstructure package to decode query parameters
into struct fields and supports custom decode hooks for specific types.

Example:

	type Options struct {
	    MaxRetries      int
	    MinRetryBackoff time.Duration
	    TLSConfig       *tls.Config
	}

	const tlsConfigStr = `{"InsecureSkipVerify":true}`

	urlStr := "fake://localhost:6379?maxretries=5&minretrybackoff=512ms&tlsconfig=" + url.QueryEscape(tlsConfigStr)
	u, _ := url.Parse(urlStr)
	options := &Options{}
	parser := NewURLParser(mapstructure.StringToTimeDurationHookFunc(), StringToTLSConfigHookFunc())
	err := parser.OptionsFromURL(u, options, map[string]bool{"db": true})

After running this code, the options struct will have MaxRetries set to 5,
MinRetryBackoff set to 512ms, and TLSConfig set to the corresponding tls.Config object.

Note: This package does not handle URL parsing itself. It expects a *url.URL
as input. It also does not set any fields in the struct that are not present
in the URL query parameters.
*/
package urlparser

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/bartventer/gocache/internal/logext"
	"github.com/mitchellh/mapstructure"
)

// URLParser is a utility for parsing [url.URL] query parameters into a given struct.
// It uses the [mapstructure] package to decode query parameters into the struct fields.
// It also supports custom decode hooks for specific types.
type URLParser struct {
	decodeHooks []mapstructure.DecodeHookFunc
	metadata    mapstructure.Metadata
	log         logext.Logger
	once        sync.Once
}

// NewURLParser creates a new [URLParser] with the given [mapstructure.DecodeHookFunc] hooks.
// Decode hooks are functions that can convert query parameters into specific types.
// They are called in the order they are provided.
func NewURLParser(decodeHooks ...mapstructure.DecodeHookFunc) *URLParser {
	parser := &URLParser{}
	parser.init(decodeHooks...)
	return parser
}

func (p *URLParser) init(decodeHooks ...mapstructure.DecodeHookFunc) {
	p.once.Do(func() {
		if len(decodeHooks) == 0 {
			decodeHooks = []mapstructure.DecodeHookFunc{
				mapstructure.StringToTimeDurationHookFunc(),
				mapstructure.StringToSliceHookFunc(","),
				mapstructure.StringToTimeHookFunc(time.RFC3339),
				mapstructure.StringToIPNetHookFunc(),
				mapstructure.StringToIPHookFunc(),
				mapstructure.RecursiveStructToMapHookFunc(),
			}
		}
		p.log = logext.NewLogger(os.Stdout)
		p.decodeHooks = decodeHooks
	})
}

// OptionsFromURL parses the query parameters from the given [url.URL] into the provided
// options struct. It uses the [mapstructure.DecodeHookFunc] hooks provided when creating the [URLParser]
// to convert query parameters into the correct types for the struct fields.
// It ignores any query parameters whose keys are in the paramKeyBlacklist.
// It returns an error if it fails to parse the [url.URL] or convert the query parameters.
//
// Example:
//
//	type Options struct {
//	    MaxRetries      int
//	    MinRetryBackoff time.Duration
//		DB              int
//	}
//
//	u, _ := url.Parse("fake://localhost:6379?maxretries=5&minretrybackoff=512ms&db=4")
//	options := &Options{}
//	bl := map[string]bool{"db": true}
//	err := parser.OptionsFromURL(u, options, bl)
//
// After running this code, the options struct will be:
//
//	Options{
//		MaxRetries:      5,
//		MinRetryBackoff: 512 * time.Millisecond,
//		DB:              0, // db is blacklisted and not set
//	}
func (p *URLParser) OptionsFromURL(u *url.URL, options interface{}, paramKeyBlacklist map[string]bool) error {
	// Parse the query parameters into a map
	queryParams := make(map[string]string)
	for key, values := range u.Query() {
		if len(values) > 0 {
			queryParams[key] = values[0]
		}
	}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		ZeroFields:       true, // ensure that the options struct is zeroed out before decoding
		WeaklyTypedInput: true,
		Result:           options,
		Squash:           true,
		DecodeHook:       mapstructure.ComposeDecodeHookFunc(p.decodeHooks...),
		Metadata:         &p.metadata,
		MatchName: func(mapKey, fieldName string) bool {
			return strings.EqualFold(mapKey, fieldName) && !paramKeyBlacklist[mapKey]
		},
	})

	if err != nil {
		return fmt.Errorf("urlparser: failed to create decoder: %w", err)
	}
	err = decoder.Decode(queryParams)
	if err != nil {
		return err
	}
	p.logMetadata(options)
	return nil
}

// logMetadata logs useful information about the decoded result.
func (p *URLParser) logMetadata(dest interface{}) {
	// Get the actual type of dest
	destType := reflect.TypeOf(dest).Elem()

	// Successful decoded keys
	if len(p.metadata.Keys) > 0 {
		log.Printf("Successfully decoded url keys for %v: %v", destType, strings.Join(p.metadata.Keys, ", "))
	}

	// Unused keys
	if len(p.metadata.Unused) > 0 {
		log.Printf("Unused options keys for %v: %v", destType, strings.Join(p.metadata.Unused, ", "))
	}

	// Unset keys
	if len(p.metadata.Unset) > 0 {
		log.Printf("Unset options keys for %v: %v", destType, strings.Join(p.metadata.Unset, ", "))
	}
}

// DefaultHooks returns the default decode hooks used by the [URLParser].
func DefaultHooks() mapstructure.DecodeHookFunc {
	return mapstructure.ComposeDecodeHookFunc(
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.StringToSliceHookFunc(","),
		mapstructure.StringToTimeHookFunc(time.RFC3339),
		mapstructure.StringToIPNetHookFunc(),
		mapstructure.StringToIPHookFunc(),
		mapstructure.RecursiveStructToMapHookFunc(),
	)
}