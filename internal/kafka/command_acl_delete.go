package kafka

import (
	"fmt"

	"github.com/spf13/cobra"

	pacl "github.com/confluentinc/cli/v3/pkg/acl"
	"github.com/confluentinc/cli/v3/pkg/ccstructs"
	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/deletion"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/resource"
)

const validACLSuggestion = "To check for valid ACLs, use `confluent kafka acl list`"

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

	cobra.CheckErr(cmd.MarkFlagRequired("operations"))

	cmd.MarkFlagsMutuallyExclusive("service-account", "principal")

	return cmd
}

func (c *aclCommand) delete(cmd *cobra.Command, _ []string) error {
	acls, err := parse(c.Context, cmd)
	if err != nil {
		return err
	}

	filters := make([]*ccstructs.ACLFilter, len(acls))
	for i, acl := range acls {
		validateAddAndDelete(acl)
		if acl.errors != nil {
			return acl.errors
		}
		filters[i] = convertToFilter(acl.ACLBinding)
	}

	kafkaREST, err := c.GetKafkaREST()
	if err != nil {
		return err
	}

	if err := c.provisioningClusterCheck(kafkaREST.GetClusterId()); err != nil {
		return err
	}

	count := 0
	for _, acl := range acls {
		aclDataList, err := kafkaREST.CloudClient.GetKafkaAcls(acl.ACLBinding)
		if err != nil {
			return err
		}
		if len(aclDataList.Data) == 0 {
			return errors.NewErrorWithSuggestions("one or more ACLs matching these parameters not found", validACLSuggestion)
		}
		count += len(aclDataList.Data)
	}

	promptMsg := fmt.Sprintf(pacl.DeleteACLConfirmMsg, resource.ACL)
	if count > 1 {
		promptMsg = fmt.Sprintf(pacl.DeleteACLConfirmMsg, resource.Plural(resource.ACL))
	}
	if err := deletion.ConfirmDeletionYesNo(cmd, promptMsg); err != nil {
		return err
	}

	count = 0
	for i, filter := range filters {
		deleteResp, err := kafkaREST.CloudClient.DeleteKafkaAcls(filter)
		if err != nil {
			if i > 0 {
				output.ErrPrintln(c.Config.EnableColor, printAclsDeleted(count))
			}
			return err
		}

		count += len(deleteResp.Data)
	}

	output.ErrPrintln(c.Config.EnableColor, printAclsDeleted(count))
	return nil
}

func printAclsDeleted(count int) string {
	switch count {
	case 0:
		return "ACL not found. ACL may have been misspelled or already deleted."
	case 1:
		return "Deleted 1 ACL."
	default:
		return fmt.Sprintf("Deleted %d ACLs.", count)
	}
}
