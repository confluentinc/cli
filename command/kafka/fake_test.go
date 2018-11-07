package kafka

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cc-structs/kafka/org/v1"
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/cli/shared"
	"github.com/confluentinc/cli/shared/kafka"
)

var conf *shared.Config

/*************** TEST command_acl ***************/
var resourcePatterns = []struct {
	args    []string
	pattern *kafka.ResourcePatternConfig
}{
	{
		args: []string{"--cluster"},
		pattern: &kafka.ResourcePatternConfig{ResourceType: "CLUSTER", Name: "kafka-cluster",
			PatternType: kafka.ResourcePatternConfig_LITERAL.String()},
	},
	{
		args: []string{"--topic", "test-topic"},
		pattern: &kafka.ResourcePatternConfig{ResourceType: "TOPIC", Name: "test-topic",
			PatternType: kafka.ResourcePatternConfig_LITERAL.String()},
	},
	{
		args: []string{"--topic", "test-topic*"},
		pattern: &kafka.ResourcePatternConfig{ResourceType: "TOPIC", Name: "test-topic",
			PatternType: kafka.ResourcePatternConfig_PREFIXED.String()},
	},
	{
		args: []string{"--consumer_group", "test-group"},
		pattern: &kafka.ResourcePatternConfig{ResourceType: "GROUP", Name: "test-group",
			PatternType: kafka.ResourcePatternConfig_LITERAL.String()},
	},
	{
		args: []string{"--consumer_group", "test-group*"},
		pattern: &kafka.ResourcePatternConfig{ResourceType: "GROUP", Name: "test-group",
			PatternType: kafka.ResourcePatternConfig_PREFIXED.String()},
	},
	{
		args: []string{"--transactional_id", "test-transactional_id"},
		pattern: &kafka.ResourcePatternConfig{ResourceType: "TRANSACTIONAL_ID", Name: "test-transactional_id",
			PatternType: kafka.ResourcePatternConfig_LITERAL.String()},
	},
	{
		args: []string{"--transactional_id", "test-transactional_id*"},
		pattern: &kafka.ResourcePatternConfig{ResourceType: "TRANSACTIONAL_ID", Name: "test-transactional_id",
			PatternType: kafka.ResourcePatternConfig_PREFIXED.String()},
	},
}

