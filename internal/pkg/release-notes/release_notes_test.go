package release_notes

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"runtime"
	"strings"
	"testing"

	testUtils "github.com/confluentinc/cli/test"
)

func Test_Prep_Reader_Imp_Read_File(t *testing.T) {

	tests := []struct {
		name                     string
		prepFile                 string
		wantBothNewFeatures      []string
		wantBothBugFixes         []string
		wantCCloudNewFeatures    []string
		wantCCloudBugFixes       []string
		wantConfluentNewFeatures []string
		wantConfluentBugFixes    []string
	}{
		{
			name:                     "test get sections map",
			prepFile:                 "test_files/prep1",
			wantBothNewFeatures:      []string{"both feature1", "both feature2"},
			wantBothBugFixes:         []string{"both bug1", "both bug2"},
			wantCCloudNewFeatures:    []string{"ccloud feature1", "ccloud feature2"},
			wantCCloudBugFixes:       []string{"ccloud bug1", "ccloud bug2"},
			wantConfluentNewFeatures: []string{"confluent new feature1", "confluent new feature2"},
			wantConfluentBugFixes:    []string{"confluent bug1", "confluent bug2"},
		},
		{
			name:                     "test get sections map",
			prepFile:                 "test_files/prep2",
			wantBothNewFeatures:      []string{"both feature1", "both feature2"},
			wantBothBugFixes:         []string{"both bug1", "both bug2"},
			wantCCloudNewFeatures:    []string{"ccloud feature1", "ccloud feature2"},
			wantCCloudBugFixes:       []string{"ccloud bug1", "ccloud bug2"},
			wantConfluentNewFeatures: []string{"confluent new feature1", "confluent new feature2"},
			wantConfluentBugFixes:    []string{"confluent bug1", "confluent bug2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prepReader := PrepFileReaderImpl{}
			err := prepReader.ReadPrepFile(tt.prepFile)
			require.NoError(t, err)
			require.Equal(t, prepReader.sections[bothNewFeatures], tt.wantBothNewFeatures)
			require.Equal(t, prepReader.sections[bothBugFixes], tt.wantBothBugFixes)
			require.Equal(t, prepReader.sections[ccloudNewFeatures], tt.wantCCloudNewFeatures)
			require.Equal(t, prepReader.sections[ccloudBugFixes], tt.wantCCloudBugFixes)
			require.Equal(t, prepReader.sections[confluentNewFeatures], tt.wantConfluentNewFeatures)
			require.Equal(t, prepReader.sections[confluentBugFixes], tt.wantConfluentBugFixes)
		})
	}
}

func Test_Prep_Reader_Impl_Get_Content_Funcs(t *testing.T) {
	sections := map[SectionType][]string{
		bothNewFeatures:      {"both feature1", "both feature2"},
		bothBugFixes:         {"both bug1", "both bug2"},
		ccloudNewFeatures:    {"ccloud feature1", "ccloud feature2"},
		ccloudBugFixes:       {"ccloud bug1", "ccloud bug2"},
		confluentNewFeatures: {"confluent new feature1", "confluent new feature2"},
		confluentBugFixes:    {"confluent bug1", "confluent bug2"},
	}
	sectionsNoConfluentBugFix := map[SectionType][]string{
		bothNewFeatures:      {"both feature1", "both feature2"},
		bothBugFixes:         {},
		ccloudNewFeatures:    {"ccloud feature1", "ccloud feature2"},
		ccloudBugFixes:       {"ccloud bug1", "ccloud bug2"},
		confluentNewFeatures: {"confluent new feature1", "confluent new feature2"},
		confluentBugFixes:    {},
	}
	tests := []struct {
		name                     string
		sections                 map[SectionType][]string
		wantCCloudNewFeatures    []string
		wantCCloudBugFixes       []string
		wantConfluentNewFeatures []string
		wantConfleuntBugFixes    []string
	}{
		{
			name:                     "basics release notes",
			sections:                 sections,
			wantCCloudNewFeatures:    []string{"ccloud feature1", "ccloud feature2", "both feature1", "both feature2"},
			wantCCloudBugFixes:       []string{"ccloud bug1", "ccloud bug2", "both bug1", "both bug2"},
			wantConfluentNewFeatures: []string{"confluent new feature1", "confluent new feature2", "both feature1", "both feature2"},
			wantConfleuntBugFixes:    []string{"confluent bug1", "confluent bug2", "both bug1", "both bug2"},
		},
		{
			name:                     "empty bug fixes",
			sections:                 sectionsNoConfluentBugFix,
			wantCCloudNewFeatures:    []string{"ccloud feature1", "ccloud feature2", "both feature1", "both feature2"},
			wantCCloudBugFixes:       []string{"ccloud bug1", "ccloud bug2"},
			wantConfluentNewFeatures: []string{"confluent new feature1", "confluent new feature2", "both feature1", "both feature2"},
			wantConfleuntBugFixes:    []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prepReader := PrepFileReaderImpl{}
			prepReader.sections = tt.sections
			ccloudContent, err := prepReader.GetCCloudReleaseNotesContent()
			require.NoError(t, err)
			require.Equal(t, tt.wantCCloudNewFeatures, ccloudContent.newFeatures)
			require.Equal(t, tt.wantCCloudBugFixes, ccloudContent.bugFixes)

			confluentContent, err := prepReader.GetConfluentReleaseNotesContent()
			require.NoError(t, err)
			require.Equal(t, tt.wantConfluentNewFeatures, confluentContent.newFeatures)
			require.Equal(t, tt.wantConfleuntBugFixes, confluentContent.bugFixes)
		})
	}
}

