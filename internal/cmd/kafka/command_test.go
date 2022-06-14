package kafka

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/ccloud-sdk-go-v1"
	"github.com/confluentinc/ccloud-sdk-go-v1/mock"
	cmkv2 "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2"
	cmkmock "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2/mock"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	krsdk "github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	cliMock "github.com/confluentinc/cli/mock"
)

const (
	serviceAccountId         = int32(123)
	serviceAccountResourceId = "sa-123"
	serviceAccountName       = "service-account"
	userId                   = int32(456)
	userResourceId           = "sa-456"
)

var conf *v1.Config

/*************** TEST command_acl ***************/
var resourcePatterns = []struct {
	args    []string
	pattern *schedv1.ResourcePatternConfig
}{
	{
		args: []string{"--cluster-scope"},
		pattern: &schedv1.ResourcePatternConfig{ResourceType: schedv1.ResourceTypes_CLUSTER, Name: "kafka-cluster",
			PatternType: schedv1.PatternTypes_LITERAL},
	},
	{
		args: []string{"--topic", "test-topic"},
		pattern: &schedv1.ResourcePatternConfig{ResourceType: schedv1.ResourceTypes_TOPIC, Name: "test-topic",
			PatternType: schedv1.PatternTypes_LITERAL},
	},
	{
		args: []string{"--topic", "test-topic", "--prefix"},
		pattern: &schedv1.ResourcePatternConfig{ResourceType: schedv1.ResourceTypes_TOPIC, Name: "test-topic",
			PatternType: schedv1.PatternTypes_PREFIXED},
	},
	{
		args: []string{"--consumer-group", "test-group"},
		pattern: &schedv1.ResourcePatternConfig{ResourceType: schedv1.ResourceTypes_GROUP, Name: "test-group",
			PatternType: schedv1.PatternTypes_LITERAL},
	},
	{
		args: []string{"--consumer-group", "test-group", "--prefix"},
		pattern: &schedv1.ResourcePatternConfig{ResourceType: schedv1.ResourceTypes_GROUP, Name: "test-group",
			PatternType: schedv1.PatternTypes_PREFIXED},
	},
	{
		args: []string{"--transactional-id", "test-transactional-id"},
		pattern: &schedv1.ResourcePatternConfig{ResourceType: schedv1.ResourceTypes_TRANSACTIONAL_ID, Name: "test-transactional-id",
			PatternType: schedv1.PatternTypes_LITERAL},
	},
	{
		args: []string{"--transactional-id", "test-transactional-id", "--prefix"},
		pattern: &schedv1.ResourcePatternConfig{ResourceType: schedv1.ResourceTypes_TRANSACTIONAL_ID, Name: "test-transactional-id",
			PatternType: schedv1.PatternTypes_PREFIXED},
	},
	{
		args: []string{"--prefix", "--topic", "test-topic"},
		pattern: &schedv1.ResourcePatternConfig{ResourceType: schedv1.ResourceTypes_TOPIC, Name: "test-topic",
			PatternType: schedv1.PatternTypes_PREFIXED},
	},
}

