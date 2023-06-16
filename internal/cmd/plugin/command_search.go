package plugin

import (
	"fmt"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-yaml/yaml"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

type Manifest struct {
	Name         string `human:"Name" serialized:"name"`
	Description  string `yaml:"description" human:"Description" serialized:"description"`
	Dependencies string `yaml:"dependencies" human:"Dependencies" serialized:"dependencies"`
}

func (c *command) newSearchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search for available Confluent CLI plugins.",
		Long:  "Search for available Confluent CLI plugins in the Confluent CLI plugin repository.",
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

	if err := clonePluginRepo(dir); err != nil {
		return err
	}

	manifests, err := getPluginManifests(dir)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, manifest := range manifests {
		list.Add(manifest)
	}

	return list.Print()
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
			manifest.Name = file.Name()
		}
	}

	return manifests, nil
}
