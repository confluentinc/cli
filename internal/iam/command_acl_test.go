package iam

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/confluentinc/mds-sdk-go-public/mdsv1"
	"github.com/confluentinc/mds-sdk-go-public/mdsv1/mock"

	climock "github.com/confluentinc/cli/v3/mock"
	"github.com/confluentinc/cli/v3/pkg/config"
	"github.com/confluentinc/cli/v3/pkg/errors"
)

/*************** TEST command_acl ***************/
var mdsResourcePatterns = []struct {
	args    []string
	pattern mdsv1.KafkaResourcePattern
}{
	{
		args: []string{"--cluster-scope"},
		pattern: mdsv1.KafkaResourcePattern{
			ResourceType: mdsv1.ACLRESOURCETYPE_CLUSTER, Name: "kafka-cluster",
			PatternType: mdsv1.PATTERNTYPE_LITERAL,
		},
	},
	{
		args: []string{"--topic", "test-topic"},
		pattern: mdsv1.KafkaResourcePattern{
			ResourceType: mdsv1.ACLRESOURCETYPE_TOPIC, Name: "test-topic",
			PatternType: mdsv1.PATTERNTYPE_LITERAL,
		},
	},
	{
		args: []string{"--topic", "test-topic", "--prefix"},
		pattern: mdsv1.KafkaResourcePattern{
			ResourceType: mdsv1.ACLRESOURCETYPE_TOPIC, Name: "test-topic",
			PatternType: mdsv1.PATTERNTYPE_PREFIXED,
		},
	},
	{
		args: []string{"--consumer-group", "test-group"},
		pattern: mdsv1.KafkaResourcePattern{
			ResourceType: mdsv1.ACLRESOURCETYPE_GROUP, Name: "test-group",
			PatternType: mdsv1.PATTERNTYPE_LITERAL,
		},
	},
	{
		args: []string{"--consumer-group", "test-group", "--prefix"},
		pattern: mdsv1.KafkaResourcePattern{
			ResourceType: mdsv1.ACLRESOURCETYPE_GROUP, Name: "test-group",
			PatternType: mdsv1.PATTERNTYPE_PREFIXED,
		},
	},
	{
		args: []string{"--transactional-id", "test-transactional-id"},
		pattern: mdsv1.KafkaResourcePattern{
			ResourceType: mdsv1.ACLRESOURCETYPE_TRANSACTIONAL_ID, Name: "test-transactional-id",
			PatternType: mdsv1.PATTERNTYPE_LITERAL,
		},
	},
	{
		args: []string{"--transactional-id", "test-transactional-id", "--prefix"},
		pattern: mdsv1.KafkaResourcePattern{
			ResourceType: mdsv1.ACLRESOURCETYPE_TRANSACTIONAL_ID, Name: "test-transactional-id",
			PatternType: mdsv1.PATTERNTYPE_PREFIXED,
		},
	},
}

