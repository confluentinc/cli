package iam

import (
	"context"
	net_http "net/http"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/log"
	cliMock "github.com/confluentinc/cli/mock"
	"github.com/confluentinc/mds-sdk-go"
	"github.com/confluentinc/mds-sdk-go/mock"
)

type AclTestSuite struct {
	suite.Suite
	conf *config.Config
	kafkaApi *mock.KafkaACLManagementApi
}

/*************** TEST command_acl ***************/
var mdsResourcePatterns = []struct {
	args    []string
	pattern mds.KafkaResourcePattern
}{
	{
		args: []string{"--cluster-scope"},
		pattern: mds.KafkaResourcePattern{ResourceType: mds.ACL_RESOURCE_TYPE_CLUSTER, Name: "kafka-cluster",
			PatternType: mds.PATTERN_TYPE_LITERAL},
	},
	{
		args: []string{"--topic", "test-topic"},
		pattern: mds.KafkaResourcePattern{ResourceType: mds.ACL_RESOURCE_TYPE_TOPIC, Name: "test-topic",
			PatternType: mds.PATTERN_TYPE_LITERAL},
	},
	{
		args: []string{"--topic", "test-topic", "--prefix"},
		pattern: mds.KafkaResourcePattern{ResourceType: mds.ACL_RESOURCE_TYPE_TOPIC, Name: "test-topic",
			PatternType: mds.PATTERN_TYPE_PREFIXED},
	},
	{
		args: []string{"--consumer-group", "test-group"},
		pattern: mds.KafkaResourcePattern{ResourceType: mds.ACL_RESOURCE_TYPE_GROUP, Name: "test-group",
			PatternType: mds.PATTERN_TYPE_LITERAL},
	},
	{
		args: []string{"--consumer-group", "test-group", "--prefix"},
		pattern: mds.KafkaResourcePattern{ResourceType: mds.ACL_RESOURCE_TYPE_GROUP, Name: "test-group",
			PatternType: mds.PATTERN_TYPE_PREFIXED},
	},
	{
		args: []string{"--transactional-id", "test-transactional-id"},
		pattern: mds.KafkaResourcePattern{ResourceType: mds.ACL_RESOURCE_TYPE_TRANSACTIONAL_ID, Name: "test-transactional-id",
			PatternType: mds.PATTERN_TYPE_LITERAL},
	},
	{
		args: []string{"--transactional-id", "test-transactional-id", "--prefix"},
		pattern: mds.KafkaResourcePattern{ResourceType: mds.ACL_RESOURCE_TYPE_TRANSACTIONAL_ID, Name: "test-transactional-id",
			PatternType: mds.PATTERN_TYPE_PREFIXED},
	},
	{
		args: []string{"--prefix", "--topic", "test-topic"},
		pattern: mds.KafkaResourcePattern{ResourceType: mds.ACL_RESOURCE_TYPE_TOPIC, Name: "test-topic",
			PatternType: mds.PATTERN_TYPE_PREFIXED},
	},
}

