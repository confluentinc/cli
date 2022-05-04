package main

import (
	"os"

	"github.com/confluentinc/cli/internal/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/docs"
	"github.com/confluentinc/cli/internal/pkg/version"
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

	// Auto-generate documentation for cloud and on-prem commands.
	configs := []*v1.Config{
		{
			Contexts:       map[string]*v1.Context{"Cloud": {PlatformName: v1.CCloudHostnames[0]}},
			CurrentContext: "Cloud",
		},
		{
			Contexts:       map[string]*v1.Context{"On-Prem": {PlatformName: "https://example.com"}},
			CurrentContext: "On-Prem",
		},
	}

	tabs := make([]docs.Tab, len(configs))
	for i, cfg := range configs {
		tabs[i] = docs.Tab{
			Name:    cfg.CurrentContext,
			Command: cmd.NewConfluentCommand(cfg, false, new(version.Version)),
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
