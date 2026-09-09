package flink

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
)

type artifactOutList struct {
	Id          string `human:"ID" serialized:"id"`
	Name        string `human:"Name" serialized:"name"`
	Cloud       string `human:"Cloud" serialized:"cloud"`
	Region      string `human:"Region" serialized:"region"`
	Environment string `human:"Environment" serialized:"environment"`
}

func (c *command) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Flink UDF artifacts.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List Flink UDF artifacts.",
				Code: "confluent flink artifact list --cloud aws --region us-west-2 --environment env-123456",
			},
		),
	}

	pcmd.AddCloudFlag(cmd)
	pcmd.AddRegionFlagFlink(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("cloud"))
	cobra.CheckErr(cmd.MarkFlagRequired("region"))

	return cmd
}

func (c *command) list(cmd *cobra.Command, _ []string) error {
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

	artifacts, err := c.V2Client.ListFlinkArtifacts(cloud, region, environment)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	list.Sort(false)
	for _, artifact := range artifacts {
		list.Add(&artifactOutList{
			Name:        artifact.GetDisplayName(),
			Id:          artifact.GetId(),
			Cloud:       artifact.GetCloud(),
			Region:      artifact.GetRegion(),
			Environment: artifact.GetEnvironment(),
		})
	}
	return list.Print()
}
