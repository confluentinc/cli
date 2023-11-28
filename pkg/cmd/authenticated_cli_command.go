package cmd

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"
	metricsv2 "github.com/confluentinc/ccloud-sdk-go-v2/metrics/v2"
	"github.com/confluentinc/mds-sdk-go-public/mdsv1"
	"github.com/confluentinc/mds-sdk-go-public/mdsv2alpha1"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	"github.com/confluentinc/cli/v3/pkg/auth"
	"github.com/confluentinc/cli/v3/pkg/ccloudv2"
	"github.com/confluentinc/cli/v3/pkg/config"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/hub"
	schemaregistry "github.com/confluentinc/cli/v3/pkg/schema-registry"
	"github.com/confluentinc/cli/v3/pkg/utils"
	testserver "github.com/confluentinc/cli/v3/test/test-server"
)

type AuthenticatedCLICommand struct {
	*CLICommand

	Client            *ccloudv1.Client
	KafkaRESTProvider *KafkaRESTProvider
	MDSClient         *mdsv1.APIClient
	MDSv2Client       *mdsv2alpha1.APIClient
	V2Client          *ccloudv2.Client

	flinkGatewayClient   *ccloudv2.FlinkGatewayClient
	hubClient            *hub.Client
	metricsClient        *ccloudv2.MetricsClient
	schemaRegistryClient *schemaregistry.Client

	Context *config.Context
}

func NewAuthenticatedCLICommand(cmd *cobra.Command, prerunner PreRunner) *AuthenticatedCLICommand {
	c := &AuthenticatedCLICommand{CLICommand: NewCLICommand(cmd)}
	cmd.PersistentPreRunE = chain(prerunner.Authenticated(c), prerunner.ParseFlagsIntoContext(c.CLICommand))
	return c
}

func NewAuthenticatedWithMDSCLICommand(cmd *cobra.Command, prerunner PreRunner) *AuthenticatedCLICommand {
	c := &AuthenticatedCLICommand{CLICommand: NewCLICommand(cmd)}
	cmd.PersistentPreRunE = chain(prerunner.AuthenticatedWithMDS(c), prerunner.ParseFlagsIntoContext(c.CLICommand))
	return c
}

func (c *AuthenticatedCLICommand) GetFlinkGatewayClient(computePoolOnly bool) (*ccloudv2.FlinkGatewayClient, error) {
	if c.flinkGatewayClient == nil {
		var url string
		var err error

		if computePoolOnly {
			if computePoolId := c.Context.GetCurrentFlinkComputePool(); computePoolId != "" {
				url, err = c.getGatewayUrlForComputePool(computePoolId, c.Context)
				if err != nil {
					return nil, err
				}
			} else {
				return nil, errors.NewErrorWithSuggestions("no compute pool selected", "Select a compute pool with `confluent flink compute-pool use` or `--compute-pool`.")
			}
		} else if c.Context.GetCurrentFlinkRegion() != "" && c.Context.GetCurrentFlinkCloudProvider() != "" {
			url, err = c.getGatewayUrlForRegion(c.Context.GetCurrentFlinkCloudProvider(), c.Context.GetCurrentFlinkRegion())
			if err != nil {
				return nil, err
			}
		} else {
			return nil, errors.NewErrorWithSuggestions("no cloud provider and region selected", "Select a cloud provider and region with `confluent flink region use` or `--cloud` and `--region`.")
		}

		unsafeTrace, err := c.Flags().GetBool("unsafe-trace")
		if err != nil {
			return nil, err
		}

		dataplaneToken, err := auth.GetDataplaneToken(c.Context)
		if err != nil {
			return nil, err
		}

		c.flinkGatewayClient = ccloudv2.NewFlinkGatewayClient(url, c.Version.UserAgent, unsafeTrace, dataplaneToken)
	}

	return c.flinkGatewayClient, nil
}

func (c *AuthenticatedCLICommand) getGatewayUrlForComputePool(computePoolId string, ctx *config.Context) (string, error) {
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
	if err != nil {
		return "", err
	}

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
	if c.hubClient == nil {
		unsafeTrace, err := c.Flags().GetBool("unsafe-trace")
		if err != nil {
			return nil, err
		}

		c.hubClient = hub.NewClient(c.Config.Version.UserAgent, c.Config.IsTest, unsafeTrace)
	}

	return c.hubClient, nil
}

