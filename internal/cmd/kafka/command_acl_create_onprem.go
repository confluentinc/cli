package kafka

import (
	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	"github.com/spf13/cobra"

	aclutil "github.com/confluentinc/cli/internal/pkg/acl"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/kafkarest"
)

func (c *aclCommand) newCreateCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a Kafka ACL.",
		Args:  cobra.NoArgs,
		RunE:  c.createOnPrem,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "You can specify only one of the following flags per command invocation: `--cluster-scope`, `--consumer-group`, `--topic`, or `--transactional-id`. For example, for a consumer to read a topic, you need to grant \"READ\" and \"DESCRIBE\" both on the `--consumer-group` and the `--topic` resources, issuing two separate commands:",
				Code: "confluent kafka acl create --allow --principal User:Jane --operation READ --operation DESCRIBE --consumer-group java_example_group_1",
			},
			examples.Example{
				Code: "confluent kafka acl create --allow --principal User:Jane --operation READ --operation DESCRIBE --topic '*'",
			},
		),
	}

	cmd.Flags().AddFlagSet(aclutil.AclFlags())
	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	_ = cmd.MarkFlagRequired("principal")
	_ = cmd.MarkFlagRequired("operation")

	return cmd
}

func (c *aclCommand) createOnPrem(cmd *cobra.Command, _ []string) error {
	acl := aclutil.ParseAclRequest(cmd)
	acl = aclutil.ValidateCreateDeleteAclRequestData(acl)
	if acl.Errors != nil {
		return acl.Errors
	}

	restClient, restContext, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}

	clusterId, err := getClusterIdForRestRequests(restClient, restContext)
	if err != nil {
		return err
	}

	opts := aclutil.AclRequestToCreateAclRequest(acl)
	httpResp, err := restClient.ACLV3Api.CreateKafkaAcls(restContext, clusterId, opts)
	if err != nil {
		return kafkarest.NewError(restClient.GetConfig().BasePath, err, httpResp)
	}

	aclData := aclutil.CreateAclRequestDataToAclData(acl)
	return aclutil.PrintACLsFromKafkaRestResponse(cmd, []kafkarestv3.AclData{aclData})
}
