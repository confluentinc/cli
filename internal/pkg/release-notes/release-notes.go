package release_notes

import (
	"io"
	"os"
)

const (
	docsPageFileName       = "release-notes.rst"
	s3ReleaseNotesFilePath = "./release-notes/latest-release.rst"
	updatedDocsFilePath    = "./release-notes/release-notes.rst"

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

	s3CLIName   = "Confluent CLI"
	docsCLIName = "|confluent-cli|"
)

var (
	s3ReleaseNotesBuilderParams = &ReleaseNotesBuilderParams{
		cliDisplayName:      s3CLIName,
		titleFormat:         s3ReleaseNotesTitleFormat,
		sectionHeaderFormat: s3SectionHeaderFormat,
	}
	docsReleaseNotesBuilderParmas = &ReleaseNotesBuilderParams{
		cliDisplayName:      docsCLIName,
		titleFormat:         docsReleaseNotesTitleFormat,
		sectionHeaderFormat: docsSectionHeaderFormat,
	}
)

func WriteReleaseNotes(docsPath, releaseVersion string) error {
	releaseNotesContent, err := getReleaseNotesContent()
	if err != nil {
		return err
	}
	err = buildAndWriteReleaseNotes(releaseVersion, releaseNotesContent, docsPath)
	if err != nil {
		return err
	}
	return nil
}

func getReleaseNotesContent() (*ReleaseNotesContent, error) {
	prepFileReader := NewPrepFileReader()
	err := prepFileReader.ReadPrepFile(prepFileName)
	if err != nil {
		return nil, err
	}
	return prepFileReader.GetReleaseNotesContent()
}

func buildAndWriteReleaseNotes(version string, content *ReleaseNotesContent, docsPath string) error {
	s3ReleaseNotes := buildReleaseNotes(version, s3ReleaseNotesBuilderParams, content)
	err := writeFile(s3ReleaseNotesFilePath, s3ReleaseNotes)
	if err != nil {
		return err
	}
	docsReleaseNotes := buildReleaseNotes(version, docsReleaseNotesBuilderParmas, content)
	updatedDocsPage, err := buildDocsPage(docsPath, docsPageHeader, docsReleaseNotes)
	if err != nil {
		return err
	}
	err = writeFile(updatedDocsFilePath, updatedDocsPage)
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
