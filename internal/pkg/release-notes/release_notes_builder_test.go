package releasenotes

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ReleaseNotesBuilderTestSuite struct {
	suite.Suite
	version                    string
	newFeatureAndBugFixContent *ReleaseNotesContent
	noBugFixContent            *ReleaseNotesContent
	noNewFeatureContent        *ReleaseNotesContent
	noChangeContent            *ReleaseNotesContent
}

func TestReleaseNotesBuilderTestSuite(t *testing.T) {
	suite.Run(t, new(ReleaseNotesBuilderTestSuite))
}

func (suite *ReleaseNotesBuilderTestSuite) SetupSuite() {
	suite.version = "v1.2.3"
	bugFixList := []string{"bug fixes1", "bug fixes2"}
	newFeatureList := []string{"new feature1", "new feature2"}
	suite.newFeatureAndBugFixContent = &ReleaseNotesContent{
		newFeatures: newFeatureList,
		bugFixes:    bugFixList,
	}
	suite.noBugFixContent = &ReleaseNotesContent{
		newFeatures: newFeatureList,
		bugFixes:    []string{},
	}
	suite.noNewFeatureContent = &ReleaseNotesContent{
		newFeatures: []string{},
		bugFixes:    bugFixList,
	}
	suite.noChangeContent = &ReleaseNotesContent{
		newFeatures: []string{},
		bugFixes:    []string{},
	}
}

func (suite *ReleaseNotesBuilderTestSuite) TestS3() {
	suite.runTest("s3", s3ReleaseNotesBuilderParams)
}

func (suite *ReleaseNotesBuilderTestSuite) TestDocs() {
	suite.runTest("docs", docsReleaseNotesBuilderParams)
}

func (suite *ReleaseNotesBuilderTestSuite) runTest(testNamePrefix string, releaseNotesBuilderParams *ReleaseNotesBuilderParams) {
	tests := []struct {
		name     string
		content  *ReleaseNotesContent
		wantFile string
	}{
		{
			name:     fmt.Sprintf("%s new features and bug fixes", testNamePrefix),
			content:  suite.newFeatureAndBugFixContent,
			wantFile: fmt.Sprintf("test_files/output/%s_release_notes_builder_both", testNamePrefix),
		},
		{
			name:     fmt.Sprintf("%s no bug fixes", testNamePrefix),
			content:  suite.noBugFixContent,
			wantFile: fmt.Sprintf("test_files/output/%s_release_notes_builder_no_bug_fixes", testNamePrefix),
		},
		{
			name:     fmt.Sprintf("%s no new features", testNamePrefix),
			content:  suite.noNewFeatureContent,
			wantFile: fmt.Sprintf("test_files/output/%s_release_notes_builder_no_new_features", testNamePrefix),
		},
		{
			name:     fmt.Sprintf("%s no changes", testNamePrefix),
			content:  suite.noChangeContent,
			wantFile: fmt.Sprintf("test_files/output/%s_release_notes_builder_no_changes", testNamePrefix),
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			builder := NewReleaseNotesBuilder(suite.version, releaseNotesBuilderParams)
			builder.date = time.Date(1999, time.February, 24, 0, 0, 0, 0, time.UTC)

			var releaseNotes string
			if testNamePrefix == "s3" {
				releaseNotes = builder.buildS3ReleaseNotes(tt.content)
			} else {
				releaseNotes = builder.buildDocsReleaseNotes(tt.content)
			}

			want, err := readTestFile(tt.wantFile)
			require.NoError(t, err)
			require.Equal(t, want, releaseNotes)
		})
	}
}
