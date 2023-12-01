package plugin

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-version"
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
	pluginDir := filepath.Join(repositoryDir, name)

	entries, err := os.ReadDir(pluginDir)
	if err != nil {
		return err
	}

	found := false
	for _, entry := range entries {
		fmt.Println("DEBUG", entry.Name())
		if nameFromEntry(entry) != "" {
			found = true

			fileData, err := os.ReadFile(filepath.Join(pluginDir, entry.Name()))
			if err != nil {
				return err
			}

			if err := os.WriteFile(filepath.Join(installDir, entry.Name()), fileData, 0755); err != nil {
				return err
			}
		}
	}

	if !found {
		return fmt.Errorf("unable to find %s file for plugin %s", language, name)
	}
	return nil
}