var mdsACLEntries = []struct {
	args  []string
	entry mdsv1.AccessControlEntry
}{
	{
		args: []string{"--allow", "--principal", "User:42", "--operation", "read"},
		entry: mdsv1.AccessControlEntry{
			PermissionType: mdsv1.ACLPERMISSIONTYPE_ALLOW,
			Principal:      "User:42", Operation: mdsv1.ACLOPERATION_READ, Host: "*",
		},
	},
	{
		args: []string{"--deny", "--principal", "User:42", "--host", "testhost", "--operation", "read"},
		entry: mdsv1.AccessControlEntry{
			PermissionType: mdsv1.ACLPERMISSIONTYPE_DENY,
			Principal:      "User:42", Operation: mdsv1.ACLOPERATION_READ, Host: "testhost",
		},
	},
	{
		args: []string{"--allow", "--principal", "User:42", "--host", "*", "--operation", "write"},
		entry: mdsv1.AccessControlEntry{
			PermissionType: mdsv1.ACLPERMISSIONTYPE_ALLOW,
			Principal:      "User:42", Operation: mdsv1.ACLOPERATION_WRITE, Host: "*",
		},
	},
	{
		args: []string{"--deny", "--principal", "User:42", "--operation", "write"},
		entry: mdsv1.AccessControlEntry{
			PermissionType: mdsv1.ACLPERMISSIONTYPE_DENY,
			Principal:      "User:42", Operation: mdsv1.ACLOPERATION_WRITE, Host: "*",
		},
	},
	{
		args: []string{"--allow", "--principal", "User:42", "--operation", "create"},
		entry: mdsv1.AccessControlEntry{
			PermissionType: mdsv1.ACLPERMISSIONTYPE_ALLOW,
			Principal:      "User:42", Operation: mdsv1.ACLOPERATION_CREATE, Host: "*",
		},
	},
	{
		args: []string{"--deny", "--principal", "User:42", "--operation", "create"},
		entry: mdsv1.AccessControlEntry{
			PermissionType: mdsv1.ACLPERMISSIONTYPE_DENY,
			Principal:      "User:42", Operation: mdsv1.ACLOPERATION_CREATE, Host: "*",
		},
	},
	{
		args: []string{"--allow", "--principal", "User:42", "--operation", "delete"},
		entry: mdsv1.AccessControlEntry{
			PermissionType: mdsv1.ACLPERMISSIONTYPE_ALLOW,
			Principal:      "User:42", Operation: mdsv1.ACLOPERATION_DELETE, Host: "*",
		},
	},
	{
		args: []string{"--deny", "--principal", "User:42", "--operation", "delete"},
		entry: mdsv1.AccessControlEntry{
			PermissionType: mdsv1.ACLPERMISSIONTYPE_DENY,
			Principal:      "User:42", Operation: mdsv1.ACLOPERATION_DELETE, Host: "*",
		},
	},
	{
		args: []string{"--allow", "--principal", "User:42", "--operation", "alter"},
		entry: mdsv1.AccessControlEntry{
			PermissionType: mdsv1.ACLPERMISSIONTYPE_ALLOW,
			Principal:      "User:42", Operation: mdsv1.ACLOPERATION_ALTER, Host: "*",
		},
	},
	{
		args: []string{"--deny", "--principal", "User:42", "--operation", "alter"},
		entry: mdsv1.AccessControlEntry{
			PermissionType: mdsv1.ACLPERMISSIONTYPE_DENY,
			Principal:      "User:42", Operation: mdsv1.ACLOPERATION_ALTER, Host: "*",
		},
	},
	{
		args: []string{"--allow", "--principal", "User:42", "--operation", "describe"},
		entry: mdsv1.AccessControlEntry{
			PermissionType: mdsv1.ACLPERMISSIONTYPE_ALLOW,
			Principal:      "User:42", Operation: mdsv1.ACLOPERATION_DESCRIBE, Host: "*",
		},
	},
	{
		args: []string{"--deny", "--principal", "User:42", "--operation", "describe"},
		entry: mdsv1.AccessControlEntry{
			PermissionType: mdsv1.ACLPERMISSIONTYPE_DENY,
			Principal:      "User:42", Operation: mdsv1.ACLOPERATION_DESCRIBE, Host: "*",
		},
	},
	{
		args: []string{"--allow", "--principal", "User:42", "--operation", "cluster-action"},
		entry: mdsv1.AccessControlEntry{
			PermissionType: mdsv1.ACLPERMISSIONTYPE_ALLOW,
			Principal:      "User:42", Operation: mdsv1.ACLOPERATION_CLUSTER_ACTION, Host: "*",
		},
	},
	{
		args: []string{"--deny", "--principal", "User:42", "--operation", "cluster-action"},
		entry: mdsv1.AccessControlEntry{
			PermissionType: mdsv1.ACLPERMISSIONTYPE_DENY,
			Principal:      "User:42", Operation: mdsv1.ACLOPERATION_CLUSTER_ACTION, Host: "*",
		},
	},
	{
		args: []string{"--allow", "--principal", "User:42", "--operation", "describe-configs"},
		entry: mdsv1.AccessControlEntry{
			PermissionType: mdsv1.ACLPERMISSIONTYPE_ALLOW,
			Principal:      "User:42", Operation: mdsv1.ACLOPERATION_DESCRIBE_CONFIGS, Host: "*",
		},
	},
	{
		args: []string{"--deny", "--principal", "User:42", "--operation", "describe-configs"},
		entry: mdsv1.AccessControlEntry{
			PermissionType: mdsv1.ACLPERMISSIONTYPE_DENY,
			Principal:      "User:42", Operation: mdsv1.ACLOPERATION_DESCRIBE_CONFIGS, Host: "*",
		},
	},
	{
		args: []string{"--allow", "--principal", "User:42", "--operation", "alter-configs"},
		entry: mdsv1.AccessControlEntry{
			PermissionType: mdsv1.ACLPERMISSIONTYPE_ALLOW,
			Principal:      "User:42", Operation: mdsv1.ACLOPERATION_ALTER_CONFIGS, Host: "*",
		},
	},
	{
		args: []string{"--deny", "--principal", "User:42", "--operation", "alter-configs"},
		entry: mdsv1.AccessControlEntry{
			PermissionType: mdsv1.ACLPERMISSIONTYPE_DENY,
			Principal:      "User:42", Operation: mdsv1.ACLOPERATION_ALTER_CONFIGS, Host: "*",
		},
	},
	{
		args: []string{"--allow", "--principal", "User:42", "--operation", "idempotent-write"},
		entry: mdsv1.AccessControlEntry{
			PermissionType: mdsv1.ACLPERMISSIONTYPE_ALLOW,
			Principal:      "User:42", Operation: mdsv1.ACLOPERATION_IDEMPOTENT_WRITE, Host: "*",
		},
	},
	{
		args: []string{"--deny", "--principal", "User:42", "--operation", "idempotent-write"},
		entry: mdsv1.AccessControlEntry{
			PermissionType: mdsv1.ACLPERMISSIONTYPE_DENY,
			Principal:      "User:42", Operation: mdsv1.ACLOPERATION_IDEMPOTENT_WRITE, Host: "*",
		},
	},
}

