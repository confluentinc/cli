package release_notes

import (
	"io"
	"os"
)

const (
	docsPageFileName = "release-notes.rst"

	s3CCloudReleaseNotesFilePath    = "./release-notes/ccloud/latest-release.rst"
	s3ConfluentReleaseNotesFilePath = "./release-notes/confluent/latest-release.rst"

	updatedCCloudDocsFilePath    = "./release-notes/ccloud/release-notes.rst"
	updatedConflunetDocsFilePath = "./release-notes/confluent/release-notes.rst"

	s3ReleaseNotesTitleFormat = `
===================================
%s %s Release Notes
===================================
`
	docsReleaseNotesTitleFormat = `
%s %s Release Notes
=====================================`

	s3SectionHeaderFormat   = "%s\n-------------"
	docsSectionHeaderFormat = "**%s**"

	s3CCloudCLIName   = "CCloud CLI"
	docsCCloudCLIName = "|ccloud| CLI"

	s3ConfluentCLIName   = "Confluent CLI"
	docsConfluentCLIName = "|confluent-cli|"
)

var (
	s3CCloudReleaseNotesBuilderParams = &ReleaseNotesBuilderParams{
		cliDisplayName:      s3CCloudCLIName,
		titleFormat:         s3ReleaseNotesTitleFormat,
		sectionHeaderFormat: s3SectionHeaderFormat,
	}
	s3ConfluentReleaseNotesBuilderParams = &ReleaseNotesBuilderParams{
		cliDisplayName:      s3ConfluentCLIName,
		titleFormat:         s3ReleaseNotesTitleFormat,
		sectionHeaderFormat: s3SectionHeaderFormat,
	}
	docsCCloudReleaseNotesBuilderParams = &ReleaseNotesBuilderParams{
		cliDisplayName:      docsCCloudCLIName,
		titleFormat:         docsReleaseNotesTitleFormat,
		sectionHeaderFormat: docsSectionHeaderFormat,
	}
	docsConfluentReleaseNotesBuilderParmas = &ReleaseNotesBuilderParams{
		cliDisplayName:      docsConfluentCLIName,
		titleFormat:         docsReleaseNotesTitleFormat,
		sectionHeaderFormat: docsSectionHeaderFormat,
	}
)

func WriteReleaseNotes(ccloudDocsPath, confluentDocsPath, releaseVersion string) error {
	ccloudReleaseNotesContent, confluentReleaseNotesContent, err := getCCloudAndConfluentReleaseNotesContent()
	if err != nil {
		return err
	}
	err = buildAndWriteCCloudReleaseNotes(releaseVersion, ccloudReleaseNotesContent, ccloudDocsPath)
	if err != nil {
		return err
	}
	err = buildAndWriteConfluentReleaseNotes(releaseVersion, confluentReleaseNotesContent, confluentDocsPath)
	if err != nil {
		return err
	}
	return nil
}

func getCCloudAndConfluentReleaseNotesContent() (*ReleaseNotesContent, *ReleaseNotesContent, error) {
	prepFileReader := NewPrepFileReader()
	err := prepFileReader.ReadPrepFile(prepFileName)
	if err != nil {
		return nil, nil, err
	}
	ccloudReleaseNotesContent, err := prepFileReader.GetCCloudReleaseNotesContent()
	if err != nil {
		return nil, nil, err
	}
	confluentReleaseNotesContent, err := prepFileReader.GetConfluentReleaseNotesContent()
	if err != nil {
		return nil, nil, err
	}
	return ccloudReleaseNotesContent, confluentReleaseNotesContent, nil
}

func buildAndWriteCCloudReleaseNotes(version string, content *ReleaseNotesContent, docsPath string) error {
	s3ReleaseNotes := buildReleaseNotes(version, s3CCloudReleaseNotesBuilderParams, content)
	err := writeFile(s3CCloudReleaseNotesFilePath, s3ReleaseNotes)
	if err != nil {
		return err
	}
	ccloudDocsReleaseNotes := buildReleaseNotes(version, docsCCloudReleaseNotesBuilderParams, content)
	ccloudDocsPage, err := buildDocsPage(docsPath, ccloudDocsPageHeader, ccloudDocsReleaseNotes)
	if err != nil {
		return err
	}
	err = writeFile(updatedCCloudDocsFilePath, ccloudDocsPage)
	if err != nil {
		return err
	}
	return nil
}

func buildAndWriteConfluentReleaseNotes(version string, content *ReleaseNotesContent, docsPath string) error {
	s3ReleaseNotes := buildReleaseNotes(version, s3ConfluentReleaseNotesBuilderParams, content)
	err := writeFile(s3ConfluentReleaseNotesFilePath, s3ReleaseNotes)
	if err != nil {
		return err
	}
	docsReleaseNotes := buildReleaseNotes(version, docsConfluentReleaseNotesBuilderParmas, content)
	updatedDocsPage, err := buildDocsPage(docsPath, confluentDocsPageHeader, docsReleaseNotes)
	if err != nil {
		return err
	}
	err = writeFile(updatedConflunetDocsFilePath, updatedDocsPage)
	if err != nil {
		return err
	}
	return nil
}

func buildReleaseNotes(version string, releaseNotesBuildParams *ReleaseNotesBuilderParams, content *ReleaseNotesContent) string {
	releaseNotesBuilder := NewReleaseNotesBuilder(version, releaseNotesBuildParams)
	return releaseNotesBuilder.buildReleaseNotes(content)
}

func buildDocsPage(docsFilePath string, docsHeader string, latestReleaseNotes string) (string, error) {
	docsUpdateHandler := NewDocsUpdateHandler(docsHeader, docsFilePath+"/"+docsPageFileName)
	updatedDocsPage, err := docsUpdateHandler.getUpdatedDocsPage(latestReleaseNotes)
	if err != nil {
		return "", err
	}
	return updatedDocsPage, nil
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
