package auth

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/analytics"
	pauth "github.com/confluentinc/cli/internal/pkg/auth"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/netrc"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

type loginCommand struct {
	*pcmd.CLICommand
	cliName         string
	Logger          *log.Logger
	analyticsClient analytics.Client
	// for testing
	MDSClientManager    pauth.MDSClientManager
	ccloudClientFactory pauth.CCloudClientFactory
	netrcHandler        netrc.NetrcHandler
	loginTokenHandler   pauth.LoginTokenHandler
}

func NewLoginCommand(cliName string, prerunner pcmd.PreRunner, log *log.Logger, ccloudClientFactory pauth.CCloudClientFactory,
	mdsClientManager pauth.MDSClientManager, analyticsClient analytics.Client, netrcHandler netrc.NetrcHandler,
	loginTokenHandler pauth.LoginTokenHandler) *loginCommand {
	cmd := &loginCommand{
		cliName:             cliName,
		Logger:              log,
		analyticsClient:     analyticsClient,
		MDSClientManager:    mdsClientManager,
		ccloudClientFactory: ccloudClientFactory,
		netrcHandler:        netrcHandler,
		loginTokenHandler:   loginTokenHandler,
	}
	cmd.init(prerunner)
	return cmd
}

func (a *loginCommand) init(prerunner pcmd.PreRunner) {
	remoteAPIName := getRemoteAPIName(a.cliName)
	loginCmd := &cobra.Command{
		Use:   "login",
		Short: fmt.Sprintf("Log in to %s.", remoteAPIName),
		Args:  cobra.NoArgs,
		PersistentPreRunE: pcmd.NewCLIPreRunnerE(func(cmd *cobra.Command, args []string) error {
			a.analyticsClient.SetCommandType(analytics.Login)
			return a.CLICommand.PersistentPreRunE(cmd, args)
		}),
	}
	if a.cliName == "ccloud" {
		loginCmd.RunE = pcmd.NewCLIRunE(a.login)
		loginCmd.Flags().String("url", pauth.CCloudURL, "Confluent Cloud service URL.")
	} else {
		loginCmd.RunE = pcmd.NewCLIRunE(a.loginMDS)
		loginCmd.Flags().String("url", "", "Metadata service URL.")
		loginCmd.Flags().String("ca-cert-path", "", "Self-signed certificate chain in PEM format.")
		loginCmd.Short = strings.ReplaceAll(loginCmd.Short, ".", " (required for RBAC).")
		loginCmd.Long = strings.ReplaceAll(loginCmd.Long, ".", " (required for RBAC).")
		check(loginCmd.MarkFlagRequired("url")) // because https://confluent.cloud isn't an MDS endpoint
	}
	loginCmd.Flags().Bool("no-browser", false, "Do not open browser when authenticating via Single Sign-On.")
	loginCmd.Flags().Bool("save", false, "Save login credentials or refresh token (in the case of SSO) to local netrc file.")
	loginCmd.Flags().SortFlags = false
	cliLoginCmd := pcmd.NewAnonymousCLICommand(loginCmd, prerunner)
	a.CLICommand = cliLoginCmd
}

func getRemoteAPIName(cliName string) string {
	if cliName == "ccloud" {
		return "Confluent Cloud"
	}
	return "Confluent Platform"
}

func (a *loginCommand) login(cmd *cobra.Command, _ []string) error {
	url, err := a.getURL(cmd)
	if err != nil {
		return err
	}

	token, creds, err := a.getCCloudTokenAndCredentials(cmd, url)
	if err != nil {
		return err
	}

	currentEnv, err := pauth.PersistCCloudLoginToConfig(a.Config.Config, creds.Username, url, token,
		a.ccloudClientFactory.JwtHTTPClientFactory(context.Background(), token, url))
	if err != nil {
		return err
	}

	err = a.saveLoginToNetrc(cmd, creds)
	if err != nil {
		return err
	}

	utils.Printf(cmd, errors.LoggedInAsMsg, creds.Username)
	utils.Printf(cmd, errors.LoggedInUsingEnvMsg, currentEnv.Id, currentEnv.Name)
	return err
}

func (a *loginCommand) getCCloudTokenAndCredentials(cmd *cobra.Command, url string) (string, *pauth.Credentials, error) {
	client := a.ccloudClientFactory.AnonHTTPClientFactory(url)

	token, creds, err := a.loginTokenHandler.GetCCloudTokenAndCredentialsFromEnvVar(cmd, client)
	if err == nil && len(token) > 0 {
		return token, creds, nil
	}

	token, creds, err = a.loginTokenHandler.GetCCloudTokenAndCredentialsFromNetrc(cmd, client, url, netrc.GetMatchingNetrcMachineParams{
		CLIName: a.cliName,
		URL:     url,
	})
	if err == nil && len(token) > 0 {
		return token, creds, nil
	}

	return a.loginTokenHandler.GetCCloudTokenAndCredentialsFromPrompt(cmd, client, url)
}