type ACLTestSuite struct {
	suite.Suite
	conf     *config.Config
	kafkaApi mdsv1.KafkaACLManagementApi
}

func (suite *ACLTestSuite) SetupSuite() {
	suite.conf = config.AuthenticatedOnPremConfigMock()
}

func (suite *ACLTestSuite) newMockIamCmd(expect chan any, message string) *cobra.Command {
	suite.kafkaApi = &mock.KafkaACLManagementApi{
		AddAclBindingFunc: func(ctx context.Context, createAclRequest mdsv1.CreateAclRequest) (*http.Response, error) {
			assert.Equal(suite.T(), createAclRequest, <-expect, message)
			return nil, nil
		},
		RemoveAclBindingsFunc: func(ctx context.Context, aclFilterRequest mdsv1.AclFilterRequest) ([]mdsv1.AclBinding, *http.Response, error) {
			assert.Equal(suite.T(), aclFilterRequest, <-expect, message)
			return nil, nil, nil
		},
		SearchAclBindingFunc: func(ctx context.Context, aclFilterRequest mdsv1.AclFilterRequest) ([]mdsv1.AclBinding, *http.Response, error) {
			assert.Equal(suite.T(), aclFilterRequest, <-expect, message)
			return nil, nil, nil
		},
	}
	mdsClient := mdsv1.NewAPIClient(mdsv1.NewConfiguration())
	mdsClient.KafkaACLManagementApi = suite.kafkaApi
	return New(suite.conf, climock.NewPreRunnerMock(nil, nil, mdsClient, nil, suite.conf))
}

