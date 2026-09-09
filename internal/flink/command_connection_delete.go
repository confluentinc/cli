package flink

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/deletion"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/resource"
)

func (c *command) newConnectionDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <name-1> [name-2] ... [name-n]",
		Short:             "Delete one or more Flink connections.",
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validConnectionArgsMultiple),
		RunE:              c.connectionDelete,
	}

	pcmd.AddCloudFlag(cmd)
	pcmd.AddRegionFlagFlink(cmd, c.AuthenticatedCLICommand)
	pcmd.AddForceFlag(cmd)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)

	cobra.CheckErr(cmd.MarkFlagRequired("cloud"))
	cobra.CheckErr(cmd.MarkFlagRequired("region"))

	return cmd
}

func (c *command) connectionDelete(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	if _, err := c.V2Client.GetOrgEnvironment(environmentId); err != nil {
		return errors.NewErrorWithSuggestions(err.Error(), fmt.Sprintf(envNotFoundErrorMsg, environmentId))
	}

	client, err := c.GetFlinkGatewayClient(false)
	if err != nil {
		return err
	}

	existenceFunc := func(id string) bool {
		_, err := client.GetConnection(environmentId, id, c.Context.GetCurrentOrganization())
		return err == nil
	}

	if err := deletion.ValidateAndConfirm(cmd, args, existenceFunc, resource.FlinkConnection); err != nil {
		return err
	}

	deleteFunc := func(id string) error {
		return client.DeleteConnection(environmentId, id, c.Context.GetCurrentOrganization())
	}

	_, err = deletion.Delete(cmd, args, deleteFunc, resource.FlinkConnection)
	return err
}
