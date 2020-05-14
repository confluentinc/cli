package release_notes

import (
	"fmt"
	"io"
	"os"
)

const (
	releaseNotesLocalFilePathFormat = "./release-notes/%s/%s"
	latestReleaseNotesFileName      = "latest-release.rst"
	docsPageFileName                = "release-notes.rst"
)

type ReleaseNotesContent struct {
	newFeatures []string
	bugFixes    []string
}

func WriteReleaseNotes(ccloudReleaseNotesPath, confluentReleaseNotesPath, releaseVersion string) error {
	prepFileReader := NewPrepFileReader(prepFileName)
	sections, err := prepFileReader.getSectionsMap()
	if err != nil {
		return err
	}
	err = constructAndWriteCLIReleaseNotes(ccloudReleaseNotesPath, "ccloud", releaseVersion, sections)
	if err != nil {
		return err
	}
	err = constructAndWriteCLIReleaseNotes(confluentReleaseNotesPath, "confluent", releaseVersion, sections)
	if err != nil {
		return err
	}
	return nil
}

func constructAndWriteCLIReleaseNotes(docsPath string, cliName string, releaseVersion string, sectionsMap map[SectionType][]string) error {
	content := extractReleaseNotesContent(sectionsMap, cliName)
	err := constructAndWriteLatestReleaseNotesForS3(content, cliName, releaseVersion)
	if err != nil{
		return err
	}
	err = constructAndWriteUpdatedDocsPage(docsPath, content, cliName, releaseVersion)
	if err != nil {
		return err
	}
	return nil
}

func extractReleaseNotesContent(sectionsMap map[SectionType][]string, cliName string) ReleaseNotesContent {
	if cliName == "ccloud" {
		return ReleaseNotesContent{
			newFeatures: getSectionContentList(sectionsMap, ccloudNewFeatures, bothNewFeatures),
			bugFixes:    getSectionContentList(sectionsMap, ccloudBugFixes, bothBugFixes),
		}
	} else {
		return ReleaseNotesContent{
			newFeatures: getSectionContentList(sectionsMap, confluentNewFeatures, bothNewFeatures),
			bugFixes:    getSectionContentList(sectionsMap, confluentBugFixes, bothBugFixes),
		}
	}
}

func getSectionContentList(sectionsMap map[SectionType][]string, exclusiveSection, bothSection SectionType, ) []string {
	var contentList []string
	contentList = append(contentList, sectionsMap[exclusiveSection]...)
	contentList = append(contentList, sectionsMap[bothSection]...)
	return contentList
}

func constructAndWriteLatestReleaseNotesForS3(content ReleaseNotesContent, cliName string, version string) error {
	releaseNotesBuilder := NewReleaseNotesBuilder(s3ReleaseNotesTitleFormat, s3SectionHeaderFormat, cliName, version)
	releaseNotes := releaseNotesBuilder.buildReleaseNotes(content)
	destFile := fmt.Sprintf(releaseNotesLocalFilePathFormat, cliName, latestReleaseNotesFileName)
	return writeFile(destFile, releaseNotes)
}

func constructAndWriteUpdatedDocsPage(docsFilePath string, content ReleaseNotesContent, cliName string, version string) error {
	releaseNotesBuilder := NewReleaseNotesBuilder(docsReleaseNotesTitleFormat, docsSectionHeaderFormat, cliName, version)
	releaseNotes := releaseNotesBuilder.buildReleaseNotes(content)
	docsUpdateHandler := NewDocsUpdateHandler(cliName, docsFilePath + "/" + docsPageFileName)
	updatedDocsPage, err := docsUpdateHandler.getUpdatedDocsPage(releaseNotes)
	if err != nil {
		return err
	}
	destFile := fmt.Sprintf(releaseNotesLocalFilePathFormat, cliName, docsPageFileName)
	return writeFile(destFile, updatedDocsPage)
}

func writeFile(filePath, fileContent string) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.WriteString(f, fileContent)
	if err != nil {
		return err
	}
	return nil
}