var mdsAclEntries = []struct {
	args  []string
	entry mds.AccessControlEntry
}{
	{
		args: []string{"--allow", "--principal", "User:Bob", "--operation", "read"},
		entry: mds.AccessControlEntry{PermissionType: mds.ACL_PERMISSION_TYPE_ALLOW,
			Principal: "User:Bob", Operation: mds.ACL_OPERATION_READ, Host: "*"},
	},
	{
		args: []string{"--deny", "--principal", "User:Bob", "--operation", "read"},
		entry: mds.AccessControlEntry{PermissionType: mds.ACL_PERMISSION_TYPE_DENY,
			Principal: "User:Bob", Operation: mds.ACL_OPERATION_READ, Host: "*"},
	},
	{
		args: []string{"--allow", "--principal", "User:Bob", "--operation", "write"},
		entry: mds.AccessControlEntry{PermissionType: mds.ACL_PERMISSION_TYPE_ALLOW,
			Principal: "User:Bob", Operation: mds.ACL_OPERATION_WRITE, Host: "*"},
	},
	{
		args: []string{"--deny", "--principal", "User:Bob", "--operation", "write"},
		entry: mds.AccessControlEntry{PermissionType: mds.ACL_PERMISSION_TYPE_DENY,
			Principal: "User:Bob", Operation: mds.ACL_OPERATION_WRITE, Host: "*"},
	},
	{
		args: []string{"--allow", "--principal", "User:Bob", "--operation", "create"},
		entry: mds.AccessControlEntry{PermissionType: mds.ACL_PERMISSION_TYPE_ALLOW,
			Principal: "User:Bob", Operation: mds.ACL_OPERATION_CREATE, Host: "*"},
	},
	{
		args: []string{"--deny", "--principal", "User:Bob", "--operation", "create"},
		entry: mds.AccessControlEntry{PermissionType: mds.ACL_PERMISSION_TYPE_DENY,
			Principal: "User:Bob", Operation: mds.ACL_OPERATION_CREATE, Host: "*"},
	},
	{
		args: []string{"--allow", "--principal", "User:Bob", "--operation", "delete"},
		entry: mds.AccessControlEntry{PermissionType: mds.ACL_PERMISSION_TYPE_ALLOW,
			Principal: "User:Bob", Operation: mds.ACL_OPERATION_DELETE, Host: "*"},
	},
	{
		args: []string{"--deny", "--principal", "User:Bob", "--operation", "delete"},
		entry: mds.AccessControlEntry{PermissionType: mds.ACL_PERMISSION_TYPE_DENY,
			Principal: "User:Bob", Operation: mds.ACL_OPERATION_DELETE, Host: "*"},
	},
	{
		args: []string{"--allow", "--principal", "User:Bob", "--operation", "alter"},
		entry: mds.AccessControlEntry{PermissionType: mds.ACL_PERMISSION_TYPE_ALLOW,
			Principal: "User:Bob", Operation: mds.ACL_OPERATION_ALTER, Host: "*"},
	},
	{
		args: []string{"--deny", "--principal", "User:Bob", "--operation", "alter"},
		entry: mds.AccessControlEntry{PermissionType: mds.ACL_PERMISSION_TYPE_DENY,
			Principal: "User:Bob", Operation: mds.ACL_OPERATION_ALTER, Host: "*"},
	},
	{
		args: []string{"--allow", "--principal", "User:Bob", "--operation", "describe"},
		entry: mds.AccessControlEntry{PermissionType: mds.ACL_PERMISSION_TYPE_ALLOW,
			Principal: "User:Bob", Operation: mds.ACL_OPERATION_DESCRIBE, Host: "*"},
	},
	{
		args: []string{"--deny", "--principal", "User:Bob", "--operation", "describe"},
		entry: mds.AccessControlEntry{PermissionType: mds.ACL_PERMISSION_TYPE_DENY,
			Principal: "User:Bob", Operation: mds.ACL_OPERATION_DESCRIBE, Host: "*"},
	},
	{
		args: []string{"--allow", "--principal", "User:Bob", "--operation", "cluster-action"},
		entry: mds.AccessControlEntry{PermissionType: mds.ACL_PERMISSION_TYPE_ALLOW,
			Principal: "User:Bob", Operation: mds.ACL_OPERATION_CLUSTER_ACTION, Host: "*"},
	},
	{
		args: []string{"--deny", "--principal", "User:Bob", "--operation", "cluster-action"},
		entry: mds.AccessControlEntry{PermissionType: mds.ACL_PERMISSION_TYPE_DENY,
			Principal: "User:Bob", Operation: mds.ACL_OPERATION_CLUSTER_ACTION, Host: "*"},
	},
	{
		args: []string{"--allow", "--principal", "User:Bob", "--operation", "describe-configs"},
		entry: mds.AccessControlEntry{PermissionType: mds.ACL_PERMISSION_TYPE_ALLOW,
			Principal: "User:Bob", Operation: mds.ACL_OPERATION_DESCRIBE_CONFIGS, Host: "*"},
	},
	{
		args: []string{"--deny", "--principal", "User:Bob", "--operation", "describe-configs"},
		entry: mds.AccessControlEntry{PermissionType: mds.ACL_PERMISSION_TYPE_DENY,
			Principal: "User:Bob", Operation: mds.ACL_OPERATION_DESCRIBE_CONFIGS, Host: "*"},
	},
	{
		args: []string{"--allow", "--principal", "User:Bob", "--operation", "alter-configs"},
		entry: mds.AccessControlEntry{PermissionType: mds.ACL_PERMISSION_TYPE_ALLOW,
			Principal: "User:Bob", Operation: mds.ACL_OPERATION_ALTER_CONFIGS, Host: "*"},
	},
	{
		args: []string{"--deny", "--principal", "User:Bob", "--operation", "alter-configs"},
		entry: mds.AccessControlEntry{PermissionType: mds.ACL_PERMISSION_TYPE_DENY,
			Principal: "User:Bob", Operation: mds.ACL_OPERATION_ALTER_CONFIGS, Host: "*"},
	},
	{
		args: []string{"--allow", "--principal", "User:Bob", "--operation", "idempotent-write"},
		entry: mds.AccessControlEntry{PermissionType: mds.ACL_PERMISSION_TYPE_ALLOW,
			Principal: "User:Bob", Operation: mds.ACL_OPERATION_IDEMPOTENT_WRITE, Host: "*"},
	},
	{
		args: []string{"--deny", "--principal", "User:Bob", "--operation", "idempotent-write"},
		entry: mds.AccessControlEntry{PermissionType: mds.ACL_PERMISSION_TYPE_DENY,
			Principal: "User:Bob", Operation: mds.ACL_OPERATION_IDEMPOTENT_WRITE, Host: "*"},
	},
}

