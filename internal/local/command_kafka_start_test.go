package local

import (
	"testing"

	"github.com/moby/moby/api/types/network"
	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/v4/pkg/config"
)

func TestGetNatPlaintextPorts(t *testing.T) {
	tests := []struct {
		name  string
		ports []string
		want  []network.Port
	}{
		{
			name:  "empty",
			ports: nil,
			want:  []network.Port{},
		},
		{
			name:  "single broker",
			ports: []string{"9092"},
			want:  []network.Port{network.MustParsePort("9092/tcp")},
		},
		{
			name:  "multiple brokers",
			ports: []string{"9092", "9093", "9094"},
			want: []network.Port{
				network.MustParsePort("9092/tcp"),
				network.MustParsePort("9093/tcp"),
				network.MustParsePort("9094/tcp"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getNatPlaintextPorts(&config.LocalPorts{PlaintextPorts: tt.ports})
			require.Equal(t, tt.want, got)
		})
	}
}
