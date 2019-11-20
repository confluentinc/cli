package iam

import (
	"context"
	net_http "net/http"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/confluentinc/mds-sdk-go"
	"github.com/confluentinc/mds-sdk-go/mock"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/update"
	cliMock "github.com/confluentinc/cli/mock"
)

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
}

var mdsAclEntries = []struct {
	args  []string
	entry mds.AccessControlEntry
}{
	{
		args: []string{"--allow", "--principal", "User:42", "--operation", "read"},
		entry: mds.AccessControlEntry{PermissionType: mds.ACL_PERMISSION_TYPE_ALLOW,
			Principal: "User:42", Operation: mds.ACL_OPERATION_READ, Host: "*"},
	},
	{
		args: []string{"--deny", "--principal", "User:42", "--host", "testhost", "--operation", "read"},
		entry: mds.AccessControlEntry{PermissionType: mds.ACL_PERMISSION_TYPE_DENY,
			Principal: "User:42", Operation: mds.ACL_OPERATION_READ, Host: "testhost"},
	},
	{
		args: []string{"--allow", "--principal", "User:42", "--host", "*", "--operation", "write"},
		entry: mds.AccessControlEntry{PermissionType: mds.ACL_PERMISSION_TYPE_ALLOW,
			Principal: "User:42", Operation: mds.ACL_OPERATION_WRITE, Host: "*"},
	},
	{
		args: []string{"--deny", "--principal", "User:42", "--operation", "write"},
		entry: mds.AccessControlEntry{PermissionType: mds.ACL_PERMISSION_TYPE_DENY,
			Principal: "User:42", Operation: mds.ACL_OPERATION_WRITE, Host: "*"},
	},
	{
		args: []string{"--allow", "--principal", "User:42", "--operation", "create"},
		entry: mds.AccessControlEntry{PermissionType: mds.ACL_PERMISSION_TYPE_ALLOW,
			Principal: "User:42", Operation: mds.ACL_OPERATION_CREATE, Host: "*"},
	},
	{
		args: []string{"--deny", "--principal", "User:42", "--operation", "create"},
		entry: mds.AccessControlEntry{PermissionType: mds.ACL_PERMISSION_TYPE_DENY,
			Principal: "User:42", Operation: mds.ACL_OPERATION_CREATE, Host: "*"},
	},
	{
		args: []string{"--allow", "--principal", "User:42", "--operation", "delete"},
		entry: mds.AccessControlEntry{PermissionType: mds.ACL_PERMISSION_TYPE_ALLOW,
			Principal: "User:42", Operation: mds.ACL_OPERATION_DELETE, Host: "*"},
	},
	{
		args: []string{"--deny", "--principal", "User:42", "--operation", "delete"},
		entry: mds.AccessControlEntry{PermissionType: mds.ACL_PERMISSION_TYPE_DENY,
			Principal: "User:42", Operation: mds.ACL_OPERATION_DELETE, Host: "*"},
	},
	{
		args: []string{"--allow", "--principal", "User:42", "--operation", "alter"},
		entry: mds.AccessControlEntry{PermissionType: mds.ACL_PERMISSION_TYPE_ALLOW,
			Principal: "User:42", Operation: mds.ACL_OPERATION_ALTER, Host: "*"},
	},
	{
		args: []string{"--deny", "--principal", "User:42", "--operation", "alter"},
		entry: mds.AccessControlEntry{PermissionType: mds.ACL_PERMISSION_TYPE_DENY,
			Principal: "User:42", Operation: mds.ACL_OPERATION_ALTER, Host: "*"},
	},
	{
		args: []string{"--allow", "--principal", "User:42", "--operation", "describe"},
		entry: mds.AccessControlEntry{PermissionType: mds.ACL_PERMISSION_TYPE_ALLOW,
			Principal: "User:42", Operation: mds.ACL_OPERATION_DESCRIBE, Host: "*"},
	},
	{
		args: []string{"--deny", "--principal", "User:42", "--operation", "describe"},
		entry: mds.AccessControlEntry{PermissionType: mds.ACL_PERMISSION_TYPE_DENY,
			Principal: "User:42", Operation: mds.ACL_OPERATION_DESCRIBE, Host: "*"},
	},
	{
		args: []string{"--allow", "--principal", "User:42", "--operation", "cluster-action"},
		entry: mds.AccessControlEntry{PermissionType: mds.ACL_PERMISSION_TYPE_ALLOW,
			Principal: "User:42", Operation: mds.ACL_OPERATION_CLUSTER_ACTION, Host: "*"},
	},
	{
		args: []string{"--deny", "--principal", "User:42", "--operation", "cluster-action"},
		entry: mds.AccessControlEntry{PermissionType: mds.ACL_PERMISSION_TYPE_DENY,
			Principal: "User:42", Operation: mds.ACL_OPERATION_CLUSTER_ACTION, Host: "*"},
	},
	{
		args: []string{"--allow", "--principal", "User:42", "--operation", "describe-configs"},
		entry: mds.AccessControlEntry{PermissionType: mds.ACL_PERMISSION_TYPE_ALLOW,
			Principal: "User:42", Operation: mds.ACL_OPERATION_DESCRIBE_CONFIGS, Host: "*"},
	},
	{
		args: []string{"--deny", "--principal", "User:42", "--operation", "describe-configs"},
		entry: mds.AccessControlEntry{PermissionType: mds.ACL_PERMISSION_TYPE_DENY,
			Principal: "User:42", Operation: mds.ACL_OPERATION_DESCRIBE_CONFIGS, Host: "*"},
	},
	{
		args: []string{"--allow", "--principal", "User:42", "--operation", "alter-configs"},
		entry: mds.AccessControlEntry{PermissionType: mds.ACL_PERMISSION_TYPE_ALLOW,
			Principal: "User:42", Operation: mds.ACL_OPERATION_ALTER_CONFIGS, Host: "*"},
	},
	{
		args: []string{"--deny", "--principal", "User:42", "--operation", "alter-configs"},
		entry: mds.AccessControlEntry{PermissionType: mds.ACL_PERMISSION_TYPE_DENY,
			Principal: "User:42", Operation: mds.ACL_OPERATION_ALTER_CONFIGS, Host: "*"},
	},
	{
		args: []string{"--allow", "--principal", "User:42", "--operation", "idempotent-write"},
		entry: mds.AccessControlEntry{PermissionType: mds.ACL_PERMISSION_TYPE_ALLOW,
			Principal: "User:42", Operation: mds.ACL_OPERATION_IDEMPOTENT_WRITE, Host: "*"},
	},
	{
		args: []string{"--deny", "--principal", "User:42", "--operation", "idempotent-write"},
		entry: mds.AccessControlEntry{PermissionType: mds.ACL_PERMISSION_TYPE_DENY,
			Principal: "User:42", Operation: mds.ACL_OPERATION_IDEMPOTENT_WRITE, Host: "*"},
	},
}