func (suite *AclTestSuite) SetupSuite() {
	_ = os.Setenv("XX_FLAG_CENTRALIZED_ACL_ENABLE", "true")
	suite.conf = config.New()
	suite.conf.CLIName = "confluent"
	suite.conf.Logger = log.New()
	suite.conf.AuthURL = "http://test"
	suite.conf.AuthToken = "T0k3n"
}

func (suite *AclTestSuite) newMockIamCmd(expect chan interface{}) *cobra.Command {
	suite.kafkaApi = &mock.KafkaACLManagementApi{
		AddAclBindingFunc:  func(ctx context.Context, createAclRequest mds.CreateAclRequest) (*net_http.Response, error) {
			assert.Equal(suite.T(), createAclRequest, <-expect, "")
			return nil, nil
		},
		RemoveAclBindingsFunc: func(ctx context.Context, aclFilterRequest mds.AclFilterRequest) ([]mds.AclBinding, *net_http.Response, error){
			assert.Equal(suite.T(), aclFilterRequest, <-expect, "")
			return nil, nil, nil
		},
		SearchAclBindingFunc: func(ctx context.Context, aclFilterRequest mds.AclFilterRequest) ([]mds.AclBinding, *net_http.Response, error){
			assert.Equal(suite.T(), aclFilterRequest, <-expect, "")
			return nil, nil, nil
		},
	}
	mdsClient := mds.NewAPIClient(mds.NewConfiguration())
	mdsClient.KafkaACLManagementApi = suite.kafkaApi
	return New(&cliMock.Commander{}, suite.conf, mdsClient)
}

func (suite *AclTestSuite) TestMdsCreateACL() {
	expect := make(chan interface{})
	for _, mdsResourcePattern := range mdsResourcePatterns {
		args := append([]string{"acl", "create", "--kafka-cluster-id", "testcluster"}, mdsResourcePattern.args...)
		for _, mdsAclEntry := range mdsAclEntries {
			cmd := suite.newMockIamCmd(expect)
			cmd.SetArgs(append(args, mdsAclEntry.args...))

			go func() {
				expect <- mds.CreateAclRequest {
					Scope: mds.KafkaScope {
						Clusters: mds.KafkaScopeClusters{
							KafkaCluster: "testcluster",
						},
					},
					AclBinding: mds.AclBinding{Pattern: mdsResourcePattern.pattern, Entry: mdsAclEntry.entry},
				}
			}()

			err := cmd.Execute()
			req := require.New(suite.T())
			req.Nil(err)
		}
	}
}

func TestAclTestSuite(t *testing.T) {
	suite.Run(t, new(AclTestSuite))
}