var aclEntries = []struct {
	args    []string
	entries []*schedv1.AccessControlEntryConfig
	err     error
}{
	{
		args: []string{"--allow", "--service-account", "sa-42", "--operation", "read"},
		entries: []*schedv1.AccessControlEntryConfig{
			{
				PermissionType: schedv1.ACLPermissionTypes_ALLOW,
				Principal:      "User:42", Operation: schedv1.ACLOperations_READ, Host: "*",
			},
		},
	},
	{
		args: []string{"--deny", "--service-account", "sa-42", "--operation", "read"},
		entries: []*schedv1.AccessControlEntryConfig{
			{
				PermissionType: schedv1.ACLPermissionTypes_DENY,
				Principal:      "User:42", Operation: schedv1.ACLOperations_READ, Host: "*",
			},
		},
	},
	{
		args: []string{"--allow", "--service-account", "sa-42", "--operation", "write"},
		entries: []*schedv1.AccessControlEntryConfig{
			{
				PermissionType: schedv1.ACLPermissionTypes_ALLOW,
				Principal:      "User:42", Operation: schedv1.ACLOperations_WRITE, Host: "*",
			},
		},
	},
	{
		args: []string{"--deny", "--service-account", "sa-42", "--operation", "write"},
		entries: []*schedv1.AccessControlEntryConfig{
			{
				PermissionType: schedv1.ACLPermissionTypes_DENY,
				Principal:      "User:42", Operation: schedv1.ACLOperations_WRITE, Host: "*",
			},
		},
	},
	{
		args: []string{"--allow", "--service-account", "sa-42", "--operation", "create"},
		entries: []*schedv1.AccessControlEntryConfig{
			{
				PermissionType: schedv1.ACLPermissionTypes_ALLOW,
				Principal:      "User:42", Operation: schedv1.ACLOperations_CREATE, Host: "*",
			},
		},
	},
	{
		args: []string{"--deny", "--service-account", "sa-42", "--operation", "create"},
		entries: []*schedv1.AccessControlEntryConfig{
			{
				PermissionType: schedv1.ACLPermissionTypes_DENY,
				Principal:      "User:42", Operation: schedv1.ACLOperations_CREATE, Host: "*",
			},
		},
	},
	{
		args: []string{"--allow", "--service-account", "sa-42", "--operation", "delete"},
		entries: []*schedv1.AccessControlEntryConfig{
			{
				PermissionType: schedv1.ACLPermissionTypes_ALLOW,
				Principal:      "User:42", Operation: schedv1.ACLOperations_DELETE, Host: "*",
			},
		},
	},
	{
		args: []string{"--deny", "--service-account", "sa-42", "--operation", "delete"},
		entries: []*schedv1.AccessControlEntryConfig{
			{
				PermissionType: schedv1.ACLPermissionTypes_DENY,
				Principal:      "User:42", Operation: schedv1.ACLOperations_DELETE, Host: "*",
			},
		},
	},
	{
		args: []string{"--allow", "--service-account", "sa-42", "--operation", "alter"},
		entries: []*schedv1.AccessControlEntryConfig{
			{
				PermissionType: schedv1.ACLPermissionTypes_ALLOW,
				Principal:      "User:42", Operation: schedv1.ACLOperations_ALTER, Host: "*",
			},
		},
	},
	{
		args: []string{"--deny", "--service-account", "sa-42", "--operation", "alter"},
		entries: []*schedv1.AccessControlEntryConfig{
			{
				PermissionType: schedv1.ACLPermissionTypes_DENY,
				Principal:      "User:42", Operation: schedv1.ACLOperations_ALTER, Host: "*",
			},
		},
	},
	{
		args: []string{"--allow", "--service-account", "sa-42", "--operation", "describe"},
		entries: []*schedv1.AccessControlEntryConfig{
			{
				PermissionType: schedv1.ACLPermissionTypes_ALLOW,
				Principal:      "User:42", Operation: schedv1.ACLOperations_DESCRIBE, Host: "*",
			},
		},
	},
	{
		args: []string{"--deny", "--service-account", "sa-42", "--operation", "describe"},
		entries: []*schedv1.AccessControlEntryConfig{
			{
				PermissionType: schedv1.ACLPermissionTypes_DENY,
				Principal:      "User:42", Operation: schedv1.ACLOperations_DESCRIBE, Host: "*",
			},
		},
	},
	{
		args: []string{"--allow", "--service-account", "sa-42", "--operation", "cluster-action"},
		entries: []*schedv1.AccessControlEntryConfig{
			{
				PermissionType: schedv1.ACLPermissionTypes_ALLOW,
				Principal:      "User:42", Operation: schedv1.ACLOperations_CLUSTER_ACTION, Host: "*",
			},
		},
	},
	{
		args: []string{"--deny", "--service-account", "sa-42", "--operation", "cluster-action"},
		entries: []*schedv1.AccessControlEntryConfig{
			{
				PermissionType: schedv1.ACLPermissionTypes_DENY,
				Principal:      "User:42", Operation: schedv1.ACLOperations_CLUSTER_ACTION, Host: "*",
			},
		},
	},
	{
		args: []string{"--allow", "--service-account", "sa-42", "--operation", "describe-configs"},
		entries: []*schedv1.AccessControlEntryConfig{
			{
				PermissionType: schedv1.ACLPermissionTypes_ALLOW,
				Principal:      "User:42", Operation: schedv1.ACLOperations_DESCRIBE_CONFIGS, Host: "*",
			},
		},
	},
	{
		args: []string{"--deny", "--service-account", "sa-42", "--operation", "describe-configs"},
		entries: []*schedv1.AccessControlEntryConfig{
			{
				PermissionType: schedv1.ACLPermissionTypes_DENY,
				Principal:      "User:42", Operation: schedv1.ACLOperations_DESCRIBE_CONFIGS, Host: "*",
			},
		},
	},
	{
		args: []string{"--allow", "--service-account", "sa-42", "--operation", "alter-configs"},
		entries: []*schedv1.AccessControlEntryConfig{
			{
				PermissionType: schedv1.ACLPermissionTypes_ALLOW,
				Principal:      "User:42", Operation: schedv1.ACLOperations_ALTER_CONFIGS, Host: "*",
			},
		},
	},
	{
		args: []string{"--deny", "--service-account", "sa-42", "--operation", "alter-configs"},
		entries: []*schedv1.AccessControlEntryConfig{
			{
				PermissionType: schedv1.ACLPermissionTypes_DENY,
				Principal:      "User:42", Operation: schedv1.ACLOperations_ALTER_CONFIGS, Host: "*",
			},
		},
	},
	{
		args: []string{"--allow", "--service-account", "sa-42", "--operation", "idempotent-write"},
		entries: []*schedv1.AccessControlEntryConfig{
			{
				PermissionType: schedv1.ACLPermissionTypes_ALLOW,
				Principal:      "User:42", Operation: schedv1.ACLOperations_IDEMPOTENT_WRITE, Host: "*",
			},
		},
	},
	{
		args: []string{"--deny", "--service-account", "sa-42", "--operation", "idempotent-write"},
		entries: []*schedv1.AccessControlEntryConfig{
			{
				PermissionType: schedv1.ACLPermissionTypes_DENY,
				Principal:      "User:42", Operation: schedv1.ACLOperations_IDEMPOTENT_WRITE, Host: "*",
			},
		},
	},
	{
		args: []string{"--deny", "--service-account", "sa-42", "--operation", "alter-configs", "--operation", "idempotent-write", "--operation", "create"},
		entries: []*schedv1.AccessControlEntryConfig{
			{
				PermissionType: schedv1.ACLPermissionTypes_DENY,
				Principal:      "User:42", Operation: schedv1.ACLOperations_ALTER_CONFIGS, Host: "*",
			},
			{
				PermissionType: schedv1.ACLPermissionTypes_DENY,
				Principal:      "User:42", Operation: schedv1.ACLOperations_IDEMPOTENT_WRITE, Host: "*",
			},
			{
				PermissionType: schedv1.ACLPermissionTypes_DENY,
				Principal:      "User:42", Operation: schedv1.ACLOperations_CREATE, Host: "*",
			},
		},
	},
	{
		args: []string{"--allow", "--principal", "User:sa-456", "--operation", "read"},
		entries: []*schedv1.AccessControlEntryConfig{
			{
				PermissionType: schedv1.ACLPermissionTypes_ALLOW,
				Principal:      "User:456", Operation: schedv1.ACLOperations_READ, Host: "*",
			},
		},
	},
	{
		args: []string{"--deny", "--principal", "User:sa-456", "--operation", "read"},
		entries: []*schedv1.AccessControlEntryConfig{
			{
				PermissionType: schedv1.ACLPermissionTypes_DENY,
				Principal:      "User:456", Operation: schedv1.ACLOperations_READ, Host: "*",
			},
		},
	},
	{
		args: []string{"--allow", "--principal", "User:sa-456", "--operation", "write"},
		entries: []*schedv1.AccessControlEntryConfig{
			{
				PermissionType: schedv1.ACLPermissionTypes_ALLOW,
				Principal:      "User:456", Operation: schedv1.ACLOperations_WRITE, Host: "*",
			},
		},
	},
	{
		args: []string{"--deny", "--principal", "User:sa-456", "--operation", "write"},
		entries: []*schedv1.AccessControlEntryConfig{
			{
				PermissionType: schedv1.ACLPermissionTypes_DENY,
				Principal:      "User:456", Operation: schedv1.ACLOperations_WRITE, Host: "*",
			},
		},
	},
	{
		args: []string{"--allow", "--principal", "User:sa-456", "--operation", "create"},
		entries: []*schedv1.AccessControlEntryConfig{
			{
				PermissionType: schedv1.ACLPermissionTypes_ALLOW,
				Principal:      "User:456", Operation: schedv1.ACLOperations_CREATE, Host: "*",
			},
		},
	},
	{
		args: []string{"--deny", "--principal", "User:sa-456", "--operation", "create"},
		entries: []*schedv1.AccessControlEntryConfig{
			{
				PermissionType: schedv1.ACLPermissionTypes_DENY,
				Principal:      "User:456", Operation: schedv1.ACLOperations_CREATE, Host: "*",
			},
		},
	},
	{
		args: []string{"--allow", "--principal", "User:sa-456", "--operation", "delete"},
		entries: []*schedv1.AccessControlEntryConfig{
			{
				PermissionType: schedv1.ACLPermissionTypes_ALLOW,
				Principal:      "User:456", Operation: schedv1.ACLOperations_DELETE, Host: "*",
			},
		},
	},
	{
		args: []string{"--deny", "--principal", "User:sa-456", "--operation", "delete"},
		entries: []*schedv1.AccessControlEntryConfig{
			{
				PermissionType: schedv1.ACLPermissionTypes_DENY,
				Principal:      "User:456", Operation: schedv1.ACLOperations_DELETE, Host: "*",
			},
		},
	},
	{
		args: []string{"--allow", "--principal", "User:sa-456", "--operation", "alter"},
		entries: []*schedv1.AccessControlEntryConfig{
			{
				PermissionType: schedv1.ACLPermissionTypes_ALLOW,
				Principal:      "User:456", Operation: schedv1.ACLOperations_ALTER, Host: "*",
			},
		},
	},
	{
		args: []string{"--deny", "--principal", "User:sa-456", "--operation", "alter"},
		entries: []*schedv1.AccessControlEntryConfig{
			{
				PermissionType: schedv1.ACLPermissionTypes_DENY,
				Principal:      "User:456", Operation: schedv1.ACLOperations_ALTER, Host: "*",
			},
		},
	},
}

