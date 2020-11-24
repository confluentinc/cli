package kafka

import (
	"context"
	"fmt"
	"os"

	"github.com/confluentinc/cli/internal/pkg/examples"
	krsdk "github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"

	aclutil "github.com/confluentinc/cli/internal/pkg/acl"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

var (
	createCmd *cobra.Command
	deleteCmd *cobra.Command
	listCmd   *cobra.Command
)

type aclCommand struct {
	*pcmd.AuthenticatedCLICommand
}

// NewACLCommand returns the Cobra command for Kafka ACL.
func NewACLCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cliCmd := pcmd.NewAuthenticatedCLICommand(
		&cobra.Command{
			Use:   "acl",
			Short: "Manage Kafka ACLs.",
		}, prerunner)
	cmd := &aclCommand{AuthenticatedCLICommand: cliCmd}
	cmd.init()
	return cmd.Command
}

func (c *aclCommand) init() {
	c.Command.PersistentFlags().String("cluster", "", "Kafka cluster ID.")

	createCmd = &cobra.Command{
		Use:   "create",
		Short: "Create a Kafka ACL.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.create),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "You can only specify one of these flags per command invocation: ``cluster``, ``consumer-group``, ``topic``, or ``transactional-id``. For example, if you want to specify both ``consumer-group`` and ``topic``, you must specify this as two separate commands:",
				Code: "ccloud kafka acl create --allow --service-account 1522 --operation READ --consumer-group java_example_group_1\nccloud kafka acl create --allow --service-account 1522 --operation READ --topic '*'",
			},
		),
	}
	createCmd.Flags().AddFlagSet(aclConfigFlags())
	createCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	createCmd.Flags().SortFlags = false

	c.AddCommand(createCmd)

	deleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete a Kafka ACL.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.delete),
	}
	deleteCmd.Flags().AddFlagSet(aclConfigFlags())
	deleteCmd.Flags().SortFlags = false

	c.AddCommand(deleteCmd)

	listCmd = &cobra.Command{
		Use:   "list",
		Short: "List Kafka ACLs for a resource.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.list),
	}
	listCmd.Flags().AddFlagSet(resourceFlags())
	listCmd.Flags().Int("service-account", 0, "Service account ID.")
	listCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	listCmd.Flags().SortFlags = false

	c.AddCommand(listCmd)
}

func (c *aclCommand) list(cmd *cobra.Command, _ []string) error {
	acl, err := parse(cmd)
	if err != nil {
		return err
	}

	kafkaRestGetConfig := convertAclBindingToGetParams(acl[0].ACLBinding)

	kafkaClusterConfig, err := c.AuthenticatedCLICommand.Context.GetKafkaClusterForCommand(cmd)
	if err != nil {
		return err
	}
	lkc := kafkaClusterConfig.ID

	kafkaRestURL, err := bootstrapServersToRestURL(kafkaClusterConfig.Bootstrap)
	if err != nil {
		return err
	}

	// Set Kafka-REST client to correct URL
	c.KafkaRestClient.ChangeBasePath(kafkaRestURL)
	kafkaRestClient := c.KafkaRestClient

	state, err := c.AuthenticatedCLICommand.Context.AuthenticatedState(cmd)
	if err != nil {
		return err
	}

	accessToken, err := getAccessToken(state, c.Context.Platform.Server)
	if err != nil {
		return err
	}

	// create new context with access token to be used in Kafka-REST call
	newCtx := context.WithValue(context.Background(), krsdk.ContextAccessToken, accessToken)

	aclGetResp, httpResp, err := kafkaRestClient.ACLApi.ClustersClusterIdAclsGet(newCtx, lkc, &kafkaRestGetConfig)

	// Kafka-REST exists and no error
	if err == nil && httpResp != nil && httpResp.StatusCode == 200 {
		fmt.Println("using kafka rest list")
		aclDatas := aclGetResp.Data
		aclListFields := []string{"ServiceAccountId", "Permission", "Operation", "Resource", "Name", "Type"}
		aclListStructuredRenames := []string{"principal", "permission", "operation", "resource_type", "resource_name", "pattern_type"}
		outputWriter, err := output.NewListOutputCustomizableWriter(cmd, aclListFields, aclListFields, aclListStructuredRenames, os.Stdout)
		if err != nil {
			return err
		}
		for _, aclData := range aclDatas {
			record := &struct {
				ServiceAccountId string
				Permission       string
				Operation        string
				Resource         string
				Name             string
				Type             string
			}{
				aclData.Principal,
				string(aclData.Permission),
				string(aclData.Operation),
				string(aclData.ResourceType),
				aclData.Host,
				string(aclData.PatternType),
			}
			outputWriter.AddElement(record)
		}
		return outputWriter.Out()

	}

	// Kafka-REST exists but Kafka-REST error occurred
	if err != nil && httpResp != nil && httpResp.StatusCode >= 400 && httpResp.StatusCode != 404 {
		return handleCommonKafkaRestClientErrors(kafkaRestURL, err)
	}

	// Kafka-REST does not exist, use Kafka-API
	cluster, err := pcmd.KafkaCluster(cmd, c.Context)
	if err != nil {
		return err
	}
	resp, err := c.Client.Kafka.ListACLs(context.Background(), cluster, convertToFilter(acl[0].ACLBinding))

	if err != nil {
		return err
	}
	return aclutil.PrintACLs(cmd, resp, os.Stdout)
}

