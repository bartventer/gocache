package urlparser

// Hooks for converting query parameters into specific types.

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"reflect"

	"github.com/mitchellh/mapstructure"
)

// StringToCertificateHookFunc creates a decode hook for converting a [pem] encoded
// [x509.Certificate] string into a pointer to an [x509.Certificate].
func StringToCertificateHookFunc() mapstructure.DecodeHookFuncType {
	return func(f, t reflect.Type, data interface{}) (interface{}, error) {
		if f.Kind() != reflect.String || t != reflect.TypeOf(&x509.Certificate{}) {
			return data, nil
		}

		certPEMBlock, _ := pem.Decode([]byte(data.(string)))
		if certPEMBlock == nil {
			return nil, errors.New("gocache: failed to decode certificate PEM block")
		}
		cert, err := x509.ParseCertificate(certPEMBlock.Bytes)
		if err != nil {
			return nil, fmt.Errorf("gocache: failed to parse certificate: %w", err)
		}

		return cert, nil
	}
}

// StringToTLSConfigHookFunc creates a decode hook for converting a [json] encoded
// [tls.Config] string into a pointer to a [tls.Config].
func StringToTLSConfigHookFunc() mapstructure.DecodeHookFuncType {
	return func(f, t reflect.Type, data interface{}) (interface{}, error) {
		if f.Kind() != reflect.String || t != reflect.TypeOf(&tls.Config{}) { //nolint:gosec // TLS MinVersion gets set later
			return data, nil
		}

		// Here we're assuming that the TLS config is represented as a JSON string
		var config tls.Config
		err := json.Unmarshal([]byte(data.(string)), &config)
		if err != nil {
			return nil, fmt.Errorf("gocache: failed to parse TLS config: %w", err)
		}
		return &config, nil
	}
}
