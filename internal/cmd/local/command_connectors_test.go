package local

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/mock"
)

func TestListConnectors(t *testing.T) {
	req := require.New(t)

	out, err := mockConnectorsCommand("list")
	req.NoError(err)
	req.Contains(out, buildTabbedList(connectors))
}

func mockConnectorsCommand(args... string) (string, error) {
	mockPrerunner := mock.NewPreRunnerMock(nil, nil)
	mockCfg := &v3.Config{}

	command := cmd.BuildRootCommand()
	command.AddCommand(NewCommand(mockPrerunner, mockCfg))

	args = append([]string{"local-v2", "connectors"}, args...)
	return cmd.ExecuteCommand(command, args...)
}
