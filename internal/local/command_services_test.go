package local

import (
	"os"
	"path/filepath"
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
		"confluent.telemetry.exporter._c3.metrics.include":                 ".*",
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
		"confluent.telemetry.exporter._c3.metrics.include":                 ".*",
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
