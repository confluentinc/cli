package iam

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	mdsv2 "github.com/confluentinc/ccloud-sdk-go-v2/mds/v2"
	"github.com/confluentinc/mds-sdk-go-public/mdsv1"
	"github.com/confluentinc/mds-sdk-go-public/mdsv2alpha1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
	presource "github.com/confluentinc/cli/v3/pkg/resource"
	"github.com/confluentinc/cli/v3/pkg/types"
)

const (
	httpStatusCodeErrorMsg         = "no error but received HTTP status code %d"
	httpStatusCodeSuggestions      = "Please file a support ticket with details."
	invalidResourceTypeErrorMsg    = `invalid resource type "%s"`
	invalidResourceTypeSuggestions = "The available resource types are: %s."
	lookUpRoleSuggestions          = "To check for valid roles, use `confluent iam rbac role list`."
	principalFormatErrorMsg        = "incorrect principal format specified"
	principalFormatSuggestions     = "Principal must be specified in this format: \"<Principal Type>:<Principal Name>\".\nFor example, \"User:u-xxxxxx\" or \"User:sa-xxxxxx\"."
	resourceFormatErrorMsg         = "incorrect resource format specified"
	resourceFormatSuggestions      = "Resource must be specified in this format: `<Resource Type>:<Resource Name>`."
	specifyCloudClusterErrorMsg    = "must specify `--cloud-cluster` to indicate role binding scope"
	specifyEnvironmentErrorMsg     = "must specify `--environment` to indicate role binding scope"
)

var (
	resourcePatternListFields = []string{"Principal", "Role", "ResourceType", "Name", "PatternType"}

	// ccloud has Email as additional field
	ccloudResourcePatternListFields = []string{"Principal", "Email", "Role", "Environment", "CloudCluster", "ClusterType", "LogicalCluster", "ResourceType", "Name", "PatternType"}

	clusterScopedRoles = types.NewSet(
		"ClusterAdmin",
		"Operator",
		"SecurityAdmin",
		"SystemAdmin",
		"UserAdmin",
	)
	clusterScopedRolesV2   = types.NewSet("CloudClusterAdmin")
	environmentScopedRoles = types.NewSet("EnvironmentAdmin")

	literalPatternType  = "LITERAL"
	prefixedPatternType = "PREFIXED"
)

type roleBindingOptions struct {
	role             string
	resource         string
	prefix           bool
	principal        string
	mdsScope         mdsv1.MdsScope
	resourcesRequest mdsv1.ResourcesRequest
}

type roleBindingCommand struct {
	*pcmd.AuthenticatedCLICommand
	cfg *config.Config
}

type roleBindingOut struct {
	Id             string `human:"ID,omitempty" serialized:"id,omitempty"`
	Principal      string `human:"Principal" serialized:"principal"`
	Email          string `human:"Email" serialized:"email"`
	Role           string `human:"Role" serialized:"role"`
	Environment    string `human:"Environment" serialized:"environment"`
	CloudCluster   string `human:"Cloud Cluster" serialized:"cloud_cluster"`
	ClusterType    string `human:"Cluster Type" serialized:"cluster_type"`
	LogicalCluster string `human:"Logical Cluster" serialized:"logical_cluster"`
	ResourceType   string `human:"Resource Type" serialized:"resource_type"`
	Name           string `human:"Name" serialized:"name"`
	PatternType    string `human:"Pattern Type" serialized:"pattern_type"`
}

func newRoleBindingCommand(cfg *config.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "role-binding",
		Aliases: []string{"rb"},
		Short:   "Manage RBAC and IAM role bindings.",
		Long:    "Manage Role-Based Access Control (RBAC) and Identity and Access Management (IAM) role bindings.",
	}

	c := &roleBindingCommand{cfg: cfg}

	if cfg.IsOnPremLogin() {
		c.AuthenticatedCLICommand = pcmd.NewAuthenticatedWithMDSCLICommand(cmd, prerunner)
	} else {
		c.AuthenticatedCLICommand = pcmd.NewAuthenticatedCLICommand(cmd, prerunner)
	}

	cmd.AddCommand(c.newCreateCommand())
	cmd.AddCommand(c.newDeleteCommand())
	cmd.AddCommand(c.newListCommand())

	return cmd
}

