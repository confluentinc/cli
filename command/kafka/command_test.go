package kafka

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"github.com/spf13/cobra"

	authv1 "github.com/confluentinc/ccloudapis/auth/v1"
	kafkav1 "github.com/confluentinc/ccloudapis/kafka/v1"
	orgv1 "github.com/confluentinc/ccloudapis/org/v1"
	chttp "github.com/confluentinc/ccloud-sdk-go"
	"github.com/confluentinc/cli/shared"
)

var conf *shared.Config

/*************** TEST command_acl ***************/
var resourcePatterns = []struct {
	args    []string
	pattern *kafkav1.ResourcePatternConfig
}{
	{
		args: []string{"--cluster"},
		pattern: &kafkav1.ResourcePatternConfig{ResourceType: kafkav1.ResourceTypes_CLUSTER, Name: "kafka-cluster",
			PatternType: kafkav1.PatternTypes_LITERAL},
	},
	{
		args: []string{"--topic", "test-topic"},
		pattern: &kafkav1.ResourcePatternConfig{ResourceType: kafkav1.ResourceTypes_TOPIC, Name: "test-topic",
			PatternType: kafkav1.PatternTypes_LITERAL},
	},
	{
		args: []string{"--topic", "test-topic*", "--pattern-type", "prefixed"},
		pattern: &kafkav1.ResourcePatternConfig{ResourceType: kafkav1.ResourceTypes_TOPIC, Name: "test-topic",
			PatternType: kafkav1.PatternTypes_PREFIXED},
	},
	{
		args: []string{"--consumer_group", "test-group"},
		pattern: &kafkav1.ResourcePatternConfig{ResourceType: kafkav1.ResourceTypes_GROUP, Name: "test-group",
			PatternType: kafkav1.PatternTypes_LITERAL},
	},
	{
		args: []string{"--consumer_group", "test-group*",  "--pattern-type", "prefixed"},
		pattern: &kafkav1.ResourcePatternConfig{ResourceType: kafkav1.ResourceTypes_GROUP, Name: "test-group",
			PatternType: kafkav1.PatternTypes_PREFIXED},
	},
	{
		args: []string{"--transactional_id", "test-transactional_id"},
		pattern: &kafkav1.ResourcePatternConfig{ResourceType: kafkav1.ResourceTypes_TRANSACTIONAL_ID, Name: "test-transactional_id",
			PatternType: kafkav1.PatternTypes_LITERAL},
	},
	{
		args: []string{"--transactional_id", "test-transactional_id*",  "--pattern-type", "prefixed"},
		pattern: &kafkav1.ResourcePatternConfig{ResourceType: kafkav1.ResourceTypes_TRANSACTIONAL_ID, Name: "test-transactional_id",
			PatternType: kafkav1.PatternTypes_PREFIXED},
	},
}

