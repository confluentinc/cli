package feedback

import (
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	cliMock "github.com/confluentinc/cli/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestFeedback(t *testing.T) {
	cmd := pcmd.BuildRootCommand()
	feedbackCmd := NewFeedbackCmd(cliMock.NewPreRunnerMock(nil, nil), nil)
	cmd.AddCommand(feedbackCmd)

	req := require.New(t)
	out, err := pcmd.ExecuteCommand(cmd, "feedback", "This feedback command is great!")
	req.NoError(err)
	req.Regexp("Thanks for your feedback.", out)
}
