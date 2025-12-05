package cmd

import (
	"crypto/tls"
	"fmt"
	purl "net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"
	metricsv2 "github.com/confluentinc/ccloud-sdk-go-v2/metrics/v2"
	srcmv3 "github.com/confluentinc/ccloud-sdk-go-v2/srcm/v3"
	"github.com/confluentinc/mds-sdk-go-public/mdsv1"
	"github.com/confluentinc/mds-sdk-go-public/mdsv2alpha1"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	"github.com/confluentinc/cli/v4/pkg/auth"
	"github.com/confluentinc/cli/v4/pkg/ccloudv2"
	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/kafkausagelimits"
	"github.com/confluentinc/cli/v4/pkg/log"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/schemaregistry"
	"github.com/confluentinc/cli/v4/pkg/utils"
	testserver "github.com/confluentinc/cli/v4/test/test-server"
)

type AuthenticatedCLICommand struct {
	*CLICommand

	Client            *ccloudv1.Client
	KafkaRESTProvider *KafkaRESTProvider
	MDSClient         *mdsv1.APIClient
	MDSv2Client       *mdsv2alpha1.APIClient
	V2Client          *ccloudv2.Client

	flinkGatewayClient   *ccloudv2.FlinkGatewayClient
	metricsClient        *ccloudv2.MetricsClient
	schemaRegistryClient *schemaregistry.Client
	usageLimitsClient    *kafkausagelimits.UsageLimitsClient

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

		if c.Context.GetCurrentFlinkEndpoint() != "" {
			url = c.Context.GetCurrentFlinkEndpoint()
		} else if computePoolOnly {
			if computePoolId := c.Context.GetCurrentFlinkComputePool(); computePoolId != "" {
				url, err = c.getGatewayUrlForComputePool(c.Context.GetCurrentFlinkAccessType(), computePoolId)
				if err != nil {
					return nil, err
				}
			} else {
				return nil, errors.NewErrorWithSuggestions("no compute pool selected", "Select a compute pool with `confluent flink compute-pool use` or `--compute-pool`.")
			}
		} else if c.Context.GetCurrentFlinkRegion() != "" && c.Context.GetCurrentFlinkCloudProvider() != "" {
			url, err = c.getGatewayUrlForRegion(c.Context.GetCurrentFlinkAccessType(), c.Context.GetCurrentFlinkCloudProvider(), c.Context.GetCurrentFlinkRegion())
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

		insecureSkipVerify, err := c.Flags().GetBool("insecure-skip-verify")
		if err != nil {
			return nil, err
		}

		dataplaneToken, err := auth.GetDataplaneToken(c.Context)
		if err != nil {
			return nil, err
		}

		log.CliLogger.Debugf("Insecure skip verify: %t\n", insecureSkipVerify)
		tlsClientConfig := &tls.Config{InsecureSkipVerify: insecureSkipVerify}

		log.CliLogger.Debugf("The final url used for setting up Flink dataplane client is: %s\n", url)
		c.flinkGatewayClient = ccloudv2.NewFlinkGatewayClient(url, c.Version.UserAgent, unsafeTrace, dataplaneToken, tlsClientConfig)
	}

	return c.flinkGatewayClient, nil
}

func (c *AuthenticatedCLICommand) getGatewayUrlForComputePool(access, id string) (string, error) {
	if c.Config.IsTest {
		return testserver.TestFlinkGatewayUrl.String(), nil
	}

	computePool, err := c.V2Client.DescribeFlinkComputePool(id, c.Context.GetCurrentEnvironment())
	if err != nil {
		return "", err
	}

	u, err := purl.Parse(c.Context.GetPlatformServer())
	if err != nil {
		return "", err
	}

	privateURL := fmt.Sprintf("https://flink.%s.%s.private.%s", computePool.Spec.GetRegion(), strings.ToLower(computePool.Spec.GetCloud()), u.Host)
	publicURL := fmt.Sprintf("https://flink.%s.%s.%s", computePool.Spec.GetRegion(), strings.ToLower(computePool.Spec.GetCloud()), u.Host)

	if access == "private" {
		return privateURL, nil
	}
	if access == "" {
		output.ErrPrintf(c.Config.EnableColor, "No Flink endpoint is specified, defaulting to public endpoint: `%s`\n", publicURL)
	}
	return publicURL, nil
}

