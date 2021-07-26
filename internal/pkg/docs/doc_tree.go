package docs

import (
	"os"
	"path/filepath"
)

// GenerateDocTree recursively builds a nested hierarchy of folders and files for a CLI's documentation.
// An index page is created for any command with subcommands, which links to its children's documentation pages.
func GenerateDocTree(tabs []Tab, dir string, depth int) error {
	if tabs[0].Command.HasSubCommands() {
		// This command has subcommands. Create a new directory and add an index page.
		// We assume that if one tab's command has subcommands, they all have subcommands.
		name := tabs[0].Command.Name()
		dir = filepath.Join(dir, name)

		if err := os.Mkdir(dir, os.ModePerm); err != nil {
			return err
		}

		if err := generateIndexPage(tabs, dir); err != nil {
			return err
		}

		// Recursively generate documentation for subcommands.
		tabsByName := make(map[string][]Tab)
		for _, tab := range tabs {
			for _, subcommand := range tab.Command.Commands() {
				if subcommand.IsAvailableCommand() {
					name := subcommand.Name()
					tabsByName[name] = append(tabsByName[name], Tab{Name: tab.Name, Command: subcommand})
				}
			}
		}

		for _, tabs := range tabsByName {
			if err := GenerateDocTree(tabs, dir, depth+1); err != nil {
				return err
			}
		}

		return nil
	} else {
		// The command has no subcommands. Generate its documentation page.
		return generateDocPage(tabs, dir, depth)
	}
}