type AclTestSuite struct {
	suite.Suite
	conf      *config.Config
	kafkaApi  mds.KafkaACLManagementApi
	preRunner pcmd.PreRunner
}

func (suite *AclTestSuite) SetupSuite() {
	suite.conf = config.New()
	suite.conf.CLIName = "confluent"
	suite.conf.Logger = log.New()
	suite.conf.AuthURL = "http://test"
	suite.conf.AuthToken = "T0k3n"
}

func (suite *AclTestSuite) SetupTest() {
	suite.preRunner = &cliMock.Commander{}
}

func (suite *AclTestSuite) newMockIamCmd(expect chan interface{}, message string) *cobra.Command {
	suite.kafkaApi = &mock.KafkaACLManagementApi{
		AddAclBindingFunc:  func(ctx context.Context, createAclRequest mds.CreateAclRequest) (*net_http.Response, error) {
			assert.Equal(suite.T(), createAclRequest, <-expect, message)
			return nil, nil
		},
		RemoveAclBindingsFunc: func(ctx context.Context, aclFilterRequest mds.AclFilterRequest) ([]mds.AclBinding, *net_http.Response, error){
			assert.Equal(suite.T(), aclFilterRequest, <-expect, message)
			return nil, nil, nil
		},
		SearchAclBindingFunc: func(ctx context.Context, aclFilterRequest mds.AclFilterRequest) ([]mds.AclBinding, *net_http.Response, error){
			assert.Equal(suite.T(), aclFilterRequest, <-expect, message)
			return nil, nil, nil
		},
	}
	mdsClient := mds.NewAPIClient(mds.NewConfiguration())
	mdsClient.KafkaACLManagementApi = suite.kafkaApi
	return New(suite.preRunner, suite.conf, mdsClient)
}