func (c *AuthenticatedCLICommand) GetKafkaREST() (*KafkaREST, error) {
	return (*c.KafkaRESTProvider)()
}

func (c *AuthenticatedCLICommand) GetMetricsClient() (*ccloudv2.MetricsClient, error) {
	if c.metricsClient == nil {
		unsafeTrace, err := c.Flags().GetBool("unsafe-trace")
		if err != nil {
			return nil, err
		}

		url := "https://api.telemetry.confluent.cloud"
		if c.Config.IsTest {
			url = testserver.TestV2CloudUrl.String()
		} else if strings.Contains(c.Context.GetPlatformServer(), "devel") {
			url = "https://devel-sandbox-api.telemetry.aws.confluent.cloud"
		} else if strings.Contains(c.Context.GetPlatformServer(), "stag") {
			url = "https://stag-sandbox-api.telemetry.aws.confluent.cloud"
		}

		configuration := metricsv2.NewConfiguration()
		configuration.Debug = unsafeTrace
		configuration.HTTPClient = ccloudv2.NewRetryableHttpClient(c.Config, unsafeTrace)
		configuration.Servers = metricsv2.ServerConfigurations{{URL: url}}
		configuration.UserAgent = c.Config.Version.UserAgent

		c.metricsClient = ccloudv2.NewMetricsClient(configuration, c.Config)
	}

	return c.metricsClient, nil
}

func (c *AuthenticatedCLICommand) GetSchemaRegistryClient(cmd *cobra.Command) (*schemaregistry.Client, error) {
	if c.schemaRegistryClient == nil {
		unsafeTrace, err := cmd.Flags().GetBool("unsafe-trace")
		if err != nil {
			return nil, err
		}

		configuration := srsdk.NewConfiguration()
		configuration.UserAgent = c.Config.Version.UserAgent
		configuration.Debug = unsafeTrace
		configuration.HTTPClient = ccloudv2.NewRetryableHttpClient(c.Config, unsafeTrace)

		schemaRegistryEndpoint, _ := cmd.Flags().GetString("schema-registry-endpoint")
		if schemaRegistryEndpoint != "" {
			u, err := url.Parse(schemaRegistryEndpoint)
			if err != nil {
				return nil, err
			}
			if u.Scheme != "http" && u.Scheme != "https" {
				u.Scheme = "https"
			}
			configuration.BasePath = u.String()

			caLocation, err := cmd.Flags().GetString("ca-location")
			if err != nil {
				return nil, err
			}
			if caLocation == "" {
				caLocation = auth.GetEnvWithFallback(auth.ConfluentPlatformCACertPath, auth.DeprecatedConfluentPlatformCACertPath)
			}
			if caLocation != "" {
				caClient, err := utils.GetCAClient(caLocation)
				if err != nil {
					return nil, err
				}
				configuration.HTTPClient = caClient
			}
		} else if c.Config.IsCloudLogin() {
			clusters, err := c.V2Client.GetSchemaRegistryClustersByEnvironment(c.Context.GetCurrentEnvironment())
			if err != nil {
				return nil, err
			}
			if len(clusters) == 0 {
				return nil, errors.NewSRNotEnabledError()
			}
			configuration.BasePath = clusters[0].Spec.GetHttpEndpoint()
			configuration.DefaultHeader = map[string]string{"target-sr-cluster": clusters[0].GetId()}
		} else {
			return nil, errors.NewErrorWithSuggestions(
				"Schema Registry endpoint not found",
				"Log in to Confluent Cloud with `confluent login`.\nSupply a Schema Registry endpoint with `--schema-registry-endpoint`.",
			)
		}

		schemaRegistryApiKey, _ := cmd.Flags().GetString("schema-registry-api-key")
		schemaRegistryApiSecret, _ := cmd.Flags().GetString("schema-registry-api-secret")

		if schemaRegistryApiKey != "" && schemaRegistryApiSecret != "" {
			apiKey := srsdk.BasicAuth{
				UserName: schemaRegistryApiKey,
				Password: schemaRegistryApiSecret,
			}
			c.schemaRegistryClient = schemaregistry.NewClientWithApiKey(configuration, apiKey)
		} else {
			c.schemaRegistryClient = schemaregistry.NewClient(configuration, c.Config)
		}

		if err := c.schemaRegistryClient.Get(); err != nil {
			return nil, fmt.Errorf("failed to validate Schema Registry client: %w", err)
		}
	}

	return c.schemaRegistryClient, nil
}
