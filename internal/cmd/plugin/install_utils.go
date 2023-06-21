package plugin

import (
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/go-version"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

var supportedLanguages = []string{"Python", "Go"}

func getLanguage(manifest *Manifest) (string, *version.Version, error) {
	if manifest == nil {
		return "", nil, nil
	}

	if manifest.Dependencies == "" {
		return "", nil, nil
	}

	dependencySlice := strings.Split(manifest.Dependencies, " ")
	if len(dependencySlice) == 1 {
		return dependencySlice[0], nil, nil
	}

	ver, err := version.NewVersion(dependencySlice[1])
	if err != nil {
		return dependencySlice[0], nil, nil
	}

	return dependencySlice[0], ver, nil
}

func installPythonPlugin(name, repoDir, installDir string) error {
	pluginDir := fmt.Sprintf("%s/%s", repoDir, name)
	files, err := os.ReadDir(pluginDir)
	if err != nil {
		return err
	}

	found := false
	for _, file := range files {
		if strings.HasPrefix(file.Name(), "confluent-") && strings.HasSuffix(file.Name(), ".py") {
			found = true

			fileData, err := os.ReadFile(fmt.Sprintf("%s/%s", pluginDir, file.Name()))
			if err != nil {
				return err
			}

			if err := os.WriteFile(fmt.Sprintf("%s/%s", installDir, file.Name()), fileData, 0755); err != nil {
				return err
			}
		}
	}

	if !found {
		return errors.Errorf("unable to find .py file for plugin %s", name)
	}
	return nil
}
