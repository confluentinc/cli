package errors

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/go-multierror"

	corev1 "github.com/confluentinc/cc-structs/kafka/core/v1"
	"github.com/confluentinc/ccloud-sdk-go"
	mds "github.com/confluentinc/mds-sdk-go/mdsv1"
)

/*
	HANDLECOMMON HELPERS
*/

func catchTypedErrors(err error) error {
	if err == nil {
		return nil
	}
	if typedErr, ok := err.(CLITypedError); ok {
		return typedErr.UserFacingError()
	}
	return err
}

func catchMDSErrors(err error) error {
	if err == nil {
		return nil
	}
	e, ok := err.(mds.GenericOpenAPIError)
	if ok {
		return Errorf(GenericOpenAPIErrorMsg, e.Error(), string(e.Body()))
	}
	return err
}

// All errors from CCloud backend services will be of corev1.Error type
// This catcher function should then be used last to not accidentally convert errors that
// are supposed to be caught by more specific catchers.
func catchCoreV1Errors(err error) error {
	if err == nil {
		return nil
	}
	e, ok := err.(*corev1.Error)
	if ok {
		var result error
		result = multierror.Append(result, e)
		return Wrap(result, CCloudBackendErrorPrefix)
	}
	return err
}

func catchCCloudTokenErrors(err error) error {
	if err == nil {
		return nil
	}
	switch err.(type) {
	case *ccloud.InvalidLoginError:
		return NewErrorWithSuggestions(InvalidLoginErrorMsg, CCloudInvalidLoginSuggestions)
	case *ccloud.InvalidTokenError:
		return NewErrorWithSuggestions(CorruptedTokenErrorMsg, CorruptedTokenSuggestions)
	case *ccloud.ExpiredTokenError:
		return NewErrorWithSuggestions(ExpiredTokenErrorMsg, ExpiredTokenSuggestions)
	}
	return err
}

/*
Error: 1 error occurred:
	* error creating ACLs: reply error: invalid character 'C' looking for beginning of value
Error: 1 error occurred:
	* error updating topic ENTERPRISE.LOANALT2-ALTERNATE-LOAN-MASTER-2.DLQ: reply error: invalid character '<' looking for beginning of value
*/
func catchCCloudBackendUnmarshallingError(err error) error {
	if err == nil {
		return nil
	}
	backendUnmarshllingErrorRegex := regexp.MustCompile(`reply error: invalid character '.' looking for beginning of value`)
	if backendUnmarshllingErrorRegex.MatchString(err.Error()) {
		errorMsg := fmt.Sprintf(prefixFormat, UnexpectedBackendOutputPrefix, BackendUnmarshallingErrorMsg)
		return NewErrorWithSuggestions(errorMsg, UnexpectedBackendOutputSuggestions)
	}
	return err
}

/*
	CCLOUD-SDK-GO CLIENT ERROR CATCHING
*/

/*
Error: 1 error occurred:
	* error checking email: User Not Found
*/
func CatchEmailNotFoundError(err error, email string) error {
	if err == nil {
		return nil
	}
	if strings.Contains(err.Error(), "error checking email: User Not Found") {
		errorMsg := fmt.Sprintf(InvalidEmailErrorMsg, email)
		return NewErrorWithSuggestions(errorMsg, InvalidEmailSuggestions)
	}
	return err
}

func CatchResourceNotFoundError(err error, resourceId string) error {
	if err == nil {
		return nil
	}
	_, isKafkaNotFound := err.(*KafkaClusterNotFoundError)
	if isResourceNotFoundError(err) || isKafkaNotFound {
		errorMsg := fmt.Sprintf(ResourceNotFoundErrorMsg, resourceId)
		suggestionsMsg := fmt.Sprintf(ResourceNotFoundSuggestions, resourceId)
		return NewErrorWithSuggestions(errorMsg, suggestionsMsg)
	}
	return err
}

func CatchKafkaNotFoundError(err error, clusterId string) error {
	if err == nil {
		return nil
	}
	if isResourceNotFoundError(err) {
		return &KafkaClusterNotFoundError{ClusterID: clusterId}
	}
	return err
}

