package feedback

import (
	"testing"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	pmock "github.com/confluentinc/cli/internal/pkg/mock"
	climock "github.com/confluentinc/cli/mock"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestFeedback(t *testing.T) {
	command := pcmd.BuildRootCommand()
	command.AddCommand(mockFeedbackCommand("This feedback tool is great!"))

	req := require.New(t)
	out, err := pcmd.ExecuteCommand(command, "feedback")
	req.NoError(err)
	req.Contains(out, "Enter feedback: ")
	req.Contains(out, errors.ThanksForFeedbackMsg)
}

func TestFeedbackEmptyMessage(t *testing.T) {
	command := pcmd.BuildRootCommand()
	command.AddCommand(mockFeedbackCommand(""))

	req := require.New(t)
	out, err := pcmd.ExecuteCommand(command, "feedback")
	req.NoError(err)
	req.Contains(out, "Enter feedback: ")
}

func mockFeedbackCommand(msg string) *cobra.Command {
	mockPreRunner := climock.NewPreRunnerMock(nil, nil, nil, nil, nil)
	mockPrompt := pmock.NewPromptMock(msg)
	return NewFeedbackCmdWithPrompt(mockPreRunner, mockPrompt)
}
