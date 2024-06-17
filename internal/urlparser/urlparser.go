/*
Package urlparser provides a utility for parsing URL query parameters into
a given struct. It uses the mapstructure package to decode query parameters
into the struct fields. It also supports custom decode hooks for specific types.

Example usage:

	type Options struct {
	    MaxRetries      int
	    MinRetryBackoff time.Duration
	    TLSConfig       *tls.Config
	}

	// Define a TLS config
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		MinVersion:         tls.VersionTLS13,
		CipherSuites:       []uint16{tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256, tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384},
		PreferServerCipherSuites: true,
		ServerName: "localhost",
		ClientAuth: tls.RequireAndVerifyClientCert,
	}

	// Encode the tls.Config as a JSON string
	tlsConfigJSON, _ := json.Marshal(tlsConfig)
	tlsConfigStr := string(tlsConfigJSON)

	// Create a URL with query parameters
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
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"flag"
	"log"
	"net/url"
	"reflect"
	"strings"
	"sync"

	"github.com/mitchellh/mapstructure"
)

// debug is a flag that enables debug logging.
var debug = flag.Bool("gocache-debug", false, "enable debug logging")

// URLParser is a utility for parsing [url.URL] query parameters into a given struct.
// It uses the [mapstructure] package to decode query parameters into the struct fields.
// It also supports custom decode hooks for specific types.
type URLParser struct {
	decodeHooks []mapstructure.DecodeHookFunc
	metadata    mapstructure.Metadata
	log         Logger
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
		if *debug {
			logger := defaultLogger()
			logger.Println("Debug logging enabled")
			p.log = logger
		} else {
			p.log = nopLogger{}
		}
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
		DecodeHook:       mapstructure.ComposeDecodeHookFunc(p.decodeHooks...),
		Metadata:         &p.metadata,
		MatchName: func(mapKey, fieldName string) bool {
			return strings.EqualFold(mapKey, fieldName) && !paramKeyBlacklist[mapKey]
		},
	})

	if err != nil {
		return err
	}
	err = decoder.Decode(queryParams)
	if err != nil {
		return err
	}
	p.logMetadata(options)
	return nil
}

func (p *URLParser) logMetadata(dest interface{}) {
	// Get the actual type of dest
	destType := reflect.TypeOf(dest).Elem()

	// Successful decoded keys
	if len(p.metadata.Keys) > 0 {
		log.Printf("Successfully decoded url keys for %v: %v", destType, strings.Join(p.metadata.Keys, ", "))
	} else {
		log.Println("No keys were successfully decoded from the url.")
	}

	// Unused keys
	if len(p.metadata.Unused) > 0 {
		log.Printf("Unused options keys for %v: %v", destType, strings.Join(p.metadata.Unused, ", "))
	} else {
		log.Println("No unused keys were found.")
	}

	// Unset keys
	if len(p.metadata.Unset) > 0 {
		log.Printf("Unset options keys for %v: %v", destType, strings.Join(p.metadata.Unset, ", "))
	} else {
		log.Println("No unset options keys were found.")
	}
}

// StringToCertificateHookFunc creates a decode hook for converting a [pem] encoded
// [x509.Certificate] string into a pointer to an [x509.Certificate].
func StringToCertificateHookFunc() mapstructure.DecodeHookFuncType {
	return func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		if f.Kind() != reflect.String || t != reflect.TypeOf(&x509.Certificate{}) {
			return data, nil
		}

		certPEMBlock, _ := pem.Decode([]byte(data.(string)))
		if certPEMBlock == nil {
			return nil, errors.New("failed to parse certificate PEM")
		}
		cert, err := x509.ParseCertificate(certPEMBlock.Bytes)
		if err != nil {
			return nil, err
		}

		return cert, nil
	}
}

// StringToTLSConfigHookFunc creates a decode hook for converting a [json] encoded
// [tls.Config] string into a pointer to a [tls.Config].
func StringToTLSConfigHookFunc() mapstructure.DecodeHookFuncType {
	return func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		if f.Kind() != reflect.String || t != reflect.TypeOf(&tls.Config{}) { //nolint:gosec // TLS MinVersion gets set later
			return data, nil
		}

		// Here we're assuming that the TLS config is represented as a JSON string
		var config tls.Config
		err := json.Unmarshal([]byte(data.(string)), &config)
		if err != nil {
			return nil, err
		}
		return &config, nil
	}
}