var aclEntries = []struct {
	args  []string
	entry *kafkav1.AccessControlEntryConfig
}{
	{
		args: []string{"--allow", "--principal", "test_user", "--operation", "read"},
		entry: &kafkav1.AccessControlEntryConfig{PermissionType: kafkav1.ACLPermissionTypes_ALLOW,
			Principal: "user:test_user", Operation: kafkav1.ACLOperations_READ, Host: "*"},
	},
	{
		args: []string{"--deny", "--principal", "test_user", "--operation", "read"},
		entry: &kafkav1.AccessControlEntryConfig{PermissionType: kafkav1.ACLPermissionTypes_DENY,
			Principal: "user:test_user", Operation: kafkav1.ACLOperations_READ, Host: "*"},
	},
	{
		args: []string{"--allow", "--principal", "test_user", "--operation", "write"},
		entry: &kafkav1.AccessControlEntryConfig{PermissionType: kafkav1.ACLPermissionTypes_ALLOW,
			Principal: "user:test_user", Operation: kafkav1.ACLOperations_WRITE, Host: "*"},
	},
	{
		args: []string{"--deny", "--principal", "test_user", "--operation", "write"},
		entry: &kafkav1.AccessControlEntryConfig{PermissionType: kafkav1.ACLPermissionTypes_DENY,
			Principal: "user:test_user", Operation: kafkav1.ACLOperations_WRITE, Host: "*"},
	},
	{
		args: []string{"--allow", "--principal", "test_user", "--operation", "create"},
		entry: &kafkav1.AccessControlEntryConfig{PermissionType: kafkav1.ACLPermissionTypes_ALLOW,
			Principal: "user:test_user", Operation: kafkav1.ACLOperations_CREATE, Host: "*"},
	},
	{
		args: []string{"--deny", "--principal", "test_user", "--operation", "create"},
		entry: &kafkav1.AccessControlEntryConfig{PermissionType: kafkav1.ACLPermissionTypes_DENY,
			Principal: "user:test_user", Operation: kafkav1.ACLOperations_WRITE, Host: "*"},
	},
	{
		args: []string{"--allow", "--principal", "test_user", "--operation", "delete"},
		entry: &kafkav1.AccessControlEntryConfig{PermissionType: kafkav1.ACLPermissionTypes_ALLOW,
			Principal: "user:test_user", Operation: kafkav1.ACLOperations_DELETE, Host: "*"},
	},
	{
		args: []string{"--deny", "--principal", "test_user", "--operation", "delete"},
		entry: &kafkav1.AccessControlEntryConfig{PermissionType: kafkav1.ACLPermissionTypes_DENY,
			Principal: "user:test_user", Operation: kafkav1.ACLOperations_DELETE, Host: "*"},
	},
	{
		args: []string{"--allow", "--principal", "test_user", "--operation", "alter"},
		entry: &kafkav1.AccessControlEntryConfig{PermissionType: kafkav1.ACLPermissionTypes_ALLOW,
			Principal: "user:test_user", Operation: kafkav1.ACLOperations_ALTER, Host: "*"},
	},
	{
		args: []string{"--deny", "--principal", "test_user", "--operation", "alter"},
		entry: &kafkav1.AccessControlEntryConfig{PermissionType: kafkav1.ACLPermissionTypes_DENY,
			Principal: "user:test_user", Operation: kafkav1.ACLOperations_ALTER, Host: "*"},
	},
	{
		args: []string{"--allow", "--principal", "test_user", "--operation", "describe"},
		entry: &kafkav1.AccessControlEntryConfig{PermissionType: kafkav1.ACLPermissionTypes_ALLOW,
			Principal: "user:test_user", Operation: kafkav1.ACLOperations_DESCRIBE, Host: "*"},
	},
	{
		args: []string{"--deny", "--principal", "test_user", "--operation", "describe"},
		entry: &kafkav1.AccessControlEntryConfig{PermissionType: kafkav1.ACLPermissionTypes_DENY,
			Principal: "user:test_user", Operation: kafkav1.ACLOperations_DESCRIBE, Host: "*"},
	},
	{
		args: []string{"--allow", "--principal", "test_user", "--operation", "cluster_action"},
		entry: &kafkav1.AccessControlEntryConfig{PermissionType: kafkav1.ACLPermissionTypes_ALLOW,
			Principal: "user:test_user", Operation: kafkav1.ACLOperations_CLUSTER_ACTION, Host: "*"},
	},
	{
		args: []string{"--deny", "--principal", "test_user", "--operation", "cluster_action"},
		entry: &kafkav1.AccessControlEntryConfig{PermissionType: kafkav1.ACLPermissionTypes_DENY,
			Principal: "user:test_user", Operation: kafkav1.ACLOperations_CLUSTER_ACTION, Host: "*"},
	},
	{
		args: []string{"--allow", "--principal", "test_user", "--operation", "describe_configs"},
		entry: &kafkav1.AccessControlEntryConfig{PermissionType: kafkav1.ACLPermissionTypes_ALLOW,
			Principal: "user:test_user", Operation: kafkav1.ACLOperations_DESCRIBE_CONFIGS, Host: "*"},
	},
	{
		args: []string{"--deny", "--principal", "test_user", "--operation", "describe_configs"},
		entry: &kafkav1.AccessControlEntryConfig{PermissionType: kafkav1.ACLPermissionTypes_DENY,
			Principal: "user:test_user", Operation: kafkav1.ACLOperations_DESCRIBE_CONFIGS, Host: "*"},
	},
	{
		args: []string{"--allow", "--principal", "test_user", "--operation", "alter_configs"},
		entry: &kafkav1.AccessControlEntryConfig{PermissionType: kafkav1.ACLPermissionTypes_ALLOW,
			Principal: "user:test_user", Operation: kafkav1.ACLOperations_ALTER_CONFIGS, Host: "*"},
	},
	{
		args: []string{"--deny", "--principal", "test_user", "--operation", "alter_configs"},
		entry: &kafkav1.AccessControlEntryConfig{PermissionType: kafkav1.ACLPermissionTypes_DENY,
			Principal: "user:test_user", Operation: kafkav1.ACLOperations_ALTER_CONFIGS, Host: "*"},
	},
	{
		args: []string{"--allow", "--principal", "test_user", "--operation", "idempotent_write"},
		entry: &kafkav1.AccessControlEntryConfig{PermissionType: kafkav1.ACLPermissionTypes_ALLOW,
			Principal: "user:test_user", Operation: kafkav1.ACLOperations_IDEMPOTENT_WRITE, Host: "*"},
	},
	{
		args: []string{"--deny", "--principal", "test_user", "--operation", "idempotent_write"},
		entry: &kafkav1.AccessControlEntryConfig{PermissionType: kafkav1.ACLPermissionTypes_DENY,
			Principal: "user:test_user", Operation: kafkav1.ACLOperations_IDEMPOTENT_WRITE, Host: "*"},
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
				expect <- &kafkav1.ACLBinding{Pattern: resource.pattern, Entry: entry.entry}
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
				expect <- convertToFilter(&kafkav1.ACLBinding{Pattern: resource.pattern, Entry: entry.entry})
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
			expect <- convertToFilter(&kafkav1.ACLBinding{Pattern: resource.pattern, Entry: &kafkav1.AccessControlEntryConfig{}})
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
			expect <- convertToFilter(&kafkav1.ACLBinding{Entry: &kafkav1.AccessControlEntryConfig{Principal: entry.entry.Principal}})
		}()

		if err := cmd.Execute(); err != nil {
			t.Errorf("error: %s", err)
		}
	}
}

