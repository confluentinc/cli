package local

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	climock "github.com/confluentinc/cli/v4/mock"
)

const (
	exampleDir  = "dir"
	exampleFile = "file"
)

func TestGetConnectConfig(t *testing.T) {
	want := map[string]string{
		"bootstrap.servers":            "localhost:9092",
		"plugin.path":                  exampleFile,
		"consumer.interceptor.classes": "io.confluent.monitoring.clients.interceptor.MonitoringConsumerInterceptor",
		"producer.interceptor.classes": "io.confluent.monitoring.clients.interceptor.MonitoringProducerInterceptor",
		"rest.extension.classes":       "io.confluent.connect.replicator.monitoring.ReplicatorMonitoringExtension",
	}
	testGetConfig(t, "connect", want)

	req := require.New(t)
	req.Equal(exampleFile, os.Getenv("CLASSPATH"))
}

func TestGetControlCenterConfig(t *testing.T) {
	want := map[string]string{
		"confluent.controlcenter.data.dir":                 exampleDir,
		"confluent.controlcenter.alertmanager.config.file": "dir/abc",
		"confluent.controlcenter.prometheus.rules.file":    "dir/def",
	}
	os.Setenv("CONTROL_CENTER_HOME", "dir")
	dir := os.Getenv("CONTROL_CENTER_HOME")

	path := filepath.Join(dir, "/etc/confluent-control-center/control-center-local.properties")
	err := os.MkdirAll(filepath.Dir(path), 0777)
	if err != nil {
		return
	}
	err = os.WriteFile(path, []byte("confluent.controlcenter.alertmanager.config.file=abc\n"+"confluent.controlcenter.prometheus.rules.file=def\n"), 0644)
	if err != nil {
		return
	}
	testGetConfigC3(t, "control-center", want)
}

func TestGetKafkaConfig(t *testing.T) {
	want := map[string]string{
		"log.dirs":         exampleDir,
		"metric.reporters": "io.confluent.metrics.reporter.ConfluentMetricsReporter",
		"confluent.metrics.reporter.bootstrap.servers": "localhost:9092",
		"confluent.metrics.reporter.topic.replicas":    "1",
	}
	testGetConfig(t, "kafka", want)
}

func TestGetKafkaConfigC3(t *testing.T) {
	want := map[string]string{
		"log.dirs": filepath.Join("dir", "kraft-broker-logs"),
		"confluent.metrics.reporter.bootstrap.servers":                     "localhost:9092",
		"confluent.metrics.reporter.topic.replicas":                        "1",
		"metric.reporters":                                                 "io.confluent.telemetry.reporter.TelemetryReporter",
		"confluent.telemetry.exporter._c3.type":                            "http",
		"confluent.telemetry.exporter._c3.enabled":                         "true",
		"confluent.telemetry.exporter._c3.metrics.include":                 c3TelemetryMetricsInclude,
		"confluent.telemetry.exporter._c3.client.base.url":                 "http://localhost:9090/api/v1/otlp",
		"confluent.telemetry.exporter._c3.client.compression":              "gzip",
		"confluent.telemetry.exporter._c3.api.key":                         "dummy",
		"confluent.telemetry.exporter._c3.api.secret":                      "dummy",
		"confluent.telemetry.exporter._c3.buffer.pending.batches.max":      "80",
		"confluent.telemetry.exporter._c3.buffer.batch.items.max":          "4000",
		"confluent.telemetry.exporter._c3.buffer.inflight.submissions.max": "10",
		"confluent.telemetry.metrics.collector.interval.ms":                "60000",
		"confluent.telemetry.remoteconfig._confluent.enabled":              "false",
		"confluent.consumer.lag.emitter.enabled":                           "true",
	}
	testGetConfigC3(t, "kafka", want)
}