func (c *roleBindingCommand) parseCommon(cmd *cobra.Command) (*roleBindingOptions, error) {
	role, err := cmd.Flags().GetString("role")
	if err != nil {
		return nil, err
	}

	resource, err := cmd.Flags().GetString("resource")
	if err != nil {
		return nil, err
	}

	// The err is ignored here since the --prefix flag is not defined by the list subcommand
	prefix, _ := cmd.Flags().GetBool("prefix")

	principal, err := cmd.Flags().GetString("principal")
	if err != nil {
		return nil, err
	}

	if cmd.Flags().Changed("principal") {
		if err := c.validatePrincipalFormat(principal); err != nil {
			return nil, err
		}
	}

	scope, err := c.parseAndValidateScope(cmd)
	if err != nil {
		return nil, err
	}

	var resourcesRequest mdsv1.ResourcesRequest
	if resource != "" {
		parsedResourcePattern, err := parseAndValidateResourcePattern(resource, prefix)
		if err != nil {
			return nil, err
		}

		// Resource types are defined under roles' access policies, so if no role is specified,
		// we have to loop over the possible resource types for all roles (this is what
		// validateResourceTypeV1 does).
		if role != "" {
			if err := c.validateRoleAndResourceTypeV1(role, parsedResourcePattern.ResourceType); err != nil {
				return nil, err
			}
		} else {
			if err := c.validateResourceTypeV1(parsedResourcePattern.ResourceType); err != nil {
				return nil, err
			}
		}

		resourcesRequest = mdsv1.ResourcesRequest{
			Scope:            *scope,
			ResourcePatterns: []mdsv1.ResourcePattern{parsedResourcePattern},
		}
	}
	return &roleBindingOptions{
			role,
			resource,
			prefix,
			principal,
			*scope,
			resourcesRequest,
		},
		nil
}

/*
Helper function to add flags for all the legal scopes/clusters for the command.
*/
func addClusterFlags(cmd *cobra.Command, cfg *config.Config, cliCommand *pcmd.CLICommand) {
	if cfg.IsCloudLogin() {
		cmd.Flags().String("environment", "", "Environment ID for scope of role-binding operation.")
		cmd.Flags().Bool("current-environment", false, "Use current environment ID for scope.")
		cmd.Flags().String("cloud-cluster", "", "Cloud cluster ID for the role binding.")
		cmd.Flags().String("flink-region", "", "Flink region ID for the role binding.")
		cmd.Flags().String("kafka-cluster", "", "Kafka cluster ID for the role binding.")
		cmd.Flags().String("schema-registry-cluster", "", "Schema Registry cluster ID for the role binding.")
		cmd.Flags().String("ksql-cluster", "", "ksqlDB cluster name for the role binding.")
	} else {
		cmd.Flags().String("kafka-cluster", "", "Kafka cluster ID for the role binding.")
		cmd.Flags().String("schema-registry-cluster", "", "Schema Registry cluster ID for the role binding.")
		cmd.Flags().String("ksql-cluster", "", "ksqlDB cluster ID for the role binding.")
		cmd.Flags().String("connect-cluster", "", "Kafka Connect cluster ID for the role binding.")
		cmd.Flags().String("cluster-name", "", "Cluster name to uniquely identify the cluster for role binding listings.")
		pcmd.AddContextFlag(cmd, cliCommand)
	}
}

func (c *roleBindingCommand) validatePrincipalFormat(principal string) error {
	if len(strings.Split(principal, ":")) == 1 {
		return errors.NewErrorWithSuggestions(principalFormatErrorMsg, principalFormatSuggestions)
	}

	return nil
}

