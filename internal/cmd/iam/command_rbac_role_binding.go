package iam

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/antihax/optional"
	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	mdsv2 "github.com/confluentinc/ccloud-sdk-go-v2/mds/v2"
	mds "github.com/confluentinc/mds-sdk-go/mdsv1"
	"github.com/confluentinc/mds-sdk-go/mdsv2alpha1"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
)

var (
	resourcePatternListFields           = []string{"Principal", "Role", "ResourceType", "Name", "PatternType"}
	resourcePatternHumanListLabels      = []string{"Principal", "Role", "Resource Type", "Name", "Pattern Type"}
	resourcePatternStructuredListLabels = []string{"principal", "role", "resource_type", "name", "pattern_type"}

	// ccloud has Email as additional field
	ccloudResourcePatternListFields           = []string{"Principal", "Email", "Role", "Environment", "CloudCluster", "ClusterType", "LogicalCluster", "ResourceType", "Name", "PatternType"}
	ccloudResourcePatternHumanListLabels      = []string{"Principal", "Email", "Role", "Environment", "Cloud Cluster", "Cluster Type", "Logical Cluster", "Resource Type", "Name", "Pattern Type"}
	ccloudResourcePatternStructuredListLabels = []string{"principal", "email", "role", "environment", "cloud_cluster", "cluster_type", "logical_cluster", "resource_type", "resource_name", "pattern_type"}

	//TODO: please move this to a backend route (https://confluentinc.atlassian.net/browse/CIAM-890)
	clusterScopedRoles = map[string]bool{
		"SystemAdmin":   true,
		"ClusterAdmin":  true,
		"SecurityAdmin": true,
		"UserAdmin":     true,
		"Operator":      true,
	}

	clusterScopedRolesV2 = map[string]bool{
		"CloudClusterAdmin": true,
	}

	environmentScopedRoles = map[string]bool{
		"EnvironmentAdmin": true,
	}

	dataplaneNamespace = optional.NewString("dataplane")
)

type roleBindingOptions struct {
	role               string
	resource           string
	prefix             bool
	principal          string
	scopeV2            mdsv2alpha1.Scope
	mdsScope           mds.MdsScope
	resourcesRequest   mds.ResourcesRequest
	resourcesRequestV2 mdsv2alpha1.ResourcesRequest
}

type roleBindingCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
	cfg *v1.Config
}

type listDisplay struct {
	Principal      string `json:"principal"`
	Email          string `json:"email"`
	Role           string `json:"role"`
	Environment    string `json:"environment"`
	CloudCluster   string `json:"cloud_cluster"`
	ClusterType    string `json:"cluster_type"`
	LogicalCluster string `json:"logical_cluster"`
	ResourceType   string `json:"resource_type"`
	Name           string `json:"resource_name"`
	PatternType    string `json:"pattern_type"`
}

func newRoleBindingCommand(cfg *v1.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "role-binding",
		Aliases: []string{"rb"},
		Short:   "Manage RBAC and IAM role bindings.",
		Long:    "Manage Role-Based Access Control (RBAC) and Identity and Access Management (IAM) role bindings.",
	}

	c := &roleBindingCommand{
		cfg: cfg,
	}

	if cfg.IsOnPremLogin() {
		c.AuthenticatedStateFlagCommand = pcmd.NewAuthenticatedWithMDSStateFlagCommand(cmd, prerunner)
	} else {
		c.AuthenticatedStateFlagCommand = pcmd.NewAuthenticatedStateFlagCommand(cmd, prerunner)
	}

	c.AddCommand(c.newCreateCommand())
	c.AddCommand(c.newDeleteCommand())
	c.AddCommand(c.newListCommand())

	return c.Command
}

