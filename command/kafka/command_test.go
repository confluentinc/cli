package kafka

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	kafkav1 "github.com/confluentinc/ccloudapis/kafka/v1"
	orgv1 "github.com/confluentinc/ccloudapis/org/v1"
	cliMock "github.com/confluentinc/cli/mock"

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
		args: []string{"--topic", "test-topic", "--prefix"},
		pattern: &kafkav1.ResourcePatternConfig{ResourceType: kafkav1.ResourceTypes_TOPIC, Name: "test-topic",
			PatternType: kafkav1.PatternTypes_PREFIXED},
	},
	{
		args: []string{"--consumer-group", "test-group"},
		pattern: &kafkav1.ResourcePatternConfig{ResourceType: kafkav1.ResourceTypes_GROUP, Name: "test-group",
			PatternType: kafkav1.PatternTypes_LITERAL},
	},
	{
		args: []string{"--consumer-group", "test-group", "--prefix"},
		pattern: &kafkav1.ResourcePatternConfig{ResourceType: kafkav1.ResourceTypes_GROUP, Name: "test-group",
			PatternType: kafkav1.PatternTypes_PREFIXED},
	},
	{
		args: []string{"--transactional-id", "test-transactional-id"},
		pattern: &kafkav1.ResourcePatternConfig{ResourceType: kafkav1.ResourceTypes_TRANSACTIONAL_ID, Name: "test-transactional-id",
			PatternType: kafkav1.PatternTypes_LITERAL},
	},
	{
		args: []string{"--transactional-id", "test-transactional-id", "--prefix"},
		pattern: &kafkav1.ResourcePatternConfig{ResourceType: kafkav1.ResourceTypes_TRANSACTIONAL_ID, Name: "test-transactional-id",
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
			Principal: "User:test_user", Operation: kafkav1.ACLOperations_READ, Host: "*"},
	},
	{
		args: []string{"--deny", "--principal", "test_user", "--operation", "read"},
		entry: &kafkav1.AccessControlEntryConfig{PermissionType: kafkav1.ACLPermissionTypes_DENY,
			Principal: "User:test_user", Operation: kafkav1.ACLOperations_READ, Host: "*"},
	},
	{
		args: []string{"--allow", "--principal", "test_user", "--operation", "write"},
		entry: &kafkav1.AccessControlEntryConfig{PermissionType: kafkav1.ACLPermissionTypes_ALLOW,
			Principal: "User:test_user", Operation: kafkav1.ACLOperations_WRITE, Host: "*"},
	},
	{
		args: []string{"--deny", "--principal", "test_user", "--operation", "write"},
		entry: &kafkav1.AccessControlEntryConfig{PermissionType: kafkav1.ACLPermissionTypes_DENY,
			Principal: "User:test_user", Operation: kafkav1.ACLOperations_WRITE, Host: "*"},
	},
	{
		args: []string{"--allow", "--principal", "test_user", "--operation", "create"},
		entry: &kafkav1.AccessControlEntryConfig{PermissionType: kafkav1.ACLPermissionTypes_ALLOW,
			Principal: "User:test_user", Operation: kafkav1.ACLOperations_CREATE, Host: "*"},
	},
	{
		args: []string{"--deny", "--principal", "test_user", "--operation", "create"},
		entry: &kafkav1.AccessControlEntryConfig{PermissionType: kafkav1.ACLPermissionTypes_DENY,
			Principal: "User:test_user", Operation: kafkav1.ACLOperations_CREATE, Host: "*"},
	},
	{
		args: []string{"--allow", "--principal", "test_user", "--operation", "delete"},
		entry: &kafkav1.AccessControlEntryConfig{PermissionType: kafkav1.ACLPermissionTypes_ALLOW,
			Principal: "User:test_user", Operation: kafkav1.ACLOperations_DELETE, Host: "*"},
	},
	{
		args: []string{"--deny", "--principal", "test_user", "--operation", "delete"},
		entry: &kafkav1.AccessControlEntryConfig{PermissionType: kafkav1.ACLPermissionTypes_DENY,
			Principal: "User:test_user", Operation: kafkav1.ACLOperations_DELETE, Host: "*"},
	},
	{
		args: []string{"--allow", "--principal", "test_user", "--operation", "alter"},
		entry: &kafkav1.AccessControlEntryConfig{PermissionType: kafkav1.ACLPermissionTypes_ALLOW,
			Principal: "User:test_user", Operation: kafkav1.ACLOperations_ALTER, Host: "*"},
	},
	{
		args: []string{"--deny", "--principal", "test_user", "--operation", "alter"},
		entry: &kafkav1.AccessControlEntryConfig{PermissionType: kafkav1.ACLPermissionTypes_DENY,
			Principal: "User:test_user", Operation: kafkav1.ACLOperations_ALTER, Host: "*"},
	},
	{
		args: []string{"--allow", "--principal", "test_user", "--operation", "describe"},
		entry: &kafkav1.AccessControlEntryConfig{PermissionType: kafkav1.ACLPermissionTypes_ALLOW,
			Principal: "User:test_user", Operation: kafkav1.ACLOperations_DESCRIBE, Host: "*"},
	},
	{
		args: []string{"--deny", "--principal", "test_user", "--operation", "describe"},
		entry: &kafkav1.AccessControlEntryConfig{PermissionType: kafkav1.ACLPermissionTypes_DENY,
			Principal: "User:test_user", Operation: kafkav1.ACLOperations_DESCRIBE, Host: "*"},
	},
	{
		args: []string{"--allow", "--principal", "test_user", "--operation", "cluster_action"},
		entry: &kafkav1.AccessControlEntryConfig{PermissionType: kafkav1.ACLPermissionTypes_ALLOW,
			Principal: "User:test_user", Operation: kafkav1.ACLOperations_CLUSTER_ACTION, Host: "*"},
	},
	{
		args: []string{"--deny", "--principal", "test_user", "--operation", "cluster_action"},
		entry: &kafkav1.AccessControlEntryConfig{PermissionType: kafkav1.ACLPermissionTypes_DENY,
			Principal: "User:test_user", Operation: kafkav1.ACLOperations_CLUSTER_ACTION, Host: "*"},
	},
	{
		args: []string{"--allow", "--principal", "test_user", "--operation", "describe_configs"},
		entry: &kafkav1.AccessControlEntryConfig{PermissionType: kafkav1.ACLPermissionTypes_ALLOW,
			Principal: "User:test_user", Operation: kafkav1.ACLOperations_DESCRIBE_CONFIGS, Host: "*"},
	},
	{
		args: []string{"--deny", "--principal", "test_user", "--operation", "describe_configs"},
		entry: &kafkav1.AccessControlEntryConfig{PermissionType: kafkav1.ACLPermissionTypes_DENY,
			Principal: "User:test_user", Operation: kafkav1.ACLOperations_DESCRIBE_CONFIGS, Host: "*"},
	},
	{
		args: []string{"--allow", "--principal", "test_user", "--operation", "alter_configs"},
		entry: &kafkav1.AccessControlEntryConfig{PermissionType: kafkav1.ACLPermissionTypes_ALLOW,
			Principal: "User:test_user", Operation: kafkav1.ACLOperations_ALTER_CONFIGS, Host: "*"},
	},
	{
		args: []string{"--deny", "--principal", "test_user", "--operation", "alter_configs"},
		entry: &kafkav1.AccessControlEntryConfig{PermissionType: kafkav1.ACLPermissionTypes_DENY,
			Principal: "User:test_user", Operation: kafkav1.ACLOperations_ALTER_CONFIGS, Host: "*"},
	},
	{
		args: []string{"--allow", "--principal", "test_user", "--operation", "idempotent_write"},
		entry: &kafkav1.AccessControlEntryConfig{PermissionType: kafkav1.ACLPermissionTypes_ALLOW,
			Principal: "User:test_user", Operation: kafkav1.ACLOperations_IDEMPOTENT_WRITE, Host: "*"},
	},
	{
		args: []string{"--deny", "--principal", "test_user", "--operation", "idempotent_write"},
		entry: &kafkav1.AccessControlEntryConfig{PermissionType: kafkav1.ACLPermissionTypes_DENY,
			Principal: "User:test_user", Operation: kafkav1.ACLOperations_IDEMPOTENT_WRITE, Host: "*"},
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
		cmd.SetArgs(append([]string{"acl", "list", "--principal"}, strings.TrimPrefix(entry.entry.Principal, "User:")))

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
		args: []string{"test_topic", "--config", "a=b", "--partitions", strconv.Itoa(1), "--replication-factor", strconv.Itoa(2)},
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
			t.Errorf("error: %s", err)
			t.Fail()
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
			t.Errorf("error: %s", err)
			t.Fail()
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
			t.Errorf("error: %s", err)
			t.Fail()
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
			t.Errorf("error: %s", err)
			t.Fail()
			return
		}
	}
}

