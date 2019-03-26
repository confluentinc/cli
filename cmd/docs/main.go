// +build docs

// Above is a hack to avoid having to upgrade the CLI to go 1.12 in CI for a docs dependency.
// Without this, `make test` will include this file in $(go list ./...) and fail the tests.
package main

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra/doc"

	"github.com/confluentinc/cli/internal/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/version"
)

const fmTemplate = `---
date: %s
title: "%s"
slug: %s
url: %s
---
`

func filePrepender(filename string) string {
	now := time.Now().Format(time.RFC3339)
	name := filepath.Base(filename)
	base := strings.TrimSuffix(name, path.Ext(name))
	url := "/commands/" + strings.ToLower(base) + "/"
	return fmt.Sprintf(fmTemplate, now, strings.Replace(base, "_", " ", -1), base, url)
}

// Sphinx cross-referencing format
func linkHandler(name, ref string) string {
	return fmt.Sprintf(":ref:`%s <%s>`", name, ref)
}

func main() {
	confluent := cmd.NewConfluentCommand(&config.Config{}, &version.Version{}, log.New())
	// See https://github.com/spf13/cobra/blob/master/doc/rest_docs.md
	err := doc.GenReSTTreeCustom(confluent, "./docs", filePrepender, linkHandler)
	if err != nil {
		panic(err)
	}
}
