package init

import (
	"fmt"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/spf13/cobra"
	"strings"
)

type command struct {
	*cobra.Command
	Config *config.Config
	prompt pcmd.Prompt
}

const longDescription = "Initialize and set a current context."

func New(prerunner pcmd.PreRunner, config *config.Config, prompt pcmd.Prompt) *cobra.Command {
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
	}
	cmd.init()
	return cmd.Command
}

func (c *command) init() {
	c.Flags().Bool("kafka-auth", false, "Interactively initialize with bootstrap url, api key, and api secret.")
	c.RunE = c.initContext
}

func (c *command) initContext(cmd *cobra.Command, args []string) error {
	contextName := args[0]
	interactive, err := c.Flags().GetBool("kafka-auth")
	if !interactive {
		return errors.HandleCommon(errors.New("Must be interactive"), c.Command)
	}
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	pcmd.Print(cmd, "Bootstrap URL: ")
	bootstrapURL, err := c.prompt.ReadString('\n')
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	bootstrapURL = strings.TrimSpace(bootstrapURL)
	if len(bootstrapURL) == 0 {
		return errors.HandleCommon(fmt.Errorf("%s cannot be empty", "Bootstrap url"), cmd)
	}
	pcmd.Print(cmd, "API Key: ")
	apiKey, err := c.prompt.ReadString('\n')
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	apiKey = strings.TrimSpace(apiKey)
	if len(apiKey) == 0 {
		return errors.HandleCommon(fmt.Errorf("%s cannot be empty", "API key"), cmd)
	}
	pcmd.Print(cmd, "API Secret: ")
	apiSecret, err := c.prompt.ReadString('\n')
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	apiSecret = strings.TrimSpace(apiSecret)
	if len(apiSecret) == 0 {
		return errors.HandleCommon(fmt.Errorf("%s cannot be empty", "API secret"), cmd)
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
	const anonClusterId = "anonymous-id"
	const anonClusterName = "anonymous-cluster"
	apiKeyPair := &config.APIKeyPair{
		Key:    apiKey,
		Secret: apiSecret,
	}
	apiKeys := map[string]*config.APIKeyPair{
		apiKey: apiKeyPair,
	}
	kafkaClusterCfg := &config.KafkaClusterConfig{
		ID:          anonClusterId,
		Name:        anonClusterName,
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
	return c.Config.AddContext(name, platform, credential, kafkaClusters,
		kafkaClusterCfg.ID, nil)
}
