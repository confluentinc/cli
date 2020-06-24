package errors

import (
	"fmt"
	"regexp"
	"strings"
)

/*
	ccloud-sdk-go errors
 */

func CatchResourceNotFoundError(err error, resourceId string) error {
	if isResourceNotFoundError(err) {
		errorMsg := fmt.Sprintf(ResourceNotFoundErrorMsg, resourceId)
		suggestionsMsg := fmt.Sprintf(ResourceNotFoundSuggestions, resourceId)
		return NewErrorWithSuggestions(errorMsg, suggestionsMsg)
	}
	return err
}

func CatchKafkaNotFoundError(err error, clusterId string) error {
	if isResourceNotFoundError(err) {
		errorMsg := fmt.Sprintf(ResourceNotFoundErrorMsg, clusterId)
		suggestionsMsg := fmt.Sprintf(KafkaNotFoundSuggestions)
		return NewErrorWithSuggestions(errorMsg, suggestionsMsg)
	}
	return err
}

func CatchKSQLNotFoundError(err error, clusterId string) error {
	if isResourceNotFoundError(err) {
		errorMsg := fmt.Sprintf(ResourceNotFoundErrorMsg, clusterId)
		return NewErrorWithSuggestions(errorMsg, KSQLNotFoundSuggestions)
	}
	return err
}

func CatchSchemaRegistryNotFoundError(err error, clusterId string) error {
	if isResourceNotFoundError(err) {
		errorMsg := fmt.Sprintf(ResourceNotFoundErrorMsg, clusterId)
		return NewErrorWithSuggestions(errorMsg, SchemaRegistryNotFoundSuggestions)
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
	Sarama Errors
 */

func CatchTopicNotExistError(err error, topicName string, clusterId string) (bool, error) {
	if strings.Contains(err.Error(), "kafka server: Request was for a topic or partition that does not exist on this broker.") {
		errorMsg := fmt.Sprintf(TopicNotExistsErrorMsg, topicName)
		suggestionsMsg := fmt.Sprintf(TopicNotExistsSuggestions, clusterId, clusterId)
		return true, NewErrorWithSuggestions(errorMsg, suggestionsMsg)
	}
	return false, err
}

func CatchClusterUnreachableError(err error, clusterId string, apiKey string) error {
	if strings.Contains(err.Error(), "kafka: client has run out of available brokers to talk to (Is your cluster reachable?)") {
		suggestionsMsg := fmt.Sprintf(UnableToConnectToKafkaSuggestions, clusterId, apiKey, apiKey, clusterId)
		return NewErrorWithSuggestions(UnableToConnectToKafkaErrorMsg, suggestionsMsg)
	}
	return err
}


