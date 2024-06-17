package urlparser

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/mitchellh/mapstructure"
)

type FakeOptions struct {
	Addr            string
	MaxRetries      int
	MinRetryBackoff time.Duration
	Cert            *x509.Certificate
	TLSConfig       *tls.Config
}

func TestURLParser_OptionsFromURL(t *testing.T) {
	type args struct {
		u                 *url.URL
		options           interface{}
		paramKeyBlacklist map[string]bool
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name: "parses valid URL",
			args: args{
				u:                 mustParseURL("fake://localhost:6379?maxretries=5&minretrybackoff=512ms"),
				options:           &FakeOptions{},
				paramKeyBlacklist: map[string]bool{"db": true},
			},
			want: &FakeOptions{
				MaxRetries:      5,
				MinRetryBackoff: 512 * time.Millisecond,
			},
			wantErr: false,
		},
		{
			name: "ignores blacklisted parameters",
			args: args{
				u:                 mustParseURL("fake://localhost:6379?addr=someotherhost:6379"),
				options:           &FakeOptions{},
				paramKeyBlacklist: map[string]bool{"addr": true},
			},
			want:    &FakeOptions{},
			wantErr: false,
		},
		{
			name: "returns error for invalid parameters",
			args: args{
				u:                 mustParseURL("fake://localhost:6379?maxretries=invalid"),
				options:           &FakeOptions{},
				paramKeyBlacklist: map[string]bool{},
			},
			want:    &FakeOptions{},
			wantErr: true,
		},
		{
			name: "parses certificate from PEM string",
			args: args{
				u:                 mustParseURL("fake://localhost:6379?cert=" + url.QueryEscape(testCertPEM)),
				options:           &FakeOptions{},
				paramKeyBlacklist: map[string]bool{},
			},
			want: &FakeOptions{
				Cert: mustParseCertificate(testCertPEM),
			},
			wantErr: false,
		},
		{
			name: "parses TLS config from JSON string",
			args: args{
				u:                 mustParseURL("fake://localhost:6379?tlsconfig=" + url.QueryEscape(testTLSConfigJSON)),
				options:           &FakeOptions{},
				paramKeyBlacklist: map[string]bool{},
			},
			want: &FakeOptions{
				TLSConfig: mustUnmarshalTLSConfig(testTLSConfigJSON),
			},
			wantErr: false,
		},
	}

	parser := NewURLParser(
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.StringToSliceHookFunc(","),
		mapstructure.StringToTimeHookFunc(time.RFC3339),
		mapstructure.StringToIPNetHookFunc(),
		mapstructure.StringToIPHookFunc(),
		mapstructure.RecursiveStructToMapHookFunc(),
		StringToTLSConfigHookFunc(),
		StringToCertificateHookFunc(),
	)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser.log.Printf("%s\n\turl: %s\n", t.Name(), tt.args.u.String())
			err := parser.OptionsFromURL(tt.args.u, tt.args.options, tt.args.paramKeyBlacklist)
			if (err != nil) != tt.wantErr {
				t.Errorf("OptionsFromURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, tt.args.options, cmpopts.IgnoreUnexported(FakeOptions{}, tls.Config{})); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestStringToCertificateHookFunc(t *testing.T) {
	hook := StringToCertificateHookFunc()

	// Test that the hook correctly parses a certificate from a PEM string
	cert, err := hook(reflect.TypeOf(""), reflect.TypeOf(&x509.Certificate{}), testCertPEM)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if diff := cmp.Diff(mustParseCertificate(testCertPEM), cert, cmpopts.IgnoreUnexported(x509.Certificate{})); diff != "" {
		t.Errorf("mismatch (-want +got):\n%s", diff)
	}

	// Test that the hook returns an error for an invalid PEM string
	_, err = hook(reflect.TypeOf(""), reflect.TypeOf(&x509.Certificate{}), "invalid")
	if err == nil {
		t.Error("expected an error, got nil")
	}
}

func TestStringToTLSConfigHookFunc(t *testing.T) {
	hook := StringToTLSConfigHookFunc()

	// Test that the hook correctly unmarshals a TLS config from a JSON string
	config, err := hook(reflect.TypeOf(""), reflect.TypeOf(&tls.Config{}), testTLSConfigJSON)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if diff := cmp.Diff(mustUnmarshalTLSConfig(testTLSConfigJSON), config, cmpopts.IgnoreUnexported(tls.Config{})); diff != "" {
		t.Errorf("mismatch (-want +got):\n%s", diff)
	}

	// Test that the hook returns an error for an invalid JSON string
	_, err = hook(reflect.TypeOf(""), reflect.TypeOf(&tls.Config{}), "invalid")
	if err == nil {
		t.Error("expected an error, got nil")
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
