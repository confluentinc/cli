package acl

import (
	"fmt"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	"github.com/confluentinc/cli/v4/pkg/ccstructs"
	"github.com/confluentinc/cli/v4/pkg/errors"
)

func TestParseRequest(t *testing.T) {
	suite := []struct {
		args        []string
		expectedAcl RequestDataWithError
	}{
		{
			args: []string{"--operation", "read", "--principal", "User:Alice", "--cluster-scope", "--host", "127.0.0.1", "--allow"},
			expectedAcl: RequestDataWithError{
				ResourceType: kafkarestv3.ACLRESOURCETYPE_CLUSTER,
				ResourceName: "kafka-cluster",
				PatternType:  "LITERAL",
				Principal:    "User:Alice",
				Host:         "127.0.0.1",
				Operation:    "READ",
				Permission:   "ALLOW",
				Errors:       nil,
			},
		},
		{
			args: []string{"--operation", "delete", "--principal", "Group:Admin", "--topic", "Test", "--prefix", "--deny"},
			expectedAcl: RequestDataWithError{
				ResourceType: kafkarestv3.ACLRESOURCETYPE_TOPIC,
				ResourceName: "Test",
				PatternType:  "PREFIXED",
				Principal:    "Group:Admin",
				Host:         "*",
				Operation:    "DELETE",
				Permission:   "DENY",
				Errors:       nil,
			},
		},
		{
			args: []string{"--operation", "fake", "--principal", "User:Alice", "--cluster-scope", "--transactional-id", "123"},
			expectedAcl: RequestDataWithError{
				Errors: multierror.Append(fmt.Errorf("invalid operation value: FAKE"), fmt.Errorf("exactly one of %v must be set",
					convertToFlags(kafkarestv3.ACLRESOURCETYPE_TOPIC, kafkarestv3.ACLRESOURCETYPE_GROUP,
						kafkarestv3.ACLRESOURCETYPE_CLUSTER, kafkarestv3.ACLRESOURCETYPE_TRANSACTIONAL_ID))),
			},
		},
		{
			args: []string{"--operation", "read", "--principal", "User:Alice", "--transactional-id", "123", "--allow", "--deny"},
			expectedAcl: RequestDataWithError{
				Errors: multierror.Append(fmt.Errorf("only `--allow` or `--deny` may be set when adding or deleting an ACL")),
			},
		},
	}
	req := require.New(t)
	for _, s := range suite {
		cmd := &cobra.Command{}
		cmd.Flags().AddFlagSet(Flags())
		_ = cmd.ParseFlags(s.args)
		acl := ParseRequest(cmd)
		if s.expectedAcl.Errors != nil {
			req.NotNil(acl.Errors)
			req.Equal(s.expectedAcl.Errors.Error(), acl.Errors.Error())
		} else {
			req.Nil(acl.Errors)
			req.Equal(s.expectedAcl, *acl)
		}
	}
}

func TestValidateCreateDeleteAclRequestData(t *testing.T) {
	suite := []struct {
		initialAcl  RequestDataWithError
		expectedAcl RequestDataWithError
	}{
		{
			initialAcl: RequestDataWithError{
				ResourceType: kafkarestv3.ACLRESOURCETYPE_CLUSTER,
				Permission:   "ALLOW",
			},
			expectedAcl: RequestDataWithError{
				PatternType:  "LITERAL",
				ResourceType: kafkarestv3.ACLRESOURCETYPE_CLUSTER,
				Permission:   "ALLOW",
			},
		},
		{
			initialAcl: RequestDataWithError{Host: "*"},
			expectedAcl: RequestDataWithError{Errors: multierror.Append(
				fmt.Errorf(errors.MustSetAllowOrDenyErrorMsg),
				fmt.Errorf(errors.MustSetResourceTypeErrorMsg, convertToFlags(kafkarestv3.ACLRESOURCETYPE_TOPIC, kafkarestv3.ACLRESOURCETYPE_GROUP, kafkarestv3.ACLRESOURCETYPE_CLUSTER, kafkarestv3.ACLRESOURCETYPE_TRANSACTIONAL_ID)),
			)},
		},
	}
	req := require.New(t)
	for _, s := range suite {
		validatedAcl := ValidateCreateDeleteAclRequestData(&s.initialAcl)
		if s.expectedAcl.Errors != nil {
			req.NotNil(validatedAcl.Errors)
			req.Equal(s.expectedAcl.Errors.Error(), validatedAcl.Errors.Error())
		} else {
			req.Nil(validatedAcl.Errors)
			req.Equal(s.expectedAcl, *validatedAcl)
		}
	}
}

func TestAclBindingToClustersClusterIdAclsPostOpts(t *testing.T) {
	req := require.New(t)

	binding := &ccstructs.ACLBinding{
		Pattern: &ccstructs.ResourcePatternConfig{
			ResourceType: ccstructs.ResourceTypes_CLUSTER,
			Name:         "mycluster",
			PatternType:  ccstructs.PatternTypes_LITERAL,
		},
		Entry: &ccstructs.AccessControlEntryConfig{
			Principal:      "myprincipal",
			Operation:      ccstructs.ACLOperations_READ,
			Host:           "myhost",
			PermissionType: ccstructs.ACLPermissionTypes_DENY,
		},
	}

	data := GetCreateAclRequestData(binding)
	req.True(data.Host == "myhost")
	req.True(data.Operation == "READ")
	req.True(data.ResourceName == "mycluster")
	req.True(data.Principal == "myprincipal")
	req.True(data.Permission == "DENY")
	req.True(data.PatternType == "LITERAL")
}
