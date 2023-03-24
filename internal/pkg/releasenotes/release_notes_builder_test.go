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
	newFeatureAndBugFixContent *ReleaseNotes
	noBugFixContent            *ReleaseNotes
	noNewFeatureContent        *ReleaseNotes
	noChangeContent            *ReleaseNotes
}

func TestReleaseNotesBuilderTestSuite(t *testing.T) {
	suite.Run(t, new(ReleaseNotesBuilderTestSuite))
}

func (suite *ReleaseNotesBuilderTestSuite) SetupSuite() {
	suite.version = "1.2.3"
	bugFixList := []string{"bug fixes1", "bug fixes2"}
	newFeatureList := []string{"add `confluent command-1`", "add `confluent command-2`"}
	suite.newFeatureAndBugFixContent = &ReleaseNotes{
		minor: newFeatureList,
		patch: bugFixList,
	}
	suite.noBugFixContent = &ReleaseNotes{
		minor: newFeatureList,
		patch: []string{},
	}
	suite.noNewFeatureContent = &ReleaseNotes{
		minor: []string{},
		patch: bugFixList,
	}
	suite.noChangeContent = &ReleaseNotes{
		minor: []string{},
		patch: []string{},
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
		content  *ReleaseNotes
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
	}
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			builder := NewReleaseNotesBuilder(suite.version, releaseNotesBuilderParams)
			builder.date = time.Date(1999, time.February, 24, 0, 0, 0, 0, time.UTC)

			want, err := readTestFile(tt.wantFile)
			require.NoError(t, err)

			releaseNotes := builder.buildReleaseNotes(tt.content)
			require.Equal(t, want, releaseNotes)
		})
	}
}
