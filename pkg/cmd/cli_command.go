package cmd

import (
	"net/http"
	"os"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v3/pkg/auth"
	"github.com/confluentinc/cli/v3/pkg/config"
	"github.com/confluentinc/cli/v3/pkg/utils"
	"github.com/confluentinc/cli/v3/pkg/version"
	testserver "github.com/confluentinc/cli/v3/test/test-server"
	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"
)

type CLICommand struct {
	*cobra.Command
	Config  *config.Config
	Version *version.Version

	cmfClient *cmfsdk.APIClient
}

func NewCLICommand(cmd *cobra.Command) *CLICommand {
	return &CLICommand{
		Command: cmd,
		Config:  &config.Config{},
	}
}

func NewAnonymousCLICommandWithoutContext(cmd *cobra.Command, prerunner PreRunner) *CLICommand {
	c := NewCLICommand(cmd)
	cmd.PersistentPreRunE = prerunner.Anonymous(c, false)
	return c
}

func NewAnonymousCLICommand(cmd *cobra.Command, prerunner PreRunner) *CLICommand {
	c := NewCLICommand(cmd)
	cmd.PersistentPreRunE = chain(prerunner.Anonymous(c, false), prerunner.ParseFlagsIntoContext(c))
	return c
}

func (c *CLICommand) GetCmfClient(cmd *cobra.Command) (*cmfsdk.APIClient, error) {
	if c.cmfClient == nil {
		cfg := cmfsdk.NewConfiguration()

		unsafeTrace, err := cmd.Flags().GetBool("unsafe-trace")
		if err != nil {
			return nil, err
		}
		cfg.Debug = unsafeTrace
		if c.Config.IsTest {
			cfg.BasePath = testserver.TestCmfUrl.String() + "/cmf/api/v1"
		} else {
			flags, err := resolveOnPremCMFRestFlags(cmd)
			if err != nil {
				return nil, err
			}
			cfg.BasePath = flags.url + "/cmf/api/v1"
		}

		restFlags, err := resolveOnPremCMFRestFlags(cmd)
		if err != nil {
			return nil, err
		}
		cfg.HTTPClient, err = createCmfRestClient(restFlags.caCertPath, restFlags.clientCertPath, restFlags.clientKeyPath)
		if err != nil {
			return nil, err
		}
		client := cmfsdk.NewAPIClient(cfg)
		c.cmfClient = client

	}
	return c.cmfClient, nil
}

type onPremCMFRestFlagValues struct {
	url            string
	caCertPath     string
	clientCertPath string
	clientKeyPath  string
}

func resolveOnPremCMFRestFlags(cmd *cobra.Command) (*onPremCMFRestFlagValues, error) {
	url, _ := cmd.Flags().GetString("url")
	certificateAuthorityPath, _ := cmd.Flags().GetString("certificate-authority-path")
	clientCertPath, _ := cmd.Flags().GetString("client-cert-path")
	clientKeyPath, _ := cmd.Flags().GetString("client-key-path")
	values := &onPremCMFRestFlagValues{
		url:            url,
		caCertPath:     certificateAuthorityPath,
		clientCertPath: clientCertPath,
		clientKeyPath:  clientKeyPath,
	}
	return values, nil
}

func createCmfRestClient(caCertPath, clientCertPath, clientKeyPath string) (*http.Client, error) {
	// If caCertPath is not provided via flag, check if it is set in the environment
	if caCertPath == "" {
		caCertPath = os.Getenv(auth.ConfluentPlatformCmfCertificateAuthorityPath)
	}
	// If we find a caCertPath, we will use it to create the client using the custom certificate authority
	if caCertPath != "" {
		client, err := utils.CustomCAAndClientCertClient(caCertPath, clientCertPath, clientKeyPath)
		if err != nil {
			return nil, err
		}
		return client, nil
		// use cert path from config if available
	} else if clientCertPath != "" && clientKeyPath != "" {
		client, err := utils.CustomCAAndClientCertClient("", clientCertPath, clientKeyPath)
		if err != nil {
			return nil, err
		}
		return client, nil
	}
	return utils.DefaultClient(), nil
}
