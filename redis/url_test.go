package redis

import (
	"net/url"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/redis/go-redis/v9"
)

func Test_optionsFromURL(t *testing.T) {
	type args struct {
		u              *url.URL
		paramOverrides map[string]string
	}
	tests := []struct {
		name    string
		args    args
		want    redis.Options
		wantErr bool
	}{
		{
			name: "parses valid URL",
			args: args{
				u:              mustParseURL("redis://localhost:6379?maxretries=5&minretrybackoff=512ms"),
				paramOverrides: map[string]string{"db": "0"},
			},
			want: redis.Options{
				Addr:            "localhost:6379",
				MaxRetries:      5,
				MinRetryBackoff: 512 * time.Millisecond,
				DB:              0,
			},
			wantErr: false,
		},
		{
			name: "ignores blacklisted parameters",
			args: args{
				u:              mustParseURL("redis://localhost:6379?addr=someotherhost:6379"),
				paramOverrides: map[string]string{"newclient": "true"},
			},
			want: redis.Options{
				Addr: "localhost:6379",
			},
			wantErr: false,
		},
		{
			name: "returns error for invalid parameters",
			args: args{
				u:              mustParseURL("redis://localhost:6379?maxretries=invalid"),
				paramOverrides: map[string]string{},
			},
			want:    redis.Options{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := optionsFromURL(tt.args.u, tt.args.paramOverrides)
			if (err != nil) != tt.wantErr {
				t.Errorf("optionsFromURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got, cmpopts.IgnoreUnexported(redis.Options{})); diff != "" {
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
