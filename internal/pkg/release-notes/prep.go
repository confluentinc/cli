package release_notes

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
)

const (
	bothBreakingChangesTitle      = "Breaking Changes for Both CLIs"
	bothNewFeaturesTitle          = "New Features for Both CLIs"
	bothBugFixesTitle             = "Bug Fixes for Both CLIs"
	ccloudBreakingChangesTitle    = "CCloud Breaking Changes"
	ccloudNewFeaturesTitle        = "CCloud New Features"
	ccloudBugFixesTitle           = "CCloud Bug Fixes"
	confluentBreakingChangesTitle = "Confluent Breaking Changes"
	confluentNewFeaturesTitle     = "Confluent New Features"
	confluentBugFixesTitle        = "Confluent Bug Fixes"

	prepFileName = "./release-notes/prep"
	placeHolder  = "<PLACEHOLDER>"
)

func WriteReleaseNotesPrep(filename string, releaseVersion string, prevVersion string) error {
	prepBaseFile := path.Join(".", "internal", "pkg", "release-notes", "prep-base")
	prepBaseBytes, err := ioutil.ReadFile(prepBaseFile)
	if err != nil {
		return fmt.Errorf("unable to load release prep-base")
	}
	prepBaseString := string(prepBaseBytes)
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	mergedPRs := getMergedPRs(prevVersion)
	prepFile := fmt.Sprintf(prepBaseString, releaseVersion, prevVersion, mergedPRs,
		bothBreakingChangesTitle,
		bothNewFeaturesTitle,
		bothBugFixesTitle,
		ccloudBreakingChangesTitle,
		ccloudNewFeaturesTitle,
		ccloudBugFixesTitle,
		confluentBreakingChangesTitle,
		confluentNewFeaturesTitle,
		confluentBugFixesTitle,
	)
	_, err = io.WriteString(f, prepFile)
	return err
}

func getMergedPRs(prevVersion string) string {
	cmd := fmt.Sprintf("git log %s..HEAD | grep -e \"(#[0-9]*)\"", prevVersion)
	out, _ := exec.Command("bash", "-c", cmd).Output()
	return string(out)
}
