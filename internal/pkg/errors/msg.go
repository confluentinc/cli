package errors

var (
	ConfigUnableToLoadError = "Unable to load config: %s"
	ConfigUnspecifiedPlatformError = "Context \"%s\" has a corrupted platform. To fix, please remove the config file, and run `login` or `init`."
	ConfigUnspecifiedCredentialError = "Context \"%s\" has corrupted credentials. To fix, please remove the config file, and run `login` or `init`."

	APIKeyCommandResourceTypeNotImplementedErrorMsg = "Command not yet available for non -Kafka cluster resources."
)
