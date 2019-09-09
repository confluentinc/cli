package init

import (
	"strings"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/spf13/cobra"
)

type command struct {
	*cobra.Command
	config   *config.Config
	prompt   pcmd.Prompt
	resolver pcmd.FlagResolver
}

// TODO: Make long description better.
const longDescription = "Initialize and set a current context."

func New(prerunner pcmd.PreRunner, config *config.Config, prompt pcmd.Prompt, resolver pcmd.FlagResolver) *cobra.Command {
	cmd := &command{
		&cobra.Command{
			Use:               "init <context-name>",
			Short:             "Initialize a context.",
			Long:              longDescription,
			PersistentPreRunE: prerunner.Anonymous(),
			Args:              cobra.ExactArgs(1),
		},
		config,
		prompt,
		resolver,
	}
	cmd.init()
	return cmd.Command
}

func (c *command) init() {
	c.Flags().Bool("kafka-auth", false, "Interactively initialize with bootstrap url, API key, and API secret.")
	c.Flags().String("bootstrap", "", "Bootstrap URL.")
	c.Flags().String("api-key", "", "API key.")
	c.Flags().String("api-secret", "", "API secret file, starting with '@'.")
	c.Flags().SortFlags = false
	c.RunE = c.initContext
}

func (c *command) initContext(cmd *cobra.Command, args []string) error {
	contextName := args[0]
	kafkaAuth, err := c.Flags().GetBool("kafka-auth")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	if !kafkaAuth {
		return errors.HandleCommon(errors.New("only kafka-auth is currently supported"), cmd)
	}
	bootstrapURL, err := c.Flags().GetString("bootstrap")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	bootstrapURL, err = c.resolver.ValueFrom(bootstrapURL, "Bootstrap URL: ", false)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	bootstrapURL = strings.TrimSpace(bootstrapURL)
	apiKey, err := c.Flags().GetString("api-key")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	apiKey, err = c.resolver.ValueFrom(apiKey, "API Key: ", false)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	apiKey = strings.TrimSpace(apiKey)
	apiSecret, err := c.Flags().GetString("api-secret")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	apiSecret, err = c.resolver.ValueFrom(apiSecret, "API Secret: ", true)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	apiSecret = strings.TrimSpace(apiSecret)
	err = c.addContext(contextName, bootstrapURL, apiKey, apiSecret)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	// Set current context.
	err = c.config.SetContext(contextName)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	return nil
}

func (c *command) addContext(name string, bootstrapURL string, apiKey string, apiSecret string) error {
	apiKeyPair := &config.APIKeyPair{
		Key:    apiKey,
		Secret: apiSecret,
	}
	apiKeys := map[string]*config.APIKeyPair{
		apiKey: apiKeyPair,
	}
	kafkaClusterCfg := &config.KafkaClusterConfig{
		ID:          "anonymous-id",
		Name:        "anonymous-cluster",
		Bootstrap:   bootstrapURL,
		APIEndpoint: "",
		APIKeys:     apiKeys,
		APIKey:      apiKey,
	}
	kafkaClusters := map[string]*config.KafkaClusterConfig{
		kafkaClusterCfg.ID: kafkaClusterCfg,
	}
	platform := &config.Platform{Server: bootstrapURL}
	// Hardcoded for now, since username/password isn't implemented yet.
	credential := &config.Credential{
		Username:       "",
		Password:       "",
		APIKeyPair:     apiKeyPair,
		CredentialType: config.APIKey,
	}
	return c.config.AddContext(name, platform, credential, kafkaClusters,
		kafkaClusterCfg.ID, nil)
}