func TestGetKafkaRestConfig(t *testing.T) {
	want := map[string]string{
		"schema.registry.url":          "http://localhost:8081",
		"zookeeper.connect":            "localhost:2181",
		"consumer.interceptor.classes": "io.confluent.monitoring.clients.interceptor.MonitoringConsumerInterceptor",
		"producer.interceptor.classes": "io.confluent.monitoring.clients.interceptor.MonitoringProducerInterceptor",
	}
	testGetConfig(t, "kafka-rest", want)
}

func TestGetKraftConfigC3(t *testing.T) {
	want := map[string]string{
		"log.dirs": filepath.Join("dir", "kraft-controller-logs"),
		"confluent.metrics.reporter.bootstrap.servers":                     "localhost:9092",
		"confluent.metrics.reporter.topic.replicas":                        "1",
		"metric.reporters":                                                 "io.confluent.telemetry.reporter.TelemetryReporter",
		"confluent.telemetry.exporter._c3.type":                            "http",
		"confluent.telemetry.exporter._c3.enabled":                         "true",
		"confluent.telemetry.exporter._c3.metrics.include":                 c3TelemetryMetricsInclude,
		"confluent.telemetry.exporter._c3.client.base.url":                 "http://localhost:9090/api/v1/otlp",
		"confluent.telemetry.exporter._c3.client.compression":              "gzip",
		"confluent.telemetry.exporter._c3.api.key":                         "dummy",
		"confluent.telemetry.exporter._c3.api.secret":                      "dummy",
		"confluent.telemetry.exporter._c3.buffer.pending.batches.max":      "80",
		"confluent.telemetry.exporter._c3.buffer.batch.items.max":          "4000",
		"confluent.telemetry.exporter._c3.buffer.inflight.submissions.max": "10",
		"confluent.telemetry.metrics.collector.interval.ms":                "60000",
		"confluent.telemetry.remoteconfig._confluent.enabled":              "false",
		"confluent.consumer.lag.emitter.enabled":                           "true",
	}
	testGetConfigC3(t, "kraft-controller", want)
}

// TestC3MetricsIncludeExcludesDelta guards against regressing to ".*", which exported
// delta-temporality metrics that the Control Center next-gen Prometheus OTLP receiver rejects
// (500 "invalid temporality and type combination"), dropping all broker metrics.
func TestC3MetricsIncludeExcludesDelta(t *testing.T) {
	req := require.New(t)

	req.NotEqual(".*", c3TelemetryMetricsInclude, "_c3.metrics.include must not be .* (sends delta metrics Prometheus rejects)")
	req.Contains(c3TelemetryMetricsInclude, "(?!.*delta)", "_c3.metrics.include must exclude delta-temporality metrics")

	// A few metrics Control Center relies on must stay in the allow-list (guards against truncation).
	for _, metric := range []string{
		"io.confluent.kafka.server.partition.under.replicated",
		"io.confluent.kafka.server.partition.under.min.isr",
		"io.confluent.kafka.server.controller.active.controller.count",
	} {
		req.Contains(c3TelemetryMetricsInclude, metric)
	}
}

// TestPre8NoTelemetryExporter protects the version gating: on Confluent Platform < 8.0 the broker
// and KRaft controller must NOT be configured with the Control Center next-gen telemetry exporter
// (they use the classic Kafka-topic metrics reporter instead), so the metrics.include allow-list is
// never emitted and those versions are unaffected by this change.
func TestPre8NoTelemetryExporter(t *testing.T) {
	req := require.New(t)

	for _, service := range []string{"kafka", "kraft-controller"} {
		config := getConfigForVersion(t, service, "7.9.0")

		for key := range config {
			req.NotContains(key, "confluent.telemetry.exporter._c3", "service %q on CP < 8.0 must not set any C3 telemetry exporter config", service)
		}
		req.Equal("io.confluent.metrics.reporter.ConfluentMetricsReporter", config["metric.reporters"], "service %q on CP < 8.0 should use the classic metrics reporter", service)
	}
}