func (c *aclCommand) create(cmd *cobra.Command, _ []string) error {
	acls, err := parse(cmd)
	if err != nil {
		return err
	}
	var bindings []*schedv1.ACLBinding
	for _, acl := range acls {
		validateAddDelete(acl)
		if acl.errors != nil {
			return acl.errors
		}
		bindings = append(bindings, acl.ACLBinding)
	}

	kafkaClusterConfig, err := c.AuthenticatedCLICommand.Context.GetKafkaClusterForCommand(cmd)
	if err != nil {
		return err
	}
	lkc := kafkaClusterConfig.ID

	kafkaRestURL, err := bootstrapServersToRestURL(kafkaClusterConfig.Bootstrap)
	if err != nil {
		return err
	}

	// Set Kafka-REST client to correct URL
	c.KafkaRestClient.ChangeBasePath(kafkaRestURL)
	kafkaRestClient := c.KafkaRestClient

	state, err := c.AuthenticatedCLICommand.Context.AuthenticatedState(cmd)
	if err != nil {
		return err
	}

	accessToken, err := getAccessToken(state, c.Context.Platform.Server)
	if err != nil {
		return err
	}

	// create new context with access token to be used in Kafka-REST call
	newCtx := context.WithValue(context.Background(), krsdk.ContextAccessToken, accessToken)

	kafkaRestExists := true
	for _, binding := range bindings {
		kafkaRestPostConfig := convertAclBindingToPostParams(binding)

		httpResp, err := kafkaRestClient.ACLApi.ClustersClusterIdAclsPost(newCtx, lkc, &kafkaRestPostConfig)

		if err != nil && httpResp == nil {
			kafkaRestExists = false
			break
		}
		// Kafka-REST exists but Kafka-REST error occurred
		if err != nil && httpResp != nil && httpResp.StatusCode >= 400 && httpResp.StatusCode != 404 {
			return handleCommonKafkaRestClientErrors(kafkaRestURL, err)
		}
	}

	// Kafka REST sent all requests successfully
	if kafkaRestExists {
		fmt.Println("using kafka rest create")
		return aclutil.PrintACLs(cmd, bindings, os.Stdout)
	}

	// Kafka-REST does not exist, use Kafka-API
	cluster, err := pcmd.KafkaCluster(cmd, c.Context)
	if err != nil {
		return err
	}

	err = c.Client.Kafka.CreateACLs(context.Background(), cluster, bindings)
	if err != nil {
		return err
	}
	return aclutil.PrintACLs(cmd, bindings, os.Stdout)
}

