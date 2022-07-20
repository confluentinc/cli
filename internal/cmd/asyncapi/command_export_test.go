package asyncapi

import (
	"context"
	"net/http"
	"testing"
	"time"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/ccloud-sdk-go-v1"
	ccsdkmock "github.com/confluentinc/ccloud-sdk-go-v1/mock"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	srMock "github.com/confluentinc/schema-registry-sdk-go/mock"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	pasyncapi "github.com/confluentinc/cli/internal/pkg/asyncapi"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	dynamicconfig "github.com/confluentinc/cli/internal/pkg/dynamic-config"
	"github.com/confluentinc/cli/internal/pkg/version"
	testserver "github.com/confluentinc/cli/test/test-server"
)

var kafkaCluster = &schedv1.KafkaCluster{
	Id:        "lkc-asyncapi",
	Name:      "AsyncAPI Cluster",
	Endpoint:  "http://kafka-endpoint",
	AccountId: "env-asyncapi",
}

const BACKWARD = "BACKWARD"

var srClient = &srsdk.APIClient{
	DefaultApi: &srMock.DefaultApi{
		ListFunc: func(ctx context.Context, opts *srsdk.ListOpts) ([]string, *http.Response, error) {
			return []string{"subject 1", "subject 2"}, nil, nil
		},
		ListVersionsFunc: func(ctx context.Context, subject string, opts *srsdk.ListVersionsOpts) (int32s []int32, response *http.Response, e error) {
			return []int32{1234, 4567}, nil, nil
		},
		GetSchemaByVersionFunc: func(ctx context.Context, subject string, version string, opts *srsdk.GetSchemaByVersionOpts) (srsdk.Schema, *http.Response, error) {
			return srsdk.Schema{
				Subject:    "subject1",
				Version:    1,
				Id:         1,
				SchemaType: "avro",
				Schema:     `{"doc":"Sample schema to help you get started.","fields":[{"doc":"The int type is a 32-bit signed integer.","name":"my_field1","type":"int"},{"doc":"The double type is a double precision(64-bit) IEEE754 floating-point number.","name":"my_field2","type":"double"},{"doc":"The string is a unicode character sequence.","name":"my_field3","type":"string"}],"name":"sampleRecord","namespace":"com.mycorp.mynamespace","type":"record"}`,
			}, nil, nil
		},
		GetSubjectLevelConfigFunc: func(ctx context.Context, subject string, localVarOptionals *srsdk.GetSubjectLevelConfigOpts) (srsdk.Config, *http.Response, error) {
			return srsdk.Config{CompatibilityLevel: BACKWARD}, nil, nil
		},
		GetTopLevelConfigFunc: func(ctx context.Context) (srsdk.Config, *http.Response, error) {
			return srsdk.Config{CompatibilityLevel: BACKWARD}, nil, nil
		},
	},
}

func newCmd() *command {
	cfg := &v1.Config{
		BaseConfig: &config.BaseConfig{},
		Contexts: map[string]*v1.Context{
			"asyncapi": {
				PlatformName: "confluent.cloud",
				Credential:   &v1.Credential{CredentialType: v1.Username},
				State: &v1.ContextState{
					Auth: &v1.AuthConfig{
						Organization: testserver.RegularOrg,
						Account:      &orgv1.Account{Id: "env-asyncapi", Name: "asyncapi"},
					},
					AuthToken: "env-asyncapi",
				},
				SchemaRegistryClusters: map[string]*v1.SchemaRegistryCluster{
					"lsrc-asyncapi": {
						Id:                     "lsrc-asyncapi",
						SchemaRegistryEndpoint: "schema-registry-endpoint",
						SrCredentials:          &v1.APIKeyPair{Key: "ASYNCAPIKEY", Secret: "ASYNCAPISECRET"},
					},
				},
				KafkaClusterContext: &v1.KafkaClusterContext{
					EnvContext:         false,
					ActiveKafkaCluster: "lkc-asyncapi",
					KafkaClusterConfigs: map[string]*v1.KafkaClusterConfig{
						"lkc-asyncapi": {
							ID:           "lkc-asyncapi",
							Name:         "AsyncAPI Cluster",
							Bootstrap:    "",
							APIEndpoint:  "kafka-endpoint",
							RestEndpoint: "kafka-endpoint",
							APIKeys: map[string]*v1.APIKeyPair{
								"AsyncAPI": {Key: "ASYNCAPIKEY", Secret: "ASYNCAPISECRET"},
							},
							APIKey:     "AsyncAPI",
							LastUpdate: time.Now(),
						},
					},
					KafkaEnvContexts: nil,
					Context:          nil,
				},
			},
		},
		CurrentContext: "asyncapi",
	}
	prerunner := &pcmd.PreRun{Config: cfg}
	cmd := cobra.Command{}
	c := &command{AuthenticatedStateFlagCommand: pcmd.NewAuthenticatedStateFlagCommand(&cmd, prerunner)}
	c.State = cfg.Context().State
	c.Config = dynamicconfig.New(cfg, nil, nil)
	c.Config.CurrentContext = cfg.CurrentContext
	c.Client = &ccloud.Client{
		Account: &ccsdkmock.Account{
			CreateFunc: func(context.Context, *orgv1.Account) (*orgv1.Account, error) {
				return nil, nil
			},
			GetFunc: func(context.Context, *orgv1.Account) (*orgv1.Account, error) {
				return nil, nil
			},
			ListFunc: func(context.Context, *orgv1.Account) ([]*orgv1.Account, error) {
				return nil, nil
			},
		},
		Kafka: &ccsdkmock.Kafka{
			DescribeFunc: func(ctx context.Context, cluster *schedv1.KafkaCluster) (*schedv1.KafkaCluster, error) {
				return kafkaCluster, nil
			},
			ListFunc: func(ctx context.Context, cluster *schedv1.KafkaCluster) (clusters []*schedv1.KafkaCluster, e error) {
				return []*schedv1.KafkaCluster{kafkaCluster}, nil
			},
			ListTopicsFunc: func(ctx context.Context, cluster *schedv1.KafkaCluster) ([]*schedv1.TopicDescription, error) {
				return []*schedv1.TopicDescription{
					{
						Name: "topic1",
						Config: []*schedv1.TopicConfigEntry{
							{
								Name:  "cleanup.policy",
								Value: "delete",
							},
							{
								Name:  "delete.retention.ms",
								Value: "86400000",
							},
						},
						Partitions: []*schedv1.TopicPartitionInfo{
							{Partition: 0, Leader: &schedv1.KafkaNode{Id: 1001}, Replicas: []*schedv1.KafkaNode{{Id: 1001}, {Id: 1002}, {Id: 1003}}, Isr: []*schedv1.KafkaNode{{Id: 1001}, {Id: 1002}, {Id: 1003}}},
							{Partition: 1, Leader: &schedv1.KafkaNode{Id: 1002}, Replicas: []*schedv1.KafkaNode{{Id: 1001}, {Id: 1002}, {Id: 1003}}, Isr: []*schedv1.KafkaNode{{Id: 1002}, {Id: 1003}}},
							{Partition: 2, Leader: &schedv1.KafkaNode{Id: 1003}, Replicas: []*schedv1.KafkaNode{{Id: 1001}, {Id: 1002}, {Id: 1003}}, Isr: []*schedv1.KafkaNode{{Id: 1003}}},
						},
					},
				}, nil
			},
			ListTopicConfigFunc: func(ctx context.Context, cluster *schedv1.KafkaCluster, topic *schedv1.Topic) (*schedv1.TopicConfig, error) {
				return &schedv1.TopicConfig{Entries: []*schedv1.TopicConfigEntry{
					{
						Name:  "cleanup.policy",
						Value: "delete",
					},
					{
						Name:  "delete.retention.ms",
						Value: "86400000",
					},
				}}, nil
			},
		}}
	return c
}