func CreateACLsTest(t *testing.T, enableREST bool) {
	expect := make(chan interface{})
	for _, resource := range resourcePatterns {
		args := append([]string{"acl", "create"}, resource.args...)
		for _, aclEntry := range aclEntries {
			cmd := newCmd(expect, enableREST)
			cmd.SetArgs(append(args, aclEntry.args...))

			// TODO: better testing of KafkaREST
			if !enableREST {
				go func() {
					var bindings []*schedv1.ACLBinding
					for _, entry := range aclEntry.entries {
						bindings = append(bindings, &schedv1.ACLBinding{Pattern: resource.pattern, Entry: entry})
					}
					expect <- bindings
				}()
			}

			if err := cmd.Execute(); err != nil {
				t.Errorf("error: %s", err)
			}
		}
	}
}

func TestCreateACLs1(t *testing.T) {
	CreateACLsTest(t, true)
}

func TestCreateACLs2(t *testing.T) {
	CreateACLsTest(t, false)
}

func DeleteACLsTest(t *testing.T, enableREST bool) {
	for i := range resourcePatterns {
		args := append([]string{"acl", "delete"}, resourcePatterns[i].args...)
		for j := range aclEntries {
			expect := make(chan interface{})
			cmd := newCmd(expect, enableREST)
			cmd.SetArgs(append(args, aclEntries[j].args...))

			var filters []*schedv1.ACLFilter
			for _, entry := range aclEntries[j].entries {
				filters = append(filters, convertToFilter(&schedv1.ACLBinding{Pattern: resourcePatterns[i].pattern, Entry: entry}))
			}

			go func() {
				expect <- filters
			}()

			if err := cmd.Execute(); err != nil {
				t.Errorf("error: %s", err)
			}
		}
	}
}

func TestDeleteACLs1(t *testing.T) {
	DeleteACLsTest(t, true)
}

func TestDeleteACLs2(t *testing.T) {
	DeleteACLsTest(t, false)
}

