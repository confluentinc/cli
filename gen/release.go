// +build ignore

package main

// This program binds cli features to the goreleaser configuration.

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/goreleaser/goreleaser/pkg/config"
)

type Module struct {
	Name string
	Package string
}

var interfaceTemplate *template.Template

var Formatters = template.FuncMap{
	"ToTitle": strings.Title,
	"ToLower": strings.ToLower,
}

func main() {
	release, err := config.Load(".goreleaser.yml")
	if err != nil {
		log.Fatalf("Failed to load goreleaser configuration", err)
	}

	module := new(Module)
	for _, build := range release.Builds {
		module.Name = ""
		module.Package = ""

		if !strings.Contains(build.Main, "plugin/") {
			continue
		}

		module.Name = build.Binary
		module.Package = extractPackage(build.Binary)

		fileHandle, err := os.Create(fmt.Sprintf(path.Join("shared", module.Package, "interface.go")))
		if err != nil {
			fileHandle.Close()
			log.Fatalf("Failed to generate %s", module.Package, err)
		}

		err = interfaceTemplate.ExecuteTemplate(fileHandle, "src", module)
		if err != nil {
			fileHandle.Close()
			log.Fatalf("Failed to write modules.go", err)
		}
		fileHandle.Close()

	}
}

// Extract package name based on naming convention outlined in README.md
func extractPackage(binary string) (string) {
	start := strings.Index(binary, "-")
	end := strings.LastIndex(binary, "-")

	// Name should be [platform]-[name]-plugin
	if (start <= 1 || end <= start) {
		return ""
	}

	return binary[start + 1:end]
}

func init() {
	interfaceTemplate = template.Must(
		template.New("interface").Funcs(Formatters).ParseFiles(
			path.Join("gen","templates", "shared", "interface.txt")))
}
