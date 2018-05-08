package shared

import (
	metrics "github.com/armon/go-metrics"
	plugin "github.com/hashicorp/go-plugin"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
)

type AuthConfig struct {
	User      *orgv1.User    `json:"user" hcl:"user"`
	Account   *orgv1.Account `json:"account" hcl:"account"`
}

type Label = metrics.Label

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

var Handshake = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "CLI_PLUGIN",
	MagicCookieValue: "hello",
}

// PluginMap is the map of plugins we can dispense.
var PluginMap = map[string]plugin.Plugin{}
