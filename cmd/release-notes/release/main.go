package main

import (
	rn "github.com/confluentinc/cli/internal/pkg/release-notes"
)

var (
	releaseVersion            = "v0.0.0"
	ccloudReleaseNotesPath    = ""
	confluentReleaseNotesPath = ""
)

func main() {
	err := rn.WriteReleaseNotes(ccloudReleaseNotesPath, confluentReleaseNotesPath, releaseVersion)
	if err != nil {
		panic(err)
	}
}
