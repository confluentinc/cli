package main

import (
	"path"

	rn "github.com/confluentinc/cli/internal/pkg/release-notes"
)

var (
	releaseVersion = "v0.0.0"
	prevVersion    = "v0.0.0"
)

func main() {
	fileName := path.Join(".", "release-notes", "prep")
	err := rn.WriteReleaseNotesPrep(fileName, releaseVersion, prevVersion)
	if err != nil {
		panic(err)
	}
}
