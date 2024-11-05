package flink

import (
	"github.com/spf13/cobra"

	flinkartifactv1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-artifact/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

type flinkArtifactOut struct {
	Id            string `human:"ID" serialized:"id"`
	Name          string `human:"Name" serialized:"name"`
	Version       string `human:"Version" serialized:"version"`
	Cloud         string `human:"Cloud" serialized:"cloud"`
	Region        string `human:"Region" serialized:"region"`
	Environment   string `human:"Environment" serialized:"environment"`
	ContentFormat string `human:"Content Format" serialized:"content_format"`
	Description   string `human:"Description" serialized:"description"`
	DocLink       string `human:"Documentation link" serialized:"doc_link"`
}

func (c *command) newArtifactCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "artifact",
		Short:       "Manage Flink UDF artifacts.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	cmd.AddCommand(c.newCreateCommand())
	cmd.AddCommand(c.newDeleteCommand())
	cmd.AddCommand(c.newDescribeCommand())
	cmd.AddCommand(c.newListCommand())

	return cmd
}

func printTable(cmd *cobra.Command, artifact flinkartifactv1.ArtifactV1FlinkArtifact) error {
	table := output.NewTable(cmd)

	var artifactVersion = ""
	if versions := artifact.GetVersions(); len(versions) > 0 {
		artifactVersion = versions[0].GetVersion()
	}

	table.Add(&flinkArtifactOut{
		Name:          artifact.GetDisplayName(),
		Id:            artifact.GetId(),
		Version:       artifactVersion,
		Cloud:         artifact.GetCloud(),
		Region:        artifact.GetRegion(),
		Environment:   artifact.GetEnvironment(),
		ContentFormat: artifact.GetContentFormat(),
		Description:   artifact.GetDescription(),
		DocLink:       artifact.GetDocumentationLink(),
	})

	return table.Print()
}
