package kafka

import (
	"context"
	"github.com/confluentinc/cli/internal/pkg/cmd"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	cliMock "github.com/confluentinc/cli/mock"
	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	kafkarestv3mock "github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3/mock"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"net/http"
	"strings"
	"testing"
)

type KafkaGroupOnPremTestSuite struct {
	suite.Suite
	testClient *kafkarestv3.APIClient
	// Data returned by APIClient
	clusterList *kafkarestv3.ClusterDataList
	consumerGroupList *kafkarestv3.ConsumerGroupDataList
	consumerGroup *kafkarestv3.ConsumerGroupData
	consumerList *kafkarestv3.ConsumerDataList
	lagList *kafkarestv3.ConsumerLagDataList
	lag *kafkarestv3.ConsumerLagData
	lagSummary *kafkarestv3.ConsumerGroupLagSummaryData

}

type testCase struct {
	input string
	expectedOutput string
	expectError bool
	errorMsgContainsAll []string
	message string
}

func (suite *KafkaGroupOnPremTestSuite) SetupSuite() {
	// Define canned test server response data
	clusterId := "cluster-1"
	suite.clusterList = &kafkarestv3.ClusterDataList{
		Data: []kafkarestv3.ClusterData{
			{
				ClusterId: clusterId,
			},
		},
	}
	suite.consumerGroupList = &kafkarestv3.ConsumerGroupDataList{
		Data: []kafkarestv3.ConsumerGroupData{
			{
				ClusterId:       clusterId,
				ConsumerGroupId: "consumer-group-1",
				IsSimple:        true,
				State:           "STABLE",
			},
			{
				ClusterId: clusterId,
				ConsumerGroupId: "consumer-group-2",
				IsSimple: false,
				State: "DEAD",
			},
		},
	}

	suite.consumerGroup = &suite.consumerGroupList.Data[0]

	suite.consumerList = &kafkarestv3.ConsumerDataList{
		Data: []kafkarestv3.ConsumerData{
			{
				ConsumerId: "consumer-1",
			},
			{
				ConsumerId: "consumer-2",
			},

		},
	}

	suite.lagList = &kafkarestv3.ConsumerLagDataList{
		Data: []kafkarestv3.ConsumerLagData{
			{
				ConsumerGroupId: "consumer-group-1",
				TopicName: "topic-1",
				PartitionId: 1,
				CurrentOffset: 1,
				LogEndOffset: 101,
				Lag: 100,
			},
			{
				ConsumerGroupId: "consumer-group-1",
				TopicName: "topic-1",
				PartitionId: 2,
				CurrentOffset: 1,
				LogEndOffset: 11,
				Lag: 10,
			},
		},
	}
	suite.lag = &suite.lagList.Data[0]

	suite.lagSummary = &kafkarestv3.ConsumerGroupLagSummaryData{
		ConsumerGroupId: "consumer-group-1",
		MaxLag: 100,
		TotalLag: 110,
		MaxLagTopicName: "topic-1",
		MaxLagPartitionId: 1,
	}
}