func (c *AuthenticatedCLICommand) getGatewayUrlForRegion(accessType, provider, region string) (string, error) {
	regions, err := c.V2Client.ListFlinkRegions(provider, region)
	if err != nil {
		return "", err
	}

	var hostUrl string
	for _, flinkRegion := range regions {
		if flinkRegion.GetRegionName() == region {
			if accessType == "private" {
				hostUrl = flinkRegion.GetPrivateHttpEndpoint()
			} else {
				hostUrl = flinkRegion.GetHttpEndpoint()
			}
			break
		}
	}
	if hostUrl == "" {
		return "", errors.NewErrorWithSuggestions("invalid region", "Please select a valid region - use `confluent flink region list` to see available regions")
	}
	if accessType == "" {
		output.ErrPrintf(c.Config.EnableColor, "No Flink endpoint is specified, defaulting to public endpoint: `%s`\n", hostUrl)
	}

	u, err := purl.Parse(hostUrl)
	if err != nil {
		return "", err
	}
	u.Path = ""

	return u.String(), nil
}

func (c *AuthenticatedCLICommand) GetKafkaREST(cmd *cobra.Command) (*KafkaREST, error) {
	return (*c.KafkaRESTProvider)(cmd)
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
		schemaRegistryApiKey, _ := cmd.Flags().GetString("schema-registry-api-key")
		schemaRegistryApiSecret, _ := cmd.Flags().GetString("schema-registry-api-secret")
		isCloudEndpointAutoFound := false
		if c.Config.IsCloudLogin() {
			if schemaRegistryEndpoint != "" {
				clusters, err := c.V2Client.GetSchemaRegistryClustersByEnvironment(c.Context.GetCurrentEnvironment())
				if err != nil {
					return nil, err
				}
				if len(clusters) == 0 {
					return nil, schemaregistry.ErrNotEnabled
				}
				u, err := purl.Parse(schemaRegistryEndpoint)
				if err != nil {
					return nil, err
				}
				if u.Scheme != "http" && u.Scheme != "https" {
					u.Scheme = "https"
				}
				configuration.Servers = srsdk.ServerConfigurations{{URL: u.String()}}
				configuration.DefaultHeader = map[string]string{"target-sr-cluster": clusters[0].GetId()}
			} else if c.Context.GetSchemaRegistryEndpoint() != "" {
				configuration.Servers = srsdk.ServerConfigurations{{URL: c.Context.GetSchemaRegistryEndpoint()}}
				clusters, err := c.V2Client.GetSchemaRegistryClustersByEnvironment(c.Context.GetCurrentEnvironment())
				if err != nil {
					return nil, err
				}
				if len(clusters) == 0 {
					return nil, schemaregistry.ErrNotEnabled
				}
				configuration.DefaultHeader = map[string]string{"target-sr-cluster": clusters[0].GetId()}
			} else {
				clusters, err := c.V2Client.GetSchemaRegistryClustersByEnvironment(c.Context.GetCurrentEnvironment())
				if err != nil {
					return nil, err
				}
				if len(clusters) == 0 {
					return nil, schemaregistry.ErrNotEnabled
				}
				for _, urlPrivate := range clusters[0].Spec.PrivateNetworkingConfig.GetRegionalEndpoints() {
					isCloudEndpointAutoFound = c.validateSchemaRegistryEndpointAndSetClient(clusters[0], schemaRegistryApiKey, schemaRegistryApiSecret, urlPrivate, *configuration)
					if isCloudEndpointAutoFound {
						err := c.saveToConfig(urlPrivate)
						if err != nil {
							return nil, err
						}
						break
					}
				}
				if !isCloudEndpointAutoFound {
					isCloudEndpointAutoFound = c.validateSchemaRegistryEndpointAndSetClient(clusters[0], schemaRegistryApiKey, schemaRegistryApiSecret, clusters[0].Spec.GetHttpEndpoint(), *configuration)
					if isCloudEndpointAutoFound {
						err := c.saveToConfig(clusters[0].Spec.GetHttpEndpoint())
						if err != nil {
							return nil, err
						}
					}
				}
				if !isCloudEndpointAutoFound {
					return nil, errors.NewErrorWithSuggestions(
						"Schema Registry could not be reached. Check if Schema Registry is accessible.",
						"If Schema Registry is accessible, supply a Schema Registry endpoint with `--schema-registry-endpoint`.",
					)
				}
			}
		} else { //this is the on-prem case
			if schemaRegistryEndpoint != "" {
				u, err := purl.Parse(schemaRegistryEndpoint)
				if err != nil {
					return nil, err
				}
				if u.Scheme != "http" && u.Scheme != "https" {
					u.Scheme = "https"
				}
				configuration.Servers = srsdk.ServerConfigurations{{URL: u.String()}}

				certificateAuthorityPath, err := cmd.Flags().GetString("certificate-authority-path")
				if err != nil {
					return nil, err
				}
				clientCertificatePath, clientKeyPath, err := GetClientCertAndKeyPaths(cmd)
				if err != nil {
					return nil, err
				}
				if certificateAuthorityPath == "" {
					certificateAuthorityPath = os.Getenv(auth.ConfluentPlatformCertificateAuthorityPath)
				}
				if certificateAuthorityPath != "" {
					client, err := utils.GetCAAndClientCertClient(certificateAuthorityPath, clientCertificatePath, clientKeyPath)
					if err != nil {
						return nil, err
					}
					configuration.HTTPClient = client
				}
			} else {
				return nil, errors.NewErrorWithSuggestions(
					"Schema Registry endpoint not found",
					"Log in to Confluent Cloud with `confluent login`.\nSupply a Schema Registry endpoint with `--schema-registry-endpoint`.",
				)
			}
		}

		if !isCloudEndpointAutoFound {
			if schemaRegistryApiKey != "" && schemaRegistryApiSecret != "" {
				apiKey := srsdk.BasicAuth{
					UserName: schemaRegistryApiKey,
					Password: schemaRegistryApiSecret,
				}
				c.schemaRegistryClient = schemaregistry.NewClientWithApiKey(configuration, apiKey)
			} else {
				c.schemaRegistryClient = schemaregistry.NewClient(configuration, c.Config)
			}
		}

		if err := c.schemaRegistryClient.Get(); err != nil {
			return nil, fmt.Errorf("failed to validate Schema Registry client: %w", err)
		}
	}

	return c.schemaRegistryClient, nil
}

