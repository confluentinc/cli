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
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/errors"
)

type command struct {
	*pcmd.CLICommand
	prompt   pcmd.Prompt
	resolver pcmd.FlagResolver
}

// TODO: Make long description better.
const longDescription = "Initialize and set a current context."

func New(prerunner pcmd.PreRunner, config *v3.Config, prompt pcmd.Prompt, resolver pcmd.FlagResolver, analyticsClient analytics.Client) *cobra.Command {
	cobraCmd := &cobra.Command{
		Use:   "init <context-name>",
		Short: "Initialize a context.",
		Long:  longDescription,
		Args:  cobra.ExactArgs(1),
	}
	cliCmd := pcmd.NewAnonymousCLICommand(cobraCmd, config, prerunner)
	cobraCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		analyticsClient.SetCommandType(analytics.Init)
		return prerunner.Anonymous(cliCmd)(cmd, args)
	}
	cmd := &command{
		CLICommand: cliCmd,
		prompt:     prompt,
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
	c.RunE = c.initContext
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
		return "", fmt.Errorf("%s cannot be empty", displayName)
	}
	return val, nil
}

func (c *command) initContext(cmd *cobra.Command, args []string) error {
	contextName := args[0]
	kafkaAuth, err := c.Flags().GetBool("kafka-auth")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	eh := new(errors.Handler)
	if !kafkaAuth {
		return errors.HandleCommon(errors.New("only kafka-auth is currently supported"), cmd)
	}
	bootstrapURL := eh.HandleString(c.parseStringFlag("bootstrap", "Bootstrap URL: ", false,
		"Bootstrap URL"))
	apiKey := eh.HandleString(c.parseStringFlag("api-key", "API Key: ", false,
		"API key"))
	apiSecret := eh.HandleString(c.parseStringFlag("api-secret", "API Secret: ", true,
		"API secret"))
	if err := eh.Reset(); err != nil {
		return errors.HandleCommon(err, cmd)
	}
	err = c.addContext(contextName, bootstrapURL, apiKey, apiSecret)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	// Set current context.
	err = c.Config.SetContext(contextName)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
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
		return fmt.Errorf("credential type %d unknown", credential.CredentialType)
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