// Create a new groupCommand. Should be called before each test case.
func (suite *KafkaGroupOnPremTestSuite) createGroupCommand() *cobra.Command {
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
	suite.testClient.ConsumerGroupApi = &kafkarestv3mock.ConsumerGroupApi{
		ClustersClusterIdConsumerGroupsGetFunc: func(ctx context.Context, clusterId string) (kafkarestv3.ConsumerGroupDataList, *http.Response, error) {
			// Check if URL is valid
			err := checkURL(suite.testClient.GetConfig().BasePath)
			if err != nil {
				return kafkarestv3.ConsumerGroupDataList{}, nil, err
			}
			// Return canned data
			return *suite.consumerGroupList, nil, nil
		},
		ClustersClusterIdConsumerGroupsConsumerGroupIdGetFunc: func(ctx context.Context, clusterId string, consumerGroupId string) (kafkarestv3.ConsumerGroupData, *http.Response, error) {
			// Check if URL is valid
			err := checkURL(suite.testClient.GetConfig().BasePath)
			if err != nil {
				return kafkarestv3.ConsumerGroupData{}, nil, err
			}
			// Return canned data
			return *suite.consumerGroup, nil, nil
		},
		ClustersClusterIdConsumerGroupsConsumerGroupIdLagsGetFunc: func(ctx context.Context, clusterId string, consumerGroupId string) (kafkarestv3.ConsumerLagDataList, *http.Response, error) {
			// Check if URL is valid
			err := checkURL(suite.testClient.GetConfig().BasePath)
			if err != nil {
				return kafkarestv3.ConsumerLagDataList{}, nil, err
			}
			// Return canned data
			return *suite.lagList, nil, nil
		},
		ClustersClusterIdConsumerGroupsConsumerGroupIdLagSummaryGetFunc: func(ctx context.Context, clusterId string, consumerGroupId string) (kafkarestv3.ConsumerGroupLagSummaryData, *http.Response, error) {
			// Check if URL is valid
			err := checkURL(suite.testClient.GetConfig().BasePath)
			if err != nil {
				return kafkarestv3.ConsumerGroupLagSummaryData{}, nil, err
			}
			// Return canned data
			return *suite.lagSummary, nil, nil
		},
		ClustersClusterIdConsumerGroupsConsumerGroupIdConsumersGetFunc: func(ctx context.Context, clusterId string, consumerGroupId string) (kafkarestv3.ConsumerDataList, *http.Response, error) {
			// Check if URL is valid
			err := checkURL(suite.testClient.GetConfig().BasePath)
			if err != nil {
				return kafkarestv3.ConsumerDataList{}, nil, err
			}
			// Return canned data
			return *suite.consumerList, nil, nil
		},
	}
	suite.testClient.PartitionApi = &kafkarestv3mock.PartitionApi{
		ClustersClusterIdConsumerGroupsConsumerGroupIdLagsTopicNamePartitionsPartitionIdGetFunc: func(ctx context.Context, clusterId string, consumerGroupId string, topicName string, partitionId int32) (kafkarestv3.ConsumerLagData, *http.Response, error) {
			// Check if URL is valid
			err := checkURL(suite.testClient.GetConfig().BasePath)
			if err != nil {
				return kafkarestv3.ConsumerLagData{}, nil, err
			}
			// Return canned data
			return *suite.lag, nil, nil
		},

	}

	provider := suite.getRestProvider()
	testPrerunner := cliMock.NewPreRunnerMock(nil, nil, &provider, v3.AuthenticatedConfluentConfigMock())
	return NewGroupCommandOnPrem(testPrerunner)
}

func (suite *KafkaGroupOnPremTestSuite) TestConfluentListGroups() {
	// Define test cases
	testCases := []testCase {
		// Correct input
		{
			input: "list --url http://localhost:8082",
			expectedOutput:
				"   Cluster  |  ConsumerGroup   | Simple | State   \n" +
				"+-----------+------------------+--------+--------+\n" +
				"  cluster-1 | consumer-group-1 | true   | STABLE  \n" +
				"  cluster-1 | consumer-group-2 | false  | DEAD    \n",
			expectError: false,
			errorMsgContainsAll: []string{},
			message: "correct argument should match expected output",
		},
		// Variable output format
		{
			input: "list --url http://localhost:8082 -o yaml",
			expectedOutput:
				"- cluster: cluster-1\n" +
				"  consumer_group: consumer-group-1\n" +
				"  simple: \"true\"\n" +
				"  state: STABLE\n" +
				"- cluster: cluster-1\n" +
				"  consumer_group: consumer-group-2\n" +
				"  simple: \"false\"\n" +
				"  state: DEAD\n",
			expectError: false,
			errorMsgContainsAll: []string{},
			message: "correct argument should match expected output",
		},
		{
			input: "list --url http://localhost:8082 -o json",
			expectedOutput: "[\n  {\n" +
				"    \"cluster\": \"cluster-1\",\n" +
				"    \"consumer_group\": \"consumer-group-1\",\n" +
				"    \"simple\": \"true\",\n" +
				"    \"state\": \"STABLE\"\n" +
				"  }, \n  {\n" +
				"    \"cluster\": \"cluster-1\",\n" +
				"    \"consumer_group\": \"consumer-group-2\",\n" +
				"    \"simple\": \"false\",\n" +
				"    \"state\": \"DEAD\"\n  }\n]\n",
			expectError: false,
			errorMsgContainsAll: []string{},
			message: "correct argument should match expected output",
		},
		// Invalid url should throw error
		{
			input: "list --url https://localhost:8082",
			expectedOutput: "",
			expectError: true,
			errorMsgContainsAll: []string{"http: server gave HTTP response to HTTPS client"},
			message: "mismatching protocol in url should throw error",
		},
		{
			input: "list --url http:///localhost:8082",
			expectedOutput: "",
			expectError: true,
			errorMsgContainsAll: []string{"no Host"},
			message: "invalid url should throw error",
		},
		{
			input: "list --url http://localhos:8082",
			expectedOutput: "",
			expectError: true,
			errorMsgContainsAll: []string{"no such host"},
			message: "incorrect host in url should throw ierror",
		},
		{
			input: "list --url http://localhost:808",
			expectedOutput: "",
			expectError: true,
			errorMsgContainsAll: []string{"connection refused"},
			message: "incorrect port in url should throw error",
		},
		{
			input: "list --url http://localhost:808a",
			expectedOutput: "",
			expectError: true,
			errorMsgContainsAll: []string{"invalid port"},
			message: "invalid url should throw error",
		},
		// Invalid format string should throw error
		{
			input: "list --url http://localhost:8082 -o hello --no-auth",
			expectedOutput: "",
			expectError: true,
			errorMsgContainsAll: []string{"invalid value", "--output", "hello"},
			message: "invalid format string should throw error",
		},
	}

	// Test test cases
	for _, test := range testCases {
		suite.RunTestCase(test)
	}
}