var aclEntries = []struct {
	args  []string
	entry *kafka.AccessControlEntryConfig
}{
	{
		args: []string{"--allow", "--principal", "test_user", "--operation", "read"},
		entry: &kafka.AccessControlEntryConfig{PermissionType: kafka.AccessControlEntryConfig_ALLOW.String(),
			Principal: "user:test_user", Operation: kafka.AccessControlEntryConfig_READ.String(), Host: "*"},
	},
	{
		args: []string{"--deny", "--principal", "test_user", "--operation", "read"},
		entry: &kafka.AccessControlEntryConfig{PermissionType: kafka.AccessControlEntryConfig_DENY.String(),
			Principal: "user:test_user", Operation: kafka.AccessControlEntryConfig_READ.String(), Host: "*"},
	},
	{
		args: []string{"--allow", "--principal", "test_user", "--operation", "write"},
		entry: &kafka.AccessControlEntryConfig{PermissionType: kafka.AccessControlEntryConfig_ALLOW.String(),
			Principal: "user:test_user", Operation: kafka.AccessControlEntryConfig_WRITE.String(), Host: "*"},
	},
	{
		args: []string{"--deny", "--principal", "test_user", "--operation", "write"},
		entry: &kafka.AccessControlEntryConfig{PermissionType: kafka.AccessControlEntryConfig_DENY.String(),
			Principal: "user:test_user", Operation: kafka.AccessControlEntryConfig_WRITE.String(), Host: "*"},
	},
	{
		args: []string{"--allow", "--principal", "test_user", "--operation", "create"},
		entry: &kafka.AccessControlEntryConfig{PermissionType: kafka.AccessControlEntryConfig_ALLOW.String(),
			Principal: "user:test_user", Operation: kafka.AccessControlEntryConfig_CREATE.String(), Host: "*"},
	},
	{
		args: []string{"--deny", "--principal", "test_user", "--operation", "create"},
		entry: &kafka.AccessControlEntryConfig{PermissionType: kafka.AccessControlEntryConfig_DENY.String(),
			Principal: "user:test_user", Operation: kafka.AccessControlEntryConfig_CREATE.String(), Host: "*"},
	},
	{
		args: []string{"--allow", "--principal", "test_user", "--operation", "delete"},
		entry: &kafka.AccessControlEntryConfig{PermissionType: kafka.AccessControlEntryConfig_ALLOW.String(),
			Principal: "user:test_user", Operation: kafka.AccessControlEntryConfig_DELETE.String(), Host: "*"},
	},
	{
		args: []string{"--deny", "--principal", "test_user", "--operation", "delete"},
		entry: &kafka.AccessControlEntryConfig{PermissionType: kafka.AccessControlEntryConfig_DENY.String(),
			Principal: "user:test_user", Operation: kafka.AccessControlEntryConfig_DELETE.String(), Host: "*"},
	},
	{
		args: []string{"--allow", "--principal", "test_user", "--operation", "alter"},
		entry: &kafka.AccessControlEntryConfig{PermissionType: kafka.AccessControlEntryConfig_ALLOW.String(),
			Principal: "user:test_user", Operation: kafka.AccessControlEntryConfig_ALTER.String(), Host: "*"},
	},
	{
		args: []string{"--deny", "--principal", "test_user", "--operation", "alter"},
		entry: &kafka.AccessControlEntryConfig{PermissionType: kafka.AccessControlEntryConfig_DENY.String(),
			Principal: "user:test_user", Operation: kafka.AccessControlEntryConfig_ALTER.String(), Host: "*"},
	},
	{
		args: []string{"--allow", "--principal", "test_user", "--operation", "describe"},
		entry: &kafka.AccessControlEntryConfig{PermissionType: kafka.AccessControlEntryConfig_ALLOW.String(),
			Principal: "user:test_user", Operation: kafka.AccessControlEntryConfig_DESCRIBE.String(), Host: "*"},
	},
	{
		args: []string{"--deny", "--principal", "test_user", "--operation", "describe"},
		entry: &kafka.AccessControlEntryConfig{PermissionType: kafka.AccessControlEntryConfig_DENY.String(),
			Principal: "user:test_user", Operation: kafka.AccessControlEntryConfig_DESCRIBE.String(), Host: "*"},
	},
	{
		args: []string{"--allow", "--principal", "test_user", "--operation", "cluster_action"},
		entry: &kafka.AccessControlEntryConfig{PermissionType: kafka.AccessControlEntryConfig_ALLOW.String(),
			Principal: "user:test_user", Operation: kafka.AccessControlEntryConfig_CLUSTER_ACTION.String(), Host: "*"},
	},
	{
		args: []string{"--deny", "--principal", "test_user", "--operation", "cluster_action"},
		entry: &kafka.AccessControlEntryConfig{PermissionType: kafka.AccessControlEntryConfig_DENY.String(),
			Principal: "user:test_user", Operation: kafka.AccessControlEntryConfig_CLUSTER_ACTION.String(), Host: "*"},
	},
	{
		args: []string{"--allow", "--principal", "test_user", "--operation", "describe_configs"},
		entry: &kafka.AccessControlEntryConfig{PermissionType: kafka.AccessControlEntryConfig_ALLOW.String(),
			Principal: "user:test_user", Operation: kafka.AccessControlEntryConfig_DESCRIBE_CONFIGS.String(), Host: "*"},
	},
	{
		args: []string{"--deny", "--principal", "test_user", "--operation", "describe_configs"},
		entry: &kafka.AccessControlEntryConfig{PermissionType: kafka.AccessControlEntryConfig_DENY.String(),
			Principal: "user:test_user", Operation: kafka.AccessControlEntryConfig_DESCRIBE_CONFIGS.String(), Host: "*"},
	},
	{
		args: []string{"--allow", "--principal", "test_user", "--operation", "alter_configs"},
		entry: &kafka.AccessControlEntryConfig{PermissionType: kafka.AccessControlEntryConfig_ALLOW.String(),
			Principal: "user:test_user", Operation: kafka.AccessControlEntryConfig_ALTER_CONFIGS.String(), Host: "*"},
	},
	{
		args: []string{"--deny", "--principal", "test_user", "--operation", "alter_configs"},
		entry: &kafka.AccessControlEntryConfig{PermissionType: kafka.AccessControlEntryConfig_DENY.String(),
			Principal: "user:test_user", Operation: kafka.AccessControlEntryConfig_ALTER_CONFIGS.String(), Host: "*"},
	},
	{
		args: []string{"--allow", "--principal", "test_user", "--operation", "idempotent_write"},
		entry: &kafka.AccessControlEntryConfig{PermissionType: kafka.AccessControlEntryConfig_ALLOW.String(),
			Principal: "user:test_user", Operation: kafka.AccessControlEntryConfig_IDEMPOTENT_WRITE.String(), Host: "*"},
	},
	{
		args: []string{"--deny", "--principal", "test_user", "--operation", "idempotent_write"},
		entry: &kafka.AccessControlEntryConfig{PermissionType: kafka.AccessControlEntryConfig_DENY.String(),
			Principal: "user:test_user", Operation: kafka.AccessControlEntryConfig_IDEMPOTENT_WRITE.String(), Host: "*"},
	},
}

