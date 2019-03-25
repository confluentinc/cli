package kafka

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/confluentinc/cli/log"
	"github.com/spf13/cobra"

	"github.com/confluentinc/ccloud-sdk-go/mock"
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
	{
		args: []string{"--prefix", "--topic", "test-topic"},
		pattern: &kafkav1.ResourcePatternConfig{ResourceType: kafkav1.ResourceTypes_TOPIC, Name: "test-topic",
			PatternType: kafkav1.PatternTypes_PREFIXED},
	},
}

var aclEntries = []struct {
	args  []string
	entry *kafkav1.AccessControlEntryConfig
}{
	{
		args: []string{"--allow", "--service-account-id", "42", "--operation", "read"},
		entry: &kafkav1.AccessControlEntryConfig{PermissionType: kafkav1.ACLPermissionTypes_ALLOW,
			Principal: "User:42", Operation: kafkav1.ACLOperations_READ, Host: "*"},
	},
	{
		args: []string{"--deny", "--service-account-id", "42", "--operation", "read"},
		entry: &kafkav1.AccessControlEntryConfig{PermissionType: kafkav1.ACLPermissionTypes_DENY,
			Principal: "User:42", Operation: kafkav1.ACLOperations_READ, Host: "*"},
	},
	{
		args: []string{"--allow", "--service-account-id", "42", "--operation", "write"},
		entry: &kafkav1.AccessControlEntryConfig{PermissionType: kafkav1.ACLPermissionTypes_ALLOW,
			Principal: "User:42", Operation: kafkav1.ACLOperations_WRITE, Host: "*"},
	},
	{
		args: []string{"--deny", "--service-account-id", "42", "--operation", "write"},
		entry: &kafkav1.AccessControlEntryConfig{PermissionType: kafkav1.ACLPermissionTypes_DENY,
			Principal: "User:42", Operation: kafkav1.ACLOperations_WRITE, Host: "*"},
	},
	{
		args: []string{"--allow", "--service-account-id", "42", "--operation", "create"},
		entry: &kafkav1.AccessControlEntryConfig{PermissionType: kafkav1.ACLPermissionTypes_ALLOW,
			Principal: "User:42", Operation: kafkav1.ACLOperations_CREATE, Host: "*"},
	},
	{
		args: []string{"--deny", "--service-account-id", "42", "--operation", "create"},
		entry: &kafkav1.AccessControlEntryConfig{PermissionType: kafkav1.ACLPermissionTypes_DENY,
			Principal: "User:42", Operation: kafkav1.ACLOperations_CREATE, Host: "*"},
	},
	{
		args: []string{"--allow", "--service-account-id", "42", "--operation", "delete"},
		entry: &kafkav1.AccessControlEntryConfig{PermissionType: kafkav1.ACLPermissionTypes_ALLOW,
			Principal: "User:42", Operation: kafkav1.ACLOperations_DELETE, Host: "*"},
	},
	{
		args: []string{"--deny", "--service-account-id", "42", "--operation", "delete"},
		entry: &kafkav1.AccessControlEntryConfig{PermissionType: kafkav1.ACLPermissionTypes_DENY,
			Principal: "User:42", Operation: kafkav1.ACLOperations_DELETE, Host: "*"},
	},
	{
		args: []string{"--allow", "--service-account-id", "42", "--operation", "alter"},
		entry: &kafkav1.AccessControlEntryConfig{PermissionType: kafkav1.ACLPermissionTypes_ALLOW,
			Principal: "User:42", Operation: kafkav1.ACLOperations_ALTER, Host: "*"},
	},
	{
		args: []string{"--deny", "--service-account-id", "42", "--operation", "alter"},
		entry: &kafkav1.AccessControlEntryConfig{PermissionType: kafkav1.ACLPermissionTypes_DENY,
			Principal: "User:42", Operation: kafkav1.ACLOperations_ALTER, Host: "*"},
	},
	{
		args: []string{"--allow", "--service-account-id", "42", "--operation", "describe"},
		entry: &kafkav1.AccessControlEntryConfig{PermissionType: kafkav1.ACLPermissionTypes_ALLOW,
			Principal: "User:42", Operation: kafkav1.ACLOperations_DESCRIBE, Host: "*"},
	},
	{
		args: []string{"--deny", "--service-account-id", "42", "--operation", "describe"},
		entry: &kafkav1.AccessControlEntryConfig{PermissionType: kafkav1.ACLPermissionTypes_DENY,
			Principal: "User:42", Operation: kafkav1.ACLOperations_DESCRIBE, Host: "*"},
	},
	{
		args: []string{"--allow", "--service-account-id", "42", "--operation", "cluster-action"},
		entry: &kafkav1.AccessControlEntryConfig{PermissionType: kafkav1.ACLPermissionTypes_ALLOW,
			Principal: "User:42", Operation: kafkav1.ACLOperations_CLUSTER_ACTION, Host: "*"},
	},
	{
		args: []string{"--deny", "--service-account-id", "42", "--operation", "cluster-action"},
		entry: &kafkav1.AccessControlEntryConfig{PermissionType: kafkav1.ACLPermissionTypes_DENY,
			Principal: "User:42", Operation: kafkav1.ACLOperations_CLUSTER_ACTION, Host: "*"},
	},
	{
		args: []string{"--allow", "--service-account-id", "42", "--operation", "describe-configs"},
		entry: &kafkav1.AccessControlEntryConfig{PermissionType: kafkav1.ACLPermissionTypes_ALLOW,
			Principal: "User:42", Operation: kafkav1.ACLOperations_DESCRIBE_CONFIGS, Host: "*"},
	},
	{
		args: []string{"--deny", "--service-account-id", "42", "--operation", "describe-configs"},
		entry: &kafkav1.AccessControlEntryConfig{PermissionType: kafkav1.ACLPermissionTypes_DENY,
			Principal: "User:42", Operation: kafkav1.ACLOperations_DESCRIBE_CONFIGS, Host: "*"},
	},
	{
		args: []string{"--allow", "--service-account-id", "42", "--operation", "alter-configs"},
		entry: &kafkav1.AccessControlEntryConfig{PermissionType: kafkav1.ACLPermissionTypes_ALLOW,
			Principal: "User:42", Operation: kafkav1.ACLOperations_ALTER_CONFIGS, Host: "*"},
	},
	{
		args: []string{"--deny", "--service-account-id", "42", "--operation", "alter-configs"},
		entry: &kafkav1.AccessControlEntryConfig{PermissionType: kafkav1.ACLPermissionTypes_DENY,
			Principal: "User:42", Operation: kafkav1.ACLOperations_ALTER_CONFIGS, Host: "*"},
	},
	{
		args: []string{"--allow", "--service-account-id", "42", "--operation", "idempotent-write"},
		entry: &kafkav1.AccessControlEntryConfig{PermissionType: kafkav1.ACLPermissionTypes_ALLOW,
			Principal: "User:42", Operation: kafkav1.ACLOperations_IDEMPOTENT_WRITE, Host: "*"},
	},
	{
		args: []string{"--deny", "--service-account-id", "42", "--operation", "idempotent-write"},
		entry: &kafkav1.AccessControlEntryConfig{PermissionType: kafkav1.ACLPermissionTypes_DENY,
			Principal: "User:42", Operation: kafkav1.ACLOperations_IDEMPOTENT_WRITE, Host: "*"},
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
		cmd.SetArgs(append([]string{"acl", "list", "--service-account-id"}, strings.TrimPrefix(entry.entry.Principal, "User:")))

		go func() {
			expect <- convertToFilter(&kafkav1.ACLBinding{Entry: &kafkav1.AccessControlEntryConfig{Principal: entry.entry.Principal}})
		}()

		if err := cmd.Execute(); err != nil {
			t.Errorf("error: %s", err)
		}
	}
}

