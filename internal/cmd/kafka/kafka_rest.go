package kafka

import (
	"encoding/json"
	"fmt"
	"net/http"
	neturl "net/url"
	"strings"

	"github.com/antihax/optional"
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

const KafkaRestBadRequestErrorCode = 40002
const KafkaRestUnknownTopicOrPartitionErrorCode = 40403
const SelfSignedCertError = "x509: certificate is not authorized to sign other certificates"
const UnauthorizedCertError = "x509: certificate signed by unknown authority"

type kafkaRestV3Error struct {
	Code    int    `json:"error_code"`
	Message string `json:"message"`
}

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
	switch e := err.(type) {
	case *neturl.Error:
		if strings.Contains(e.Error(), SelfSignedCertError) || strings.Contains(e.Error(), UnauthorizedCertError) {
			return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.KafkaRestConnectionErrorMsg, url, e.Err), errors.KafkaRestCertErrorSuggestions)
		}
		return errors.Errorf(errors.KafkaRestConnectionErrorMsg, url, e.Err)
	case kafkarestv3.GenericOpenAPIError:
		openAPIError, parseErr := parseOpenAPIError(err)
		if parseErr == nil {
			if strings.Contains(openAPIError.Message, "invalid_token") {
				return errors.NewErrorWithSuggestions(errors.InvalidMDSTokenErrorMsg, errors.InvalidMDSTokenSuggestions)
			}
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
