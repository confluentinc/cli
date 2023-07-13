package cmd

import (
	"github.com/confluentinc/cli/internal/pkg/errors"
	"net/url"
	"strings"

	"github.com/spf13/cobra"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"
	mds "github.com/confluentinc/mds-sdk-go-public/mdsv1"
	"github.com/confluentinc/mds-sdk-go-public/mdsv2alpha1"

	"github.com/confluentinc/cli/internal/pkg/auth"
	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	dynamicconfig "github.com/confluentinc/cli/internal/pkg/dynamic-config"
	"github.com/confluentinc/cli/internal/pkg/hub"
	testserver "github.com/confluentinc/cli/test/test-server"
)

type AuthenticatedCLICommand struct {
	*CLICommand

	Client             *ccloudv1.Client
	HubClient          *hub.Client
	KafkaRESTProvider  *KafkaRESTProvider
	MDSClient          *mds.APIClient
	MDSv2Client        *mdsv2alpha1.APIClient
	V2Client           *ccloudv2.Client
	flinkGatewayClient *ccloudv2.FlinkGatewayClient
	metricsClient      *ccloudv2.MetricsClient

	Context *dynamicconfig.DynamicContext
	State   *v1.ContextState
}

func NewAuthenticatedCLICommand(cmd *cobra.Command, prerunner PreRunner) *AuthenticatedCLICommand {
	c := &AuthenticatedCLICommand{CLICommand: NewCLICommand(cmd)}
	cmd.PersistentPreRunE = Chain(prerunner.Authenticated(c), prerunner.ParseFlagsIntoContext(c))
	return c
}

func NewAuthenticatedWithMDSCLICommand(cmd *cobra.Command, prerunner PreRunner) *AuthenticatedCLICommand {
	c := &AuthenticatedCLICommand{CLICommand: NewCLICommand(cmd)}
	cmd.PersistentPreRunE = Chain(prerunner.AuthenticatedWithMDS(c), prerunner.ParseFlagsIntoContext(c))
	return c
}

func (c *AuthenticatedCLICommand) GetFlinkGatewayClient() (*ccloudv2.FlinkGatewayClient, error) {
	ctx := c.Config.Context()

	if c.flinkGatewayClient == nil {
		var gatewayUrl string
		var err error
		computePoolId := ctx.GetCurrentFlinkComputePool()
		if computePoolId != "" {
			gatewayUrl, err = c.getGatewayUrlForComputePool(computePoolId, ctx)
			if err != nil {
				return nil, err
			}
		} else if ctx.GetCurrentFlinkRegion() != "" && ctx.GetCurrentFlinkCloudProvider() != "" {
			gatewayUrl, err = c.getGatewayUrlForRegion(ctx.GetCurrentFlinkCloudProvider(), ctx.GetCurrentFlinkRegion())
			if err != nil {
				return nil, err
			}
		} else {
			return nil, errors.NewErrorWithSuggestions("no compute pool or cloud provider and region selected", "Select a compute pool with `confluent flink compute-pool use` or `--compute-pool`. Alternatively you can also select a cloud provider and region with `--cloud` and `--region`")
		}

		unsafeTrace, err := c.Command.Flags().GetBool("unsafe-trace")
		if err != nil {
			return nil, err
		}

		dataplaneToken, err := auth.GetDataplaneToken(ctx.GetState(), ctx.GetPlatformServer())
		if err != nil {
			return nil, err
		}

		c.flinkGatewayClient = ccloudv2.NewFlinkGatewayClient(gatewayUrl, c.Version.UserAgent, unsafeTrace, dataplaneToken)
	}

	return c.flinkGatewayClient, nil
}

func (c *AuthenticatedCLICommand) getGatewayUrlForComputePool(computePoolId string, ctx *dynamicconfig.DynamicContext) (string, error) {
	computePool, err := c.V2Client.DescribeFlinkComputePool(computePoolId, ctx.GetCurrentEnvironment())
	if err != nil {
		return "", err
	}

	u, err := url.Parse(computePool.Spec.GetHttpEndpoint())
	if err != nil {
		return "", err
	}
	u.Path = ""
	return u.String(), nil
}

func (c *AuthenticatedCLICommand) getGatewayUrlForRegion(provider, region string) (string, error) {
	regions, err := c.V2Client.ListFlinkRegions(provider)
	var hostUrl string
	for _, flinkRegion := range regions {
		if flinkRegion.GetRegionName() == region {
			hostUrl = flinkRegion.GetHttpEndpoint()
			break
		}
	}
	if hostUrl == "" {
		return "", errors.NewErrorWithSuggestions("invalid region", "Please select a valid region - use `confluent flink region list` to see available regions")
	}

	u, err := url.Parse(hostUrl)
	if err != nil {
		return "", err
	}
	u.Path = ""

	return u.String(), nil
}

func (c *AuthenticatedCLICommand) GetHubClient() (*hub.Client, error) {
	if c.HubClient == nil {
		unsafeTrace, err := c.Flags().GetBool("unsafe-trace")
		if err != nil {
			return nil, err
		}

		c.HubClient = hub.NewClient(c.Config.Version.UserAgent, c.Config.IsTest, unsafeTrace)
	}

	return c.HubClient, nil
}

func (c *AuthenticatedCLICommand) GetKafkaREST() (*KafkaREST, error) {
	return (*c.KafkaRESTProvider)()
}

func (c *AuthenticatedCLICommand) GetMetricsClient() (*ccloudv2.MetricsClient, error) {
	ctx := c.Config.Context()

	if c.metricsClient == nil {
		url := "https://api.telemetry.confluent.cloud"
		if c.Config.IsTest {
			url = testserver.TestV2CloudUrl.String()
		} else if strings.Contains(ctx.GetPlatformServer(), "devel") {
			url = "https://devel-sandbox-api.telemetry.aws.confluent.cloud"
		} else if strings.Contains(ctx.GetPlatformServer(), "stag") {
			url = "https://stag-sandbox-api.telemetry.aws.confluent.cloud"
		}

		unsafeTrace, err := c.Command.Flags().GetBool("unsafe-trace")
		if err != nil {
			return nil, err
		}

		dataplaneToken, err := auth.GetDataplaneToken(ctx.GetState(), ctx.GetPlatformServer())
		if err != nil {
			return nil, err
		}

		c.metricsClient = ccloudv2.NewMetricsClient(url, c.Version.UserAgent, unsafeTrace, dataplaneToken)
	}

	return c.metricsClient, nil
}

func (c *AuthenticatedCLICommand) AuthToken() string {
	return c.Context.GetAuthToken()
}
