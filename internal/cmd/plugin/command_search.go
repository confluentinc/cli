package plugin

import (
	"fmt"
	"os"
	"strconv"

	"github.com/go-git/go-git/v5"
	"github.com/go-yaml/yaml"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/output"
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

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) search(cmd *cobra.Command, _ []string) error {
	dir, err := os.MkdirTemp("", "plugin-search")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	cloneOptions := &git.CloneOptions{
		URL:          "https://github.com/confluentinc/cli-plugins.git",
		SingleBranch: true, // this should be redundant w/ Depth=1, but specify it just in case
		Depth:        1,
	}
	if _, err := git.PlainClone(dir, false, cloneOptions); err != nil {
		return err
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	i := 1
	for _, file := range files {
		manifestPath := fmt.Sprintf("%s/%s/manifest.yml", dir, file.Name())
		if file.IsDir() && utils.DoesPathExist(manifestPath) {
			manifestFile, err := os.ReadFile(manifestPath)
			if err != nil {
				return err
			}

			manifest := new(Manifest)
			if err := yaml.Unmarshal(manifestFile, manifest); err != nil {
				return err
			}
			manifest.Number = strconv.Itoa(i)
			i++

			list.Add(manifest)
		}
	}

	listStr, err := list.PrintString()
	if err != nil {
		return err
	}

	prompt := form.NewPrompt(os.Stdin)
	promptMsg := "Enter a single number or a comma-separated list of numbers to install plugins:\n%sTo cancel, press Ctrl-C"
	f := form.New(form.Field{
		ID:     "plugin numbers",
		Prompt: fmt.Sprintf(promptMsg, listStr),
		Regex:  `^\d$`,
	})
	if err := f.Prompt(prompt); err != nil {
		return err
	}

	// TODO: Fix prompt regex and handle user input
	return nil
}
