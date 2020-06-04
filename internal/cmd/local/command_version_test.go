package local

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/mock"
)

func TestConfluentCommunitySoftwareVersion(t *testing.T) {
	req := require.New(t)

	confluentHome := mock.NewConfluentHomeMock()
	req.NoError(confluentHome.Setup())
	defer req.NoError(confluentHome.TearDown())

	file := strings.Replace(versionFiles["Confluent Community Software"], "*", "0.0.0", 1)
	req.NoError(confluentHome.AddFile(file))

	out, err := mockVersionCommand()
	req.NoError(err)
	req.Contains(out, "Confluent Community Software: 0.0.0")
}

func TestConfluentPlatformVersion(t *testing.T) {
	req := require.New(t)

	confluentHome := mock.NewConfluentHomeMock()
	req.NoError(confluentHome.Setup())
	defer req.NoError(confluentHome.TearDown())

	file := strings.Replace(confluentControlCenter, "*", "0.0.0", 1)
	req.NoError(confluentHome.AddFile(file))

	file = strings.Replace(versionFiles["Confluent Platform"], "*", "1.0.0", 1)
	req.NoError(confluentHome.AddFile(file))

	out, err := mockVersionCommand()
	req.NoError(err)
	req.Contains(out, "Confluent Platform: 1.0.0")
}

func mockVersionCommand() (string, error) {
	mockPrerunner := mock.NewPreRunnerMock(nil, nil)
	mockCfg := &v3.Config{}

	command := cmd.BuildRootCommand()
	command.AddCommand(NewCommand(mockPrerunner, mockCfg))

	return cmd.ExecuteCommand(command, "local-v2", "version")
}
