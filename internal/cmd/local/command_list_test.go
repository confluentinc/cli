package local

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/mock"
)

func TestListConfluentPlatform(t *testing.T) {
	req := require.New(t)

	confluentHome := mock.NewConfluentHomeMock()
	req.NoError(confluentHome.Setup())
	defer req.NoError(confluentHome.TearDown())

	file := strings.Replace(confluentControlCenter, "*", "0.0.0", 1)
	req.NoError(confluentHome.AddFile(file))

	out, err := runList()
	req.NoError(err)
	allServices := append(services, confluentPlatformServices...)
	req.Contains(out, buildTabbedList(allServices))
}

func TestListNoConfluentPlatform(t *testing.T) {
	req := require.New(t)

	confluentHome := mock.NewConfluentHomeMock()
	req.NoError(confluentHome.Setup())
	defer req.NoError(confluentHome.TearDown())

	out, err := runList()
	req.NoError(err)
	req.Contains(out, buildTabbedList(services))
}

func TestListConnectors(t *testing.T) {
	req := require.New(t)

	out, err := runList("connectors")
	req.NoError(err)
	req.Contains(out, buildTabbedList(connectors))
}

func runList(args... string) (string, error) {
	mockPrerunner := mock.NewPreRunnerMock(nil, nil)
	mockCfg := &v3.Config{}

	command := cmd.BuildRootCommand()
	command.AddCommand(NewCommand(mockPrerunner, mockCfg))

	args = append([]string{"local-v2", "list"}, args...)
	return cmd.ExecuteCommand(command, args...)
}
