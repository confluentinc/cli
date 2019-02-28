package shared

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"

	authv1 "github.com/confluentinc/ccloudapis/auth/v1"
	kafka1 "github.com/confluentinc/ccloudapis/kafka/v1"
	"github.com/confluentinc/cli/log"
)

const (
	defaultConfigFile = "~/.ccloud/config.json"
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
		e := fmt.Errorf("no auth found for Kafka %s, please run `ccloud kafka cluster auth` first", cfg.Kafka)
		return KafkaClusterConfig{}, NotAuthenticatedError(e)
	}
	return *cluster, nil
}

// SetContextKey stores apiKey within the specified cluster context
func (c *Config) SetKafkaClusterKey(clusterId string, apiKey *authv1.ApiKey) error {
	kafkaConf, ok := c.Platforms[c.CurrentContext].KafkaClusters[clusterId]
	if !ok {
		return ErrNotFound
	}

	// There is no clean way to update cluster specifics (api endpoint and bootstrap) from here
	// Instead we will use `describe cluster` to keep these fields updated
	kafkaConf.APIKey = apiKey.GetKey()
	kafkaConf.APISecret = apiKey.GetSecret()

	c.Platforms[c.CurrentContext].KafkaClusters[clusterId] = kafkaConf

	return c.Save()
}

// AddCluster adds a cluster to the current platform context
func (c *Config) AddCluster(cluster *kafka1.KafkaCluster) error {
	if _, ok := c.Platforms[c.CurrentContext].KafkaClusters[cluster.Id]; !ok {
		c.Platforms[c.CurrentContext].KafkaClusters[cluster.Id] = new(KafkaClusterConfig)
	}

	return c.UpdateCluster(cluster)
}

// UpdateCluster updates a cluster configuration within the current platform context
func (c *Config) UpdateCluster(cluster *kafka1.KafkaCluster) error {
	clusterConf, ok := c.Platforms[c.CurrentContext].KafkaClusters[cluster.Id]
	if !ok {
		return ErrNotFound
	}

	clusterConf.APIEndpoint = cluster.ApiEndpoint
	clusterConf.Bootstrap = addBootstrapProtocol(cluster.Endpoint)

	c.Platforms[c.CurrentContext].KafkaClusters[cluster.Id] = clusterConf

	return c.Save()
}

// UpdateClusters updates the local cluster configurations for the current platform context
func (c *Config) UpdateClusters(clusters []*kafka1.KafkaCluster) error {
	platformClusters := c.Platforms[c.CurrentContext].KafkaClusters

	if platformClusters == nil {
		platformClusters = map[string]*KafkaClusterConfig{}
	}

	// index clusters which are still valid
	valid := map[string]struct{}{}

	for _, cluster := range clusters {
		if _, ok := platformClusters[cluster.Id]; !ok {
			platformClusters[cluster.Id] = &KafkaClusterConfig{
				Bootstrap:   addBootstrapProtocol(cluster.GetEndpoint()),
				APIEndpoint: cluster.GetApiEndpoint(),
			}

		}
		valid[cluster.Id] = struct{}{}
	}

	//Remove clusters which are not longer valid
	for id := range platformClusters {
		if _, ok := valid[id]; !ok {
			delete(platformClusters, id)
		}
	}

	c.Platforms[c.CurrentContext].KafkaClusters = platformClusters

	return c.Save()
}

// MaybeDeleteCluster removes a cluster from the current platform context
func (c *Config) MaybeDeleteCluster(cluster *kafka1.KafkaCluster) error {
	platformClusters := c.Platforms[c.CurrentContext].KafkaClusters

	if _, ok := platformClusters[cluster.Id]; ok {
		delete(platformClusters, cluster.Id)
	}
	c.Platforms[c.CurrentContext].KafkaClusters = platformClusters

	return c.Save()
}

// UpdateClusterAPIKey updates the clusters api key within the current platform context
func (c *Config) UpdateClusterAPIKey(clusterId string, apiKey *authv1.ApiKey) error {
	return c.SetKafkaClusterKey(clusterId, apiKey)
}

// MaybeDeleteKey removes an ApiKey from the current platform context
func (c *Config) MaybeDeleteKey(apikey string) {
	for platformKey, platform := range c.Platforms {
		for clusterKey, cluster := range platform.KafkaClusters {
			if cluster.APIKey == apikey {
				cluster.APIKey = ""
				cluster.APISecret = ""
				c.Platforms[platformKey].KafkaClusters[clusterKey] = cluster
			}
		}
	}
	_ = c.Save()
	return
}

func addBootstrapProtocol(endpoint string) string {
	return strings.TrimPrefix(endpoint, "SASL_SSL://")
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
