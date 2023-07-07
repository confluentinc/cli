package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/confluentinc/cli/internal/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/docs"
	pversion "github.com/confluentinc/cli/internal/pkg/version"
)

// Auto-generate documentation files for all CLI commands. Documentation uses reStructured Text (ReST) format, and is
// nested in a hierarchical format. Commands may differ based on the login context found in the user's config file.
// We use tabs to show differing cloud and on-prem commands on the same page.
// This code is adapted from https://github.com/spf13/cobra/blob/master/doc/rest_docs.md

func main() {
	// Prevent printing the user's HOME in docs when generating confluent local services kafka
	currentHOME := os.Getenv("HOME")
	if err := os.Setenv("HOME", "$HOME"); err != nil {
		panic(err)
	}

	// Generate documentation for both subsets of commands: cloud and on-prem
	configs := []*v1.Config{
		{CurrentContext: "Cloud", Contexts: map[string]*v1.Context{"Cloud": {PlatformName: "https://confluent.cloud"}}},
		{CurrentContext: "On-Premises", Contexts: map[string]*v1.Context{"On-Premises": {PlatformName: "https://example.com"}}},
	}

	tabs := make([]docs.Tab, len(configs))
	for i, cfg := range configs {
		cfg.IsTest = true
		cfg.Version = new(pversion.Version)

		tabs[i] = docs.Tab{
			Name:    cfg.CurrentContext,
			Command: cmd.NewConfluentCommand(cfg),
		}
	}

	if err := os.Mkdir("docs", os.ModePerm); err != nil {
		panic(err)
	}

	if err := docs.GenerateDocTree(tabs, "docs", 0); err != nil {
		panic(err)
	}

	removeUnreleasedDocs()

	if err := os.Setenv("HOME", currentHOME); err != nil {
		panic(err)
	}
}

// removeUnreleasedDocs hides documentation for unreleased features
func removeUnreleasedDocs() {
	removeUnreleasedCommands("flink")
}

func removeUnreleasedCommands(command string) {
	if err := removeLineFromFile(fmt.Sprintf(`\s{3}%s/index\n`, command), filepath.Join("docs", "index.rst")); err != nil {
		panic(err)
	}

	if err := removeLineFromFile(fmt.Sprintf("\\s{7}:ref:`confluent_%s`\\s+.+\\s+\n", command), filepath.Join("docs", "overview.rst")); err != nil {
		panic(err)
	}

	if err := os.RemoveAll(filepath.Join("docs", command)); err != nil {
		panic(err)
	}
}

func removeLineFromFile(line, file string) error {
	out, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	re := regexp.MustCompile(line)
	out = re.ReplaceAll(out, []byte(""))

	return os.WriteFile(file, out, 0644)
}
