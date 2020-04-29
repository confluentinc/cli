package kafka

import (
	"bytes"
	"context"
	"io/ioutil"
	"strconv"
	"strings"
	"testing"

	"github.com/confluentinc/ccloud-sdk-go"
	"github.com/confluentinc/ccloud-sdk-go/mock"
	kafkav1 "github.com/confluentinc/ccloudapis/kafka/v1"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
	cliMock "github.com/confluentinc/cli/mock"
)

var conf *v3.Config

func init() {
	stdin = bytes.NewBuffer(nil)
	stdout = bytes.NewBuffer(nil)
}

/*************** TEST command_acl ***************/
var resourcePatterns = []struct {
	args    []string
	pattern *kafkav1.ResourcePatternConfig
}{
	{
		args: []string{"--cluster-scope"},
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
	args    []string
	entries []*kafkav1.AccessControlEntryConfig
	err     error
}{
	{
		args: []string{"--allow", "--service-account", "42", "--operation", "read"},
		entries: []*kafkav1.AccessControlEntryConfig{
			{
				PermissionType: kafkav1.ACLPermissionTypes_ALLOW,
				Principal:      "User:42", Operation: kafkav1.ACLOperations_READ, Host: "*",
			},
		},
	},
	{
		args: []string{"--deny", "--service-account", "42", "--operation", "read"},
		entries: []*kafkav1.AccessControlEntryConfig{
			{
				PermissionType: kafkav1.ACLPermissionTypes_DENY,
				Principal:      "User:42", Operation: kafkav1.ACLOperations_READ, Host: "*",
			},
		},
	},
	{
		args: []string{"--allow", "--service-account", "42", "--operation", "write"},
		entries: []*kafkav1.AccessControlEntryConfig{
			{
				PermissionType: kafkav1.ACLPermissionTypes_ALLOW,
				Principal:      "User:42", Operation: kafkav1.ACLOperations_WRITE, Host: "*",
			},
		},
	},
	{
		args: []string{"--deny", "--service-account", "42", "--operation", "write"},
		entries: []*kafkav1.AccessControlEntryConfig{
			{
				PermissionType: kafkav1.ACLPermissionTypes_DENY,
				Principal:      "User:42", Operation: kafkav1.ACLOperations_WRITE, Host: "*",
			},
		},
	},
	{
		args: []string{"--allow", "--service-account", "42", "--operation", "create"},
		entries: []*kafkav1.AccessControlEntryConfig{
			{
				PermissionType: kafkav1.ACLPermissionTypes_ALLOW,
				Principal:      "User:42", Operation: kafkav1.ACLOperations_CREATE, Host: "*",
			},
		},
	},
	{
		args: []string{"--deny", "--service-account", "42", "--operation", "create"},
		entries: []*kafkav1.AccessControlEntryConfig{
			{
				PermissionType: kafkav1.ACLPermissionTypes_DENY,
				Principal:      "User:42", Operation: kafkav1.ACLOperations_CREATE, Host: "*",
			},
		},
	},
	{
		args: []string{"--allow", "--service-account", "42", "--operation", "delete"},
		entries: []*kafkav1.AccessControlEntryConfig{
			{
				PermissionType: kafkav1.ACLPermissionTypes_ALLOW,
				Principal:      "User:42", Operation: kafkav1.ACLOperations_DELETE, Host: "*",
			},
		},
	},
	{
		args: []string{"--deny", "--service-account", "42", "--operation", "delete"},
		entries: []*kafkav1.AccessControlEntryConfig{
			{
				PermissionType: kafkav1.ACLPermissionTypes_DENY,
				Principal:      "User:42", Operation: kafkav1.ACLOperations_DELETE, Host: "*",
			},
		},
	},
	{
		args: []string{"--allow", "--service-account", "42", "--operation", "alter"},
		entries: []*kafkav1.AccessControlEntryConfig{
			{
				PermissionType: kafkav1.ACLPermissionTypes_ALLOW,
				Principal:      "User:42", Operation: kafkav1.ACLOperations_ALTER, Host: "*",
			},
		},
	},
	{
		args: []string{"--deny", "--service-account", "42", "--operation", "alter"},
		entries: []*kafkav1.AccessControlEntryConfig{
			{
				PermissionType: kafkav1.ACLPermissionTypes_DENY,
				Principal:      "User:42", Operation: kafkav1.ACLOperations_ALTER, Host: "*",
			},
		},
	},
	{
		args: []string{"--allow", "--service-account", "42", "--operation", "describe"},
		entries: []*kafkav1.AccessControlEntryConfig{
			{
				PermissionType: kafkav1.ACLPermissionTypes_ALLOW,
				Principal:      "User:42", Operation: kafkav1.ACLOperations_DESCRIBE, Host: "*",
			},
		},
	},
	{
		args: []string{"--deny", "--service-account", "42", "--operation", "describe"},
		entries: []*kafkav1.AccessControlEntryConfig{
			{
				PermissionType: kafkav1.ACLPermissionTypes_DENY,
				Principal:      "User:42", Operation: kafkav1.ACLOperations_DESCRIBE, Host: "*",
			},
		},
	},
	{
		args: []string{"--allow", "--service-account", "42", "--operation", "cluster-action"},
		entries: []*kafkav1.AccessControlEntryConfig{
			{
				PermissionType: kafkav1.ACLPermissionTypes_ALLOW,
				Principal:      "User:42", Operation: kafkav1.ACLOperations_CLUSTER_ACTION, Host: "*",
			},
		},
	},
	{
		args: []string{"--deny", "--service-account", "42", "--operation", "cluster-action"},
		entries: []*kafkav1.AccessControlEntryConfig{
			{
				PermissionType: kafkav1.ACLPermissionTypes_DENY,
				Principal:      "User:42", Operation: kafkav1.ACLOperations_CLUSTER_ACTION, Host: "*",
			},
		},
	},
	{
		args: []string{"--allow", "--service-account", "42", "--operation", "describe-configs"},
		entries: []*kafkav1.AccessControlEntryConfig{
			{
				PermissionType: kafkav1.ACLPermissionTypes_ALLOW,
				Principal:      "User:42", Operation: kafkav1.ACLOperations_DESCRIBE_CONFIGS, Host: "*",
			},
		},
	},
	{
		args: []string{"--deny", "--service-account", "42", "--operation", "describe-configs"},
		entries: []*kafkav1.AccessControlEntryConfig{
			{
				PermissionType: kafkav1.ACLPermissionTypes_DENY,
				Principal:      "User:42", Operation: kafkav1.ACLOperations_DESCRIBE_CONFIGS, Host: "*",
			},
		},
	},
	{
		args: []string{"--allow", "--service-account", "42", "--operation", "alter-configs"},
		entries: []*kafkav1.AccessControlEntryConfig{
			{
				PermissionType: kafkav1.ACLPermissionTypes_ALLOW,
				Principal:      "User:42", Operation: kafkav1.ACLOperations_ALTER_CONFIGS, Host: "*",
			},
		},
	},
	{
		args: []string{"--deny", "--service-account", "42", "--operation", "alter-configs"},
		entries: []*kafkav1.AccessControlEntryConfig{
			{
				PermissionType: kafkav1.ACLPermissionTypes_DENY,
				Principal:      "User:42", Operation: kafkav1.ACLOperations_ALTER_CONFIGS, Host: "*",
			},
		},
	},
	{
		args: []string{"--allow", "--service-account", "42", "--operation", "idempotent-write"},
		entries: []*kafkav1.AccessControlEntryConfig{
			{
				PermissionType: kafkav1.ACLPermissionTypes_ALLOW,
				Principal:      "User:42", Operation: kafkav1.ACLOperations_IDEMPOTENT_WRITE, Host: "*",
			},
		},
	},
	{
		args: []string{"--deny", "--service-account", "42", "--operation", "idempotent-write"},
		entries: []*kafkav1.AccessControlEntryConfig{
			{
				PermissionType: kafkav1.ACLPermissionTypes_DENY,
				Principal:      "User:42", Operation: kafkav1.ACLOperations_IDEMPOTENT_WRITE, Host: "*",
			},
		},
	},
	{
		args: []string{"--deny", "--service-account", "42", "--operation", "alter-configs", "--operation", "idempotent-write", "--operation", "create"},
		entries: []*kafkav1.AccessControlEntryConfig{
			{
				PermissionType: kafkav1.ACLPermissionTypes_DENY,
				Principal:      "User:42", Operation: kafkav1.ACLOperations_ALTER_CONFIGS, Host: "*",
			},
			{
				PermissionType: kafkav1.ACLPermissionTypes_DENY,
				Principal:      "User:42", Operation: kafkav1.ACLOperations_IDEMPOTENT_WRITE, Host: "*",
			},
			{
				PermissionType: kafkav1.ACLPermissionTypes_DENY,
				Principal:      "User:42", Operation: kafkav1.ACLOperations_CREATE, Host: "*",
			},
		},
	},
}

func TestCreateACLs(t *testing.T) {
	expect := make(chan interface{})
	for _, resource := range resourcePatterns {
		args := append([]string{"acl", "create"}, resource.args...)
		for _, aclEntry := range aclEntries {
			cmd := NewCMD(expect)
			cmd.SetArgs(append(args, aclEntry.args...))

			go func() {
				bindings := []*kafkav1.ACLBinding{}
				for _, entry := range aclEntry.entries {
					bindings = append(bindings, &kafkav1.ACLBinding{Pattern: resource.pattern, Entry: entry})
				}
				expect <- bindings
			}()

			if err := cmd.Execute(); err != nil {
				t.Errorf("error: %s", err)
			}
		}
	}
}

func TestDeleteACLs(t *testing.T) {
	expect := make(chan interface{})
	for _, resource := range resourcePatterns {
		args := append([]string{"acl", "delete"}, resource.args...)
		for _, aclEntry := range aclEntries {
			cmd := NewCMD(expect)
			cmd.SetArgs(append(args, aclEntry.args...))

			go func() {
				filters := []*kafkav1.ACLFilter{}
				for _, entry := range aclEntry.entries {
					filters = append(filters, convertToFilter(&kafkav1.ACLBinding{Pattern: resource.pattern, Entry: entry}))
				}
				expect <- filters
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
	for _, aclEntry := range aclEntries {
		if len(aclEntry.entries) != 1 {
			continue
		}
		entry := aclEntry.entries[0]
		cmd := NewCMD(expect)
		cmd.SetArgs(append([]string{"acl", "list", "--service-account"}, strings.TrimPrefix(entry.Principal, "User:")))

		go func() {
			expect <- convertToFilter(&kafkav1.ACLBinding{Entry: &kafkav1.AccessControlEntryConfig{Principal: entry.Principal}})
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
		for _, aclEntry := range aclEntries {
			if len(aclEntry.entries) != 1 {
				continue
			}
			entry := aclEntry.entries[0]
			cmd := NewCMD(expect)
			cmd.SetArgs(append(args, "--service-account", strings.TrimPrefix(entry.Principal, "User:")))

			go func() {
				expect <- convertToFilter(&kafkav1.ACLBinding{Pattern: resource.pattern, Entry: entry})
			}()

			if err := cmd.Execute(); err != nil {
				t.Errorf("error: %s", err)
			}
		}
	}
}

func TestMultipleResourceACL(t *testing.T) {
	expect := "exactly one of cluster-scope, consumer-group, topic, transactional-id must be set"
	args := []string{"acl", "create", "--allow", "--operation", "read", "--service-account", "42",
		"--topic", "resource1", "--consumer-group", "resource2"}

	cmd := NewCMD(nil)
	cmd.SetArgs(args)

	err := cmd.Execute()
	if !strings.Contains(err.Error(), expect) {
		t.Errorf("expected: %s got: %s", expect, err.Error())
	}
}

/*************** TEST command_topic ***************/
var Topics = []struct {
	args []string
	spec *kafkav1.TopicSpecification
}{
	{
		args: []string{"test_topic", "--config", "a=b", "--partitions", strconv.Itoa(1)},
		spec: &kafkav1.TopicSpecification{Name: "test_topic", ReplicationFactor: 3, NumPartitions: 1, Configs: map[string]string{"a": "b"}},
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
	cmd.SetArgs([]string{"acl", "create", "--allow", "--service-account", "42",
		"--operation", "read", "--topic", "dan"})
	go func() {
		expect <- []*kafkav1.ACLBinding{
			{
				Pattern: &kafkav1.ResourcePatternConfig{ResourceType: kafkav1.ResourceTypes_TOPIC, Name: "dan",
					PatternType: kafkav1.PatternTypes_LITERAL},
				Entry: &kafkav1.AccessControlEntryConfig{Host: "*", Principal: "User:42",
					Operation: kafkav1.ACLOperations_READ, PermissionType: kafkav1.ACLPermissionTypes_ALLOW},
			},
		}
	}()

	if err := cmd.Execute(); err != nil {
		t.Errorf("Topic PatternType was not set to default value of PatternTypes_LITERAL")
	}

	cmd = NewCMD(expect)
	cmd.SetArgs([]string{"acl", "create", "--cluster-scope", "--allow", "--service-account", "42",
		"--operation", "read"})

	go func() {
		expect <- []*kafkav1.ACLBinding{
			{
				Pattern: &kafkav1.ResourcePatternConfig{ResourceType: kafkav1.ResourceTypes_CLUSTER, Name: "kafka-cluster",
					PatternType: kafkav1.PatternTypes_LITERAL},
				Entry: &kafkav1.AccessControlEntryConfig{Host: "*", Principal: "User:42",
					Operation: kafkav1.ACLOperations_READ, PermissionType: kafkav1.ACLPermissionTypes_ALLOW},
			},
		}
	}()

	if err := cmd.Execute(); err != nil {
		t.Errorf("Cluster PatternType was not set to default value of PatternTypes_LITERAL")
	}
}

/*************** TEST command_cluster ***************/
// TODO: do this for all commands/subcommands... and for all common error messages
func Test_HandleError_NotLoggedIn(t *testing.T) {
	kafka := &mock.Kafka{
		ListFunc: func(ctx context.Context, cluster *kafkav1.KafkaCluster) ([]*kafkav1.KafkaCluster, error) {
			return nil, errors.ErrNotLoggedIn
		},
	}
	client := &ccloud.Client{Kafka: kafka}
	cmd := New(cliMock.NewPreRunnerMock(client, nil), conf, log.New(), "test-client")
	cmd.PersistentFlags().CountP("verbose", "v", "Increase output verbosity")
	cmd.SetArgs(append([]string{"cluster", "list"}))
	buf := new(bytes.Buffer)
	cmd.SetOutput(buf)

	err := cmd.Execute()
	want := "You must log in to run that command."
	if err.Error() != want {
		t.Errorf("unexpected output, got %s, want %s", err, want)
	}
}

/*************** TEST setup/helpers ***************/
func NewCMD(expect chan interface{}) *cobra.Command {
	client := &ccloud.Client{
		Kafka: cliMock.NewKafkaMock(expect),
		EnvironmentMetadata: &mock.EnvironmentMetadata{
			GetFunc: func(ctx context.Context) ([]*kafkav1.CloudMetadata, error) {
				return []*kafkav1.CloudMetadata{{
					Id:       "aws",
					Accounts: []*kafkav1.AccountMetadata{{Id: "account-xyz"}},
					Regions:  []*kafkav1.Region{{IsSchedulable: true, Id: "us-west-2"}},
				}}, nil
			},
		},
	}
	cmd := New(cliMock.NewPreRunnerMock(client, nil), conf, log.New(), "test-client")
	cmd.PersistentFlags().CountP("verbose", "v", "Increase output verbosity")

	return cmd
}

func TestCreateEncryptionKeyId(t *testing.T) {
	c := make(chan interface{})

	_, err := stdin.Write([]byte("y\n"))
	require.NoError(t, err)

	cmd := NewCMD(c)
	// err: not dedicated, the api validates this too
	cmd.SetArgs([]string{
		"cluster",
		"create",
		"name-xyz",
		"--region=us-west-2",
		"--cloud=aws",
		"--encryption-key=xyz",
		"--type=dedicated",
		"--cku=4",
	})
	err = cmd.Execute()
	require.NoError(t, err)

	b, err := ioutil.ReadAll(stdout)
	require.NoError(t, err)
	require.Equal(t, "Please confirm you've authorized the key for these accounts account-xyz (y/n): ", string(b))
}

func init() {
	conf = v3.AuthenticatedCloudConfigMock()
}
