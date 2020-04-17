package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"
)

var (
	cliName = "confluent"
	releaseVersion = "v0.0.0"
	prevVersion = "v0.0.0"
)


func main() {

	fileName := path.Join(".", "release-notes", cliName, "index.rst")

	err := writeReleaseNotes(fileName, cliName, releaseVersion, prevVersion)
	if err != nil {
		panic(err)
	}
}

func writeReleaseNotes(filename string, cliName string, releaseVersion string, prevVersion string) error {
	content := `

%s CLI %s Release Notes
==============================

Merged PR List <REMOVE WHEN DONE>
%s

New Features
- new feature


Bug Fixes
- bug
`
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	fmt.Println("PREV:", prevVersion)
	mergedPRs := getMergedPRs(prevVersion)
	fmt.Println(mergedPRs)
	_, err = io.WriteString(f, fmt.Sprintf(content, strings.ToUpper(cliName), releaseVersion, mergedPRs))
	return err
}


func getMergedPRs(prevVersion string) string {
	//cmd := exec.Command("git", "log", version+"..HEAD")
	//var out bytes.Buffer
	//grep := exec.Command("grep", "-e", "\"(#[0-9]*)\"")
	//r, w := io.Pipe()
	//cmd.Stdout = w
	//grep.Stdin = r

	cmd := fmt.Sprintf("git log %s..HEAD | grep -e \"(#[0-9]*)\"", prevVersion)
	out, _ := exec.Command("bash", "-c", cmd).Output()
	return string(out)
}

