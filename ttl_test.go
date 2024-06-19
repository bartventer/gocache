package cache

import (
	"testing"
	"time"
)

func TestValidateTTL(t *testing.T) {
	type args struct {
		ttl time.Duration
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Valid TTL",
			args: args{
				ttl: 10 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "Zero TTL",
			args: args{
				ttl: 0,
			},
			wantErr: false,
		},
		{
			name: "Negative TTL",
			args: args{
				ttl: -1 * time.Second,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateTTL(tt.args.ttl); (err != nil) != tt.wantErr {
				t.Errorf("ValidateTTL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
