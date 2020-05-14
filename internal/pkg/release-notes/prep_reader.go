package release_notes

import (
	"bufio"
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
)

type PrepFileReader interface {
	getSectionsMap() (map[SectionType][]string, error)
}

type PrepFileReaderImpl struct {
	prepFilePath string
	scanner      *bufio.Scanner
	sections     map[SectionType][]string
}

func NewPrepFileReader(prepFilePath string) PrepFileReader {
	return &PrepFileReaderImpl{
		prepFilePath: prepFilePath,
	}
}


func (p *PrepFileReaderImpl) getSectionsMap() (map[SectionType][]string, error) {
	err := p.initializeFileScanner()
	if err != nil {
		return nil, err
	}
	p.sections = make(map[SectionType][]string)
	err = p.extractSections()
	if err != nil {
		return nil, err
	}
	return p.sections, nil
}

func (p *PrepFileReaderImpl) initializeFileScanner() error {
	f, err := os.Open(p.prepFilePath)
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
