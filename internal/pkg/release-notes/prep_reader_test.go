package releasenotes

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Prep_Reader_Impl_Read_File(t *testing.T) {
	tests := []struct {
		name                string
		prepFile            string
		wantBreakingChanges []string
		wantNewFeatures     []string
		wantBugFixes        []string
	}{
		{
			name:                "test get sections with good formatting",
			prepFile:            "test_files/prep1",
			wantBreakingChanges: []string{"breakingchange1", "breakingchange2"},
			wantNewFeatures:     []string{"feature1", "feature2"},
			wantBugFixes:        []string{"bug1", "bug2"},
		},
		{
			name:                "test get sections with bad formatting",
			prepFile:            "test_files/prep2",
			wantBreakingChanges: []string{"breakingchange1", "breakingchange2"},
			wantNewFeatures:     []string{"feature1", "feature2"},
			wantBugFixes:        []string{"bug1", "bug2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prepReader := PrepFileReaderImpl{}
			err := prepReader.ReadPrepFile(tt.prepFile)
			require.NoError(t, err)
			require.Equal(t, prepReader.sections[newFeatures], tt.wantNewFeatures)
			require.Equal(t, prepReader.sections[bugFixes], tt.wantBugFixes)
			require.Equal(t, prepReader.sections[breakingChanges], tt.wantBreakingChanges)
		})
	}
}

func Test_Prep_Reader_Impl_Get_Content_Funcs(t *testing.T) {
	tests := []struct {
		name                string
		sections            map[SectionType][]string
		wantBreakingChanges []string
		wantNewFeatures     []string
		wantBugFixes        []string
	}{
		{
			name: "basic release notes",
			sections: map[SectionType][]string{
				breakingChanges: {"breakingchange1", "breakingchange2"},
				newFeatures:     {"feature1", "feature2"},
				bugFixes:        {"bug1", "bug2"},
			},
			wantBreakingChanges: []string{"breakingchange1", "breakingchange2"},
			wantNewFeatures:     []string{"feature1", "feature2"},
			wantBugFixes:        []string{"bug1", "bug2"},
		},
		{
			name: "empty bug fixes",
			sections: map[SectionType][]string{
				breakingChanges: {"breakingchange1", "breakingchange2"},
				newFeatures:     {"feature1", "feature2"},
				bugFixes:        {},
			},
			wantBreakingChanges: []string{"breakingchange1", "breakingchange2"},
			wantNewFeatures:     []string{"feature1", "feature2"},
			wantBugFixes:        []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prepReader := PrepFileReaderImpl{}
			prepReader.sections = tt.sections

			confluentContent, err := prepReader.GetReleaseNotesContent()
			require.NoError(t, err)
			require.Equal(t, tt.wantNewFeatures, confluentContent.newFeatures)
			require.Equal(t, tt.wantBugFixes, confluentContent.bugFixes)
		})
	}
}
