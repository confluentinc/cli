package prompt

import (
	"fmt"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

type Quotation int

const (
	NoQuotes Quotation = iota
	SingleQuotes
	DoubleQuotes
)

func TestPromptExecutorFunc(t *testing.T) {
	tests := []struct {
		name         string
		flagValue    string
		expectedFlag string
		quoteType    Quotation
	}{
		{
			name:         "no quotes basic flag value",
			flagValue:    `describing`,
			expectedFlag: `describing`,
			quoteType:    NoQuotes,
		},
		{
			name:         "single quotes basic flag value",
			flagValue:    `describing`,
			expectedFlag: `describing`,
			quoteType:    SingleQuotes,
		},
		{
			name:         "double quotes basic flag value",
			flagValue:    `describing`,
			expectedFlag: `describing`,
			quoteType:    DoubleQuotes,
		},
		{
			name:         "no quotes with escaped quotes",
			flagValue:    `\"describing\'`,
			expectedFlag: `"describing'`,
			quoteType:    NoQuotes,
		},
		{
			name:         "no quotes value with space in between splits flag value",
			flagValue:    `describing stuff`,
			expectedFlag: `describing`,
			quoteType:    NoQuotes,
		},
		{
			name:         "double quotes flag value with space in between",
			flagValue:    `describing stuff`,
			expectedFlag: `describing stuff`,
			quoteType:    DoubleQuotes,
		},
		{
			name:         "single quotes flag value with space in between",
			flagValue:    `describing stuff`,
			expectedFlag: `describing stuff`,
			quoteType:    SingleQuotes,
		},

		{
			name:         "single quotes nested in double quotes",
			flagValue:    `describing 'complex' stuff`,
			expectedFlag: `describing 'complex' stuff`,
			quoteType:    DoubleQuotes,
		},
		{
			name:         "escaped double quotes nested in double quotes",
			flagValue:    `describing \"complex\" stuff`,
			expectedFlag: `describing "complex" stuff`,
			quoteType:    DoubleQuotes,
		},
		{
			name:         "single quotes including escape character",
			flagValue:    `describing \"complex\" stuff`,
			expectedFlag: `describing \"complex\" stuff`,
			quoteType:    SingleQuotes,
		},
		{
			name:         "double quotes nested in single quotes",
			flagValue:    `describing "complex" stuff`,
			expectedFlag: `describing "complex" stuff`,
			quoteType:    SingleQuotes,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commandCalled := false
			cli := newTestCommandWithExpectedFlag(t, tt.expectedFlag, &commandCalled)
			command := &instrumentedCommand{
				Command: cli,
			}
			shellPrompt := &ShellPrompt{RootCmd: command}
			executorFunc := promptExecutorFunc(shellPrompt)
			var format string
			switch tt.quoteType {
			case NoQuotes:
				format = `api --description %s`
			case SingleQuotes:
				format = `api --description '%s'`
			case DoubleQuotes:
				format = `api --description "%s"`
			}
			executorFunc(fmt.Sprintf(format, tt.flagValue))
			require.True(t, commandCalled)
		})
	}
}

func newTestCommandWithExpectedFlag(t *testing.T, expectedFlag string, commandCalled *bool) *cobra.Command {
	cli := new(cobra.Command)
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
