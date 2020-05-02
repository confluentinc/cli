package release_notes

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
)


func WriteReleaseNotes(releaseNotesPath, cliName, releaseVersion string) error {
	releaseNotesContent, err := getReleaseNotesContent(cliName)
	if err != nil {
		return err
	}

	latestReleaseNotes := fmt.Sprintf(releaseNotesContentFormat, strings.ToUpper(cliName), releaseVersion, releaseNotesContent)

	latestReleaseNotesFilePath := fmt.Sprintf(releaseNotesLocalFile, cliName, latestReleaseNotesFileName)
	err = writeReleaseNotesToLocalFile(latestReleaseNotesFilePath, latestReleaseNotes)
	if err != nil {
		return err
	}

	pastReleaseNotes, err := getPastReleaseNotes(releaseNotesPath)
	if err != nil {
		return err
	}
	fullReleaseNotes  := latestReleaseNotes + pastReleaseNotes

	fullReleaseNotesFilePath := fmt.Sprintf(releaseNotesLocalFile, cliName, fullReleaseNotesFileName)
	err = writeReleaseNotesToLocalFile(fullReleaseNotesFilePath, fullReleaseNotes)
	if err != nil {
		return err
	}
	return nil
}

func writeReleaseNotesToLocalFile(fileName, releaseNotes string) error {
	f, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.WriteString(f, releaseNotes)
	if err != nil {
		return err
	}
	return nil
}

func getPastReleaseNotes(releaseNotesPath string) (string, error) {
	oldReleaseFileName := path.Join(releaseNotesPath, "index.rst")
	b, err := ioutil.ReadFile(oldReleaseFileName)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func getReleaseNotesContent(cliName string) (string, error) {
	sectionContentMap, err := getSectionContentMap()
	if err != nil {
		return "", err
	}

	newFeaturesList, bugFixesList := getNewFeatureAndBugFixesList(sectionContentMap, cliName)

	newFeaturesString := assembleSectionString(newFeaturesSectionTitle, newFeaturesList)
	bugFixesString := assembleSectionString(bugFixesSectionTitle, bugFixesList)

	var sectionString string
	if newFeaturesString != "" {
		sectionString += newFeaturesString + "\n\n"
	}
	if bugFixesString != "" {
		sectionString += bugFixesString
	}
	if sectionString == "" {
		sectionString = fmt.Sprintf(noChangeContentFormat, strings.ToUpper(cliName))
	}
	return sectionString, nil
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

func getNewFeatureAndBugFixesList(sectionContentMap map[ReleaseNotesSection][]string, cliName string) ([]string, []string) {
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
	return newFeatures, bugFixes
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
