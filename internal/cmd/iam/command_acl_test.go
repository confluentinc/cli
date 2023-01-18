package iam

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	mds "github.com/confluentinc/mds-sdk-go-public/mdsv1"
	"github.com/confluentinc/mds-sdk-go-public/mdsv1/mock"

	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	climock "github.com/confluentinc/cli/mock"
)

/*************** TEST command_acl ***************/
var mdsResourcePatterns = []struct {
	args    []string
	pattern mds.KafkaResourcePattern
}{
	{
		args: []string{"--cluster-scope"},
		pattern: mds.KafkaResourcePattern{ResourceType: mds.ACLRESOURCETYPE_CLUSTER, Name: "kafka-cluster",
			PatternType: mds.PATTERNTYPE_LITERAL},
	},
	{
		args: []string{"--topic", "test-topic"},
		pattern: mds.KafkaResourcePattern{ResourceType: mds.ACLRESOURCETYPE_TOPIC, Name: "test-topic",
			PatternType: mds.PATTERNTYPE_LITERAL},
	},
	{
		args: []string{"--topic", "test-topic", "--prefix"},
		pattern: mds.KafkaResourcePattern{ResourceType: mds.ACLRESOURCETYPE_TOPIC, Name: "test-topic",
			PatternType: mds.PATTERNTYPE_PREFIXED},
	},
	{
		args: []string{"--consumer-group", "test-group"},
		pattern: mds.KafkaResourcePattern{ResourceType: mds.ACLRESOURCETYPE_GROUP, Name: "test-group",
			PatternType: mds.PATTERNTYPE_LITERAL},
	},
	{
		args: []string{"--consumer-group", "test-group", "--prefix"},
		pattern: mds.KafkaResourcePattern{ResourceType: mds.ACLRESOURCETYPE_GROUP, Name: "test-group",
			PatternType: mds.PATTERNTYPE_PREFIXED},
	},
	{
		args: []string{"--transactional-id", "test-transactional-id"},
		pattern: mds.KafkaResourcePattern{ResourceType: mds.ACLRESOURCETYPE_TRANSACTIONAL_ID, Name: "test-transactional-id",
			PatternType: mds.PATTERNTYPE_LITERAL},
	},
	{
		args: []string{"--transactional-id", "test-transactional-id", "--prefix"},
		pattern: mds.KafkaResourcePattern{ResourceType: mds.ACLRESOURCETYPE_TRANSACTIONAL_ID, Name: "test-transactional-id",
			PatternType: mds.PATTERNTYPE_PREFIXED},
	},
}