func (suite *KafkaGroupOnPremTestSuite) TestConfluentDescribeGroup() {
	// Define test cases
	testCases := []testCase {
		{
			// Consumers list is not in expectedOutput because it is generated with an external printer package function
			// which hardcodes output to os.Stdout (rather than cobra.Command.outWriter). We can verify Consumers list
			// data with the following two tests which output json/yaml, and Consumers list output format with the integ
			// tests in kafka_test.go.
			input:          "describe consumer-group-1 --url http://localhost:8082",
			expectedOutput:
				"+-------------------+------------------+\n" +
				"| Cluster           | cluster-1        |\n" +
				"| ConsumerGroup     | consumer-group-1 |\n" +
				"| Coordinator       |                  |\n" +
				"| Simple            | true             |\n" +
				"| PartitionAssignor |                  |\n" +
				"| State             | STABLE           |\n" +
				"+-------------------+------------------+\n\n" +
				"Consumers\n\n",
		},
		{
			input:			"describe consumer-group-1 --url http://localhost:8082 -o json",
			expectedOutput: "{\n" +
				"  \"cluster\": \"cluster-1\",\n" +
				"  \"consumer_group\": \"consumer-group-1\",\n" +
				"  \"coordinator\": \"\",\n" +
				"  \"simple\": true,\n" +
				"  \"partition_assignor\": \"\",\n" +
				"  \"state\": \"STABLE\",\n" +
				"  \"consumers\": [\n    {\n" +
				"      \"consumer_group\": \"consumer-group-1\",\n" +
				"      \"consumer\": \"consumer-1\",\n" +
				"      \"instance\": \"\",\n" +
				"      \"client\": \"\"\n" +
				"    }, \n    {\n" +
				"      \"consumer_group\": \"consumer-group-1\",\n" +
				"      \"consumer\": \"consumer-2\",\n" +
				"      \"instance\": \"\",\n" +
				"      \"client\": \"\"\n    }\n  ]\n}\n",
		},
		{
			input:			"describe consumer-group-1 --url http://localhost:8082 -o yaml",
			expectedOutput:
				"cluster: cluster-1\n" +
				"consumer_group: consumer-group-1\n" +
				"coordinator: \"\"\n" +
				"simple: true\n" +
				"partition_assignor: \"\"\n" +
				"state: STABLE\n" +
				"consumers:\n" +
				"- consumer_group: consumer-group-1\n" +
				"  consumer: consumer-1\n" +
				"  instance: \"\"\n" +
				"  client: \"\"\n" +
				"- consumer_group: consumer-group-1\n" +
				"  consumer: consumer-2\n" +
				"  instance: \"\"\n" +
				"  client: \"\"\n",
		},
		{
			input:               "describe --url http://localhost:8082",
			expectError:         true,
			errorMsgContainsAll: []string{"accepts 1 arg(s), received 0"},
		},
		{
			input:               "describe --group --url http://localhost:8082",
			expectError:         true,
			errorMsgContainsAll: []string{"unknown flag: --group"},
		},
	}

	// Test test cases
	for _, test := range testCases {
		suite.RunTestCase(test)
	}
}

