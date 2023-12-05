package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/confluentinc/cli/v3/internal"
	"github.com/confluentinc/cli/v3/pkg/config"
	"github.com/confluentinc/cli/v3/pkg/docs"
	pversion "github.com/confluentinc/cli/v3/pkg/version"
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
	configs := []*config.Config{
		{CurrentContext: "Cloud", Contexts: map[string]*config.Context{"Cloud": {PlatformName: "https://confluent.cloud"}}},
		{CurrentContext: "On-Premises", Contexts: map[string]*config.Context{"On-Premises": {PlatformName: "https://example.com"}}},
	}

	tabs := make([]docs.Tab, len(configs))
	for i, cfg := range configs {
		cfg.IsTest = true
		cfg.Version = new(pversion.Version)

		tabs[i] = docs.Tab{
			Name:    cfg.CurrentContext,
			Command: internal.NewConfluentCommand(cfg),
		}
	}

	if err := os.Mkdir("docs", os.ModePerm); err != nil {
		panic(err)
	}

	if err := docs.GenerateDocTree(tabs, "docs", 0); err != nil {
		panic(err)
	}

	if err := os.Setenv("HOME", currentHOME); err != nil {
		panic(err)
	}
}

func removeUnreleasedCommands(command string) { //nolint:unused
	subcommands := strings.Split(command, " ")

	line := fmt.Sprintf(`\s{3}%s/index\n`, subcommands[len(subcommands)-1])
	file := filepath.Join(append(append([]string{"docs"}, subcommands[:len(subcommands)-1]...), "index.rst")...)
	if err := removeLineFromFile(line, file); err != nil {
		panic(err)
	}

	line = fmt.Sprintf("\\s{7}:ref:`confluent_%s`\\s+.+\\s+\n", strings.Join(subcommands, "_"))
	if len(subcommands) == 1 {
		file = filepath.Join("docs", "overview.rst")
		if err := removeLineFromFile(line, file); err != nil {
			panic(err)
		}
	} else {
		if err := removeLineFromFile(line, file); err != nil {
			panic(err)
		}
	}

	path := filepath.Join(append([]string{"docs"}, subcommands...)...)
	if err := os.RemoveAll(path); err != nil {
		panic(err)
	}
}

func removeLineFromFile(line, file string) error { //nolint:unused
	out, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	re := regexp.MustCompile(line)
	out = re.ReplaceAll(out, []byte(""))

	return os.WriteFile(file, out, 0644)
}
