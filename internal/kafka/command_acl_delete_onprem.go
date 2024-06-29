package kafka

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v3/pkg/acl"
	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/deletion"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/kafkarest"
	"github.com/confluentinc/cli/v3/pkg/resource"
)

func (c *aclCommand) newDeleteCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete Kafka ACLs matching the search criteria.",
		Args:  cobra.NoArgs,
		RunE:  c.deleteOnPrem,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Delete all "read" access ACLs for the specified user (providing embedded Kafka REST Proxy endpoint).`,
				Code: `confluent kafka acl delete --url http://localhost:8090/kafka --operation read --allow --topic Test --principal User:Jane --host "*"`,
			},
			examples.Example{
				Text: `Delete all "read" access ACLs for the specified user (providing Kafka REST Proxy endpoint).`,
				Code: `confluent kafka acl delete --url http://localhost:8082 --operation read --allow --topic Test --principal User:Jane --host "*"`,
			},
			examples.Example{
				Text: `Delete all "read" access ACLs for the specified user.`,
				Code: `confluent kafka acl delete --operation read --allow --topic Test --principal User:Jane --host "*"`,
			},
		),
	}

	cmd.Flags().AddFlagSet(acl.Flags())
	pcmd.AddForceFlag(cmd)
	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("principal"))
	cobra.CheckErr(cmd.MarkFlagRequired("operation"))
	cobra.CheckErr(cmd.MarkFlagRequired("host"))

	return cmd
}

func (c *aclCommand) deleteOnPrem(cmd *cobra.Command, _ []string) error {
	data := acl.ValidateCreateDeleteAclRequestData(acl.ParseRequest(cmd))
	if data.Errors != nil {
		return data.Errors
	}

	restClient, restContext, clusterId, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}

	optsForList := acl.RequestToListRequest(data)
	aclDataList, httpResp, err := restClient.ACLV3Api.GetKafkaAcls(restContext, clusterId, optsForList)
	if err != nil {
		return kafkarest.NewError(restClient.GetConfig().BasePath, err, httpResp)
	}
	if len(aclDataList.Data) == 0 {
		return errors.NewErrorWithSuggestions("ACL matching these parameters not found", validACLSuggestion)
	}

	promptMsg := fmt.Sprintf(acl.DeleteACLConfirmMsg, resource.ACL)
	if err := deletion.ConfirmPrompt(cmd, promptMsg); err != nil {
		return err
	}

	opts := acl.RequestToDeleteRequest(data)
	aclDeleteResp, httpResp, err := restClient.ACLV3Api.DeleteKafkaAcls(restContext, clusterId, opts)
	if err != nil {
		return kafkarest.NewError(restClient.GetConfig().BasePath, err, httpResp)
	}

	return acl.PrintACLsFromKafkaRestResponseOnPrem(cmd, aclDeleteResp.Data)
}