func ListResourceACLTest(t *testing.T, enableREST bool) {
	expect := make(chan interface{})
	for _, resource := range resourcePatterns {
		cmd := newCmd(expect, enableREST)
		cmd.SetArgs(append([]string{"acl", "list"}, resource.args...))

		// TODO: better testing of KafkaREST
		if !enableREST {
			go func() {
				expect <- convertToFilter(&schedv1.ACLBinding{Pattern: resource.pattern, Entry: &schedv1.AccessControlEntryConfig{}})
			}()
		}

		if err := cmd.Execute(); err != nil {
			t.Errorf("error: %s", err)
		}
	}
}

func TestListResourceACL1(t *testing.T) {
	ListResourceACLTest(t, true)
}

func TestListResourceACL2(t *testing.T) {
	ListResourceACLTest(t, false)
}

func ListPrincipalACLTest(t *testing.T, enableREST bool) {
	expect := make(chan interface{})
	for _, aclEntry := range aclEntries {
		if len(aclEntry.entries) != 1 {
			continue
		}
		entry := aclEntry.entries[0]
		cmd := newCmd(expect, enableREST)
		cmd.SetArgs(append([]string{"acl", "list", "--service-account"}, "sa-"+strings.TrimPrefix(entry.Principal, "User:")))

		go func() {
			expect <- convertToFilter(&schedv1.ACLBinding{Entry: &schedv1.AccessControlEntryConfig{Principal: entry.Principal}})
		}()

		if err := cmd.Execute(); err != nil {
			t.Errorf("error: %s", err)
		}
	}
}

func TestListPrincipalACL1(t *testing.T) {
	ListPrincipalACLTest(t, true)
}

func TestListPrincipalACL2(t *testing.T) {
	ListPrincipalACLTest(t, false)
}

func ListResourcePrincipalFilterACLTest(t *testing.T, enableREST bool) {
	expect := make(chan interface{})
	for _, resource := range resourcePatterns {
		args := append([]string{"acl", "list"}, resource.args...)
		for _, aclEntry := range aclEntries {
			if len(aclEntry.entries) != 1 {
				continue
			}
			entry := aclEntry.entries[0]
			cmd := newCmd(expect, enableREST)
			cmd.SetArgs(append(args, "--service-account", "sa-"+strings.TrimPrefix(entry.Principal, "User:")))

			// TODO: better testing of KafkaREST
			if !enableREST {
				go func() {
					expect <- convertToFilter(&schedv1.ACLBinding{Pattern: resource.pattern, Entry: entry})
				}()
			}

			if err := cmd.Execute(); err != nil {
				t.Errorf("error: %s", err)
			}
		}
	}
}

func TestListResourcePrincipalFilterACL1(t *testing.T) {
	ListResourcePrincipalFilterACLTest(t, true)
}

func TestListResourcePrincipalFilterACL2(t *testing.T) {
	ListResourcePrincipalFilterACLTest(t, false)
}

func MultipleResourceACLTest(t *testing.T, enableREST bool) {
	args := []string{"acl", "create", "--allow", "--operation", "read", "--service-account", "sa-42",
		"--topic", "resource1", "--consumer-group", "resource2"}

	cmd := newCmd(nil, enableREST)
	cmd.SetArgs(args)

	err := cmd.Execute()
	expect := fmt.Sprintf(errors.ExactlyOneSetErrorMsg, "cluster-scope, consumer-group, topic, transactional-id")
	if !strings.Contains(err.Error(), expect) {
		t.Errorf("expected: %s got: %s", expect, err.Error())
	}
}

func TestMultipleResourceACL1(t *testing.T) {
	MultipleResourceACLTest(t, true)
}

func TestMultipleResourceACL2(t *testing.T) {
	MultipleResourceACLTest(t, false)
}

/*************** TEST command_topic ***************/
var Topics = []struct {
	args []string
	spec *schedv1.TopicSpecification
}{
	{
		args: []string{"test_topic", "--config", "a=b", "--partitions", strconv.Itoa(1)},
		spec: &schedv1.TopicSpecification{Name: "test_topic", ReplicationFactor: 3, NumPartitions: 1, Configs: map[string]string{"a": "b"}},
	},
}

func ListTopicTest(t *testing.T, enableREST bool) {
	expect := make(chan interface{})
	for _, topic := range Topics {
		cmd := newCmd(expect, enableREST)
		cmd.SetArgs([]string{"topic", "list"})
		go func() {
			expect <- &schedv1.Topic{Spec: &schedv1.TopicSpecification{Name: topic.spec.Name}}
		}()

		if err := cmd.Execute(); err != nil {
			t.Errorf("error: %s", err)
			t.Fail()
			return
		}
	}
}

func TestListTopics1(t *testing.T) {
	ListTopicTest(t, true)
}

func TestListTopics2(t *testing.T) {
	ListTopicTest(t, false)
}

func CreateTopicTest(t *testing.T, enableREST bool) {
	expect := make(chan interface{})
	for _, topic := range Topics {
		cmd := newCmd(expect, enableREST)
		cmd.SetArgs(append([]string{"topic", "create"}, topic.args...))

		go func() {
			expect <- &schedv1.Topic{Spec: topic.spec}
		}()

		if err := cmd.Execute(); err != nil {
			t.Errorf("error: %s", err)
			t.Fail()
			return
		}
	}
}

func TestCreateTopic1(t *testing.T) {
	CreateTopicTest(t, true)
}

func TestCreateTopic2(t *testing.T) {
	CreateTopicTest(t, false)
}

func DescribeTopicTest(t *testing.T, enableREST bool) {
	expect := make(chan interface{})
	for _, topic := range Topics {
		cmd := newCmd(expect, enableREST)
		cmd.SetArgs(append([]string{"topic", "describe"}, topic.args[0]))

		go func() {
			expect <- &schedv1.Topic{Spec: &schedv1.TopicSpecification{Name: topic.spec.Name}}
		}()

		if err := cmd.Execute(); err != nil {
			t.Errorf("error: %s", err)
			t.Fail()
			return
		}
	}
}

