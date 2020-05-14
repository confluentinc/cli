package release_notes

import (
	"fmt"
	"strings"
)

const (
	newFeaturesSectionTitle = "New Features"
	bugFixesSectionTitle = "Bug Fixes"
	noChangeContentFormat = "No changes relating to %s CLI for this version."

	s3ReleaseNotesTitleFormat = `
===============================
%s %s Release Notes
===============================
`
	docsReleaseNotesTitleFormat = `
%s %s Release Notes
=============================`

	s3SectionHeaderFormat = "%s\n-------------"
	docsSectionHeaderFormat = "**%s**"
)

type ReleaseNotesBuilder interface {
	buildReleaseNotes(content ReleaseNotesContent) string
}


type ReleaseNotesBuilderImpl struct {
	titleFormat         string
	sectionHeaderFormat string
	cliName             string
	version             string
}

func NewReleaseNotesBuilder(titleFormat string, sectionHeaderFormat string, cliName string, version string) ReleaseNotesBuilder {
	return &ReleaseNotesBuilderImpl{
		titleFormat:         titleFormat,
		sectionHeaderFormat: sectionHeaderFormat,
		cliName:             cliName,
		version:             version,
	}
}


func (b *ReleaseNotesBuilderImpl) buildReleaseNotes(content ReleaseNotesContent) string {
	newFeaturesSection := b.buildSection(newFeaturesSectionTitle, content.newFeatures)
	bugFixesSection := b.buildSection(bugFixesSectionTitle, content.bugFixes)
	title := fmt.Sprintf(b.titleFormat, b.cliName, b.version)
	return b.assembleReleaseNotes(title, newFeaturesSection, bugFixesSection)
}

func (b *ReleaseNotesBuilderImpl) buildSection(sectionTitle string, sectionElements []string) string {
	if len(sectionElements) == 0 {
		return ""
	}
	sectionHeader := fmt.Sprintf(b.sectionHeaderFormat, sectionTitle)
	bulletPoints := b.buildBulletPoints(sectionElements)
	return sectionHeader + "\n" + bulletPoints
}

func (b *ReleaseNotesBuilderImpl) buildBulletPoints(elements []string) string {
	var bulletPointList []string
	for _, element := range elements {
		bulletPointList = append(bulletPointList, fmt.Sprintf("  - %s", element))
	}
	return strings.Join(bulletPointList, "\n")
}

func (b *ReleaseNotesBuilderImpl) assembleReleaseNotes(title string, newFeaturesSection string, bugFixesSection string) string {
	content := b.getReleaseNotesContent(newFeaturesSection, bugFixesSection)
	return title + "\n" + content
}

func (b *ReleaseNotesBuilderImpl) getReleaseNotesContent(newFeaturesSection string, bugFixesSection string) string {
	if newFeaturesSection == "" && bugFixesSection == "" {
		return fmt.Sprintf(noChangeContentFormat, b.cliName)
	}
	return newFeaturesSection + "\n\n" + bugFixesSection
}