func (c *roleBindingCommand) parseCommon(cmd *cobra.Command) (*roleBindingOptions, error) {
	role, err := cmd.Flags().GetString("role")
	if err != nil {
		return nil, err
	}

	isCloud := c.cfg.IsCloudLogin()

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
	if isCloud {
		if strings.HasPrefix(principal, "User:") {
			principalValue := strings.TrimLeft(principal, "User:")
			if strings.Contains(principalValue, "@") {
				user, err := c.Client.User.Describe(context.Background(), &orgv1.User{Email: principalValue})
				if err != nil {
					return nil, err
				}
				principal = "User:" + user.ResourceId
			}
		}
	}

	if cmd.Flags().Changed("principal") {
		err = c.validatePrincipalFormat(principal)
		if err != nil {
			return nil, err
		}
	}

	scope := &mds.MdsScope{}
	scopeV2 := &mdsv2alpha1.Scope{}
	if !isCloud {
		scope, err = c.parseAndValidateScope(cmd)
	} else {
		scopeV2, err = c.parseAndValidateScopeV2(cmd)
	}
	if err != nil {
		return nil, err
	}

	resourcesRequest := mds.ResourcesRequest{}
	resourcesRequestV2 := mdsv2alpha1.ResourcesRequest{}
	if resource != "" {
		if isCloud {
			parsedResourcePattern, err := parseAndValidateResourcePatternV2(resource, prefix)
			if err != nil {
				return nil, err
			}
			// Resource types are defined under roles' access policies, so if no role is specified,
			// we have to loop over the possible resource types for all roles (this is what
			// validateResourceTypeV2 does).
			if role != "" {
				if err := c.validateRoleAndResourceTypeV2(role, parsedResourcePattern.ResourceType); err != nil {
					return nil, err
				}
			} else {
				if err := c.validateResourceTypeV2(parsedResourcePattern.ResourceType); err != nil {
					return nil, err
				}
			}

			resourcesRequestV2 = mdsv2alpha1.ResourcesRequest{
				Scope:            *scopeV2,
				ResourcePatterns: []mdsv2alpha1.ResourcePattern{parsedResourcePattern},
			}
		} else {
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

			resourcesRequest = mds.ResourcesRequest{
				Scope:            *scope,
				ResourcePatterns: []mds.ResourcePattern{parsedResourcePattern},
			}
		}
	}
	return &roleBindingOptions{
			role,
			resource,
			prefix,
			principal,
			*scopeV2,
			*scope,
			resourcesRequest,
			resourcesRequestV2,
		},
		nil
}

/*
Helper function to add flags for all the legal scopes/clusters for the command.
*/
func addClusterFlags(cmd *cobra.Command, isCloudLogin bool, cliCommand *pcmd.CLICommand) {
	if isCloudLogin {
		cmd.Flags().String("environment", "", "Environment ID for scope of role-binding operation.")
		cmd.Flags().Bool("current-env", false, "Use current environment ID for scope.")
		cmd.Flags().String("cloud-cluster", "", "Cloud cluster ID for the role binding.")
		cmd.Flags().String("kafka-cluster-id", "", "Kafka cluster ID for the role binding.")
		if os.Getenv("XX_DATAPLANE_3_ENABLE") != "" {
			cmd.Flags().String("schema-registry-cluster-id", "", "Schema Registry cluster ID for the role binding.")
			cmd.Flags().String("ksql-cluster-id", "", "ksqlDB cluster ID for the role binding.")
		}
	} else {
		cmd.Flags().String("kafka-cluster-id", "", "Kafka cluster ID for the role binding.")
		cmd.Flags().String("schema-registry-cluster-id", "", "Schema Registry cluster ID for the role binding.")
		cmd.Flags().String("ksql-cluster-id", "", "ksqlDB cluster ID for the role binding.")
		cmd.Flags().String("connect-cluster-id", "", "Kafka Connect cluster ID for the role binding.")
		cmd.Flags().String("cluster-name", "", "Cluster name to uniquely identify the cluster for role binding listings.")
		pcmd.AddContextFlag(cmd, cliCommand)
	}
}

func (c *roleBindingCommand) validatePrincipalFormat(principal string) error {
	if len(strings.Split(principal, ":")) == 1 {
		return errors.NewErrorWithSuggestions(errors.PrincipalFormatErrorMsg, errors.PrincipalFormatSuggestions)
	}

	return nil
}

func (c *roleBindingCommand) parseAndValidateScope(cmd *cobra.Command) (*mds.MdsScope, error) {
	scope := &mds.MdsScopeClusters{}
	nonKafkaScopesSet := 0

	clusterName, err := cmd.Flags().GetString("cluster-name")
	if err != nil {
		return nil, err
	}

	cmd.Flags().Visit(func(flag *pflag.Flag) {
		switch flag.Name {
		case "kafka-cluster-id":
			scope.KafkaCluster = flag.Value.String()
		case "schema-registry-cluster-id":
			scope.SchemaRegistryCluster = flag.Value.String()
			nonKafkaScopesSet++
		case "ksql-cluster-id":
			scope.KsqlCluster = flag.Value.String()
			nonKafkaScopesSet++
		case "connect-cluster-id":
			scope.ConnectCluster = flag.Value.String()
			nonKafkaScopesSet++
		}
	})

	if clusterName != "" && (scope.KafkaCluster != "" || nonKafkaScopesSet > 0) {
		return nil, errors.New(errors.BothClusterNameAndScopeErrorMsg)
	}

	if clusterName == "" {
		if scope.KafkaCluster == "" && nonKafkaScopesSet > 0 {
			return nil, errors.New(errors.SpecifyKafkaIDErrorMsg)
		}

		if scope.KafkaCluster == "" && nonKafkaScopesSet == 0 {
			return nil, errors.New(errors.SpecifyClusterErrorMsg)
		}

		if nonKafkaScopesSet > 1 {
			return nil, errors.New(errors.MoreThanOneNonKafkaErrorMsg)
		}

		return &mds.MdsScope{Clusters: *scope}, nil
	}

	return &mds.MdsScope{ClusterName: clusterName}, nil
}

