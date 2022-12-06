package releasenotes

import (
	"io"
	"os"
	"path/filepath"
)

const (
	docsPageFileName       = "release-notes.rst"
	s3ReleaseNotesFilePath = "./release-notes/latest-release.rst"
	updatedDocsFilePath    = "./release-notes/release-notes.rst"

	titleFormat = "[%s] %s %s Release Notes"

	s3SectionHeaderFormat   = "%s\n-------------"
	docsSectionHeaderFormat = "**%s**"

	s3CLIName   = "Confluent CLI"
	docsCLIName = "|confluent-cli|"
)

var (
	s3ReleaseNotesBuilderParams = &ReleaseNotesBuilderParams{
		cliDisplayName:      s3CLIName,
		sectionHeaderFormat: s3SectionHeaderFormat,
	}
	docsReleaseNotesBuilderParams = &ReleaseNotesBuilderParams{
		cliDisplayName:      docsCLIName,
		sectionHeaderFormat: docsSectionHeaderFormat,
	}
)

func WriteReleaseNotes(docsPath, releaseVersion string) error {
	releaseNotesContent, err := getReleaseNotesContent()
	if err != nil {
		return err
	}
	return buildAndWriteReleaseNotes(releaseVersion, releaseNotesContent, docsPath)
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
	s3ReleaseNotesBuilder := NewReleaseNotesBuilder(version, s3ReleaseNotesBuilderParams)
	s3ReleaseNotes := s3ReleaseNotesBuilder.buildS3ReleaseNotes(content)
	err := writeFile(s3ReleaseNotesFilePath, s3ReleaseNotes)
	if err != nil {
		return err
	}
	docsReleaseNotesBuilder := NewReleaseNotesBuilder(version, docsReleaseNotesBuilderParams)
	docsReleaseNotes := docsReleaseNotesBuilder.buildDocsReleaseNotes(content)
	updatedDocsPage, err := buildDocsPage(docsPath, docsPageHeader, docsReleaseNotes)
	if err != nil {
		return err
	}
	return writeFile(updatedDocsFilePath, updatedDocsPage)
}

func buildDocsPage(docsFilePath, docsHeader, latestReleaseNotes string) (string, error) {
	docsUpdateHandler := NewDocsUpdateHandler(docsHeader, filepath.Join(docsFilePath, docsPageFileName))
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
