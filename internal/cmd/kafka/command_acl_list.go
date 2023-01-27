package kafka

import (
	"github.com/spf13/cobra"

	aclutil "github.com/confluentinc/cli/internal/pkg/acl"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/kafkarest"
)

func (c *aclCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Kafka ACLs for a resource.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}

	cmd.Flags().AddFlagSet(resourceFlags())
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddServiceAccountFlag(cmd, c.AuthenticatedCLICommand)
	cmd.Flags().String("principal", "", `Principal for this operation, prefixed with "User:".`)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *aclCommand) list(cmd *cobra.Command, _ []string) error {
	acl, err := parse(cmd)
	if err != nil {
		return err
	}

	if acl[0].errors != nil {
		return acl[0].errors
	}

	resourceIdMap, err := c.mapUserIdToResourceId()
	if err != nil {
		return err
	}

	kafkaClusterConfig, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	if err := c.provisioningClusterCheck(kafkaClusterConfig.ID); err != nil {
		return err
	}

	kafkaREST, err := c.GetKafkaREST()
	if err != nil {
		return err
	}

	aclDataList, httpResp, err := kafkaREST.CloudClient.GetKafkaAcls(kafkaClusterConfig.ID, acl[0].ACLBinding)
	if err != nil {
		return kafkarest.NewError(kafkaREST.CloudClient.GetUrl(), err, httpResp)
	}

	return aclutil.PrintACLsFromKafkaRestResponseWithResourceIdMap(cmd, aclDataList.Data, resourceIdMap)
}