func (c *aclCommand) delete(cmd *cobra.Command, _ []string) error {
	acls, err := parse(cmd)
	if err != nil {
		return err
	}
	var filters []*schedv1.ACLFilter
	for _, acl := range acls {
		validateAddDelete(acl)
		if acl.errors != nil {
			return acl.errors
		}
		filters = append(filters, convertToFilter(acl.ACLBinding))
	}

	kafkaClusterConfig, err := c.AuthenticatedCLICommand.Context.GetKafkaClusterForCommand(cmd)
	if err != nil {
		return err
	}
	lkc := kafkaClusterConfig.ID

	kafkaRestURL, err := bootstrapServersToRestURL(kafkaClusterConfig.Bootstrap)
	if err != nil {
		return err
	}

	// Set Kafka-REST client to correct URL
	c.KafkaRestClient.ChangeBasePath(kafkaRestURL)
	kafkaRestClient := c.KafkaRestClient

	state, err := c.AuthenticatedCLICommand.Context.AuthenticatedState(cmd)
	if err != nil {
		return err
	}

	accessToken, err := getAccessToken(state, c.Context.Platform.Server)
	if err != nil {
		return err
	}

	// create new context with access token to be used in Kafka-REST call
	newCtx := context.WithValue(context.Background(), krsdk.ContextAccessToken, accessToken)

	matchingBindingCount := 0
	kafkaRestExists := true

	for _, filter := range filters {
		// Check to see if Kafka-REST is enabled and if specified ACL to be deleted exists
		kafkaRestGetConfig := convertAclFilterToGetParams(filter)
		aclGetResp, httpResp, err := kafkaRestClient.ACLApi.ClustersClusterIdAclsGet(newCtx, lkc, &kafkaRestGetConfig)

		// Kafka-REST exists and ACL response is not empty(ACL exists), so delete
		if err == nil && httpResp != nil && httpResp.StatusCode == 200 && len(aclGetResp.Data) > 0 {
			kafkaRestPostConfig := convertAclFilterToPostParams(filter)

			_, deleteHttpResp, err := kafkaRestClient.ACLApi.ClustersClusterIdAclsDelete(newCtx, lkc, &kafkaRestPostConfig)
			fmt.Println("using kafka rest delete")
			if err == nil && deleteHttpResp != nil && deleteHttpResp.StatusCode == 200 {
				matchingBindingCount += len(aclGetResp.Data)
			} else {
				handleCommonKafkaRestClientErrors(kafkaRestURL, err)
			}
		}

		// Kafka-REST is not enabled, use Kafka-API failover
		if err != nil && httpResp == nil {
			kafkaRestExists = false
			break
		}
	}

	// Kafka-REST exists and was able to delete at least one ACL
	if kafkaRestExists && matchingBindingCount > 0 {
		utils.ErrPrintf(cmd, errors.DeletedACLsMsg)
		return nil
	}

	// Use Kafka-API
	cluster, err := pcmd.KafkaCluster(cmd, c.Context)
	if err != nil {
		return err
	}

	matchingBindingCount = 0
	for _, acl := range acls {
		// For the tests it's useful to know that the ListACLs call is coming from the delete call.
		resp, err := c.Client.Kafka.ListACLs(context.WithValue(context.Background(), "requestor", "delete"), cluster, convertToFilter(acl.ACLBinding))
		if err != nil {
			return err
		}
		matchingBindingCount += len(resp)
	}
	if matchingBindingCount == 0 {
		utils.ErrPrintf(cmd, errors.ACLsNotFoundMsg)
		return nil
	}

	err = c.Client.Kafka.DeleteACLs(context.Background(), cluster, filters)
	if err != nil {
		return err
	}
	utils.ErrPrintf(cmd, errors.DeletedACLsMsg)
	return nil
}

// validateAddDelete ensures the minimum requirements for acl add and delete are met
func validateAddDelete(binding *ACLConfiguration) {
	if binding.Entry.PermissionType == schedv1.ACLPermissionTypes_UNKNOWN {
		binding.errors = multierror.Append(binding.errors, fmt.Errorf(errors.MustSetAllowOrDenyErrorMsg))
	}

	if binding.Pattern.PatternType == schedv1.PatternTypes_UNKNOWN {
		binding.Pattern.PatternType = schedv1.PatternTypes_LITERAL
	}

	if binding.Pattern == nil || binding.Pattern.ResourceType == schedv1.ResourceTypes_UNKNOWN {
		binding.errors = multierror.Append(binding.errors, fmt.Errorf(errors.MustSetResourceTypeErrorMsg,
			listEnum(schedv1.ResourceTypes_ResourceType_name, []string{"ANY", "UNKNOWN"})))
	}
}

// convertToFilter converts a ACLBinding to a KafkaAPIACLFilterRequest
func convertToFilter(binding *schedv1.ACLBinding) *schedv1.ACLFilter {
	// ACE matching rules
	// https://github.com/apache/kafka/blob/trunk/clients/src/main/java/org/apache/kafka/common/acl/AccessControlEntryFilter.java#L102-L113
	if binding.Entry == nil {
		binding.Entry = new(schedv1.AccessControlEntryConfig)
	}

	if binding.Entry.Operation == schedv1.ACLOperations_UNKNOWN {
		binding.Entry.Operation = schedv1.ACLOperations_ANY
	}

	if binding.Entry.PermissionType == schedv1.ACLPermissionTypes_UNKNOWN {
		binding.Entry.PermissionType = schedv1.ACLPermissionTypes_ANY
	}

	// ResourcePattern matching rules
	// https://github.com/apache/kafka/blob/trunk/clients/src/main/java/org/apache/kafka/common/resource/ResourcePatternFilter.java#L42-L56
	if binding.Pattern == nil {
		binding.Pattern = &schedv1.ResourcePatternConfig{}
	}

	binding.Entry.Host = "*"

	if binding.Pattern.ResourceType == schedv1.ResourceTypes_UNKNOWN {
		binding.Pattern.ResourceType = schedv1.ResourceTypes_ANY
	}

	if binding.Pattern.PatternType == schedv1.PatternTypes_UNKNOWN {
		binding.Pattern.PatternType = schedv1.PatternTypes_ANY
	}

	return &schedv1.ACLFilter{
		EntryFilter:   binding.Entry,
		PatternFilter: binding.Pattern,
	}
}