func TestGetEnv(t *testing.T) {
	require.Equal(t, "dev", getEnv("pkc-0wg55.us-central1.gcp.devel.cpdev.cloud:9092"))
	require.Equal(t, "local", getEnv("localhost:8081"))
}

func TestGetClusterDetails(t *testing.T) {
	c := newCmd()

	_, _, _, err := c.getClusterDetails("", "")
	if err != nil {
		require.NoError(t, err)
	}
}

func TestGetBroker(t *testing.T) {
	require.Equal(t, "kafka-endpoint", getBroker(kafkaCluster))
}

func TestGetSchemaRegistry(t *testing.T) {
	c := newCmd()
	c.Command.Flags().String("resource", "lsrc-asyncapi", "resource flag for SR testing")
	c.Version = &version.Version{Version: "1", UserAgent: "asyncapi"}
	pcmd.AddApiKeyFlag(c.Command, c.AuthenticatedCLICommand)
	pcmd.AddApiSecretFlag(c.Command)
	err := c.Command.Flags().Set("api-key", "ASYNCAPIKEY")
	require.NoError(t, err)
	err = c.Command.Flags().Set("api-secret", "ASYNCAPISECRET")
	require.NoError(t, err)
	c.Command.Flags().String("sr-endpoint", "schema-registry-endpoint", "SR endpoint")
	c.Client.SchemaRegistry = &ccsdkmock.SchemaRegistry{GetSchemaRegistryClusterFunc: func(ctx context.Context, clusterConfig *schedv1.SchemaRegistryCluster) (*schedv1.SchemaRegistryCluster, error) {
		return nil, nil
	}}
	_, _, _, err = getSchemaRegistry(c, c.Command, "", "")
	require.EqualError(t, err, "EOF")
}

func TestGetChannelDetails(t *testing.T) {
	c := newCmd()
	topics, _ := c.Client.Kafka.ListTopics(*new(context.Context), new(schedv1.KafkaCluster))

	contentType, _, _, err := getChannelDetails(topics[0], srClient, *new(context.Context), "subject1")
	require.NoError(t, err)
	require.Equal(t, "avro", contentType)
}

func TestGetBindings(t *testing.T) {
	c := newCmd()
	topics, _ := c.Client.Kafka.ListTopics(*new(context.Context), new(schedv1.KafkaCluster))
	_, err := c.getBindings(kafkaCluster, topics[0], "group1")
	require.NoError(t, err)
}

func TestGetTags(t *testing.T) {
	c := newCmd()
	schema, _, _ := srClient.DefaultApi.GetSchemaByVersion(*new(context.Context), "subject1", "1", nil)
	catalog := pasyncapi.Catalog{
		GetSchemaLevelTagsRequest: func(srEndpoint, schemaClusterId, schemaId, apiKey, apiSecret string) ([]byte, error) {
			return []byte(`[{"typeName":"trial","entityType":"sr_schema","entityName":"lsrc-asyncapi:.:100001"}]`), nil
		},
		GetTagDefinitionsRequest: func(srEndpoint, tagName, apiKey, apiSecret string) ([]byte, error) {
			return []byte(`{"name":"trial","description":"Tag trial"}`), nil
		},
	}
	_, err := getTags(c.Config.Context().SchemaRegistryClusters["lsrc-asyncapi"], schema, "ASYNCAPIKEY", "ASYNCAPISECRET", catalog)
	require.NoError(t, err)
}

func TestAddMessageCompatibility(t *testing.T) {
	_, err := addMessageCompatibility(srClient, *new(context.Context), "subject1")
	require.NoError(t, err)
}
