package connect

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/deletion"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/resource"
)

var extraWarning = "\nThis action is irreversible and is going to break all custom SMTs using this Artifact!\nDo you still want to proceed?"

func (c *artifactCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id-1> [id-2] ... [id-n]",
		Short: "Delete one or more connect artifacts.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  c.delete,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Delete connect artifact.",
				Code: "confluent connect artifact delete --cloud aws --region us-west-2 --environment env-abc123 cfa-abc123",
			},
		),
	}

	pcmd.AddCloudFlag(cmd)
	cmd.Flags().String("region", "", `Cloud region for connect artifact, ex. "us-west-2".`)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddForceFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("cloud"))
	cobra.CheckErr(cmd.MarkFlagRequired("region"))

	return cmd
}

func (c *artifactCommand) delete(cmd *cobra.Command, args []string) error {
	cloud, err := cmd.Flags().GetString("cloud")
	if err != nil {
		return err
	}

	region, err := cmd.Flags().GetString("region")
	if err != nil {
		return err
	}

	environment, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}
	if _, err = c.V2Client.GetOrgEnvironment(environment); err != nil {
		return fmt.Errorf("environment '%s' not found", environment)
	}

	existenceFunc := func(id string) bool {
		artifactIdToName, err := c.mapArtifactIdToName(cloud, region, environment)
		if err != nil {
			return false
		}
		_, ok := artifactIdToName[id]
		return ok
	}

	if err := deletion.ValidateAndConfirmWithExtraWarning(cmd, args, existenceFunc, resource.ConnectArtifact, extraWarning); err != nil {
		return err
	}

	deleteFunc := func(id string) error {
		return c.V2Client.DeleteConnectArtifact(cloud, region, environment, id)
	}

	_, err = deletion.Delete(args, deleteFunc, resource.ConnectArtifact)

	return err
}

func (c *artifactCommand) mapArtifactIdToName(cloud string, region string, environment string) (map[string]string, error) {

	artifacts, err := c.V2Client.ListConnectArtifacts(cloud, region, environment)
	if err != nil {
		return nil, err
	}

	artifactIdToName := make(map[string]string)
	for _, artifact := range artifacts {
		artifactIdToName[artifact.GetId()] = artifact.Spec.GetDisplayName()
	}

	return artifactIdToName, nil
}