func TestAclTestSuite(t *testing.T) {
	suite.Run(t, new(ACLTestSuite))
}

func (suite *ACLTestSuite) TestMdsCreateACL() {
	expect := make(chan any)
	for _, mdsResourcePattern := range mdsResourcePatterns {
		args := append([]string{"acl", "create", "--kafka-cluster", "testcluster"},
			mdsResourcePattern.args...)
		for _, mdsAclEntry := range mdsACLEntries {
			cmd := suite.newMockIamCmd(expect, "")
			cmd.SetArgs(append(args, mdsAclEntry.args...))

			go func() {
				expect <- mdsv1.CreateAclRequest{
					Scope: mdsv1.KafkaScope{
						Clusters: mdsv1.KafkaScopeClusters{
							KafkaCluster: "testcluster",
						},
					},
					AclBinding: mdsv1.AclBinding{Pattern: mdsResourcePattern.pattern, Entry: mdsAclEntry.entry},
				}
			}()

			err := cmd.Execute()
			assert.Nil(suite.T(), err)
		}
	}
}

func (suite *ACLTestSuite) TestMdsDeleteACL() {
	expect := make(chan any)
	for _, mdsResourcePattern := range mdsResourcePatterns {
		args := append([]string{"acl", "delete", "--kafka-cluster", "testcluster", "--host", "*", "--force"},
			mdsResourcePattern.args...)
		for _, mdsAclEntry := range mdsACLEntries {
			cmd := suite.newMockIamCmd(expect, "")
			cmd.SetArgs(append(args, mdsAclEntry.args...))

			go func() {
				expect <- convertToAclFilterRequest(
					&mdsv1.CreateAclRequest{
						Scope: mdsv1.KafkaScope{
							Clusters: mdsv1.KafkaScopeClusters{
								KafkaCluster: "testcluster",
							},
						},
						AclBinding: mdsv1.AclBinding{
							Pattern: mdsResourcePattern.pattern,
							Entry:   mdsAclEntry.entry,
						},
					},
				)
				expect <- convertToAclFilterRequest(
					&mdsv1.CreateAclRequest{
						Scope: mdsv1.KafkaScope{
							Clusters: mdsv1.KafkaScopeClusters{
								KafkaCluster: "testcluster",
							},
						},
						AclBinding: mdsv1.AclBinding{
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
	expect := make(chan any)
	for _, mdsResourcePattern := range mdsResourcePatterns {
		cmd := suite.newMockIamCmd(expect, "")
		cmd.SetArgs(append([]string{"acl", "list", "--kafka-cluster", "testcluster"}, mdsResourcePattern.args...))

		go func() {
			expect <- convertToAclFilterRequest(
				&mdsv1.CreateAclRequest{
					Scope: mdsv1.KafkaScope{
						Clusters: mdsv1.KafkaScopeClusters{
							KafkaCluster: "testcluster",
						},
					},
					AclBinding: mdsv1.AclBinding{
						Pattern: mdsResourcePattern.pattern,
						Entry:   mdsv1.AccessControlEntry{},
					},
				},
			)
		}()

		err := cmd.Execute()
		assert.Nil(suite.T(), err)
	}
}

func (suite *ACLTestSuite) TestMdsListPrincipalACL() {
	expect := make(chan any)
	for _, mdsAclEntry := range mdsACLEntries {
		cmd := suite.newMockIamCmd(expect, "")
		cmd.SetArgs(append([]string{"acl", "list", "--kafka-cluster", "testcluster", "--principal"}, mdsAclEntry.entry.Principal))

		go func() {
			expect <- convertToAclFilterRequest(
				&mdsv1.CreateAclRequest{
					Scope: mdsv1.KafkaScope{
						Clusters: mdsv1.KafkaScopeClusters{
							KafkaCluster: "testcluster",
						},
					},
					AclBinding: mdsv1.AclBinding{
						Entry: mdsv1.AccessControlEntry{
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
	expect := make(chan any)
	for _, mdsResourcePattern := range mdsResourcePatterns {
		args := append([]string{"acl", "list", "--kafka-cluster", "testcluster"}, mdsResourcePattern.args...)
		for _, mdsAclEntry := range mdsACLEntries {
			cmd := suite.newMockIamCmd(expect, "")
			cmd.SetArgs(append(args, "--principal", mdsAclEntry.entry.Principal))

			go func() {
				expect <- convertToAclFilterRequest(
					&mdsv1.CreateAclRequest{
						Scope: mdsv1.KafkaScope{
							Clusters: mdsv1.KafkaScopeClusters{
								KafkaCluster: "testcluster",
							},
						},
						AclBinding: mdsv1.AclBinding{
							Pattern: mdsResourcePattern.pattern,
							Entry: mdsv1.AccessControlEntry{
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
	cmd := suite.newMockIamCmd(nil, "")
	cmd.SetArgs([]string{"acl", "create", "--kafka-cluster", "testcluster", "--allow", "--operation", "read", "--principal", "User:42", "--topic", "resource1", "--consumer-group", "resource2"})

	err := cmd.Execute()
	assert.NotNil(suite.T(), err)
	expect := fmt.Sprintf(errors.ExactlyOneSetErrorMsg, "`--cluster-scope`, `--consumer-group`, `--topic`, `--transactional-id`")
	assert.Contains(suite.T(), err.Error(), expect)
}

func (suite *ACLTestSuite) TestMdsDefaults() {
	expect := make(chan any)
	cmd := suite.newMockIamCmd(expect, "Topic PatternType was not set to default value of PatternTypes_LITERAL")
	cmd.SetArgs([]string{"acl", "create", "--kafka-cluster", "testcluster", "--allow", "--principal", "User:42", "--operation", "read", "--topic", "dan"})
	go func() {
		expect <- mdsv1.CreateAclRequest{
			Scope: mdsv1.KafkaScope{
				Clusters: mdsv1.KafkaScopeClusters{
					KafkaCluster: "testcluster",
				},
			},
			AclBinding: mdsv1.AclBinding{
				Pattern: mdsv1.KafkaResourcePattern{
					ResourceType: mdsv1.ACLRESOURCETYPE_TOPIC,
					Name:         "dan",
					PatternType:  mdsv1.PATTERNTYPE_LITERAL,
				},
				Entry: mdsv1.AccessControlEntry{
					Principal:      "User:42",
					PermissionType: mdsv1.ACLPERMISSIONTYPE_ALLOW,
					Operation:      mdsv1.ACLOPERATION_READ,
					Host:           "*",
				},
			},
		}
	}()

	err := cmd.Execute()
	assert.Nil(suite.T(), err)
	cmd = suite.newMockIamCmd(expect, "Cluster PatternType was not set to default value of PatternTypes_LITERAL")

	cmd.SetArgs([]string{"acl", "create", "--kafka-cluster", "testcluster", "--cluster-scope", "--allow", "--principal", "User:42", "--operation", "read"})

	go func() {
		expect <- mdsv1.CreateAclRequest{
			Scope: mdsv1.KafkaScope{
				Clusters: mdsv1.KafkaScopeClusters{
					KafkaCluster: "testcluster",
				},
			},
			AclBinding: mdsv1.AclBinding{
				Pattern: mdsv1.KafkaResourcePattern{
					ResourceType: mdsv1.ACLRESOURCETYPE_CLUSTER,
					Name:         "kafka-cluster",
					PatternType:  mdsv1.PATTERNTYPE_LITERAL,
				},
				Entry: mdsv1.AccessControlEntry{
					Principal:      "User:42",
					PermissionType: mdsv1.ACLPERMISSIONTYPE_ALLOW,
					Operation:      mdsv1.ACLOPERATION_READ,
					Host:           "*",
				},
			},
		}
	}()

	err = cmd.Execute()
	assert.Nil(suite.T(), err)
}
