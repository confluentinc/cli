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

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	dynamicconfig "github.com/confluentinc/cli/internal/pkg/dynamic-config"
	"github.com/confluentinc/cli/internal/pkg/utils"
	"github.com/confluentinc/cli/internal/pkg/version"
	testserver "github.com/confluentinc/cli/test/test-server"
)

const BackwardCompatibilityLevel = "BACKWARD"

var details = &accountDetails{
	cluster: &schedv1.KafkaCluster{
		Id:          "lkc-asyncapi",
		Name:        "AsyncAPI Cluster",
		Endpoint:    "http://kafka-endpoint",
		ApiEndpoint: "http://kafka-endpoint",
		AccountId:   "env-asyncapi",
	},
	srClient: &srsdk.APIClient{
		DefaultApi: &srMock.DefaultApi{
			GetByUniqueAttributesFunc: func(_ context.Context, typeName string, qualifiedName string, localVarOptionals *srsdk.GetByUniqueAttributesOpts) (srsdk.AtlasEntityWithExtInfo, *http.Response, error) {
				if typeName == "kafka_topic" {
					return srsdk.AtlasEntityWithExtInfo{Entity: srsdk.AtlasEntity{Attributes: map[string]interface{}{"description": "kafka topic description"}}}, nil, nil
				}
				return srsdk.AtlasEntityWithExtInfo{}, nil, nil
			},
			ListFunc: func(_ context.Context, _ *srsdk.ListOpts) ([]string, *http.Response, error) {
				return []string{"subject 1", "subject 2"}, nil, nil
			},
			ListVersionsFunc: func(_ context.Context, _ string, _ *srsdk.ListVersionsOpts) ([]int32, *http.Response, error) {
				return []int32{1234, 4567}, nil, nil
			},
			GetSchemaByVersionFunc: func(_ context.Context, _ string, _ string, _ *srsdk.GetSchemaByVersionOpts) (srsdk.Schema, *http.Response, error) {
				return srsdk.Schema{
					Subject:    "subject1",
					Version:    1,
					Id:         1,
					SchemaType: "avro",
					Schema:     `{"doc":"Sample schema to help you get started.","fields":[{"doc":"The int type is a 32-bit signed integer.","name":"my_field1","type":"int"},{"doc":"The double type is a double precision(64-bit) IEEE754 floating-point number.","name":"my_field2","type":"double"},{"doc":"The string is a unicode character sequence.","name":"my_field3","type":"string"}],"name":"sampleRecord","namespace":"com.mycorp.mynamespace","type":"record"}`,
				}, nil, nil
			},
			GetSubjectLevelConfigFunc: func(_ context.Context, _ string, _ *srsdk.GetSubjectLevelConfigOpts) (srsdk.Config, *http.Response, error) {
				return srsdk.Config{CompatibilityLevel: BackwardCompatibilityLevel}, nil, nil
			},
			GetTopLevelConfigFunc: func(ctx context.Context) (srsdk.Config, *http.Response, error) {
				return srsdk.Config{CompatibilityLevel: BackwardCompatibilityLevel}, nil, nil
			},
			GetTagsFunc: func(_ context.Context, _, _ string) ([]srsdk.TagResponse, *http.Response, error) {
				return []srsdk.TagResponse{
					{
						TypeName: "Public",
					},
				}, nil, nil
			},
			GetTagDefByNameFunc: func(_ context.Context, _ string) (srsdk.TagDef, *http.Response, error) {
				return srsdk.TagDef{Name: "Public", Description: "Public tag"}, nil, nil
			},
		},
	},
}

