package release_notes

import (
	"bufio"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"os"
	"strings"
)

type SectionType int

const (
	bothNewFeatures SectionType = iota
	bothBugFixes
	ccloudNewFeatures
	ccloudBugFixes
	confluentNewFeatures
	confluentBugFixes
)

var (
	sectionNameToSectionTypeMap = map[string]SectionType{
		bothNewFeaturesTitle:      bothNewFeatures,
		bothBugFixesTitle:         bothBugFixes,
		ccloudNewFeaturesTitle:    ccloudNewFeatures,
		ccloudBugFixesTitle:       ccloudBugFixes,
		confluentNewFeaturesTitle: confluentNewFeatures,
		confluentBugFixesTitle:    confluentBugFixes,
	}
	prepFileNotReadErrorMsg = "Prep file has not been read."
)

type PrepFileReader interface {
	ReadPrepFile(prepFilePath string) error
	GetCCloudReleaseNotesContent() (*ReleaseNotesContent, error)
	GetConfluentReleaseNotesContent() (*ReleaseNotesContent, error)
}

type PrepFileReaderImpl struct {
	scanner  *bufio.Scanner
	sections map[SectionType][]string
}

type ReleaseNotesContent struct {
	newFeatures []string
	bugFixes    []string
}

func NewPrepFileReader() PrepFileReader {
	return &PrepFileReaderImpl{}
}

func (p *PrepFileReaderImpl) ReadPrepFile(prepFilePath string) error {
	err := p.initializeFileScanner(prepFilePath)
	if err != nil {
		return err
	}
	p.sections = make(map[SectionType][]string)
	err = p.extractSections()
	if err != nil {
		return err
	}
	return nil
}

func (p *PrepFileReaderImpl) initializeFileScanner(prepFilePath string) error {
	f, err := os.Open(prepFilePath)
	if err != nil {
		return err
	}
	p.scanner = bufio.NewScanner(f)
	return nil
}

func (p *PrepFileReaderImpl) extractSections() error {
	var line string
	for p.isSectionName(line) || p.scanner.Scan() {
		line = p.scanner.Text()
		if section, isSectionName := p.checkForSectionName(line); isSectionName {
			line = p.extractSectionContent(section)
		}
	}
	err := p.scanner.Err()
	if err != nil {
		return err
	}
	return nil
}

func (p *PrepFileReaderImpl) checkForSectionName(line string) (SectionType, bool) {
	line = strings.TrimSpace(line)
	section, ok := sectionNameToSectionTypeMap[line]
	return section, ok
}

func (p *PrepFileReaderImpl) isSectionName(line string) bool {
	line = strings.TrimSpace(line)
	_, ok := sectionNameToSectionTypeMap[line]
	return ok
}

func (p *PrepFileReaderImpl) extractSectionContent(section SectionType) (lastLine string) {
	var sectionContent []string
	var line string
	for p.scanner.Scan() {
		line = p.scanner.Text()
		if !strings.HasPrefix(line, "-") {
			break
		}
		element := line[1:]
		element = strings.TrimSpace(element)
		if p.isPlaceHolder(element) {
			break
		}
		sectionContent = append(sectionContent, element)
	}
	p.sections[section] = sectionContent
	return line
}

func (p *PrepFileReaderImpl) isPlaceHolder(element string) bool {
	return element == placeHolder ||
		(strings.HasPrefix(element, "<") && strings.HasSuffix(element, ">"))
}

func (p *PrepFileReaderImpl) GetCCloudReleaseNotesContent() (*ReleaseNotesContent, error) {
	if p.sections == nil {
		return nil, errors.Errorf(prepFileNotReadErrorMsg)
	}
	content := &ReleaseNotesContent{
		newFeatures: p.getSectionContentList(ccloudNewFeatures, bothNewFeatures),
		bugFixes:    p.getSectionContentList(ccloudBugFixes, bothBugFixes),
	}
	return content, nil
}

func (p *PrepFileReaderImpl) GetConfluentReleaseNotesContent() (*ReleaseNotesContent, error) {
	if p.sections == nil {
		return nil, errors.Errorf(prepFileNotReadErrorMsg)
	}
	content := &ReleaseNotesContent{
		newFeatures: p.getSectionContentList(confluentNewFeatures, bothNewFeatures),
		bugFixes:    p.getSectionContentList(confluentBugFixes, bothBugFixes),
	}
	return content, nil
}

func (p *PrepFileReaderImpl) getSectionContentList(exclusiveSection, bothSection SectionType) []string {
	exclusiveContent := p.sections[exclusiveSection]
	bothContent := p.sections[bothSection]
	if len(exclusiveContent)+len(bothContent) == 0 {
		return []string{}
	}
	var contentList []string
	contentList = append(contentList, exclusiveContent...)
	contentList = append(contentList, bothContent...)
	return contentList
}
