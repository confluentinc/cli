package connect

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/output"
)

type artifactCommand struct {
	*pcmd.AuthenticatedCLICommand
}

type artifactOut struct {
	// TODO: double check all the fields
	Id           string `human:"ID" serialized:"id"`
	Name         string `human:"Name" serialized:"name"`
	Description  string `human:"Description" serialized:"description"`
	ArtifactFile string `human:"Artifact File" serialized:"artifact_file"`
	Cloud        string `human:"Cloud" serialized:"cloud"`
	Region       string `human:"Region" serialized:"region"`
	Environment  string `human:"Environment" serialized:"environment"`
}

func newArtifactCommand(cfg *config.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "artifact",
		Short: "Manage custom SMT artifacts.",
		// TODO: double check this
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	c := &artifactCommand{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	cmd.AddCommand(c.newCreateCommand())
	cmd.AddCommand(c.newDescribeCommand())
	// TODO: delete operation is out of scope?
	// cmd.AddCommand(c.newDeleteCommand())
	cmd.AddCommand(c.newListCommand())
	// cmd.AddCommand(c.newUpdateCommand())

	return cmd
}

func printTable(cmd *cobra.Command, artifact connectartifactv1.ConnectV1CustomConnectArtifact) error {
	table := output.NewTable(cmd)
	table.Add(&artifactOut{
		// TODO: double check all the fields
		Id:           artifact.GetId(),
		Name:         artifact.GetDisplayName(),
		Description:  artifact.GetDescription(),
		ArtifactFile: artifact.GetArtifactFile(),
		Cloud:        artifact.GetCloud(),
		Region:       artifact.GetRegion(),
		Environment:  artifact.GetEnvironment(),
	})
	return table.Print()
}