func (c *roleBindingCommand) parseAndValidateScope(cmd *cobra.Command) (*mdsv1.MdsScope, error) {
	scope := &mdsv1.MdsScopeClusters{}
	nonKafkaScopesSet := 0

	clusterName, err := cmd.Flags().GetString("cluster-name")
	if err != nil {
		return nil, err
	}

	cmd.Flags().Visit(func(flag *pflag.Flag) {
		switch flag.Name {
		case "kafka-cluster":
			scope.KafkaCluster = flag.Value.String()
		case "schema-registry-cluster":
			scope.SchemaRegistryCluster = flag.Value.String()
			nonKafkaScopesSet++
		case "ksql-cluster":
			scope.KsqlCluster = flag.Value.String()
			nonKafkaScopesSet++
		case "connect-cluster":
			scope.ConnectCluster = flag.Value.String()
			nonKafkaScopesSet++
		}
	})

	if clusterName != "" && (scope.KafkaCluster != "" || nonKafkaScopesSet > 0) {
		return nil, errors.New("cannot specify both cluster name and cluster scope")
	}

	if clusterName == "" {
		if scope.KafkaCluster == "" && nonKafkaScopesSet > 0 {
			return nil, errors.New(errors.SpecifyKafkaIdErrorMsg)
		}

		if scope.KafkaCluster == "" && nonKafkaScopesSet == 0 {
			return nil, errors.New("must specify either cluster ID to indicate role binding scope or the cluster name")
		}

		if nonKafkaScopesSet > 1 {
			return nil, errors.New(errors.MoreThanOneNonKafkaErrorMsg)
		}

		return &mdsv1.MdsScope{Clusters: *scope}, nil
	}

	return &mdsv1.MdsScope{ClusterName: clusterName}, nil
}

func (c *roleBindingCommand) validateResourceTypeV2(resourceType string) error {
	ctx := c.createContext()
	roles, _, err := c.MDSv2Client.RBACRoleDefinitionsApi.Roles(ctx, nil)
	if err != nil {
		return err
	}

	allResourceTypes := make(map[string]bool)
	found := false
	for _, role := range roles {
		for _, policies := range role.Policies {
			for _, operation := range policies.AllowedOperations {
				allResourceTypes[operation.ResourceType] = true
				if operation.ResourceType == resourceType {
					found = true
					break
				}
			}
		}
	}

	if !found {
		uniqueResourceTypes := []string{}
		for rt := range allResourceTypes {
			uniqueResourceTypes = append(uniqueResourceTypes, rt)
		}
		suggestionsMsg := fmt.Sprintf(invalidResourceTypeSuggestions, strings.Join(uniqueResourceTypes, ", "))
		return errors.NewErrorWithSuggestions(fmt.Sprintf(invalidResourceTypeErrorMsg, resourceType), suggestionsMsg)
	}

	return nil
}

func parseAndValidateResourcePattern(resource string, prefix bool) (mdsv1.ResourcePattern, error) {
	result := mdsv1.ResourcePattern{PatternType: "LITERAL"}
	if prefix {
		result.PatternType = "PREFIXED"
	}

	parts := strings.SplitN(resource, ":", 2)
	if len(parts) != 2 {
		return result, errors.NewErrorWithSuggestions(resourceFormatErrorMsg, resourceFormatSuggestions)
	}
	result.ResourceType = parts[0]
	result.Name = parts[1]

	return result, nil
}

func (c *roleBindingCommand) validateRoleAndResourceTypeV1(roleName, resourceType string) error {
	errorMsg := fmt.Sprintf(`failed to look up role "%s"`, roleName)

	role, resp, err := c.MDSClient.RBACRoleDefinitionsApi.RoleDetail(c.createContext(), roleName)
	if err != nil {
		return errors.NewWrapErrorWithSuggestions(err, errorMsg, lookUpRoleSuggestions)
	}
	if resp.StatusCode == http.StatusNoContent {
		return errors.NewErrorWithSuggestions(errorMsg, lookUpRoleSuggestions)
	}

	allResourceTypes := make([]string, len(role.AccessPolicy.AllowedOperations))
	found := false
	for i, operation := range role.AccessPolicy.AllowedOperations {
		if operation.ResourceType == resourceType {
			found = true
			break
		}
		allResourceTypes[i] = operation.ResourceType
	}

	if !found {
		return errors.NewErrorWithSuggestions(
			fmt.Sprintf(invalidResourceTypeErrorMsg, resourceType),
			fmt.Sprintf(invalidResourceTypeSuggestions, strings.Join(allResourceTypes, ", ")),
		)
	}

	return nil
}

func (c *roleBindingCommand) validateResourceTypeV1(resourceType string) error {
	ctx := c.createContext()
	roles, _, err := c.MDSClient.RBACRoleDefinitionsApi.Roles(ctx)
	if err != nil {
		return err
	}

	allResourceTypes := []string{}
	found := false
	for _, role := range roles {
		for _, operation := range role.AccessPolicy.AllowedOperations {
			allResourceTypes = append(allResourceTypes, operation.ResourceType)
			if operation.ResourceType == resourceType {
				found = true
				break
			}
		}
	}

	if !found {
		uniqueResourceTypes := types.RemoveDuplicates(allResourceTypes)
		suggestionsMsg := fmt.Sprintf(invalidResourceTypeSuggestions, strings.Join(uniqueResourceTypes, ", "))
		return errors.NewErrorWithSuggestions(fmt.Sprintf(invalidResourceTypeErrorMsg, resourceType), suggestionsMsg)
	}

	return nil
}

