package shell

import (
	"testing"

	goprompt "github.com/c-bata/go-prompt"
	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/internal/pkg/config"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
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
				config: v3.AuthenticatedCloudConfigMock(),
			},
			wantText:  "ccloud > ",
			wantColor: candyAppleGreen,
		},
		{
			name: "prefix when logged out",
			args: args{
				config: &v3.Config{
					BaseConfig: &config.BaseConfig{
						Params: &config.Params{CLIName: "testingtesting"},
					},
				},
			},
			wantText:  "testingtesting > ",
			wantColor: watermelonRed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text, color := prefixState(tt.args.config)
			require.Equal(t, tt.wantText, text)
			require.Equal(t, tt.wantColor, color)
		})
	}
}