var mdsACLEntries = []struct {
	args  []string
	entry mds.AccessControlEntry
}{
	{
		args: []string{"--allow", "--principal", "User:42", "--operation", "read"},
		entry: mds.AccessControlEntry{PermissionType: mds.ACLPERMISSIONTYPE_ALLOW,
			Principal: "User:42", Operation: mds.ACLOPERATION_READ, Host: "*"},
	},
	{
		args: []string{"--deny", "--principal", "User:42", "--host", "testhost", "--operation", "read"},
		entry: mds.AccessControlEntry{PermissionType: mds.ACLPERMISSIONTYPE_DENY,
			Principal: "User:42", Operation: mds.ACLOPERATION_READ, Host: "testhost"},
	},
	{
		args: []string{"--allow", "--principal", "User:42", "--host", "*", "--operation", "write"},
		entry: mds.AccessControlEntry{PermissionType: mds.ACLPERMISSIONTYPE_ALLOW,
			Principal: "User:42", Operation: mds.ACLOPERATION_WRITE, Host: "*"},
	},
	{
		args: []string{"--deny", "--principal", "User:42", "--operation", "write"},
		entry: mds.AccessControlEntry{PermissionType: mds.ACLPERMISSIONTYPE_DENY,
			Principal: "User:42", Operation: mds.ACLOPERATION_WRITE, Host: "*"},
	},
	{
		args: []string{"--allow", "--principal", "User:42", "--operation", "create"},
		entry: mds.AccessControlEntry{PermissionType: mds.ACLPERMISSIONTYPE_ALLOW,
			Principal: "User:42", Operation: mds.ACLOPERATION_CREATE, Host: "*"},
	},
	{
		args: []string{"--deny", "--principal", "User:42", "--operation", "create"},
		entry: mds.AccessControlEntry{PermissionType: mds.ACLPERMISSIONTYPE_DENY,
			Principal: "User:42", Operation: mds.ACLOPERATION_CREATE, Host: "*"},
	},
	{
		args: []string{"--allow", "--principal", "User:42", "--operation", "delete"},
		entry: mds.AccessControlEntry{PermissionType: mds.ACLPERMISSIONTYPE_ALLOW,
			Principal: "User:42", Operation: mds.ACLOPERATION_DELETE, Host: "*"},
	},
	{
		args: []string{"--deny", "--principal", "User:42", "--operation", "delete"},
		entry: mds.AccessControlEntry{PermissionType: mds.ACLPERMISSIONTYPE_DENY,
			Principal: "User:42", Operation: mds.ACLOPERATION_DELETE, Host: "*"},
	},
	{
		args: []string{"--allow", "--principal", "User:42", "--operation", "alter"},
		entry: mds.AccessControlEntry{PermissionType: mds.ACLPERMISSIONTYPE_ALLOW,
			Principal: "User:42", Operation: mds.ACLOPERATION_ALTER, Host: "*"},
	},
	{
		args: []string{"--deny", "--principal", "User:42", "--operation", "alter"},
		entry: mds.AccessControlEntry{PermissionType: mds.ACLPERMISSIONTYPE_DENY,
			Principal: "User:42", Operation: mds.ACLOPERATION_ALTER, Host: "*"},
	},
	{
		args: []string{"--allow", "--principal", "User:42", "--operation", "describe"},
		entry: mds.AccessControlEntry{PermissionType: mds.ACLPERMISSIONTYPE_ALLOW,
			Principal: "User:42", Operation: mds.ACLOPERATION_DESCRIBE, Host: "*"},
	},
	{
		args: []string{"--deny", "--principal", "User:42", "--operation", "describe"},
		entry: mds.AccessControlEntry{PermissionType: mds.ACLPERMISSIONTYPE_DENY,
			Principal: "User:42", Operation: mds.ACLOPERATION_DESCRIBE, Host: "*"},
	},
	{
		args: []string{"--allow", "--principal", "User:42", "--operation", "cluster-action"},
		entry: mds.AccessControlEntry{PermissionType: mds.ACLPERMISSIONTYPE_ALLOW,
			Principal: "User:42", Operation: mds.ACLOPERATION_CLUSTER_ACTION, Host: "*"},
	},
	{
		args: []string{"--deny", "--principal", "User:42", "--operation", "cluster-action"},
		entry: mds.AccessControlEntry{PermissionType: mds.ACLPERMISSIONTYPE_DENY,
			Principal: "User:42", Operation: mds.ACLOPERATION_CLUSTER_ACTION, Host: "*"},
	},
	{
		args: []string{"--allow", "--principal", "User:42", "--operation", "describe-configs"},
		entry: mds.AccessControlEntry{PermissionType: mds.ACLPERMISSIONTYPE_ALLOW,
			Principal: "User:42", Operation: mds.ACLOPERATION_DESCRIBE_CONFIGS, Host: "*"},
	},
	{
		args: []string{"--deny", "--principal", "User:42", "--operation", "describe-configs"},
		entry: mds.AccessControlEntry{PermissionType: mds.ACLPERMISSIONTYPE_DENY,
			Principal: "User:42", Operation: mds.ACLOPERATION_DESCRIBE_CONFIGS, Host: "*"},
	},
	{
		args: []string{"--allow", "--principal", "User:42", "--operation", "alter-configs"},
		entry: mds.AccessControlEntry{PermissionType: mds.ACLPERMISSIONTYPE_ALLOW,
			Principal: "User:42", Operation: mds.ACLOPERATION_ALTER_CONFIGS, Host: "*"},
	},
	{
		args: []string{"--deny", "--principal", "User:42", "--operation", "alter-configs"},
		entry: mds.AccessControlEntry{PermissionType: mds.ACLPERMISSIONTYPE_DENY,
			Principal: "User:42", Operation: mds.ACLOPERATION_ALTER_CONFIGS, Host: "*"},
	},
	{
		args: []string{"--allow", "--principal", "User:42", "--operation", "idempotent-write"},
		entry: mds.AccessControlEntry{PermissionType: mds.ACLPERMISSIONTYPE_ALLOW,
			Principal: "User:42", Operation: mds.ACLOPERATION_IDEMPOTENT_WRITE, Host: "*"},
	},
	{
		args: []string{"--deny", "--principal", "User:42", "--operation", "idempotent-write"},
		entry: mds.AccessControlEntry{PermissionType: mds.ACLPERMISSIONTYPE_DENY,
			Principal: "User:42", Operation: mds.ACLOPERATION_IDEMPOTENT_WRITE, Host: "*"},
	},
}