func TestDescribeTopic1(t *testing.T) {
	DescribeTopicTest(t, true)
}

func TestDescribeTopic2(t *testing.T) {
	DescribeTopicTest(t, false)
}

func DeleteTopicTest(t *testing.T, enableREST bool) {
	expect := make(chan interface{})
	for _, topic := range Topics {
		cmd := newCmd(expect, enableREST)
		cmd.SetArgs(append([]string{"topic", "delete"}, topic.args[0]))

		go func() {
			expect <- &schedv1.Topic{Spec: &schedv1.TopicSpecification{Name: topic.spec.Name}}
		}()

		if err := cmd.Execute(); err != nil {
			t.Errorf("error: %s", err)
			t.Fail()
			return
		}
	}
}

func TestDeleteTopic1(t *testing.T) {
	DeleteTopicTest(t, true)
}

func TestDeleteTopic2(t *testing.T) {
	DeleteTopicTest(t, false)
}

func UpdateTopicTest(t *testing.T, enableREST bool) {
	expect := make(chan interface{})
	for _, topic := range Topics {
		cmd := newCmd(expect, enableREST)
		cmd.SetArgs(append([]string{"topic", "update"}, topic.args[0:3]...))
		go func() {
			expect <- &schedv1.Topic{Spec: &schedv1.TopicSpecification{Name: topic.spec.Name, Configs: topic.spec.Configs}}
		}()

		if err := cmd.Execute(); err != nil {
			t.Errorf("error: %s", err)
			t.Fail()
			return
		}
	}
}

func TestUpdateTopic1(t *testing.T) {
	UpdateTopicTest(t, true)
}

func TestUpdateTopic2(t *testing.T) {
	UpdateTopicTest(t, false)
}

func DefaultsTest(t *testing.T, enableREST bool) {
	expect := make(chan interface{})
	cmd := newCmd(expect, enableREST)
	cmd.SetArgs([]string{"acl", "create", "--allow", "--service-account", "sa-42",
		"--operation", "read", "--topic", "dan"})
	go func() {
		expect <- []*schedv1.ACLBinding{
			{
				Pattern: &schedv1.ResourcePatternConfig{ResourceType: schedv1.ResourceTypes_TOPIC, Name: "dan",
					PatternType: schedv1.PatternTypes_LITERAL},
				Entry: &schedv1.AccessControlEntryConfig{Host: "*", Principal: "User:42",
					Operation: schedv1.ACLOperations_READ, PermissionType: schedv1.ACLPermissionTypes_ALLOW},
			},
		}
	}()

	if err := cmd.Execute(); err != nil {
		t.Errorf("Topic PatternType was not set to default value of PatternTypes_LITERAL")
	}

	cmd = newCmd(expect, enableREST)
	cmd.SetArgs([]string{"acl", "create", "--cluster-scope", "--allow", "--service-account", "sa-42",
		"--operation", "read"})

	go func() {
		expect <- []*schedv1.ACLBinding{
			{
				Pattern: &schedv1.ResourcePatternConfig{ResourceType: schedv1.ResourceTypes_CLUSTER, Name: "kafka-cluster",
					PatternType: schedv1.PatternTypes_LITERAL},
				Entry: &schedv1.AccessControlEntryConfig{Host: "*", Principal: "User:42",
					Operation: schedv1.ACLOperations_READ, PermissionType: schedv1.ACLPermissionTypes_ALLOW},
			},
		}
	}()

	if err := cmd.Execute(); err != nil {
		t.Errorf("Cluster PatternType was not set to default value of PatternTypes_LITERAL")
	}
}

func TestDefaults1(t *testing.T) {
	DefaultsTest(t, true)
}

func TestDefaults2(t *testing.T) {
	DefaultsTest(t, false)
}

/*************** TEST command_cluster ***************/
// TODO: do this for all commands/subcommands... and for all common error messages
func Test_HandleError_NotLoggedIn(t *testing.T) {
	cmkApiMock := &cmkmock.ClustersCmkV2Api{
		ListCmkV2ClustersFunc: func(ctx context.Context) cmkv2.ApiListCmkV2ClustersRequest {
			return cmkv2.ApiListCmkV2ClustersRequest{}
		},
		ListCmkV2ClustersExecuteFunc: func(req cmkv2.ApiListCmkV2ClustersRequest) (cmkv2.CmkV2ClusterList, *http.Response, error) {
			return cmkv2.CmkV2ClusterList{}, nil, new(errors.NotLoggedInError)
		},
	}
	cmkClient := &cmkv2.APIClient{ClustersCmkV2Api: cmkApiMock}
	cmd := New(conf, cliMock.NewPreRunnerMock(nil, &ccloudv2.Client{CmkClient: cmkClient, AuthToken: "auth-token"}, nil, nil, conf), "test-client")
	cmd.PersistentFlags().CountP("verbose", "v", "Increase output verbosity")
	cmd.SetArgs([]string{"cluster", "list"})
	buf := new(bytes.Buffer)
	cmd.SetOutput(buf)

	err := cmd.Execute()
	require.Error(t, err)
	require.Equal(t, errors.NotLoggedInErrorMsg, err.Error())
}

/*************** TEST command_links ***************/
type testLink struct {
	name   string
	source string
}

var Links = []testLink{
	{
		name:   "test_link",
		source: "myhost:1234",
	},
}

