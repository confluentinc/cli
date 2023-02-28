package kafka

import (
	"fmt"

	"github.com/spf13/cobra"

	pacl "github.com/confluentinc/cli/internal/pkg/acl"
	"github.com/confluentinc/cli/internal/pkg/ccstructs"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/kafkarest"
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

	_ = cmd.MarkFlagRequired("operations")

	return cmd
}

func (c *aclCommand) create(cmd *cobra.Command, _ []string) error {
	acls, err := parse(cmd)
	if err != nil {
		return err
	}

	users, userIdMap, err := c.mapResourceIdToUserId(nil)
	if err != nil {
		return err
	}

	if err := c.aclResourceIdToNumericId(acls, userIdMap); err != nil {
		return err
	}

	_, resourceIdMap, err := c.mapUserIdToResourceId(users)
	if err != nil {
		return err
	}

	var bindings []*ccstructs.ACLBinding
	for _, acl := range acls {
		validateAddAndDelete(acl)
		if acl.errors != nil {
			return acl.errors
		}
		bindings = append(bindings, acl.ACLBinding)
	}

	kafkaClusterConfig, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	if err := c.provisioningClusterCheck(cmd, kafkaClusterConfig.ID); err != nil {
		return err
	}

	kafkaREST, err := c.GetKafkaREST()
	if err != nil {
		return err
	}

	for i, binding := range bindings {
		data := pacl.GetCreateAclRequestData(binding)
		if httpResp, err := kafkaREST.CloudClient.CreateKafkaAcls(kafkaClusterConfig.ID, data); err != nil {
			if i > 0 {
				_ = pacl.PrintACLsWithResourceIdMap(cmd, bindings[:i], resourceIdMap)
			}
			return kafkarest.NewError(kafkaREST.CloudClient.GetUrl(), err, httpResp)
		}
	}

	return pacl.PrintACLsWithResourceIdMap(cmd, bindings, resourceIdMap)
}
