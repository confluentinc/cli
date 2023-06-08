package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-yaml/yaml"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/types"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

type Manifest struct {
	Number       string
	Name         string `yaml:"name"`
	Description  string `yaml:"description"`
	Requirements string `yaml:"requirements"`
	Location     string
}

type manifestOut struct {
	Number       string `human:"Number" serialized:"number"`
	Name         string `human:"Name" serialized:"name"`
	Description  string `human:"Description" serialized:"description"`
	Requirements string `human:"Requirements" serialized:"requirements"`
}

func (c *command) newSearchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search for and install Confluent CLI plugins.",
		Long:  "Search for Confluent CLI plugins in the Confluent CLI plugin repository, and prompt the user to select plugins to install.",
		Args:  cobra.NoArgs,
		RunE:  c.search,
	}

	cmd.Flags().String("plugin-directory", "", "The plugin installation directory; this must be a directory in your $PATH. If not specified, a default will be selected based on your OS.")

	return cmd
}

func (c *command) search(cmd *cobra.Command, _ []string) error {
	installDir, err := getPluginInstallDir(cmd)
	if err != nil {
		return err
	}

	dir, err := os.MkdirTemp("", "plugin-search")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	if err := clonePluginRepo(dir); err != nil {
		return err
	}

	manifests, err := getPluginManifests(dir)
	if err != nil {
		return err
	}

	manifests, err = selectPlugins(cmd, manifests)
	if err != nil {
		return err
	}

	return installPlugins(installDir, manifests)
}

func getPluginInstallDir(cmd *cobra.Command) (string, error) {
	if !cmd.Flags().Changed("plugin-directory") {
		return getDefaultPluginInstallDir()
	}

	pluginDir, err := cmd.Flags().GetString("plugin-directory")
	if err != nil {
		return "", err
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
	if runtime.GOOS == "windows" {
		// TODO: IMPLEMENT THIS BEFORE MERGING
		return "", nil
	}

	defaultDir := "/usr/local/bin"
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

func clonePluginRepo(dir string) error {
	cloneOptions := &git.CloneOptions{
		URL:          "https://github.com/confluentinc/cli-plugins.git",
		SingleBranch: true, // this should be redundant w/ Depth=1, but specify it just in case
		Depth:        1,
	}
	_, err := git.PlainClone(dir, false, cloneOptions)

	return err
}

func getPluginManifests(dir string) ([]*Manifest, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	manifests := []*Manifest{}
	for _, file := range files {
		manifestPath := fmt.Sprintf("%s/%s/manifest.yml", dir, file.Name())
		if file.IsDir() && utils.DoesPathExist(manifestPath) {
			manifestFile, err := os.ReadFile(manifestPath)
			if err != nil {
				return nil, err
			}

			manifest := new(Manifest)
			manifests = append(manifests, manifest)
			if err := yaml.Unmarshal(manifestFile, manifest); err != nil {
				return nil, err
			}
			manifest.Number = strconv.Itoa(len(manifests))
			manifest.Location = fmt.Sprintf("%s/%s", dir, file.Name())
		}
	}

	return manifests, nil
}

func selectPlugins(cmd *cobra.Command, manifests []*Manifest) ([]*Manifest, error) {
	list := output.NewList(cmd)
	for _, manifest := range manifests {
		list.Add(&manifestOut{
			Number:       manifest.Number,
			Name:         manifest.Name,
			Description:  manifest.Description,
			Requirements: manifest.Requirements,
		})
	}
	listStr, err := list.PrintString()
	if err != nil {
		return nil, err
	}

	prompt := form.NewPrompt(os.Stdin)
	promptMsg := "Enter a single number or a comma-separated list of numbers to install plugins:\n%sTo cancel, press Ctrl-C"
	f := form.New(form.Field{
		ID:     "plugin numbers",
		Prompt: fmt.Sprintf(promptMsg, listStr),
		Regex:  `^(?:\d+)(?:,\d+)*$`,
	})
	if err := f.Prompt(prompt); err != nil {
		return nil, err
	}

	inputStrings := types.RemoveDuplicates(strings.Split(f.Responses["plugin numbers"].(string), ","))
	selectedManifests := make([]*Manifest, len(inputStrings))
	for i, inputStr := range inputStrings {
		num, err := strconv.Atoi(inputStr)
		if err != nil {
			return nil, err
		}
		if num < 1 || num > len(manifests) {
			return nil, errors.Errorf(`Input "%s" must be a number between 1 and %d`, inputStr, len(manifests))
		}

		selectedManifests[i] = manifests[num-1]
	}

	return selectedManifests, nil
}

func installPlugins(installDir string, manifests []*Manifest) error {
	return nil
}
