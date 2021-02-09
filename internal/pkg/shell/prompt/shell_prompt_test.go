package prompt

import (
	"fmt"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	cliMock "github.com/confluentinc/cli/mock"
)

type Quotation int
const (
	NO_QUOTES Quotation = iota
	SINGLE_QUOTES
	DOUBLE_QUOTES
)

func TestPromptExecutorFunc(t *testing.T) {
	tests := []struct {
		name      string
		flagValue string
		expectedFlag string
		quoteType Quotation
	}{
		{
			name:      "no quotes basic flag value",
			flagValue: `describing`,
			expectedFlag: `describing`,
			quoteType: NO_QUOTES,
		},
		{
			name:      "single quotes basic flag value",
			flagValue: `describing`,
			expectedFlag: `describing`,
			quoteType: SINGLE_QUOTES,
		},
		{
			name:      "double quotes basic flag value",
			flagValue: `describing`,
			expectedFlag: `describing`,
			quoteType: DOUBLE_QUOTES,
		},
		{
			name:      "no quotes with escaped quotes",
			flagValue: `\"describing\'`,
			expectedFlag: `"describing'`,
			quoteType: NO_QUOTES,
		},
		{
			name:      "no quotes value with space in between splits flag value",
			flagValue: `describing stuff`,
			expectedFlag: `describing`,
			quoteType: NO_QUOTES,
		},
		{
			name:      "double quotes flag value with space in between",
			flagValue: `describing stuff`,
			expectedFlag: `describing stuff`,
			quoteType: DOUBLE_QUOTES,
		},
		{
			name:      "single quotes flag value with space in between",
			flagValue: `describing stuff`,
			expectedFlag: `describing stuff`,
			quoteType: SINGLE_QUOTES,
		},

		{
			name:      "single quotes nested in double quotes",
			flagValue: `describing 'complex' stuff`,
			expectedFlag: `describing 'complex' stuff`,
			quoteType: DOUBLE_QUOTES,
		},
		{
			name:      "escaped double quotes nested in double quotes",
			flagValue: `describing \"complex\" stuff`,
			expectedFlag: `describing "complex" stuff`,
			quoteType: DOUBLE_QUOTES,
		},
		{
			name:      "single quotes including escape character",
			flagValue: `describing \"complex\" stuff`,
			expectedFlag: `describing \"complex\" stuff`,
			quoteType: SINGLE_QUOTES,
		},
		{
			name:      "double quotes nested in single quotes",
			flagValue: `describing "complex" stuff`,
			expectedFlag: `describing "complex" stuff`,
			quoteType: SINGLE_QUOTES,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commandCalled := false
			cli := newTestCommandWithExpectedFlag(t, tt.expectedFlag, &commandCalled)
			config := v3.AuthenticatedCloudConfigMock()
			command := &instrumentedCommand{
				Command:   cli,
				analytics: cliMock.NewDummyAnalyticsMock(),
			}
			shellPrompt := &ShellPrompt{RootCmd: command}
			executorFunc := promptExecutorFunc(config, shellPrompt)
			var format string
			switch tt.quoteType {
			case NO_QUOTES:
				format = `api --description %s`
			case SINGLE_QUOTES:
				format = `api --description '%s'`
			case DOUBLE_QUOTES:
				format = `api --description "%s"`
			}
			executorFunc(fmt.Sprintf(format, tt.flagValue))
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
