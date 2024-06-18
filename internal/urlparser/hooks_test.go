package urlparser

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net/url"
	"reflect"
	"testing"
)

func TestStringToCertificateHookFunc(t *testing.T) {
	hook := StringToCertificateHookFunc()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "Valid PEM string",
			input:   testCertPEM,
			wantErr: false,
		},
		{
			name:    "Invalid PEM string",
			input:   "invalid",
			wantErr: true,
		},
		{
			name:    "Valid PEM string without valid certificate",
			input:   "-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := hook(reflect.TypeOf(""), reflect.TypeOf(&x509.Certificate{}), tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("StringToCertificateHookFunc() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStringToTLSConfigHookFunc(t *testing.T) {
	hook := StringToTLSConfigHookFunc()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "Valid JSON string",
			input:   testTLSConfigJSON,
			wantErr: false,
		},
		{
			name:    "Invalid JSON string",
			input:   "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := hook(reflect.TypeOf(""), reflect.TypeOf(&tls.Config{}), tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("StringToTLSConfigHookFunc() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func mustParseURL(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return u
}

func mustParseCertificate(pemStr string) *x509.Certificate {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		panic("failed to decode PEM block containing certificate")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		panic(err)
	}
	return cert
}

func mustUnmarshalTLSConfig(jsonStr string) *tls.Config {
	var config tls.Config
	err := json.Unmarshal([]byte(jsonStr), &config)
	if err != nil {
		panic(err)
	}
	return &config
}

// testCertPEM is a PEM-encoded certificate.
//
// Generated with:
//
//	openssl req -x509 -newkey rsa:512 -keyout key.pem -out cert.pem -days 365 -nodes.
const testCertPEM = `
-----BEGIN CERTIFICATE-----
MIIBfzCCASmgAwIBAgIUCIdg9oq/sX8obdNZkxkrxuGOyr8wDQYJKoZIhvcNAQEL
BQAwFDESMBAGA1UEAwwJbG9jYWxob3N0MB4XDTI0MDYxNzA4MDgzM1oXDTI1MDYx
NzA4MDgzM1owFDESMBAGA1UEAwwJbG9jYWxob3N0MFwwDQYJKoZIhvcNAQEBBQAD
SwAwSAJBAOBO8thWynustgb/5cvGbRraAZ305q4UKOThemEIsI/WhsXK0fqJ/Emq
8AVAJaogrD3yGYROJKEJ9loqh4D3jZ0CAwEAAaNTMFEwHQYDVR0OBBYEFILgGDUB
Hx9t+vv3mzp+yUlv4YsdMB8GA1UdIwQYMBaAFILgGDUBHx9t+vv3mzp+yUlv4Ysd
MA8GA1UdEwEB/wQFMAMBAf8wDQYJKoZIhvcNAQELBQADQQBiJro5f93F5CuoDpI9
Lu+hGkgtwvizqONHBGmSo4mX4M0f8n65gJn3qBpyYQIJTVKtL0VXsjxrOnuj8DUx
jmtJ
-----END CERTIFICATE-----
`

// testTLSConfigJSON is a JSON-encoded TLS config.
const testTLSConfigJSON = `{
    "InsecureSkipVerify": true,
    "MinVersion": 771,
    "CipherSuites": [4865, 4866],
    "PreferServerCipherSuites": true,
    "ServerName": "localhost",
	"ClientAuth": 1
}`