func Test_Release_Notes_Builder(t *testing.T) {
	content := &ReleaseNotesContent{
		newFeatures: []string{"new feature1", "new feature2"},
		bugFixes:    []string{"bug fixes1", "bug fixes2"},
	}
	contentNoBugFix := &ReleaseNotesContent{
		newFeatures: []string{"new feature1", "new feature2"},
		bugFixes:    []string{},
	}
	contentNoChange := &ReleaseNotesContent{
		newFeatures: []string{},
		bugFixes:    []string{},
	}
	tests := []struct {
		name     string
		content  *ReleaseNotesContent
		wantFile string
	}{
		{
			name:     "basics release notes",
			content:  content,
			wantFile: "test_files/release_notes_builder_output1",
		},
		{
			name:     "empty bug fixes",
			content:  contentNoBugFix,
			wantFile: "test_files/release_notes_builder_output2",
		},
		{
			name:     "empty bug fixes",
			content:  contentNoChange,
			wantFile: "test_files/release_notes_builder_output3",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s3CCloudReleaseNotesBuildParams.version = "v1.2.3"
			releaseNotesBuilder := NewReleaseNotesBuilder(s3CCloudReleaseNotesBuildParams)
			releaseNotes := releaseNotesBuilder.buildReleaseNotes(tt.content)
			want, err := readTestFile(tt.wantFile)
			require.NoError(t, err)
			require.Equal(t, want, releaseNotes)
		})
	}
}

func Test_Docs_Update_Handler(t *testing.T) {
	newReleaseNotes := `|ccloud| CLI v1.2.0 Release Notes
=================================

New Features
------------------------
- 1.2 cloud feature
- 1.2 both feat

Bug Fixes
------------------------
- 1.2 cloud bug
- 1.2 two both bugs`

	if runtime.GOOS == "windows" {
		newReleaseNotes = strings.ReplaceAll(newReleaseNotes, "\n", "\r\n")
	}

	tests := []struct {
		name            string
		newReleaseNotes string
		docsFile        string
		wantFile        string
	}{
		{
			name:            "basics release notes",
			newReleaseNotes: newReleaseNotes,
			docsFile:        "test_files/release-notes.rst",
			wantFile:        "test_files/docs_update_handler_output",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			docsUpdateHandler := NewDocsUpdateHandler(ccloudDocsPageHeader, tt.docsFile)
			docs, err := docsUpdateHandler.getUpdatedDocsPage(tt.newReleaseNotes)
			require.NoError(t, err)
			want, err := readTestFile(tt.wantFile)
			require.NoError(t, err)
			// got windows docs result will contain /r/n but readTestfile already uses NormalizeNewLines
			docs = testUtils.NormalizeNewLines(docs)
			require.Equal(t, want, docs)
		})
	}
}

func readTestFile(filePath string) (string, error) {
	fileBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("Unable to load output file.")
	}
	fileContent := string(fileBytes)
	return testUtils.NormalizeNewLines(fileContent), nil
}
