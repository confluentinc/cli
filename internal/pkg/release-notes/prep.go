package release_notes

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
)

func WriteReleaseNotesPrep(filename string, releaseVersion string, prevVersion string) error {
	prepBaseFile := path.Join(".", "internal", "pkg", "release-notes", "prep-base")
	prepBaseBytes, err := ioutil.ReadFile(prepBaseFile)
	if err != nil {
		return fmt.Errorf("Unable to load release prep-base.")
	}
	prepBaseString := string(prepBaseBytes)
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	mergedPRs := getMergedPRs(prevVersion)
	_, err = io.WriteString(f, fmt.Sprintf(prepBaseString, releaseVersion, prevVersion, mergedPRs))
	return err
}


func getMergedPRs(prevVersion string) string {
	cmd := fmt.Sprintf("git log %s..HEAD | grep -e \"(#[0-9]*)\"", prevVersion)
	out, _ := exec.Command("bash", "-c", cmd).Output()
	return string(out)
}
