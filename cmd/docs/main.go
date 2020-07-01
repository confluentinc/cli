package main

import (
	"bytes"
	"fmt"
	"path"
	"strings"

	"github.com/confluentinc/cli/internal/cmd"
	"github.com/confluentinc/cli/internal/pkg/auth"
	"github.com/confluentinc/cli/internal/pkg/doc"
	"github.com/confluentinc/cli/internal/pkg/version"
)

var (
	// Injected from linker flags like `go build -ldflags "-X main.cliName=$NAME"`
	cliName = "confluent"
)

// See https://github.com/spf13/cobra/blob/master/doc/rest_docs.md
func main() {
	emptyStr := func(filename string) string { return "" }
	sphinxRef := func(name, ref string) string { return fmt.Sprintf(":ref:`%s`", ref) }
	confluent, err := cmd.NewConfluentCommand(
		cliName,
		true,
		&version.Version{},
		auth.NewNetrcHandler(""))
	if err != nil {
		panic(err)
	}
	err = doc.GenReSTTreeCustom(confluent.Command, path.Join(".", "docs", cliName), emptyStr, sphinxRef)
	if err != nil {
		panic(err)
	}

	indexHeader := func(filename string) string {
		buf := new(bytes.Buffer)

		buf.WriteString(fmt.Sprintf(".. _%s-ref:\n\n", cliName))
		title := fmt.Sprintf("|%s| CLI Command Reference\n", cliName)
		buf.WriteString(title)
		buf.WriteString(strings.Repeat("=", len(title)) + "\n\n")
		buf.WriteString(fmt.Sprintf("The available |%s| CLI commands are documented here.\n\n", cliName))

		return buf.String()
	}

	if err := doc.GenReSTIndex(confluent.Command, path.Join(".", "docs", cliName, "index.rst"), indexHeader, sphinxRef); err != nil {
		panic(err)
	}
}
