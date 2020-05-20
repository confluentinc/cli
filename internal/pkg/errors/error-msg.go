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
	ProhibitedFlagCombinationMsg = "cannot use `--%s` and `--%s` flags at the same time"

	UserNotLoggedInErrorMsg =  "You must log in to run that command."


	APIKeyNotValidForResourceErrorMsg = "invalid API key \"%s\" for resource \"%s\""
	FlagUseProhibitedCombinationErrorMsg = "flag \"%s\" cannot be use with flag \"%s\""

	// Direction messages
	CorruptedConfigErrorDirectionMsg = "Please remove the config file, and run `ccloud login` or `ccloud init`."

	//// ErrMsg and Directions pair
	//UserNotLoggedInErrorPair = ErrorMsgAndDirectionMsgPair{
	//	Msg:        "login required",
	//	Directions: "Please login to use the command.",
	//}
	//APIKeyNotSelectedForResourceErrorPair = ErrorMsgAndDirectionMsgPair{
	//	Msg:        "no API key selected for resource \"%s\"",
	//	Directions: "Please select an API key for the resource with `ccloud api-key use`.",
	//}
	//APISecretNotStoredForAPIKeyErrorPair = ErrorMsgAndDirectionMsgPair{
	//	Msg:        "no API key secret stored for resource API key \"%s\", resource \"%s\"",
	//	Directions: "Please add API secret for the API key using `ccloud api-key store`.",
	//}

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
	KafkaClusterCommandCKUFlagValueErrorMsg = ""
)