func TestUpdateTopic(t *testing.T) {
	expect := make(chan interface{})
	for _, topic := range Topics {
		cmd := NewCMD(expect)
		cmd.SetArgs(append([]string{"topic", "update"}, topic.args[0:3]...))
		go func() {
			expect <- &kafkav1.Topic{Spec: &kafkav1.TopicSpecification{Name: topic.spec.Name, Configs: topic.spec.Configs}}
		}()

		if err := cmd.Execute(); err != nil {
			t.Errorf("error: %s", err)
			t.Fail()
			return
		}
	}
}

/*************** TEST setup/helpers ***************/
func NewCMD(expect chan interface{}) *cobra.Command {
	cmd, _ := NewKafkaCommand(conf, &cliMock.GRPCPlugin {
		LookupPathFunc: func() (string, error) {
			return "", nil
		},
		LoadFunc: func(value interface{}) error {
			return cliMock.NewKafkaMock(value, expect)
		},
	})

	return cmd
}

func init() {
	conf = shared.NewConfig()
	conf.AuthURL = "http://test"
	conf.Auth = &shared.AuthConfig{
		User:    new(orgv1.User),
		Account: &orgv1.Account{Id: "testAccount"},
	}
	initContext(conf)
}

// initContext mimics logging in with a configured context
// TODO: create auth mock
func initContext(config *shared.Config) {
	user := config.Auth
	name := fmt.Sprintf("login-%s-%s", user.User.Email, config.AuthURL)

	config.Platforms[name] = &shared.Platform{
		Server:        config.AuthURL,
		KafkaClusters: map[string]shared.KafkaClusterConfig{name: {}},
	}

	config.Credentials[name] = &shared.Credential{
		Username: user.User.Email,
	}

	config.Contexts[name] = &shared.Context{
		Platform:   name,
		Credential: name,
		Kafka:      name,
	}

	config.CurrentContext = name
}
