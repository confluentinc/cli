package releasenotes

import (
	"fmt"
	"strings"
	"time"
)

const titleFormat = "[%s] %s v%s Release Notes"

const (
	majorSectionTitle = "Breaking Changes"
	minorSectionTitle = "New Features"
	patchSectionTitle = "Bug Fixes"
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

func (b *ReleaseNotesBuilder) buildS3ReleaseNotes(content *ReleaseNotes) string {
	title := fmt.Sprintf(titleFormat, b.buildDate(), b.cliDisplayName, b.version)
	underline := strings.Repeat("=", len(title))
	title = "\n" + title + "\n" + underline + "\n"

	majorSection := b.buildSection(majorSectionTitle, content.major)
	minorSection := b.buildSection(minorSectionTitle, content.minor)
	patchSection := b.buildSection(patchSectionTitle, content.patch)
	return title + "\n" + b.getReleaseNotesContent(majorSection, minorSection, patchSection)
}

func (b *ReleaseNotesBuilder) buildDocsReleaseNotes(content *ReleaseNotes) string {
	title := fmt.Sprintf(titleFormat, b.buildDate(), b.cliDisplayName, b.version)
	underline := strings.Repeat("=", len(title))
	title = "\n" + title + "\n" + underline

	breakingChangesSection := b.buildSection(majorSectionTitle, content.major)
	newFeaturesSection := b.buildSection(minorSectionTitle, content.minor)
	bugFixesSection := b.buildSection(patchSectionTitle, content.patch)
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
	bulletPointList := make([]string, len(elements))
	for i, element := range elements {
		bulletPointList[i] = fmt.Sprintf("  - %s", element)
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
		return fmt.Sprintf("No changes relating to %s for this version.", b.cliDisplayName)
	}

	return strings.Join(fullSections, "\n\n")
}