func (a *loginCommand) loginMDS(cmd *cobra.Command, _ []string) error {
	url, err := a.getURL(cmd)
	if err != nil {
		return err
	}

	caCertPath, err := a.getCaCertPath(cmd)
	if err != nil {
		return err
	}

	token, creds, err := a.getConfluentTokenAndCredentials(cmd, url, caCertPath)
	if err != nil {
		return err
	}

	err = pauth.PersistConfluentLoginToConfig(a.Config.Config, creds.Username, url, token, caCertPath)
	if err != nil {
		return err
	}

	err = a.saveLoginToNetrc(cmd, creds)
	if err != nil {
		return err
	}

	utils.Printf(cmd, errors.LoggedInAsMsg, creds.Username)
	return nil
}

// if ca-cert-path flag is not used, returns caCertPath value from config
// if user passes empty string for ca-cert-path flag then user intends to reset the ca-cert-path
func (a *loginCommand) getCaCertPath(cmd *cobra.Command) (string, error) {
	caCertPath, err := cmd.Flags().GetString("ca-cert-path")
	if err != nil {
		return "", err
	}
	if caCertPath == "" {
		changed := cmd.Flags().Changed("ca-cert-path")
		if changed {
			return "", nil
		}
		return a.getCaCertPathFromConfig(cmd)
	}
	return caCertPath, nil
}

func (a *loginCommand) getCaCertPathFromConfig(cmd *cobra.Command) (string, error) {
	ctx, err := a.getContext(cmd)
	if err != nil {
		return "", err
	}
	if ctx != nil {
		return ctx.Platform.CaCertPath, nil
	}
	return "", nil
}

func (a *loginCommand) getContext(cmd *cobra.Command) (*v3.Context, error) {
	dynamicContext, err := a.Config.Context(cmd)
	if err != nil {
		return nil, err
	}
	var ctx *v3.Context
	if dynamicContext != nil {
		ctx = dynamicContext.Context
	}
	return ctx, nil
}

// Order of precedence: env vars > netrc > prompt
// i.e. if login credentials found in env vars then acquire token using env vars and skip checking for credentials else where
func (a *loginCommand) getConfluentTokenAndCredentials(cmd *cobra.Command, url string, caCertPath string) (string, *pauth.Credentials, error) {
	client, err := a.MDSClientManager.GetMDSClient(url, caCertPath, a.Logger)
	if err != nil {
		return "", nil, err
	}

	token, creds, err := a.loginTokenHandler.GetConfluentTokenAndCredentialsFromEnvVar(cmd, client)
	if err == nil && len(token) > 0 {
		return token, creds, nil
	}

	token, creds, err = a.loginTokenHandler.GetConfluentTokenAndCredentialsFromNetrc(cmd, client, netrc.GetMatchingNetrcMachineParams{
		CLIName: a.cliName,
		URL:     url,
	})
	if err == nil && len(token) > 0 {
		return token, creds, nil
	}

	return a.loginTokenHandler.GetConfluentTokenAndCredentialsFromPrompt(cmd, client)
}

func (a *loginCommand) getURL(cmd *cobra.Command) (string, error) {
	url, err := cmd.Flags().GetString("url")
	if err != nil {
		return "", err
	}
	url, valid, errMsg := validateURL(url, a.cliName)
	if !valid {
		return "", errors.Errorf(errors.InvalidLoginURLMsg)
	}
	if errMsg != "" {
		utils.ErrPrintf(cmd, errors.UsingLoginURLDefaults, errMsg)
	}
	return url, nil
}

func (a *loginCommand) saveLoginToNetrc(cmd *cobra.Command, creds *pauth.Credentials) error {
	saveToNetrc, err := cmd.Flags().GetBool("save")
	if err != nil {
		return err
	}
	if saveToNetrc {
		var err error
		if creds.RefreshToken == "" {
			err = a.netrcHandler.WriteNetrcCredentials(a.Config.CLIName, false, a.Config.Config.Context().Name, creds.Username, creds.Password)
		} else {
			err = a.netrcHandler.WriteNetrcCredentials(a.Config.CLIName, true, a.Config.Config.Context().Name, creds.Username, creds.RefreshToken)
		}
		if err != nil {
			return err
		}
		utils.ErrPrintf(cmd, errors.WroteCredentialsToNetrcMsg, a.netrcHandler.GetFileName())
	}
	return nil
}

func validateURL(url string, cli string) (string, bool, string) {
	protocol_rgx, _ := regexp.Compile(`(\w+)://`)
	port_rgx, _ := regexp.Compile(`:(\d+\/?)`)

	protocol_match := protocol_rgx.MatchString(url)
	port_match := port_rgx.MatchString(url)

	var msg []string
	if !protocol_match {
		if cli == "ccloud" {
			url = "https://" + url
			msg = append(msg, "https protocol")
		} else {
			url = "http://" + url
			msg = append(msg, "http protocol")
		}
	}
	if !port_match && cli == "confluent" {
		url = url + ":8090"
		msg = append(msg, "default MDS port 8090")
	}
	var pattern string
	if cli == "confluent" {
		pattern = `^\w+://[^/ ]+:\d+(?:\/|$)`
	} else {
		pattern = `^\w+://[^/ ]+`
	}
	matched, _ := regexp.Match(pattern, []byte(url))

	return url, matched, strings.Join(msg, " and ")
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
