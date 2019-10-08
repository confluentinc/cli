package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/atrox/homedir"
	v1 "github.com/confluentinc/ccloudapis/org/v1"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/metric"
)

const (
	defaultConfigFileFmt = "~/.%s/config.json"
)

// AuthConfig represents an authenticated user.
type AuthConfig struct {
	User     *v1.User      `json:"user" hcl:"user"`
	Account  *v1.Account   `json:"account" hcl:"account"`
	Accounts []*v1.Account `json:"accounts" hcl:"accounts"`
}

// Config represents the CLI configuration.
type Config struct {
	CLIName        string                   `json:"-" hcl:"-"`
	MetricSink     metric.Sink              `json:"-" hcl:"-"`
	Logger         *log.Logger              `json:"-" hcl:"-"`
	Filename       string                   `json:"-" hcl:"-"`
	Platforms      map[string]*Platform     `json:"platforms" hcl:"platforms"`
	Credentials    map[string]*Credential   `json:"credentials" hcl:"credentials"`
	Contexts       map[string]*Context      `json:"contexts" hcl:"contexts"`
	ContextStates  map[string]*ContextState `json:"context_states" hcl:"context_states"`
	CurrentContext string                   `json:"current_context" hcl:"current_context"`
}

// New initializes a new Config object
func New(config ...*Config) *Config {
	var c *Config
	if config == nil {
		c = &Config{}
	} else {
		c = config[0]
	}
	if c.CLIName == "" {
		// HACK: this is a workaround while we're building multiple binaries off one codebase
		c.CLIName = "confluent"
	}
	c.Platforms = map[string]*Platform{}
	c.Credentials = map[string]*Credential{}
	c.Contexts = map[string]*Context{}
	c.ContextStates = map[string]*ContextState{}
	return c
}

// Load reads the CLI config from disk.
func (c *Config) Load() error {
	filename, err := c.getFilename()
	if err != nil {
		return err
	}
	input, err := ioutil.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			// Save a default version if none exists yet.
			if err := c.Save(); err != nil {
				return errors.Wrapf(err, "unable to create config: %v", err)
			}
			return nil
		}
		return errors.Wrapf(err, "unable to read config file: %s", filename)
	}
	err = json.Unmarshal(input, c)
	if err != nil {
		return errors.Wrapf(err, "unable to parse config file: %s", filename)
	}
	for _, context := range c.Contexts {
		context.State = c.ContextStates[context.Name]
		context.Credential = c.Credentials[context.CredentialName]
		context.Platform = c.Platforms[context.PlatformName]
	}
	err = c.Validate()
	if err != nil {
		return err
	}
	return nil
}

// Save writes the CLI config to disk.
func (c *Config) Save() error {
	err := c.Validate()
	if err != nil {
		return err
	}
	cfg, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return errors.Wrapf(err, "unable to marshal config")
	}
	filename, err := c.getFilename()
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Dir(filename), 0700)
	if err != nil {
		return errors.Wrapf(err, "unable to create config directory: %s", filename)
	}
	err = ioutil.WriteFile(filename, cfg, 0600)
	if err != nil {
		return errors.Wrapf(err, "unable to write config to file: %s", filename)
	}
	return nil
}

func (c *Config) Validate() error {
	// Validate that current context exists.
	if c.CurrentContext != "" {
		if _, ok := c.Contexts[c.CurrentContext]; !ok {
			c.Logger.Trace("current context does not exist")
			return errors.Errorf("the current context does not exist.")
		}
	}
	// Validate that every context:
	// 1. Has a credential, a platform, and a state. 
	// 2. Has no hanging references between the context and the config.
	// 3. Is mapped by name correctly in the config.
	for _, context := range c.Contexts {
		err := context.Validate()
		if err != nil {
			c.Logger.Trace("context validation error")
			return err
		}
		if _, ok := c.Credentials[context.CredentialName]; !ok || context.CredentialName == "" {
			c.Logger.Trace("unspecified credential error")
			return &errors.UnspecifiedCredentialError{ContextName: context.Name}
		}
		if _, ok := c.Platforms[context.PlatformName]; !ok || context.PlatformName == "" {
			c.Logger.Trace("unspecified platform error")
			return &errors.UnspecifiedPlatformError{ContextName: context.Name}
		}
		if _, ok := c.ContextStates[context.Name]; !ok {
			c.ContextStates[context.Name] = new(ContextState)
		}

	}
	// Validate that all context states are mapped to an existing context.
	for contextName, _ := range c.ContextStates {
		if _, ok := c.Contexts[contextName]; !ok {
			c.Logger.Trace("context state mapped to nonexistent context")
			return c.corruptedConfigError()
		}
	}
	return nil
}

// DeleteContext deletes the specified context, and returns an error if it's not found.
func (c *Config) DeleteContext(name string) error {
	_, err := c.FindContext(name)
	if err != nil {
		return err
	}
	delete(c.Contexts, name)
	if c.CurrentContext == name {
		c.CurrentContext = ""
	}
	delete(c.ContextStates, name)
	return nil
}

// FindContext finds a context by name, and returns nil if not found.
func (c *Config) FindContext(name string) (*Context, error) {
	context, ok := c.Contexts[name]
	if !ok {
		return nil, fmt.Errorf("context \"%s\" does not exist", name)
	}
	return context, nil
}

