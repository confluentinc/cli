package kafka

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	purl "net/url"
	"strings"
	"testing"

	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"

	"github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	kafkarestv3mock "github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3/mock"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	cliMock "github.com/confluentinc/cli/mock"
)

const (
	// Expected output of tests
	ExpectedListTopicsOutput     = "   Name    \n-----------\n  topic-1  \n  topic-2  \n  topic-3  \n"
	ExpectedListTopicsYamlOutput = `- name: topic-1
- name: topic-2
- name: topic-3
`
	ExpectedListTopicsJsonOutput = "[\n  {\n    \"name\": \"topic-1\"\n  }, \n  {\n    \"name\": \"topic-2\"\n  }, \n  {\n    \"name\": \"topic-3\"\n  }\n]\n"
)

type KafkaTopicOnPremTestSuite struct {
	suite.Suite
	testClient *kafkarestv3.APIClient
	// Data returned by APIClient
	clusterList    *kafkarestv3.ClusterDataList
	topicList      *kafkarestv3.TopicDataList
	partitionList  *kafkarestv3.PartitionDataList
	configDataList *kafkarestv3.TopicConfigDataList
	replicaList    *kafkarestv3.ReplicaDataList

	createTopicName              string
	createTopicPartitionsCount   int32
	createTopicReplicationFactor int32
	createTopicConfigs           []kafkarestv3.CreateTopicRequestDataConfigs
	updateTopicName              string
	updateTopicData              []kafkarestv3.AlterConfigBatchRequestDataData
}

func (suite *KafkaTopicOnPremTestSuite) SetupSuite() {
	// Define canned test server response data
	suite.clusterList = &kafkarestv3.ClusterDataList{
		Data: []kafkarestv3.ClusterData{
			{
				ClusterId: "cluster-1",
			}},
	}

	suite.topicList = &kafkarestv3.TopicDataList{
		Data: []kafkarestv3.TopicData{
			{
				TopicName: "topic-1",
			},
			{
				TopicName: "topic-2",
			},
			{
				TopicName: "topic-3",
			},
		},
	}

	suite.partitionList = &kafkarestv3.PartitionDataList{
		Data: []kafkarestv3.PartitionData{
			{
				TopicName: "topic-1",
			},
			{
				TopicName: "topic-2",
			},
			{
				TopicName: "topic-3",
			},
		},
	}
	value := "testValue"
	suite.configDataList = &kafkarestv3.TopicConfigDataList{
		Data: []kafkarestv3.TopicConfigData{
			{
				ClusterId: clusterId,
				TopicName: topicName,
				Value:     &value,
			},
		},
	}
	suite.replicaList = &kafkarestv3.ReplicaDataList{
		Data: []kafkarestv3.ReplicaData{
			{
				PartitionId: 1,
				ClusterId:   clusterId,
				TopicName:   topicName,
				BrokerId:    50,
				IsLeader:    true,
				IsInSync:    true,
			},
		},
	}
}

// Helper for testAPIClient for parsing URL
func checkURL(url string) error {
	// Assumes the address is: localhost:8082
	parsedUrl, err := purl.Parse(url)
	if err != nil {
		return err
	} else if strings.Contains(parsedUrl.Scheme, "https") { // if not http
		return &purl.Error{Op: "", URL: "", Err: fmt.Errorf("http: server gave HTTP response to HTTPS client")}
	} else if parsedUrl.Hostname() == "" {
		return &purl.Error{Op: "", URL: "", Err: fmt.Errorf("http: no Host in request URL")}
	} else if parsedUrl.Hostname() != "localhost" { // if not localhost
		return &purl.Error{Op: "", URL: "", Err: fmt.Errorf("dial tcp: lookup %s: no such host", parsedUrl.Hostname())}
	} else if parsedUrl.Port() != "8082" { // if not 8082
		return &purl.Error{Op: "", URL: "", Err: fmt.Errorf(" dial tcp [::1]:%s: connect: connection refused", parsedUrl.Port())}
	}
	return nil
}