func CatchKSQLNotFoundError(err error, clusterId string) error {
	if err == nil {
		return nil
	}
	if isResourceNotFoundError(err) {
		errorMsg := fmt.Sprintf(ResourceNotFoundErrorMsg, clusterId)
		return NewErrorWithSuggestions(errorMsg, KSQLNotFoundSuggestions)
	}
	return err
}

func CatchSchemaRegistryNotFoundError(err error, clusterId string) error {
	if err == nil {
		return nil
	}
	if isResourceNotFoundError(err) {
		errorMsg := fmt.Sprintf(ResourceNotFoundErrorMsg, clusterId)
		return NewErrorWithSuggestions(errorMsg, SRNotFoundSuggestions)
	}
	return err
}

/*
Error: 1 error occurred:
	* error describing kafka cluster: resource not found
Error: 1 error occurred:
	* error describing kafka cluster: resource not found
Error: 1 error occurred:
	* error listing schema-registry cluster: resource not found
Error: 1 error occurred:
	* error describing ksql cluster: resource not found
*/
func isResourceNotFoundError(err error) bool {
	resourceNotFoundRegex := regexp.MustCompile(`error .* cluster: resource not found`)
	return resourceNotFoundRegex.MatchString(err.Error())
}

/*
Error: 1 error occurred:
	* error creating topic bob: Topic 'bob' already exists.
*/
func CatchTopicExistsError(err error, clusterId string, topicName string, ifNotExistsFlag bool) error {
	if err == nil {
		return nil
	}
	compiledRegex := regexp.MustCompile(`error creating topic .*: Topic '.*' already exists\.`)
	if compiledRegex.MatchString(err.Error()) {
		if ifNotExistsFlag {
			return nil
		}
		errorMsg := fmt.Sprintf(TopicExistsErrorMsg, topicName, clusterId)
		suggestions := fmt.Sprintf(TopicExistsSuggestions, clusterId, clusterId)
		return NewErrorWithSuggestions(errorMsg, suggestions)
	}
	return err
}

/*
Error: 1 error occurred:
	* error listing topics: Authentication failed: 1 extensions are invalid! They are: logicalCluster: Authentication failed
Error: 1 error occurred:
	* error creating topic test-topic: Authentication failed: 1 extensions are invalid! They are: logicalCluster: Authentication failed
*/
func CatchClusterNotReadyError(err error, clusterId string) error {
	if err == nil {
		return nil
	}
	if strings.Contains(err.Error(), "Authentication failed: 1 extensions are invalid! They are: logicalCluster: Authentication failed") {
		errorMsg := fmt.Sprintf(KafkaNotReadyErrorMsg, clusterId)
		return NewErrorWithSuggestions(errorMsg, KafkaNotReadySuggestions)
	}
	return err
}

/*
	SARAMA ERROR CATCHING
*/

/*
kafka server: Request was for a topic or partition that does not exist on this broker.
*/
func CatchTopicNotExistError(err error, topicName string, clusterId string) (bool, error) {
	if err == nil {
		return false, nil
	}
	if strings.Contains(err.Error(), "kafka server: Request was for a topic or partition that does not exist on this broker.") {
		errorMsg := fmt.Sprintf(TopicNotExistsErrorMsg, topicName)
		suggestionsMsg := fmt.Sprintf(TopicNotExistsSuggestions, clusterId, clusterId)
		return true, NewErrorWithSuggestions(errorMsg, suggestionsMsg)
	}
	return false, err
}

/*
Error: "kafka: client has run out of available brokers to talk to (Is your cluster reachable?)"
*/
func CatchClusterUnreachableError(err error, clusterId string, apiKey string) error {
	if err == nil {
		return nil
	}
	if strings.Contains(err.Error(), "kafka: client has run out of available brokers to talk to (Is your cluster reachable?)") {
		suggestionsMsg := fmt.Sprintf(UnableToConnectToKafkaSuggestions, clusterId, apiKey, apiKey, clusterId)
		return NewErrorWithSuggestions(UnableToConnectToKafkaErrorMsg, suggestionsMsg)
	}
	return err
}
