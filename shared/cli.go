package shared

import (
	metrics "github.com/armon/go-metrics"
	plugin "github.com/hashicorp/go-plugin"

	orgv1 "github.com/confluentinc/ccloudapis/org/v1"
)

// AuthConfig represents an authenticated user.
type AuthConfig struct {
	User     *orgv1.User      `json:"user" hcl:"user"`
	Account  *orgv1.Account   `json:"account" hcl:"account"`
	Accounts []*orgv1.Account `json:"accounts" hcl:"accounts"`
}

// KafkaClusterConfig represents a connection to a Kafka cluster.
type KafkaClusterConfig struct {
	Bootstrap   string `json:"bootstrap_servers" hcl:"bootstrap_servers"`
	APIEndpoint string `json:"api_endpoint,omitempty" hcl:"api_endpoint"`
	APIKey      string `json:"api_key" hcl:"api_key"`
	APISecret   string `json:"api_secret" hcl:"api_secret"`
}

// Platform represents a Confluent Platform deployment
type Platform struct {
	Server string `json:"server" hcl:"server"`
	// KafkaClusters store connection info for interacting directly with Kafka (e.g., topic mgmt, consume/produce, etc)
	// N.B. These may later be exposed in the CLI to directly register kafkas (outside a Control Plane)
	KafkaClusters map[string]KafkaClusterConfig `json:"kafka_clusters" hcl:"kafka_clusters"`
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
	Kafka      string `json:"kafka_cluster" hcl:"kafka_cluster"`
}

// Label represents a key-value pair for a metric.
type Label = metrics.Label

// The MetricSink interface is used to transmit metrics information to an external system.
type MetricSink interface {
	// A Gauge should retain the last value it is set to
	SetGauge(key []string, val float32)
	SetGaugeWithLabels(key []string, val float32, labels []Label)

	// Should emit a Key/Value pair for each call
	EmitKey(key []string, val float32)

	// Counters should accumulate values
	IncrCounter(key []string, val float32)
	IncrCounterWithLabels(key []string, val float32, labels []Label)

	// Samples are for timing information, where quantiles are used
	AddSample(key []string, val float32)
	AddSampleWithLabels(key []string, val float32, labels []Label)
}

// Handshake is a configuration for CLI to communicate with SDK components.
var Handshake = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "CLI_PLUGIN",
	MagicCookieValue: "hello",
}

// PluginMap is the map of plugins we can dispense.
var PluginMap = map[string]plugin.Plugin{}
