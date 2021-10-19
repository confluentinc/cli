package shell

import (
	"testing"

	goprompt "github.com/c-bata/go-prompt"
	"github.com/stretchr/testify/require"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/log"
)

const (
	validAuthToken = "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJPbmxpbmUgSldUIEJ1aWxkZXIiLCJpYXQiO" +
		"jE1NjE2NjA4NTcsImV4cCI6MjUzMzg2MDM4NDU3LCJhdWQiOiJ3d3cuZXhhbXBsZS5jb20iLCJzdWIiOiJqcm9ja2V0QGV4YW1w" +
		"bGUuY29tIn0.G6IgrFm5i0mN7Lz9tkZQ2tZvuZ2U7HKnvxMuZAooPmE"
)

func Test_prefixState(t *testing.T) {
	type args struct {
		config *v1.Config
	}
	tests := []struct {
		name      string
		args      args
		wantText  string
		wantColor goprompt.Color
	}{
		{
			name: "prefix when logged in",
			args: args{
				config: func() *v1.Config {
					cfg := v1.AuthenticatedCloudConfigMock()
					cfg.Context().State.AuthToken = validAuthToken
					return cfg
				}(),
			},
			wantText:  "confluent > ",
			wantColor: candyAppleGreen,
		},
		{
			name: "prefix when logged out",
			args: args{
				config: v1.UnauthenticatedCloudConfigMock(),
			},
			wantText:  "confluent > ",
			wantColor: watermelonRed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text, color := prefixState(pcmd.NewJWTValidator(log.New()), tt.args.config)
			require.Equal(t, tt.wantText, text)
			require.Equal(t, tt.wantColor, color)
		})
	}
}