type ACLTestSuite struct {
	suite.Suite
	conf     *v1.Config
	kafkaApi mds.KafkaACLManagementApi
}

func (suite *ACLTestSuite) SetupSuite() {
	suite.conf = v1.AuthenticatedOnPremConfigMock()
}

func (suite *ACLTestSuite) newMockIamCmd(expect chan interface{}, message string) *cobra.Command {
	suite.kafkaApi = &mock.KafkaACLManagementApi{
		AddAclBindingFunc: func(ctx context.Context, createAclRequest mds.CreateAclRequest) (*http.Response, error) {
			assert.Equal(suite.T(), createAclRequest, <-expect, message)
			return nil, nil
		},
		RemoveAclBindingsFunc: func(ctx context.Context, aclFilterRequest mds.AclFilterRequest) ([]mds.AclBinding, *http.Response, error) {
			assert.Equal(suite.T(), aclFilterRequest, <-expect, message)
			return nil, nil, nil
		},
		SearchAclBindingFunc: func(ctx context.Context, aclFilterRequest mds.AclFilterRequest) ([]mds.AclBinding, *http.Response, error) {
			assert.Equal(suite.T(), aclFilterRequest, <-expect, message)
			return nil, nil, nil
		},
	}
	mdsClient := mds.NewAPIClient(mds.NewConfiguration())
	mdsClient.KafkaACLManagementApi = suite.kafkaApi
	return New(suite.conf, climock.NewPreRunnerMock(nil, nil, mdsClient, nil, suite.conf))
}

func TestAclTestSuite(t *testing.T) {
	suite.Run(t, new(ACLTestSuite))
}