func (c *roleBindingCommand) parseAndValidateScopeV2(cmd *cobra.Command) (*mdsv2alpha1.Scope, error) {
	scopeV2 := &mdsv2alpha1.Scope{}
	orgResourceId := c.State.Auth.Organization.GetResourceId()
	scopeV2.Path = []string{"organization=" + orgResourceId}

	if cmd.Flags().Changed("current-env") {
		scopeV2.Path = append(scopeV2.Path, "environment="+c.EnvironmentId())
	} else if cmd.Flags().Changed("environment") {
		env, err := cmd.Flags().GetString("environment")
		if err != nil {
			return nil, err
		}
		scopeV2.Path = append(scopeV2.Path, "environment="+env)
	}

	if cmd.Flags().Changed("cloud-cluster") {
		cluster, err := cmd.Flags().GetString("cloud-cluster")
		if err != nil {
			return nil, err
		}
		scopeV2.Path = append(scopeV2.Path, "cloud-cluster="+cluster)
	}

	if cmd.Flags().Changed("kafka-cluster-id") {
		kafkaCluster, err := cmd.Flags().GetString("kafka-cluster-id")
		if err != nil {
			return nil, err
		}
		scopeV2.Clusters.KafkaCluster = kafkaCluster
	}

	if cmd.Flags().Changed("schema-registry-cluster-id") {
		srCluster, err := cmd.Flags().GetString("schema-registry-cluster-id")
		if err != nil {
			return nil, err
		}
		scopeV2.Clusters.SchemaRegistryCluster = srCluster
	}

	if cmd.Flags().Changed("ksql-cluster-id") {
		ksqlCluster, err := cmd.Flags().GetString("ksql-cluster-id")
		if err != nil {
			return nil, err
		}
		scopeV2.Clusters.KsqlCluster = ksqlCluster
	}

	if cmd.Flags().Changed("role") {
		role, err := cmd.Flags().GetString("role")
		if err != nil {
			return nil, err
		}
		if clusterScopedRolesV2[role] && !cmd.Flags().Changed("cloud-cluster") {
			return nil, errors.New(errors.SpecifyCloudClusterErrorMsg)
		}
		if (environmentScopedRoles[role] || clusterScopedRolesV2[role]) && !cmd.Flags().Changed("current-env") && !cmd.Flags().Changed("environment") {
			return nil, errors.New(errors.SpecifyEnvironmentErrorMsg)
		}
	}

	if cmd.Flags().Changed("cloud-cluster") && !cmd.Flags().Changed("current-env") && !cmd.Flags().Changed("environment") {
		return nil, errors.New(errors.SpecifyEnvironmentErrorMsg)
	}
	return scopeV2, nil
}

func parseAndValidateResourcePatternV2(resource string, prefix bool) (mdsv2alpha1.ResourcePattern, error) {
	var result mdsv2alpha1.ResourcePattern
	if prefix {
		result.PatternType = "PREFIXED"
	} else {
		result.PatternType = "LITERAL"
	}

	parts := strings.SplitN(resource, ":", 2)
	if len(parts) != 2 {
		return result, errors.NewErrorWithSuggestions(errors.ResourceFormatErrorMsg, errors.ResourceFormatSuggestions)
	}
	result.ResourceType = parts[0]
	result.Name = parts[1]

	return result, nil
}