// Create a new topicCommand. Should be called before each test case.
func (suite *KafkaTopicOnPremTestSuite) createCommand() *cobra.Command {
	// Define testAPIClient
	suite.testClient = kafkarestv3.NewAPIClient(kafkarestv3.NewConfiguration())
	suite.testClient.ClusterApi = &kafkarestv3mock.ClusterApi{
		ClustersGetFunc: func(ctx context.Context) (kafkarestv3.ClusterDataList, *http.Response, error) {
			// Check if URL is valid
			err := checkURL(suite.testClient.GetConfig().BasePath)
			if err != nil {
				return kafkarestv3.ClusterDataList{}, nil, err
			}
			// Return canned data
			return *suite.clusterList, nil, nil
		},
	}
	suite.testClient.TopicApi = &kafkarestv3mock.TopicApi{
		ClustersClusterIdTopicsGetFunc: func(ctx context.Context, clusterId string) (kafkarestv3.TopicDataList, *http.Response, error) {
			// Check if URL is valid
			err := checkURL(suite.testClient.GetConfig().BasePath)
			if err != nil {
				return kafkarestv3.TopicDataList{}, nil, err
			}
			// Return canned data
			return *suite.topicList, nil, nil
		},
		ClustersClusterIdTopicsPostFunc: func(ctx context.Context, clusterId string, localVarOptionals *kafkarestv3.ClustersClusterIdTopicsPostOpts) (kafkarestv3.TopicData, *http.Response, error) {
			err := checkURL(suite.testClient.GetConfig().BasePath)
			if err != nil {
				return kafkarestv3.TopicData{}, nil, err
			}
			t := suite.T()
			require.True(t, localVarOptionals.CreateTopicRequestData.IsSet())
			topicCreateData := localVarOptionals.CreateTopicRequestData.Value().(kafkarestv3.CreateTopicRequestData)
			require.Equal(t, suite.createTopicName, topicCreateData.TopicName)
			require.Equal(t, suite.createTopicPartitionsCount, topicCreateData.PartitionsCount)
			require.Equal(t, suite.createTopicReplicationFactor, topicCreateData.ReplicationFactor)
			require.Equal(t, len(suite.createTopicConfigs), len(topicCreateData.Configs))
			values := make(map[string]string)
			for _, requestEntry := range topicCreateData.Configs {
				values[requestEntry.Name] = *requestEntry.Value
			}
			for _, expectedEntry := range suite.createTopicConfigs {
				require.Equal(t, values[expectedEntry.Name], *expectedEntry.Value)
			}
			return kafkarestv3.TopicData{
				ClusterId:         clusterId,
				TopicName:         topicCreateData.TopicName,
				ReplicationFactor: topicCreateData.ReplicationFactor,
			}, nil, nil
		},
		ClustersClusterIdTopicsTopicNameDeleteFunc: func(ctx context.Context, clusterId string, topicName string) (*http.Response, error) {
			// Check if URL is valid
			err := checkURL(suite.testClient.GetConfig().BasePath)
			if err != nil {
				return nil, err
			}
			return nil, nil
		},
	}
	suite.testClient.ConfigsApi = &kafkarestv3mock.ConfigsApi{
		ClustersClusterIdTopicsTopicNameConfigsalterPostFunc: func(ctx context.Context, clusterId string, topicName string, localVarOptionals *kafkarestv3.ClustersClusterIdTopicsTopicNameConfigsalterPostOpts) (*http.Response, error) {
			err := checkURL(suite.testClient.GetConfig().BasePath)
			if err != nil {
				return nil, err
			}
			t := suite.T()
			topicUpdateOpts := localVarOptionals.AlterConfigBatchRequestData.Value().(kafkarestv3.AlterConfigBatchRequestData)
			require.Equal(t, suite.updateTopicName, topicName)
			require.Equal(t, len(suite.updateTopicData), len(topicUpdateOpts.Data))
			values := make(map[string]string)
			for _, requestEntry := range topicUpdateOpts.Data {
				values[requestEntry.Name] = *requestEntry.Value
			}
			for _, expectedEntry := range suite.updateTopicData {
				require.Equal(t, values[expectedEntry.Name], *expectedEntry.Value)
			}
			return nil, nil
		},
		ClustersClusterIdTopicsTopicNameConfigsGetFunc: func(ctx context.Context, clusterId string, topicName string) (kafkarestv3.TopicConfigDataList, *http.Response, error) {
			err := checkURL(suite.testClient.GetConfig().BasePath)
			if err != nil {
				return kafkarestv3.TopicConfigDataList{}, nil, err
			}
			return *suite.configDataList, nil, nil
		},
	}
	suite.testClient.PartitionApi = &kafkarestv3mock.PartitionApi{
		ClustersClusterIdTopicsTopicNamePartitionsGetFunc: func(ctx context.Context, clusterId string, topicName string) (kafkarestv3.PartitionDataList, *http.Response, error) {
			err := checkURL(suite.testClient.GetConfig().BasePath)
			if err != nil {
				return kafkarestv3.PartitionDataList{}, nil, err
			}
			return *suite.partitionList, nil, nil
		},
	}
	suite.testClient.ReplicaApi = &kafkarestv3mock.ReplicaApi{
		ClustersClusterIdTopicsTopicNamePartitionsPartitionIdReplicasGetFunc: func(ctx context.Context, clusterId string, topicName string, partitionId int32) (kafkarestv3.ReplicaDataList, *http.Response, error) {
			err := checkURL(suite.testClient.GetConfig().BasePath)
			if err != nil {
				return kafkarestv3.ReplicaDataList{}, nil, err
			}
			return *suite.replicaList, nil, nil
		},
	}
	provider := suite.getRestProvider()
	testPrerunner := cliMock.NewPreRunnerMock(nil, nil, &provider, v3.AuthenticatedConfluentConfigMock())
	return NewTopicCommandOnPrem(testPrerunner)
}

