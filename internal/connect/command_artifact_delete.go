package connect

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/deletion"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/resource"
)

func (c *artifactCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id-1> [id-2] ... [id-n]",
		Short: "Delete one or more Connect artifacts.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  c.delete,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Delete Connect artifact.",
				Code: "confluent connect artifact delete cfa-abc123 --cloud aws --environment env-abc123",
			},
		),
	}

	pcmd.AddCloudAwsAzureFlag(cmd)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddForceFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("cloud"))

	return cmd
}

func (c *artifactCommand) delete(cmd *cobra.Command, args []string) error {
	cloud, err := cmd.Flags().GetString("cloud")
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
		artifactIdToName, err := c.mapArtifactIdToName(cloud, environment)
		if err != nil {
			return false
		}
		_, ok := artifactIdToName[id]
		return ok
	}

	if err := deletion.ValidateAndConfirm(cmd, args, existenceFunc, resource.ConnectArtifact); err != nil {
		return err
	}

	deleteFunc := func(id string) error {
		return c.V2Client.DeleteConnectArtifact(cloud, environment, id)
	}

	_, err = deletion.Delete(cmd, args, deleteFunc, resource.ConnectArtifact)

	return err
}

func (c *artifactCommand) mapArtifactIdToName(cloud string, environment string) (map[string]string, error) {
	artifacts, err := c.V2Client.ListConnectArtifacts(cloud, environment)
	if err != nil {
		return nil, err
	}

	artifactIdToName := make(map[string]string)
	for _, artifact := range artifacts {
		artifactIdToName[artifact.GetId()] = artifact.Spec.GetDisplayName()
	}

	return artifactIdToName, nil
}
