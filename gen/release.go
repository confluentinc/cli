// +build ignore

package main

// This program binds cli features to the goreleaser configuration.

import (
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

var (
	interfaceTemplate *template.Template
)

var Formatters = template.FuncMap{
	"ToTitle": strings.Title,
	"ToLower": strings.ToLower,
	"ToUpper": strings.ToUpper,
}

func main() {
	release, err := config.Load(".goreleaser.yml")
	if err != nil {
		log.Fatalf("Failed to load goreleaser configuration", err)
	}

	writeShared(release)
	writeMain(release)
}

func writeMain(release config.Project) {
	out := openFile("main","main.go")
	defer out.Close()
	writeTemplate(out, interfaceTemplate, "main", extractModules(release.Builds))
}

func writeShared(release config.Project) {
	for _, module := range extractModules(release.Builds) {
		out := openFile("shared_interface", "shared", module.Package, "interface.go")
		defer out.Close()
		err := interfaceTemplate.ExecuteTemplate(out, "plugin_interface", module)
		if err != nil {
			log.Fatalf("Failed to write modules.go", err)
		}
	}

}
func openFile(name string, outputPath ...string) *os.File {
	fileHandle, err := os.Create(path.Join(outputPath...))
	if err != nil {
		log.Fatalf("Failed to generate %s", name, err)
	}
	return fileHandle
}

func writeTemplate(out *os.File, tmpl *template.Template, name string, data interface{}) {
	err := tmpl.ExecuteTemplate(out, name, data)
	if err != nil {
		log.Fatalf("Failed to write modules.go", err)
	}
}

func extractModules(builds []config.Build) []Module  {
	var modules []Module
	for _, build := range builds {
		if ok, _ := path.Match("*/plugin/*", build.Main); !ok {
			continue
		}
		modules = append(modules, Module{Name: build.Binary, Package: extractPackage(build.Binary)})
	}
	return modules
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
			path.Join("gen", "templates", "generated.txt"),
			path.Join("gen", "templates", "shared", "interface.txt"),
			path.Join("gen", "templates", "main.txt")))
}