func TestAclTestSuite(t *testing.T) {
	suite.Run(t, new(AclTestSuite))
}

func (suite *AclTestSuite) TestMdsCreateACL() {
	expect := make(chan interface{})
	for _, mdsResourcePattern := range mdsResourcePatterns {
		args := append([]string{"acl", "create", "--kafka-cluster-id", "testcluster"},
					mdsResourcePattern.args...)
		for _, mdsAclEntry := range mdsAclEntries {
			cmd := suite.newMockIamCmd(expect, "")
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
			assert.Nil(suite.T(), err)
		}
	}
}

func (suite *AclTestSuite) TestMdsDeleteACL() {
	expect := make(chan interface{})
	for _, mdsResourcePattern := range mdsResourcePatterns {
		args := append([]string{"acl", "delete", "--kafka-cluster-id", "testcluster", "--host", "*"},
					mdsResourcePattern.args...)
		for _, mdsAclEntry := range mdsAclEntries {
			cmd := suite.newMockIamCmd(expect, "")
			cmd.SetArgs(append(args, mdsAclEntry.args...))

			go func() {
				expect <- convertToAclFilterRequest(
					&mds.CreateAclRequest {
						Scope: mds.KafkaScope {
							Clusters: mds.KafkaScopeClusters{
								KafkaCluster: "testcluster",
							},
						},
						AclBinding: mds.AclBinding{
							Pattern: mdsResourcePattern.pattern,
							Entry: mdsAclEntry.entry,
						},
					},
				)
			}()

			err := cmd.Execute()
			assert.Nil(suite.T(), err)
		}
	}
}

func (suite *AclTestSuite) TestMdsListACL() {
	expect := make(chan interface{})
	for _, mdsResourcePattern := range mdsResourcePatterns {
		cmd := suite.newMockIamCmd(expect, "")
		cmd.SetArgs(append([]string{"acl", "list", "--kafka-cluster-id", "testcluster"}, mdsResourcePattern.args...))

		go func() {
			expect <- convertToAclFilterRequest(
				&mds.CreateAclRequest {
					Scope: mds.KafkaScope {
						Clusters: mds.KafkaScopeClusters{
							KafkaCluster: "testcluster",
						},
					},
					AclBinding: mds.AclBinding{
						Pattern: mdsResourcePattern.pattern,
						Entry: mds.AccessControlEntry{},
					},
				},
			)
		}()

		err := cmd.Execute()
		assert.Nil(suite.T(), err)
	}
}

func (suite *AclTestSuite) TestMdsListPrincipalACL() {
	expect := make(chan interface{})
	for _, mdsAclEntry := range mdsAclEntries {
		cmd := suite.newMockIamCmd(expect, "")
		cmd.SetArgs(append([]string{"acl", "list", "--kafka-cluster-id", "testcluster", "--principal"}, mdsAclEntry.entry.Principal))

		go func() {
			expect <- convertToAclFilterRequest(
				&mds.CreateAclRequest {
					Scope: mds.KafkaScope {
						Clusters: mds.KafkaScopeClusters{
							KafkaCluster: "testcluster",
						},
					},
					AclBinding: mds.AclBinding{
						Entry: mds.AccessControlEntry{
							Principal: mdsAclEntry.entry.Principal,
						},
					},
				},
				)
		}()

		err := cmd.Execute()
		assert.Nil(suite.T(), err)
	}
}

func (suite *AclTestSuite) TestMdsListPrincipalFilterACL() {
	expect := make(chan interface{})
	for _, mdsResourcePattern := range mdsResourcePatterns {
		args := append([]string{"acl", "list", "--kafka-cluster-id", "testcluster"}, mdsResourcePattern.args...)
		for _, mdsAclEntry := range mdsAclEntries {
			cmd := suite.newMockIamCmd(expect, "")
			cmd.SetArgs(append(args, "--principal", mdsAclEntry.entry.Principal))

			go func() {
				expect <- convertToAclFilterRequest(
					&mds.CreateAclRequest {
						Scope: mds.KafkaScope {
							Clusters: mds.KafkaScopeClusters{
								KafkaCluster: "testcluster",
							},
						},
						AclBinding: mds.AclBinding{
							Pattern: mdsResourcePattern.pattern,
							Entry: mds.AccessControlEntry{
								Principal: mdsAclEntry.entry.Principal,
							},
						},
					},
				)			}()

			err := cmd.Execute()
			assert.Nil(suite.T(), err)
		}
	}
}

