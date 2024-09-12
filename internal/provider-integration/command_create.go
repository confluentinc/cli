package providerintegration

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	providerintegrationv1 "github.com/confluentinc/ccloud-sdk-go-v2/provider-integration/v1"

	"github.com/confluentinc/cli/v3/pkg/ccloudv2"
	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/publiccloud"
	"github.com/confluentinc/cli/v3/pkg/utils"
)

func (c *command) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a provider integration.",
		Long:  "Create a Provider Integration that allow users to manage access to public cloud service provider resources through Confluent resources.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.create,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Create a provider integration "s3-provider-integration" associated with AWS IAM role arn "arn:aws:iam::000000000000:role/my-test-aws-role in current environment".`,
				Code: "confluent provider-integration create s3-provider-integration --cloud aws --customer-role-arn arn:aws:iam::000000000000:role/my-test-aws-role",
			},
			examples.Example{
				Text: `Create a provider integration "s3-provider-integration" associated with AWS IAM role arn "arn:aws:iam::000000000000:role/my-test-aws-role in environment env-abcdef".`,
				Code: "confluent provider-integration create s3-provider-integration --cloud aws --customer-role-arn arn:aws:iam::000000000000:role/my-test-aws-role --environment env-abcdef",
			},
		),
	}

	// Handle the flags, for cloud flag only AWS is supported now
	cmd.Flags().String("customer-role-arn", "", "Amazon Resource Name (ARN) that identifies the AWS Identity and Access Management (IAM) role that Confluent Cloud assumes when it accesses resources in your AWS account, having to be unique in the same environment.")
	c.addCloudFlag(cmd)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	// Check the argument from flags
	cobra.CheckErr(cmd.MarkFlagRequired("cloud"))
	cobra.CheckErr(cmd.MarkFlagRequired("customer-role-arn"))

	return cmd
}

func (c *command) create(cmd *cobra.Command, args []string) error {
	name := args[0]

	cloud, err := cmd.Flags().GetString("cloud")
	if err != nil {
		return err
	}

	customerIamRoleArn, err := cmd.Flags().GetString("customer-role-arn")
	if err != nil {
		return err
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	kind, err := getProviderConfigKind(cloud)
	if err != nil {
		return err
	}

	// Populate the PimV1Integration request object
	request := &providerintegrationv1.PimV1Integration{
		DisplayName: providerintegrationv1.PtrString(name),
		Provider:    providerintegrationv1.PtrString(cloud),
		Environment: &providerintegrationv1.GlobalObjectReference{Id: environmentId},
		Config: &providerintegrationv1.PimV1IntegrationConfigOneOf{
			PimV1AwsIntegrationConfig: &providerintegrationv1.PimV1AwsIntegrationConfig{
				CustomerIamRoleArn: providerintegrationv1.PtrString(customerIamRoleArn),
				Kind:               kind,
			},
		},
	}

	providerIntegration, err := c.V2Client.CreateProviderIntegration(*request)
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)

	resp := providerIntegrationOut{
		Id:                 providerIntegration.GetId(),
		Name:               providerIntegration.GetDisplayName(),
		Provider:           providerIntegration.GetProvider(),
		Environment:        providerIntegration.Environment.GetId(),
		IamRoleArn:         providerIntegration.GetConfig().PimV1AwsIntegrationConfig.GetIamRoleArn(),
		ExternalId:         providerIntegration.GetConfig().PimV1AwsIntegrationConfig.GetExternalId(),
		CustomerIamRoleArn: providerIntegration.GetConfig().PimV1AwsIntegrationConfig.GetCustomerIamRoleArn(),
	}

	// `PimV1Integration.Usages` field is empty after create() and this field should be hidden
	table.Add(&resp)
	table.Filter([]string{"Id", "Name", "Provider", "Environment", "IamRoleArn", "ExternalId", "CustomerIamRoleArn"})
	return table.Print()
}

func getProviderConfigKind(provider string) (string, error) {
	switch strings.ToUpper(provider) {
	case publiccloud.CloudAws:
		return "AwsIntegrationConfig", nil
	default:
		return "", fmt.Errorf(`cloud provider "%s" is not supported`, provider)
	}
}

func (c *command) addCloudFlag(cmd *cobra.Command) {
	cmd.Flags().String("cloud", "", fmt.Sprintf("Specify the cloud provider as %s.", utils.ArrayToCommaDelimitedString(ccloudv2.ProviderIntegrationSupportClouds, "or")))
	pcmd.RegisterFlagCompletionFunc(cmd, "cloud", func(_ *cobra.Command, _ []string) []string { return ccloudv2.ProviderIntegrationSupportClouds })
}