func TestListResourcePrincipalFilterACL(t *testing.T) {
	expect := make(chan interface{})
	for _, resource := range resourcePatterns {
		args := append([]string{"acl", "list"}, resource.args...)
		for _, entry := range aclEntries {
			cmd := NewCMD(expect)
			cmd.SetArgs(append(args, "--service-account-id", strings.TrimPrefix(entry.entry.Principal, "User:")))

			go func() {
				expect <- convertToFilter(&kafkav1.ACLBinding{Pattern: resource.pattern, Entry: entry.entry})
			}()

			if err := cmd.Execute(); err != nil {
				t.Errorf("error: %s", err)
			}
		}
	}
}

func TestMultipleResourceACL(t *testing.T) {
	expect := "exactly one of"
	args := []string{"acl", "create", "--allow", "--operation", "read", "--service-account-id", "42",
		"--topic", "resource1", "--consumer-group", "resource2"}

	cmd := NewCMD(nil)
	cmd.SetArgs(args)

	err := cmd.Execute()
	if !strings.Contains(err.Error(), expect) {
		t.Logf("expected: %s got: %s", expect, err.Error())
		t.Fail()
		return
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

func TestDefaults(t *testing.T) {
	expect := make(chan interface{})
	cmd := NewCMD(expect)
	cmd.SetArgs([]string{"acl", "create", "--allow", "--service-account-id", "42",
		"--operation", "read" , "--topic", "dan"})
	go func() {
		expect <- &kafkav1.ACLBinding{
			Pattern: &kafkav1.ResourcePatternConfig{ResourceType: kafkav1.ResourceTypes_TOPIC, Name:"dan",
				PatternType: kafkav1.PatternTypes_LITERAL},
			Entry: &kafkav1.AccessControlEntryConfig{Host:"*", Principal:"User:42",
				Operation:kafkav1.ACLOperations_READ, PermissionType:kafkav1.ACLPermissionTypes_ALLOW},
		}
	}()

	if err:= cmd.Execute(); err != nil {
		t.Errorf("Topic PatternType was not set to default value of PatternTypes_LITERAL")
	}

	cmd = NewCMD(expect)
	cmd.SetArgs([]string{"acl", "create", "--cluster", "--allow", "--service-account-id", "42",
		"--operation", "read"})

	go func() {
		expect <- &kafkav1.ACLBinding{
			Pattern: &kafkav1.ResourcePatternConfig{ResourceType: kafkav1.ResourceTypes_CLUSTER, Name:"kafka-cluster",
				PatternType: kafkav1.PatternTypes_LITERAL},
			Entry: &kafkav1.AccessControlEntryConfig{Host:"*", Principal:"User:42",
				Operation:kafkav1.ACLOperations_READ, PermissionType:kafkav1.ACLPermissionTypes_ALLOW},
		}
	}()

	if err:= cmd.Execute(); err != nil {
		t.Errorf("Cluster PatternType was not set to default value of PatternTypes_LITERAL")
	}
}

/*************** TEST command_cluster ***************/
// TODO: do this for all commands/subcommands... and for all common error messages
func Test_HandleError_NotLoggedIn(t *testing.T) {
	cmd := New(conf, &mock.Kafka{
		ListFunc: func(ctx context.Context, cluster *kafkav1.KafkaCluster) ([]*kafkav1.KafkaCluster, error) {
			return nil, shared.ErrUnauthorized
		},
	})
	cmd.PersistentFlags().CountP("verbose", "v", "increase output verbosity")
	cmd.SetArgs(append([]string{"cluster", "list"}))
	buf := new(bytes.Buffer)
	cmd.SetOutput(buf)

	err := cmd.Execute()
	want := "You must login to access Confluent Cloud."
	if err.Error() != want {
		t.Errorf("unexpected output, got %s, want %s", err, want)
	}
}

/*************** TEST setup/helpers ***************/
func NewCMD(expect chan interface{}) *cobra.Command {
	cmd := New(conf, cliMock.NewKafkaMock(expect))
	cmd.PersistentFlags().CountP("verbose", "v", "increase output verbosity")

	return cmd
}

func init() {
	conf = shared.NewConfig()
	conf.Logger = log.New()
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
