package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type releaseNotes struct {
	date     time.Time
	version  string
	sections map[string][]string
}

var sections = []string{
	"Breaking Changes",
	"New Features",
	"Bug Fixes",
}

// Usage: go run main.go 3.10.0.json docs
func main() {
	path, format := os.Args[1], os.Args[2]

	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	version := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))

	r := &releaseNotes{
		date:     time.Now(),
		version:  version,
		sections: make(map[string][]string),
	}

	if err := json.NewDecoder(file).Decode(&r.sections); err != nil {
		panic(err)
	}

	switch format {
	case "docs":
		fmt.Print(docs(r))
	case "github":
		fmt.Print(github(r))
	case "s3":
		fmt.Print(s3(r))
	}
}

func docs(r *releaseNotes) string {
	title := fmt.Sprintf("[%s] Confluent CLI v%s Release Notes", r.date.Format("1/2/2006"), r.version)

	out := underline(title, "=") + "\n"
	out += "\n"

	for _, section := range sections {
		if len(r.sections[section]) == 0 {
			continue
		}

		out += underline(section, "-") + "\n"
		for _, line := range r.sections[section] {
			line = strings.ReplaceAll(line, "`", "``")
			out += bullet(line) + "\n"
		}
		out += "\n"
	}

	return out
}

func github(r *releaseNotes) string {
	out := ""

	for _, section := range sections {
		if len(r.sections[section]) == 0 {
			continue
		}

		out += underline(section, "-") + "\n"
		for _, line := range r.sections[section] {
			out += bullet(line) + "\n"
		}
		out += "\n"
	}

	return out
}

func s3(r *releaseNotes) string {
	title := fmt.Sprintf("[%s] Confluent CLI v%s Release Notes", r.date.Format("1/2/2006"), r.version)

	out := underline(title, "=") + "\n"
	out += "\n"

	for _, section := range sections {
		if len(r.sections[section]) == 0 {
			continue
		}

		out += underline(section, "-") + "\n"
		for _, line := range r.sections[section] {
			out += bullet(line) + "\n"
		}
		out += "\n"
	}

	return out
}

func underline(s string, style string) string {
	out := s + "\n"
	out += strings.Repeat(style, len(s))
	return out
}

func bullet(s string) string {
	return fmt.Sprintf("- %s", s)
}
