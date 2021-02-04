package prompt

import (
	"fmt"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	cliMock "github.com/confluentinc/cli/mock"
)

func TestPromptExecutorFunc(t *testing.T) {
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
			cli := newTestCommandWithExpectedFlag(t, tt.flagValue, &commandCalled)
			config := v3.AuthenticatedCloudConfigMock()
			command := &instrumentedCommand{
				Command:   cli,
				analytics: cliMock.NewDummyAnalyticsMock(),
			}
			shellPrompt := &ShellPrompt{RootCmd: command}
			executorFunc := promptExecutorFunc(config, shellPrompt)
			executorFunc(fmt.Sprintf(`api --description "%s"`, tt.flagValue))
			require.True(t, commandCalled)
		})
	}
}

func newTestCommandWithExpectedFlag(t *testing.T, expectedFlag string, commandCalled *bool) *cobra.Command {
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
		},
	}
	apiCommand.Flags().String("description", "", "Description of API key.")
	cli.AddCommand(apiCommand)
	return cli
}
