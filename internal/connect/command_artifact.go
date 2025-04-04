package connect

import (
	"github.com/spf13/cobra"

	camv1 "github.com/confluentinc/ccloud-sdk-go-v2/cam/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/output"
)

type artifactCommand struct {
	*pcmd.AuthenticatedCLICommand
}

type artifactOut struct {
	// TODO: double check all the fields
	Id string `human:"ID" serialized:"id"`
	//ArtifactFile  string `human:"Artifact File" serialized:"artifact_file"`
	Name          string `human:"Name" serialized:"name"`
	Description   string `human:"Description" serialized:"description"`
	Cloud         string `human:"Cloud" serialized:"cloud"`
	Region        string `human:"Region" serialized:"region"`
	Environment   string `human:"Environment" serialized:"environment"`
	ContentFormat string `human:"Content Format" serialized:"content_format"`
	//UploadSource  string `human:"Upload Source" serialized:"upload_source"`
	//Plugins       string `human:"Plugins" serialized:"plugins"`
	//Usages        string `human:"Usages" serialized:"usages"`
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

func printArtifactTable(cmd *cobra.Command, artifact camv1.CamV1ConnectArtifact) error {
	table := output.NewTable(cmd)

	table.Add(&artifactOut{
		// TODO: double check all the fields
		Id: artifact.GetId(),
		//ArtifactFile:  artifact.GetArtifactFile(),
		Name:          artifact.Spec.GetDisplayName(),
		Description:   artifact.Spec.GetDescription(),
		Cloud:         artifact.Spec.GetCloud(),
		Region:        artifact.Spec.GetRegion(),
		Environment:   artifact.Spec.GetEnvironment(),
		ContentFormat: artifact.Spec.GetContentFormat(),
		//UploadSource:  artifact.Spec.GetUploadSource(),
		//Plugins: artifact.Spec.GetPlugins(),
		//Usages:  artifact.Spec.GetUsages(),
	})

	return table.Print()
}
