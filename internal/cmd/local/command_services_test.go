package local

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/mock"
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
		"confluent.controlcenter.data.dir": exampleDir,
	}
	testGetConfig(t, "control-center", want)
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

func TestGetKafkaRestConfig(t *testing.T) {
	want := map[string]string{
		"schema.registry.url":          "http://localhost:8081",
		"zookeeper.connect":            "localhost:2181",
		"consumer.interceptor.classes": "io.confluent.monitoring.clients.interceptor.MonitoringConsumerInterceptor",
		"producer.interceptor.classes": "io.confluent.monitoring.clients.interceptor.MonitoringProducerInterceptor",
	}
	testGetConfig(t, "kafka-rest", want)
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

	c := &LocalCommand{
		ch: &mock.MockConfluentHome{
			IsConfluentPlatformFunc: func() (bool, error) {
				return true, nil
			},
			GetFileFunc: func(path ...string) (string, error) {
				return exampleFile, nil
			},
			FindFileFunc: func(pattern string) ([]string, error) {
				return []string{exampleFile}, nil
			},
		},
		cc: &mock.MockConfluentCurrent{
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

	c := &LocalCommand{
		ch: &mock.MockConfluentHome{
			IsConfluentPlatformFunc: func() (bool, error) {
				return true, nil
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

	c := &LocalCommand{
		ch: &mock.MockConfluentHome{
			IsConfluentPlatformFunc: func() (bool, error) {
				return false, nil
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
