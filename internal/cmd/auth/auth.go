package auth

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/confluentinc/ccloud-sdk-go"
	orgv1 "github.com/confluentinc/ccloudapis/org/v1"

	auth_server "github.com/confluentinc/cli/internal/pkg/auth-server"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"

	"github.com/confluentinc/mds-sdk-go"
)

type commands struct {
	Commands   []*cobra.Command
	config     *config.Config
	mdsClient  *mds.APIClient
	Logger     *log.Logger
	// @VisibleForTesting, defaults to the OS filesystem
	certReader io.Reader
	// for testing
	prompt                pcmd.Prompt
	anonHTTPClientFactory func(baseURL string, logger *log.Logger) *ccloud.Client
	jwtHTTPClientFactory  func(ctx context.Context, authToken string, baseURL string, logger *log.Logger) *ccloud.Client
}

// New returns a list of auth-related Cobra commands.
func New(prerunner pcmd.PreRunner, config *config.Config, logger *log.Logger, mdsClient *mds.APIClient, userAgent string) []*cobra.Command {
	var defaultAnonHTTPClientFactory = func(baseURL string, logger *log.Logger) *ccloud.Client {
		return ccloud.NewClient(&ccloud.Params{BaseURL: baseURL, HttpClient: ccloud.BaseClient, Logger: logger, UserAgent: userAgent})
	}
	var defaultJwtHTTPClientFactory = func(ctx context.Context, jwt string, baseURL string, logger *log.Logger) *ccloud.Client {
		return ccloud.NewClientWithJWT(ctx, jwt, &ccloud.Params{BaseURL: baseURL, Logger: logger, UserAgent: userAgent})
	}
	return newCommands(prerunner, config, logger, mdsClient, pcmd.NewPrompt(os.Stdin),
		defaultAnonHTTPClientFactory, defaultJwtHTTPClientFactory,
	).Commands
}

func newCommands(prerunner pcmd.PreRunner, config *config.Config, log *log.Logger, mdsClient *mds.APIClient, prompt pcmd.Prompt,
	anonHTTPClientFactory func(baseURL string, logger *log.Logger) *ccloud.Client,
	jwtHTTPClientFactory func(ctx context.Context, authToken string, baseURL string, logger *log.Logger) *ccloud.Client,
) *commands {
	cmd := &commands{
		config:                config,
		mdsClient:             mdsClient,
		Logger:                log,
		prompt:                prompt,
		anonHTTPClientFactory: anonHTTPClientFactory,
		jwtHTTPClientFactory:  jwtHTTPClientFactory,
	}
	cmd.init(prerunner)
	return cmd
}

func (a *commands) init(prerunner pcmd.PreRunner) {
	loginCmd := &cobra.Command{
		Use:   "login",
		Short: fmt.Sprintf("Log in to %s.", a.config.APIName()),
		Long:  fmt.Sprintf("Log in to %s.", a.config.APIName()),
		Args:  cobra.NoArgs,
	}
	if a.config.CLIName == "ccloud" {
		loginCmd.RunE = a.login
		loginCmd.Flags().String("url", "https://confluent.cloud", "Confluent Cloud service URL.")
	} else {
		loginCmd.RunE = a.loginMDS
		loginCmd.Flags().String("url", "", "Metadata service URL.")
		loginCmd.Flags().String("ca-cert-path", "", "Self-signed certificate in PEM format.")
		loginCmd.Short = strings.Replace(loginCmd.Short, ".", " (required for RBAC).", -1)
		loginCmd.Long = strings.Replace(loginCmd.Long, ".", " (required for RBAC).", -1)
		check(loginCmd.MarkFlagRequired("url")) // because https://confluent.cloud isn't an MDS endpoint
	}
	loginCmd.Flags().SortFlags = false
	loginCmd.PersistentPreRunE = prerunner.Anonymous()
	logoutCmd := &cobra.Command{
		Use:   "logout",
		Short: fmt.Sprintf("Logout of %s.", a.config.APIName()),
		Long:  fmt.Sprintf("Logout of %s.", a.config.APIName()),

		RunE: a.logout,
		Args: cobra.NoArgs,
	}
	logoutCmd.PersistentPreRunE = prerunner.Anonymous()
	a.Commands = []*cobra.Command{loginCmd, logoutCmd}
}

