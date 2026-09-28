package flink

import (
	"testing"

	"github.com/stretchr/testify/require"

	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"
)

func TestArtifactOutOnPrem(t *testing.T) {
	t.Run("fills every column from a complete response", func(t *testing.T) {
		artifact := cmfsdk.Artifact{
			Metadata: cmfsdk.ArtifactMetadata{
				Name:              "my-udf.jar",
				CreationTimestamp: cmfsdk.PtrString("2026-01-01T00:00:00Z"),
			},
			Status: &cmfsdk.ArtifactStatus{
				Version:           cmfsdk.PtrInt32(3),
				Phase:             cmfsdk.PtrString("READY"),
				Size:              cmfsdk.PtrInt64(1024),
				Checksum:          cmfsdk.PtrString("sha256:abc"),
				CreationTimestamp: cmfsdk.PtrString("2026-02-01T00:00:00Z"),
			},
		}

		require.Equal(t, &artifactOutOnPrem{Name: "my-udf.jar", Version: "3", Phase: "READY", Size: "1024", CreationTime: "2026-01-01T00:00:00Z"}, newArtifactOutOnPrem(artifact))
		require.Equal(t, &artifactVersionOutOnPrem{Version: "3", Phase: "READY", Size: "1024", Checksum: "sha256:abc", CreationTime: "2026-02-01T00:00:00Z"}, newArtifactVersionOutOnPrem(artifact))
	})

	// CMF sends an empty status for an artifact listed while its last version is being deleted.
	for name, status := range map[string]*cmfsdk.ArtifactStatus{"no status": nil, "empty status": {}} {
		t.Run(name+" leaves the columns blank instead of 0", func(t *testing.T) {
			artifact := cmfsdk.Artifact{Metadata: cmfsdk.ArtifactMetadata{Name: "my-udf.jar"}, Status: status}

			require.Equal(t, &artifactOutOnPrem{Name: "my-udf.jar"}, newArtifactOutOnPrem(artifact))
			require.Equal(t, &artifactVersionOutOnPrem{}, newArtifactVersionOutOnPrem(artifact))
		})
	}
}