// Executes the given command with the given args, returns the command executed, stdout and error.
func executeCommand(command *cobra.Command, args []string) (*cobra.Command, string, error) {
	buf := new(bytes.Buffer)
	command.SetOut(buf)
	command.SetErr(buf)
	command.SetArgs(args)
	c, err := command.ExecuteC()

	return c, buf.String(), err
}

func (suite *KafkaTopicOnPremTestSuite) TestConfluentListTopics() {
	// Define test cases
	cases := []struct {
		input               string
		expectedOutput      string
		expectError         bool
		errorMsgContainsAll []string
		message             string
	}{
		// Correct input
		{input: "list --url http://localhost:8082", expectedOutput: ExpectedListTopicsOutput, expectError: false, errorMsgContainsAll: []string{}, message: "correct argument should match expected output"},
		// Variable output format
		{input: "list --url http://localhost:8082 -o yaml", expectedOutput: ExpectedListTopicsYamlOutput, expectError: false, errorMsgContainsAll: []string{}, message: "correct argument should match expected output"},
		{input: "list --url http://localhost:8082 -o json", expectedOutput: ExpectedListTopicsJsonOutput, expectError: false, errorMsgContainsAll: []string{}, message: "correct argument should match expected output"},
		// Invalid url should throw error
		{input: "list --url https://localhost:8082", expectedOutput: "", expectError: true, errorMsgContainsAll: []string{"http: server gave HTTP response to HTTPS client"}, message: "mismatching protocol in url should throw error"},
		{input: "list --url http:///localhost:8082", expectedOutput: "", expectError: true, errorMsgContainsAll: []string{"no Host"}, message: "invalid url should throw error"},
		{input: "list --url http://localhos:8082", expectedOutput: "", expectError: true, errorMsgContainsAll: []string{"no such host"}, message: "incorrect host in url should throw ierror"},
		{input: "list --url http://localhost:808", expectedOutput: "", expectError: true, errorMsgContainsAll: []string{"connection refused"}, message: "incorrect port in url should throw error"},
		{input: "list --url http://localhost:808a", expectedOutput: "", expectError: true, errorMsgContainsAll: []string{"invalid port"}, message: "invalid url should throw error"},
		// Invalid format string should throw error
		{input: "list --url http://localhost:8082 -o hello --no-auth", expectedOutput: "", expectError: true, errorMsgContainsAll: []string{"invalid value", "--output", "hello"}, message: "invalid format string should throw error"},
	}

	// Test test cases
	for _, testCase := range cases {
		topicCommand := suite.createCommand()
		_, output, err := executeCommand(topicCommand, strings.Split(testCase.input, " "))

		if testCase.expectError == false {
			require.NoError(suite.T(), err, testCase.message)
			require.Equal(suite.T(), testCase.expectedOutput, output, testCase.message)
		} else {
			require.Error(suite.T(), err, testCase.message)
			for _, errorMsgContains := range testCase.errorMsgContainsAll {
				require.Contains(suite.T(), err.Error(), errorMsgContains, testCase.message)
			}
		}
	}
}