func (suite *AclTestSuite) TestMdsMultipleResourceACL() {
	expect := "exactly one of cluster-scope, consumer-group, topic, transactional-id must be set"
	args := []string{"acl", "create", "--kafka-cluster-id", "testcluster",
		"--allow", "--operation", "read", "--principal", "User:42",
		"--topic", "resource1", "--consumer-group", "resource2"}

	cmd := suite.newMockIamCmd(nil, "")
	cmd.SetArgs(args)

	err := cmd.Execute()
	assert.NotNil(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), expect)
}

func (suite *AclTestSuite) TestMdsDefaults() {
	expect := make(chan interface{})
	cmd := suite.newMockIamCmd(expect,"Topic PatternType was not set to default value of PatternTypes_LITERAL")
	cmd.SetArgs([]string{"acl", "create", "--kafka-cluster-id", "testcluster",
		"--allow", "--principal", "User:42",
		"--operation", "read", "--topic", "dan"})
	go func() {
		expect <- mds.CreateAclRequest {
			Scope: mds.KafkaScope {
				Clusters: mds.KafkaScopeClusters{
					KafkaCluster: "testcluster",
				},
			},
			AclBinding: mds.AclBinding{
				Pattern: mds.KafkaResourcePattern {
					ResourceType: mds.ACL_RESOURCE_TYPE_TOPIC,
					Name: "dan",
					PatternType: mds.PATTERN_TYPE_LITERAL,
				},
				Entry: mds.AccessControlEntry{
					Principal: "User:42",
					PermissionType: mds.ACL_PERMISSION_TYPE_ALLOW,
					Operation: mds.ACL_OPERATION_READ,
					Host: "*",
				},
			},
		}
	}()

	err := cmd.Execute()
	assert.Nil(suite.T(), err)

	cmd = suite.newMockIamCmd(expect,"Cluster PatternType was not set to default value of PatternTypes_LITERAL")
	cmd.SetArgs([]string{"acl", "create", "--kafka-cluster-id", "testcluster",
		"--cluster-scope", "--allow", "--principal", "User:42",
		"--operation", "read"})

	go func() {
		expect <- mds.CreateAclRequest {
			Scope: mds.KafkaScope {
				Clusters: mds.KafkaScopeClusters{
					KafkaCluster: "testcluster",
				},
			},
			AclBinding: mds.AclBinding{
				Pattern: mds.KafkaResourcePattern {
					ResourceType: mds.ACL_RESOURCE_TYPE_CLUSTER,
					Name: "kafka-cluster",
					PatternType: mds.PATTERN_TYPE_LITERAL,
				},
				Entry: mds.AccessControlEntry{
					Principal: "User:42",
					PermissionType: mds.ACL_PERMISSION_TYPE_ALLOW,
					Operation: mds.ACL_OPERATION_READ,
					Host: "*",
				},
			},
		}
	}()

	err = cmd.Execute()
	assert.Nil(suite.T(), err)
}

func (suite *AclTestSuite) TestMdsHandleErrorNotLoggedIn() {
	// Use the real auth preRunner
	suite.preRunner = &pcmd.PreRun{
		Logger: log.New(),
		Config: suite.conf,
		UpdateClient: update.NewClient(&update.ClientParams{CheckInterval: 0}),
	}
	expect := make(chan interface{})

	cmd := suite.newMockIamCmd(expect, "")
	cmd.PersistentFlags().CountP("verbose", "v", "Increase output verbosity")

	for _, aclCmd := range []string { "list", "create", "delete" } {
		cmd.SetArgs([]string{"acl", aclCmd, "--kafka-cluster-id", "testcluster"})
		go func() {
			expect <- nil
		}()
		err := cmd.Execute()
		assert.NotNil(suite.T(), err)
		assert.Contains(suite.T(), err.Error(), errors.HandleCommon(errors.ErrNotLoggedIn, cmd).Error())
	}
}
