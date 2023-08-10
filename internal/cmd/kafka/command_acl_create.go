package kafka

import (
	"fmt"

	"github.com/spf13/cobra"

	pacl "github.com/confluentinc/cli/internal/pkg/acl"
	"github.com/confluentinc/cli/internal/pkg/ccstructs"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
)

func (c *aclCommand) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a Kafka ACL.",
		Args:  cobra.NoArgs,
		RunE:  c.create,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "You can specify only one of the following flags per command invocation: `--cluster-scope`, `--consumer-group`, `--topic`, or `--transactional-id`. For example, for a consumer to read a topic, you need to grant \"read\" and \"describe\" both on the `--consumer-group` and the `--topic` resources, issuing two separate commands:",
				Code: "confluent kafka acl create --allow --service-account sa-55555 --operations read,describe --consumer-group java_example_group_1",
			},
			examples.Example{
				Code: `confluent kafka acl create --allow --service-account sa-55555 --operations read,describe --topic "*"`,
			},
		),
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
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("operations"))

	cmd.MarkFlagsMutuallyExclusive("service-account", "principal")

	return cmd
}

func (c *aclCommand) create(cmd *cobra.Command, _ []string) error {
	acls, err := parse(cmd)
	if err != nil {
		return err
	}

	bindings := make([]*ccstructs.ACLBinding, len(acls))
	for i, acl := range acls {
		validateAddAndDelete(acl)
		if acl.errors != nil {
			return acl.errors
		}
		bindings[i] = acl.ACLBinding
	}

	kafkaREST, err := c.GetKafkaREST()
	if err != nil {
		return err
	}

	if err := c.provisioningClusterCheck(kafkaREST.GetClusterId()); err != nil {
		return err
	}

	users, err := c.getAllUsers()
	if err != nil {
		return err
	}

	numericIdToResourceId := mapNumericIdToResourceId(users)

	for i, binding := range bindings {
		data := pacl.GetCreateAclRequestData(binding)
		if err := kafkaREST.CloudClient.CreateKafkaAcls(data); err != nil {
			if i > 0 {
				_ = pacl.PrintACLsWithResourceIdMap(cmd, bindings[:i], numericIdToResourceId)
			}
			return err
		}
	}

	return pacl.PrintACLsWithResourceIdMap(cmd, bindings, numericIdToResourceId)
}