func (suite *KafkaTopicOnPremTestSuite) TestConfluentCreateTopic() {
	retentionValue := "1"
	compressionValue := "gzip"
	// Define test cases
	cases := []struct {
		input                        string
		expectedOutput               string
		expectError                  bool
		errorMsgContainsAll          []string
		message                      string
		createTopicName              string
		createTopicReplicationFactor int32
		createTopicPartitionsCount   int32
		createTopicConfigs           []kafkarestv3.CreateTopicRequestDataConfigs
	}{
		{
			input:               "create topic-X --url http://localhost:8082 --config retention.ms=1,compression",
			expectError:         true,
			errorMsgContainsAll: []string{`failed to parse "key=value" pattern from configuration: compression`},
			createTopicName:     "topic-X",
		},
		{
			input:                        "create topic-X --url http://localhost:8082 --config retention.ms=1,compression.type=gzip --replication-factor 2 --partitions 4",
			expectedOutput:               fmt.Sprintf(errors.CreatedTopicMsg, "topic-X"),
			createTopicName:              "topic-X",
			createTopicPartitionsCount:   4,
			createTopicReplicationFactor: 2,
			createTopicConfigs: []kafkarestv3.CreateTopicRequestDataConfigs{
				{
					Name:  "retention.ms",
					Value: &retentionValue,
				},
				{
					Name:  "compression.type",
					Value: &compressionValue,
				},
			},
		},
	}

	// Test test cases
	for _, testCase := range cases {
		suite.createTopicName = testCase.createTopicName
		suite.createTopicPartitionsCount = testCase.createTopicPartitionsCount
		suite.createTopicReplicationFactor = testCase.createTopicReplicationFactor
		suite.createTopicConfigs = testCase.createTopicConfigs
		topicCommand := suite.createCommand()
		_, output, err := executeCommand(topicCommand, strings.Split(testCase.input, " "))

		if testCase.expectError == false {
			require.NoError(suite.T(), err, testCase.message)
			require.Equal(suite.T(), testCase.expectedOutput, output, testCase.message)
		} else {
			require.Error(suite.T(), err, testCase.message)
			for _, errorMsgContains := range testCase.errorMsgContainsAll {
				require.Contains(suite.T(), err.Error(), errorMsgContains, testCase.message)
			}
		}
	}
}