func (a *commands) login(cmd *cobra.Command, args []string) error {
	url, err := cmd.Flags().GetString("url")
	if err != nil {
		return err
	}
	a.config.AuthURL = url

	client := a.anonHTTPClientFactory(a.config.AuthURL, a.config.Logger)
	email, password, err := a.credentials(cmd, "Email", client)
	if err != nil {
		return err
	}

	// Check if user has an enterprise SSO connection enabled.  If so we need to start
	// a background HTTP server to support the authorization code flow with PKCE
	// described at https://auth0.com/docs/flows/guides/auth-code-pkce/call-api-auth-code-pkce
	userSSO, err := client.User.CheckEmail(context.Background(), &orgv1.User{Email: email})
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	token := ""

	if userSSO != nil && userSSO.Sso != nil && userSSO.Sso.Enabled && userSSO.Sso.Auth0ConnectionName != "" {
		// Be conservative: only bother trying to launch server if we have to
		server := &auth_server.AuthServer{}
		err = server.Start(a.config.AuthURL)
		if err != nil {
			return errors.HandleCommon(err, cmd)
		}

		// Get authorization code for making subsequent token request
		err = server.GetAuthorizationCode(userSSO.Sso.Auth0ConnectionName)
		if err != nil {
			return errors.HandleCommon(err, cmd)
		}

		// Exchange authorization code for OAuth token from SSO provider
		err := server.GetOAuthToken()
		if err != nil {
			return errors.HandleCommon(err, cmd)
		}

		token, err = client.Auth.Login(context.Background(), server.SSOProviderIDToken, "", "")
		if err != nil {
			return errors.HandleCommon(err, cmd)
		}
	} else {
		token, err = client.Auth.Login(context.Background(), "", email, password)
		if err != nil {
			return errors.HandleCommon(err, cmd)
		}
	}

	a.config.AuthToken = token

	client = a.jwtHTTPClientFactory(context.Background(), a.config.AuthToken, a.config.AuthURL, a.config.Logger)
	user, err := client.Auth.User(context.Background())
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	if len(user.Accounts) == 0 {
		return errors.HandleCommon(errors.New("No environments found for authenticated user!"), cmd)
	}

	// If no auth config exists, initialize it
	if a.config.Auth == nil {
		a.config.Auth = &config.AuthConfig{}
	}

	// Always overwrite the user and list of accounts when logging in -- but don't necessarily
	// overwrite `Account` (current/active environment) since we want that to be remembered
	// between CLI sessions.
	a.config.Auth.User = user.User
	a.config.Auth.Accounts = user.Accounts
	// Default to 0th environment if no suitable environment is already configured
	hasGoodEnv := false
	if a.config.Auth.Account != nil {
		for _, acc := range a.config.Auth.Accounts {
			if acc.Id == a.config.Auth.Account.Id {
				hasGoodEnv = true
			}
		}
	}
	if !hasGoodEnv {
		a.config.Auth.Account = a.config.Auth.Accounts[0]
	}

	err = a.setContextAndAddContextIfAbsent(a.config.Auth.User.Email, "")
	if err != nil {
		return err
	}
	err = a.config.Save()
	if err != nil {
		return errors.Wrap(err, "Unable to save user authentication.")
	}
	pcmd.Println(cmd, "Logged in as", email)
	pcmd.Print(cmd, "Using environment ", a.config.Auth.Account.Id,
		" (\"", a.config.Auth.Account.Name, "\")\n")
	return err
}

func (a *commands) loginMDS(cmd *cobra.Command, args []string) error {
	url, err := cmd.Flags().GetString("url")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	a.config.AuthURL = url
	caCertPath := ""
	if cmd.Flags().Changed("ca-cert-path") {
		caCertPath, err = cmd.Flags().GetString("ca-cert-path")
		if err != nil {
			return errors.HandleCommon(err, cmd)
		}
		if caCertPath == "" {
			// revert to default client regardless of previously configured client
			a.mdsClient.GetConfig().HTTPClient = DefaultClient()
		} else {
			// override previously configured httpclient if a new cert path was specified
			if a.certReader == nil {
				// if a certReader wasn't already set, eg. for testing, then create one now
				caCertPath, err = filepath.Abs(caCertPath)
				if err != nil {
					return errors.HandleCommon(err, cmd)
				}
				caCertFile, err := os.Open(caCertPath)
				if err != nil {
					return errors.HandleCommon(err, cmd)
				}
				defer caCertFile.Close()
				a.certReader = caCertFile
			}
			a.mdsClient.GetConfig().HTTPClient, err = SelfSignedCertClient(a.certReader, a.Logger)
			if err != nil {
				return errors.HandleCommon(err, cmd)
			}
			a.Logger.Debugf("Successfully loaded certificate from %s", caCertPath)
		}
	}
	a.mdsClient.ChangeBasePath(a.config.AuthURL)
	email, password, err := a.credentials(cmd, "Username", nil)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	basicContext := context.WithValue(context.Background(), mds.ContextBasicAuth, mds.BasicAuth{UserName: email, Password: password})
	resp, _, err := a.mdsClient.TokensAuthenticationApi.GetToken(basicContext, "")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	a.config.AuthToken = resp.AuthToken

	err = a.config.Save()
	if err != nil {
		return errors.Wrap(err, "Unable to save user authentication.")
	}
	err = a.setContextAndAddContextIfAbsent(email, caCertPath)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	pcmd.Println(cmd, "Logged in as", email)

	return errors.HandleCommon(err, cmd)
}

