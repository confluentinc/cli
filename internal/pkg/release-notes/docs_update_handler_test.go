package releasenotes

import (
	"testing"

	"github.com/confluentinc/cli/internal/pkg/utils"

	"github.com/stretchr/testify/require"
)

func Test_Docs_Update_Handler(t *testing.T) {
	newReleaseNotes := `|confluent-cli| v1.2.0 Release Notes
====================================

Breaking Changes
------------------------
- 1.2 breaking change

New Features
------------------------
- v1.2.0 feature

Bug Fixes
------------------------
- v1.2.0 bug`

	tests := []struct {
		name     string
		docsFile string
		wantFile string
	}{
		{
			name:     "basic release notes",
			docsFile: "test_files/release-notes.rst",
			wantFile: "test_files/output/docs_update_handler_output",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			docsUpdateHandler := NewDocsUpdateHandler(docsPageHeader, tt.docsFile)
			docs, err := docsUpdateHandler.getUpdatedDocsPage(newReleaseNotes)
			require.NoError(t, err)
			want, err := readTestFile(tt.wantFile)
			require.NoError(t, err)
			// got windows docs result will contain /r/n but readTestfile already uses NormalizeNewLines
			docs = utils.NormalizeNewLines(docs)
			require.Equal(t, want, docs)
		})
	}
}