func newCmd() (*command, error) {
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
							Bootstrap:    "kafka-endpoint",
							APIEndpoint:  "kafka-endpoint",
							RestEndpoint: "kafka-endpoint",
							APIKeys: map[string]*v1.APIKeyPair{
								"AsyncAPI": {Key: "ASYNCAPIKEY", Secret: "ASYNCAPISECRET"},
							},
							APIKey:     "AsyncAPI",
							LastUpdate: time.Now(),
						},
					},
				},
			},
		},
		CurrentContext: "asyncapi",
	}
	prerunner := &pcmd.PreRun{Config: cfg}
	cmd := new(cobra.Command)
	c := &command{AuthenticatedStateFlagCommand: pcmd.NewAuthenticatedStateFlagCommand(cmd, prerunner)}
	c.Command.Flags().String("resource", "lsrc-asyncapi", "resource flag for SR testing")
	c.Version = &version.Version{Version: "1", UserAgent: "asyncapi"}
	pcmd.AddApiKeyFlag(c.Command, c.AuthenticatedCLICommand)
	pcmd.AddApiSecretFlag(c.Command)
	err := c.Command.Flags().Set("api-key", "ASYNCAPIKEY")
	if err != nil {
		return nil, err
	}
	err = c.Command.Flags().Set("api-secret", "ASYNCAPISECRET")
	if err != nil {
		return nil, err
	}
	c.Command.Flags().String("sr-endpoint", "schema-registry-endpoint", "SR endpoint")
	c.State = cfg.Context().State
	c.Config = dynamicconfig.New(cfg, nil, nil)
	c.Config.CurrentContext = cfg.CurrentContext
	c.Context = c.Config.Context()
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
		SchemaRegistry: &ccsdkmock.SchemaRegistry{
			GetSchemaRegistryClusterFunc: func(ctx context.Context, clusterConfig *schedv1.SchemaRegistryCluster) (*schedv1.SchemaRegistryCluster, error) {
				return nil, nil
			},
		},
		Kafka: &ccsdkmock.Kafka{
			DescribeFunc: func(ctx context.Context, cluster *schedv1.KafkaCluster) (*schedv1.KafkaCluster, error) {
				return details.cluster, nil
			},
			ListFunc: func(ctx context.Context, cluster *schedv1.KafkaCluster) (clusters []*schedv1.KafkaCluster, e error) {
				return []*schedv1.KafkaCluster{details.cluster}, nil
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
							{Partition: 0,
								Leader: &schedv1.KafkaNode{Id: 1001},
								Replicas: []*schedv1.KafkaNode{
									{Id: 1001},
									{Id: 1002},
									{Id: 1003},
								},
								Isr: []*schedv1.KafkaNode{
									{Id: 1001},
									{Id: 1002},
									{Id: 1003},
								},
							},
							{Partition: 1,
								Leader: &schedv1.KafkaNode{Id: 1002},
								Replicas: []*schedv1.KafkaNode{
									{Id: 1001},
									{Id: 1002},
									{Id: 1003},
								},
								Isr: []*schedv1.KafkaNode{
									{Id: 1002},
									{Id: 1003},
								},
							},
							{Partition: 2,
								Leader: &schedv1.KafkaNode{Id: 1003},
								Replicas: []*schedv1.KafkaNode{
									{Id: 1001},
									{Id: 1002},
									{Id: 1003},
								},
								Isr: []*schedv1.KafkaNode{
									{Id: 1003},
								},
							},
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
	details.srCluster = c.Config.Context().SchemaRegistryClusters["lsrc-asyncapi"]
	return c, nil
}

func TestGetTopicDescription(t *testing.T) {
	c, err := newCmd()
	require.NoError(t, err)
	details.topics, _ = c.Client.Kafka.ListTopics(*new(context.Context), new(schedv1.KafkaCluster))
	details.channelDetails.currentSubject = "subject1"
	details.channelDetails.currentTopic = details.topics[0]
	err = details.getTopicDescription()
	require.NoError(t, err)
	require.Equal(t, "kafka topic description", details.channelDetails.currentTopicDescription)
}

func TestGetClusterDetails(t *testing.T) {
	c, err := newCmd()
	require.NoError(t, err)
	err = c.getClusterDetails(details)
	require.NoError(t, err)
}

func TestGetSchemaRegistry(t *testing.T) {
	c, err := newCmd()
	require.NoError(t, err)
	flags := &flags{apiKey: "ASYNCAPIKEY", apiSecret: "ASYNCAPISECRET"}
	err = c.getSchemaRegistry(details, flags)
	utils.Println(c.Command, "")
	require.Error(t, err)
}

func TestGetSchemaDetails(t *testing.T) {
	c, err := newCmd()
	require.NoError(t, err)
	details.topics, _ = c.Client.Kafka.ListTopics(*new(context.Context), new(schedv1.KafkaCluster))
	details.channelDetails.currentSubject = "subject1"
	details.channelDetails.currentTopic = details.topics[0]
	schema, _, _ := details.srClient.DefaultApi.GetSchemaByVersion(*new(context.Context), "subject1", "1", nil)
	details.channelDetails.schema = &schema
	err = details.getSchemaDetails()
	require.NoError(t, err)
}

func TestGetChannelDetails(t *testing.T) {
	c, err := newCmd()
	require.NoError(t, err)
	details.topics, _ = c.Client.Kafka.ListTopics(*new(context.Context), new(schedv1.KafkaCluster))
	details.channelDetails.currentSubject = "subject1"
	details.channelDetails.currentTopic = details.topics[0]
	schema, _, _ := details.srClient.DefaultApi.GetSchemaByVersion(*new(context.Context), "subject1", "1", nil)
	details.channelDetails.schema = &schema
	flags := &flags{apiKey: "ASYNCAPIKEY", apiSecret: "ASYNCAPISECRET"}
	err = c.getChannelDetails(details, flags)
	require.NoError(t, err)
}

func TestGetBindings(t *testing.T) {
	c, err := newCmd()
	require.NoError(t, err)
	topics, _ := c.Client.Kafka.ListTopics(*new(context.Context), new(schedv1.KafkaCluster))
	_, err = c.getBindings(details.cluster, topics[0])
	require.NoError(t, err)
}

func TestGetTags(t *testing.T) {
	c, err := newCmd()
	require.NoError(t, err)
	schema, _, _ := details.srClient.DefaultApi.GetSchemaByVersion(*new(context.Context), "subject1", "1", nil)
	details.srCluster = c.Config.Context().SchemaRegistryClusters["lsrc-asyncapi"]
	details.channelDetails.schema = &schema
	err = details.getTags()
	require.NoError(t, err)
}

func TestGetMessageCompatibility(t *testing.T) {
	_, err := getMessageCompatibility(details.srClient, *new(context.Context), "subject1")
	require.NoError(t, err)
}

func TestMsgName(t *testing.T) {
	require.Equal(t, "TopicNameMessage", msgName("topic name"))
}
