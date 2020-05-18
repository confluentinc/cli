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
	s3CCloudReleaseNotesBuildParams = &ReleaseNotesBuilderParams{
		cliDisplayName:      s3CCloudCLIName,
		titleFormat:         s3ReleaseNotesTitleFormat,
		sectionHeaderFormat: s3SectionHeaderFormat,
		version:             "",
	}
	s3ConfluentReleaseNotesBuildParams = &ReleaseNotesBuilderParams{
		cliDisplayName:      s3ConfluentCLIName,
		titleFormat:         s3ReleaseNotesTitleFormat,
		sectionHeaderFormat: s3SectionHeaderFormat,
		version:             "",
	}
	docsCCloudReleaseNotesBuildParams = &ReleaseNotesBuilderParams{
		cliDisplayName:      docsCCloudCLIName,
		titleFormat:         docsReleaseNotesTitleFormat,
		sectionHeaderFormat: docsSectionHeaderFormat,
		version:             "",
	}
	docsConfluentReleaseNotesBuildParmas = &ReleaseNotesBuilderParams{
		cliDisplayName:      docsConfluentCLIName,
		titleFormat:         docsReleaseNotesTitleFormat,
		sectionHeaderFormat: docsSectionHeaderFormat,
		version:             "",
	}
)

func WriteReleaseNotes(ccloudDocsPath, confluentDocsPath, releaseVersion string) error {
	ccloudReleaseNotesContent, confluentReleaseNotesContent, err := getCCloudAndConfluentReleaseNotesContent()
	if err != nil {
		return err
	}
	setReleaseNotesBuildParamsVersion(releaseVersion)
	err = buildAndWriteCCloudReleaseNotes(ccloudReleaseNotesContent, ccloudDocsPath)
	if err != nil {
		return err
	}
	err = buildAndWriteConfluentReleaseNotes(confluentReleaseNotesContent, confluentDocsPath)
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

func setReleaseNotesBuildParamsVersion(version string) {
	s3CCloudReleaseNotesBuildParams.version = version
	s3ConfluentReleaseNotesBuildParams.version = version
	docsCCloudReleaseNotesBuildParams.version = version
	docsConfluentReleaseNotesBuildParmas.version = version
}

func buildAndWriteCCloudReleaseNotes(content *ReleaseNotesContent, docsPath string) error {
	s3ReleaseNotes := buildReleaseNotes(s3CCloudReleaseNotesBuildParams, content)
	err := writeFile(s3CCloudReleaseNotesFilePath, s3ReleaseNotes)
	if err != nil {
		return err
	}
	ccloudDocsReleaseNotes := buildReleaseNotes(docsCCloudReleaseNotesBuildParams, content)
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

func buildAndWriteConfluentReleaseNotes(content *ReleaseNotesContent, docsPath string) error {
	s3ReleaseNotes := buildReleaseNotes(s3ConfluentReleaseNotesBuildParams, content)
	err := writeFile(s3ConfluentReleaseNotesFilePath, s3ReleaseNotes)
	if err != nil {
		return err
	}
	docsReleaseNotes := buildReleaseNotes(docsConfluentReleaseNotesBuildParmas, content)
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

func buildReleaseNotes(releaseNotesBuildParams *ReleaseNotesBuilderParams, content *ReleaseNotesContent) string {
	releaseNotesBuilder := NewReleaseNotesBuilder(releaseNotesBuildParams)
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
