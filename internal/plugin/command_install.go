package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-yaml/yaml"
	"github.com/hashicorp/go-version"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/plugin"
	"github.com/confluentinc/cli/v3/pkg/utils"
)

func (c *command) newInstallCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "install <plugin>",
		Short: "Install or update official Confluent CLI plugins.",
		Long:  "Install official Confluent CLI plugins from the confluentinc/cli-plugins repository or update existing plugins.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.install,
	}
}

func (c *command) install(cmd *cobra.Command, args []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	confluentDir := filepath.Join(home, ".confluent")
	dir, err := os.MkdirTemp(confluentDir, "cli-plugins")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	installDir := filepath.Join(confluentDir, "plugins")
	if err := os.MkdirAll(installDir, os.ModePerm); err != nil {
		return errors.Wrapf(err, "failed to create plugin install directory %s", installDir)
	}

	if _, err := clonePluginRepo(dir, cliPluginsUrl); err != nil {
		return err
	}

	manifest, err := getPluginManifest(args[0], dir)
	if err != nil {
		return err
	}

	if err := installPlugin(manifest, dir, installDir); err != nil {
		return err
	}

	output.Printf("Installed plugin \"%s\".\n", args[0])

	return nil
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

	var pluginInstaller plugin.PluginInstaller
	switch language {
	case "go":
		pluginInstaller = &plugin.GoPluginInstaller{Name: manifest.Name}
	case "python":
		pluginInstaller = &plugin.PythonPluginInstaller{
			Name:          manifest.Name,
			RepositoryDir: repositoryDir,
			InstallDir:    installDir,
		}
	case "bash":
		pluginInstaller = &plugin.BashPluginInstaller{
			Name:          manifest.Name,
			RepositoryDir: repositoryDir,
			InstallDir:    installDir,
		}
	default:
		return errors.Errorf("installation of plugins using %s is not yet supported", language)
	}

	if err := pluginInstaller.CheckVersion(ver); err != nil {
		return err
	}
	return pluginInstaller.Install()
}

func getLanguage(manifest *Manifest) (string, *version.Version, error) {
	if manifest == nil || len(manifest.Dependencies) == 0 {
		return "", nil, errors.New("no language found in plugin manifest")
	}

	language := manifest.Dependencies[0]
	language.Name = strings.ToLower(language.Name)
	if language.Version == "" {
		return "", nil, errors.New(errors.NoVersionFoundErrorMsg)
	}

	ver, err := version.NewVersion(language.Version)
	if err != nil {
		return "", nil, err
	}

	return language.Name, ver, nil
}
