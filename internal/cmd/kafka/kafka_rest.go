package kafka

import (
	"encoding/json"
	"fmt"
	"net/http"
	neturl "net/url"

	"github.com/antihax/optional"
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
)

const KafkaRestBadRequestErrorCode = 40002
const KafkaRestUnknownTopicOrPartitionErrorCode = 40403

func kafkaRestHttpError(httpResp *http.Response) error {
	return errors.NewErrorWithSuggestions(
		fmt.Sprintf(errors.KafkaRestErrorMsg, httpResp.Request.Method, httpResp.Request.URL, httpResp.Status),
		errors.InternalServerErrorSuggestions)
}

func parseOpenAPIError(err error) (*kafkaRestV3Error, error) {
	if openAPIError, ok := err.(kafkarestv3.GenericOpenAPIError); ok {
		var decodedError kafkaRestV3Error
		err = json.Unmarshal(openAPIError.Body(), &decodedError)
		if err != nil {
			return nil, err
		}
		return &decodedError, nil
	}
	return nil, fmt.Errorf("unexpected type")
}

func kafkaRestError(url string, err error, httpResp *http.Response) error {
	switch err.(type) {
	case *neturl.Error:
		if e, ok := err.(*neturl.Error); ok {
			return errors.Errorf(errors.KafkaRestConnectionMsg, url, e.Err)
		}
	case kafkarestv3.GenericOpenAPIError:
		openAPIError, parseErr := parseOpenAPIError(err)
		if parseErr == nil {
			return fmt.Errorf("REST request failed: %v (%v)", openAPIError.Message, openAPIError.Code)
		}
		if httpResp != nil && httpResp.StatusCode >= 400 {
			return kafkaRestHttpError(httpResp)
		}
		return errors.NewErrorWithSuggestions(errors.UnknownErrorMsg, errors.InternalServerErrorSuggestions)
	}
	return err
}

// Converts ACLBinding to Kafka REST ClustersClusterIdAclsGetOpts
func aclBindingToClustersClusterIdAclsGetOpts(acl *schedv1.ACLBinding) kafkarestv3.ClustersClusterIdAclsGetOpts {
	var opts kafkarestv3.ClustersClusterIdAclsGetOpts

	if acl.Pattern.ResourceType != schedv1.ResourceTypes_UNKNOWN {
		opts.ResourceType = optional.NewInterface(kafkarestv3.AclResourceType(acl.Pattern.ResourceType.String()))
	}

	opts.ResourceName = optional.NewString(acl.Pattern.Name)

	if acl.Pattern.PatternType != schedv1.PatternTypes_UNKNOWN {
		opts.PatternType = optional.NewInterface(kafkarestv3.AclPatternType(acl.Pattern.PatternType.String()))
	}

	opts.Principal = optional.NewString(acl.Entry.Principal)
	opts.Host = optional.NewString(acl.Entry.Host)

	if acl.Entry.Operation != schedv1.ACLOperations_UNKNOWN {
		opts.Operation = optional.NewInterface(kafkarestv3.AclOperation(acl.Entry.Operation.String()))
	}

	if acl.Entry.PermissionType != schedv1.ACLPermissionTypes_UNKNOWN {
		opts.Permission = optional.NewInterface(kafkarestv3.AclPermission(acl.Entry.PermissionType.String()))
	}

	return opts
}

// Converts ACLBinding to Kafka REST ClustersClusterIdAclsPostOpts
func aclBindingToClustersClusterIdAclsPostOpts(acl *schedv1.ACLBinding) kafkarestv3.ClustersClusterIdAclsPostOpts {
	var aclRequestData kafkarestv3.CreateAclRequestData

	if acl.Pattern.ResourceType != schedv1.ResourceTypes_UNKNOWN {
		aclRequestData.ResourceType = kafkarestv3.AclResourceType(acl.Pattern.ResourceType.String())
	}

	if acl.Pattern.PatternType != schedv1.PatternTypes_UNKNOWN {
		aclRequestData.PatternType = kafkarestv3.AclPatternType(acl.Pattern.PatternType.String())
	}

	aclRequestData.ResourceName = acl.Pattern.Name
	aclRequestData.Principal = acl.Entry.Principal
	aclRequestData.Host = acl.Entry.Host

	if acl.Entry.Operation != schedv1.ACLOperations_UNKNOWN {
		aclRequestData.Operation = kafkarestv3.AclOperation(acl.Entry.Operation.String())
	}

	if acl.Entry.PermissionType != schedv1.ACLPermissionTypes_UNKNOWN {
		aclRequestData.Permission = kafkarestv3.AclPermission(acl.Entry.PermissionType.String())
	}

	var opts kafkarestv3.ClustersClusterIdAclsPostOpts
	opts.CreateAclRequestData = optional.NewInterface(aclRequestData)

	return opts
}

// Converts ACLFilter to Kafka REST ClustersClusterIdAclsDeleteOpts
func aclFilterToClustersClusterIdAclsDeleteOpts(acl *schedv1.ACLFilter) kafkarestv3.ClustersClusterIdAclsDeleteOpts {
	var opts kafkarestv3.ClustersClusterIdAclsDeleteOpts

	if acl.PatternFilter.ResourceType != schedv1.ResourceTypes_UNKNOWN {
		opts.ResourceType = optional.NewInterface(kafkarestv3.AclResourceType(acl.PatternFilter.ResourceType.String()))
	}

	opts.ResourceName = optional.NewString(acl.PatternFilter.Name)

	if acl.PatternFilter.PatternType != schedv1.PatternTypes_UNKNOWN {
		opts.PatternType = optional.NewInterface(kafkarestv3.AclPatternType(acl.PatternFilter.PatternType.String()))
	}

	opts.Principal = optional.NewString(acl.EntryFilter.Principal)
	opts.Host = optional.NewString(acl.EntryFilter.Host)

	if acl.EntryFilter.Operation != schedv1.ACLOperations_UNKNOWN {
		opts.Operation = optional.NewInterface(kafkarestv3.AclOperation(acl.EntryFilter.Operation.String()))
	}

	if acl.EntryFilter.PermissionType != schedv1.ACLPermissionTypes_UNKNOWN {
		opts.Permission = optional.NewInterface(kafkarestv3.AclPermission(acl.EntryFilter.PermissionType.String()))
	}

	return opts
}
