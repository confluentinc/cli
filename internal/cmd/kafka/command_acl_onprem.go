package kafka

import (
	"fmt"
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	aclutil "github.com/confluentinc/cli/internal/pkg/acl"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/spf13/cobra"
	"net/http"
)

type aclOnPremCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
}

func NewAclCommandOnPrem(prerunner pcmd.PreRunner) *cobra.Command {
	aclCmd := &aclOnPremCommand{
		pcmd.NewAuthenticatedStateFlagCommand(
			&cobra.Command{
				Use: "acl",
				Short: "Manage Kafka ACLs with Confluent Rest Proxy.",
			}, prerunner, OnPremTopicSubcommandFlags),
	}
	aclCmd.SetPersistentPreRunE(prerunner.InitializeOnPremKafkaRest(aclCmd.AuthenticatedCLICommand))
	aclCmd.init()
	return aclCmd.Command
}

func (aclCmd *aclOnPremCommand) init() {
	createCmd = &cobra.Command{
		Use:   "create",
		Short: "Create a Kafka ACL.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(aclCmd.create),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "You can specify only one of the following flags per command invocation: ``cluster``, ``consumer-group``, ``topic``, or ``transactional-id``. For example, to modify both ``consumer-group`` and ``topic`` resources, you need to issue two separate commands:",
				Code: "ccloud kafka acl create --allow --service-account 1522 --operation READ --consumer-group java_example_group_1\nccloud kafka acl create --allow --service-account 1522 --operation READ --topic '*'",
			}),
	}
	createCmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	createCmd.Flags().AddFlagSet(aclConfigFlags())
	createCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	createCmd.Flags().SortFlags = false
	aclCmd.AddCommand(createCmd)

	deleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete a Kafka ACL.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(aclCmd.delete),
	}
	deleteCmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	deleteCmd.Flags().AddFlagSet(aclConfigFlags())
	deleteCmd.Flags().SortFlags = false
	aclCmd.AddCommand(deleteCmd)

	listCmd = &cobra.Command{
		Use:   "list",
		Short: "List Kafka ACLs for a resource.",
		Args: cobra.NoArgs,
		RunE: pcmd.NewCLIRunE(aclCmd.list),
	}
	listCmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	listCmd.Flags().AddFlagSet(resourceFlags())
	listCmd.Flags().Int("service-account", 0, "Service account ID.")
	listCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	aclCmd.AddCommand(listCmd)
}

func (aclCmd *aclOnPremCommand) list(cmd *cobra.Command, _ []string) error {
	acl, err := parse(cmd)
	if err != nil {
		return err
	}
	restClient, restContext, err := initKafkaRest(aclCmd.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}
	opts := aclBindingToClustersClusterIdAclsGetOpts(acl[0].ACLBinding)
	clusterId, err := getClusterIdForRestRequests(restClient, restContext)
	if err != nil {
		return err
	}
	aclGetResp, httpResp, err := restClient.ACLApi.ClustersClusterIdAclsGet(restContext, clusterId, &opts)
	if err != nil {
		return kafkaRestError(restClient.GetConfig().BasePath, err, httpResp)
	}
	return aclutil.PrintACLsFromKafkaRestResponse(cmd, aclGetResp, cmd.OutOrStdout())
}

func (aclCmd *aclOnPremCommand) create(cmd *cobra.Command, _ []string) error {
	acls, err := parse(cmd)
	if err != nil {
		return err
	}
	restClient, restContext, err := initKafkaRest(aclCmd.AuthenticatedCLICommand, cmd)
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
	clusterId, err := getClusterIdForRestRequests(restClient, restContext)
	if err != nil {
		return err
	}
	for i, binding := range bindings {
		opts := aclBindingToClustersClusterIdAclsPostOpts(binding)
		httpResp, err := restClient.ACLApi.ClustersClusterIdAclsPost(restContext, clusterId, &opts)
		if err != nil {
			if i > 0 {
				_ = aclutil.PrintACLs(cmd, bindings[:i], cmd.OutOrStdout())
			}
			return kafkaRestError(restClient.GetConfig().BasePath, err, httpResp)
		} else if httpResp != nil && httpResp.StatusCode != http.StatusCreated {
			if i > 0 {
				_ = aclutil.PrintACLs(cmd, bindings[:i], cmd.OutOrStdout())
			}
			return errors.NewErrorWithSuggestions(
				fmt.Sprintf(errors.KafkaRestUnexpectedStatusMsg, httpResp.Request.URL, httpResp.StatusCode),
				errors.InternalServerErrorSuggestions)
		}
	}
	return aclutil.PrintACLs(cmd, bindings, cmd.OutOrStdout())
}

func (aclCmd *aclOnPremCommand) delete(cmd *cobra.Command, _ []string) error {
	return nil
}
