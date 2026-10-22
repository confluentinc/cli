package connect

import (
	"fmt"

	"github.com/spf13/cobra"

	camv1 "github.com/confluentinc/ccloud-sdk-go-v2/cam/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
)

func (c *artifactCommand) newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a Connect artifact.",
		Long:  "Update the name or description of a Connect artifact. All other fields are immutable.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.update,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Update the name and description of Connect artifact "cfa-abc123".`,
				Code: "confluent connect artifact update cfa-abc123 --name my-new-name --description \"My new description\" --cloud aws --environment env-abc123",
			},
		),
	}

	cmd.Flags().String("name", "", "Specify the new Connect artifact name.")
	cmd.Flags().String("description", "", "Specify the new Connect artifact description.")
	pcmd.AddCloudFlag(cmd)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("cloud"))
	cmd.MarkFlagsOneRequired("name", "description")

	return cmd
}

func (c *artifactCommand) update(cmd *cobra.Command, args []string) error {
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

	spec := &camv1.CamV1ConnectArtifactSpecUpdate{}
	if cmd.Flags().Changed("name") {
		name, err := cmd.Flags().GetString("name")
		if err != nil {
			return err
		}
		spec.SetDisplayName(name)
	}
	if cmd.Flags().Changed("description") {
		description, err := cmd.Flags().GetString("description")
		if err != nil {
			return err
		}
		spec.SetDescription(description)
	}

	artifact, err := c.V2Client.UpdateConnectArtifact(cloud, environment, args[0], camv1.CamV1ConnectArtifactUpdate{Spec: spec})
	if err != nil {
		return err
	}

	return printArtifactTable(cmd, artifact)
}
