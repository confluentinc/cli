package main

import (
	"github.com/confluentinc/cli/internal/pkg/releasenotes"
)

// releaseNotesPath is populated with ldflags
var releaseNotesPath string

func main() {
	r := releasenotes.New()

	if err := r.ReadFromGithub(); err != nil {
		panic(err)
	}

	if err := r.Write(releaseNotesPath); err != nil {
		panic(err)
	}
}
