package plugin

import (
	"fmt"
	"os"

	"github.com/hashicorp/go-version"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

const (
	programNotFoundErrorMsg      = "unable to find %s"
	programNotFoundSuggestions   = "Check that it is installed in a directory in your $PATH."
	unableToParseVersionErrorMsg = "unable to parse %s version"
	insufficientVersionErrorMsg  = "installed %s version %s is less than the required version %s"
)

type PluginInstaller interface {
	IsVersion(word string) bool
	CheckVersion(ver *version.Version) error
	Install() error
}

func installSimplePlugin(name, repositoryDir, installDir, language string) error {
	pluginDir := fmt.Sprintf("%s/%s", repositoryDir, name)
	entries, err := os.ReadDir(pluginDir)
	if err != nil {
		return err
	}

	found := false
	for _, entry := range entries {
		if PluginFromEntry(entry) != "" {
			found = true

			fileData, err := os.ReadFile(fmt.Sprintf("%s/%s", pluginDir, entry.Name()))
			if err != nil {
				return err
			}

			if err := os.WriteFile(fmt.Sprintf("%s/%s", installDir, entry.Name()), fileData, 0755); err != nil {
				return err
			}
		}
	}

	if !found {
		return errors.Errorf("unable to find %s file for plugin %s", language, name)
	}
	return nil
}
