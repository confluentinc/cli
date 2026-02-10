package flink

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/deletion"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/resource"
)

func (c *command) newMaterializedTableDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <name-1> [name-2] ... [name-n]",
		Short:             "Delete one or more materialized tables.",
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validConnectionArgsMultiple),
		RunE:              c.materializedTableDelete,
	}

	cmd.Flags().String("database", "", "The ID of Kafka cluster hosting the Materialized Table's topic.")
	pcmd.AddCloudFlag(cmd)
	pcmd.AddRegionFlagFlink(cmd, c.AuthenticatedCLICommand)
	pcmd.AddForceFlag(cmd)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)

	cobra.CheckErr(cmd.MarkFlagRequired("database"))

	return cmd
}

func (c *command) materializedTableDelete(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	if _, err := c.V2Client.GetOrgEnvironment(environmentId); err != nil {
		return errors.NewErrorWithSuggestions(err.Error(), fmt.Sprintf(envNotFoundErrorMsg, environmentId))
	}

	client, err := c.GetFlinkGatewayClientInternal(false)
	if err != nil {
		return err
	}

	kafkaId, err := cmd.Flags().GetString("database")
	if err != nil {
		return err
	}

	existenceFunc := func(id string) bool {
		_, err := client.DescribeMaterializedTable(environmentId, id, c.Context.GetCurrentOrganization(), kafkaId)
		return err == nil
	}

	if err := deletion.ValidateAndConfirm(cmd, args, existenceFunc, resource.MaterializedTable); err != nil {
		return err
	}

	deleteFunc := func(id string) error {
		return client.DeleteMaterializedTable(environmentId, c.Context.GetCurrentOrganization(), kafkaId, id)
	}

	_, err = deletion.Delete(cmd, args, deleteFunc, resource.MaterializedTable)
	return err
}