func (c *roleBindingCommand) displayCCloudCreateAndDeleteOutput(cmd *cobra.Command, roleBinding *mdsv2.IamV2RoleBinding) error {
	userResourceId := strings.TrimPrefix(roleBinding.GetPrincipal(), "User:")

	out := &roleBindingOut{
		Id:        roleBinding.GetId(),
		Principal: roleBinding.GetPrincipal(),
		Role:      roleBinding.GetRoleName(),
	}

	// The err is ignored here since the --prefix flag is not defined by the list subcommand
	prefix, _ := cmd.Flags().GetBool("prefix")

	resource, err := cmd.Flags().GetString("resource")
	if err != nil {
		return err
	}
	if resource != "" {
		parts := strings.SplitN(resource, ":", 2)
		if len(parts) != 2 {
			return errors.NewErrorWithSuggestions(resourceFormatErrorMsg, resourceFormatSuggestions)
		}
		resourceType := parts[0]
		if resourceType == "Cluster" {
			resourceType = "kafka"
		}

		out.ResourceType = resourceType
		out.Name = parts[1]
		if prefix {
			out.PatternType = prefixedPatternType
		} else {
			out.PatternType = literalPatternType
		}
	}

	var fields []string
	principalType := presource.LookupType(userResourceId)
	if principalType == presource.ServiceAccount || principalType == presource.IdentityPool {
		if resource != "" {
			fields = resourcePatternListFields
		} else {
			fields = []string{"Id", "Principal", "Role"}
		}
	} else {
		if resource != "" {
			fields = ccloudResourcePatternListFields
		} else {
			user, err := c.V2Client.GetIamUserById(userResourceId)
			if err != nil {
				return err
			}
			out.Email = user.GetEmail()
			fields = []string{"Id", "Principal", "Email", "Role"}
		}
	}

	table := output.NewTable(cmd)
	table.Add(out)
	table.Filter(fields)
	return table.Print()
}

func displayCreateAndDeleteOutput(cmd *cobra.Command, options *roleBindingOptions) error {
	var fieldsSelected []string

	out := &roleBindingOut{
		Principal: options.principal,
		Role:      options.role,
	}
	if options.resource != "" {
		fieldsSelected = resourcePatternListFields
		if len(options.resourcesRequest.ResourcePatterns) != 1 {
			return errors.New("display error: number of resource pattern is not 1")
		}
		resourcePattern := options.resourcesRequest.ResourcePatterns[0]
		out.ResourceType = resourcePattern.ResourceType
		out.Name = resourcePattern.Name
		out.PatternType = resourcePattern.PatternType
	} else {
		fieldsSelected = []string{"Principal", "Role", "ResourceType"}
		out.ResourceType = "Cluster"
	}

	table := output.NewTable(cmd)
	table.Add(out)
	table.Filter(fieldsSelected)
	return table.Print()
}

func (c *roleBindingCommand) createContext() context.Context {
	if c.cfg.IsCloudLogin() {
		return context.WithValue(context.Background(), mdsv2alpha1.ContextAccessToken, c.Context.GetAuthToken())
	} else {
		return context.WithValue(context.Background(), mdsv1.ContextAccessToken, c.Context.GetAuthToken())
	}
}