func (c *AuthenticatedCLICommand) GetUsageLimitsClient() *kafkausagelimits.UsageLimitsClient {
	if c.usageLimitsClient == nil {
		httpClient := ccloudv2.NewRetryableHttpClient(c.Config, false)
		c.usageLimitsClient = kafkausagelimits.NewUsageLimitsClient(c.Config, httpClient)
	}

	return c.usageLimitsClient
}

func (c *AuthenticatedCLICommand) validateSchemaRegistryEndpointAndSetClient(cluster srcmv3.SrcmV3Cluster, schemaRegistryApiKey, schemaRegistryApiSecret, url string, configuration srsdk.Configuration) bool {
	configuration.Servers = srsdk.ServerConfigurations{{URL: url}}
	configuration.DefaultHeader = map[string]string{"target-sr-cluster": cluster.GetId()}

	if schemaRegistryApiKey != "" && schemaRegistryApiSecret != "" {
		apiKey := srsdk.BasicAuth{
			UserName: schemaRegistryApiKey,
			Password: schemaRegistryApiSecret,
		}
		c.schemaRegistryClient = schemaregistry.NewClientWithApiKey(&configuration, apiKey)
	} else {
		c.schemaRegistryClient = schemaregistry.NewClient(&configuration, c.Config)
	}
	_, err := c.schemaRegistryClient.GetTopLevelConfig()
	if err == nil {
		return true
	}
	log.CliLogger.Infof("Trying endpoint: %s : Failed\n", url)

	return false
}

func (c *AuthenticatedCLICommand) GetCurrentSchemaRegistryClusterIdAndEndpoint(cmd *cobra.Command) (string, string, error) {
	clusters, err := c.V2Client.GetSchemaRegistryClustersByEnvironment(c.Context.GetCurrentEnvironment())
	if err != nil {
		return "", "", err
	}
	if len(clusters) == 0 {
		return "", "", schemaregistry.ErrNotEnabled
	}
	cluster := clusters[0]
	var endpoint string
	if c.Context.GetSchemaRegistryEndpoint() != "" {
		endpoint = c.Context.GetSchemaRegistryEndpoint()
	} else {
		endpoint, err = c.GetValidSchemaRegistryClusterEndpoint(cmd, cluster)
		if err != nil {
			return "", "", err
		}
	}
	return cluster.GetId(), endpoint, nil
}

