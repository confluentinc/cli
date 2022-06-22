package releasenotes

import (
	"fmt"
	"strings"
	"time"
)

const (
	breakingChangesSectionTitle = "Breaking Changes"
	newFeaturesSectionTitle     = "New Features"
	bugFixesSectionTitle        = "Bug Fixes"
	noChangeContentFormat       = "No changes relating to %s for this version."
	bulletPointFormat           = "  - %s"
)

type ReleaseNotesBuilderParams struct {
	cliDisplayName      string
	sectionHeaderFormat string
}

type ReleaseNotesBuilder struct {
	*ReleaseNotesBuilderParams
	date    time.Time
	version string
}

func NewReleaseNotesBuilder(version string, params *ReleaseNotesBuilderParams) *ReleaseNotesBuilder {
	return &ReleaseNotesBuilder{
		ReleaseNotesBuilderParams: params,
		date:                      time.Now(),
		version:                   version,
	}
}

func (b *ReleaseNotesBuilder) buildS3ReleaseNotes(content *ReleaseNotesContent) string {
	title := fmt.Sprintf(titleFormat, b.buildDate(), b.cliDisplayName, b.version)
	underline := strings.Repeat("=", len(title))
	title = "\n" + underline + "\n" + title + "\n" + underline + "\n"

	breakingChangesSection := b.buildSection(breakingChangesSectionTitle, content.breakingChanges)
	newFeaturesSection := b.buildSection(newFeaturesSectionTitle, content.newFeatures)
	bugFixesSection := b.buildSection(bugFixesSectionTitle, content.bugFixes)
	return title + "\n" + b.getReleaseNotesContent(breakingChangesSection, newFeaturesSection, bugFixesSection)
}

func (b *ReleaseNotesBuilder) buildDocsReleaseNotes(content *ReleaseNotesContent) string {
	title := fmt.Sprintf(titleFormat, b.buildDate(), b.cliDisplayName, b.version)
	underline := strings.Repeat("=", len(title))
	title = "\n" + title + "\n" + underline

	breakingChangesSection := b.buildSection(breakingChangesSectionTitle, content.breakingChanges)
	newFeaturesSection := b.buildSection(newFeaturesSectionTitle, content.newFeatures)
	bugFixesSection := b.buildSection(bugFixesSectionTitle, content.bugFixes)
	return title + "\n" + b.getReleaseNotesContent(breakingChangesSection, newFeaturesSection, bugFixesSection)
}

func (b *ReleaseNotesBuilder) buildDate() string {
	return b.date.Format("1/2/2006")
}

func (b *ReleaseNotesBuilder) buildSection(sectionTitle string, sectionElements []string) string {
	if len(sectionElements) == 0 {
		return ""
	}
	sectionHeader := fmt.Sprintf(b.sectionHeaderFormat, sectionTitle)
	bulletPoints := b.buildBulletPoints(sectionElements)
	return sectionHeader + "\n" + bulletPoints
}

func (b *ReleaseNotesBuilder) buildBulletPoints(elements []string) string {
	var bulletPointList []string
	for _, element := range elements {
		bulletPointList = append(bulletPointList, fmt.Sprintf(bulletPointFormat, element))
	}
	return strings.Join(bulletPointList, "\n")
}

func (b *ReleaseNotesBuilder) getReleaseNotesContent(sections ...string) string {
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