/*************** TEST command_topic ***************/
var Topics = []struct {
	args []string
	spec *kafkav1.TopicSpecification
}{
	{
		args: []string{"test_topic", "--partitions", strconv.Itoa(1), "--replication-factor", strconv.Itoa(2), "--config", "a=b"},
		spec: &kafkav1.TopicSpecification{Name: "test_topic", ReplicationFactor: 2, NumPartitions: 1, Configs: map[string]string{"a": "b"}},
	},
}

func TestListTopics(t *testing.T) {
	expect := make(chan interface{})
	for _, topic := range Topics {
		cmd := NewCMD(expect)
		cmd.SetArgs([]string{"topic", "list"})

		go func() {
			expect <- &kafkav1.Topic{Spec: &kafkav1.TopicSpecification{Name: topic.spec.Name}}
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
			expect <- &kafkav1.Topic{Spec: topic.spec}
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
			expect <- &kafkav1.Topic{Spec: &kafkav1.TopicSpecification{Name: topic.spec.Name}}
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
			expect <- &kafkav1.Topic{Spec: &kafkav1.TopicSpecification{Name: topic.spec.Name}}
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
			expect <- &kafkav1.Topic{Spec: &kafkav1.TopicSpecification{Name: topic.spec.Name, Configs: topic.spec.Configs}}
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
		User:    new(orgv1.User),
		Account: &orgv1.Account{Id: "testAccount"},
	}
}

// Compile-time check interface adherence
var _ chttp.Kafka = (*kafkaPluginMock)(nil)

type kafkaPluginMock struct {
	Expect chan interface{}
}

func NewPluginMock(value interface{}, expect chan interface{}) error {
	client := &kafkaPluginMock{expect}
	rv := reflect.ValueOf(value)
	rv.Elem().Set(reflect.ValueOf(client))
	return nil
}

func (m *kafkaPluginMock) CreateAPIKey(_ context.Context, apiKey *authv1.APIKey) (*authv1.APIKey, error) {
	return apiKey, nil
}

func (m *kafkaPluginMock) List(_ context.Context, cluster *kafkav1.Cluster) ([]*kafkav1.Cluster, error) {
	return []*kafkav1.Cluster{cluster}, nil
}

func (m *kafkaPluginMock) Describe(_ context.Context, cluster *kafkav1.Cluster) (*kafkav1.Cluster, error) {
	return cluster, nil
}

func (m *kafkaPluginMock) Create(_ context.Context, config *kafkav1.ClusterConfig) (*kafkav1.Cluster, error) {
	return &kafkav1.Cluster{}, nil
}

func (m *kafkaPluginMock) Delete(_ context.Context, cluster *kafkav1.Cluster) error {
	return nil
}

func (m *kafkaPluginMock) ListTopics(ctx context.Context, cluster *kafkav1.Cluster) ([]*kafkav1.TopicDescription, error) {
	return []*kafkav1.TopicDescription{
		{Name:"test1"},
		{Name:"test2"},
		{Name:"test3"}}, nil
}

func (m *kafkaPluginMock) DescribeTopic(ctx context.Context, cluster *kafkav1.Cluster, topic *kafkav1.Topic) (*kafkav1.TopicDescription, error) {
	return &kafkav1.TopicDescription{}, assertEquals(topic, <-m.Expect)
}*
func (m *kafkaPluginMock) DeleteACL(ctx context.Context, cluster *kafkav1.Cluster, filter *kafkav1.ACLFilter) error {
	return assertEquals(filter, <-m.Expect)
}

func assertEquals(actual interface{}, expected interface{}) error {
	if !reflect.DeepEqual(actual, expected) {
		return fmt.Errorf("actual: %+v\nexpected: %+v", actual, expected)
	}
	return nil
}
