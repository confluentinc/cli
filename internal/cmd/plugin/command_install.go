package plugin

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"

	"github.com/go-yaml/yaml"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/types"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *command) newInstallCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install official Confluent CLI plugins.",
		Long:  "Install official Confluent CLI plugins from the confluentinc/cli-plugins repository.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.install,
	}

	cmd.Flags().String("plugin-directory", "", "The plugin installation directory; this must be a directory in your $PATH. If not specified, a default will be selected based on your OS.")

	return cmd
}

func (c *command) install(cmd *cobra.Command, args []string) error {
	installDir, err := getPluginInstallDir(cmd)
	if err != nil {
		return err
	}

	dir, err := os.MkdirTemp("", "plugin-install")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	if _, err = clonePluginRepo(dir, cliPluginsUrl); err != nil {
		return err
	}

	manifest, err := getPluginManifest(args[0], dir)
	if err != nil {
		return err
	}

	return installPlugin(manifest, dir, installDir)
}

func getPluginInstallDir(cmd *cobra.Command) (string, error) {
	pluginDir, err := cmd.Flags().GetString("plugin-directory")
	if err != nil {
		return "", err
	}
	if pluginDir == "" {
		return getDefaultPluginInstallDir()
	}

	if pluginDir, err = filepath.Abs(pluginDir); err != nil {
		return "", err
	}

	if !utils.DoesPathExist(pluginDir) {
		return "", errors.Errorf(`plugin directory "%s" does not exist`, pluginDir)
	}

	if !inPath(pluginDir) {
		output.Printf("WARNING: failed to find installation directory \"%s\" in your $PATH.\n\n", pluginDir)
	}

	return pluginDir, nil
}

func getDefaultPluginInstallDir() (string, error) {
	// Windows: CLI installation directory
	// Unix:    /usr/local/bin
	defaultDir := "/usr/local/bin"
	if runtime.GOOS == "windows" {
		cliPath, err := os.Executable()
		if err != nil {
			return "", err
		}

		// Check if the path is a symlink, since os.Executable does not always return
		// the actual path if the process is started from a symlink
		file, err := os.Lstat(cliPath)
		if err != nil {
			return "", err
		}
		if file.Mode()&fs.ModeSymlink != 0 {
			return "", errors.NewErrorWithSuggestions("unable to select a suitable default installation directory", "Pass an installation directory in your $PATH with the `--plugin-directory` flag.")
		}

		defaultDir = filepath.Dir(cliPath)
	}

	if !inPath(defaultDir) {
		output.Printf("WARNING: failed to find default directory \"%s\" in your $PATH.\n\n", defaultDir)
	}

	return defaultDir, nil
}

func inPath(dir string) bool {
	pathDirectories := filepath.SplitList(os.Getenv("PATH"))
	for i := range pathDirectories {
		absPath, _ := filepath.Abs(pathDirectories[i])
		pathDirectories[i] = absPath
	}
	return types.Contains(pathDirectories, dir)
}

func getPluginManifest(pluginName, dir string) (*Manifest, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.Name() != pluginName || !file.IsDir() {
			continue
		}

		manifestPath := fmt.Sprintf("%s/%s/manifest.yml", dir, file.Name())
		if !utils.DoesPathExist(manifestPath) {
			return nil, errors.Errorf("manifest not found for plugin %s", pluginName)
		}

		manifestFile, err := os.ReadFile(manifestPath)
		if err != nil {
			return nil, err
		}

		manifest := new(Manifest)
		if err := yaml.Unmarshal(manifestFile, manifest); err != nil {
			return nil, err
		}
		manifest.Name = file.Name()

		return manifest, nil
	}

	return nil, errors.Errorf("plugin %s not found", pluginName)
}

func installPlugin(manifest *Manifest, repositoryDir, installDir string) error {
	language, ver, err := getLanguage(manifest)
	if err != nil {
		return err
	}

	if !types.Contains(supportedLanguages, language) {
		return errors.Errorf("installation of plugins using %s is not yet supported", language)
	}

	switch language {
	case "python":
		checkPythonVersion(ver)
		return installPythonPlugin(manifest.Name, repoDir, installDir)
	case "go":
		checkGoVersion(ver)
		return installGoPlugin(manifest.Name)
	}

	return nil
}
