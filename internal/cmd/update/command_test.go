package update

import (
	"testing"

	"github.com/stretchr/testify/require"

	updatemock "github.com/confluentinc/cli/internal/pkg/update/mock"
	"github.com/confluentinc/cli/internal/pkg/version"
)

func TestGetReleaseNotes_MultipleReleaseNotes(t *testing.T) {
	t.Parallel()

	client := &updatemock.Client{
		GetLatestReleaseNotesFunc: func(_, _ string) (string, []string, error) {
			notes := []string{
				"v0.1.0 changes\n",
				"v1.0.0 changes\n",
			}
			return "1.0.0", notes, nil
		},
	}

	c := &command{
		client:  client,
		version: &version.Version{Version: "0.0.0"},
	}

	require.Equal(t, "v0.1.0 changes\n\nv1.0.0 changes\n", c.getReleaseNotes("confluent", "1.0.0"))
}
