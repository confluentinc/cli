package kafka

import (
	"context"
	"fmt"
	"github.com/confluentinc/cli/internal/pkg/dynamic-config"
	"net/http"
	"os"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/spf13/cobra"

	aclutil "github.com/confluentinc/cli/internal/pkg/acl"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
)

func (c *aclCommand) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a Kafka ACL.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.create),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "You can specify only one of the following flags per command invocation: `cluster`, `consumer-group`, `topic`, or `transactional-id`. For example, for a consumer to read a topic, you need to grant `READ` and `DESCRIBE` both on the `consumer-group` and the `topic` resources, issuing two separate commands:",
				Code: "confluent kafka acl create --allow --service-account sa-55555 --operation READ --operation DESCRIBE --consumer-group java_example_group_1",
			},
			examples.Example{
				Code: "confluent kafka acl create --allow --service-account sa-55555 --operation READ --operation DESCRIBE --topic '*'",
			},
		),
	}

	cmd.Flags().AddFlagSet(aclConfigFlags())
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *aclCommand) create(cmd *cobra.Command, _ []string) error {
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

	resourceIdMap, err := c.mapUserIdToResourceId()
	if err != nil {
		return err
	}

	var bindings []*schedv1.ACLBinding
	for _, acl := range acls {
		validateAddAndDelete(acl)
		if acl.errors != nil {
			return acl.errors
		}
		bindings = append(bindings, acl.ACLBinding)
	}

	kafkaREST, _ := c.GetKafkaREST()
	if kafkaREST != nil {
		kafkaClusterConfig, err := c.AuthenticatedCLICommand.Context.GetKafkaClusterForCommand()
		if err != nil {
			return err
		}
		lkc := kafkaClusterConfig.ID

		kafkaRestExists := true
		for i, binding := range bindings {
			opts := aclBindingToClustersClusterIdAclsPostOpts(binding)
			httpResp, err := kafkaREST.Client.ACLV3Api.CreateKafkaAcls(kafkaREST.Context, lkc, &opts)

			if err != nil && httpResp == nil {
				if i == 0 {
					// assume Kafka REST is not available, fallback to KafkaAPI
					kafkaRestExists = false
					break
				}
				// i > 0: unlikely
				_ = aclutil.PrintACLsWithResourceIdMap(cmd, bindings[:i], os.Stdout, resourceIdMap)
				return kafkaRestError(kafkaREST.Client.GetConfig().BasePath, err, httpResp)
			}

			if err != nil {
				if i > 0 {
					// unlikely
					_ = aclutil.PrintACLsWithResourceIdMap(cmd, bindings[:i], os.Stdout, resourceIdMap)
				}
				return kafkaRestError(kafkaREST.Client.GetConfig().BasePath, err, httpResp)
			}

			if httpResp != nil && httpResp.StatusCode != http.StatusCreated {
				if i > 0 {
					_ = aclutil.PrintACLsWithResourceIdMap(cmd, bindings[:i], os.Stdout, resourceIdMap)
				}
				return errors.NewErrorWithSuggestions(
					fmt.Sprintf(errors.KafkaRestUnexpectedStatusMsg, httpResp.Request.URL, httpResp.StatusCode),
					errors.InternalServerErrorSuggestions)
			}
		}

		if kafkaRestExists {
			return aclutil.PrintACLsWithResourceIdMap(cmd, bindings, os.Stdout, resourceIdMap)
		}
	}

	// Kafka REST is not available, fallback to KafkaAPI
	cluster, err := dynamic_config.KafkaCluster(c.Context)
	if err != nil {
		return err
	}

	if err := c.Client.Kafka.CreateACLs(context.Background(), cluster, bindings); err != nil {
		return err
	}

	return aclutil.PrintACLsWithResourceIdMap(cmd, bindings, os.Stdout, resourceIdMap)
}
