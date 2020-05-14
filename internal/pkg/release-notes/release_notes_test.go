package release_notes

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"testing"
)

func Test_Prep_Reader(t *testing.T) {

	tests := []struct {
		name                     string
		prepFile                 string
		wantBothNewFeatures      []string
		wantBothBugFixes         []string
		wantCCloudNewFeatures    []string
		wantCCloudBugFixes       []string
		wantConfluentNewFeatures []string
		wantConfleuntBugFixes    []string
	}{
		{
			name:                     "test get sections map",
			prepFile:                 "test_files/prep1",
			wantBothNewFeatures:      []string{"both feature1", "both feature2"},
			wantBothBugFixes:         []string{"both bug1", "both bug2"},
			wantCCloudNewFeatures:    []string{"ccloud feature1", "ccloud feature2"},
			wantCCloudBugFixes:       []string{"ccloud bug1", "ccloud bug2"},
			wantConfluentNewFeatures: []string{"confluent new feature1", "confluent new feature2"},
			wantConfleuntBugFixes:    []string{"confluent bug1", "confluent bug2"},
		},
		{
			name:                     "test get sections map",
			prepFile:                 "test_files/prep2",
			wantBothNewFeatures:      []string{"both feature1", "both feature2"},
			wantBothBugFixes:         []string{"both bug1", "both bug2"},
			wantCCloudNewFeatures:    []string{"ccloud feature1", "ccloud feature2"},
			wantCCloudBugFixes:       []string{"ccloud bug1", "ccloud bug2"},
			wantConfluentNewFeatures: []string{"confluent new feature1", "confluent new feature2"},
			wantConfleuntBugFixes:    []string{"confluent bug1", "confluent bug2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prepReader := NewPrepFileReader(tt.prepFile)
			sections, err := prepReader.getSectionsMap()
			require.NoError(t, err)
			require.Equal(t, sections[bothNewFeatures], tt.wantBothNewFeatures)
			require.Equal(t, sections[bothBugFixes], tt.wantBothBugFixes)
			require.Equal(t, sections[ccloudNewFeatures], tt.wantCCloudNewFeatures)
			require.Equal(t, sections[ccloudBugFixes], tt.wantCCloudBugFixes)
			require.Equal(t, sections[confluentNewFeatures], tt.wantConfluentNewFeatures)
			require.Equal(t, sections[confluentBugFixes], tt.wantConfleuntBugFixes)
		})
	}
}

func Test_Extract_Release_Notes_Content(t *testing.T) {
	sections := map[SectionType][]string {
		bothNewFeatures:      {"both feature1", "both feature2"},
		bothBugFixes:         {"both bug1", "both bug2"},
		ccloudNewFeatures:    {"ccloud feature1", "ccloud feature2"},
		ccloudBugFixes:       {"ccloud bug1", "ccloud bug2"},
		confluentNewFeatures: {"confluent new feature1", "confluent new feature2"},
		confluentBugFixes:    {"confluent bug1", "confluent bug2"},
	}
	sectionsNoConfluentBugFix := map[SectionType][]string {
		bothNewFeatures:      {"both feature1", "both feature2"},
		bothBugFixes:         {},
		ccloudNewFeatures:    {"ccloud feature1", "ccloud feature2"},
		ccloudBugFixes:       {"ccloud bug1", "ccloud bug2"},
		confluentNewFeatures: {"confluent new feature1", "confluent new feature2"},
		confluentBugFixes:    {},
	}
	tests := []struct {
		name     string
		sections  map[SectionType][]string
		wantCCloudNewFeatures    []string
		wantCCloudBugFixes       []string
		wantConfluentNewFeatures []string
		wantConfleuntBugFixes    []string
	}{
		{
			name:                     "basics release notes",
			sections: 				  sections,
			wantCCloudNewFeatures:    []string{"ccloud feature1", "ccloud feature2", "both feature1", "both feature2"},
			wantCCloudBugFixes:       []string{"ccloud bug1", "ccloud bug2", "both bug1", "both bug2"},
			wantConfluentNewFeatures: []string{"confluent new feature1", "confluent new feature2", "both feature1", "both feature2"},
			wantConfleuntBugFixes:    []string{"confluent bug1", "confluent bug2", "both bug1", "both bug2"},
		},
		{
			name:                     "empty bug fixes",
			sections: sectionsNoConfluentBugFix,
			wantCCloudNewFeatures:    []string{"ccloud feature1", "ccloud feature2", "both feature1", "both feature2"},
			wantCCloudBugFixes:       []string{"ccloud bug1", "ccloud bug2"},
			wantConfluentNewFeatures: []string{"confluent new feature1", "confluent new feature2", "both feature1", "both feature2"},
			wantConfleuntBugFixes:    []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ccloudContent := extractReleaseNotesContent(tt.sections, "ccloud")
			require.Equal(t, tt.wantCCloudNewFeatures, ccloudContent.newFeatures)
			require.Equal(t, tt.wantCCloudBugFixes, ccloudContent.bugFixes)

			confluentContent := extractReleaseNotesContent(tt.sections, "confluent")
			require.Equal(t, tt.wantConfluentNewFeatures, confluentContent.newFeatures)
			require.Equal(t, tt.wantConfleuntBugFixes, confluentContent.bugFixes)
		})
	}
}