func (suite *ACLTestSuite) TestMdsCreateACL() {
	expect := make(chan interface{})
	for _, mdsResourcePattern := range mdsResourcePatterns {
		args := append([]string{"acl", "create", "--kafka-cluster", "testcluster"},
			mdsResourcePattern.args...)
		for _, mdsAclEntry := range mdsACLEntries {
			cmd := suite.newMockIamCmd(expect, "")
			cmd.SetArgs(append(args, mdsAclEntry.args...))

			go func() {
				expect <- mds.CreateAclRequest{
					Scope: mds.KafkaScope{
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

func (suite *ACLTestSuite) TestMdsDeleteACL() {
	expect := make(chan interface{})
	for _, mdsResourcePattern := range mdsResourcePatterns {
		args := append([]string{"acl", "delete", "--kafka-cluster", "testcluster", "--host", "*", "--force"},
			mdsResourcePattern.args...)
		for _, mdsAclEntry := range mdsACLEntries {
			cmd := suite.newMockIamCmd(expect, "")
			cmd.SetArgs(append(args, mdsAclEntry.args...))

			go func() {
				expect <- convertToACLFilterRequest(
					&mds.CreateAclRequest{
						Scope: mds.KafkaScope{
							Clusters: mds.KafkaScopeClusters{
								KafkaCluster: "testcluster",
							},
						},
						AclBinding: mds.AclBinding{
							Pattern: mdsResourcePattern.pattern,
							Entry:   mdsAclEntry.entry,
						},
					},
				)
				expect <- convertToACLFilterRequest(
					&mds.CreateAclRequest{
						Scope: mds.KafkaScope{
							Clusters: mds.KafkaScopeClusters{
								KafkaCluster: "testcluster",
							},
						},
						AclBinding: mds.AclBinding{
							Pattern: mdsResourcePattern.pattern,
							Entry:   mdsAclEntry.entry,
						},
					},
				)
			}()

			err := cmd.Execute()
			assert.Nil(suite.T(), err)
		}
	}
}

func (suite *ACLTestSuite) TestMdsListACL() {
	expect := make(chan interface{})
	for _, mdsResourcePattern := range mdsResourcePatterns {
		cmd := suite.newMockIamCmd(expect, "")
		cmd.SetArgs(append([]string{"acl", "list", "--kafka-cluster", "testcluster"}, mdsResourcePattern.args...))

		go func() {
			expect <- convertToACLFilterRequest(
				&mds.CreateAclRequest{
					Scope: mds.KafkaScope{
						Clusters: mds.KafkaScopeClusters{
							KafkaCluster: "testcluster",
						},
					},
					AclBinding: mds.AclBinding{
						Pattern: mdsResourcePattern.pattern,
						Entry:   mds.AccessControlEntry{},
					},
				},
			)
		}()

		err := cmd.Execute()
		assert.Nil(suite.T(), err)
	}
}

func (suite *ACLTestSuite) TestMdsListPrincipalACL() {
	expect := make(chan interface{})
	for _, mdsAclEntry := range mdsACLEntries {
		cmd := suite.newMockIamCmd(expect, "")
		cmd.SetArgs(append([]string{"acl", "list", "--kafka-cluster", "testcluster", "--principal"}, mdsAclEntry.entry.Principal))

		go func() {
			expect <- convertToACLFilterRequest(
				&mds.CreateAclRequest{
					Scope: mds.KafkaScope{
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

func (suite *ACLTestSuite) TestMdsListPrincipalFilterACL() {
	expect := make(chan interface{})
	for _, mdsResourcePattern := range mdsResourcePatterns {
		args := append([]string{"acl", "list", "--kafka-cluster", "testcluster"}, mdsResourcePattern.args...)
		for _, mdsAclEntry := range mdsACLEntries {
			cmd := suite.newMockIamCmd(expect, "")
			cmd.SetArgs(append(args, "--principal", mdsAclEntry.entry.Principal))

			go func() {
				expect <- convertToACLFilterRequest(
					&mds.CreateAclRequest{
						Scope: mds.KafkaScope{
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
				)
			}()
			err := cmd.Execute()
			assert.Nil(suite.T(), err)
		}
	}
}

func (suite *ACLTestSuite) TestMdsMultipleResourceACL() {
	args := []string{"acl", "create", "--kafka-cluster", "testcluster",
		"--allow", "--operation", "read", "--principal", "User:42",
		"--topic", "resource1", "--consumer-group", "resource2"}

	cmd := suite.newMockIamCmd(nil, "")
	cmd.SetArgs(args)

	err := cmd.Execute()
	assert.NotNil(suite.T(), err)
	expect := fmt.Sprintf(errors.ExactlyOneSetErrorMsg, "cluster-scope, consumer-group, topic, transactional-id")
	assert.Contains(suite.T(), err.Error(), expect)
}

func (suite *ACLTestSuite) TestMdsDefaults() {
	expect := make(chan interface{})
	cmd := suite.newMockIamCmd(expect, "Topic PatternType was not set to default value of PatternTypes_LITERAL")
	cmd.SetArgs([]string{"acl", "create", "--kafka-cluster", "testcluster",
		"--allow", "--principal", "User:42",
		"--operation", "read", "--topic", "dan"})
	go func() {
		expect <- mds.CreateAclRequest{
			Scope: mds.KafkaScope{
				Clusters: mds.KafkaScopeClusters{
					KafkaCluster: "testcluster",
				},
			},
			AclBinding: mds.AclBinding{
				Pattern: mds.KafkaResourcePattern{
					ResourceType: mds.ACLRESOURCETYPE_TOPIC,
					Name:         "dan",
					PatternType:  mds.PATTERNTYPE_LITERAL,
				},
				Entry: mds.AccessControlEntry{
					Principal:      "User:42",
					PermissionType: mds.ACLPERMISSIONTYPE_ALLOW,
					Operation:      mds.ACLOPERATION_READ,
					Host:           "*",
				},
			},
		}
	}()

	err := cmd.Execute()
	assert.Nil(suite.T(), err)
	cmd = suite.newMockIamCmd(expect, "Cluster PatternType was not set to default value of PatternTypes_LITERAL")

	cmd.SetArgs([]string{"acl", "create", "--kafka-cluster", "testcluster",
		"--cluster-scope", "--allow", "--principal", "User:42",
		"--operation", "read"})

	go func() {
		expect <- mds.CreateAclRequest{
			Scope: mds.KafkaScope{
				Clusters: mds.KafkaScopeClusters{
					KafkaCluster: "testcluster",
				},
			},
			AclBinding: mds.AclBinding{
				Pattern: mds.KafkaResourcePattern{
					ResourceType: mds.ACLRESOURCETYPE_CLUSTER,
					Name:         "kafka-cluster",
					PatternType:  mds.PATTERNTYPE_LITERAL,
				},
				Entry: mds.AccessControlEntry{
					Principal:      "User:42",
					PermissionType: mds.ACLPERMISSIONTYPE_ALLOW,
					Operation:      mds.ACLOPERATION_READ,
					Host:           "*",
				},
			},
		}
	}()

	err = cmd.Execute()
	assert.Nil(suite.T(), err)
}
