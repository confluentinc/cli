package local

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/mock"
)

func findConfluentFile(pattern string) ([]string, error) {
	confluentHome, err := getConfluentHome()
	if err != nil {
		return []string{}, err
	}

	path := filepath.Join(confluentHome, pattern)
	matches, err := filepath.Glob(path)
	if err != nil {
		return []string{}, err
	}

	for i := range matches {
		matches[i], err = filepath.Rel(confluentHome, matches[i])
		if err != nil {
			return []string{}, err
		}
	}
	return matches, nil
}

func getConfluentHome() (string, error) {
	confluentHome := os.Getenv("CONFLUENT_HOME")
	if confluentHome == "" {
		return "", fmt.Errorf("set environment variable CONFLUENT_HOME")
	}
	return confluentHome, nil
}

func isConfluentPlatform() (bool, error) {
	controlCenter := "share/java/confluent-control-center/control-center-*.jar"
	files, err := findConfluentFile(controlCenter)
	if err != nil {
		return false, err
	}
	return len(files) > 0, nil
}

func buildTabbedList(slice []string) string {
	sort.Strings(slice)

	var list strings.Builder
	for _, x := range slice {
		fmt.Fprintf(&list, "  %s\n", x)
	}
	return list.String()
}

func mockLocalCommand(args... string) (string, error) {
	mockPrerunner := mock.NewPreRunnerMock(nil, nil)
	mockCfg := &v3.Config{}

	command := cmd.BuildRootCommand()
	command.AddCommand(NewCommand(mockPrerunner, mockCfg))

	args = append([]string{"local-v2"}, args...)
	return cmd.ExecuteCommand(command, args...)
}
