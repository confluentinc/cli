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


func WriteReleaseNotes(releaseNotesDocsPagePath, cliName, releaseVersion string) error {
	releaseNotes, err := getReleaseNotes(cliName, releaseVersion)
	if err != nil {
		return err
	}

	err = writeLatestReleaseNotesFile(cliName, releaseNotes)
	if err != nil {
		return err
	}

	pastReleaseNotes, err := getPastReleaseNotesDocsPage(releaseNotesDocsPagePath)
	if err != nil {
		return err
	}
	var releaseNotesDocsPage string
	if cliName == "ccloud" {
		releaseNotesDocsPage = constructReleaseNotesDocsPage(ccloudHeader, releaseNotes, pastReleaseNotes)
	} else {
		releaseNotesDocsPage = constructReleaseNotesDocsPage(confluentHeader, releaseNotes, pastReleaseNotes)
	}
	return writeLocalReleaseNotesDocsPage(cliName, releaseNotesDocsPage)
}

func getReleaseNotes(cliName, releaseVersion string) (string, error) {
	releaseNotesContent, err := getReleaseNotesContent(cliName)
	if err != nil {
		return "", err
	}
	var format string
	if cliName == "ccloud" {
		format = ccloudReleaseNotesFormat
	} else {
		format = confluentReleaseNotesFormat
	}
	releaseNotes := fmt.Sprintf(format, releaseVersion, releaseNotesContent)
	return releaseNotes, nil
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
		sectionString += newFeaturesString + "\n"
	}
	if bugFixesString != "" {
		sectionString += bugFixesString
	}
	if sectionString == "" {
		sectionString = fmt.Sprintf(noChangeContentFormat, strings.ToUpper(cliName))
	}
	return sectionString, nil
}

func writeLatestReleaseNotesFile(cliName, releaseNotes string) error {
	latestReleaseNotesFilePath := fmt.Sprintf(releaseNotesLocalFileFormat, cliName, latestReleaseNotesFileName)
	return writeReleaseNotesToLocalFile(latestReleaseNotesFilePath, releaseNotes)
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

func writeLocalReleaseNotesDocsPage(cliName, releaseNotesDocsPage string) error {
	filePath := fmt.Sprintf(releaseNotesLocalFileFormat, cliName, releaseNotesDocsPageFileName)
	return writeReleaseNotesToLocalFile(filePath, releaseNotesDocsPage)
}

func getPastReleaseNotesDocsPage(filePath string) (string, error) {
	oldReleaseFileName := path.Join(filePath, releaseNotesDocsPageFileName)
	b, err := ioutil.ReadFile(oldReleaseFileName)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func constructReleaseNotesDocsPage(header, latestReleaseNotes, pastReleaseNotes string) string {
	// remove old header
	pastReleaseNotes = strings.ReplaceAll(pastReleaseNotes, header, "")
	return header + latestReleaseNotes + pastReleaseNotes
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
		return sectionTitle + "\n------------------------\n" + sectionString
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
