package asyncapi

import (
	"context"
	"fmt"
	"net/http"

	v3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"
	kafkarestv3mock "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3/mock"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	srMock "github.com/confluentinc/schema-registry-sdk-go/mock"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	dynamicconfig "github.com/confluentinc/cli/internal/pkg/dynamic-config"
)

const BackwardCompatibilityLevel = "BACKWARD"

var detailsMock = &accountDetails{
	srCluster: &v1.SchemaRegistryCluster{
		Id:                     "lsrc-asyncapi",
		SchemaRegistryEndpoint: "schema-registry-endpoint",
		SrCredentials:          &v1.APIKeyPair{Key: "ASYNCAPIKEY", Secret: "ASYNCAPISECRET"},
	},
	clusterId: "lkc-asyncapi",
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

func mockAsyncApiCommand() *command {
	cmd := new(cobra.Command)
	cfg := v1.AuthenticatedCloudConfigMock()
	prerunner := &pcmd.PreRun{Config: cfg}
	c := &command{pcmd.NewAuthenticatedStateFlagCommand(cmd, prerunner)}
	c.Config = dynamicconfig.New(cfg, nil, nil)
	c.Context = c.Config.Context()
	apiClient := v3.NewAPIClient(v3.NewConfiguration())
	apiClient.ConfigsV3Api = &kafkarestv3mock.ConfigsV3Api{
		ListKafkaTopicConfigsFunc: func(_ context.Context, _, _ string) v3.ApiListKafkaTopicConfigsRequest {
			return v3.ApiListKafkaTopicConfigsRequest{}
		},
		ListKafkaTopicConfigsExecuteFunc: func(_ v3.ApiListKafkaTopicConfigsRequest) (v3.TopicConfigDataList, *http.Response, error) {
			return v3.TopicConfigDataList{}, nil, nil
		},
	}
	apiClient.TopicV3Api = &kafkarestv3mock.TopicV3Api{
		ListKafkaTopicsFunc: func(_ context.Context, _ string) v3.ApiListKafkaTopicsRequest {
			return v3.ApiListKafkaTopicsRequest{}
		},
		ListKafkaTopicsExecuteFunc: func(_ v3.ApiListKafkaTopicsRequest) (v3.TopicDataList, *http.Response, error) {
			return v3.TopicDataList{Data: []v3.TopicData{{TopicName: "topic1"}}}, nil, nil
		},
		CreateKafkaTopicFunc: func(_ context.Context, _ string) v3.ApiCreateKafkaTopicRequest {
			return v3.ApiCreateKafkaTopicRequest{}
		},
		CreateKafkaTopicExecuteFunc: func(request v3.ApiCreateKafkaTopicRequest) (v3.TopicData, *http.Response, error) {
			return v3.TopicData{}, nil, nil
		},
	}
	kafkaRestProvider := pcmd.KafkaRESTProvider(func() (*pcmd.KafkaREST, error) {
		return &pcmd.KafkaREST{CloudClient: &ccloudv2.KafkaRestClient{APIClient: apiClient}}, nil
	})
	c.KafkaRESTProvider = &kafkaRestProvider
	return c
}