func Test_Release_Notes_Builder(t *testing.T) {
	content := ReleaseNotesContent{
		newFeatures: []string{"new feature1", "new feature2"},
		bugFixes:    []string{"bug fixes1", "bug fixes2"},
	}
	contentNoBugFix := ReleaseNotesContent{
		newFeatures: []string{"new feature1", "new feature2"},
		bugFixes:    []string{},
	}
	contentNoChange := ReleaseNotesContent{
		newFeatures: []string{},
		bugFixes:    []string{},
	}
	tests := []struct {
		name     string
		content  ReleaseNotesContent
		wantFile  string
	}{
		{
			name:                     "basics release notes",
			content: content,
			wantFile:                 "test_files/release_notes_builder_output1",
		},
		{
			name:                     "empty bug fixes",
			content: contentNoBugFix,
			wantFile:                 "test_files/release_notes_builder_output2",
		},
		{
			name:                     "empty bug fixes",
			content: contentNoChange,
			wantFile:                 "test_files/release_notes_builder_output3",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			releaseNotesBuilder := NewReleaseNotesBuilder(s3ReleaseNotesTitleFormat, s3SectionHeaderFormat,
				"ccloud", "v1.2.3")
			releaseNotes := releaseNotesBuilder.buildReleaseNotes(tt.content)
			want, err := readTestFile(tt.wantFile)
			require.NoError(t, err)
			require.Equal(t, want, releaseNotes)
		})
	}
}

func readTestFile(filePath string) (string, error) {
	fileBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("Unable to load output file.")
	}
	fileContent := string(fileBytes)
	return fileContent, nil
}

func Test_Docs_Update_Handler(t *testing.T) {
	newReleaseNotes := `v1.3.0 Release Notes
==========================

New Features
------------------------
- 1.3 cloud feature
- 1.3 both feat

Bug Fixes
------------------------
- 1.3 cloud bug
- 1.3 two both bugs`

	tests := []struct {
		name     string
		newReleaseNotes string
		docsFile string
		wantFile  string
	}{
		{
			name:                     "basics release notes",
			newReleaseNotes: newReleaseNotes,
			docsFile:                 "test_files/release-notes.rst",
			wantFile:                 "test_files/docs_update_handler_output",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			docsUpdateHandler := NewDocsUpdateHandler("ccloud", tt.docsFile)
			docs, err := docsUpdateHandler.getUpdatedDocsPage(tt.newReleaseNotes)
			require.NoError(t, err)
			want, err := readTestFile(tt.wantFile)
			require.NoError(t, err)
			require.Equal(t, want, docs)
		})
	}
}

