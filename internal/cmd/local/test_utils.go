package local

import (
	"github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/mock"
)

func mockLocalCommand(args... string) (string, error) {
	mockPrerunner := mock.NewPreRunnerMock(nil, nil)
	mockCfg := &v3.Config{}

	command := cmd.BuildRootCommand()
	command.AddCommand(NewCommand(mockPrerunner, mockCfg))

	args = append([]string{"local-v2"}, args...)
	return cmd.ExecuteCommand(command, args...)
}
