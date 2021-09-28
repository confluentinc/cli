package release_notes

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
)

const (
	breakingChangesTitle = "Breaking Changes"
	newFeaturesTitle     = "New Features"
	bugFixesTitle        = "Bug Fixes"

	prepFileName = "./release-notes/prep"
	placeHolder  = "<PLACEHOLDER>"
)

func WriteReleaseNotesPrep(filename, releaseVersion, prevVersion string) error {
	prepBaseFile := path.Join(".", "internal", "pkg", "release-notes", "prep-base")
	prepBaseBytes, err := ioutil.ReadFile(prepBaseFile)
	if err != nil {
		return fmt.Errorf("unable to load release prep-base")
	}
	prepBaseString := string(prepBaseBytes)

	if err := os.MkdirAll(filepath.Dir(filename), os.ModePerm); err != nil {
		return err
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	mergedPRs := getMergedPRs(prevVersion)
	prepFile := fmt.Sprintf(prepBaseString, releaseVersion, prevVersion, mergedPRs, breakingChangesTitle, newFeaturesTitle, bugFixesTitle)
	_, err = io.WriteString(file, prepFile)
	return err
}

func getMergedPRs(prevVersion string) string {
	cmd := fmt.Sprintf("git log %s..HEAD | grep -e \"(#[0-9]*)\"", prevVersion)
	out, _ := exec.Command("bash", "-c", cmd).Output()
	return string(out)
}
