package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
)

func (c *command) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <id>",
		Short: "Describe a Flink UDF artifact.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.describe,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Describe Flink UDF artifact.",
				Code: "confluent flink artifact describe --cloud aws --region us-west-2 cfa-123456",
			},
		),
	}

	pcmd.AddCloudFlag(cmd)
	pcmd.AddRegionFlagFlink(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("cloud"))
	cobra.CheckErr(cmd.MarkFlagRequired("region"))

	return cmd
}

func (c *command) describe(cmd *cobra.Command, args []string) error {
	cloud, err := cmd.Flags().GetString("cloud")
	if err != nil {
		return err
	}

	region, err := cmd.Flags().GetString("region")
	if err != nil {
		return err
	}

	artifact, err := c.V2Client.DescribeFlinkArtifact(cloud, region, args[0])
	if err != nil {
		return err
	}

	return printTable(cmd, artifact)
}
