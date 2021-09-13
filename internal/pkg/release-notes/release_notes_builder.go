package release_notes

import (
	"fmt"
	"strings"
)

const (
	breakingChangesSectionTitle = "Breaking Changes"
	newFeaturesSectionTitle     = "New Features"
	bugFixesSectionTitle        = "Bug Fixes"
	noChangeContentFormat       = "No changes relating to %s for this version."
	bulletPointFormat           = "  - %s"
)

type ReleaseNotesBuilder interface {
	buildReleaseNotes(content *ReleaseNotesContent) string
}

type ReleaseNotesBuilderParams struct {
	cliDisplayName      string
	titleFormat         string
	sectionHeaderFormat string
}

type ReleaseNotesBuilderImpl struct {
	*ReleaseNotesBuilderParams
	version string
}

func NewReleaseNotesBuilder(version string, params *ReleaseNotesBuilderParams) ReleaseNotesBuilder {
	return &ReleaseNotesBuilderImpl{
		ReleaseNotesBuilderParams: params,
		version:                   version,
	}
}

func (b *ReleaseNotesBuilderImpl) buildReleaseNotes(content *ReleaseNotesContent) string {
	title := fmt.Sprintf(b.titleFormat, b.cliDisplayName, b.version)
	breakingChangesSection := b.buildSection(breakingChangesSectionTitle, content.breakingChanges)
	newFeaturesSection := b.buildSection(newFeaturesSectionTitle, content.newFeatures)
	bugFixesSection := b.buildSection(bugFixesSectionTitle, content.bugFixes)
	return title + "\n" + b.getReleaseNotesContent(breakingChangesSection, newFeaturesSection, bugFixesSection)
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

func (b *ReleaseNotesBuilderImpl) getReleaseNotesContent(sections ...string) string {
	var fullSections []string
	for _, section := range sections {
		if section != "" {
			fullSections = append(fullSections, section)
		}
	}

	if len(fullSections) == 0 {
		return fmt.Sprintf(noChangeContentFormat, b.cliDisplayName)
	}

	return strings.Join(fullSections, "\n\n")
}
