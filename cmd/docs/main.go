package main

import (
	"os"
	"testing"

	"github.com/confluentinc/cli/v4/internal"
	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/docs"
	pversion "github.com/confluentinc/cli/v4/pkg/version"
	testserver "github.com/confluentinc/cli/v4/test/test-server"
)

// Auto-generate documentation files for all CLI commands. Documentation uses reStructured Text (ReST) format, and is
// nested in a hierarchical format. Commands may differ based on the login context found in the user's config file.
// We use tabs to show differing cloud and on-prem commands on the same page.
// This code is adapted from https://github.com/spf13/cobra/blob/master/doc/rest_docs.md

func main() {
	// Set up test server for feature flags called by the code
	testBackend := testserver.StartTestCloudServer(&testing.T{}, true)
	defer testBackend.Close()

	// Prevent printing the user's $HOME in docs when generating confluent local services kafka
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

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

	if err := os.Setenv("HOME", home); err != nil {
		panic(err)
	}
}
