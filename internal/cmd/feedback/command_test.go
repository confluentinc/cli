package feedback

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/mock"
)

func TestFeedback(t *testing.T) {
	cmd := pcmd.BuildRootCommand()
	cmd.AddCommand(mockFeedbackCommand("This feedback tool is great!"))

	req := require.New(t)
	out, err := pcmd.ExecuteCommand(cmd, "feedback")
	req.NoError(err)
	req.Contains(out, "Enter feedback: ")
	req.Contains(out, "Thanks for your feedback.")
}

func TestFeedbackEmptyMessage(t *testing.T) {
	cmd := pcmd.BuildRootCommand()
	cmd.AddCommand(mockFeedbackCommand(""))

	req := require.New(t)
	out, err := pcmd.ExecuteCommand(cmd, "feedback")
	req.NoError(err)
	req.Contains(out, "Enter feedback: ")
}

func mockFeedbackCommand(msg string) *cobra.Command {
	mockPreRunner := mock.NewPreRunnerMock(nil, nil)
	mockAnalytics := mock.NewDummyAnalyticsMock()
	mockPrompt := mock.NewPromptMock(msg)
	return NewFeedbackCmdWithPrompt(mockPreRunner, nil, mockAnalytics, mockPrompt)
}
