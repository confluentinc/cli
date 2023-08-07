package errors

import (
	"fmt"

	"github.com/confluentinc/cli/internal/pkg/log"
)

type CLITypedError interface {
	error
	UserFacingError() error
}

type NotLoggedInError struct{}

func (e *NotLoggedInError) Error() string {
	return NotLoggedInErrorMsg
}

func (e *NotLoggedInError) UserFacingError() error {
	return NewErrorWithSuggestions(NotLoggedInErrorMsg, NotLoggedInSuggestions)
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
	return NewErrorWithSuggestions(SRNotAuthenticatedErrorMsg, SRNotAuthenticatedSuggestions)
}

type SRNotEnabledError struct {
	ErrorMsg       string
	SuggestionsMsg string
}

func NewSRNotEnabledError() CLITypedError {
	return &SRNotEnabledError{ErrorMsg: SRNotEnabledErrorMsg, SuggestionsMsg: SRNotEnabledSuggestions}
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
	errMsg := fmt.Sprintf(KafkaNotFoundErrorMsg, e.ClusterID)
	return NewErrorWithSuggestions(errMsg, KafkaNotFoundSuggestions)
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
	errorMsg := fmt.Sprintf(NoAPISecretStoredErrorMsg, e.APIKey, e.ClusterID)
	suggestionsMsg := fmt.Sprintf(NoAPISecretStoredSuggestions, e.APIKey, e.ClusterID)
	return NewErrorWithSuggestions(errorMsg, suggestionsMsg).Error()
}

func (e *UnconfiguredAPISecretError) UserFacingError() error {
	errorMsg := fmt.Sprintf(NoAPISecretStoredErrorMsg, e.APIKey, e.ClusterID)
	suggestionsMsg := fmt.Sprintf(NoAPISecretStoredSuggestions, e.APIKey, e.ClusterID)
	return NewErrorWithSuggestions(errorMsg, suggestionsMsg)
}

func NewCorruptedConfigError(format, contextName, configFile string) CLITypedError {
	e := &CorruptedConfigError{}
	var errorWithStackTrace error
	if contextName != "" {
		errorWithStackTrace = Errorf(format, contextName)
	} else {
		errorWithStackTrace = Errorf(format)
	}
	// logging stack trace of the error use pkg/errors error type
	log.CliLogger.Debugf("%+v", errorWithStackTrace)
	e.errorMsg = fmt.Sprintf(prefixFormat, CorruptedConfigErrorPrefix, errorWithStackTrace.Error())
	e.suggestionsMsg = fmt.Sprintf(CorruptedConfigSuggestions, configFile)
	return e
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
	return &UpdateClientError{errorMsg: Wrap(err, errorMsg).Error()}
}

type UpdateClientError struct {
	errorMsg string
}

func (e *UpdateClientError) Error() string {
	return e.errorMsg
}

func (e *UpdateClientError) UserFacingError() error {
	errMsg := fmt.Sprintf(prefixFormat, UpdateClientFailurePrefix, e.errorMsg)
	return NewErrorWithSuggestions(errMsg, UpdateClientFailureSuggestions)
}

type MDSV2Alpha1ErrorType1 struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
	Type       string `json:"type"`
	Err        error
}

func (e *MDSV2Alpha1ErrorType1) Error() string { return e.Message }

func (e *MDSV2Alpha1ErrorType1) UserFacingError() error {
	return Errorf(ParsedGenericOpenAPIErrorMsg, e.Message)
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
		errorMessage = fmt.Sprintf("%s ", error.Error()) + errorMessage
	}
	return errorMessage
}

func (e *MDSV2Alpha1ErrorType2Array) UserFacingError() error {
	return Errorf(ParsedGenericOpenAPIErrorMsg, e.Error())
}
