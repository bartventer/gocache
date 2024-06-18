package ramcache

import (
	"net/url"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func Test_optionsFromURL(t *testing.T) {
	type args struct {
		u              *url.URL
		paramOverrides map[string]string
	}
	tests := []struct {
		name    string
		args    args
		want    Options
		wantErr bool
	}{
		{
			name: "parses valid URL",
			args: args{
				u:              mustParseURL("ramcache://?defaultttl=5m"),
				paramOverrides: map[string]string{},
			},
			want: Options{
				DefaultTTL: 5 * time.Minute,
			},
			wantErr: false,
		},
		{
			name: "ignores blacklisted parameters",
			args: args{
				u:              mustParseURL("ramcache://?defaultttl=5m"),
				paramOverrides: map[string]string{"blacklistedParam": "value"},
			},
			want: Options{
				DefaultTTL: 5 * time.Minute,
			},
			wantErr: false,
		},
		{
			name: "returns error for invalid parameters",
			args: args{
				u:              mustParseURL("ramcache://?defaultttl=invalid"),
				paramOverrides: map[string]string{},
			},
			want:    Options{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := optionsFromURL(tt.args.u)
			if (err != nil) != tt.wantErr {
				t.Errorf("optionsFromURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got, cmpopts.IgnoreUnexported(Options{})); diff != "" {
				t.Errorf("optionsFromURL() mismatch (-want +got):\n%s", diff)
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
