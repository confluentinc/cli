package connect

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
)

type artifactOutList struct {
	Id            string `human:"ID" serialized:"id"`
	Name          string `human:"Name" serialized:"name"`
	Description   string `human:"Description" serialized:"description"`
	Cloud         string `human:"Cloud" serialized:"cloud"`
	Region        string `human:"Region" serialized:"region"`
	Environment   string `human:"Environment" serialized:"environment"`
	ContentFormat string `human:"Content Format" serialized:"content_format"`
}

func (c *artifactCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List connect artifacts.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List connect artifacts.",
				Code: "confluent connect artifact list --cloud aws --region us-west-2 --environment env-abc123",
			},
		),
	}

	pcmd.AddCloudFlag(cmd)
	//TODO: see if we can autocomplete similar to pcmd.AddRegionFlagFlink(cmd, c.AuthenticatedCLICommand)
	cmd.Flags().String("region", "", `Cloud region for connect artifact.`)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("cloud"))
	cobra.CheckErr(cmd.MarkFlagRequired("region"))

	return cmd
}

func (c *artifactCommand) list(cmd *cobra.Command, _ []string) error {

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

	artifacts, err := c.V2Client.ListConnectArtifacts(cloud, region, environment)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	list.Sort(false)
	for _, artifact := range artifacts {
		list.Add(&artifactOutList{
			Id:            artifact.GetId(),
			Name:          artifact.Spec.GetDisplayName(),
			Description:   artifact.Spec.GetDescription(),
			Cloud:         artifact.Spec.GetCloud(),
			Region:        artifact.Spec.GetRegion(),
			Environment:   artifact.Spec.GetEnvironment(),
			ContentFormat: artifact.Spec.GetContentFormat(),
		})
	}

	return list.Print()
}