// TestC3TelemetryExporterVersionGate documents the exact version boundary so the change is safe to
// roll out: the Control Center next-gen telemetry exporter (with the metrics.include allow-list) is
// emitted only on Confluent Platform 8.0 and later. On 7.x and earlier the broker and KRaft controller
// keep the classic Kafka-topic metrics reporter and never see this config, so upgrading the CLI cannot
// change metrics behavior for an existing pre-8.0 deployment.
func TestC3TelemetryExporterVersionGate(t *testing.T) {
	req := require.New(t)
	const includeKey = "confluent.telemetry.exporter._c3.metrics.include"

	cases := []struct {
		version  string
		expectC3 bool
	}{
		{"6.2.0", false},
		{"7.0.0", false},
		{"7.9.0", false},
		{"8.0.0", true},
		{"8.2.1", true},
		{"9.0.0", true},
	}

	for _, service := range []string{"kafka", "kraft-controller"} {
		for _, c := range cases {
			config := getConfigForVersion(t, service, c.version)
			if c.expectC3 {
				req.Equal(c3TelemetryMetricsInclude, config[includeKey], "service %q on %s should export the C3 allow-list", service, c.version)
				req.Equal("io.confluent.telemetry.reporter.TelemetryReporter", config["metric.reporters"], "service %q on %s should use the telemetry reporter", service, c.version)
			} else {
				req.NotContains(config, includeKey, "service %q on %s must not set the C3 metrics include", service, c.version)
				req.Equal("io.confluent.metrics.reporter.ConfluentMetricsReporter", config["metric.reporters"], "service %q on %s should keep the classic metrics reporter", service, c.version)
			}
		}
	}
}

// TestC3TelemetryConfigIdenticalForKafkaAndKraft confirms the broker and the KRaft controller receive
// the exact same telemetry configuration (including the metrics.include fix). Both export to the same
// Control Center Prometheus, so a divergence between the two would silently break one of them.
func TestC3TelemetryConfigIdenticalForKafkaAndKraft(t *testing.T) {
	req := require.New(t)

	kafka := getConfigForVersion(t, "kafka", "8.1.0")
	kraft := getConfigForVersion(t, "kraft-controller", "8.1.0")

	telemetry := func(config map[string]string) map[string]string {
		out := make(map[string]string)
		for key, val := range config {
			if strings.HasPrefix(key, "confluent.telemetry.") || key == "metric.reporters" {
				out[key] = val
			}
		}
		return out
	}

	req.Equal(telemetry(kafka), telemetry(kraft), "kafka and kraft-controller must get identical telemetry config")
	req.Equal(c3TelemetryMetricsInclude, kafka["confluent.telemetry.exporter._c3.metrics.include"])
}

func getConfigForVersion(t *testing.T, service, version string) map[string]string {
	t.Helper()
	req := require.New(t)

	c := &command{
		ch: &climock.MockConfluentHome{
			IsConfluentPlatformFunc: func() (bool, error) {
				return true, nil
			},
			GetConfluentVersionFunc: func() (string, error) {
				return version, nil
			},
			GetFileFunc: func(path ...string) (string, error) {
				return exampleFile, nil
			},
			FindFileFunc: func(pattern string) ([]string, error) {
				return []string{exampleFile}, nil
			},
			ReadServiceConfigFunc: func(service string, _ bool) ([]byte, error) {
				return []byte("plugin.path=share/java"), nil
			},
		},
		cc: &climock.MockConfluentCurrent{
			GetDataDirFunc: func(service string) (string, error) {
				return exampleDir, nil
			},
		},
	}

	got, err := c.getConfig(service)
	req.NoError(err)
	return got
}

func TestGetKsqlServerConfig(t *testing.T) {
	want := map[string]string{
		"kafkastore.connection.url":    "localhost:2181",
		"ksql.schema.registry.url":     "http://localhost:8081",
		"state.dir":                    exampleDir,
		"consumer.interceptor.classes": "io.confluent.monitoring.clients.interceptor.MonitoringConsumerInterceptor",
		"producer.interceptor.classes": "io.confluent.monitoring.clients.interceptor.MonitoringProducerInterceptor",
	}
	testGetConfig(t, "ksql-server", want)
}