func (suite *KafkaGroupOnPremTestSuite) TestConfluentSummarizeLag() {
	// Define test cases
	testCases := []testCase {
		{
			input:          "lag summarize consumer-group-1 --url http://localhost:8082",
			expectedOutput:
				"+-----------------+------------------+\n" +
				"| Cluster         |                  |\n" +
				"| ConsumerGroup   | consumer-group-1 |\n" +
				"| TotalLag        |              110 |\n" +
				"| MaxLag          |              100 |\n" +
				"| MaxLagConsumer  |                  |\n" +
				"| MaxLagInstance  |                  |\n" +
				"| MaxLagClient    |                  |\n" +
				"| MaxLagTopic     | topic-1          |\n" +
				"| MaxLagPartition |                1 |\n" +
				"+-----------------+------------------+\n",
		},
		{
			input:			"lag summarize consumer-group-1 --url http://localhost:8082 -o json",
			expectedOutput: "{\n" +
				"  \"cluster\": \"\",\n" +
				"  \"consumer_group\": \"consumer-group-1\",\n" +
				"  \"total_lag\": 110,\n" +
				"  \"max_lag\": 100,\n" +
				"  \"max_lag_consumer\": \"\",\n" +
				"  \"max_lag_instance\": \"\",\n" +
				"  \"max_lag_client\": \"\",\n" +
				"  \"max_lag_topic\": \"topic-1\",\n" +
				"  \"max_lag_partition\": 1\n}\n",
		},
		{
			input:			"lag summarize consumer-group-1 --url http://localhost:8082 -o yaml",
			expectedOutput:
				"cluster: \"\"\n" +
				"consumer_group: consumer-group-1\n" +
				"max_lag: 100\n" +
				"max_lag_client: \"\"\n" +
				"max_lag_consumer: \"\"\n" +
				"max_lag_instance: \"\"\n" +
				"max_lag_partition: 1\n" +
				"max_lag_topic: topic-1\n" +
				"total_lag: 110\n",
		},
	}

	// Test test cases
	for _, test := range testCases {
		suite.RunTestCase(test)
	}
}

func (suite *KafkaGroupOnPremTestSuite) TestConfluentListLags() {
	// Define test cases
	testCases := []testCase {
		{
			input:          "lag list consumer-group-1 --url http://localhost:8082",
			expectedOutput:
				"  Cluster |  ConsumerGroup   | Lag | LogEndOffset | CurrentOffset | Consumer | Instance | Client |  Topic  | Partition  \n" +
				"+---------+------------------+-----+--------------+---------------+----------+----------+--------+---------+-----------+\n" +
				"          | consumer-group-1 | 100 |          101 |             1 |          |          |        | topic-1 |         1  \n" +
				"          | consumer-group-1 |  10 |           11 |             1 |          |          |        | topic-1 |         2  \n",
		},
		{
			input:			"lag list consumer-group-1 --url http://localhost:8082 -o json",
			expectedOutput: "[\n  {\n" +
				"    \"client\": \"\",\n" +
				"    \"cluster\": \"\",\n" +
				"    \"consumer\": \"\",\n" +
				"    \"consumer_group\": \"consumer-group-1\",\n" +
				"    \"current_offset\": \"1\",\n" +
				"    \"instance\": \"\",\n" +
				"    \"lag\": \"100\",\n" +
				"    \"log_end_offset\": \"101\",\n" +
				"    \"partition\": \"1\",\n" +
				"    \"topic\": \"topic-1\"\n" +
				"  }, \n  {\n" +
				"    \"client\": \"\",\n" +
				"    \"cluster\": \"\",\n" +
				"    \"consumer\": \"\",\n" +
				"    \"consumer_group\": \"consumer-group-1\",\n" +
				"    \"current_offset\": \"1\",\n" +
				"    \"instance\": \"\",\n" +
				"    \"lag\": \"10\",\n" +
				"    \"log_end_offset\": \"11\",\n" +
				"    \"partition\": \"2\",\n" +
				"    \"topic\": \"topic-1\"\n  }\n]\n",
		},
		{
			input:			"lag list consumer-group-1 --url http://localhost:8082 -o yaml",
			expectedOutput:
				"- client: \"\"\n" +
				"  cluster: \"\"\n" +
				"  consumer: \"\"\n" +
				"  consumer_group: consumer-group-1\n" +
				"  current_offset: \"1\"\n" +
				"  instance: \"\"\n" +
				"  lag: \"100\"\n" +
				"  log_end_offset: \"101\"\n" +
				"  partition: \"1\"\n" +
				"  topic: topic-1\n" +
				"- client: \"\"\n" +
				"  cluster: \"\"\n" +
				"  consumer: \"\"\n" +
				"  consumer_group: consumer-group-1\n" +
				"  current_offset: \"1\"\n" +
				"  instance: \"\"\n" +
				"  lag: \"10\"\n" +
				"  log_end_offset: \"11\"\n" +
				"  partition: \"2\"\n" +
				"  topic: topic-1\n",
		},
	}

	// Test test cases
	for _, test := range testCases {
		suite.RunTestCase(test)
	}
}

