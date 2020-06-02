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

	testVersion(req, []string{}, "Confluent Community Software: 0.0.0")
}

func TestConfluentPlatformVersion(t *testing.T) {
	req := require.New(t)

	confluentHome := mock.NewConfluentHomeMock()
	req.NoError(confluentHome.Setup())
	defer req.NoError(confluentHome.TearDown())

	file := strings.Replace(versionFiles["Confluent Platform"], "*", "1.0.0", 1)
	req.NoError(confluentHome.AddFile(file))

	testVersion(req, []string{}, "Confluent Platform: 1.0.0")
}

func TestServiceVersions(t *testing.T) {
	req := require.New(t)

	confluentHome := mock.NewConfluentHomeMock()
	req.NoError(confluentHome.Setup())
	defer req.NoError(confluentHome.TearDown())

	services := []string{"kafka", "zookeeper"}
	versions := []string{"2.0.0", "3.0.0"}

	for i := 0; i < len(services); i++ {
		service := services[i]
		version := versions[i]

		file := strings.Replace(versionFiles[service], "*", version, 1)
		req.NoError(confluentHome.AddFile(file))
		testVersion(req, []string{service}, version)
	}
}

func testVersion(req *require.Assertions, args []string, version string) {
	mockPrerunner := mock.NewPreRunnerMock(nil, nil)
	mockCfg := &v3.Config{}

	command := cmd.BuildRootCommand()
	command.AddCommand(NewCommand(mockPrerunner, mockCfg))

	args = append([]string{"local-v2", "version"}, args...)
	out, err := cmd.ExecuteCommand(command, args...)

	req.NoError(err)
	req.Contains(out, version)
}
