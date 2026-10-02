package flink

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	pflink "github.com/confluentinc/cli/v4/pkg/flink"
)

func TestApplicationCreateRejectsMissingMetadataNameFromResourceFile(t *testing.T) {
	resourceFilePath := filepath.Join(t.TempDir(), "application.yaml")
	err := os.WriteFile(resourceFilePath, []byte("metadata: {}\n"), 0600)
	require.NoError(t, err)

	application, err := readApplicationResourceFile(resourceFilePath)
	require.NoError(t, err)

	_, err = (&pflink.CmfRestClient{}).CreateApplication(context.Background(), "environment", application)
	require.EqualError(t, err, `application name is required: ensure the resource file contains a non-empty "metadata.name" field`)
}