func linkTestHelper(t *testing.T, argmaker func(testLink) []string, expector func(chan interface{}, testLink)) {
	expect := make(chan interface{})
	for _, link := range Links {
		cmd := newRestCmd(expect)
		cmd.SetArgs(argmaker(link))

		go expector(expect, link)

		if err := cmd.Execute(); err != nil {
			t.Errorf("error: %s", err)
			t.Fail()
			return
		}
	}
}

func TestListLinks(t *testing.T) {
	linkTestHelper(
		t,
		func(link testLink) []string {
			return []string{"link", "list"}
		},
		func(expect chan interface{}, link testLink) {
		},
	)
}

func TestListLinksWithTopics(t *testing.T) {
	linkTestHelper(
		t,
		func(link testLink) []string {
			return []string{"link", "list", "--include-topics"}
		},
		func(expect chan interface{}, link testLink) {
		},
	)
}

func TestDescribeLink(t *testing.T) {
	linkTestHelper(
		t,
		func(link testLink) []string {
			return []string{"link", "describe", link.name}
		},
		func(expect chan interface{}, link testLink) {
			expect <- cliMock.DescribeLinkMatcher{
				LinkName: link.name,
			}
			expect <- cliMock.DescribeLinkMatcher{
				LinkName: link.name,
			}
		},
	)
}

func TestDeleteLink(t *testing.T) {
	linkTestHelper(
		t,
		func(link testLink) []string {
			return []string{"link", "delete", link.name}
		},
		func(expect chan interface{}, link testLink) {
			expect <- cliMock.DeleteLinkMatcher{
				LinkName: link.name,
			}
		},
	)
}

func configMapWithJsonConfigValues() map[string]string {
	type child struct {
		Name        string `json:"name"`
		PatternType string `json:"patternType"`
		FilterType  string `json:"filterType"`
	}

	type aJson struct {
		Children []child `json:"children"`
	}

	jsonConfig, _ := json.Marshal(aJson{
		Children: []child{
			{
				Name:        "*",
				PatternType: "LITERAL",
				FilterType:  "INCLUDE",
			},
			{
				Name:        "*",
				PatternType: "LITERAL",
				FilterType:  "INCLUDE",
			},
		},
	})

	println(string(jsonConfig))

	return map[string]string{
		"key1": "val1",
		"key2": "val2",
		"key3": string(jsonConfig),
	}
}

func TestBatchAlterLink(t *testing.T) {
	const configFileName = "link-config.in"
	configs := configMapWithJsonConfigValues()
	path, err := createTestConfigFile(configFileName, configs)
	defer os.Remove(path)
	if err != nil {
		log.Fatalf("failed to create test config file: %v", err)
	}

	linkTestHelper(
		t,
		func(link testLink) []string {
			return []string{"link", "update", link.name, "--config-file", configFileName}
		},
		func(expect chan interface{}, link testLink) {
			expect <- cliMock.BatchUpdateLinkConfigMatcher{
				LinkName: link.name,
				Configs:  configs,
			}
		},
	)
}

func TestCreateLink(t *testing.T) {
	linkTestHelper(
		t,
		func(link testLink) []string {
			return []string{"link", "create", link.name, "--source-bootstrap-server", link.source, "--source-cluster-id", "id1"}
		},
		func(expect chan interface{}, link testLink) {
			expect <- cliMock.CreateLinkMatcher{
				LinkName:        link.name,
				ValidateLink:    true,
				ValidateOnly:    false,
				SourceClusterId: "id1",
				Configs:         map[string]string{"bootstrap.servers": link.source},
			}
		})
}

func TestCreateMirror(t *testing.T) {
	const configFileName = "mirror-topic-config.in"
	configs := configMapWithJsonConfigValues()
	path, err := createTestConfigFile(configFileName, configs)
	defer os.Remove(path)
	if err != nil {
		log.Fatalf("failed to create test config file: %v", err)
	}

	linkTestHelper(
		t,
		func(link testLink) []string {
			return []string{"mirror", "create", "topic-1", "--link", "link-1", "--replication-factor", "2", "--config-file", configFileName}
		},
		func(expect chan interface{}, link testLink) {
			expect <- cliMock.CreateMirrorMatcher{
				LinkName:        "link-1",
				SourceTopicName: "topic-1",
				MirrorTopicName: "topic-1",
				Configs:         configs,
			}
		},
	)
}

func TestCreateMirrorWithLinkPrefix(t *testing.T) {
	const configFileName, topicName, clusterLinkPrefix = "prefixed-mirror-topic-config.in", "topic-1", "src_"
	configs := configMapWithJsonConfigValues()
	path, err := createTestConfigFile(configFileName, configs)
	defer os.Remove(path)
	if err != nil {
		log.Fatalf("failed to create test config file: %v", err)
	}

	linkTestHelper(
		t,
		func(link testLink) []string {
			return []string{"mirror", "create", clusterLinkPrefix + topicName, "--link", "link-1", "--replication-factor", "2", "--config-file", configFileName, "--source-topic", topicName}
		},
		func(expect chan interface{}, link testLink) {
			expect <- cliMock.CreateMirrorMatcher{
				LinkName:        "link-1",
				SourceTopicName: topicName,
				MirrorTopicName: clusterLinkPrefix + topicName,
				Configs:         configs,
			}
		},
	)
}

func TestListAllMirror(t *testing.T) {
	linkTestHelper(
		t,
		func(link testLink) []string {
			return []string{"mirror", "list", "--mirror-status", "active"}
		},
		func(expect chan interface{}, link testLink) {
			expect <- cliMock.ListMirrorMatcher{
				Status: "active",
			}
		},
	)
}

