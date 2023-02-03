package kafka

import (
	"github.com/spf13/cobra"

	aclutil "github.com/confluentinc/cli/internal/pkg/acl"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/kafkarest"
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

	cmd.Flags().AddFlagSet(aclutil.AclFlags())
	pcmd.AddForceFlag(cmd)
	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	_ = cmd.MarkFlagRequired("principal")
	_ = cmd.MarkFlagRequired("operation")
	_ = cmd.MarkFlagRequired("host")

	return cmd
}

func (c *aclCommand) deleteOnPrem(cmd *cobra.Command, _ []string) error {
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

	optsForList := aclutil.AclRequestToListAclRequest(acl)
	aclDataList, httpResp, err := restClient.ACLV3Api.GetKafkaAcls(restContext, clusterId, optsForList)
	if err != nil {
		return kafkarest.NewError(restClient.GetConfig().BasePath, err, httpResp)
	}
	if len(aclDataList.Data) == 0 {
		return errors.NewErrorWithSuggestions("ACL matching these parameters not found", ValidACLSuggestion)
	}

	promptMsg := errors.DeleteACLConfirmMsg
	if ok, err := form.ConfirmDeletion(cmd, promptMsg, ""); err != nil || !ok {
		return err
	}

	opts := aclutil.AclRequestToDeleteAclRequest(acl)
	aclDeleteResp, httpResp, err := restClient.ACLV3Api.DeleteKafkaAcls(restContext, clusterId, opts)
	if err != nil {
		return kafkarest.NewError(restClient.GetConfig().BasePath, err, httpResp)
	}

	return aclutil.PrintACLsFromKafkaRestResponse(cmd, aclDeleteResp.Data)
}
