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

// urlParser is a utility for parsing [url.URL] query parameters into a given struct.
// It uses the [mapstructure] package to decode query parameters into the struct fields.
// It also supports custom decode hooks for specific types.
type urlParser struct {
	decodeHooks []mapstructure.DecodeHookFunc
	log         *log.Logger
	once        sync.Once
}

func newDecoderConfig(decodeHooks ...mapstructure.DecodeHookFunc) *mapstructure.DecoderConfig {
	config := &mapstructure.DecoderConfig{
		ZeroFields:       true,
		WeaklyTypedInput: true,
		Squash:           true,
		DecodeHook:       mapstructure.ComposeDecodeHookFunc(decodeHooks...),
	}
	return config
}

// New creates a new [urlParser] with the given [mapstructure.DecodeHookFunc] hooks.
// Decode hooks are functions that can convert query parameters into specific types.
// They are called in the order they are provided.
func New(decodeHooks ...mapstructure.DecodeHookFunc) *urlParser {
	u := &urlParser{}
	u.init(decodeHooks...)
	return u
}

func (p *urlParser) init(decodeHooks ...mapstructure.DecodeHookFunc) {
	p.once.Do(func() {
		p.log = logext.NewLogger(os.Stdout)
		if len(decodeHooks) > 0 {
			p.decodeHooks = decodeHooks
		} else {
			p.decodeHooks = []mapstructure.DecodeHookFunc{
				mapstructure.StringToTimeDurationHookFunc(),
				mapstructure.StringToSliceHookFunc(","),
				mapstructure.StringToTimeHookFunc(time.RFC3339),
				mapstructure.StringToIPNetHookFunc(),
				mapstructure.StringToIPHookFunc(),
				mapstructure.RecursiveStructToMapHookFunc(),
			}
		}
	})
}

func inBlacklist(bl map[string]struct{}, key string) bool {
	_, ok := bl[strings.ToLower(key)]
	return ok
}

// OptionsFromURL parses the query parameters from the given [url.URL] into the provided
// options struct. It uses the [mapstructure.DecodeHookFunc] hooks provided when creating the [urlParser]
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
func (p *urlParser) OptionsFromURL(u *url.URL, options interface{}, paramKeyBlacklist map[string]struct{}) error {
	// Parse the query parameters into a map
	queryParams := make(map[string]string)
	for key, values := range u.Query() {
		if inBlacklist(paramKeyBlacklist, key) {
			continue
		}
		if len(values) > 0 {
			queryParams[key] = values[0]
		}
	}

	// Set the decoder options
	config := newDecoderConfig(p.decodeHooks...)
	metadata := &mapstructure.Metadata{}
	config.Result = options
	config.Metadata = metadata
	config.MatchName = func(mapKey, fieldName string) bool {
		return strings.EqualFold(mapKey, fieldName) && !inBlacklist(paramKeyBlacklist, mapKey)
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return fmt.Errorf("urlparser: failed to create decoder: %w", err)
	}

	err = decoder.Decode(queryParams)
	if err != nil {
		return err
	}

	p.logMetadata(options, metadata)

	return nil
}

// logMetadata logs useful information about the decoded result.
func (p *urlParser) logMetadata(dest interface{}, metadata *mapstructure.Metadata) {
	// Get the actual type of dest
	destType := reflect.TypeOf(dest).Elem()

	// Successful decoded keys
	if len(metadata.Keys) > 0 {
		p.log.Printf("Successfully decoded url keys for %v: %v", destType, strings.Join(metadata.Keys, ", "))
	}

	// Unused keys
	if len(metadata.Unused) > 0 {
		p.log.Printf("Unused options keys for %v: %v", destType, strings.Join(metadata.Unused, ", "))
	}

	// Unset keys
	if len(metadata.Unset) > 0 {
		p.log.Printf("Unset options keys for %v: %v", destType, strings.Join(metadata.Unset, ", "))
	}
}
