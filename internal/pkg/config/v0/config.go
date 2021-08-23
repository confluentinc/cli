package v0

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/blang/semver"

	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/errors"
	pversion "github.com/confluentinc/cli/internal/pkg/version"
)

const (
	defaultConfigFileFmt = "%s/.confluent/config.json"
)

var (
	Version = semver.MustParse("0.0.0")
)

// Config represents the CLI configuration.
type Config struct {
	*config.BaseConfig
	AuthURL        string                 `json:"auth_url" hcl:"auth_url"`
	AuthToken      string                 `json:"auth_token" hcl:"auth_token"`
	Auth           *AuthConfig            `json:"auth" hcl:"auth"`
	Platforms      map[string]*Platform   `json:"platforms" hcl:"platforms"`
	Credentials    map[string]*Credential `json:"credentials" hcl:"credentials"`
	Contexts       map[string]*Context    `json:"contexts" hcl:"contexts"`
	CurrentContext string                 `json:"current_context" hcl:"current_context"`
}

// NewBaseConfig initializes a new Config object
func New(params *config.Params) *Config {
	return &Config{
		BaseConfig:  config.NewBaseConfig(params, Version),
		Platforms:   make(map[string]*Platform),
		Credentials: make(map[string]*Credential),
		Contexts:    make(map[string]*Context),
	}
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
	return nil
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

// V0 Config does not have validation functionality.
func (c *Config) Validate() error {
	return nil
}

// Binary returns the display name for the CLI
func (c *Config) Name() string {
	return pversion.FullCLIName
}

// Context returns the current Context object.
func (c *Config) Context() (*Context, error) {
	if c.CurrentContext == "" {
		return nil, new(errors.NotLoggedInError)
	}
	return c.Contexts[c.CurrentContext], nil
}

func (c *Config) SchemaRegistryCluster() (*SchemaRegistryCluster, error) {
	context, err := c.Context()
	if err != nil {
		return nil, err
	}
	if c.Auth == nil || c.Auth.Account == nil {
		return nil, new(errors.NotLoggedInError)
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

// KafkaClusterConfig returns the KafkaClusterConfig for the current Context
// or nil if there is none set.
func (c *Config) KafkaClusterConfig() (*KafkaClusterConfig, error) {
	context, err := c.Context()
	if err != nil {
		return nil, err
	}
	kafka := context.Kafka
	if kafka == "" {
		return nil, nil
	} else {
		return context.KafkaClusters[kafka], nil
	}
}

// CheckLogin returns an error if the user is not logged in.
func (c *Config) CheckLogin() error {
	if c.AuthToken == "" && (c.Auth == nil || c.Auth.Account == nil || c.Auth.Account.Id == "") {
		return new(errors.NotLoggedInError)
	}
	return nil
}

func (c *Config) CheckHasAPIKey(clusterID string) error {
	cfg, err := c.Context()
	if err != nil {
		return err
	}

	cluster, found := cfg.KafkaClusters[clusterID]
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
		homedir, _ := os.UserHomeDir()
		c.Filename = filepath.FromSlash(fmt.Sprintf(defaultConfigFileFmt, homedir))
	}
	return c.Filename, nil
}
