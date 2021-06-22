package init

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/analytics"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v0 "github.com/confluentinc/cli/internal/pkg/config/v0"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	v2 "github.com/confluentinc/cli/internal/pkg/config/v2"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

type command struct {
	*pcmd.CLICommand
	resolver pcmd.FlagResolver
}

// TODO: Make long description better.
const longDescription = "Initialize and set a current context."

func New(prerunner pcmd.PreRunner, resolver pcmd.FlagResolver, analyticsClient analytics.Client) *cobra.Command {
	cobraCmd := &cobra.Command{
		Use:   "init <context-name>",
		Short: "Initialize a context.",
		Long:  longDescription,
		Args:  cobra.ExactArgs(1),
	}
	cliCmd := pcmd.NewAnonymousCLICommand(cobraCmd, prerunner)
	cobraCmd.PersistentPreRunE = pcmd.NewCLIPreRunnerE(func(cmd *cobra.Command, args []string) error {
		analyticsClient.SetCommandType(analytics.Init)
		return prerunner.Anonymous(cliCmd)(cmd, args)
	})
	cmd := &command{
		CLICommand: cliCmd,
		resolver:   resolver,
	}
	cmd.init()
	return cmd.Command
}

func (c *command) init() {
	c.Flags().Bool("kafka-auth", false, "Initialize with bootstrap url, API key, and API secret. "+
		"Can be done interactively, with flags, or both.")
	c.Flags().String("bootstrap", "", "Bootstrap URL.")
	c.Flags().String("api-key", "", "API key.")
	c.Flags().String("api-secret", "", "API secret. Can be specified as plaintext, "+
		"as a file, starting with '@', or as stdin, starting with '-'.")
	c.Flags().SortFlags = false
	c.RunE = pcmd.NewCLIRunE(c.initContext)
}

func (c *command) parseStringFlag(name string, prompt string, secure bool, displayName string) (string, error) {
	str, err := c.Flags().GetString(name)
	if err != nil {
		return "", err
	}
	val, err := c.resolver.ValueFrom(str, prompt, secure)
	if err != nil {
		return "", err
	}
	val = strings.TrimSpace(val)
	if len(val) == 0 {
		return "", errors.Errorf(errors.CannotBeEmptyErrorMsg, displayName)
	}
	return val, nil
}

func (c *command) initContext(cmd *cobra.Command, args []string) error {
	contextName := args[0]
	kafkaAuth, err := c.Flags().GetBool("kafka-auth")
	if err != nil {
		return err
	}
	if !kafkaAuth {
		return errors.New(errors.OnlyKafkaAuthErrorMsg)
	}
	bootstrapURL, err := c.parseStringFlag("bootstrap", "Bootstrap URL: ", false,
		"Bootstrap URL")
	if err != nil {
		return err
	}
	apiKey, err := c.parseStringFlag("api-key", "API Key: ", false,
		"API key")
	if err != nil {
		return err
	}
	apiSecret, err := c.parseStringFlag("api-secret", "API Secret: ", true,
		"API secret")
	if err != nil {
		return err
	}
	err = c.addContext(contextName, bootstrapURL, apiKey, apiSecret)
	if err != nil {
		return err
	}
	// Set current context.
	err = c.Config.SetContext(contextName)
	if err != nil {
		return err
	}
	utils.Printf(cmd, errors.InitContextMsg, contextName)
	return nil
}

func (c *command) addContext(name string, bootstrapURL string, apiKey string, apiSecret string) error {
	apiKeyPair := &v0.APIKeyPair{
		Key:    apiKey,
		Secret: apiSecret,
	}
	apiKeys := map[string]*v0.APIKeyPair{
		apiKey: apiKeyPair,
	}
	kafkaClusterCfg := &v1.KafkaClusterConfig{
		ID:          "anonymous-id",
		Name:        "anonymous-cluster",
		Bootstrap:   bootstrapURL,
		APIEndpoint: "",
		APIKeys:     apiKeys,
		APIKey:      apiKey,
	}
	kafkaClusters := map[string]*v1.KafkaClusterConfig{
		kafkaClusterCfg.ID: kafkaClusterCfg,
	}
	platform := &v2.Platform{Server: bootstrapURL}
	// Inject credential and platforms name for now, until users can provide custom names.
	platform.Name = strings.TrimPrefix(platform.Server, "https://")
	// Hardcoded for now, since username/password isn't implemented yet.
	credential := &v2.Credential{
		Username:       "",
		Password:       "",
		APIKeyPair:     apiKeyPair,
		CredentialType: v2.APIKey,
	}
	switch credential.CredentialType {
	case v2.Username:
		credential.Name = fmt.Sprintf("%s-%s", &credential.CredentialType, credential.Username)
	case v2.APIKey:
		credential.Name = fmt.Sprintf("%s-%s", &credential.CredentialType, credential.APIKeyPair.Key)
	default:
		return errors.Errorf(errors.UnknownCredentialTypeErrorMsg, credential.CredentialType)
	}
	err := c.Config.SaveCredential(credential)
	if err != nil {
		return err
	}
	err = c.Config.SavePlatform(platform)
	if err != nil {
		return err
	}
	return c.Config.AddContext(name, platform.Name, credential.Name, kafkaClusters,
		kafkaClusterCfg.ID, nil, nil)
}