func TestListMirror(t *testing.T) {
	linkTestHelper(
		t,
		func(link testLink) []string {
			return []string{"mirror", "list", "--link", "link-1", "--mirror-status", "active"}
		},
		func(expect chan interface{}, link testLink) {
			expect <- cliMock.ListMirrorMatcher{
				LinkName: "link-1",
				Status:   "active",
			}
		},
	)
}

func TestDescribeMirror(t *testing.T) {
	linkTestHelper(
		t,
		func(link testLink) []string {
			return []string{"mirror", "describe", "dest-topic-1", "--link", "link-1"}
		},
		func(expect chan interface{}, link testLink) {
			expect <- cliMock.DescribeMirrorMatcher{
				LinkName:        "link-1",
				MirrorTopicName: "dest-topic-1",
			}
		})
}

func TestPromoteMirror(t *testing.T) {
	linkTestHelper(
		t,
		func(link testLink) []string {
			return []string{"mirror", "promote", "dest-topic-1", "dest-topic-2", "dest-topic-3", "--link", "link-1"}
		},
		func(expect chan interface{}, link testLink) {
			expect <- cliMock.AlterMirrorMatcher{
				LinkName: "link-1",
				MirrorTopicNames: map[string]bool{
					"dest-topic-1": true,
					"dest-topic-2": true,
					"dest-topic-3": true,
				},
			}
		})
}

func TestFailoverMirror(t *testing.T) {
	linkTestHelper(
		t,
		func(link testLink) []string {
			return []string{"mirror", "failover", "dest-topic-1", "dest-topic-2", "dest-topic-3", "--link", "link-1"}
		},
		func(expect chan interface{}, link testLink) {
			expect <- cliMock.AlterMirrorMatcher{
				LinkName: "link-1",
				MirrorTopicNames: map[string]bool{
					"dest-topic-1": true,
					"dest-topic-2": true,
					"dest-topic-3": true,
				},
			}
		})
}

func TestPauseMirror(t *testing.T) {
	linkTestHelper(
		t,
		func(link testLink) []string {
			return []string{"mirror", "pause", "dest-topic-1", "dest-topic-2", "dest-topic-3", "--link", "link-1"}
		},
		func(expect chan interface{}, link testLink) {
			expect <- cliMock.AlterMirrorMatcher{
				LinkName: "link-1",
				MirrorTopicNames: map[string]bool{
					"dest-topic-1": true,
					"dest-topic-2": true,
					"dest-topic-3": true,
				},
			}
		})
}

func TestResumeMirror(t *testing.T) {
	linkTestHelper(
		t,
		func(link testLink) []string {
			return []string{"mirror", "resume", "dest-topic-1", "dest-topic-2", "dest-topic-3", "--link", "link-1"}
		},
		func(expect chan interface{}, link testLink) {
			expect <- cliMock.AlterMirrorMatcher{
				LinkName: "link-1",
				MirrorTopicNames: map[string]bool{
					"dest-topic-1": true,
					"dest-topic-2": true,
					"dest-topic-3": true,
				},
			}
		})
}

/*************** TEST command_consumer-group_cloud ***************/
type testConsumerGroup struct {
	groupId string
}

var ConsumerGroups = []testConsumerGroup{
	{
		groupId: "consumer-group-1",
	},
}

func TestListGroups(t *testing.T) {
	expect := make(chan interface{})
	cmd := newRestCmd(expect)
	args := []string{"consumer-group", "list"}
	cmd.SetArgs(args)

	CheckIfCmdErrors(t, cmd, args, false)

	for _, args := range [][]string{
		{"consumer-group", "list", "egg"},
	} {
		cmd.SetArgs(args)
		CheckIfCmdErrors(t, cmd, args, true)
	}
}

func TestDescribeGroup(t *testing.T) {
	expect := make(chan interface{})
	cmd := newRestCmd(expect)
	for _, consumerGroup := range ConsumerGroups {
		args := []string{"consumer-group", "describe", consumerGroup.groupId}
		cmd.SetArgs(args)
		go func() {
			expect <- cliMock.GroupMatcher{
				ConsumerGroupId: consumerGroup.groupId,
			}
		}()
		CheckIfCmdErrors(t, cmd, args, false)
	}

	for _, args := range [][]string{
		{"consumer-group", "describe"},
		{"consumer-group", "describe", "consumer-group-1", "egg"},
	} {
		cmd.SetArgs(args)
		CheckIfCmdErrors(t, cmd, args, true)
	}
}

func TestSummarizeLag(t *testing.T) {
	expect := make(chan interface{})
	cmd := newRestCmd(expect)
	for _, consumerGroup := range ConsumerGroups {
		args := []string{"consumer-group", "lag", "summarize", consumerGroup.groupId}
		cmd.SetArgs(args)
		go func() {
			expect <- cliMock.GroupMatcher{
				ConsumerGroupId: consumerGroup.groupId,
			}
		}()
		CheckIfCmdErrors(t, cmd, args, false)
	}

	for _, args := range [][]string{
		{"consumer-group", "lag", "summarize"},
		{"consumer-group", "lag", "summarize", "consumer-group-1", "egg"},
	} {
		cmd.SetArgs(args)
		CheckIfCmdErrors(t, cmd, args, true)
	}
}

func TestListLag(t *testing.T) {
	expect := make(chan interface{})
	cmd := newRestCmd(expect)
	for _, consumerGroup := range ConsumerGroups {
		args := []string{"consumer-group", "lag", "list", consumerGroup.groupId}
		cmd.SetArgs(args)
		go func() {
			expect <- cliMock.GroupMatcher{
				ConsumerGroupId: consumerGroup.groupId,
			}
		}()
		CheckIfCmdErrors(t, cmd, args, false)
	}

	for _, args := range [][]string{
		{"consumer-group", "lag", "list"},
		{"consumer-group", "lag", "list", "consumer-group-1", "egg"},
	} {
		cmd.SetArgs(args)
		CheckIfCmdErrors(t, cmd, args, true)
	}
}