func (suite *KafkaTopicOnPremTestSuite) TestConfluentUpdateTopic() {
	retentionValue := "1"
	compressionValue := "gzip"
	// Define test cases
	cases := []struct {
		input               string
		expectedOutput      string
		expectError         bool
		errorMsgContainsAll []string
		message             string
		updateTopicName     string
		updateTopicData     []kafkarestv3.AlterConfigBatchRequestDataData
	}{
		{
			input:               "update topic-X --url http://localhost:8082 --config retention.ms",
			errorMsgContainsAll: []string{`failed to parse "key=value" pattern from configuration: retention.ms`},
			expectError:         true,
			updateTopicName:     "topic-X",
			updateTopicData:     []kafkarestv3.AlterConfigBatchRequestDataData{},
		},
		{
			input:           "update topic-X --url http://localhost:8082",
			updateTopicName: "topic-X",
			expectedOutput:  fmt.Sprintf(errors.UpdateTopicConfigMsg, "topic-X"),
		},
		{
			input:           "update topic-Y --url http://localhost:8082 --config retention.ms=1,compression.type=gzip",
			expectedOutput:  fmt.Sprintf(errors.UpdateTopicConfigMsg, "topic-Y"), // update table gets printed to stdout so dont include in expect
			updateTopicName: "topic-Y",
			updateTopicData: []kafkarestv3.AlterConfigBatchRequestDataData{
				{
					Name:  "retention.ms",
					Value: &retentionValue,
				},
				{
					Name:  "compression.type",
					Value: &compressionValue,
				},
			},
		},
	}

	// Test test cases
	for _, testCase := range cases {
		suite.updateTopicName = testCase.updateTopicName
		suite.updateTopicData = testCase.updateTopicData
		topicCommand := suite.createCommand()
		_, output, err := executeCommand(topicCommand, strings.Split(testCase.input, " "))

		if testCase.expectError == false {
			require.NoError(suite.T(), err, testCase.message)
			require.Equal(suite.T(), testCase.expectedOutput, output, testCase.message)
		} else {
			require.Error(suite.T(), err, testCase.message)
			for _, errorMsgContains := range testCase.errorMsgContainsAll {
				require.Contains(suite.T(), err.Error(), errorMsgContains, testCase.message)
			}
		}
	}
}

func (suite *KafkaTopicOnPremTestSuite) TestConfluentDescribeTopic() {
	// Define test cases
	cases := []struct {
		input               string
		expectedOutput      string
		expectError         bool
		errorMsgContainsAll []string
		message             string
	}{
		{
			input:          "describe topic --url http://localhost:8082",
			expectedOutput: "Topic: topic\nPartitionCount: 3\nReplicationFactor: 1\n\n\nConfiguration\n\n",
		},
		{
			input:               "describe --topic --url http://localhost:8082",
			expectError:         true,
			errorMsgContainsAll: []string{"unknown flag: --topic"},
		},
	}

	// Test test cases
	for _, testCase := range cases {
		topicCommand := suite.createCommand()
		_, output, err := executeCommand(topicCommand, strings.Split(testCase.input, " "))

		if testCase.expectError == false {
			require.NoError(suite.T(), err, testCase.message)
			require.Equal(suite.T(), testCase.expectedOutput, output, testCase.message)
		} else {
			require.Error(suite.T(), err, testCase.message)
			for _, errorMsgContains := range testCase.errorMsgContainsAll {
				require.Contains(suite.T(), err.Error(), errorMsgContains, testCase.message)
			}
		}
	}
}

func (suite *KafkaTopicOnPremTestSuite) TestConfluentDeleteTopic() {
	// Define test cases
	cases := []struct {
		input               string
		expectedOutput      string
		expectError         bool
		errorMsgContainsAll []string
		message             string
	}{
		{
			input:          "delete topicDelete --url http://localhost:8082",
			expectedOutput: "Deleted topic \"topicDelete\".\n",
		},
		{
			input:               "delete --topic --url http://localhost:8082",
			expectError:         true,
			errorMsgContainsAll: []string{"unknown flag: --topic"},
		},
	}

	// Test test cases
	for _, testCase := range cases {
		topicCommand := suite.createCommand()
		_, output, err := executeCommand(topicCommand, strings.Split(testCase.input, " "))

		if testCase.expectError == false {
			require.NoError(suite.T(), err, testCase.message)
			require.Equal(suite.T(), testCase.expectedOutput, output, testCase.message)
		} else {
			require.Error(suite.T(), err, testCase.message)
			for _, errorMsgContains := range testCase.errorMsgContainsAll {
				require.Contains(suite.T(), err.Error(), errorMsgContains, testCase.message)
			}
		}
	}
}

func (suite *KafkaTopicOnPremTestSuite) getRestProvider() cmd.KafkaRESTProvider {
	return func() (*cmd.KafkaREST, error) {
		return &cmd.KafkaREST{Client: suite.testClient, Context: context.Background()}, nil
	}
}

func TestConfluentKafkaTopic(t *testing.T) {
	suite.Run(t, new(KafkaTopicOnPremTestSuite))
}
