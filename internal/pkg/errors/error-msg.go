package errors

var (
	ConfigUnableToLoadError          = "Unable to load config: %s"
	ConfigUnspecifiedPlatformError   = "Context \"%s\" has a corrupted platform. To fix, please remove the config file, and run `login` or `init`."
	ConfigUnspecifiedCredentialError = "Context \"%s\" has corrupted credentials. To fix, please remove the config file, and run `login` or `init`."
	UserNotLoggedInErrMsg            = "You must log in to run that command."
	CorruptedAuthTokenErrorMsg       = "Your auth token has been corrupted. Please login again."
	NotLoggedInInternalErrorMsg      = "not logged in"
)


var (
	ProhibitedFlagCombinationErrorMsg = "cannot use `--%s` and `--%s` flags at the same time"

	OutputWriterInvalidFormatFlagErrorMsg = "invalid `format` flag value \"%s\""

	APIKeyCommandUnableToStoreAPIKeyErrorMsg          = "unable to store API key locally"
	APIKeyCommandResourceTypeNotImplementedErrorMsg   = "command not yet available for non-Kafka cluster resources"
	APIKeyStoreRefuseToOverrideExistingSecretErrorMsg = "refusing to overwrite existing secret for API Key \"%s\""
	APIKeyUseFailedErrorMsg = "unable to set active API key"

	LoginUnableToSaveUserAuthErrorMsg = "unable to save user authentication"
	LoginNoEnvironmentErrorMsg = "no environment found for authenticated user"

	FindKafkaClusterNoClientErrorMsg          = "unable to obtain Kafka cluster information for cluster \"%s\": no client"
	ResolveResourceIDResourceNotFoundErrorMsg = "resource \"%s\" not found"

	AuthNetrcHandlerWriteToNetrcFileErrorMsg  = "unable to write to netrc file \"%s\""
	AuthNetrcHandlerResolvingFilepathErrorMsg = "unable to resolve netrc filepath at \"%s\""
	AuthNetrcHandlerGetNetrcCredentialsErrorMsg = "unable to get credentials from netrc file \"%s\""
	AuthNetrcHandlerCredentialsNotFoundErrorMsg = "login credentials not found in netrc file \"%s\""
	AuthNetrcHandlerCreateFileErrorMsg = "unable to create netrc file \"%s\""

	KafkaClusterCreateCloudRegionNotAvailableErrorMsg    = "\"%s\" is not an available region for \"%s\""
	KafkaClusterCreateCloudProviderNotAvailableErrorMsg = "\"%s\" is not an available cloud provider"
)
