package release_notes

import (
	"github.com/stretchr/testify/require"
	"testing"
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
