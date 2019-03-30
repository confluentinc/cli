// +build docs

// Above is a hack to avoid having to upgrade the CLI to go 1.12 in CI for a docs dependency.
// Without this, `make test` will include this file in $(go list ./...) and fail the tests.
package main

import (
	"fmt"

	"github.com/spf13/cobra/doc"

	"github.com/confluentinc/cli/internal/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/version"
)

// See https://github.com/spf13/cobra/blob/master/doc/rest_docs.md
func main() {
	emptyStr := func(filename string) string { return "" }
	sphinxRef := func(name, ref string) string { return fmt.Sprintf(":ref:`%s`", ref) }
	confluent := cmd.NewConfluentCommand(&config.Config{}, &version.Version{}, log.New())
	err := doc.GenReSTTreeCustom(confluent, "./docs", emptyStr, sphinxRef)
	if err != nil {
		panic(err)
	}
}
