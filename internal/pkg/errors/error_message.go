package errors

// Error message and suggestions message associated with them

const (
	// Flag Errors
	ProhibitedFlagCombinationErrorMsg = "cannot use `--%s` and `--%s` flags at the same time"
	InvalidOutputFormatFlagErrorMsg = "invalid `format` flag value \"%s\""

	// API key commands
	UnableToStoreAPIKeyErrorMsg       = "unable to store API key locally"
	NonKafkaNotImplementedErrorMsg    = "command not yet available for non-Kafka cluster resources"
	RefuseToOverrideSecretErrorMsg    = "refusing to overwrite existing secret for API Key \"%s\""
	RefuseToOverrideSecretSuggestions = "If you would like to override the existing secret stored for API key \"%s\", please use `--force` flag."
	APIKeyUseFailedErrorMsg           = "unable to set active API key"

	// Login
	UnableToSaveUserAuthErrorMsg = "unable to save user authentication"
	NoEnvironmentFoundErrorMsg   = "no environment found for authenticated user"

	//
	FindKafkaNoClientErrorMsg         = "unable to obtain Kafka cluster information for cluster \"%s\": no client"
	CCloudResourceNotFoundErrorMsg    = "resource \"%s\" not found"
	CCloudResourceNotFoundSuggestions = "Please check the ID string \"%s\" for typos, and verify that the resource you are looking for exists."

	// netrc handler
	WriteToNetrcFileErrorMsg    = "unable to write to netrc file \"%s\""
	ResolvingFilepathErrorMsg   = "unable to resolve netrc filepath at \"%s\""
	GetNetrcCredentialsErrorMsg = "unable to get credentials from netrc file \"%s\""
	CredentialsNotFoundErrorMsg = "login credentials not found in netrc file \"%s\""
	CreateFileErrorMsg          = "unable to create netrc file \"%s\""

	// ccloud kafka commands
	CloudRegionNotAvailableErrorMsg      = "\"%s\" is not an available region for \"%s\""
	CloudRegionNotAvailableSuggestions   = "You can view a list of available regions for \"%s\" with `ccloud kafka region list --cloud %s` command."
	CloudProviderNotAvailableErrorMsg    = "\"%s\" is not an available cloud provider"
	CloudProviderNotAvailableSuggestions = "You can view a list of available cloud providers and regions with the `ccloud kafka region list` command."
)

const (
	ConfigUnableToLoadError          = "Unable to load config: %s"
	ConfigUnspecifiedPlatformError   = "Context \"%s\" has a corrupted platform. To fix, please remove the config file, and run `login` or `init`."
	ConfigUnspecifiedCredentialError = "Context \"%s\" has corrupted credentials. To fix, please remove the config file, and run `login` or `init`."
	UserNotLoggedInErrMsg            = "You must log in to run that command."
	CorruptedAuthTokenErrorMsg       = "Your auth token has been corrupted. Please login again."
	NotLoggedInInternalErrorMsg      = "not logged in"
)
