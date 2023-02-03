package kafka

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/ccstructs"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/kafkarest"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

const ValidACLSuggestion = "To check for valid ACLs, use `confluent kafka acl list`"

func (c *aclCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a Kafka ACL.",
		Args:  cobra.NoArgs,
		RunE:  c.delete,
	}

	cmd.Flags().StringSlice("operations", []string{""}, fmt.Sprintf("A comma-separated list of ACL operations: (%s).", listEnum(ccstructs.ACLOperations_ACLOperation_name, []string{"ANY", "UNKNOWN"})))
	cmd.Flags().String("principal", "", `Principal for this operation, prefixed with "User:".`)
	cmd.Flags().String("service-account", "", "The service account ID.")
	cmd.Flags().Bool("allow", false, "Access to the resource is allowed.")
	cmd.Flags().Bool("deny", false, "Access to the resource is denied.")
	cmd.Flags().Bool("cluster-scope", false, "Modify ACLs for the cluster.")
	cmd.Flags().String("topic", "", "Modify ACLs for the specified topic resource.")
	cmd.Flags().String("consumer-group", "", "Modify ACLs for the specified consumer group resource.")
	cmd.Flags().String("transactional-id", "", "Modify ACLs for the specified TransactionalID resource.")
	cmd.Flags().Bool("prefix", false, "When this flag is set, the specified resource name is interpreted as a prefix.")
	pcmd.AddForceFlag(cmd)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	_ = cmd.MarkFlagRequired("operations")

	return cmd
}

func (c *aclCommand) delete(cmd *cobra.Command, _ []string) error {
	acls, err := parse(cmd)
	if err != nil {
		return err
	}

	userIdMap, err := c.mapResourceIdToUserId()
	if err != nil {
		return err
	}

	if err := c.aclResourceIdToNumericId(acls, userIdMap); err != nil {
		return err
	}

	var filters []*ccstructs.ACLFilter
	for _, acl := range acls {
		validateAddAndDelete(acl)
		if acl.errors != nil {
			return acl.errors
		}
		filters = append(filters, convertToFilter(acl.ACLBinding))
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

	count := 0
	for _, acl := range acls {
		aclDataList, httpResp, err := kafkaREST.CloudClient.GetKafkaAcls(kafkaClusterConfig.ID, acl.ACLBinding)
		if err != nil {
			return kafkarest.NewError(kafkaREST.CloudClient.GetUrl(), err, httpResp)
		}
		if len(aclDataList.Data) == 0 {
			return errors.NewErrorWithSuggestions("one or more ACLs matching these parameters not found", ValidACLSuggestion)
		}
		count += len(aclDataList.Data)
	}

	promptMsg := errors.DeleteACLConfirmMsg
	if count > 1 {
		promptMsg = errors.DeleteACLsConfirmMsg
	}
	if ok, err := form.ConfirmDeletion(cmd, promptMsg, ""); err != nil || !ok {
		return err
	}

	count = 0
	for i, filter := range filters {
		deleteResp, httpResp, err := kafkaREST.CloudClient.DeleteKafkaAcls(kafkaClusterConfig.ID, filter)
		if err != nil {
			if i > 0 {
				printAclsDeleted(cmd, count)
			}
			return kafkarest.NewError(kafkaREST.CloudClient.GetUrl(), err, httpResp)
		}

		count += len(deleteResp.Data)
	}

	printAclsDeleted(cmd, count)
	return nil
}

func printAclsDeleted(cmd *cobra.Command, count int) {
	if count == 0 {
		utils.ErrPrintf(cmd, errors.ACLsNotFoundMsg)
	} else {
		utils.ErrPrintf(cmd, fmt.Sprintf(errors.DeletedACLsCountMsg, count))
	}
}
