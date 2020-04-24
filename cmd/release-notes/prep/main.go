package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
)

var (
	releaseVersion = "v0.0.0"
	prevVersion = "v0.0.0"
)


func main() {

	fileName := path.Join(".", "release-notes", "prep")

	err := writeReleaseNotesPrep(fileName, releaseVersion, prevVersion)
	if err != nil {
		panic(err)
	}
}

func writeReleaseNotesPrep(filename string, releaseVersion string, prevVersion string) error {
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
	_, err = io.WriteString(f, fmt.Sprintf(prepBaseString, releaseVersion, mergedPRs))
	return err
}


func getMergedPRs(prevVersion string) string {
	cmd := fmt.Sprintf("git log %s..HEAD | grep -e \"(#[0-9]*)\"", prevVersion)
	out, _ := exec.Command("bash", "-c", cmd).Output()
	return string(out)
}
