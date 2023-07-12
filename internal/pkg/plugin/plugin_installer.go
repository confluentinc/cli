package plugin

import (
	"fmt"
	"os"

	"github.com/hashicorp/go-version"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

const (
	programNotFoundMsg      = "[WARN] Unable to find %s. Check that it is installed in a directory in your $PATH.\n"
	unableToParseVersionMsg = "[WARN] Unable to parse %s version.\n"
	insufficientVersionMsg  = "installed %s version %s is less than the required version %s"
)

type PluginInstaller interface {
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
		if name := PluginFromEntry(entry); name != "" {
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