func (c *roleBindingCommand) validateRoleAndResourceTypeV2(roleName string, resourceType string) error {
	ctx := c.createContext()
	opts := &mdsv2alpha1.RoleDetailOpts{Namespace: dataplaneNamespace}

	// Currently we don't allow multiple namespace in opts so as a workaround we first check with dataplane
	// namespace and if we get an error try without any namespace.
	role, resp, err := c.MDSv2Client.RBACRoleDefinitionsApi.RoleDetail(ctx, roleName, opts)
	if err != nil || resp.StatusCode == http.StatusNoContent {
		role, resp, err = c.MDSv2Client.RBACRoleDefinitionsApi.RoleDetail(ctx, roleName, nil)
		if err != nil || resp.StatusCode == http.StatusNoContent {
			if err == nil {
				return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.LookUpRoleErrorMsg, roleName), errors.LookUpRoleSuggestions)
			} else {
				return errors.NewWrapErrorWithSuggestions(err, fmt.Sprintf(errors.LookUpRoleErrorMsg, roleName), errors.LookUpRoleSuggestions)
			}
		}
	}

	var allResourceTypes []string
	for _, policies := range role.Policies {
		for _, operation := range policies.AllowedOperations {
			allResourceTypes = append(allResourceTypes, operation.ResourceType)
			if operation.ResourceType == resourceType {
				return nil
			}
		}
	}

	suggestionsMsg := fmt.Sprintf(errors.InvalidResourceTypeSuggestions, strings.Join(allResourceTypes, ", "))
	return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.InvalidResourceTypeErrorMsg, resourceType), suggestionsMsg)
}

func (c *roleBindingCommand) validateResourceTypeV2(resourceType string) error {
	ctx := c.createContext()
	roles, _, err := c.MDSv2Client.RBACRoleDefinitionsApi.Roles(ctx, nil)
	if err != nil {
		return err
	}

	var allResourceTypes = make(map[string]bool)
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
		suggestionsMsg := fmt.Sprintf(errors.InvalidResourceTypeSuggestions, strings.Join(uniqueResourceTypes, ", "))
		return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.InvalidResourceTypeErrorMsg, resourceType), suggestionsMsg)
	}

	return nil
}

func parseAndValidateResourcePattern(resource string, prefix bool) (mds.ResourcePattern, error) {
	var result mds.ResourcePattern
	if prefix {
		result.PatternType = "PREFIXED"
	} else {
		result.PatternType = "LITERAL"
	}

	parts := strings.SplitN(resource, ":", 2)
	if len(parts) != 2 {
		return result, errors.NewErrorWithSuggestions(errors.ResourceFormatErrorMsg, errors.ResourceFormatSuggestions)
	}
	result.ResourceType = parts[0]
	result.Name = parts[1]

	return result, nil
}

func (c *roleBindingCommand) validateRoleAndResourceTypeV1(roleName string, resourceType string) error {
	ctx := c.createContext()
	role, resp, err := c.MDSClient.RBACRoleDefinitionsApi.RoleDetail(ctx, roleName)
	if err != nil || resp.StatusCode == 204 {
		if err == nil {
			return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.LookUpRoleErrorMsg, roleName), errors.LookUpRoleSuggestions)
		} else {
			return errors.NewWrapErrorWithSuggestions(err, fmt.Sprintf(errors.LookUpRoleErrorMsg, roleName), errors.LookUpRoleSuggestions)
		}
	}

	var allResourceTypes []string
	found := false
	for _, operation := range role.AccessPolicy.AllowedOperations {
		allResourceTypes = append(allResourceTypes, operation.ResourceType)
		if operation.ResourceType == resourceType {
			found = true
			break
		}
	}

	if !found {
		suggestionsMsg := fmt.Sprintf(errors.InvalidResourceTypeSuggestions, strings.Join(allResourceTypes, ", "))
		return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.InvalidResourceTypeErrorMsg, resourceType), suggestionsMsg)
	}

	return nil
}

func (c *roleBindingCommand) validateResourceTypeV1(resourceType string) error {
	ctx := c.createContext()
	roles, _, err := c.MDSClient.RBACRoleDefinitionsApi.Roles(ctx)
	if err != nil {
		return err
	}

	var allResourceTypes = make(map[string]bool)
	found := false
	for _, role := range roles {
		for _, operation := range role.AccessPolicy.AllowedOperations {
			allResourceTypes[operation.ResourceType] = true
			if operation.ResourceType == resourceType {
				found = true
				break
			}
		}
	}

	if !found {
		uniqueResourceTypes := []string{}
		for rt := range allResourceTypes {
			uniqueResourceTypes = append(uniqueResourceTypes, rt)
		}
		suggestionsMsg := fmt.Sprintf(errors.InvalidResourceTypeSuggestions, strings.Join(uniqueResourceTypes, ", "))
		return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.InvalidResourceTypeErrorMsg, resourceType), suggestionsMsg)
	}

	return nil
}