func (c *roleBindingCommand) parseV2RoleBinding(cmd *cobra.Command) (*mdsv2.IamV2RoleBinding, error) {
	role, err := cmd.Flags().GetString("role")
	if err != nil {
		return nil, err
	}

	principal, err := cmd.Flags().GetString("principal")
	if err != nil {
		return nil, err
	}
	if cmd.Flags().Changed("principal") {
		if err = c.validatePrincipalFormat(principal); err != nil {
			return nil, err
		}
	}

	if strings.HasPrefix(principal, "User:") {
		principalValue := strings.TrimPrefix(principal, "User:")
		if strings.Contains(principalValue, "@") {
			user, err := c.V2Client.GetIamUserByEmail(principalValue)
			if err != nil {
				return nil, err
			}
			principal = "User:" + user.GetId()
		}
	}

	crnPattern, err := c.parseV2BaseCrnPattern(cmd)
	if err != nil {
		return nil, err
	}

	// The err is ignored here since the --prefix flag is not defined by the list subcommand
	prefix, _ := cmd.Flags().GetBool("prefix")

	resource, err := cmd.Flags().GetString("resource")
	if err != nil {
		return nil, err
	}
	if resource != "" {
		parts := strings.SplitN(resource, ":", 2)
		if len(parts) != 2 {
			return nil, errors.NewErrorWithSuggestions(resourceFormatErrorMsg, resourceFormatSuggestions)
		}
		resourceType := parts[0]
		resourceName := parts[1]
		if resourceType == "Cluster" {
			resourceType = "kafka"
		}
		if resourceType == "ServiceAccount" {
			resourceType = "service-account"
		}
		if resourceType == "ComputePool" {
			resourceType = "compute-pool"
		}

		if role == "" {
			if err := c.validateResourceTypeV2(resourceType); err != nil {
				return nil, err
			}
		}

		crnPattern += fmt.Sprintf("/%s=%s", strings.ToLower(resourceType), resourceName)

		if prefix {
			crnPattern += "*"
		}
	}

	return &mdsv2.IamV2RoleBinding{
		Principal:  mdsv2.PtrString(principal),
		RoleName:   mdsv2.PtrString(role),
		CrnPattern: mdsv2.PtrString(crnPattern),
	}, nil
}

func (c *roleBindingCommand) parseV2BaseCrnPattern(cmd *cobra.Command) (string, error) {
	orgResourceId := c.Context.GetCurrentOrganization()
	crnPattern := "crn://confluent.cloud/organization=" + orgResourceId

	if cmd.Flags().Changed("current-environment") {
		environmentId, err := c.Context.EnvironmentId()
		if err != nil {
			return "", err
		}
		crnPattern += "/environment=" + environmentId
	} else if cmd.Flags().Changed("environment") {
		environment, err := cmd.Flags().GetString("environment")
		if err != nil {
			return "", err
		}
		crnPattern += "/environment=" + environment
	}

	if cmd.Flags().Changed("flink-region") {
		flinkRegion, err := cmd.Flags().GetString("flink-region")
		if err != nil {
			return "", err
		}
		crnPattern += "/flink-region=" + flinkRegion
	}

	if cmd.Flags().Changed("cloud-cluster") {
		cloudCluster, err := cmd.Flags().GetString("cloud-cluster")
		if err != nil {
			return "", err
		}
		crnPattern += "/cloud-cluster=" + cloudCluster
	}

	if cmd.Flags().Changed("schema-registry-cluster") { // route not implemented yet
		schemaRegistryCluster, err := cmd.Flags().GetString("schema-registry-cluster")
		if err != nil {
			return "", err
		}
		crnPattern += "/schema-registry=" + schemaRegistryCluster
	}

	if cmd.Flags().Changed("ksql-cluster") { // route not implemented yet
		ksqlCluster, err := cmd.Flags().GetString("ksql-cluster")
		if err != nil {
			return "", err
		}
		crnPattern += "/ksql=" + ksqlCluster
	}

	if cmd.Flags().Changed("kafka-cluster") {
		kafkaCluster, err := cmd.Flags().GetString("kafka-cluster")
		if err != nil {
			return "", err
		}
		crnPattern += "/kafka=" + kafkaCluster
	}

	if cmd.Flags().Changed("role") {
		role, err := cmd.Flags().GetString("role")
		if err != nil {
			return "", err
		}
		if clusterScopedRolesV2.Contains(role) && !cmd.Flags().Changed("cloud-cluster") {
			return "", errors.New(specifyCloudClusterErrorMsg)
		}
		if (environmentScopedRoles[role] || clusterScopedRolesV2.Contains(role)) && !cmd.Flags().Changed("current-environment") && !cmd.Flags().Changed("environment") {
			return "", errors.New(specifyEnvironmentErrorMsg)
		}
	}

	if cmd.Flags().Changed("cloud-cluster") && !cmd.Flags().Changed("current-environment") && !cmd.Flags().Changed("environment") {
		return "", errors.New(specifyEnvironmentErrorMsg)
	}
	return crnPattern, nil
}