/*
func TestMdsDeleteACL(t *testing.T) {
	expect := make(chan interface{})
	for _, resource := range mdsResourcePatterns {
		args := append([]string{"acl", "delete"}, resource.args...)
		for _, entry := range mdsAclEntries {
			cmd := NewMdsCMD(expect)
			cmd.SetArgs(append(args, entry.args...))

			go func() {
				expect <- convertToAclFilterRequest(&mds.AclBinding{Pattern: resource.pattern, Entry: entry.entry})
			}()

			if err := cmd.Execute(); err != nil {
				t.Errorf("error: %s", err)
			}
		}
	}
}

func TestMdsListResourceACL(t *testing.T) {
	expect := make(chan interface{})
	for _, resource := range mdsResourcePatterns {
		cmd := NewMdsCMD(expect)
		cmd.SetArgs(append([]string{"acl", "list"}, resource.args...))

		go func() {
			expect <- convertToAclFilterRequest(&kafkav1.ACLBinding{Pattern: resource.pattern, Entry: &mds.AccessControlEntry{}})
		}()

		if err := cmd.Execute(); err != nil {
			t.Errorf("error: %s", err)
		}
	}
}

func TestMdsListPrincipalACL(t *testing.T) {
	expect := make(chan interface{})
	for _, entry := range mdsAclEntries {
		cmd := NewMdsCMD(expect)
		cmd.SetArgs(append([]string{"acl", "list", "--service-account-id"}, strings.TrimPrefix(entry.entry.Principal, "User:")))

		go func() {
			expect <- convertToAclFilterRequest(&kafkav1.ACLBinding{Entry: &mds.AccessControlEntry{Principal: entry.entry.Principal}})
		}()

		if err := cmd.Execute(); err != nil {
			t.Errorf("error: %s", err)
		}
	}
}

func TestMdsListResourcePrincipalFilterACL(t *testing.T) {
	expect := make(chan interface{})
	for _, resource := range mdsResourcePatterns {
		args := append([]string{"acl", "list"}, resource.args...)
		for _, entry := range mdsAclEntries {
			cmd := NewMdsCMD(expect)
			cmd.SetArgs(append(args, "--service-account-id", strings.TrimPrefix(entry.entry.Principal, "User:")))

			go func() {
				expect <- convertToAclFilterRequest(&kafkav1.ACLBinding{Pattern: resource.pattern, Entry: entry.entry})
			}()

			if err := cmd.Execute(); err != nil {
				t.Errorf("error: %s", err)
			}
		}
	}
}

func TestMdsMultipleResourceACL(t *testing.T) {
	expect := "exactly one of cluster-scope, consumer-group, topic, transactional-id must be set"
	args := []string{"acl", "create", "--allow", "--operation", "read", "--principal", "User:Bob",
		"--topic", "resource1", "--consumer-group", "resource2"}

	cmd := NewMdsCMD(nil)
	cmd.SetArgs(args)

	err := cmd.Execute()
	if !strings.Contains(err.Error(), expect) {
		t.Errorf("expected: %s got: %s", expect, err.Error())
	}
}

func TestMdsDefaults(t *testing.T) {
	expect := make(chan interface{})
	cmd := NewMdsCMD(expect)
	cmd.SetArgs([]string{"acl", "create", "--allow", "--principal", "User:Bob",
		"--operation", "read", "--topic", "dan"})
	go func() {
		expect <- &kafkav1.ACLBinding{
			Pattern: &kafkav1.ResourcePatternConfig{ResourceType: kafkav1.ResourceTypes_TOPIC, Name: "dan",
				PatternType: kafkav1.PatternTypes_LITERAL},
			Entry: &mds.AccessControlEntry{Host: "*", Principal: "User:Bob",
				Operation: mds.ACL_OPERATION_READ, PermissionType: mds.ACL_PERMISSION_TYPE_ALLOW},
		}
	}()

	if err := cmd.Execute(); err != nil {
		t.Errorf("Topic PatternType was not set to default value of PatternTypes_LITERAL")
	}

	cmd = NewMdsCMD(expect)
	cmd.SetArgs([]string{"acl", "create", "--cluster-scope", "--allow", "--principal", "User:Bob",
		"--operation", "read"})

	go func() {
		expect <- &kafkav1.ACLBinding{
			Pattern: &kafkav1.ResourcePatternConfig{ResourceType: kafkav1.ResourceTypes_CLUSTER, Name: "kafka-cluster",
				PatternType: kafkav1.PatternTypes_LITERAL},
			Entry: &mds.AccessControlEntry{Host: "*", Principal: "User:Bob",
				Operation: mds.ACL_OPERATION_READ, PermissionType: mds.ACL_PERMISSION_TYPE_ALLOW},
		}
	}()

	if err := cmd.Execute(); err != nil {
		t.Errorf("Cluster PatternType was not set to default value of PatternTypes_LITERAL")
	}
}

// TODO: do this for all commands/subcommands... and for all common error messages
func Test_HandleError_NotLoggedIn(t *testing.T) {
	kafka := &mock.Kafka{
		ListFunc: func(ctx context.Context, cluster *kafkav1.KafkaCluster) ([]*kafkav1.KafkaCluster, error) {
			return nil, errors.ErrNotLoggedIn
		},
	}
	cmd := kafka2.New(&cliMock.Commander{}, kafka2.conf, kafka, &pcmd.ConfigHelper{Config: kafka2.conf, Client: &ccloud.Client{Kafka: kafka}})
	cmd.PersistentFlags().CountP("verbose", "v", "Increase output verbosity")
	cmd.SetArgs(append([]string{"cluster", "list"}))
	buf := new(bytes.Buffer)
	cmd.SetOutput(buf)

	err := cmd.Execute()
	want := "You must login to run that command."
	if err.Error() != want {
		t.Errorf("unexpected output, got %s, want %s", err, want)
	}
}

func NewMdsCMD(expect chan interface{}) *cobra.Command {
	kafka := cliMock.NewKafkaMock(expect)
	cmd := kafka2.NewMdsCommand(&cliMock.Commander{}, kafka2.conf, kafka, &pcmd.ConfigHelper{Config: kafka2.conf, Client: &ccloud.Client{Kafka: kafka}})
	cmd.PersistentFlags().CountP("verbose", "v", "Increase output verbosity")

	return cmd
}
*/
