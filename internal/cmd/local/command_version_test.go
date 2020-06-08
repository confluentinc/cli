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

func TestConfluentCommunitySoftwareVersion(t *testing.T) {
	req := require.New(t)

	dir := filepath.Join(os.TempDir(), "confluent1")
	req.NoError(setupConfluentHome(dir))
	defer req.NoError(teardownConfluentHome(dir))

	file := strings.Replace(versionFiles["Confluent Community Software"], "*", "0.0.0", 1)
	req.NoError(addFileToConfluentHome(dir, file))

	testVersion(t, []string{}, "Confluent Community Software: 0.0.0")
}

func TestConfluentPlatformVersion(t *testing.T) {
	req := require.New(t)

	dir := filepath.Join(os.TempDir(), "confluent2")
	req.NoError(setupConfluentHome(dir))
	defer req.NoError(teardownConfluentHome(dir))

	file := strings.Replace(versionFiles["Confluent Platform"], "*", "1.0.0", 1)
	req.NoError(addFileToConfluentHome(dir, file))

	testVersion(t, []string{}, "Confluent Platform: 1.0.0")
}

func TestServiceVersions(t *testing.T) {
	req := require.New(t)

	dir := filepath.Join(os.TempDir(), "confluent3")
	req.NoError(setupConfluentHome(dir))
	defer req.NoError(teardownConfluentHome(dir))

	services := []string{"kafka", "zookeeper"}
	versions := []string{"2.0.0", "3.0.0"}

	for i := 0; i < len(services); i++ {
		service := services[i]
		version := versions[i]

		file := strings.Replace(versionFiles[service], "*", version, 1)
		req.NoError(addFileToConfluentHome(dir, file))
		testVersion(t, []string{service}, version)
	}
}

func setupConfluentHome(dir string) error {
	return os.Setenv("CONFLUENT_HOME", dir)
}

func teardownConfluentHome(dir string) error {
	return os.RemoveAll(dir)
}

func addFileToConfluentHome(dir string, file string) error {
	path := filepath.Join(dir, file)

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
