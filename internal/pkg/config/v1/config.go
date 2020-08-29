package v1

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/atrox/homedir"
	"github.com/blang/semver"
	"github.com/google/uuid"

	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/errors"
)

const (
	defaultConfigFileFmt = "~/.%s/config.json"
)

var (
	Version = semver.MustParse("1.0.0")
)

// Config represents the CLI configuration.
type Config struct {
	*config.BaseConfig
	DisableUpdateCheck bool                   `json:"disable_update_check"`
	DisableUpdates     bool                   `json:"disable_updates"`
	AuthURL            string                 `json:"auth_url"`
	NoBrowser          bool                   `json:"no_browser" hcl:"no_browser"`
	AuthToken          string                 `json:"auth_token"`
	Auth               *AuthConfig            `json:"auth"`
	Platforms          map[string]*Platform   `json:"platforms"`
	Credentials        map[string]*Credential `json:"credentials"`
	Contexts           map[string]*Context    `json:"contexts"`
	CurrentContext     string                 `json:"current_context"`
	AnonymousId        string
}

// NewBaseConfig initializes a new Config object
func New(params *config.Params) *Config {
	c := &Config{}
	baseCfg := config.NewBaseConfig(params, Version)
	c.BaseConfig = baseCfg
	if c.CLIName == "" {
		// HACK: this is a workaround while we're building multiple binaries off one codebase
		c.CLIName = "confluent"
	}
	c.Platforms = map[string]*Platform{}
	c.Credentials = map[string]*Credential{}
	c.Contexts = map[string]*Context{}
	c.AnonymousId = uuid.New().String()
	return c
}

// Load reads the CLI config from disk.
func (c *Config) Load() error {
	filename, err := c.getFilename()
	if err != nil {
		return err
	}
	c.Filename = filename
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
	return c.Validate()
}

