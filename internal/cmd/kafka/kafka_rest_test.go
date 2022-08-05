package kafka

import (
	"testing"

	"github.com/spf13/cobra"

	"github.com/antihax/optional"
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	"github.com/stretchr/testify/suite"

	"github.com/confluentinc/cli/internal/pkg/acl"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
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
	req.True(r.Operation == optional.NewString("CREATE"))
	req.True(r.ResourceName == optional.NewString("mygroup"))
	req.True(r.Principal == optional.NewString("myprincipal"))
	req.True(r.Permission == optional.NewString("ALLOW"))
	req.True(r.PatternType == optional.NewString("PREFIXED"))
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

	r := acl.GetCreateAclRequestData(&binding)
	req.True(r.Host == "myhost")
	req.True(r.Operation == "READ")
	req.True(r.ResourceName == "mycluster")
	req.True(r.Principal == "myprincipal")
	req.True(r.Permission == "DENY")
	req.True(r.PatternType == "LITERAL")
}

func (suite *KafkaRestTestSuite) TestSetServerURL() {
	req := suite.Require()
	cmd := cobra.Command{Use: "command"}
	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	cmd.Flags().CountP("verbose", "v", "verbosity")
	client := kafkarestv3.NewAPIClient(kafkarestv3.NewConfiguration())

	setServerURL(&cmd, client, "localhost:8090")
	req.Equal("http://localhost:8090/v3", client.GetConfig().BasePath)

	setServerURL(&cmd, client, "localhost:8090/kafka/v3/")
	req.Equal("http://localhost:8090/kafka/v3", client.GetConfig().BasePath)

	setServerURL(&cmd, client, "localhost:8090/")
	req.Equal("http://localhost:8090/v3", client.GetConfig().BasePath)

	_ = cmd.Flags().Set("client-cert-path", "path")
	setServerURL(&cmd, client, "localhost:8090/kafka")
	req.Equal("https://localhost:8090/kafka/v3", client.GetConfig().BasePath)

	_ = cmd.Flags().Set("client-cert-path", "")
	_ = cmd.Flags().Set("ca-cert-path", "path")
	setServerURL(&cmd, client, "localhost:8090/kafka")
	req.Equal("https://localhost:8090/kafka/v3", client.GetConfig().BasePath)
}

func TestKafkaRestTestSuite(t *testing.T) {
	suite.Run(t, new(KafkaRestTestSuite))
}
