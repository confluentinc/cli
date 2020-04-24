package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

type ReleaseNotesSection int

const (
	bothNewFeature ReleaseNotesSection = iota
	bothBugFix
	ccloudNewFeature
	ccloudBugFix
	confluentNewFeature
	confluentBugFix
)

var (
	releaseVersion = "v0.0.0"
	ccloudReleaseNotesPath = ""
	confluentReleaseNotesPath = ""

	//sectionNameMap = map[ReleaseNotesSection]string{
	//	bothNewFeature: "New Features for Both CLIs",
	//	bothBugFix: "Bug Fixes for Both CLIs",
	//	ccloudNewFeature: "CCloud New Features",
	//	ccloudBugFix: "CCloud Bug Fixes",
	//	confluentNewFeature: "Confluent New Features",
	//	confluentBugFix: "Confluent Bug Fixes",
	//}
	sectionNameMap = map[string]ReleaseNotesSection{
		"New Features for Both CLIs": bothNewFeature,
		"Bug Fixes for Both CLIs": bothBugFix,
		"CCloud New Features": ccloudNewFeature,
		"CCloud Bug Fixes": ccloudBugFix,
		"Confluent New Features": confluentNewFeature,
		"Confluent Bug Fixes": confluentBugFix,
	}
	prepFileName = path.Join(".", "release-notes", "prep")
	placeHolder = "<PLACEHOLDER: REPLACE WITH CONTENT OR DO NOTHING IF NO CONTENT>"

	newFeaturesSectionTitle = "New Features"
	bugFixesSectionTitle = "Bug Fixes"
)


func main() {

	err := writeReleaseNotes("ccloud", releaseVersion)
	if err != nil {
		panic(err)
	}
	err = writeReleaseNotes("confluent", releaseVersion)
	if err != nil {
		panic(err)
	}
}


func writeReleaseNotes(cliName string, releaseVersion string) error {
	content := `
=============================================================
%s CLI %s Release Notes
=============================================================

%s
`
	fileName := path.Join(".", "release-notes", cliName, "index.rst")
	f, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer f.Close()

	updateFileName := path.Join(".", "internal", "pkg", "update", "update-release-notes")
	updateFile, err := os.Create(updateFileName)
	defer updateFile.Close()

	sectionContentMap, err := getSectionContentMap()

	sections := getSectionsString(cliName, sectionContentMap)

	releaseNotes := fmt.Sprintf(content, strings.ToUpper(cliName), releaseVersion, sections)

	_, err = io.WriteString(updateFile, releaseNotes)
	if err != nil {
		return err
	}

	var oldReleaseFileName string
	if cliName == "ccloud" {
		oldReleaseFileName = path.Join(ccloudReleaseNotesPath, "index.rst")
	} else {
		oldReleaseFileName = path.Join(confluentReleaseNotesPath, "index.rst")
	}
	b, err := ioutil.ReadFile(oldReleaseFileName)
	if err != nil {
		return err
	}

	oldReleaseNotes := string(b)

	releaseNotes += oldReleaseNotes

	_, err = io.WriteString(f, releaseNotes)
	if err != nil {
		return err
	}

	return nil
}

func getSectionsString(cliName string, sectionContentMap map[ReleaseNotesSection][]string) string {
	var newFeatures []string
	var bugFixes []string
	if cliName == "ccloud" {
		newFeatures = append(newFeatures, sectionContentMap[ccloudNewFeature]...)
		bugFixes = append(bugFixes, sectionContentMap[ccloudBugFix]...)
	} else {
		newFeatures = append(newFeatures, sectionContentMap[confluentNewFeature]...)
		bugFixes = append(bugFixes, sectionContentMap[confluentBugFix]...)
	}
	newFeatures = append(newFeatures, sectionContentMap[bothNewFeature]...)
	bugFixes = append(bugFixes, sectionContentMap[bothBugFix]...)

	newFeaturesString := assembleSectionString(newFeaturesSectionTitle, newFeatures)
	bugFixesString := assembleSectionString(bugFixesSectionTitle, bugFixes)

	var sectionString string
	if newFeaturesString != "" {
		sectionString += newFeaturesString + "\n\n"
	}
	if bugFixesString != "" {
		sectionString += bugFixesString
	}
	return sectionString
}

func assembleSectionString(sectionTitle string, sectionList []string) string {
	var sectionString string
	for _, element := range sectionList {
		sectionString += "- " + element + "\n"
	}
	if sectionString != "" {
		return sectionTitle + "\n----------------------------\n" + sectionString
	}
	return ""
}

func getSectionContentMap() (map[ReleaseNotesSection][]string, error) {
	f, err := os.Open(prepFileName)
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(f)
	sections := make(map[ReleaseNotesSection][]string)

	for scanner.Scan() {
		line := scanner.Text()
		lastLine := processLine(line, sections, scanner)
		_ = processLine(lastLine, sections, scanner)
	}
	err = scanner.Err()
	if err != nil {
		return nil, err
	}
	return sections, nil
}

func processLine(line string, sections map[ReleaseNotesSection][]string, scanner *bufio.Scanner) string {
	section, ok := sectionNameMap[line]
	if ok {
		sectionList, lastLine := getSectionList(scanner)
		sections[section] = sectionList
		return lastLine
	}
	return ""
}

// Returns list of all section elements, and the line after the section in case there is no new line between sections
func getSectionList(scanner *bufio.Scanner) ([]string, string) {
	var sectionList []string
	var line string
	for scanner.Scan() {
		line = scanner.Text()
		if !strings.HasPrefix(line, "-") {
			break
		}
		elementString := line[1:]
		elementString = strings.TrimSpace(elementString)
		if isPlaceHolder(elementString) {
			break
		}
		sectionList = append(sectionList, elementString)
	}
	return sectionList, line
}

func isPlaceHolder(element string) bool {
	return element == placeHolder ||
		(strings.HasPrefix(element, "<") && strings.HasSuffix(element, ">"))
}