func TestGetSchemaRegistryConfig(t *testing.T) {
	want := map[string]string{
		"kafkastore.connection.url":    "localhost:2181",
		"consumer.interceptor.classes": "io.confluent.monitoring.clients.interceptor.MonitoringConsumerInterceptor",
		"producer.interceptor.classes": "io.confluent.monitoring.clients.interceptor.MonitoringProducerInterceptor",
	}
	testGetConfig(t, "schema-registry", want)
}

func TestGetZookeeperConfig(t *testing.T) {
	want := map[string]string{
		"dataDir": exampleDir,
	}
	testGetConfig(t, "zookeeper", want)
}

func testGetConfig(t *testing.T, service string, want map[string]string) {
	req := require.New(t)

	c := &command{
		ch: &climock.MockConfluentHome{
			IsConfluentPlatformFunc: func() (bool, error) {
				return true, nil
			},
			GetConfluentVersionFunc: func() (string, error) {
				return "7.9.0", nil
			},
			GetFileFunc: func(path ...string) (string, error) {
				return exampleFile, nil
			},
			FindFileFunc: func(pattern string) ([]string, error) {
				return []string{exampleFile}, nil
			},
			ReadServiceConfigFunc: func(service string, _ bool) ([]byte, error) {
				return []byte("plugin.path=share/java"), nil
			},
		},
		cc: &climock.MockConfluentCurrent{
			GetDataDirFunc: func(service string) (string, error) {
				return exampleDir, nil
			},
		},
	}

	got, err := c.getConfig(service)

	req.NoError(err)
	req.Equal(want, got)
}

func testGetConfigC3(t *testing.T, service string, want map[string]string) {
	req := require.New(t)

	c := &command{
		ch: &climock.MockConfluentHome{
			IsConfluentPlatformFunc: func() (bool, error) {
				return true, nil
			},
			GetConfluentVersionFunc: func() (string, error) {
				return "8.1.0", nil
			},
			GetFileFunc: func(path ...string) (string, error) {
				return exampleFile, nil
			},
			FindFileFunc: func(pattern string) ([]string, error) {
				return []string{exampleFile}, nil
			},
			ReadServiceConfigFunc: func(service string, _ bool) ([]byte, error) {
				return []byte("plugin.path=share/java"), nil
			},
		},
		cc: &climock.MockConfluentCurrent{
			GetDataDirFunc: func(service string) (string, error) {
				return exampleDir, nil
			},
		},
	}

	got, err := c.getConfig(service)

	req.NoError(err)
	req.Equal(want, got)
}

func TestConfluentPlatformAvailableServices(t *testing.T) {
	req := require.New(t)

	c := &command{
		ch: &climock.MockConfluentHome{
			IsConfluentPlatformFunc: func() (bool, error) {
				return true, nil
			},
			GetConfluentVersionFunc: func() (string, error) {
				return "7.9.0", nil
			},
		},
	}

	got, err := c.getAvailableServices()
	req.NoError(err)

	want := []string{
		"zookeeper",
		"kafka",
		"schema-registry",
		"kafka-rest",
		"connect",
		"ksql-server",
		"control-center",
	}
	req.Equal(want, got)
}

func TestConfluentCommunitySoftwareAvailableServices(t *testing.T) {
	req := require.New(t)

	c := &command{
		ch: &climock.MockConfluentHome{
			IsConfluentPlatformFunc: func() (bool, error) {
				return false, nil
			},
			GetConfluentVersionFunc: func() (string, error) {
				return "7.9.0", nil
			},
		},
	}

	got, err := c.getAvailableServices()
	req.NoError(err)

	want := []string{
		"zookeeper",
		"kafka",
		"schema-registry",
		"kafka-rest",
		"connect",
		"ksql-server",
	}
	req.Equal(want, got)
}
