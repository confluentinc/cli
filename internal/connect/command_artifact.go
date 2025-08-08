package connect

import (
	"github.com/spf13/cobra"

	camv1 "github.com/confluentinc/ccloud-sdk-go-v2/cam/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

type artifactCommand struct {
	*pcmd.AuthenticatedCLICommand
}

type artifactOut struct {
	Id            string `human:"ID" serialized:"id"`
	Name          string `human:"Name" serialized:"name"`
	Description   string `human:"Description" serialized:"description"`
	Cloud         string `human:"Cloud" serialized:"cloud"`
	Environment   string `human:"Environment" serialized:"environment"`
	ContentFormat string `human:"Content Format" serialized:"content_format"`
	Status        string `human:"Status" serialized:"status"`
}

func newArtifactCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "artifact",
		Short:       "Manage Connect artifacts.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	c := &artifactCommand{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	cmd.AddCommand(c.newCreateCommand())
	cmd.AddCommand(c.newDeleteCommand())
	cmd.AddCommand(c.newDescribeCommand())
	cmd.AddCommand(c.newListCommand())

	return cmd
}

func convertToArtifactOut(artifact camv1.CamV1ConnectArtifact) *artifactOut {
	return &artifactOut{
		Id:            artifact.GetId(),
		Name:          artifact.Spec.GetDisplayName(),
		Description:   artifact.Spec.GetDescription(),
		Cloud:         artifact.Spec.GetCloud(),
		Environment:   artifact.Spec.GetEnvironment(),
		ContentFormat: artifact.Spec.GetContentFormat(),
		Status:        artifact.Status.GetPhase(),
	}
}

func printArtifactTable(cmd *cobra.Command, artifact camv1.CamV1ConnectArtifact) error {
	table := output.NewTable(cmd)

	table.Add(convertToArtifactOut(artifact))

	return table.Print()
}
