package errors

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"regexp"
	"strings"

	"github.com/pkg/errors"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"
	"github.com/confluentinc/mds-sdk-go-public/mdsv1"
	"github.com/confluentinc/mds-sdk-go-public/mdsv2alpha1"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"
)

/*
	HANDLECOMMON HELPERS
	see: https://github.com/confluentinc/cli/v3/blob/master/errors.md
*/

const quotaExceededRegex = ".* is currently limited to .*"

type errorResponseBody struct {
	Errors  []errorDetail `json:"errors"`
	Error   errorBody     `json:"error"`
	Message string        `json:"message"`
}

type errorDetail struct {
	Detail     string `json:"detail"`
	Resolution string `json:"resolution"`
}

type errorBody struct {
	Message string `json:"message"`
}

func catchTypedErrors(err error) error {
	if typedErr, ok := err.(CLITypedError); ok {
		return typedErr.UserFacingError()
	}
	return err
}

func parseMDSOpenAPIErrorType1(err error) (*MDSV2Alpha1ErrorType1, error) {
	if openAPIError, ok := err.(mdsv2alpha1.GenericOpenAPIError); ok {
		var decodedError MDSV2Alpha1ErrorType1
		if err := json.Unmarshal(openAPIError.Body(), &decodedError); err != nil {
			return nil, err
		}
		if reflect.DeepEqual(decodedError, MDSV2Alpha1ErrorType1{}) {
			return nil, fmt.Errorf("failed to parse")
		}
		return &decodedError, nil
	}
	return nil, fmt.Errorf("unexpected type")
}

func parseMDSOpenAPIErrorType2(err error) (*MDSV2Alpha1ErrorType2Array, error) {
	if openAPIError, ok := err.(mdsv2alpha1.GenericOpenAPIError); ok {
		var decodedError MDSV2Alpha1ErrorType2Array
		if err := json.Unmarshal(openAPIError.Body(), &decodedError); err != nil {
			return nil, err
		}
		if reflect.DeepEqual(decodedError, MDSV2Alpha1ErrorType2Array{}) {
			return nil, fmt.Errorf("failed to parse")
		}
		return &decodedError, nil
	}
	return nil, fmt.Errorf("unexpected type")
}

func catchMDSErrors(err error) error {
	switch err2 := err.(type) {
	case mdsv1.GenericOpenAPIError:
		return Errorf(GenericOpenApiErrorMsg, err.Error(), string(err2.Body()))
	case mdsv2alpha1.GenericOpenAPIError:
		if strings.Contains(err.Error(), "Forbidden Access") {
			return NewErrorWithSuggestions("user is unauthorized to perform this action", "Check the user's privileges by running `confluent iam rbac role-binding list`.\nGive the user the appropriate permissions using `confluent iam rbac role-binding create`.")
		}
		openAPIError, parseErr := parseMDSOpenAPIErrorType1(err)
		if parseErr == nil {
			return openAPIError.UserFacingError()
		} else {
			openAPIErrorType2, parseErr2 := parseMDSOpenAPIErrorType2(err)
			if parseErr2 == nil {
				return openAPIErrorType2.UserFacingError()
			} else {
				return Errorf(GenericOpenApiErrorMsg, err.Error(), string(err2.Body()))
			}
		}
	}
	return err
}

// All errors from CCloud backend services will be of ccloudv1.Error type
// This catcher function should then be used last to not accidentally convert errors that
// are supposed to be caught by more specific catchers.
func catchCcloudV1Errors(err error) error {
	if err, ok := err.(*ccloudv1.Error); ok {
		return Wrap(err, "Confluent Cloud backend error")
	}
	return err
}

func catchCCloudTokenErrors(err error) error {
	switch err.(type) {
	case *ccloudv1.InvalidLoginError:
		return NewErrorWithSuggestions(InvalidLoginErrorMsg, InvalidLoginErrorSuggestions)
	case *ccloudv1.InvalidTokenError:
		return NewErrorWithSuggestions(CorruptedTokenErrorMsg, CorruptedTokenSuggestions)
	}
	return err
}

