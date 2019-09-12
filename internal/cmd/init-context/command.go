package init

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/errors"
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
	c.Flags().Bool("kafka-auth", false, "Initialize with bootstrap url, API key, and API secret. "+
		"Can be done interactively, with flags, or both.")
	c.Flags().String("bootstrap", "", "Bootstrap URL.")
	c.Flags().String("api-key", "", "API key.")
	c.Flags().String("api-secret", "", "API secret file, starting with '@'.")
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
	eh.Handle(c.addContext(contextName, bootstrapURL, apiKey, apiSecret))
	// Set current context.
	eh.Handle(c.config.SetContext(contextName))
	if err := eh.Reset(); err != nil {
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
