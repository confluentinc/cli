package errors

import "fmt"

var (
	prefixFormat = "%s: %s"
	cliDownLoadLink = map[string]string {
		"confluent": "https://docs.confluent.io/current/cli/installing.html",
		"ccloud": "https://docs.confluent.io/current/cloud/cli/install.html",
	}
)

type CLITypedError interface {
	error
	UserFacingError() error
}

type NotLoggedInError struct {
	CLIName string
}

func (e *NotLoggedInError) Error() string {
	return e.CLIName
}

func (e *NotLoggedInError) UserFacingError() error {
	suggestionsMsg := fmt.Sprintf(NotLoggedInSuggestions, e.CLIName)
	return NewErrorWithSuggestions(NotLoggedInErrorMsg, suggestionsMsg)
}

type NoContextError struct {
	CLIName string
}

func (e *NoContextError) Error() string {
	return e.CLIName
}

func (e *NoContextError) UserFacingError() error {
	suggestionsMsg := fmt.Sprintf(NotLoggedInSuggestions, e.CLIName)
	return NewErrorWithSuggestions(NotLoggedInErrorMsg, suggestionsMsg)
}

// UnspecifiedAPIKeyError means the user needs to set an api-key for this cluster
type UnspecifiedAPIKeyError struct {
	ClusterID string
}

func (e *UnspecifiedAPIKeyError) Error() string {
	return e.ClusterID
}

func (e *UnspecifiedAPIKeyError) UserFacingError() error {
	errorMsg := fmt.Sprintf(NoAPIKeySelectedErrorMsg, e.ClusterID)
	suggestionsMsg := fmt.Sprintf(NoAPIKeySelectedSuggestions, e.ClusterID, e.ClusterID, e.ClusterID, e.ClusterID)
	return NewErrorWithSuggestions(errorMsg, suggestionsMsg)
}


// UnconfiguredAPISecretError means the user needs to store the API secret locally
type UnconfiguredAPISecretError struct {
	APIKey    string
	ClusterID string
}

func (e *UnconfiguredAPISecretError) Error() string {
	return e.APIKey
}

func (e *UnconfiguredAPISecretError) UserFacingError() error {
	errorMsg := fmt.Sprintf(NoAPISecretStoredErrorMsg, e.APIKey, e.ClusterID)
	suggestionsMsg := fmt.Sprintf(NoAPISecretStoredSuggestions, e.APIKey, e.ClusterID)
	return NewErrorWithSuggestions(errorMsg, suggestionsMsg)
}

func NewCorruptedConfigError(format string, contextName string, cliName string, configFile string) CLITypedError {
	return &CorruptedConfigError{
		messageFormat: format,
		contextName:   contextName,
		cliName:       cliName,
		configFile:    configFile,
	}
}

type CorruptedConfigError struct {
	messageFormat string
	contextName   string
	cliName       string
	configFile    string
}

func (e *CorruptedConfigError) Error() string {
	return e.contextName
}

func (e *CorruptedConfigError) UserFacingError() error {
	var errorMsg string
	if e.contextName != "" {
		errorMsg = fmt.Sprintf(e.messageFormat, e.contextName)
	} else {
		errorMsg = e.messageFormat
	}
	errorMsg = fmt.Sprintf(prefixFormat, CorruptedConfigErrorPrefix, errorMsg)
	suggestionsMsg := fmt.Sprintf(CorruptedConfigSuggestions, e.configFile, e.cliName, e.cliName)
	return NewErrorWithSuggestions(errorMsg, suggestionsMsg)
}

func NewUpdateClientWrapError(err error, errorMsg string, cliName string) CLITypedError {
	return &UpdateClientError{errorMsg:Wrap(err, errorMsg).Error(), cliName: cliName}
}

type UpdateClientError struct {
	errorMsg string
	cliName string
}

func (e *UpdateClientError) Error() string {
	return e.errorMsg
}

func (e *UpdateClientError) UserFacingError() error {
	errMsg := fmt.Sprintf(prefixFormat, UpdateClientFailurePrefix, e.errorMsg)
	suggestionsMsg := fmt.Sprintf(UpdateClientFailureSuggestions, cliDownLoadLink[e.cliName])
	return NewErrorWithSuggestions(errMsg, suggestionsMsg)
}