func (c *AuthenticatedCLICommand) GetMDSClient(cmd *cobra.Command) (*mdsv1.APIClient, error) {
	if c.MDSClient == nil {
		unsafeTrace, err := cmd.Flags().GetBool("unsafe-trace")
		if err != nil {
			return nil, err
		}

		clientCertPath, clientKeyPath, err := GetClientCertAndKeyPaths(cmd)
		if err != nil {
			return nil, err
		}

		configuration := mdsv1.NewConfiguration()
		configuration.HTTPClient = utils.DefaultClient()
		configuration.Debug = unsafeTrace
		if c.Context == nil {
			c.MDSClient = mdsv1.NewAPIClient(configuration)
		}
		configuration.BasePath = c.Context.GetPlatformServer()
		configuration.UserAgent = c.Config.Version.UserAgent
		if c.Context.Platform.CaCertPath == "" {
			c.MDSClient = mdsv1.NewAPIClient(configuration)
		}

		caCertPath := c.Context.Platform.CaCertPath
		// Try to load certs. On failure, warn, but don't error out because this may be an auth command, so there may
		// be a --certificate-authority-path flag on the cmd line that'll fix whatever issue there is with the cert file in the config
		client, err := utils.CustomCAAndClientCertClient(caCertPath, clientCertPath, clientKeyPath)
		if err != nil {
			log.CliLogger.Warnf("Unable to load certificate from %s. %s. Resulting SSL errors will be fixed by logging in with the --certificate-authority-path flag.", caCertPath, err.Error())
		} else {
			configuration.HTTPClient = client
		}
		c.MDSClient = mdsv1.NewAPIClient(configuration)
	}

	return c.MDSClient, nil
}

func GetClientCertAndKeyPaths(cmd *cobra.Command) (string, string, error) {
	// Order of precedence: flags > env vars
	clientCertPath, err := cmd.Flags().GetString("client-cert-path")
	if err != nil {
		return "", "", err
	}
	clientKeyPath, err := cmd.Flags().GetString("client-key-path")
	if err != nil {
		return "", "", err
	}

	if clientCertPath == "" && clientKeyPath == "" {
		clientCertPath = os.Getenv(auth.ConfluentPlatformClientCertPath)
		clientKeyPath = os.Getenv(auth.ConfluentPlatformClientKeyPath)
	}

	return clientCertPath, clientKeyPath, nil
}

func (c *AuthenticatedCLICommand) GetValidSchemaRegistryClusterEndpoint(cmd *cobra.Command, cluster srcmv3.SrcmV3Cluster) (string, error) {
	unsafeTrace, _ := cmd.Flags().GetBool("unsafe-trace")
	configuration := srsdk.NewConfiguration()
	configuration.UserAgent = c.Config.Version.UserAgent
	configuration.Debug = unsafeTrace
	configuration.HTTPClient = ccloudv2.NewRetryableHttpClient(c.Config, unsafeTrace)

	schemaRegistryApiKey, _ := cmd.Flags().GetString("schema-registry-api-key")
	schemaRegistryApiSecret, _ := cmd.Flags().GetString("schema-registry-api-secret")
	isCloudEndpointAutoFound := false
	var endpoint string
	for _, urlPrivate := range cluster.Spec.PrivateNetworkingConfig.GetRegionalEndpoints() {
		isCloudEndpointAutoFound = c.validateSchemaRegistryEndpointAndSetClient(cluster, schemaRegistryApiKey, schemaRegistryApiSecret, urlPrivate, *configuration)
		if isCloudEndpointAutoFound {
			endpoint = urlPrivate
			err := c.saveToConfig(urlPrivate)
			if err != nil {
				return "", err
			}
			break
		}
	}
	if !isCloudEndpointAutoFound {
		isCloudEndpointAutoFound = c.validateSchemaRegistryEndpointAndSetClient(cluster, schemaRegistryApiKey, schemaRegistryApiSecret, cluster.Spec.GetHttpEndpoint(), *configuration)
		if isCloudEndpointAutoFound {
			endpoint = cluster.Spec.GetHttpEndpoint()
			err := c.saveToConfig(cluster.Spec.GetHttpEndpoint())
			if err != nil {
				return "", err
			}
		}
	}
	if !isCloudEndpointAutoFound {
		return "", errors.NewErrorWithSuggestions(
			"Schema Registry endpoint not found",
			"Log in to Confluent Cloud with `confluent login`.\nSupply a Schema Registry endpoint with `--schema-registry-endpoint`.",
		)
	}
	return endpoint, nil
}

func (c *AuthenticatedCLICommand) saveToConfig(url string) error {
	err := c.Context.SetSchemaRegistryEndpoint(url)
	if err != nil {
		return err
	}
	if err := c.Config.Save(); err != nil {
		return err
	}
	return nil
}
