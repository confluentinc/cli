package main

import (
	"os"
	"path/filepath"
	"regexp"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"

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
		{CurrentContext: "Cloud", Contexts: map[string]*v1.Context{"Cloud": {PlatformName: "https://confluent.cloud", State: &v1.ContextState{Auth: &v1.AuthConfig{Organization: &orgv1.Organization{}}}}}},
		{CurrentContext: "On-Prem", Contexts: map[string]*v1.Context{"On-Prem": {PlatformName: "https://example.com"}}},
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

	removeDocumentation()

	if err := os.Setenv("HOME", currentHOME); err != nil {
		panic(err)
	}
}

// removeDocumentation hides documentation for unreleased features
func removeDocumentation() {
	out, err := os.ReadFile(filepath.Join("docs", "index.rst"))
	if err != nil {
		panic(err)
	}

	re := regexp.MustCompile(`\s{3}stream-share/index\n`)
	out = re.ReplaceAll(out, []byte(""))

	if err := os.WriteFile(filepath.Join("docs", "index.rst"), out, 0644); err != nil {
		panic(err)
	}

	if err := os.RemoveAll(filepath.Join("docs", "stream-share")); err != nil {
		panic(err)
	}
}
