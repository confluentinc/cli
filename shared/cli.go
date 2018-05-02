package shared

import (
	"fmt"

	metrics "github.com/armon/go-metrics"
	plugin "github.com/hashicorp/go-plugin"
)

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

var ErrNotImplemented = fmt.Errorf("not implemented")