func (a *commands) logout(cmd *cobra.Command, args []string) error {
	a.config.AuthToken = ""
	a.config.Auth = nil
	err := a.config.Save()
	if err != nil {
		return errors.Wrap(err, "Unable to delete user auth")
	}
	pcmd.Println(cmd, "You are now logged out")
	return nil
}

func (a *commands) credentials(cmd *cobra.Command, userField string, cloudClient *ccloud.Client) (string, string, error) {
	email := os.Getenv("XX_CCLOUD_EMAIL")
	if len(email) == 0 {
		email = os.Getenv("XX_CONFLUENT_USERNAME")
	}
	password := os.Getenv("XX_CCLOUD_PASSWORD")
	if len(password) == 0 {
		password = os.Getenv("XX_CONFLUENT_PASSWORD")
	}
	if len(email) == 0 || len(password) == 0 {
		pcmd.Println(cmd, "Enter your Confluent credentials:")
	}
	if len(email) == 0 {
		pcmd.Print(cmd, userField+": ")
		emailFromPrompt, err := a.prompt.ReadString('\n')
		if err != nil {
			return "", "", err
		}
		email = strings.TrimSpace(emailFromPrompt)
	}

	a.Logger.Trace("Successfully obtained email")

	// In the case of MDS login (`confluent`) or in the case of some of the mocks,
	// cloudClient will be nll, so we need this check
	if cloudClient != nil {
		// If SSO user, don't prompt for password
		userSSO, err := cloudClient.User.CheckEmail(context.Background(), &orgv1.User{Email: email})
		// Fine to ignore non-nil err for this request: e.g. what if this fails due to invalid/malicious
		// email, we want to silently continue and give the illusion of password prompt.
		// If Auth0ConnectionName is blank ("local" user) still prompt for password
		if err == nil && userSSO != nil && userSSO.Sso != nil && userSSO.Sso.Enabled && userSSO.Sso.Auth0ConnectionName != "" {
			a.Logger.Trace("User is SSO-enabled so won't prompt for password")
			return email, password, nil
		}
	}

	if len(password) == 0 {
		var err error
		pcmd.Print(cmd, "Password: ")
		bytePassword, err := a.prompt.ReadPassword()
		if err != nil {
			return "", "", err
		}
		pcmd.Println(cmd)
		password = string(bytePassword)
	}

	a.Logger.Trace("Successfully obtained password")

	return email, password, nil
}

func (a *commands) setContextAndAddContextIfAbsent(username string, caCertPath string) error {
	name := fmt.Sprintf("login-%s-%s", username, a.config.AuthURL)
	if _, ok := a.config.Contexts[name]; ok {
		err := a.config.SetContext(name)
		if err != nil {
			return err
		}
		return nil
	}
	platform := &config.Platform{
		Server:     a.config.AuthURL,
		CaCertPath: caCertPath,
	}
	credential := &config.Credential{
		Username: username,
		// don't save password if they entered it interactively.
	}
	err := a.config.AddContext(name, platform, credential, map[string]*config.KafkaClusterConfig{}, "", nil)
	if err != nil {
		return err
	}
	err = a.config.SetContext(name)
	if err != nil {
		return err
	}
	return nil
}

func SelfSignedCertClient(certReader io.Reader, logger *log.Logger) (*http.Client, error){
	certPool, err := x509.SystemCertPool()
	if err != nil {
		logger.Warnf("Unable to load system certificates. Continuing with custom certificates only.")
	}
	if certPool == nil {
		certPool = x509.NewCertPool()
	}

	if certReader == nil {
		return nil, fmt.Errorf("no reader specified for reading custom certificates")
	}
	certs, err := ioutil.ReadAll(certReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate: %v", err)
	}

	// Append new cert to the system pool
	if ok := certPool.AppendCertsFromPEM(certs); !ok {
		return nil, fmt.Errorf("no certs appended, using system certs only")
	}

	// Trust the updated cert pool in our client
	tlsClientConfig := &tls.Config{RootCAs: certPool}
	transport := &http.Transport{TLSClientConfig: tlsClientConfig}
	client := DefaultClient()
	client.Transport = transport

	return client, nil
}

func DefaultClient() *http.Client {
	return http.DefaultClient
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
