package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"

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

// APIKeyPair holds an API Key and Secret.
type APIKeyPair struct {
	Key    string `json:"api_key" hcl:"api_key"`
	Secret string `json:"api_secret" hcl:"api_secret"`
}

// KafkaClusterConfig represents a connection to a Kafka cluster.
type KafkaClusterConfig struct {
	ID          string                 `json:"id" hcl:"id"`
	Name        string                 `json:"name" hcl:"name"`
	Bootstrap   string                 `json:"bootstrap_servers" hcl:"bootstrap_servers"`
	APIEndpoint string                 `json:"api_endpoint,omitempty" hcl:"api_endpoint"`
	APIKeys     map[string]*APIKeyPair `json:"api_keys" hcl:"api_keys"`
	// APIKey is your active api key for this cluster and references a key in the APIKeys map
	APIKey string `json:"api_key,omitempty" hcl:"api_key"`
}

// Platform represents a Confluent Platform deployment
type Platform struct {
	Server string `json:"server" hcl:"server"`
}

// Credential represent an authentication mechanism for a Platform
type Credential struct {
	Username string
	Password string
}

// Context represents a specific CLI context.
type Context struct {
	Platform   string `json:"platform" hcl:"platform"`
	Credential string `json:"credentials" hcl:"credentials"`
	// KafkaClusters store connection info for interacting directly with Kafka (e.g., consume/produce, etc)
	// N.B. These may later be exposed in the CLI to directly register kafkas (outside a Control Plane)
	KafkaClusters map[string]*KafkaClusterConfig `json:"kafka_clusters" hcl:"kafka_clusters"`
	// Kafka is your active Kafka cluster and references a key in the KafkaClusters map
	Kafka string `json:"kafka_cluster" hcl:"kafka_cluster"`
}

// Config represents the CLI configuration.
type Config struct {
	CLIName        string                 `json:"-" hcl:"-"`
	MetricSink     metric.Sink            `json:"-" hcl:"-"`
	Logger         *log.Logger            `json:"-" hcl:"-"`
	Filename       string                 `json:"-" hcl:"-"`
	AuthURL        string                 `json:"auth_url" hcl:"auth_url"`
	AuthToken      string                 `json:"auth_token" hcl:"auth_token"`
	Auth           *AuthConfig            `json:"auth" hcl:"auth"`
	Platforms      map[string]*Platform   `json:"platforms" hcl:"platforms"`
	Credentials    map[string]*Credential `json:"credentials" hcl:"credentials"`
	Contexts       map[string]*Context    `json:"contexts" hcl:"contexts"`
	CurrentContext string                 `json:"current_context" hcl:"current_context"`
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

// Binary returns the display name for the CLI
func (c *Config) Name() string {
	name := "Confluent CLI"
	if c.CLIName == "ccloud" {
		name = "Confluent Cloud CLI"
	}
	return name
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
		return nil, errors.ErrNoContext
	}
	return c.Contexts[c.CurrentContext], nil
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
		return errors.ErrNotLoggedIn
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
		c.Filename = fmt.Sprintf(defaultConfigFileFmt, c.CLIName)
	}
	filename, err := homedir.Expand(c.Filename)
	if err != nil {
		return "", err
	}
	return filename, nil
}
