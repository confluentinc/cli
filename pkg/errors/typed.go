package errors

import (
	"fmt"

	"github.com/confluentinc/cli/v3/pkg/log"
)

const parsedGenericOpenApiErrorMsg = "metadata service backend error: %s"

type CLITypedError interface {
	error
	UserFacingError() error
}

type NotLoggedInError struct{}

func (e *NotLoggedInError) Error() string {
	return NotLoggedInErrorMsg
}

func (e *NotLoggedInError) UserFacingError() error {
	return NewErrorWithSuggestions(NotLoggedInErrorMsg, "You must be logged in to run this command.\n"+AvoidTimeoutSuggestions)
}

type EndOfFreeTrialError struct {
	OrgId string
}

func (e *EndOfFreeTrialError) Error() string {
	return fmt.Sprintf(EndOfFreeTrialErrorMsg, e.OrgId)
}

func (e *EndOfFreeTrialError) UserFacingError() error {
	return NewErrorWithSuggestions(fmt.Sprintf(EndOfFreeTrialErrorMsg, e.OrgId), EndOfFreeTrialSuggestions)
}

type SRNotAuthenticatedError struct{}

func (e *SRNotAuthenticatedError) Error() string {
	return SRNotAuthenticatedErrorMsg
}

func (e *SRNotAuthenticatedError) UserFacingError() error {
	return NewErrorWithSuggestions(SRNotAuthenticatedErrorMsg, "You must specify the endpoint for a Schema Registry cluster (`--schema-registry-endpoint`) or be logged in using `confluent login` to run this command.\n"+AvoidTimeoutSuggestions)
}

type SRNotEnabledError struct {
	ErrorMsg       string
	SuggestionsMsg string
}

func NewSRNotEnabledError() CLITypedError {
	return &SRNotEnabledError{
		ErrorMsg: "Schema Registry not enabled",
		SuggestionsMsg: "Schema Registry must be enabled for the environment in order to run the command.\n" +
			"You can enable Schema Registry for this environment with `confluent schema-registry cluster enable`.",
	}
}

func (e *SRNotEnabledError) Error() string {
	return e.ErrorMsg
}

func (e *SRNotEnabledError) UserFacingError() error {
	return NewErrorWithSuggestions(e.ErrorMsg, e.SuggestionsMsg)
}

type KafkaClusterNotFoundError struct {
	ClusterID string
}

func (e *KafkaClusterNotFoundError) Error() string {
	return e.ClusterID
}

func (e *KafkaClusterNotFoundError) UserFacingError() error {
	return NewErrorWithSuggestions(
		fmt.Sprintf(`Kafka cluster "%s" not found`, e.ClusterID),
		"To list Kafka clusters, use `confluent kafka cluster list`.",
	)
}

// UnspecifiedAPIKeyError means the user needs to set an api-key for this cluster
type UnspecifiedAPIKeyError struct {
	ClusterID string
}

func (e *UnspecifiedAPIKeyError) Error() string {
	return e.ClusterID
}

func (e *UnspecifiedAPIKeyError) UserFacingError() error {
	errorMsg := fmt.Sprintf(`no API key selected for resource "%s"`, e.ClusterID)
	suggestionsMsg := fmt.Sprintf("Select an API key for resource \"%s\" with `confluent api-key use <API_KEY>`.\n"+
		"To do so, you must have either already created or stored an API key for the resource.\n"+
		"To create an API key, use `confluent api-key create --resource %s`.\n"+
		"To store an existing API key, use `confluent api-key store --resource %s`.", e.ClusterID, e.ClusterID, e.ClusterID)
	return NewErrorWithSuggestions(errorMsg, suggestionsMsg)
}

// UnconfiguredAPISecretError means the user needs to store the API secret locally
type UnconfiguredAPISecretError struct {
	APIKey    string
	ClusterID string
}

func (e *UnconfiguredAPISecretError) Error() string {
	return e.UserFacingError().Error()
}

func (e *UnconfiguredAPISecretError) UserFacingError() error {
	errorMsg := fmt.Sprintf(NoAPISecretStoredErrorMsg, e.APIKey, e.ClusterID)
	suggestionsMsg := fmt.Sprintf(NoAPISecretStoredSuggestions, e.APIKey, e.ClusterID)
	return NewErrorWithSuggestions(errorMsg, suggestionsMsg)
}

func NewCorruptedConfigError(format, contextName, configFile string) CLITypedError {
	var errorWithStackTrace error
	if contextName != "" {
		errorWithStackTrace = fmt.Errorf(format, contextName)
	} else {
		errorWithStackTrace = fmt.Errorf(format)
	}
	// logging stack trace of the error use pkg/errors error type
	log.CliLogger.Debugf("%+v", errorWithStackTrace)
	return &CorruptedConfigError{
		errorMsg: fmt.Sprintf("corrupted CLI config: %v", errorWithStackTrace),
		suggestionsMsg: fmt.Sprintf("Your configuration file \"%s\" is corrupted.\n"+
			"Remove config file, and run `confluent login` or `confluent context create`.\n"+
			"Unfortunately, your active CLI state will be lost as a result.\n"+
			"Please file a support ticket with details about your configuration file to help us address this issue.\n"+
			"Please rerun the command with `--unsafe-trace` and attach the output with the support ticket.", configFile),
	}
}

type CorruptedConfigError struct {
	errorMsg       string
	suggestionsMsg string
}

func (e *CorruptedConfigError) Error() string {
	return e.errorMsg
}

func (e *CorruptedConfigError) UserFacingError() error {
	return NewErrorWithSuggestions(e.errorMsg, e.suggestionsMsg)
}

func NewUpdateClientWrapError(err error, errorMsg string) CLITypedError {
	return &UpdateClientError{errorMsg: fmt.Sprintf("%s: %v", errorMsg, err)}
}

type UpdateClientError struct {
	errorMsg string
}

func (e *UpdateClientError) Error() string {
	return e.errorMsg
}

func (e *UpdateClientError) UserFacingError() error {
	return NewErrorWithSuggestions(
		fmt.Sprintf("update client failure: %s", e.errorMsg),
		"Please submit a support ticket.\nIn the meantime, see link for other ways to download the latest CLI version:\nhttps://docs.confluent.io/current/cli/installing.html",
	)
}

type MDSV2Alpha1ErrorType1 struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
	Type       string `json:"type"`
	Err        error
}

func (e *MDSV2Alpha1ErrorType1) Error() string { return e.Message }

func (e *MDSV2Alpha1ErrorType1) UserFacingError() error {
	return fmt.Errorf(parsedGenericOpenApiErrorMsg, e.Message)
}

type MDSV2Alpha1ErrorType2 struct {
	Id     string   `json:"id"`
	Status string   `json:"status"`
	Code   string   `json:"code"`
	Detail string   `json:"detail"`
	Source []string `json:"type"`
	Err    error
}

func (e *MDSV2Alpha1ErrorType2) Error() string { return e.Detail }

type MDSV2Alpha1ErrorType2Array struct {
	Errors []MDSV2Alpha1ErrorType2 `json:"errors"`
	Err    error
}

func (e *MDSV2Alpha1ErrorType2Array) Error() string {
	errorMessage := ""
	for _, error := range e.Errors {
		errorMessage = fmt.Sprintf("%s %s", error.Error(), errorMessage)
	}
	return errorMessage
}

func (e *MDSV2Alpha1ErrorType2Array) UserFacingError() error {
	return fmt.Errorf(parsedGenericOpenApiErrorMsg, e.Error())
}
