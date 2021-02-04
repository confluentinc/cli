package shell

import (
	"fmt"
	"testing"

	goprompt "github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/shell/completer"
	cliMock "github.com/confluentinc/cli/mock"
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
			text, color := prefixState(pcmd.NewJWTValidator(log.New()), tt.args.config)
			require.Equal(t, tt.wantText, text)
			require.Equal(t, tt.wantColor, color)
		})
	}
}

func TestArgumentParsing(t *testing.T) {
	tests := []struct {
		name      string
		flagValue string
	}{
		{
			name:      "basic flag value",
			flagValue: "hi",
		},
		{
			name:      "flag value with space in between",
			flagValue: "hi hi",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commandCalled := false
			shellCmd := newTestShellCommandWithExpectedFlag(t, tt.flagValue, &commandCalled)
			shellCmd.SetArgs([]string{"ccloud", "api", "--description", tt.flagValue})
			err := shellCmd.Execute()
			require.NoError(t, err)
			require.True(t, commandCalled)
		})
	}
}

func newTestShellCommandWithExpectedFlag(t *testing.T, expectedFlag string, commandCalled *bool) *cobra.Command {
	cli := &cobra.Command{
		Use: "ccloud",
	}
	apiCommand := &cobra.Command{
		Use: "api",
		Run: func(cmd *cobra.Command, args []string) {
			description, err := cmd.Flags().GetString("description")
			require.NoError(t, err)
			require.Equal(t, expectedFlag, description)
			*commandCalled = true
			fmt.Println("CALLED")
		},
	}
	apiCommand.Flags().String("description", "", "Description of API key.")
	cli.AddCommand(apiCommand)
	config := v3.AuthenticatedCloudConfigMock()
	prerunner := &pcmd.PreRun{}
	shellCompleter := completer.NewShellCompleter(cli)
	return NewShellCmd(cli, prerunner, "ccloud", config, nil, shellCompleter, log.New(), cliMock.NewDummyAnalyticsMock(), nil)
}
