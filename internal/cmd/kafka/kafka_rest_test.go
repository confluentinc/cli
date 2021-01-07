package kafka

import (
	"fmt"
	"net/http"
	neturl "net/url"
	"testing"

	"github.com/antihax/optional"
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	"github.com/stretchr/testify/suite"
)

type KafkaRestTestSuite struct {
	suite.Suite
}

func (suite *KafkaRestTestSuite) TestAclBindingToClustersClusterIdAclsGetOpts() {
	req := suite.Require()

	binding := schedv1.ACLBinding{
		Pattern: &schedv1.ResourcePatternConfig{
			ResourceType:         schedv1.ResourceTypes_GROUP,
			Name:                 "mygroup",
			PatternType:          schedv1.PatternTypes_PREFIXED,
			XXX_NoUnkeyedLiteral: struct{}{},
			XXX_unrecognized:     []byte{},
			XXX_sizecache:        0,
		},
		Entry: &schedv1.AccessControlEntryConfig{
			Principal:            "myprincipal",
			Operation:            schedv1.ACLOperations_CREATE,
			Host:                 "myhost",
			PermissionType:       schedv1.ACLPermissionTypes_ALLOW,
			XXX_NoUnkeyedLiteral: struct{}{},
			XXX_unrecognized:     []byte{},
			XXX_sizecache:        0,
		},
		XXX_NoUnkeyedLiteral: struct{}{},
		XXX_unrecognized:     []byte{},
		XXX_sizecache:        0,
	}

	r := aclBindingToClustersClusterIdAclsGetOpts(&binding)
	req.True(r.Host == optional.NewString("myhost"))
	req.True(r.Operation == optional.NewInterface(kafkarestv3.AclOperation("CREATE")))
	req.True(r.ResourceName == optional.NewString("mygroup"))
	req.True(r.Principal == optional.NewString("myprincipal"))
	req.True(r.Permission == optional.NewInterface(kafkarestv3.AclPermission("ALLOW")))
	req.True(r.PatternType == optional.NewInterface(kafkarestv3.AclPatternType("PREFIXED")))
}

func (suite *KafkaRestTestSuite) TestAclBindingToClustersClusterIdAclsPostOpts() {
	req := suite.Require()

	binding := schedv1.ACLBinding{
		Pattern: &schedv1.ResourcePatternConfig{
			ResourceType:         schedv1.ResourceTypes_CLUSTER,
			Name:                 "mycluster",
			PatternType:          schedv1.PatternTypes_LITERAL,
			XXX_NoUnkeyedLiteral: struct{}{},
			XXX_unrecognized:     []byte{},
			XXX_sizecache:        0,
		},
		Entry: &schedv1.AccessControlEntryConfig{
			Principal:            "myprincipal",
			Operation:            schedv1.ACLOperations_READ,
			Host:                 "myhost",
			PermissionType:       schedv1.ACLPermissionTypes_DENY,
			XXX_NoUnkeyedLiteral: struct{}{},
			XXX_unrecognized:     []byte{},
			XXX_sizecache:        0,
		},
		XXX_NoUnkeyedLiteral: struct{}{},
		XXX_unrecognized:     []byte{},
		XXX_sizecache:        0,
	}

	r := aclBindingToClustersClusterIdAclsPostOpts(&binding).CreateAclRequestData.Value().(kafkarestv3.CreateAclRequestData)
	req.True(r.Host == "myhost")
	req.True(r.Operation == kafkarestv3.AclOperation("READ"))
	req.True(r.ResourceName == "mycluster")
	req.True(r.Principal == "myprincipal")
	req.True(r.Permission == kafkarestv3.AclPermission("DENY"))
	req.True(r.PatternType == kafkarestv3.AclPatternType("LITERAL"))
}

func (suite *KafkaRestTestSuite) TestAclFilterToClustersClusterIdAclsDeleteOpts() {
	req := suite.Require()

	filter := schedv1.ACLFilter{
		PatternFilter: &schedv1.ResourcePatternConfig{
			ResourceType:         schedv1.ResourceTypes_TOPIC,
			Name:                 "mytopic",
			PatternType:          schedv1.PatternTypes_LITERAL,
			XXX_NoUnkeyedLiteral: struct{}{},
			XXX_unrecognized:     []byte{},
			XXX_sizecache:        0},
		EntryFilter: &schedv1.AccessControlEntryConfig{
			Principal:            "myprincipal",
			Operation:            schedv1.ACLOperations_WRITE,
			Host:                 "myhost",
			PermissionType:       schedv1.ACLPermissionTypes_ALLOW,
			XXX_NoUnkeyedLiteral: struct{}{},
			XXX_unrecognized:     []byte{},
			XXX_sizecache:        0,
		},
		XXX_NoUnkeyedLiteral: struct{}{},
		XXX_unrecognized:     []byte{},
		XXX_sizecache:        0,
	}

	r := aclFilterToClustersClusterIdAclsDeleteOpts(&filter)
	req.Equal(r.Host, optional.NewString("myhost"))
	req.Equal(r.Operation, optional.NewInterface(kafkarestv3.AclOperation("WRITE")))
	req.Equal(r.ResourceName, optional.NewString("mytopic"))
	req.Equal(r.Principal, optional.NewString("myprincipal"))
	req.Equal(r.Permission, optional.NewInterface(kafkarestv3.AclPermission("ALLOW")))
	req.Equal(r.PatternType, optional.NewInterface(kafkarestv3.AclPatternType("LITERAL")))
}

func (suite *KafkaRestTestSuite) TestKafkaRestError() {
	req := suite.Require()
	url := "http://my-url"
	neturlMsg := "net-error"

	neturlError := neturl.Error{
		Op:  "my-op",
		URL: url,
		Err: fmt.Errorf(neturlMsg),
	}

	r := kafkaRestError(url, &neturlError, nil)
	req.NotNil(r)
	req.Contains(r.Error(), "establish")
	req.Contains(r.Error(), url)
	req.Contains(r.Error(), neturlMsg)

	openAPIError := kafkarestv3.GenericOpenAPIError{}

	r = kafkaRestError(url, openAPIError, nil)
	req.NotNil(r)
	req.Contains(r.Error(), "Unknown")

	httpResp := http.Response{
		Status:     "Code: 400",
		StatusCode: 400,
		Request: &http.Request{
			Method: "GET",
			URL: &neturl.URL{
				Host: "myhost",
				Path: "/my-path",
			},
		},
	}
	r = kafkaRestError(url, openAPIError, &httpResp)
	req.NotNil(r)
	req.Contains(r.Error(), "failed")
	req.Contains(r.Error(), "GET")
	req.Contains(r.Error(), "myhost")
	req.Contains(r.Error(), "my-path")
}
func TestKafkaRestTestSuite(t *testing.T) {
	suite.Run(t, new(KafkaRestTestSuite))
}
