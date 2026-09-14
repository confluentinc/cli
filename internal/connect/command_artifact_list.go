package connect

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *artifactCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Connect artifacts.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List Connect artifacts.",
				Code: "confluent connect artifact list --cloud aws --environment env-abc123",
			},
		),
	}

	pcmd.AddCloudFlag(cmd)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("cloud"))

	return cmd
}

func (c *artifactCommand) list(cmd *cobra.Command, _ []string) error {
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

	artifacts, err := c.V2Client.ListConnectArtifacts(cloud, environment)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	list.Sort(false)
	for _, artifact := range artifacts {
		list.Add(convertToArtifactOut(artifact))
	}

	list.Sort(true)
	return list.Print()
}