func TestCreateACL(t *testing.T) {
	expect := make(chan interface{})
	for _, resource := range resourcePatterns {
		args := append([]string{"acl", "create"}, resource.args...)
		for _, entry := range aclEntries {
			cmd := NewCMD(expect)
			cmd.SetArgs(append(args, entry.args...))

			go func() {
				expect <- &kafka.KafkaAPIACLRequest{Pattern: resource.pattern, Entry: entry.entry}
			}()

			if err := cmd.Execute(); err != nil {
				t.Errorf("error: %s", err)
			}
		}
	}
}

func TestDeleteACL(t *testing.T) {
	expect := make(chan interface{})
	for _, resource := range resourcePatterns {
		args := append([]string{"acl", "delete"}, resource.args...)
		for _, entry := range aclEntries {
			cmd := NewCMD(expect)
			cmd.SetArgs(append(args, entry.args...))

			go func() {
				expect <- ConvertToFilter(&kafka.KafkaAPIACLRequest{Pattern: resource.pattern, Entry: entry.entry})
			}()

			if err := cmd.Execute(); err != nil {
				t.Errorf("error: %s", err)
			}
		}
	}
}

func TestListResourceACL(t *testing.T) {
	expect := make(chan interface{})
	for _, resource := range resourcePatterns {
		cmd := NewCMD(expect)
		cmd.SetArgs(append([]string{"acl", "list"}, resource.args...))

		go func() {
			expect <- ConvertToFilter(&kafka.KafkaAPIACLRequest{Pattern: resource.pattern, Entry: &kafka.AccessControlEntryConfig{}})
		}()

		if err := cmd.Execute(); err != nil {
			t.Errorf("error: %s", err)
		}
	}
}

func TestListPrincipalACL(t *testing.T) {
	expect := make(chan interface{})
	for _, entry := range aclEntries {
		cmd := NewCMD(expect)
		cmd.SetArgs(append([]string{"acl", "list", "--principal"}, strings.TrimPrefix(entry.entry.Principal, "user:")))

		go func() {
			expect <- ConvertToFilter(&kafka.KafkaAPIACLRequest{Entry: &kafka.AccessControlEntryConfig{Principal: entry.entry.Principal}})
		}()

		if err := cmd.Execute(); err != nil {
			t.Errorf("error: %s", err)
		}
	}
}

/*************** TEST command_topic ***************/
var Topics = []struct {
	args []string
	spec *kafka.KafkaTopicSpecification
}{
	{
		args: []string{"test_topic", "--partitions", strconv.Itoa(1), "--replication-factor", strconv.Itoa(2), "--config", "a=b"},
		spec: &kafka.KafkaTopicSpecification{Name: "test_topic", ReplicationFactor: 2, NumPartitions: 1, Configs: map[string]string{"a": "b"}},
	},
}

func TestListTopic(t *testing.T) {
	expect := make(chan interface{})
	for _, topic := range Topics {
		cmd := NewCMD(expect)
		cmd.SetArgs([]string{"topic", "list"})

		go func() {
			expect <- &kafka.KafkaAPITopicRequest{Spec: &kafka.KafkaTopicSpecification{Name: topic.spec.Name}}
		}()

		if err := cmd.Execute(); err != nil {
			t.Logf("error: %s", err)
			t.Error()
			return
		}
	}
}

func TestCreateTopic(t *testing.T) {
	expect := make(chan interface{})
	for _, topic := range Topics {
		cmd := NewCMD(expect)
		cmd.SetArgs(append([]string{"topic", "create"}, topic.args...))

		go func() {
			expect <- &kafka.KafkaAPITopicRequest{Spec: topic.spec}
		}()

		if err := cmd.Execute(); err != nil {
			t.Logf("error: %s", err)
			t.Error()
			return
		}
	}
}

func TestDescribeTopic(t *testing.T) {
	expect := make(chan interface{})
	for _, topic := range Topics {
		cmd := NewCMD(expect)
		cmd.SetArgs(append([]string{"topic", "describe"}, topic.args[0]))

		go func() {
			expect <- &kafka.KafkaAPITopicRequest{Spec: &kafka.KafkaTopicSpecification{Name: topic.spec.Name}}
		}()

		if err := cmd.Execute(); err != nil {
			t.Logf("error: %s", err)
			t.Error()
			return
		}
	}
}

func TestDeleteTopic(t *testing.T) {
	expect := make(chan interface{})
	for _, topic := range Topics {
		cmd := NewCMD(expect)
		cmd.SetArgs(append([]string{"topic", "delete"}, topic.args[0]))

		go func() {
			expect <- &kafka.KafkaAPITopicRequest{Spec: &kafka.KafkaTopicSpecification{Name: topic.spec.Name}}
		}()

		if err := cmd.Execute(); err != nil {
			t.Logf("error: %s", err)
			t.Error()
			return
		}
	}
}