type testPartitionLag struct {
	consumerGroupId string
	topicName       string
	partitionId     int32
}

var PartitionLags = []testPartitionLag{
	{
		consumerGroupId: "consumer-group-1",
		topicName:       "topic-1",
		partitionId:     1,
	},
}

func TestGetLag(t *testing.T) {
	expect := make(chan interface{})
	cmd := newRestCmd(expect)

	// testing that properly formatted commands don't return errors
	for _, lag := range PartitionLags {
		args := []string{"consumer-group", "lag", "get", lag.consumerGroupId, "--topic", lag.topicName, "--partition", strconv.Itoa(int(lag.partitionId))}
		cmd.SetArgs(args)
		go func() {
			expect <- cliMock.PartitionLagMatcher{
				ConsumerGroupId: lag.consumerGroupId,
				TopicName:       lag.topicName,
				PartitionId:     lag.partitionId,
			}
		}()
		CheckIfCmdErrors(t, cmd, args, false)
	}

	// testing that improperly formatted commands return errors
	for _, args := range [][]string{
		{"consumer-group", "lag", "get"},
		{"consumer-group", "lag", "get", "consumer-group-1", "egg"},
		{"consumer-group", "lag", "get", "consumer-group-1", "consumer-group-1", "--topic", "topic-1"},
	} {
		cmd.SetArgs(args)
		CheckIfCmdErrors(t, cmd, args, true)
	}
}

// Executes cmd, checks if there was an error. Test will fail if there was an error but expectErr was false,
// or expectErr was true but no error was returned.
func CheckIfCmdErrors(t *testing.T, cmd *cobra.Command, args []string, expectErr bool) {
	err := cmd.Execute()
	var errMsg string
	if expectErr == true && err == nil {
		errMsg = "expected error from executing " + strings.Join(args, " ") + " but received none"
	} else if expectErr == false && err != nil {
		errMsg = "error: " + err.Error()
	}
	if errMsg != "" {
		t.Errorf(errMsg)
		t.Fail()
	}
}

/*************** TEST setup/helpers ***************/
func newMockCmd(kafkaExpect chan interface{}, kafkaRestExpect chan interface{}, enableREST bool) *cobra.Command {
	client := &ccloud.Client{
		Kafka: cliMock.NewKafkaMock(kafkaExpect),
		User: &mock.User{
			DescribeFunc: func(_ context.Context, _ *orgv1.User) (*orgv1.User, error) {
				return &orgv1.User{
					Email: "csreesangkom@confluent.io",
				}, nil
			},
			GetServiceAccountsFunc: func(arg0 context.Context) ([]*orgv1.User, error) {
				return []*orgv1.User{
					{
						Id:          serviceAccountId,
						ResourceId:  serviceAccountResourceId,
						ServiceName: serviceAccountName,
					},
					{
						Id:          42,
						ResourceId:  "sa-42",
						ServiceName: serviceAccountName,
					},
				}, nil
			},
			ListFunc: func(_ context.Context) ([]*orgv1.User, error) {
				return []*orgv1.User{
					{
						Id:         userId,
						ResourceId: userResourceId,
					},
				}, nil
			},
		},
		EnvironmentMetadata: &mock.EnvironmentMetadata{
			GetFunc: func(ctx context.Context) ([]*schedv1.CloudMetadata, error) {
				return []*schedv1.CloudMetadata{{
					Id:       "aws",
					Accounts: []*schedv1.AccountMetadata{{Id: "account-xyz"}},
					Regions:  []*schedv1.Region{{IsSchedulable: true, Id: "us-west-2"}},
				}}, nil
			},
		},
	}
	provider := (pcmd.KafkaRESTProvider)(func() (*pcmd.KafkaREST, error) {
		if enableREST {
			restMock := krsdk.NewAPIClient(&krsdk.Configuration{BasePath: "/dummy-base-path"})
			restMock.ACLV3Api = cliMock.NewACLMock()
			restMock.TopicV3Api = cliMock.NewTopicMock()
			restMock.PartitionV3Api = cliMock.NewPartitionMock(kafkaRestExpect)
			restMock.ReplicaV3Api = cliMock.NewReplicaMock()
			restMock.ConfigsV3Api = cliMock.NewConfigsMock()
			restMock.ClusterLinkingV3Api = cliMock.NewClusterLinkingMock(kafkaRestExpect)
			restMock.ConsumerGroupV3Api = cliMock.NewConsumerGroupMock(kafkaRestExpect)
			restMock.ReplicaStatusApi = cliMock.NewReplicaStatusMock()
			ctx := context.WithValue(context.Background(), krsdk.ContextAccessToken, "dummy-bearer-token")
			kafkaREST := pcmd.NewKafkaREST(restMock, ctx)
			return kafkaREST, nil
		}
		return nil, nil
	})
	cmd := New(conf, cliMock.NewPreRunnerMock(client, nil, nil, &provider, conf), "test-client")
	cmd.PersistentFlags().CountP("verbose", "v", "Increase output verbosity")
	return cmd
}

func newCmd(kafkaExpect chan interface{}, enableREST bool) *cobra.Command {
	return newMockCmd(kafkaExpect, make(chan interface{}), enableREST)
}

func newRestCmd(restExpect chan interface{}) *cobra.Command {
	return newMockCmd(make(chan interface{}), restExpect, true)
}

func init() {
	conf = v1.AuthenticatedCloudConfigMock()
}
