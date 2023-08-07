package cmd

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/spf13/cobra"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"
	mds "github.com/confluentinc/mds-sdk-go-public/mdsv1"
	"github.com/confluentinc/mds-sdk-go-public/mdsv2alpha1"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	"github.com/confluentinc/cli/internal/pkg/auth"
	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
	dynamicconfig "github.com/confluentinc/cli/internal/pkg/dynamic-config"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/hub"
	schemaregistry "github.com/confluentinc/cli/internal/pkg/schema-registry"
	"github.com/confluentinc/cli/internal/pkg/utils"
	testserver "github.com/confluentinc/cli/test/test-server"
)

type AuthenticatedCLICommand struct {
	*CLICommand

	Client            *ccloudv1.Client
	KafkaRESTProvider *KafkaRESTProvider
	MDSClient         *mds.APIClient
	MDSv2Client       *mdsv2alpha1.APIClient
	V2Client          *ccloudv2.Client

	flinkGatewayClient   *ccloudv2.FlinkGatewayClient
	hubClient            *hub.Client
	metricsClient        *ccloudv2.MetricsClient
	schemaRegistryClient *schemaregistry.Client

	Context *dynamicconfig.DynamicContext
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

func (c *AuthenticatedCLICommand) GetFlinkGatewayClient() (*ccloudv2.FlinkGatewayClient, error) {
	if c.flinkGatewayClient == nil {
		ctx := c.Config.Context()
		computePoolId := ctx.GetCurrentFlinkComputePool()

		var url string
		var err error

		if computePoolId != "" {
			url, err = c.getGatewayUrlForComputePool(computePoolId, ctx)
			if err != nil {
				return nil, err
			}
		} else if ctx.GetCurrentFlinkRegion() != "" && ctx.GetCurrentFlinkCloudProvider() != "" {
			url, err = c.getGatewayUrlForRegion(ctx.GetCurrentFlinkCloudProvider(), ctx.GetCurrentFlinkRegion())
			if err != nil {
				return nil, err
			}
		} else {
			return nil, errors.NewErrorWithSuggestions("no compute pool or cloud provider and region selected", "Select a compute pool with `confluent flink compute-pool use` or `--compute-pool`. Alternatively you can also select a cloud provider and region with `--cloud` and `--region`")
		}

		unsafeTrace, err := c.Flags().GetBool("unsafe-trace")
		if err != nil {
			return nil, err
		}

		dataplaneToken, err := auth.GetDataplaneToken(ctx.GetState(), ctx.GetPlatformServer())
		if err != nil {
			return nil, err
		}

		c.flinkGatewayClient = ccloudv2.NewFlinkGatewayClient(url, c.Version.UserAgent, unsafeTrace, dataplaneToken)
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
		ctx := c.Config.Context()

		url := "https://api.telemetry.confluent.cloud"
		if c.Config.IsTest {
			url = testserver.TestV2CloudUrl.String()
		} else if strings.Contains(ctx.GetPlatformServer(), "devel") {
			url = "https://devel-sandbox-api.telemetry.aws.confluent.cloud"
		} else if strings.Contains(ctx.GetPlatformServer(), "stag") {
			url = "https://stag-sandbox-api.telemetry.aws.confluent.cloud"
		}

		unsafeTrace, err := c.Flags().GetBool("unsafe-trace")
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

func (c *AuthenticatedCLICommand) GetSchemaRegistryClient(cmd *cobra.Command) (*schemaregistry.Client, error) {
	if c.schemaRegistryClient == nil {
		ctx := c.Config.Context()

		unsafeTrace, err := cmd.Flags().GetBool("unsafe-trace")
		if err != nil {
			return nil, err
		}

		if c.Config.IsCloudLogin() {
			configuration := srsdk.NewConfiguration()
			configuration.UserAgent = c.Config.Version.UserAgent
			configuration.Debug = unsafeTrace
			configuration.HTTPClient = ccloudv2.NewRetryableHttpClient(unsafeTrace)

			if ctx != nil && ctx.GetState() != nil {
				clusters, err := c.V2Client.GetSchemaRegistryClustersByEnvironment(ctx.GetCurrentEnvironment())
				if err != nil {
					return nil, err
				}
				if len(clusters) == 0 {
					return nil, errors.NewSRNotEnabledError()
				}
				configuration.DefaultHeader = map[string]string{"target-sr-cluster": clusters[0].GetId()}
				configuration.BasePath = clusters[0].Spec.GetHttpEndpoint()

				dataplaneToken, err := auth.GetDataplaneToken(ctx.GetState(), ctx.GetPlatformServer())
				if err != nil {
					return nil, err
				}

				c.schemaRegistryClient = schemaregistry.NewClientWithToken(configuration, dataplaneToken)
			} else {
				// Used by `asyncapi export`, `asyncapi import`, `kafka client-config create`, `kafka topic consume`, and `kafka topic produce`
				schemaRegistryEndpoint, err := cmd.Flags().GetString("schema-registry-endpoint")
				if err != nil {
					return nil, err
				}
				if schemaRegistryEndpoint == "" {
					return nil, errors.New(errors.SREndpointNotSpecifiedErrorMsg)
				}
				configuration.BasePath = schemaRegistryEndpoint

				schemaRegistryApiKey, err := cmd.Flags().GetString("schema-registry-api-key")
				if err != nil {
					return nil, err
				}
				schemaRegistryApiSecret, err := cmd.Flags().GetString("schema-registry-api-secret")
				if err != nil {
					return nil, err
				}

				c.schemaRegistryClient = schemaregistry.NewClientWithApiKey(configuration, schemaRegistryApiKey, schemaRegistryApiSecret)

				if err := c.schemaRegistryClient.Get(); err != nil {
					return nil, errors.New(errors.SRClientNotValidatedErrorMsg)
				}
			}
		} else {
			schemaRegistryEndpoint, err := cmd.Flags().GetString("schema-registry-endpoint")
			if err != nil {
				return nil, err
			}
			if schemaRegistryEndpoint == "" {
				return nil, errors.New(errors.SREndpointNotSpecifiedErrorMsg)
			}

			caLocation, err := cmd.Flags().GetString("ca-location")
			if err != nil {
				return nil, err
			}
			if caLocation == "" {
				caLocation = auth.GetEnvWithFallback(auth.ConfluentPlatformCACertPath, auth.DeprecatedConfluentPlatformCACertPath)
			}

			var client *http.Client
			if caLocation != "" {
				client, err = utils.GetCAClient(caLocation)
				if err != nil {
					return nil, err
				}
			} else {
				client = ccloudv2.NewRetryableHttpClient(unsafeTrace)
			}

			configuration := srsdk.NewConfiguration()
			configuration.BasePath = schemaRegistryEndpoint
			configuration.UserAgent = c.Config.Version.UserAgent
			configuration.Debug = unsafeTrace
			configuration.HTTPClient = client

			if ctx != nil && ctx.GetState() != nil {
				c.schemaRegistryClient = schemaregistry.NewClientWithToken(configuration, ctx.GetAuthToken())
			} else {
				c.schemaRegistryClient = schemaregistry.NewClient(configuration)
			}

			if err := c.schemaRegistryClient.Get(); err != nil {
				return nil, errors.New(errors.SRClientNotValidatedErrorMsg)
			}
		}
	}

	return c.schemaRegistryClient, nil
}
