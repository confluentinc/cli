package release_notes

import (
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	testUtils "github.com/confluentinc/cli/test"
)

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
			wantFile:        "test_files/output/docs_update_handler_output",
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
