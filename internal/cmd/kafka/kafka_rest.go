package kafka

import (
	"github.com/antihax/optional"
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
)

// Converts ACLBinding to Kafka REST ClustersClusterIdAclsGetOpts
func aclBindingToClustersClusterIdAclsGetOpts(acl *schedv1.ACLBinding) kafkarestv3.GetKafkaAclsOpts {
	var opts kafkarestv3.GetKafkaAclsOpts

	if acl.Pattern.ResourceType != schedv1.ResourceTypes_UNKNOWN {
		opts.ResourceType = optional.NewInterface(kafkarestv3.AclResourceType(acl.Pattern.ResourceType.String()))
	}

	opts.ResourceName = optional.NewString(acl.Pattern.Name)

	if acl.Pattern.PatternType != schedv1.PatternTypes_UNKNOWN {
		opts.PatternType = optional.NewString(acl.Pattern.PatternType.String())
	}

	opts.Principal = optional.NewString(acl.Entry.Principal)
	opts.Host = optional.NewString(acl.Entry.Host)

	if acl.Entry.Operation != schedv1.ACLOperations_UNKNOWN {
		opts.Operation = optional.NewString(acl.Entry.Operation.String())
	}

	if acl.Entry.PermissionType != schedv1.ACLPermissionTypes_UNKNOWN {
		opts.Permission = optional.NewString(acl.Entry.PermissionType.String())
	}

	return opts
}
