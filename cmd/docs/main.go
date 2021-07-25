package main

import (
	"os"

	"github.com/confluentinc/cli/internal/cmd"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/doc"
)

// Auto-generate documentation files for all CLI commands. Documentation uses reStructured Text (ReST) format, and is
// nested in a hierarchical format. Commands may differ based on the login context found in the user's config file. We
// use tabs to show differing cloud and on-prem commands on the same page.
// This code is adapted from https://github.com/spf13/cobra/blob/master/doc/rest_docs.md

func main() {
	// Prevent printing the user's HOME in docs when generating confluent local services kafka
	if err := os.Setenv("HOME", "$HOME"); err != nil {
		panic(err)
	}

	// Auto-generate documentation for cloud and on-prem commands.
	configs := []*v3.Config{
		{
			Contexts:       map[string]*v3.Context{"cloud": {PlatformName: v3.CCloudHostnames[0]}},
			CurrentContext: "cloud",
		},
		{
			Contexts:       map[string]*v3.Context{"onprem": {PlatformName: "https://example.com"}},
			CurrentContext: "onprem",
		},
	}

	commands := make([]doc.Tab, len(configs))
	for i, cfg := range configs {
		commands[i] = doc.Tab{
			Command: cmd.NewConfluentCommand(cfg, false, nil).Command,
			Name:    cfg.CurrentContext,
		}
	}

	if err := doc.GenerateDocTree(commands, "docs", 0); err != nil {
		panic(err)
	}
}
