package v3

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"

	"github.com/atrox/homedir"
	"github.com/blang/semver"
	"github.com/google/uuid"

	"github.com/confluentinc/cli/internal/pkg/config"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	v2 "github.com/confluentinc/cli/internal/pkg/config/v2"
	"github.com/confluentinc/cli/internal/pkg/errors"
)

const (
	defaultConfigFileFmt = "~/.%s/config.json"
	emptyFieldIndicator  = "EMPTY"
)

var (
	Version = semver.MustParse("3.0.0")
)

// Config represents the CLI configuration.
type Config struct {
	*config.BaseConfig
	DisableUpdateCheck     bool                        `json:"disable_update_check"`
	DisableUpdates         bool                        `json:"disable_updates"`
	NoBrowser              bool                        `json:"no_browser" hcl:"no_browser"`
	Platforms              map[string]*v2.Platform     `json:"platforms,omitempty"`
	Credentials            map[string]*v2.Credential   `json:"credentials,omitempty"`
	Contexts               map[string]*Context         `json:"contexts,omitempty"`
	ContextStates          map[string]*v2.ContextState `json:"context_states,omitempty"`
	CurrentContext         string                      `json:"current_context"`
	AnonymousId            string                      `json:"anonymous_id,omitempty"`
	overwrittenAccount     *orgv1.Account
	overwrittenCurrContext string
	overwrittenActiveKafka string
}

func (c *Config) SetOverwrittenAccount(acct *orgv1.Account) {
	if c.overwrittenAccount == nil {
		c.overwrittenAccount = acct
	}
}

func (c *Config) SetOverwrittenCurrContext(contextName string) {
	if contextName == "" {
		contextName = emptyFieldIndicator
	}
	if c.overwrittenCurrContext == "" {
		c.overwrittenCurrContext = contextName
	}
}

func (c *Config) SetOverwrittenActiveKafka(clusterId string) {
	if clusterId == "" {
		clusterId = emptyFieldIndicator
	}
	if c.overwrittenActiveKafka == "" {
		c.overwrittenActiveKafka = clusterId
	}
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
	c.Platforms = map[string]*v2.Platform{}
	c.Credentials = map[string]*v2.Credential{}
	c.Contexts = map[string]*Context{}
	c.ContextStates = map[string]*v2.ContextState{}
	c.AnonymousId = uuid.New().String()
	return c
}

// Load reads the CLI config from disk.
// Save a default version if none exists yet.
func (c *Config) Load() error {
	currentVersion := Version
	filename, err := c.getFilename()
	c.Filename = filename
	if err != nil {
		return err
	}
	input, err := ioutil.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			// Save a default version if none exists yet.
			if err := c.Save(); err != nil {
				return errors.Wrapf(err, errors.UnableToCreateConfigErrorMsg)
			}
			return nil
		}
		return errors.Wrapf(err, errors.UnableToReadConfigErrorMsg, filename)
	}
	err = json.Unmarshal(input, c)
	if c.Ver.Compare(currentVersion) < 0 {
		return errors.Errorf(errors.ConfigNotUpToDateErrorMsg, c.Ver, currentVersion)
	} else if c.Ver.Compare(Version) > 0 {
		return errors.Errorf(errors.InvalidConfigVersionErrorMsg, c.Ver)
	}
	if err != nil {
		return errors.Wrapf(err, errors.ParseConfigErrorMsg, filename)
	}
	for _, context := range c.Contexts {
		// Some "pre-validation"
		if context.Name == "" {
			return errors.NewCorruptedConfigError(errors.NoNameContextErrorMsg, "", c.CLIName, c.Filename, c.Logger)
		}
		if context.CredentialName == "" {
			return errors.NewCorruptedConfigError(errors.UnspecifiedCredentialErrorMsg, context.Name, c.CLIName, c.Filename, c.Logger)
		}
		if context.PlatformName == "" {
			return errors.NewCorruptedConfigError(errors.UnspecifiedPlatformErrorMsg, context.Name, c.CLIName, c.Filename, c.Logger)
		}
		context.State = c.ContextStates[context.Name]
		context.Credential = c.Credentials[context.CredentialName]
		context.Platform = c.Platforms[context.PlatformName]
		context.Logger = c.Logger
		context.Config = c
		if context.KafkaClusterContext == nil {
			return errors.NewCorruptedConfigError(errors.MissingKafkaClusterContextErrorMsg, context.Name, c.CLIName, c.Filename, c.Logger)
		}
		context.KafkaClusterContext.Context = context
	}
	err = c.Validate()
	if err != nil {
		return err
	}
	return nil
}

