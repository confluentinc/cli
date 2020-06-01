package local

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/mock"
)

var confluentHome = filepath.Join(os.TempDir(), "confluent")

func TestConfluentCommunitySoftwareVersion(t *testing.T) {
	req := require.New(t)

	req.NoError(setupConfluentHome())
	defer req.NoError(teardownConfluentHome())

	file := strings.Replace(versionFiles["Confluent Community Software"], "*", "0.0.0", 1)
	req.NoError(addFileToConfluentHome(file))

	testVersion(t, []string{}, "Confluent Community Software: 0.0.0")
}

func TestConfluentPlatformVersion(t *testing.T) {
	req := require.New(t)

	req.NoError(setupConfluentHome())
	defer req.NoError(teardownConfluentHome())

	file := strings.Replace(versionFiles["Confluent Platform"], "*", "1.0.0", 1)
	req.NoError(addFileToConfluentHome(file))

	testVersion(t, []string{}, "Confluent Platform: 1.0.0")
}

func TestServiceVersions(t *testing.T) {
	req := require.New(t)

	req.NoError(setupConfluentHome())
	defer req.NoError(teardownConfluentHome())

	services := []string{"kafka", "zookeeper"}
	versions := []string{"2.0.0", "3.0.0"}

	for i := 0; i < len(services); i++ {
		service := services[i]
		version := versions[i]

		file := strings.Replace(versionFiles[service], "*", version, 1)
		req.NoError(addFileToConfluentHome(file))
		testVersion(t, []string{service}, version)
	}
}

func setupConfluentHome() error {
	return os.Setenv("CONFLUENT_HOME", confluentHome)
}

func teardownConfluentHome() error {
	return os.RemoveAll(confluentHome)
}

func addFileToConfluentHome(file string) error {
	path := filepath.Join(confluentHome, file)

	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		return err
	}

	if _, err := os.Create(path); err != nil {
		return err
	}

	return nil
}

func testVersion(t *testing.T, args []string, version string) {
	req := require.New(t)

	mockPrerunner := mock.NewPreRunnerMock(nil, nil)
	mockCfg := &v3.Config{}

	command := cmd.BuildRootCommand()
	command.AddCommand(NewCommand(mockPrerunner, mockCfg))

	args = append([]string{"local-v2", "version"}, args...)
	out, err := cmd.ExecuteCommand(command, args...)

	req.NoError(err)
	req.Contains(out, version)
}
