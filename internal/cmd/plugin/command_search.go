package plugin

import (
	"fmt"
	"os"
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
	Number       string `human:"Number" serialized:"number"`
	Name         string `human:"Name" serialized:"name" yaml:"name"`
	Description  string `human:"Description" serialized:"description" yaml:"description"`
	Requirements string `human:"Requirements" serialized:"requirements" yaml:"requirements"`
}

func (c *command) newSearchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search for and install Confluent CLI plugins.",
		Long:  "Search for Confluent CLI plugins in the Confluent CLI plugin repository, and prompt the user to select plugins to install.",
		Args:  cobra.NoArgs,
		RunE:  c.search,
	}

	return cmd
}

func (c *command) search(cmd *cobra.Command, _ []string) error {
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

	selections, err := selectPlugins(cmd, manifests)
	if err != nil {
		return err
	}

	return installPlugins(selections)
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
		}
	}

	return manifests, nil
}

func selectPlugins(cmd *cobra.Command, manifests []*Manifest) ([]int, error) {
	list := output.NewList(cmd)
	for _, manifest := range manifests {
		list.Add(manifest)
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
	inputNumbers := make([]int, len(inputStrings))
	for i, inputStr := range inputStrings {
		num, err := strconv.Atoi(inputStr)
		if err != nil {
			return nil, err
		}
		if num < 1 || num > len(manifests) {
			return nil, errors.Errorf(`Input "%s" must be a number between 1 and %d`, inputStr, len(manifests))
		}
		inputNumbers[i] = num
	}

	return inputNumbers, nil
}

func installPlugins(selections []int) error {
	return nil
}