func catchOpenAPIError(err error) error {
	if openAPIError, ok := err.(srsdk.GenericOpenAPIError); ok {
		body := string(openAPIError.Body())

		r := strings.NewReader(body)

		formattedErr := &struct {
			ErrorCode int    `json:"error_code"`
			Message   string `json:"message"`
		}{}

		if err := json.NewDecoder(r).Decode(formattedErr); err == nil {
			return New(formattedErr.Message)
		}

		return New(body)
	}

	return err
}

/*
error creating ACLs: reply error: invalid character 'C' looking for beginning of value
error updating topic ENTERPRISE.LOANALT2-ALTERNATE-LOAN-MASTER-2.DLQ: reply error: invalid character '<' looking for beginning of value
*/
func catchCCloudBackendUnmarshallingError(err error) error {
	if regexp.MustCompile(`reply error: invalid character '.' looking for beginning of value`).MatchString(err.Error()) {
		return NewErrorWithSuggestions(
			"unexpected CCloud backend output: protobuf unmarshalling error",
			"Please submit a support ticket.",
		)
	}
	return err
}

/*
	CCLOUD-SDK-GO CLIENT ERROR CATCHING
*/

func CatchCCloudV2Error(err error, r *http.Response) error {
	if err == nil || r == nil {
		return err
	}

	body, _ := io.ReadAll(r.Body)
	var resBody errorResponseBody
	_ = json.Unmarshal(body, &resBody)
	if len(resBody.Errors) > 0 {
		detail := resBody.Errors[0].Detail
		if ok, _ := regexp.MatchString(quotaExceededRegex, detail); ok {
			return NewErrorWithSuggestions(detail, "Look up Confluent Cloud service quota limits with `confluent service-quota list`.")
		}
		if detail != "" {
			err = errors.Wrap(err, strings.TrimSuffix(detail, "\n"))
			if resolution := strings.TrimSuffix(resBody.Errors[0].Resolution, "\n"); resolution != "" {
				err = NewErrorWithSuggestions(err.Error(), resolution)
			}
			return err
		}
	}

	if resBody.Message != "" {
		return Wrap(err, strings.TrimRight(resBody.Message, "\n"))
	}

	if resBody.Error.Message != "" {
		errorMessage := strings.TrimFunc(resBody.Error.Message, func(c rune) bool {
			return c == rune('.') || c == rune('\n')
		})
		return Wrap(err, errorMessage)
	}

	return err
}

func CatchResourceNotFoundError(err error, resourceId string) error {
	if err == nil {
		return nil
	}

	if _, ok := err.(*KafkaClusterNotFoundError); ok || isResourceNotFoundError(err) {
		errorMsg := fmt.Sprintf(ResourceNotFoundErrorMsg, resourceId)
		suggestionsMsg := fmt.Sprintf(ResourceNotFoundSuggestions, resourceId)
		return NewErrorWithSuggestions(errorMsg, suggestionsMsg)
	}

	return err
}

func CatchCCloudV2ResourceNotFoundError(err error, resourceType string, r *http.Response) error {
	if err == nil {
		return nil
	}

	if r != nil && r.StatusCode == http.StatusForbidden {
		return NewWrapErrorWithSuggestions(CatchCCloudV2Error(err, r), fmt.Sprintf("%s not found or access forbidden", resourceType), fmt.Sprintf(ListResourceSuggestions, resourceType, resourceType))
	}

	return CatchCCloudV2Error(err, r)
}

func CatchComputePoolNotFoundError(err error, computePoolId string, r *http.Response) error {
	if err == nil {
		return nil
	}

	if r != nil && r.StatusCode == http.StatusForbidden {
		return NewWrapErrorWithSuggestions(
			CatchCCloudV2Error(err, r),
			fmt.Sprintf(`Flink compute pool "%s" not found or access forbidden`, computePoolId),
			"List available Flink compute pools with `confluent flink compute-pool list`.\nMake sure you have selected the compute pool's environment with `confluent environment use`.",
		)
	}

	return CatchCCloudV2Error(err, r)
}

