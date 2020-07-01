package release_notes

import (
	"fmt"
	"testing"

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
		newFeatures: nil,
		bugFixes:    bugFixList,
	}
	suite.noChangeContent = &ReleaseNotesContent{
		newFeatures: []string{},
		bugFixes:    []string{},
	}
}

func (suite *ReleaseNotesBuilderTestSuite) TestS3CCloud() {
	suite.runTest("S3 CCloud", "s3_ccloud", s3CCloudReleaseNotesBuilderParams)
}

func (suite *ReleaseNotesBuilderTestSuite) TestS3Confluent() {
	suite.runTest("S3 Confluent", "s3_confluent", s3ConfluentReleaseNotesBuilderParams)
}

func (suite *ReleaseNotesBuilderTestSuite) TestDocsCCloud() {
	suite.runTest("Docs CCloud", "docs_ccloud", docsCCloudReleaseNotesBuilderParams)
}

func (suite *ReleaseNotesBuilderTestSuite) TestDocsConfluent() {
	suite.runTest("Docs Confluent", "docs_confluent", docsConfluentReleaseNotesBuilderParmas)
}

func (suite *ReleaseNotesBuilderTestSuite) runTest(testNamePrefix string, outputFilePrefix string, releaseNotesBuilderParams *ReleaseNotesBuilderParams) {
	tests := []struct {
		name     string
		content  *ReleaseNotesContent
		wantFile string
	}{
		{
			name:     fmt.Sprintf("%s basics release notes", testNamePrefix),
			content:  suite.newFeatureAndBugFixContent,
			wantFile: fmt.Sprintf("test_files/output/%s_release_notes_builder_both", outputFilePrefix),
		},
		{
			name:     fmt.Sprintf("%s no bug fixes", testNamePrefix),
			content:  suite.noBugFixContent,
			wantFile: fmt.Sprintf("test_files/output/%s_release_notes_builder_no_bug", outputFilePrefix),
		},
		{
			name:     fmt.Sprintf("%s no new features", testNamePrefix),
			content:  suite.noNewFeatureContent,
			wantFile: fmt.Sprintf("test_files/output/%s_release_notes_builder_no_new_feature", outputFilePrefix),
		},
		{
			name:     fmt.Sprintf("%s no change", testNamePrefix),
			content:  suite.noChangeContent,
			wantFile: fmt.Sprintf("test_files/output/%s_release_notes_builder_no_change", outputFilePrefix),
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			releaseNotesBuilder := NewReleaseNotesBuilder(suite.version, releaseNotesBuilderParams)
			releaseNotes := releaseNotesBuilder.buildReleaseNotes(tt.content)
			want, err := readTestFile(tt.wantFile)
			require.NoError(t, err)
			require.Equal(t, want, releaseNotes)
		})
	}
}
