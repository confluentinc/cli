package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-yaml/yaml"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
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
	Dependency string `yaml:"dependency"`
	Version    string `yaml:"version"`
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
	confluentDir := filepath.Join(os.Getenv("HOME"), ".confluent")
	dir, err := os.MkdirTemp(confluentDir, "confluent-plugin-search")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	_, err = clonePluginRepo(dir, cliPluginsUrl)
	if err != nil {
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
		manifestPath := fmt.Sprintf("%s/%s/manifest.yml", dir, file.Name())
		if file.IsDir() && utils.DoesPathExist(manifestPath) {
			manifestFile, err := os.ReadFile(manifestPath)
			if err != nil {
				return nil, err
			}

			manifest := new(Manifest)
			if err := yaml.Unmarshal(manifestFile, manifest); err != nil {
				return nil, err
			}
			manifestOutList = append(manifestOutList, &ManifestOut{
				Name:         file.Name(),
				Description:  manifest.Description,
				Dependencies: dependenciesToString(manifest.Dependencies),
			})
		}
	}

	return manifestOutList, nil
}

func dependenciesToString(dependencies []Dependency) string {
	var dependencyString string
	for _, dependency := range dependencies {
		if dependency.Dependency != "" {
			dependencyString = fmt.Sprintf("%s %s", dependencyString, dependency.Dependency)
			if dependency.Version != "" {
				dependencyString = fmt.Sprintf("%s %s", dependencyString, dependency.Version)
			}
			dependencyString = fmt.Sprintf("%s,", dependencyString)
		}
	}

	return strings.Trim(dependencyString, ", ")
}
