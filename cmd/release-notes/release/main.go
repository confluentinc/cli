package main

import (
	rn "github.com/confluentinc/cli/internal/pkg/release-notes"
)

var (
	releaseVersion = "v0.0.0"
	ccloudReleaseNotesPath = ""
	confluentReleaseNotesPath = ""
)


func main() {
	err := rn.WriteReleaseNotes(ccloudReleaseNotesPath, "ccloud", releaseVersion)
	if err != nil {
		panic(err)
	}
	err = rn.WriteReleaseNotes(confluentReleaseNotesPath,"confluent", releaseVersion)
	if err != nil {
		panic(err)
	}
}
