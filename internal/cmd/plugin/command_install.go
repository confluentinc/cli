package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/go-yaml/yaml"
	"github.com/hashicorp/go-version"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/exec"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

const (
	programNotFoundMsg      = "[WARN] Unable to find %s. Check that it is installed in a directory in your $PATH.\n"
	unableToParseVersionMsg = "[WARN] Unable to parse %s version.\n"
	insufficientVersionMsg  = "[WARN] Installed %s version %s is less than the required version %s.\n"
)

func (c *command) newInstallCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "install",
		Short: "Install official Confluent CLI plugins.",
		Long:  "Install official Confluent CLI plugins from the confluentinc/cli-plugins repository.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.install,
	}
}

func (c *command) install(cmd *cobra.Command, args []string) error {
	confluentDir := filepath.Join(os.Getenv("HOME"), ".confluent")
	dir, err := os.MkdirTemp(confluentDir, "cli-plugins")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	installDir := filepath.Join(confluentDir, "plugins")
	if err := os.MkdirAll(installDir, os.ModePerm); err != nil {
		return errors.Wrapf(err, "failed to create plugin install directory %s", installDir)
	}

	if _, err = clonePluginRepo(dir, cliPluginsUrl); err != nil {
		return err
	}

	manifest, err := getPluginManifest(args[0], dir)
	if err != nil {
		return err
	}

	if err := installPlugin(manifest, dir, installDir); err != nil {
		return err
	}

	output.Printf("Installed plugin %s.\n", args[0])

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
	language, ver := getLanguage(manifest)

	switch language {
	case "go":
		checkGoVersion(ver)
		return installGoPlugin(manifest.Name)
	case "python":
		checkPythonVersion(ver)
		return installSimplePlugin(manifest.Name, repositoryDir, installDir, "python")
	case "shell":
		return installSimplePlugin(manifest.Name, repositoryDir, installDir, "shell")
	default:
		return errors.Errorf("installation of plugins using %s is not yet supported", language)
	}
}

func getLanguage(manifest *Manifest) (string, *version.Version) {
	if manifest == nil || len(manifest.Dependencies) == 0 {
		return "", nil
	}

	language := manifest.Dependencies[0]
	language.Dependency = strings.ToLower(language.Dependency)
	if language.Version == "" {
		return language.Dependency, nil
	}

	ver, err := version.NewVersion(language.Version)
	if err != nil {
		return language.Dependency, nil
	}

	return language.Dependency, ver
}

func checkPythonVersion(ver *version.Version) {
	versionCmd := exec.NewCommand("python", "--version")

	out, err := versionCmd.Output()
	if err != nil {
		output.ErrPrintf(programNotFoundMsg, "python")
		return
	}

	re := regexp.MustCompile(`^[1-9][0-9]*\.[0-9]+\.(0|[1-9][0-9]*)$`)
	for _, word := range strings.Split(string(out), " ") {
		if re.MatchString(word) {
			installedVer, err := version.NewVersion(strings.Trim(word, " \n"))
			if err != nil {
				output.ErrPrintf(unableToParseVersionMsg, "python")
				return
			}
			if installedVer.LessThan(ver) {
				output.ErrPrintf(insufficientVersionMsg, "python", installedVer, ver)
				return
			}
		}
	}
}

func checkGoVersion(ver *version.Version) {
	versionCmd := exec.NewCommand("go", "version")

	out, err := versionCmd.Output()
	if err != nil {
		output.ErrPrintf(programNotFoundMsg, "go")
		return
	}

	re := regexp.MustCompile(`^go[1-9][0-9]*\.[0-9]+(\.[1-9][0-9]*)?$`)
	for _, word := range strings.Split(string(out), " ") {
		if re.MatchString(word) {
			installedVer, err := version.NewVersion(strings.TrimPrefix(word, "go"))
			if err != nil {
				output.ErrPrintf(unableToParseVersionMsg, "go")
				return
			}
			if installedVer.LessThan(ver) {
				output.ErrPrintf(insufficientVersionMsg, "go", installedVer, ver)
				return
			}
		}
	}
}

func installSimplePlugin(name, repositoryDir, installDir, language string) error {
	pluginDir := fmt.Sprintf("%s/%s", repositoryDir, name)
	files, err := os.ReadDir(pluginDir)
	if err != nil {
		return err
	}

	found := false
	for _, file := range files {
		if strings.HasPrefix(file.Name(), "confluent-") {
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
		return errors.Errorf("unable to find %s file for plugin %s", language, name)
	}
	return nil
}

func installGoPlugin(name string) error {
	packageName := fmt.Sprintf("github.com/confluentinc/cli-plugins/%s@latest", name)
	installCmd := exec.NewCommand("go", "install", packageName)

	if _, err := installCmd.Output(); err != nil {
		return errors.Wrap(err, "failed to run go install command")
	}

	return nil
}