// Save writes the CLI config to disk.
func (c *Config) Save() error {
	tempKafka := c.resolveOverwrittenKafka()
	tempAccount := c.resolveOverwrittenAccount()
	tempContext := c.resolveOverwrittenContext()
	err := c.Validate()
	if err != nil {
		return err
	}

	cfg, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return errors.Wrapf(err, errors.MarshalConfigErrorMsg)
	}
	filename, err := c.getFilename()
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Dir(filename), 0700)
	if err != nil {
		return errors.Wrapf(err, errors.CreateConfigDirectoryErrorMsg, filename)
	}
	err = ioutil.WriteFile(filename, cfg, 0600)
	if err != nil {
		return errors.Wrapf(err, errors.CreateConfigFileErrorMsg, filename)
	}
	c.restoreOverwrittenContext(tempContext)
	c.restoreOverwrittenAccount(tempAccount)
	c.restoreOverwrittenKafka(tempKafka)
	return nil
}

// If active Kafka cluster has been overwritten by flag value; if so, replace with previous active kafka
// Return the flag value so that it can be restored after writing to file so that continued execution uses flag value
// This prevents flags from updating state
func (c *Config) resolveOverwrittenKafka() string {
	ctx := c.Context()
	var tempKafka string
	if c.overwrittenActiveKafka != "" && ctx != nil && ctx.KafkaClusterContext != nil {
		if c.overwrittenActiveKafka == emptyFieldIndicator {
			c.overwrittenActiveKafka = ""
		}
		tempKafka = ctx.KafkaClusterContext.GetActiveKafkaClusterId()
		ctx.KafkaClusterContext.SetActiveKafkaCluster(c.overwrittenActiveKafka)
	}
	return tempKafka
}

// Restore the flag cluster back into the struct so that it is used for any execution after Save()
func (c *Config) restoreOverwrittenKafka(tempKafka string) {
	ctx := c.Context()
	if tempKafka != "" {
		ctx.KafkaClusterContext.SetActiveKafkaCluster(tempKafka)
	}
}

// Switch the initial config context back into the struct so that it is saved and not the flag value
// Return the overwriting flag context value so that it can be restored after writing the file
func (c *Config) resolveOverwrittenContext() string {
	var tempContext string
	if c.overwrittenCurrContext != "" && c != nil {
		if c.overwrittenCurrContext == emptyFieldIndicator {
			c.overwrittenCurrContext = ""
		}
		tempContext = c.CurrentContext
		c.CurrentContext = c.overwrittenCurrContext
	}
	return tempContext
}

// Restore the flag context back into the struct so that it is used for any execution after Save()
func (c *Config) restoreOverwrittenContext(tempContext string) {
	if tempContext != "" {
		c.CurrentContext = tempContext
	}
}

// Switch the initial config account back into the struct so that it is saved and not the flag value
// Return the overwriting flag account value so that it can be restored after writing the file
func (c *Config) resolveOverwrittenAccount() *orgv1.Account {
	ctx := c.Context()
	var tempAccount *orgv1.Account
	if c.overwrittenAccount != nil && ctx != nil && ctx.State != nil && ctx.State.Auth != nil {
		tempAccount = ctx.State.Auth.Account
		ctx.State.Auth.Account = c.overwrittenAccount
	}
	return tempAccount
}

// Restore the flag account back into the struct so that it is used for any execution after Save()
func (c *Config) restoreOverwrittenAccount(tempAccount *orgv1.Account) {
	ctx := c.Context()
	if tempAccount != nil {
		ctx.State.Auth.Account = tempAccount
	}
}