func (c *Config) AddContext(name string, platformName string, credentialName string,
	kafkaClusters map[string]*KafkaClusterConfig, kafka string,
	schemaRegistryClusters map[string]*SchemaRegistryCluster, state *ContextState) error {
	if _, ok := c.Contexts[name]; ok {
		return fmt.Errorf("context \"%s\" already exists", name)
	}
	credential, ok := c.Credentials[credentialName]
	if !ok {
		return fmt.Errorf("credential \"%s\" not found", credentialName)
	}
	platform, ok := c.Platforms[platformName]
	if !ok {
		return fmt.Errorf("platform \"%s\" not found", platformName)
	}
	context, err := newContext(name, platform, credential, kafkaClusters, kafka,
		schemaRegistryClusters, state)
	if err != nil {
		return err
	}
	c.Contexts[name] = context
	c.ContextStates[name] = context.State
	err = c.Validate()
	if err != nil {
		return err
	}
	if c.CurrentContext == "" {
		c.CurrentContext = context.Name
	}
	return c.Save()
}

func (c *Config) SetContext(name string) error {
	_, err := c.FindContext(name)
	if err != nil {
		return err
	}
	c.CurrentContext = name
	return c.Save()
}

// Name returns the display name for the CLI
func (c *Config) Name() string {
	name := "Confluent CLI"
	if c.CLIName == "ccloud" {
		name = "Confluent Cloud CLI"
	}
	return name
}

func (c *Config) SaveCredential(credential *Credential) error {
	if credential.Name == "" {
		return fmt.Errorf("credential must have a name")
	}
	c.Credentials[credential.Name] = credential
	return c.Save()
}

func (c *Config) SavePlatform(platform *Platform) error {
	if platform.Name == "" {
		return fmt.Errorf("platform must have a name")
	}
	c.Platforms[platform.Name] = platform
	return c.Save()
}

func (c *Config) Support() string {
	support := "https://confluent.io; support@confluent.io"
	if c.CLIName == "ccloud" {
		support = "https://confluent.cloud; support@confluent.io"
	}
	return support
}

// APIName returns the display name of the remote API
// (e.g., Confluent Platform or Confluent Cloud)
func (c *Config) APIName() string {
	name := "Confluent Platform"
	if c.CLIName == "ccloud" {
		name = "Confluent Cloud"
	}
	return name
}

// Context returns the current Context, or nil if there's no context set.
func (c *Config) Context() *Context {
	return c.Contexts[c.CurrentContext]
}

func (c *Config) AuthenticatedState() (*ContextState, error) {
	context := c.Context()
	if context == nil {
		return nil, errors.ErrNoContext
	}
	return context.authenticatedState()
}

// SchemaRegistryCluster returns the SchemaRegistryCluster for the current Context,
// or an empty SchemaRegistryCluster if there is none set,
// or an error if no context exists/if the user is not logged in.
func (c *Config) SchemaRegistryCluster() (*SchemaRegistryCluster, error) {
	context := c.Context()
	if context == nil {
		return nil, errors.ErrNoContext
	}
	return context.schemaRegistryCluster()
}

// CheckHasAPIKey returns nil if the specified cluster exists in the current context
// and has an active API key, error otherwise.
func (c *Config) CheckHasAPIKey(clusterID string) error {
	context := c.Context()
	if context == nil {
		return errors.ErrNoContext
	}
	cluster, found := context.KafkaClusters[clusterID]
	if !found {
		return fmt.Errorf("unknown kafka cluster: %s", clusterID)
	}
	if cluster.APIKey == "" {
		return &errors.UnspecifiedAPIKeyError{ClusterID: clusterID}
	}
	return nil
}

func (c *Config) getFilename() (string, error) {
	if c.Filename == "" {
		c.Filename = fmt.Sprintf(defaultConfigFileFmt, c.CLIName)
	}
	filename, err := homedir.Expand(c.Filename)
	if err != nil {
		c.Logger.Error(err)
		// Return a more user-friendly error.
		err = fmt.Errorf("an error resolving the config filepath at %s has occurred. "+
			"Please try moving the file to a different location", c.Filename)
		return "", err
	}
	return filename, nil
}

// corruptedConfigError returns an error signaling that the config file has been corrupted,
// or another error if the config's filepath is unable to be resolved.
func (c *Config) corruptedConfigError() error {
	configPath, err := c.getFilename()
	if err != nil {
		return err
	}
	errMsg := "the config file located at %s has been corrupted. " +
		"To fix, please remove the config file, and run `login` or `init`"
	err = fmt.Errorf(errMsg, configPath)
	return err
}

// corruptedContextError returns an error signaling that the specified context's,
// config has been corrupted, or another error if the config's filepath is unable to be resolved.
func (c *Config) corruptedContextError(contextName string) error {
	configPath, err := c.getFilename()
	if err != nil {
		return err
	}
	errMsg := "the configuration of context '%s' has been corrupted. " +
		"To fix, please remove the config file located at %s, and run `login` or `init`"
	err = fmt.Errorf(errMsg, contextName, configPath)
	return err
}