func TestUpdateTopic(t *testing.T) {
	expect := make(chan interface{})
	for _, topic := range Topics {
		cmd := NewCMD(expect)
		cmd.SetArgs(append([]string{"topic", "update"}, topic.args[0]))

		go func() {
			expect <- &kafka.KafkaAPITopicRequest{Spec: &kafka.KafkaTopicSpecification{Name: topic.spec.Name, Configs: topic.spec.Configs}}
		}()

		if err := cmd.Execute(); err != nil {
			t.Logf("error: %s", err)
			t.Error()
			return
		}
	}
}

/*************** TEST setup/helpers ***************/
func NewCMD(expect chan interface{}) *cobra.Command {
	cmd, _ := NewKafkaCommand(conf, func(value interface{}) error {
		return NewPluginMock(value, expect)
	})
	return cmd
}

func init() {
	conf = shared.NewConfig()
	conf.Auth = &shared.AuthConfig{
		User:    new(v1.User),
		Account: &v1.Account{Id: "testAccount"},
	}
}

type kafkaPluginMock struct {
	Expect chan interface{}
}

func NewPluginMock(value interface{}, expect chan interface{}) error {
	client := &kafkaPluginMock{expect}
	rv := reflect.ValueOf(value)
	rv.Elem().Set(reflect.ValueOf(client))
	return nil
}

func (m *kafkaPluginMock) CreateAPIKey(_ context.Context, apiKey *schedv1.ApiKey) (*schedv1.ApiKey, error) {
	return apiKey, nil
}

func (m *kafkaPluginMock) List(_ context.Context, cluster *schedv1.KafkaCluster) ([]*schedv1.KafkaCluster, error) {
	return []*schedv1.KafkaCluster{cluster}, nil
}

func (m *kafkaPluginMock) Describe(_ context.Context, cluster *schedv1.KafkaCluster) (*schedv1.KafkaCluster, error) {
	return cluster, nil
}

func (m *kafkaPluginMock) Create(_ context.Context, config *schedv1.KafkaClusterConfig) (*schedv1.KafkaCluster, error) {
	return &schedv1.KafkaCluster{}, nil
}

func (m *kafkaPluginMock) Delete(_ context.Context, cluster *schedv1.KafkaCluster) error {
	return nil
}

func (m *kafkaPluginMock) ListTopic(_ context.Context) (*kafka.ListKafkaTopicReply, error) {
	return &kafka.ListKafkaTopicReply{Topics: []string{"test1", "test2", "test3"}}, nil
}

func (m *kafkaPluginMock) DescribeTopic(_ context.Context, actual *kafka.KafkaAPITopicRequest) (*kafka.KafkaTopicDescription, error) {
	return &kafka.KafkaTopicDescription{}, assertEquals(actual, <-m.Expect)
}

func (m *kafkaPluginMock) CreateTopic(_ context.Context, actual *kafka.KafkaAPITopicRequest) (*kafka.KafkaAPIResponse, error) {
	return &kafka.KafkaAPIResponse{}, assertEquals(actual, <-m.Expect)
}

func (m *kafkaPluginMock) DeleteTopic(_ context.Context, actual *kafka.KafkaAPITopicRequest) (*kafka.KafkaAPIResponse, error) {
	return &kafka.KafkaAPIResponse{}, assertEquals(actual, <-m.Expect)
}

func (m *kafkaPluginMock) UpdateTopic(_ context.Context, actual *kafka.KafkaAPITopicRequest) (*kafka.KafkaAPIResponse, error) {
	return &kafka.KafkaAPIResponse{}, assertEquals(actual, <-m.Expect)
}

func (m *kafkaPluginMock) ListACL(_ context.Context, actual *kafka.KafkaAPIACLFilterRequest) (*kafka.KafkaAPIACLFilterReply, error) {
	return &kafka.KafkaAPIACLFilterReply{}, assertEquals(actual, <-m.Expect)
}

func (m *kafkaPluginMock) CreateACL(_ context.Context, actual *kafka.KafkaAPIACLRequest) (*kafka.KafkaAPIResponse, error) {
	return nil, assertEquals(actual, <-m.Expect)
}

func (m *kafkaPluginMock) DeleteACL(_ context.Context, actual *kafka.KafkaAPIACLFilterRequest) (*kafka.KafkaAPIResponse, error) {
	return &kafka.KafkaAPIResponse{}, assertEquals(actual, <-m.Expect)
}

func assertEquals(actual interface{}, expected interface{}) error {
	if !reflect.DeepEqual(actual, expected) {
		return fmt.Errorf("actual: %+v\nexpected: %+v", actual, expected)
	}
	return nil
}