func (suite *KafkaGroupOnPremTestSuite) TestConfluentGetLag() {
	// Define test cases
	testCases := []testCase {
		{
			input:          "lag get consumer-group-1 --topic topic-1 --partition 1 --url http://localhost:8082",
			expectedOutput:
				"+---------------+------------------+\n" +
				"| Cluster       |                  |\n" +
				"| ConsumerGroup | consumer-group-1 |\n" +
				"| Lag           |              100 |\n" +
				"| LogEndOffset  |              101 |\n" +
				"| CurrentOffset |                1 |\n" +
				"| Consumer      |                  |\n" +
				"| Instance      |                  |\n" +
				"| Client        |                  |\n" +
				"| Topic         | topic-1          |\n" +
				"| Partition     |                1 |\n" +
				"+---------------+------------------+\n",
		},
		{
			input:			"lag get consumer-group-1 --topic topic-1 --partition 1 --url http://localhost:8082 -o json",
			expectedOutput: "{\n" +
				"  \"cluster\": \"\",\n" +
				"  \"consumer_group\": \"consumer-group-1\",\n" +
				"  \"lag\": 100,\n" +
				"  \"log_end_offset\": 101,\n" +
				"  \"current_offset\": 1,\n" +
				"  \"consumer\": \"\",\n" +
				"  \"instance\": \"\",\n" +
				"  \"client\": \"\",\n" +
				"  \"topic\": \"topic-1\",\n" +
				"  \"partition\": 1\n}\n",
		},
		{
			input:			"lag get consumer-group-1 --topic topic-1 --partition 1 --url http://localhost:8082 -o yaml",
			expectedOutput:
				"client: \"\"\n" +
				"cluster: \"\"\n" +
				"consumer: \"\"\n" +
				"consumer_group: consumer-group-1\n" +
				"current_offset: 1\n" +
				"instance: \"\"\n" +
				"lag: 100\n" +
				"log_end_offset: 101\n" +
				"partition: 1\n" +
				"topic: topic-1\n",
		},
	}

	// Test test cases
	for _, test := range testCases {
		suite.RunTestCase(test)
	}
}

func (suite *KafkaGroupOnPremTestSuite) getRestProvider() cmd.KafkaRESTProvider {
	return func() (*cmd.KafkaREST, error) {
		return &cmd.KafkaREST{Client: suite.testClient, Context: context.Background()}, nil
	}
}

func (suite *KafkaGroupOnPremTestSuite) RunTestCase(test testCase) {
	groupCommand := suite.createGroupCommand()
	_, output, err := executeCommand(groupCommand, strings.Split(test.input, " "))

	if test.expectError == false {
		require.NoError(suite.T(), err, test.message)
		require.Equal(suite.T(), test.expectedOutput, output, test.message)
	} else {
		require.Error(suite.T(), err, test.message)
		for _, errorMsgContains := range test.errorMsgContainsAll {
			require.Contains(suite.T(), err.Error(), errorMsgContains, test.message)
		}
	}
}

func TestConfluentKafkaGroup(t *testing.T) {
	suite.Run(t, new(KafkaGroupOnPremTestSuite))
}
