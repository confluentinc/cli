package providerintegration

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	providerintegrationv1 "github.com/confluentinc/ccloud-sdk-go-v2/provider-integration/v1"

	"github.com/confluentinc/cli/v4/pkg/ccloudv2"
	pcloud "github.com/confluentinc/cli/v4/pkg/cloud"
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/utils"
)

func (c *command) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a provider integration.",
		Long:  "Create a provider integration that allows users to manage access to public cloud service provider resources through Confluent resources.\n\n⚠️  DEPRECATION NOTICE: This command will be deprecated in Q4 2025. Use 'confluent provider-integration v2 create' for new integrations.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.create,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Create provider integration "s3-provider-integration" associated with AWS IAM role ARN "arn:aws:iam::000000000000:role/my-test-aws-role" in the current environment.`,
				Code: "confluent provider-integration create s3-provider-integration --cloud aws --customer-role-arn arn:aws:iam::000000000000:role/my-test-aws-role",
			},
			examples.Example{
				Text: `Create provider integration "s3-provider-integration" associated with AWS IAM role ARN "arn:aws:iam::000000000000:role/my-test-aws-role" in environment "env-abcdef".`,
				Code: "confluent provider-integration create s3-provider-integration --cloud aws --customer-role-arn arn:aws:iam::000000000000:role/my-test-aws-role --environment env-abcdef",
			},
		),
	}

	cmd.Flags().String("customer-role-arn", "", "Amazon Resource Name (ARN) that identifies the AWS Identity and Access Management (IAM) role that Confluent Cloud assumes when it accesses resources in your AWS account, and must be unique in the same environment.")
	c.addCloudFlag(cmd)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

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

	customerRoleArn, err := cmd.Flags().GetString("customer-role-arn")
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

	request := &providerintegrationv1.PimV1Integration{
		DisplayName: providerintegrationv1.PtrString(name),
		Provider:    providerintegrationv1.PtrString(cloud),
		Environment: &providerintegrationv1.GlobalObjectReference{Id: environmentId},
		Config: &providerintegrationv1.PimV1IntegrationConfigOneOf{
			PimV1AwsIntegrationConfig: &providerintegrationv1.PimV1AwsIntegrationConfig{
				CustomerIamRoleArn: providerintegrationv1.PtrString(customerRoleArn),
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
		Id:              providerIntegration.GetId(),
		Name:            providerIntegration.GetDisplayName(),
		Provider:        providerIntegration.GetProvider(),
		Environment:     providerIntegration.Environment.GetId(),
		IamRoleArn:      providerIntegration.GetConfig().PimV1AwsIntegrationConfig.GetIamRoleArn(),
		ExternalId:      providerIntegration.GetConfig().PimV1AwsIntegrationConfig.GetExternalId(),
		CustomerRoleArn: providerIntegration.GetConfig().PimV1AwsIntegrationConfig.GetCustomerIamRoleArn(),
	}

	// `PimV1Integration.Usages` field is empty after create() and this field should be hidden
	table.Add(&resp)
	table.Filter([]string{"Id", "Name", "Provider", "Environment", "IamRoleArn", "ExternalId", "CustomerRoleArn"})
	if err := table.Print(); err != nil {
		return err
	}

	// Add deprecation warning
	cmd.Println("\n⚠️  DEPRECATION NOTICE:")
	cmd.Println("This provider integration resource will be deprecated in Q4 2025.")
	cmd.Println("Please prepare to upgrade to the new provider integration v2 resource when available:")
	cmd.Println("  confluent provider-integration v2 create --help")
	
	return nil
}

func getProviderConfigKind(provider string) (string, error) {
	switch strings.ToUpper(provider) {
	case pcloud.Aws:
		return "AwsIntegrationConfig", nil
	default:
		return "", fmt.Errorf(`cloud provider "%s" is not supported`, provider)
	}
}

func (c *command) addCloudFlag(cmd *cobra.Command) {
	cmd.Flags().String("cloud", "", fmt.Sprintf("Specify the cloud provider as %s.", utils.ArrayToCommaDelimitedString(ccloudv2.ProviderIntegrationSupportClouds, "or")))
	pcmd.RegisterFlagCompletionFunc(cmd, "cloud", func(_ *cobra.Command, _ []string) []string { return ccloudv2.ProviderIntegrationSupportClouds })
}
