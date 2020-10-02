package shell

import (
	"testing"

	goprompt "github.com/c-bata/go-prompt"
	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/internal/pkg/cmd"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/log"
)

const (
	validAuthToken = "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJPbmxpbmUgSldUIEJ1aWxkZXIiLCJpYXQiO" +
		"jE1NjE2NjA4NTcsImV4cCI6MjUzMzg2MDM4NDU3LCJhdWQiOiJ3d3cuZXhhbXBsZS5jb20iLCJzdWIiOiJqcm9ja2V0QGV4YW1w" +
		"bGUuY29tIn0.G6IgrFm5i0mN7Lz9tkZQ2tZvuZ2U7HKnvxMuZAooPmE"
)

func Test_prefixState(t *testing.T) {
	type args struct {
		config *v3.Config
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
				config: func() *v3.Config {
					cfg := v3.AuthenticatedCloudConfigMock()
					cfg.Context().State.AuthToken = validAuthToken
					return cfg
				}(),
			},
			wantText:  "ccloud > ",
			wantColor: candyAppleGreen,
		},
		{
			name: "prefix when logged out",
			args: args{
				config: v3.UnauthenticatedCloudConfigMock(),
			},
			wantText:  "ccloud > ",
			wantColor: watermelonRed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text, color := prefixState(cmd.NewJWTValidator(log.New()), tt.args.config)
			require.Equal(t, tt.wantText, text)
			require.Equal(t, tt.wantColor, color)
		})
	}
}
