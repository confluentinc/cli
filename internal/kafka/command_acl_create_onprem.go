package kafka

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	"github.com/confluentinc/cli/v4/pkg/acl"
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/kafkarest"
)

func (c *aclCommand) newCreateCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a Kafka ACL.",
		Args:  cobra.NoArgs,
		RunE:  c.createOnPrem,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "You can specify only one of the following flags per command invocation: `--cluster-scope`, `--consumer-group`, `--topic`, or `--transactional-id`. For example, for a consumer to read a topic, you need to grant \"read\" and \"describe\" both on the `--consumer-group` and the `--topic` resources, issuing two separate commands:",
				Code: "confluent kafka acl create --allow --principal User:Jane --operation read --consumer-group java_example_group_1",
			},
			examples.Example{
				Code: `confluent kafka acl create --allow --principal User:Jane --operation read --topic "*"`,
			},
			examples.Example{
				Text: "You can run the previous example without logging in if you provide the embedded Kafka REST Proxy endpoint with the `--url` flag.",
				Code: "confluent kafka acl create --url http://localhost:8090/kafka --allow --principal User:Jane --operation read --consumer-group java_example_group_1",
			},
			examples.Example{
				Code: `confluent kafka acl create --url http://localhost:8090/kafka --allow --principal User:Jane --operation read --topic "*"`,
			},
			examples.Example{
				Text: "You can also run the example above without logging in if you provide the Kafka REST proxy endpoint with the `--url` flag.",
				Code: "confluent kafka acl create --url http://localhost:8082 --allow --principal User:Jane --operation read --consumer-group java_example_group_1",
			},
			examples.Example{
				Code: `confluent kafka acl create --url http://localhost:8082 --allow --principal User:Jane --operation read --topic "*"`,
			},
		),
	}

	cmd.Flags().AddFlagSet(acl.Flags())
	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("principal"))
	cobra.CheckErr(cmd.MarkFlagRequired("operation"))

	return cmd
}

func (c *aclCommand) createOnPrem(cmd *cobra.Command, _ []string) error {
	data := acl.ValidateCreateDeleteAclRequestData(acl.ParseRequest(cmd))
	if data.Errors != nil {
		return data.Errors
	}

	restClient, restContext, clusterId, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}

	opts := acl.RequestToCreateRequest(data)
	httpResp, err := restClient.ACLV3Api.CreateKafkaAcls(restContext, clusterId, opts)
	if err != nil {
		return kafkarest.NewError(restClient.GetConfig().BasePath, err, httpResp)
	}

	aclData := acl.CreateAclRequestDataToAclData(data)
	return acl.PrintACLsFromKafkaRestResponseOnPrem(cmd, []kafkarestv3.AclData{aclData})
}