func (c *roleBindingCommand) displayCCloudCreateAndDeleteOutput(cmd *cobra.Command, options *roleBindingOptions) error {
	var fieldsSelected []string
	structuredRename := map[string]string{"Principal": "principal", "Email": "email", "Role": "role", "ResourceType": "resource_type", "Name": "name", "PatternType": "pattern_type"}
	userResourceId := strings.TrimLeft(options.principal, "User:")
	user, err := c.Client.User.Describe(context.Background(), &orgv1.User{ResourceId: userResourceId})
	displayStruct := &listDisplay{
		Principal: options.principal,
		Role:      options.role,
	}

	if options.resource != "" {
		if len(options.resourcesRequestV2.ResourcePatterns) != 1 {
			return errors.New("display error: number of resource pattern is not 1")
		}
		resourcePattern := options.resourcesRequestV2.ResourcePatterns[0]
		displayStruct.ResourceType = resourcePattern.ResourceType
		displayStruct.Name = resourcePattern.Name
		displayStruct.PatternType = resourcePattern.PatternType
	}

	if err != nil {
		if options.resource != "" {
			fieldsSelected = resourcePatternListFields
		} else {
			fieldsSelected = []string{"Principal", "Role"}
		}
	} else {
		if options.resource != "" {
			fieldsSelected = ccloudResourcePatternListFields
		} else {
			displayStruct.Email = user.Email
			fieldsSelected = []string{"Principal", "Email", "Role"}
		}
	}
	return output.DescribeObject(cmd, displayStruct, fieldsSelected, map[string]string{}, structuredRename)
}

func displayCreateAndDeleteOutput(cmd *cobra.Command, options *roleBindingOptions) error {
	var fieldsSelected []string

	displayStruct := &listDisplay{
		Principal: options.principal,
		Role:      options.role,
	}
	if options.resource != "" {
		fieldsSelected = resourcePatternListFields
		if len(options.resourcesRequest.ResourcePatterns) != 1 {
			return errors.New("display error: number of resource pattern is not 1")
		}
		resourcePattern := options.resourcesRequest.ResourcePatterns[0]
		displayStruct.ResourceType = resourcePattern.ResourceType
		displayStruct.Name = resourcePattern.Name
		displayStruct.PatternType = resourcePattern.PatternType
	} else {
		fieldsSelected = []string{"Principal", "Role", "ResourceType"}
		displayStruct.ResourceType = "Cluster"
	}

	humanRenames := map[string]string{"ResourceType": "Resource Type", "PatternType": "Pattern Type"}
	structuredRenames := map[string]string{"Principal": "principal", "Role": "role", "ResourceType": "resource_type", "Name": "name", "PatternType": "pattern_type"}

	return output.DescribeObject(cmd, displayStruct, fieldsSelected, humanRenames, structuredRenames)
}

func (c *roleBindingCommand) createContext() context.Context {
	if c.cfg.IsCloudLogin() {
		return context.WithValue(context.Background(), mdsv2alpha1.ContextAccessToken, c.AuthToken())
	} else {
		return context.WithValue(context.Background(), mds.ContextAccessToken, c.AuthToken())
	}
}

func (c *roleBindingCommand) parseRoleBinding(cmd *cobra.Command) (*mdsv2.IamV2RoleBinding, error) {
	// include org, environment, cloud-cluster, and resource name and id
	role, err := cmd.Flags().GetString("role")
	if err != nil {
		return nil, err
	}

	resource, err := cmd.Flags().GetString("resource")
	if err != nil {
		return nil, err
	}

	prefix, _ := cmd.Flags().GetBool("prefix")

	principal, err := cmd.Flags().GetString("principal")
	if err != nil {
		return nil, err
	}

	if strings.HasPrefix(principal, "User:") {
		principalValue := strings.TrimLeft(principal, "User:")
		if strings.Contains(principalValue, "@") {
			user, err := c.V2Client.GetIamUserByEmail(principalValue)
			if err != nil {
				return nil, err
			}
			principal = "User:" + *user.Id
		}
	}

	if cmd.Flags().Changed("principal") {
		err = c.validatePrincipalFormat(principal)
		if err != nil {
			return nil, err
		}
	}

	// crnPattern := "crn://confluent.cloud"
	crnPattern := parseCrnPattern()
	return &mdsv2.IamV2RoleBinding{
		Principal:  mdsv2.PtrString(principal),
		RoleName:   mdsv2.PtrString(role),
		CrnPattern: crnPattern,
	}
}
