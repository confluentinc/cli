package flink

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/deletion"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/resource"
)

var extraWarning = "\nThis action is irreversible and is going to break all Flink statements using this Artifact!\nDo you still want to proceed?"

func (c *command) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id-1> [id-2] ... [id-n]",
		Short: "Delete one or more Flink UDF artifacts.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  c.delete,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Delete Flink UDF artifact.",
				Code: "confluent flink artifact delete --cloud aws --region us-west-2 --environment env-123456 cfa-123456",
			},
		),
	}

	pcmd.AddCloudFlag(cmd)
	pcmd.AddRegionFlagFlink(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddForceFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)

	cobra.CheckErr(cmd.MarkFlagRequired("cloud"))
	cobra.CheckErr(cmd.MarkFlagRequired("region"))

	return cmd
}

func (c *command) delete(cmd *cobra.Command, args []string) error {
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

	if err := deletion.ValidateAndConfirmWithExtraWarning(cmd, args, existenceFunc, resource.FlinkArtifact, extraWarning); err != nil {
		return err
	}

	deleteFunc := func(id string) error {
		return c.V2Client.DeleteFlinkArtifact(cloud, region, environment, id)
	}

	_, err = deletion.Delete(cmd, args, deleteFunc, resource.FlinkArtifact)
	return err
}

func (c *command) mapArtifactIdToName(cloud string, region string, environment string) (map[string]string, error) {
	artifacts, err := c.V2Client.ListFlinkArtifacts(cloud, region, environment)
	if err != nil {
		return nil, err
	}

	artifactIdToName := make(map[string]string)
	for _, artifact := range artifacts {
		artifactIdToName[artifact.GetId()] = artifact.GetDisplayName()
	}

	return artifactIdToName, nil
}
