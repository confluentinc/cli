package acl

import (
	"fmt"
	"testing"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/internal/pkg/ccstructs"
	errMsgs "github.com/confluentinc/cli/internal/pkg/errors"
)

func TestParseAclRequest(t *testing.T) {
	suite := []struct {
		args        []string
		expectedAcl AclRequestDataWithError
	}{
		{
			args: []string{"--operation", "read", "--principal", "User:Alice", "--cluster-scope", "--host", "127.0.0.1", "--allow"},
			expectedAcl: AclRequestDataWithError{
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
			expectedAcl: AclRequestDataWithError{
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
			expectedAcl: AclRequestDataWithError{
				Errors: multierror.Append(errors.New("invalid operation value: FAKE"), fmt.Errorf("exactly one of %v must be set",
					convertToFlags(kafkarestv3.ACLRESOURCETYPE_TOPIC, kafkarestv3.ACLRESOURCETYPE_GROUP,
						kafkarestv3.ACLRESOURCETYPE_CLUSTER, kafkarestv3.ACLRESOURCETYPE_TRANSACTIONAL_ID))),
			},
		},
		{
			args: []string{"--operation", "read", "--principal", "User:Alice", "--transactional-id", "123", "--allow", "--deny"},
			expectedAcl: AclRequestDataWithError{
				Errors: multierror.Append(errors.Errorf(errMsgs.OnlySetAllowOrDenyErrorMsg)),
			},
		},
	}
	req := require.New(t)
	for _, s := range suite {
		cmd := &cobra.Command{}
		cmd.Flags().AddFlagSet(AclFlags())
		_ = cmd.ParseFlags(s.args)
		acl := ParseAclRequest(cmd)
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
		initialAcl  AclRequestDataWithError
		expectedAcl AclRequestDataWithError
	}{
		{
			initialAcl: AclRequestDataWithError{
				ResourceType: kafkarestv3.ACLRESOURCETYPE_CLUSTER,
				Permission:   "ALLOW",
			},
			expectedAcl: AclRequestDataWithError{
				PatternType:  "LITERAL",
				ResourceType: kafkarestv3.ACLRESOURCETYPE_CLUSTER,
				Permission:   "ALLOW",
			},
		},
		{
			initialAcl: AclRequestDataWithError{Host: "*"},
			expectedAcl: AclRequestDataWithError{Errors: multierror.Append(errors.Errorf(errMsgs.MustSetAllowOrDenyErrorMsg), errors.Errorf(errMsgs.MustSetResourceTypeErrorMsg,
				convertToFlags(kafkarestv3.ACLRESOURCETYPE_TOPIC, kafkarestv3.ACLRESOURCETYPE_GROUP,
					kafkarestv3.ACLRESOURCETYPE_CLUSTER, kafkarestv3.ACLRESOURCETYPE_TRANSACTIONAL_ID)))},
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

func TestGetPrefixAndResourceIdFromPrincipal_Empty(t *testing.T) {
	prefix, resourceId, err := getPrefixAndResourceIdFromPrincipal("", nil)
	require.NoError(t, err)
	require.Equal(t, "", prefix)
	require.Equal(t, "", resourceId)
}

func TestGetPrefixAndResourceIdFromPrincipal_UnrecognizedFormat(t *testing.T) {
	_, _, err := getPrefixAndResourceIdFromPrincipal("string with no colon", nil)
	require.Error(t, err)
}

func TestGetPrefixAndResourceIdFromPrincipal_ResourceId(t *testing.T) {
	prefix, resourceId, err := getPrefixAndResourceIdFromPrincipal("User:sa-123456", nil)
	require.NoError(t, err)
	require.Equal(t, "User", prefix)
	require.Equal(t, "sa-123456", resourceId)
}

func TestGetPrefixAndResourceIdFromPrincipal_NumericId(t *testing.T) {
	prefix, resourceId, err := getPrefixAndResourceIdFromPrincipal("User:123456", map[int32]string{123456: "sa-123456"})
	require.NoError(t, err)
	require.Equal(t, "User", prefix)
	require.Equal(t, "sa-123456", resourceId)
}

func TestGetPrefixAndResourceIdFromPrincipal_UserIdNotValid(t *testing.T) {
	for _, principal := range []string{"User:123456", "User:abcdef"} {
		_, _, err := getPrefixAndResourceIdFromPrincipal(principal, nil)
		require.Error(t, err)
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
