package asyncapi

import (
	"context"
	"fmt"
	"net/http"
	"time"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"
	ccloudv1mock "github.com/confluentinc/ccloud-sdk-go-v1-public/mock"
	kafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"
	kafkarestv3mock "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3/mock"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	srMock "github.com/confluentinc/schema-registry-sdk-go/mock"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
	"github.com/confluentinc/cli/internal/pkg/ccstructs"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	dynamicconfig "github.com/confluentinc/cli/internal/pkg/dynamic-config"
	"github.com/confluentinc/cli/internal/pkg/version"
	testserver "github.com/confluentinc/cli/test/test-server"
)

const BackwardCompatibilityLevel = "BACKWARD"

var details = &accountDetails{
	cluster: &ccstructs.KafkaCluster{
		Id:        "lkc-asyncapi",
		Name:      "AsyncAPI Cluster",
		Endpoint:  "http://kafka-endpoint",
		AccountId: "env-asyncapi",
	},
	srClient: &srsdk.APIClient{
		DefaultApi: &srMock.DefaultApi{
			RegisterFunc: func(_ context.Context, subject string, _ srsdk.RegisterSchemaRequest) (srsdk.RegisterSchemaResponse, *http.Response, error) {
				if subject == "testTopic-value" {
					return srsdk.RegisterSchemaResponse{Id: 100001}, nil, nil
				}
				return srsdk.RegisterSchemaResponse{}, nil, nil
			},
			UpdateSubjectLevelConfigFunc: func(ctx context.Context, subject string, body srsdk.ConfigUpdateRequest) (srsdk.ConfigUpdateRequest, *http.Response, error) {
				if body.Compatibility == "BACKWARD" {
					return srsdk.ConfigUpdateRequest{Compatibility: body.Compatibility}, nil, nil
				}
				return srsdk.ConfigUpdateRequest{}, nil, fmt.Errorf("invalid compatibility type")
			},
			CreateTagDefsFunc: func(ctx context.Context, localVarOptionals *srsdk.CreateTagDefsOpts) ([]srsdk.TagDefResponse, *http.Response, error) {
				return []srsdk.TagDefResponse{
					{Name: "Tag1"}, {Name: "Tag2"},
				}, nil, nil
			},
			CreateTagsFunc: func(ctx context.Context, localVarOptionals *srsdk.CreateTagsOpts) ([]srsdk.TagResponse, *http.Response, error) {
				return []srsdk.TagResponse{{TypeName: "Tag1"}, {TypeName: "Tag2"}}, nil, nil
			},
			GetByUniqueAttributesFunc: func(_ context.Context, typeName string, qualifiedName string, localVarOptionals *srsdk.GetByUniqueAttributesOpts) (srsdk.AtlasEntityWithExtInfo, *http.Response, error) {
				if typeName == "kafka_topic" {
					return srsdk.AtlasEntityWithExtInfo{Entity: srsdk.AtlasEntity{Attributes: map[string]any{"description": "kafka topic description"}}}, nil, nil
				}
				return srsdk.AtlasEntityWithExtInfo{}, nil, nil
			},
			ListFunc: func(_ context.Context, _ *srsdk.ListOpts) ([]string, *http.Response, error) {
				return []string{"subject 1", "subject 2"}, nil, nil
			},
			ListVersionsFunc: func(_ context.Context, _ string, _ *srsdk.ListVersionsOpts) ([]int32, *http.Response, error) {
				return []int32{1234, 4567}, nil, nil
			},
			GetSchemaByVersionFunc: func(_ context.Context, subject string, _ string, _ *srsdk.GetSchemaByVersionOpts) (srsdk.Schema, *http.Response, error) {
				if subject == "subject2" {
					return srsdk.Schema{
						Subject:    "subject1",
						Version:    1,
						Id:         1,
						SchemaType: "PROTOBUF",
						Schema:     `syntax = "proto3"; package com.mycorp.mynamespace; message SampleRecord { int32 my_field1 = 1; double my_field2 = 2; string my_field3 = 3;}`,
					}, nil, nil
				}
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
				return []srsdk.TagResponse{{TypeName: "Public"}}, nil, nil
			},
			GetTagDefByNameFunc: func(_ context.Context, _ string) (srsdk.TagDef, *http.Response, error) {
				tag := srsdk.TagDef{
					Name:        "Public",
					Description: "Public tag",
				}
				return tag, nil, nil
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
						Account:      &ccloudv1.Account{Id: "env-asyncapi", Name: "asyncapi"},
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
	c.Command.Flags().String("schema-registry-api-key", "ASYNCAPIKEY", "API Key for schema registry")
	c.Command.Flags().String("schema-registry-api-secret", "ASYNCAPISECRET", "API Secret for Schema Registry")
	err := c.Command.Flags().Set("schema-registry-api-key", "ASYNCAPIKEY")
	if err != nil {
		return nil, err
	}
	err = c.Command.Flags().Set("schema-registry-api-secret", "ASYNCAPISECRET")
	if err != nil {
		return nil, err
	}
	c.Command.Flags().String("sr-endpoint", "schema-registry-endpoint", "SR endpoint")
	c.State = cfg.Context().State
	c.Config = dynamicconfig.New(cfg, nil, nil)
	c.Config.CurrentContext = cfg.CurrentContext
	c.Context = c.Config.Context()
	apiClient := kafkarestv3.NewAPIClient(kafkarestv3.NewConfiguration())
	apiClient.ConfigsV3Api = &kafkarestv3mock.ConfigsV3Api{
		ListKafkaTopicConfigsFunc: func(_ context.Context, _, _ string) kafkarestv3.ApiListKafkaTopicConfigsRequest {
			return kafkarestv3.ApiListKafkaTopicConfigsRequest{}
		},
		ListKafkaTopicConfigsExecuteFunc: func(_ kafkarestv3.ApiListKafkaTopicConfigsRequest) (kafkarestv3.TopicConfigDataList, *http.Response, error) {
			configs := []kafkarestv3.TopicConfigData{
				{
					Name:  "cleanup.policy",
					Value: *kafkarestv3.NewNullableString(kafkarestv3.PtrString("delete")),
				},
				{
					Name:  "delete.retention.ms",
					Value: *kafkarestv3.NewNullableString(kafkarestv3.PtrString("86400000")),
				},
			}
			return kafkarestv3.TopicConfigDataList{Data: configs}, nil, nil
		},
	}
	apiClient.TopicV3Api = &kafkarestv3mock.TopicV3Api{
		ListKafkaTopicsFunc: func(_ context.Context, _ string) kafkarestv3.ApiListKafkaTopicsRequest {
			return kafkarestv3.ApiListKafkaTopicsRequest{}
		},
		ListKafkaTopicsExecuteFunc: func(_ kafkarestv3.ApiListKafkaTopicsRequest) (kafkarestv3.TopicDataList, *http.Response, error) {
			return kafkarestv3.TopicDataList{Data: []kafkarestv3.TopicData{{TopicName: "topic1"}}}, nil, nil
		},
		CreateKafkaTopicFunc: func(_ context.Context, _ string) kafkarestv3.ApiCreateKafkaTopicRequest {
			return kafkarestv3.ApiCreateKafkaTopicRequest{}
		},
		CreateKafkaTopicExecuteFunc: func(request kafkarestv3.ApiCreateKafkaTopicRequest) (kafkarestv3.TopicData, *http.Response, error) {
			return kafkarestv3.TopicData{}, nil, nil
		},
	}
	kafkaRestProvider := pcmd.KafkaRESTProvider(func() (*pcmd.KafkaREST, error) {
		return &pcmd.KafkaREST{CloudClient: &ccloudv2.KafkaRestClient{APIClient: apiClient}}, nil
	})
	c.KafkaRESTProvider = &kafkaRestProvider
	c.Client = &ccloudv1.Client{
		SchemaRegistry: &ccloudv1mock.SchemaRegistry{
			GetSchemaRegistryClusterFunc: func(_ context.Context, _ *ccloudv1.SchemaRegistryCluster) (*ccloudv1.SchemaRegistryCluster, error) {
				return nil, nil
			},
		},
		Account: &ccloudv1mock.AccountInterface{
			CreateFunc: func(context.Context, *ccloudv1.Account) (*ccloudv1.Account, error) {
				return nil, nil
			},
			GetFunc: func(context.Context, *ccloudv1.Account) (*ccloudv1.Account, error) {
				return nil, nil
			},
			ListFunc: func(context.Context, *ccloudv1.Account) ([]*ccloudv1.Account, error) {
				return nil, nil
			},
		},
	}
	details.srCluster = c.Config.Context().SchemaRegistryClusters["lsrc-asyncapi"]
	details.kafkaRest, _ = c.GetKafkaREST()
	return c, err
}
