package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/utils"
)

type ManifestOut struct {
	Name         string `human:"Name" serialized:"name"`
	Description  string `human:"Description" serialized:"description"`
	Dependencies string `human:"Dependencies" serialized:"dependencies"`
}

type Manifest struct {
	Name         string
	Description  string       `yaml:"description"`
	Dependencies []Dependency `yaml:"dependencies"`
}

type Dependency struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

func (c *command) newSearchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search for Confluent CLI plugins.",
		Long:  "Search for Confluent CLI plugins in the confluentinc/cli-plugins repository.",
		Args:  cobra.NoArgs,
		RunE:  c.search,
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) search(cmd *cobra.Command, _ []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	dir, err := os.MkdirTemp(filepath.Join(home, ".confluent"), "cli-plugins")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	fmt.Println("DEBUG", dir)

	if _, err := clonePluginRepo(dir, cliPluginsUrl); err != nil {
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

func clonePluginRepo(dir, url string) (*git.Repository, error) {
	cloneOptions := &git.CloneOptions{
		URL:   url,
		Depth: 1,
	}

	return git.PlainClone(dir, false, cloneOptions)
}

func getPluginManifests(dir string) ([]*ManifestOut, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	manifestOutList := []*ManifestOut{}
	for _, file := range files {
		manifestPath := filepath.Join(dir, file.Name(), "manifest.yml")
		if file.IsDir() && utils.DoesPathExist(manifestPath) {
			manifestFile, err := os.ReadFile(manifestPath)
			if err != nil {
				return nil, err
			}

			manifest := new(Manifest)
			if err := yaml.Unmarshal(manifestFile, manifest); err != nil {
				return nil, err
			}
			manifestOut := ManifestOut{
				Name:         file.Name(),
				Description:  manifest.Description,
				Dependencies: strings.Join(dependenciesToStrings(manifest.Dependencies), ", "),
			}
			manifestOutList = append(manifestOutList, &manifestOut)
		}
	}

	return manifestOutList, nil
}

func dependenciesToStrings(dependencies []Dependency) []string {
	dependencyStrings := make([]string, len(dependencies))
	for i, dependency := range dependencies {
		if dependency.Name == "" {
			continue
		}
		dependencyString := dependency.Name
		if dependency.Version != "" {
			dependencyString = fmt.Sprintf("%s %s", dependencyString, dependency.Version)
		}
		dependencyStrings[i] = dependencyString
	}

	return dependencyStrings
}
