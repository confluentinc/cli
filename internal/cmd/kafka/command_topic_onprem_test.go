package kafka

import (
	"bytes"
	"context"
	"fmt"
	"github.com/confluentinc/cli/internal/pkg/cmd"
	"net/http"
	purl "net/url"
	"strings"
	"testing"

	cliMock "github.com/confluentinc/cli/mock"
	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	kafkarestv3mock "github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3/mock"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	// Expected output of tests
	ExpectedListTopicsOutput = `   Name
+---------+
  topic-1
  topic-2
  topic-3
`
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
	clusterList *kafkarestv3.ClusterDataList
	topicList   *kafkarestv3.TopicDataList
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
	}
	provider := suite.getRestProvider()
	testPrerunner := cliMock.NewPreRunnerMock(nil, nil, &provider, nil)
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
		{input: "list --url https://localhost:8082", expectedOutput: "", expectError: true, errorMsgContainsAll: []string{"invalid value", "--url", "https"}, message: "mismatching protocol in url should throw error"},
		{input: "list --url http:///localhost:8082", expectedOutput: "", expectError: true, errorMsgContainsAll: []string{"invalid value", "--url", "no Host"}, message: "invalid url should throw error"},
		{input: "list --url http://localhos:8082", expectedOutput: "", expectError: true, errorMsgContainsAll: []string{"invalid value", "--url", "no such host"}, message: "incorrect host in url should throw ierror"},
		{input: "list --url http://localhost:808", expectedOutput: "", expectError: true, errorMsgContainsAll: []string{"invalid value", "--url", "connection refused"}, message: "incorrect port in url should throw error"},
		{input: "list --url http://localhost:808a", expectedOutput: "", expectError: true, errorMsgContainsAll: []string{"invalid value", "--url", "invalid port"}, message: "invalid url should throw error"},
		// Invalid format string should throw error
		{input: "list --url http://localhost:8082 -o hello", expectedOutput: "", expectError: true, errorMsgContainsAll: []string{"invalid value", "--output", "hello"}, message: "invalid format string should throw error"},
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
	return  func() (*cmd.KafkaREST, error) {
		return &cmd.KafkaREST{Client: suite.testClient, Context: context.Background()}, nil
	}
}

func TestConfluentKafkaTopicSubcommands(t *testing.T) {
	suite.Run(t, new(KafkaTopicTestSuite))
}
