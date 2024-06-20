package urlparser

import (
	"crypto/tls"
	"crypto/x509"
	"net/url"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/mitchellh/mapstructure"
)

type Embedded struct {
	Name string
}

type FakeOptions struct {
	Embedded
	Addr            string
	MaxRetries      int
	MinRetryBackoff time.Duration
	Cert            *x509.Certificate
	TLSConfig       *tls.Config
}

func TestNewURLParser(t *testing.T) {
	parser := NewURLParser()
	if parser == nil {
		t.Errorf("NewURLParser() = nil, want non-nil")
	}
}

func TestNewDecoderConfig(t *testing.T) {
	config := newDecoderConfig()
	if config == nil {
		t.Errorf("newDecoderConfig() = nil, want non-nil")
	}
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
		{
			name: "returns error for non-pointer destination",
			args: args{
				u:                 mustParseURL("fake://localhost:6379?tlsconfig=" + url.QueryEscape(testTLSConfigJSON)),
				options:           FakeOptions{}, // non-pointer destination, should return mapstructure error
				paramKeyBlacklist: map[string]bool{},
			},
			want:    &FakeOptions{},
			wantErr: true,
		},
		{
			name: "parses embedded struct",
			args: args{
				u:                 mustParseURL("fake://localhost:6379?name=TestName"),
				options:           &FakeOptions{},
				paramKeyBlacklist: map[string]bool{},
			},
			want: &FakeOptions{
				Embedded: Embedded{
					Name: "TestName",
				},
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
			if tt.wantErr {
				return
			}
			if diff := cmp.Diff(tt.want, tt.args.options, cmpopts.IgnoreUnexported(FakeOptions{}, tls.Config{})); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
