package shared

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"

	"github.com/confluentinc/cli/log"
)

const (
	defaultConfigFile = "~/.confluent/config.json"
)

// ErrNoConfig means that no configuration exists.
var ErrNoConfig = fmt.Errorf("no config file exists")

// Config represents the CLI configuration.
type Config struct {
	MetricSink     MetricSink             `json:"-" hcl:"-"`
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

// NewConfig initializes a new Config object
func NewConfig(config ...*Config) *Config {
	var c *Config
	if config == nil {
		c = &Config{}
	} else {
		c = config[0]
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
			return ErrNoConfig
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
	err = os.MkdirAll(path.Dir(filename), 0700)
	if err != nil {
		return errors.Wrapf(err, "unable to create config directory: %s", filename)
	}
	err = ioutil.WriteFile(filename, cfg, 0600)
	if err != nil {
		return errors.Wrapf(err, "unable to write config to file: %s", filename)
	}
	return nil
}

// Context returns the current Context object.
func (c *Config) Context() (*Context, error) {
	if c.CurrentContext == "" {
		return nil, ErrNoContext
	}
	return c.Contexts[c.CurrentContext], nil
}

// KafkaClusterConfig returns the current KafkaClusterConfig
func (c *Config) KafkaClusterConfig() (KafkaClusterConfig, error) {
	cfg, err := c.Context()
	if err != nil {
		return KafkaClusterConfig{}, err
	}
	cluster, found := c.Platforms[cfg.Platform].KafkaClusters[cfg.Kafka]
	if !found {
		e := fmt.Errorf("no auth found for Kafka %s, please run `confluent kafka cluster auth` first", cfg.Kafka)
		return KafkaClusterConfig{}, NotAuthenticatedError(e)
	}
	return cluster, nil
}

// CheckLogin returns an error if the user is not logged in.
func (c *Config) CheckLogin() error {
	if c.Auth == nil || c.Auth.Account == nil || c.Auth.Account.Id == "" {
		return ErrUnauthorized
	}
	return nil
}

func (c *Config) getFilename() (string, error) {
	if c.Filename == "" {
		c.Filename = defaultConfigFile
	}
	filename, err := homedir.Expand(c.Filename)
	if err != nil {
		return "", err
	}
	return filename, nil
}