func CatchKafkaNotFoundError(err error, clusterId string, r *http.Response) error {
	if err == nil {
		return nil
	}
	if isResourceNotFoundError(err) {
		return &KafkaClusterNotFoundError{ClusterID: clusterId}
	}

	if r != nil && r.StatusCode == http.StatusForbidden {
		return NewWrapErrorWithSuggestions(
			CatchCCloudV2Error(err, r),
			fmt.Sprintf(`Kafka cluster "%s" not found or access forbidden`, clusterId),
			ChooseRightEnvironmentSuggestions+"\nThe active Kafka cluster may have been deleted. Set a new active cluster with `confluent kafka cluster use`.",
		)
	}

	return CatchCCloudV2Error(err, r)
}

func CatchClusterConfigurationNotValidError(err error, r *http.Response) error {
	if err == nil {
		return nil
	}

	if r == nil {
		return err
	}

	err = CatchCCloudV2Error(err, r)
	if strings.Contains(err.Error(), "CKU must be greater") {
		return New("CKU must be greater than 1 for multi-zone dedicated clusters")
	}
	if strings.Contains(err.Error(), "Durability must be HIGH for an Enterprise cluster") {
		return New(`availability must be "multi-zone" for enterprise clusters`)
	}

	return err
}

func CatchApiKeyForbiddenAccessError(err error, operation string, r *http.Response) error {
	if r != nil && r.StatusCode == http.StatusForbidden || strings.Contains(err.Error(), "Unknown API key") {
		return NewWrapErrorWithSuggestions(CatchCCloudV2Error(err, r), fmt.Sprintf("error %s API key", operation), ApiKeyNotFoundSuggestions)
	}
	return CatchCCloudV2Error(err, r)
}

func CatchByokKeyNotFoundError(err error, r *http.Response) error {
	if err == nil {
		return nil
	}

	if r != nil && r.StatusCode == http.StatusNotFound {
		return NewWrapErrorWithSuggestions(CatchCCloudV2Error(err, r), "Self-managed key not found or access forbidden", ByokKeyNotFoundSuggestions)
	}

	return CatchCCloudV2Error(err, r)
}

func CatchKSQLNotFoundError(err error, clusterId string) error {
	if err == nil {
		return nil
	}

	if isResourceNotFoundError(err) {
		return NewErrorWithSuggestions(
			fmt.Sprintf(ResourceNotFoundErrorMsg, clusterId),
			"To list KSQL clusters, use `confluent ksql cluster list`.",
		)
	}

	return err
}

func CatchServiceNameInUseError(err error, r *http.Response, serviceName string) error {
	if err == nil {
		return nil
	}

	if r == nil {
		return err
	}

	err = CatchCCloudV2Error(err, r)
	if strings.Contains(err.Error(), "Service name is already in use") {
		return NewErrorWithSuggestions(fmt.Sprintf(`service name "%s" is already in use`, serviceName), "To list all service account, use `confluent iam service-account list`.")
	}

	return err
}

func CatchServiceAccountNotFoundError(err error, r *http.Response, serviceAccountId string) error {
	if err == nil {
		return nil
	}

	if r != nil {
		switch r.StatusCode {
		case http.StatusNotFound:
			errorMsg := fmt.Sprintf(ServiceAccountNotFoundErrorMsg, serviceAccountId)
			return NewErrorWithSuggestions(errorMsg, ServiceAccountNotFoundSuggestions)
		case http.StatusForbidden:
			return NewWrapErrorWithSuggestions(CatchCCloudV2Error(err, r), "service account not found or access forbidden", ServiceAccountNotFoundSuggestions)
		}
	}

	return CatchCCloudV2Error(err, r)
}

func isResourceNotFoundError(err error) bool {
	return strings.Contains(err.Error(), "resource not found")
}

/*
failed to produce offset -1: Unknown error, how did this happen? Error code = 87
*/
func CatchProduceToCompactedTopicError(err error, topicName string) (bool, error) {
	if err == nil {
		return false, nil
	}
	compiledRegex := regexp.MustCompile(`Unknown error, how did this happen\? Error code = 87`)
	if compiledRegex.MatchString(err.Error()) {
		return true, NewErrorWithSuggestions(
			fmt.Sprintf("producer has detected an INVALID_RECORD error for topic %s", topicName),
			"If the topic has schema validation enabled, ensure you are producing with a schema-enabled producer.\n"+
				"If your topic is compacted, ensure you are producing a record with a key.",
		)
	}
	return false, err
}