// Save writes the CLI config to disk.
func (c *Config) Save() error {
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

// V1 Config does not have validation functionality.
func (c *Config) Validate() error {
	// Hack to differentiate between v0 and v1 configs retroactively.
	for _, context := range c.Contexts {
		if context.Name == "" {
			return errors.New("context has no name")
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
	return nil
}

// FindContext finds a context by name,
// and returns a formatted error if not found.
func (c *Config) FindContext(name string) (*Context, error) {
	context, ok := c.Contexts[name]
	if !ok {
		return nil, fmt.Errorf("context \"%s\" does not exist", name)
	}
	return context, nil
}

func newContext(name string, platform *Platform, credential *Credential,
	kafkaClusters map[string]*KafkaClusterConfig, kafka string,
	schemaRegistryClusters map[string]*SchemaRegistryCluster) *Context {
	return &Context{
		Name:                   name,
		Platform:               platform.String(),
		Credential:             credential.String(),
		KafkaClusters:          kafkaClusters,
		Kafka:                  kafka,
		SchemaRegistryClusters: schemaRegistryClusters,
	}
}

func (c *Config) AddContext(name string, platform *Platform, credential *Credential,
	kafkaClusters map[string]*KafkaClusterConfig, kafka string,
	schemaRegistryClusters map[string]*SchemaRegistryCluster) error {
	if _, ok := c.Contexts[name]; ok {
		return fmt.Errorf("context \"%s\" already exists", name)
	}
	context := newContext(name, platform, credential, kafkaClusters, kafka,
		schemaRegistryClusters)
	// Update config maps.
	c.Contexts[name] = context
	c.Credentials[context.Credential] = credential
	c.Platforms[context.Platform] = platform
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

// Context returns the current Context object.
func (c *Config) Context() (*Context, error) {
	if c.CurrentContext == "" {
		return nil, &errors.NoContextError{CLIName: c.CLIName}
	}
	context, err := c.FindContext(c.CurrentContext)
	if err != nil {
		return nil, err
	}
	return context, nil
}

// CredentialType returns the credential type of the current Context.
// It returns ErrNoContext if there's no current context,
// or UnspecifiedCredentialError if there is a current context with no credentials,
// informing the user the config file has been corrupted.
func (c *Config) CredentialType() (CredentialType, error) {
	context, err := c.Context()
	if err != nil {
		return -1, err
	}
	if cred, ok := c.Credentials[context.Credential]; ok {
		return cred.CredentialType, nil
	}
	return -1, errors.NewCorruptedConfigError(errors.UnspecifiedCredentialErrorMsg, c.CurrentContext, c.CLIName, c.Filename, c.Logger)
}

// SchemaRegistryCluster returns the SchemaRegistryCluster for the current Context,
// or an empty SchemaRegistryCluster if there is none set,
// or an error if no context exists/if the user is not logged in.
func (c *Config) SchemaRegistryCluster() (*SchemaRegistryCluster, error) {
	context, err := c.Context()
	if err != nil {
		return nil, err
	}
	if c.Auth == nil || c.Auth.Account == nil {
		return nil, &errors.NotLoggedInError{CLIName: c.CLIName}
	}
	sr := context.SchemaRegistryClusters[c.Auth.Account.Id]
	if sr == nil {
		if context.SchemaRegistryClusters == nil {
			context.SchemaRegistryClusters = map[string]*SchemaRegistryCluster{}
		}
		context.SchemaRegistryClusters[c.Auth.Account.Id] = &SchemaRegistryCluster{}
	}
	return context.SchemaRegistryClusters[c.Auth.Account.Id], nil
}

// KafkaClusterConfig returns the KafkaClusterConfig for the current Context.
// or nil if there is none set.
func (c *Config) KafkaClusterConfig() (*KafkaClusterConfig, error) {
	context, err := c.Context()
	if err != nil {
		return nil, err
	}
	kafka := context.Kafka
	if kafka == "" {
		return nil, nil
	}
	kcc, ok := context.KafkaClusters[kafka]
	if !ok {
		configPath, err := c.getFilename()
		if err != nil {
			err = fmt.Errorf("an error resolving the config filepath at %s has occurred. "+
				"Please try moving the file to a different location", c.Filename)
			return nil, err
		}
		errMsg := "the configuration of context \"%s\" has been corrupted. " +
			"To fix, please remove the config file located at %s, and run `login` or `init`"
		err = fmt.Errorf(errMsg, context.Name, configPath)
		return nil, err
	}
	return kcc, nil
}

// CheckLogin returns an error if the user is not logged in
// with a username and password.
func (c *Config) CheckLogin() error {
	credType, err := c.CredentialType()
	if err != nil {
		return err
	}
	switch credType {
	case Username:
		if c.AuthToken == "" && (c.Auth == nil || c.Auth.Account == nil || c.Auth.Account.Id == "") {
			return &errors.NotLoggedInError{CLIName: c.CLIName}
		}
	case APIKey:
		return &errors.NotLoggedInError{CLIName: c.CLIName}
	}
	return nil
}

// CheckHasAPIKey returns nil if the specified cluster exists in the current context
// and has an active API key, error otherwise.
func (c *Config) CheckHasAPIKey(clusterID string) error {
	context, err := c.Context()
	if err != nil {
		return err
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

func (c *Config) CheckSchemaRegistryHasAPIKey() bool {
	srCluster, err := c.SchemaRegistryCluster()
	if err != nil {
		return false
	}
	return !(srCluster.SrCredentials == nil || len(srCluster.SrCredentials.Key) == 0 || len(srCluster.SrCredentials.Secret) == 0)
}

func (c *Config) ResetAnonymousId() error {
	c.AnonymousId = uuid.New().String()
	return c.Save()
}

func (c *Config) DeleteUserAuth() error {
	c.AuthToken = ""
	c.Auth = nil
	err := c.Save()
	if err != nil {
		return errors.Wrap(err, "Unable to delete user auth")
	}
	return nil
}

func (c *Config) getFilename() (string, error) {
	if c.Filename == "" {
		c.Filename = fmt.Sprintf(defaultConfigFileFmt, c.CLIName)
	}
	filename, err := homedir.Expand(c.Filename)
	if err != nil {
		return "", err
	}
	return filename, nil
}
