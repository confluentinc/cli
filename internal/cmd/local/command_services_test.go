package local

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/mock"
)

func TestListConfluentPlatformServices(t *testing.T) {
	req := require.New(t)

	confluentHome := mock.NewConfluentHomeMock()
	req.NoError(confluentHome.Setup())
	defer req.NoError(confluentHome.TearDown())

	file := strings.Replace(confluentControlCenter, "*", "0.0.0", 1)
	req.NoError(confluentHome.AddFile(file))

	out, err := mockServicesCommand("list")
	req.NoError(err)
	allServices := append(services, confluentPlatformServices...)
	req.Contains(out, buildTabbedList(allServices))
}

func TestListServicesNoConfluentPlatform(t *testing.T) {
	req := require.New(t)

	confluentHome := mock.NewConfluentHomeMock()
	req.NoError(confluentHome.Setup())
	defer req.NoError(confluentHome.TearDown())

	out, err := mockServicesCommand("list")
	req.NoError(err)
	req.Contains(out, buildTabbedList(services))
}

func TestServiceVersions(t *testing.T) {
	req := require.New(t)

	confluentHome := mock.NewConfluentHomeMock()
	req.NoError(confluentHome.Setup())
	defer req.NoError(confluentHome.TearDown())

	file := strings.Replace(confluentControlCenter, "*", "0.0.0", 1)
	req.NoError(confluentHome.AddFile(file))

	versions := map[string]string{
		"Confluent Platform": "1.0.0",
		"kafka":              "2.0.0",
		"zookeeper":          "3.0.0",
	}

	for service, version := range versions {
		file := strings.Replace(versionFiles[service], "*", version, 1)
		req.NoError(confluentHome.AddFile(file))
	}

	for _, service := range services {
		out, err := mockServicesCommand(service, "version")
		req.NoError(err)

		version, ok := versions[service]
		if !ok {
			version = versions["Confluent Platform"]
		}
		req.Contains(out, version)
	}
}

func mockServicesCommand(args ...string) (string, error) {
	mockPrerunner := mock.NewPreRunnerMock(nil, nil)
	mockCfg := &v3.Config{}

	command := cmd.BuildRootCommand()
	command.AddCommand(NewCommand(mockPrerunner, mockCfg))

	args = append([]string{"local-v2", "services"}, args...)
	return cmd.ExecuteCommand(command, args...)
}