func (c *Config) Validate() error {
	// Validate that current context exists.
	if c.CurrentContext != "" {
		if _, ok := c.Contexts[c.CurrentContext]; !ok {
			c.Logger.Trace("current context does not exist")
			return errors.NewCorruptedConfigError(errors.CurrentContextNotExistErrorMsg, c.CurrentContext, c.CLIName, c.Filename, c.Logger)
		}
	}
	// Validate that every context:
	// 1. Has no hanging references between the context and the config.
	// 2. Is mapped by name correctly in the config.
	for _, context := range c.Contexts {
		err := context.validate()
		if err != nil {
			c.Logger.Trace("context validation error")
			return err
		}
		if _, ok := c.Credentials[context.CredentialName]; !ok {
			c.Logger.Trace("unspecified credential error")
			return errors.NewCorruptedConfigError(errors.UnspecifiedCredentialErrorMsg, context.Name, c.CLIName, c.Filename, c.Logger)
		}
		if _, ok := c.Platforms[context.PlatformName]; !ok {
			c.Logger.Trace("unspecified platform error")
			return errors.NewCorruptedConfigError(errors.UnspecifiedPlatformErrorMsg, context.Name, c.CLIName, c.Filename, c.Logger)
		}
		if _, ok := c.ContextStates[context.Name]; !ok {
			c.ContextStates[context.Name] = new(v2.ContextState)
		}
		if *c.ContextStates[context.Name] != *context.State {
			c.Logger.Trace(fmt.Sprintf("state of context %s in config does not match actual state of context", context.Name))
			return errors.NewCorruptedConfigError(errors.ContextStateMismatchErrorMsg, context.Name, c.CLIName, c.Filename, c.Logger)
		}
	}
	// Validate that all context states are mapped to an existing context.
	for contextName := range c.ContextStates {
		if _, ok := c.Contexts[contextName]; !ok {
			c.Logger.Trace("context state mapped to nonexistent context")
			return errors.NewCorruptedConfigError(errors.ContextStateNotMappedErrorMsg, contextName, c.CLIName, c.Filename, c.Logger)
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
		return nil, fmt.Errorf(errors.ContextNotExistErrorMsg, name)
	}
	return context, nil
}

func (c *Config) AddContext(name string, platformName string, credentialName string,
	kafkaClusters map[string]*v1.KafkaClusterConfig, kafka string,
	schemaRegistryClusters map[string]*v2.SchemaRegistryCluster, state *v2.ContextState) error {
	if _, ok := c.Contexts[name]; ok {
		return fmt.Errorf(errors.ContextNameExistsErrorMsg, name)
	}
	return c.BuildAndSaveContext(name, platformName, credentialName, kafkaClusters, kafka, schemaRegistryClusters, state)
}

func (c *Config) BuildAndSaveContext(name string, platformName string, credentialName string,
	kafkaClusters map[string]*v1.KafkaClusterConfig, kafka string,
	schemaRegistryClusters map[string]*v2.SchemaRegistryCluster, state *v2.ContextState) error {

	credential, ok := c.Credentials[credentialName]
	if !ok {
		return fmt.Errorf(errors.CredentialNotFoundErrorMsg, credentialName)
	}
	platform, ok := c.Platforms[platformName]
	if !ok {
		return fmt.Errorf(errors.PlatformNotFoundErrorMsg, platformName)
	}
	context, err := newContext(name, platform, credential, kafkaClusters, kafka,
		schemaRegistryClusters, state, c)
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

func (c *Config) SaveCredential(credential *v2.Credential) error {
	if credential.Name == "" {
		return errors.New(errors.NoNameCredentialErrorMsg)
	}
	c.Credentials[credential.Name] = credential
	return c.Save()
}

func (c *Config) SavePlatform(platform *v2.Platform) error {
	if platform.Name == "" {
		return errors.New(errors.NoNamePlatformErrorMsg)
	}
	c.Platforms[platform.Name] = platform
	return c.Save()
}

// Context returns the user specified context if it exists,
// the current Context, or nil if there's no context set.
func (c *Config) Context() *Context {
	if c == nil {
		return nil
	}
	return c.Contexts[c.CurrentContext]
}

func (c *Config) CredentialType() v2.CredentialType {
	ctx := c.Context()
	if ctx == nil {
		return v2.None
	}
	if c.Context().Credential.CredentialType == v2.APIKey {
		return v2.APIKey
	}
	if c.HasLogin() {
		return v2.Username
	}
	return v2.None
}

func (c *Config) HasLogin() bool {
	ctx := c.Context()
	if ctx == nil {
		return false
	}
	if c.CLIName == "ccloud" {
		return ctx.hasCCloudLogin()
	}
	return ctx.HasMDSLogin()
}

func (c *Config) ResetAnonymousId() error {
	c.AnonymousId = uuid.New().String()
	return c.Save()
}

func (c *Config) getFilename() (string, error) {
	if c.Filename == "" {
		c.Filename = fmt.Sprintf(defaultConfigFileFmt, c.CLIName)
	}
	filename, err := homedir.Expand(c.Filename)
	if err != nil {
		c.Logger.Error(err)
		// Return a more user-friendly error.
		return "", errors.NewErrorWithSuggestions(fmt.Sprintf(errors.ResolvingConfigPathErrorMsg, c.Filename), errors.ResolvingConfigPathSuggestions)
	}
	return filename, nil
}
