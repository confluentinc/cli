package release_notes

import (
	"fmt"
	"strings"
)

const (
	newFeaturesSectionTitle = "New Features"
	bugFixesSectionTitle    = "Bug Fixes"
	noChangeContentFormat   = "No changes relating to %s CLI for this version."
	bulletPointFormat       = "  - %s"
)

type ReleaseNotesBuilder interface {
	buildReleaseNotes(content *ReleaseNotesContent) string
}

type ReleaseNotesBuilderParams struct {
	cliDisplayName      string
	titleFormat         string
	sectionHeaderFormat string
	version             string
}

type ReleaseNotesBuilderImpl struct {
	*ReleaseNotesBuilderParams
}

func NewReleaseNotesBuilder(params *ReleaseNotesBuilderParams) ReleaseNotesBuilder {
	return &ReleaseNotesBuilderImpl{params}
}

func (b *ReleaseNotesBuilderImpl) buildReleaseNotes(content *ReleaseNotesContent) string {
	newFeaturesSection := b.buildSection(newFeaturesSectionTitle, content.newFeatures)
	bugFixesSection := b.buildSection(bugFixesSectionTitle, content.bugFixes)
	title := fmt.Sprintf(b.titleFormat, b.cliDisplayName, b.version)
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
		bulletPointList = append(bulletPointList, fmt.Sprintf(bulletPointFormat, element))
	}
	return strings.Join(bulletPointList, "\n")
}

func (b *ReleaseNotesBuilderImpl) assembleReleaseNotes(title string, newFeaturesSection string, bugFixesSection string) string {
	content := b.getReleaseNotesContent(newFeaturesSection, bugFixesSection)
	return title + "\n" + content
}

func (b *ReleaseNotesBuilderImpl) getReleaseNotesContent(newFeaturesSection string, bugFixesSection string) string {
	if newFeaturesSection == "" && bugFixesSection == "" {
		return fmt.Sprintf(noChangeContentFormat, b.cliDisplayName)
	}
	return newFeaturesSection + "\n\n" + bugFixesSection
}
