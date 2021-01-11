package kafka

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"

	linkv1 "github.com/confluentinc/cc-structs/kafka/clusterlink/v1"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/ccloud-sdk-go"
	"github.com/confluentinc/ccloud-sdk-go/mock"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
	cliMock "github.com/confluentinc/cli/mock"
	krsdk "github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
)

var conf *v3.Config

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
		args: []string{"--allow", "--service-account", "42", "--operation", "read"},
		entries: []*schedv1.AccessControlEntryConfig{
			{
				PermissionType: schedv1.ACLPermissionTypes_ALLOW,
				Principal:      "User:42", Operation: schedv1.ACLOperations_READ, Host: "*",
			},
		},
	},
	{
		args: []string{"--deny", "--service-account", "42", "--operation", "read"},
		entries: []*schedv1.AccessControlEntryConfig{
			{
				PermissionType: schedv1.ACLPermissionTypes_DENY,
				Principal:      "User:42", Operation: schedv1.ACLOperations_READ, Host: "*",
			},
		},
	},
	{
		args: []string{"--allow", "--service-account", "42", "--operation", "write"},
		entries: []*schedv1.AccessControlEntryConfig{
			{
				PermissionType: schedv1.ACLPermissionTypes_ALLOW,
				Principal:      "User:42", Operation: schedv1.ACLOperations_WRITE, Host: "*",
			},
		},
	},
	{
		args: []string{"--deny", "--service-account", "42", "--operation", "write"},
		entries: []*schedv1.AccessControlEntryConfig{
			{
				PermissionType: schedv1.ACLPermissionTypes_DENY,
				Principal:      "User:42", Operation: schedv1.ACLOperations_WRITE, Host: "*",
			},
		},
	},
	{
		args: []string{"--allow", "--service-account", "42", "--operation", "create"},
		entries: []*schedv1.AccessControlEntryConfig{
			{
				PermissionType: schedv1.ACLPermissionTypes_ALLOW,
				Principal:      "User:42", Operation: schedv1.ACLOperations_CREATE, Host: "*",
			},
		},
	},
	{
		args: []string{"--deny", "--service-account", "42", "--operation", "create"},
		entries: []*schedv1.AccessControlEntryConfig{
			{
				PermissionType: schedv1.ACLPermissionTypes_DENY,
				Principal:      "User:42", Operation: schedv1.ACLOperations_CREATE, Host: "*",
			},
		},
	},
	{
		args: []string{"--allow", "--service-account", "42", "--operation", "delete"},
		entries: []*schedv1.AccessControlEntryConfig{
			{
				PermissionType: schedv1.ACLPermissionTypes_ALLOW,
				Principal:      "User:42", Operation: schedv1.ACLOperations_DELETE, Host: "*",
			},
		},
	},
	{
		args: []string{"--deny", "--service-account", "42", "--operation", "delete"},
		entries: []*schedv1.AccessControlEntryConfig{
			{
				PermissionType: schedv1.ACLPermissionTypes_DENY,
				Principal:      "User:42", Operation: schedv1.ACLOperations_DELETE, Host: "*",
			},
		},
	},
	{
		args: []string{"--allow", "--service-account", "42", "--operation", "alter"},
		entries: []*schedv1.AccessControlEntryConfig{
			{
				PermissionType: schedv1.ACLPermissionTypes_ALLOW,
				Principal:      "User:42", Operation: schedv1.ACLOperations_ALTER, Host: "*",
			},
		},
	},
	{
		args: []string{"--deny", "--service-account", "42", "--operation", "alter"},
		entries: []*schedv1.AccessControlEntryConfig{
			{
				PermissionType: schedv1.ACLPermissionTypes_DENY,
				Principal:      "User:42", Operation: schedv1.ACLOperations_ALTER, Host: "*",
			},
		},
	},
	{
		args: []string{"--allow", "--service-account", "42", "--operation", "describe"},
		entries: []*schedv1.AccessControlEntryConfig{
			{
				PermissionType: schedv1.ACLPermissionTypes_ALLOW,
				Principal:      "User:42", Operation: schedv1.ACLOperations_DESCRIBE, Host: "*",
			},
		},
	},
	{
		args: []string{"--deny", "--service-account", "42", "--operation", "describe"},
		entries: []*schedv1.AccessControlEntryConfig{
			{
				PermissionType: schedv1.ACLPermissionTypes_DENY,
				Principal:      "User:42", Operation: schedv1.ACLOperations_DESCRIBE, Host: "*",
			},
		},
	},
	{
		args: []string{"--allow", "--service-account", "42", "--operation", "cluster-action"},
		entries: []*schedv1.AccessControlEntryConfig{
			{
				PermissionType: schedv1.ACLPermissionTypes_ALLOW,
				Principal:      "User:42", Operation: schedv1.ACLOperations_CLUSTER_ACTION, Host: "*",
			},
		},
	},
	{
		args: []string{"--deny", "--service-account", "42", "--operation", "cluster-action"},
		entries: []*schedv1.AccessControlEntryConfig{
			{
				PermissionType: schedv1.ACLPermissionTypes_DENY,
				Principal:      "User:42", Operation: schedv1.ACLOperations_CLUSTER_ACTION, Host: "*",
			},
		},
	},
	{
		args: []string{"--allow", "--service-account", "42", "--operation", "describe-configs"},
		entries: []*schedv1.AccessControlEntryConfig{
			{
				PermissionType: schedv1.ACLPermissionTypes_ALLOW,
				Principal:      "User:42", Operation: schedv1.ACLOperations_DESCRIBE_CONFIGS, Host: "*",
			},
		},
	},
	{
		args: []string{"--deny", "--service-account", "42", "--operation", "describe-configs"},
		entries: []*schedv1.AccessControlEntryConfig{
			{
				PermissionType: schedv1.ACLPermissionTypes_DENY,
				Principal:      "User:42", Operation: schedv1.ACLOperations_DESCRIBE_CONFIGS, Host: "*",
			},
		},
	},
	{
		args: []string{"--allow", "--service-account", "42", "--operation", "alter-configs"},
		entries: []*schedv1.AccessControlEntryConfig{
			{
				PermissionType: schedv1.ACLPermissionTypes_ALLOW,
				Principal:      "User:42", Operation: schedv1.ACLOperations_ALTER_CONFIGS, Host: "*",
			},
		},
	},
	{
		args: []string{"--deny", "--service-account", "42", "--operation", "alter-configs"},
		entries: []*schedv1.AccessControlEntryConfig{
			{
				PermissionType: schedv1.ACLPermissionTypes_DENY,
				Principal:      "User:42", Operation: schedv1.ACLOperations_ALTER_CONFIGS, Host: "*",
			},
		},
	},
	{
		args: []string{"--allow", "--service-account", "42", "--operation", "idempotent-write"},
		entries: []*schedv1.AccessControlEntryConfig{
			{
				PermissionType: schedv1.ACLPermissionTypes_ALLOW,
				Principal:      "User:42", Operation: schedv1.ACLOperations_IDEMPOTENT_WRITE, Host: "*",
			},
		},
	},
	{
		args: []string{"--deny", "--service-account", "42", "--operation", "idempotent-write"},
		entries: []*schedv1.AccessControlEntryConfig{
			{
				PermissionType: schedv1.ACLPermissionTypes_DENY,
				Principal:      "User:42", Operation: schedv1.ACLOperations_IDEMPOTENT_WRITE, Host: "*",
			},
		},
	},
	{
		args: []string{"--deny", "--service-account", "42", "--operation", "alter-configs", "--operation", "idempotent-write", "--operation", "create"},
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
		cmd.SetArgs(append([]string{"acl", "list", "--service-account"}, strings.TrimPrefix(entry.Principal, "User:")))

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
			cmd.SetArgs(append(args, "--service-account", strings.TrimPrefix(entry.Principal, "User:")))

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
	args := []string{"acl", "create", "--allow", "--operation", "read", "--service-account", "42",
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
	cmd.SetArgs([]string{"acl", "create", "--allow", "--service-account", "42",
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
	cmd.SetArgs([]string{"acl", "create", "--cluster-scope", "--allow", "--service-account", "42",
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
	kafka := &mock.Kafka{
		ListFunc: func(ctx context.Context, cluster *schedv1.KafkaCluster) ([]*schedv1.KafkaCluster, error) {
			return nil, &errors.NotLoggedInError{CLIName: "ccloud"}
		},
	}
	client := &ccloud.Client{Kafka: kafka}
	cmd := New(false, conf.CLIName, cliMock.NewPreRunnerMock(client, nil, nil, conf), log.New(), "test-client", &cliMock.ServerSideCompleter{})
	cmd.PersistentFlags().CountP("verbose", "v", "Increase output verbosity")
	cmd.SetArgs(append([]string{"cluster", "list"}))
	buf := new(bytes.Buffer)
	cmd.SetOutput(buf)

	err := cmd.Execute()
	want := errors.NotLoggedInErrorMsg
	require.Error(t, err)
	require.Equal(t, want, err.Error())
	errors.VerifyErrorAndSuggestions(require.New(t), err, errors.NotLoggedInErrorMsg, fmt.Sprintf(errors.NotLoggedInSuggestions, "ccloud"))
}

/*************** TEST command_links ***************/
type testLink struct {
	name       string
	source     string
	alterKey   string
	alterValue string
}

var Links = []testLink{
	{
		name:       "test_link",
		source:     "myhost:1234",
		alterKey:   "retention.ms",
		alterValue: "1234567890",
	},
}

func linkTestHelper(t *testing.T, argmaker func(testLink) []string, expector func(chan interface{}, testLink)) {
	expect := make(chan interface{})
	for _, link := range Links {
		cmd := newCmd(expect, false)
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
			expect <- link.name
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
			expect <- link.name
		},
	)
}

func TestAlterLink(t *testing.T) {
	linkTestHelper(
		t,
		func(link testLink) []string {
			return []string{"link", "update", link.name, "--config", fmt.Sprintf("%s=%s", link.alterKey, link.alterValue)}
		},
		func(expect chan interface{}, link testLink) {
			expect <- link.name
			expect <- &linkv1.LinkProperties{Properties: map[string]string{link.alterKey: link.alterValue}}
		},
	)
}

func TestCreateLink(t *testing.T) {
	linkTestHelper(
		t,
		func(link testLink) []string {
			return []string{"link", "create", link.name, "--source_cluster", link.source}
		},
		func(expect chan interface{}, link testLink) {
			expect <- &linkv1.ClusterLink{
				LinkName: link.name,
				Configs: map[string]string{
					"bootstrap.servers": link.source,
				},
			}
		})
}

/*************** TEST setup/helpers ***************/
func newCmd(expect chan interface{}, enableREST bool) *cobra.Command {
	client := &ccloud.Client{
		Kafka: cliMock.NewKafkaMock(expect),
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
			restMock.ACLApi = cliMock.NewACLMock()
			restMock.TopicApi = cliMock.NewTopicMock()
			restMock.PartitionApi = cliMock.NewPartitionMock()
			restMock.ReplicaApi = cliMock.NewReplicaMock()
			restMock.ConfigsApi = cliMock.NewConfigsMock()
			ctx := context.WithValue(context.Background(), krsdk.ContextAccessToken, "dummy-bearer-token")
			kafkaREST := pcmd.NewKafkaREST(restMock, ctx)
			return kafkaREST, nil
		}
		return nil, nil
	})

	cmd := New(false, conf.CLIName, cliMock.NewPreRunnerMock(client, nil, &provider, conf), log.New(), "test-client", &cliMock.ServerSideCompleter{})
	cmd.PersistentFlags().CountP("verbose", "v", "Increase output verbosity")

	return cmd
}

func init() {
	conf = v3.AuthenticatedCloudConfigMock()
}
