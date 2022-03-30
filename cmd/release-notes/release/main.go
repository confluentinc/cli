package main

import (
	rn "github.com/confluentinc/cli/internal/pkg/release-notes"
)

var (
	releaseVersion   = "v0.0.0"
	releaseNotesPath = ""
)

func main() {
	err := rn.WriteReleaseNotes(releaseNotesPath, releaseVersion)
	if err != nil {
		panic(err)
	}
}
