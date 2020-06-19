package errors

var (
	ConfigUnableToLoadError                         = "Unable to load config: %s"
	ConfigUnspecifiedPlatformError                  = "Context \"%s\" has a corrupted platform. To fix, please remove the config file, and run `login` or `init`."
	ConfigUnspecifiedCredentialError                = "Context \"%s\" has corrupted credentials. To fix, please remove the config file, and run `login` or `init`."
	UserNotLoggedInErrMsg                           = "You must log in to run that command."
	CorruptedAuthTokenErrorMsg                      = "Your auth token has been corrupted. Please login again."
	NotLoggedInInternalErrorMsg                     = "not logged in"
	APIKeyCommandResourceTypeNotImplementedErrorMsg = "Command not yet available for non-Kafka cluster resources."
)
