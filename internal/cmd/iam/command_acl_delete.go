package iam

import (
	"fmt"

	"github.com/spf13/cobra"

	pacl "github.com/confluentinc/cli/internal/pkg/acl"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

func (c *aclCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a centralized ACL.",
		Args:  cobra.NoArgs,
		RunE:  c.delete,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Delete an ACL that granted the specified user access to the "test" topic in the specified cluster.`,
				Code: `confluent iam acl delete --kafka-cluster <kafka-cluster-id> --allow --principal User:Jane --topic test --operation write --host "*"`,
			},
		),
	}

	cmd.Flags().AddFlagSet(aclFlags())
	pcmd.AddForceFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)

	cobra.CheckErr(cmd.MarkFlagRequired("kafka-cluster"))
	cobra.CheckErr(cmd.MarkFlagRequired("principal"))
	cobra.CheckErr(cmd.MarkFlagRequired("operation"))
	cobra.CheckErr(cmd.MarkFlagRequired("host"))

	return cmd
}

func (c *aclCommand) delete(cmd *cobra.Command, _ []string) error {
	acl := parse(cmd)
	if acl.errors != nil {
		return acl.errors
	}

	bindings, response, err := c.MDSClient.KafkaACLManagementApi.SearchAclBinding(c.createContext(), convertToACLFilterRequest(acl.CreateAclRequest))
	if err != nil {
		return c.handleACLError(cmd, err, response)
	}

	promptMsg := fmt.Sprintf(pacl.DeleteACLConfirmMsg, resource.ACL)
	if len(bindings) > 1 {
		promptMsg = fmt.Sprintf(pacl.DeleteACLConfirmMsg, resource.Plural(resource.ACL))
	}
	if ok, err := form.ConfirmDeletionYesNo(cmd, promptMsg); err != nil || !ok {
		return err
	}

	bindings, response, err = c.MDSClient.KafkaACLManagementApi.RemoveAclBindings(c.createContext(), convertToACLFilterRequest(acl.CreateAclRequest))
	if err != nil {
		return c.handleACLError(cmd, err, response)
	}

	return printACLs(cmd, acl.Scope.Clusters.KafkaCluster, bindings)
}
