package kafka

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/c-bata/go-prompt"
	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/ccloud-sdk-go-v1"

	aclutil "github.com/confluentinc/cli/internal/pkg/acl"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/shell/completer"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

var (
	createCmd *cobra.Command
	deleteCmd *cobra.Command
	listCmd   *cobra.Command
)

type aclCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
	completableFlagChildren map[string][]*cobra.Command
}

// NewACLCommand returns the Cobra command for Kafka ACL.
func NewACLCommand(cfg *v3.Config, prerunner pcmd.PreRunner) *aclCommand {
	cmd := &cobra.Command{
		Use:   "acl",
		Short: "Manage Kafka ACLs.",
	}

	flagMap := OnPremTopicSubcommandFlags
	if cfg.IsCloudLogin() {
		flagMap = AclSubcommandFlags
	}

	c := &aclCommand{AuthenticatedStateFlagCommand: pcmd.NewAuthenticatedStateFlagCommand(cmd, prerunner, flagMap)}

	if cfg.IsCloudLogin() {
		c.init()
	} else {
		c.SetPersistentPreRunE(prerunner.InitializeOnPremKafkaRest(c.AuthenticatedCLICommand))
		c.onPremInit()
	}

	return c
}

func (c *aclCommand) init() {
	createCmd = &cobra.Command{
		Use:   "create",
		Short: "Create a Kafka ACL.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.create),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "You can specify only one of the following flags per command invocation: `cluster`, `consumer-group`, `topic`, or `transactional-id`. For example, for a consumer to read a topic, you need to grant `READ` and `DESCRIBE` both on the `consumer-group` and the `topic` resources, issuing two separate commands:",
				Code: "confluent kafka acl create --allow --service-account sa-55555 --operation READ --operation DESCRIBE --consumer-group java_example_group_1\nconfluent kafka acl create --allow --service-account sa-55555 --operation READ --operation DESCRIBE --topic '*'",
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
	listCmd.Flags().String("service-account", "", "Service account ID.")
	listCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	listCmd.Flags().SortFlags = false

	c.AddCommand(listCmd)

	c.completableFlagChildren = map[string][]*cobra.Command{
		"cluster":         {createCmd, deleteCmd, listCmd},
		"service-account": {createCmd, deleteCmd, listCmd},
	}
}

func (c *aclCommand) list(cmd *cobra.Command, _ []string) error {
	acl, err := parse(cmd)
	if err != nil {
		return err
	}

	err = c.ACLResourceIdtoNumericId(acl)
	if err != nil {
		return err
	}

	IdMap, err := getUserIdMap(c.Client)
	if err != nil {
		return err
	}

	kafkaREST, _ := c.GetKafkaREST()
	if kafkaREST != nil {
		opts := aclBindingToClustersClusterIdAclsGetOpts(acl[0].ACLBinding)

		kafkaClusterConfig, err := c.Context.GetKafkaClusterForCommand(cmd)
		if err != nil {
			return err
		}
		lkc := kafkaClusterConfig.ID

		aclGetResp, httpResp, err := kafkaREST.Client.ACLApi.ClustersClusterIdAclsGet(kafkaREST.Context, lkc, &opts)

		if err != nil && httpResp != nil {
			// Kafka REST is available, but an error occurred
			return kafkaRestError(kafkaREST.Client.GetConfig().BasePath, err, httpResp)
		}

		if err == nil && httpResp != nil {
			if httpResp.StatusCode != http.StatusOK {
				return errors.NewErrorWithSuggestions(
					fmt.Sprintf(errors.KafkaRestUnexpectedStatusMsg, httpResp.Request.URL, httpResp.StatusCode),
					errors.InternalServerErrorSuggestions)
			}
			// Kafka REST is available and there was no error
			return aclutil.PrintACLsFromKafkaRestResponseWithMap(cmd, aclGetResp, cmd.OutOrStdout(), IdMap)
		}
	}

	// Kafka REST is not available, fallback to KafkaAPI

	cluster, err := pcmd.KafkaCluster(cmd, c.Context)
	if err != nil {
		return err
	}
	resp, err := c.Client.Kafka.ListACLs(context.Background(), cluster, convertToFilter(acl[0].ACLBinding))

	if err != nil {
		return err
	}
	return aclutil.PrintACLsWithMap(cmd, resp, os.Stdout, IdMap)
}

func (c *aclCommand) create(cmd *cobra.Command, _ []string) error {
	acls, err := parse(cmd)
	if err != nil {
		return err
	}

	err = c.ACLResourceIdtoNumericId(acls)
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
		kafkaClusterConfig, err := c.AuthenticatedCLICommand.Context.GetKafkaClusterForCommand(cmd)
		if err != nil {
			return err
		}
		lkc := kafkaClusterConfig.ID

		kafkaRestExists := true
		for i, binding := range bindings {
			opts := aclBindingToClustersClusterIdAclsPostOpts(binding)
			httpResp, err := kafkaREST.Client.ACLApi.ClustersClusterIdAclsPost(kafkaREST.Context, lkc, &opts)

			if err != nil && httpResp == nil {
				if i == 0 {
					// assume Kafka REST is not available, fallback to KafkaAPI
					kafkaRestExists = false
					break
				}
				// i > 0: unlikely
				_ = aclutil.PrintACLs(cmd, bindings[:i], os.Stdout)
				return kafkaRestError(kafkaREST.Client.GetConfig().BasePath, err, httpResp)
			}

			if err != nil {
				if i > 0 {
					// unlikely
					_ = aclutil.PrintACLs(cmd, bindings[:i], os.Stdout)
				}
				return kafkaRestError(kafkaREST.Client.GetConfig().BasePath, err, httpResp)
			}

			if httpResp != nil && httpResp.StatusCode != http.StatusCreated {
				if i > 0 {
					_ = aclutil.PrintACLs(cmd, bindings[:i], os.Stdout)
				}
				return errors.NewErrorWithSuggestions(
					fmt.Sprintf(errors.KafkaRestUnexpectedStatusMsg, httpResp.Request.URL, httpResp.StatusCode),
					errors.InternalServerErrorSuggestions)
			}
		}

		if kafkaRestExists {
			return aclutil.PrintACLs(cmd, bindings, os.Stdout)
		}
	}

	// Kafka REST is not available, fallback to KafkaAPI

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

	err = c.ACLResourceIdtoNumericId(acls)
	if err != nil {
		return err
	}

	var filters []*schedv1.ACLFilter
	for _, acl := range acls {
		validateAddAndDelete(acl)
		if acl.errors != nil {
			return acl.errors
		}
		filters = append(filters, convertToFilter(acl.ACLBinding))
	}

	kafkaREST, _ := c.GetKafkaREST()
	if kafkaREST != nil {
		kafkaClusterConfig, err := c.AuthenticatedCLICommand.Context.GetKafkaClusterForCommand(cmd)
		if err != nil {
			return err
		}
		lkc := kafkaClusterConfig.ID

		kafkaRestURL := kafkaREST.Client.GetConfig().BasePath

		kafkaRestExists := true
		matchingBindingCount := 0
		for i, filter := range filters {
			deleteOpts := aclFilterToClustersClusterIdAclsDeleteOpts(filter)
			deleteResp, httpResp, err := kafkaREST.Client.ACLApi.ClustersClusterIdAclsDelete(kafkaREST.Context, lkc, &deleteOpts)

			if err != nil && httpResp == nil {
				if i == 0 {
					// Kafka REST is not available, fallback to KafkaAPI
					kafkaRestExists = false
					break
				}
				// i > 0: unlikely
				printAclsDeleted(cmd, matchingBindingCount)
				return kafkaRestError(kafkaRestURL, err, httpResp)
			}

			if err != nil {
				if i > 0 {
					// unlikely
					printAclsDeleted(cmd, matchingBindingCount)
				}
				return kafkaRestError(kafkaRestURL, err, httpResp)
			}

			if httpResp.StatusCode == http.StatusOK {
				matchingBindingCount += len(deleteResp.Data)
			} else {
				printAclsDeleted(cmd, matchingBindingCount)
				return errors.NewErrorWithSuggestions(
					fmt.Sprintf(errors.KafkaRestUnexpectedStatusMsg, httpResp.Request.URL, httpResp.StatusCode),
					errors.InternalServerErrorSuggestions)
			}
		}

		if kafkaRestExists {
			// Kafka REST is available and at least one ACL was deleted
			printAclsDeleted(cmd, matchingBindingCount)
			return nil
		}
	}

	// Kafka REST is not available, fallback to KafkaAPI

	cluster, err := pcmd.KafkaCluster(cmd, c.Context)
	if err != nil {
		return err
	}

	matchingBindingCount := 0
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

// validateAddAndDelete ensures the minimum requirements for acl add and delete are met
func validateAddAndDelete(binding *ACLConfiguration) {
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

func printAclsDeleted(cmd *cobra.Command, count int) {
	if count == 0 {
		utils.ErrPrintf(cmd, errors.ACLsNotFoundMsg)
	} else {
		utils.ErrPrintf(cmd, fmt.Sprintf(errors.DeletedACLsCountMsg, count))
	}
}

func (c *aclCommand) ServerCompletableFlagChildren() map[string][]*cobra.Command {
	return c.completableFlagChildren
}

func (c *aclCommand) ServerFlagComplete() map[string]func() []prompt.Suggest {
	return map[string]func() []prompt.Suggest{
		"cluster":         completer.ClusterFlagServerCompleterFunc(c.Client, c.EnvironmentId()),
		"service-account": completer.ServiceAccountFlagCompleterFunc(c.Client),
	}
}

func (c *aclCommand) Cmd() *cobra.Command {
	return c.Command
}

func (c *aclCommand) ACLResourceIdtoNumericId(acl []*ACLConfiguration) error {
	IdMap, err := getResourceIdMap(c.Client)
	if err != nil {
		return err
	}
	for i := 0; i < len(acl); i++ {
		if acl[i].ACLBinding.Entry.Principal != "" { // it has a service-account flag
			id := acl[i].ACLBinding.Entry.Principal[5:] // extract service account id
			_, err := strconv.Atoi(id)
			if err != nil { // it's a resource id
				acl[i].ACLBinding.Entry.Principal = "User:" + strconv.Itoa(int(IdMap[id])) // translate into numeric ID
			}
		}
	}
	return nil
}

func getResourceIdMap(client *ccloud.Client) (map[string]int32, error) {
	serviceAccounts, err := client.User.GetServiceAccounts(context.Background())
	if err != nil {
		return nil, err
	}
	saMap := make(map[string]int32)
	for _, sa := range serviceAccounts {
		saMap[sa.ResourceId] = sa.Id
	}
	return saMap, nil
}

func getUserIdMap(client *ccloud.Client) (map[int32]string, error) {
	serviceAccounts, err := client.User.GetServiceAccounts(context.Background())
	if err != nil {
		return nil, err
	}
	saMap := make(map[int32]string)
	for _, sa := range serviceAccounts {
		saMap[sa.Id] = sa.ResourceId
	}
	return saMap, nil
}
