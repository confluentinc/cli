package init

import (
	"fmt"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/spf13/cobra"
	"io/ioutil"
	"strings"
)

type command struct {
	*cobra.Command
	Config *config.Config
	prompt pcmd.Prompt
}

// TODO: Make long description better.
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
	c.Flags().Bool("kafka-auth", false, "Interactively initialize with bootstrap url, API key, and API secret.")
	c.Flags().String("bootstrap", "", "Bootstrap URL.")
	c.Flags().String("api-key", "", "API key.")
	c.Flags().String("api-secret", "", "API secret file, starting with '@'.")
	c.Flags().SortFlags = false
	c.RunE = c.initContext
}

func (c *command) promptForNonemptyString(msg string) (string, error) {
	pcmd.Print(c.Command, fmt.Sprintf("%s: ", msg))
	val, err := c.prompt.ReadString('\n')
	if err != nil {
		return "", err
	}
	val = strings.TrimSpace(val)
	if len(val) == 0 {
		return "", fmt.Errorf("%s cannot be empty", msg)
	}
	return val, nil
}

func (c *command) initContext(cmd *cobra.Command, args []string) error {
	contextName := args[0]
	kafkaAuth, err := c.Flags().GetBool("kafka-auth")
	if err != nil {
		return errors.HandleCommon(err, c.Command)
	}
	if !kafkaAuth {
		return errors.HandleCommon(errors.New("only kafka-auth is currently supported"), cmd)
	}
	bootstrapURL, err := c.Flags().GetString("bootstrap")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	apiKey, err := c.Flags().GetString("api-key")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	apiSecretFilename, err := c.Flags().GetString("api-secret")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	if len(bootstrapURL) == 0 {
		bootstrapURL, err = c.promptForNonemptyString("Bootstrap URL")
		if err != nil {
			return errors.HandleCommon(err, cmd)
		}
	}
	if len(apiKey) == 0 {
		apiKey, err = c.promptForNonemptyString("API key")
		if err != nil {
			return errors.HandleCommon(err, cmd)
		}
	}
	var apiSecret string
	if len(apiSecretFilename) == 0 {
		apiSecret, err = c.promptForNonemptyString("API secret")
		if err != nil {
			return errors.HandleCommon(err, cmd)
		}
	} else {
		// Read API secret from file.
		apiSecretFilename = apiSecretFilename[1:]
		apiSecretBytes, err := ioutil.ReadFile(apiSecretFilename)
		if err != nil {
			return errors.HandleCommon(err, cmd)
		}
		apiSecret = string(apiSecretBytes)
		if len(apiSecret) == 0 {
			return fmt.Errorf("%s cannot be empty", "API secret")
		}
		apiSecret = strings.TrimSpace(apiSecret)
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
