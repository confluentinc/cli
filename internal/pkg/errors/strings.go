package errors

const (
	// api-key command
	APIKeySuccessfullyDeletedMsg = "API Key successfully deleted."

	// auth commands
	LoggedInAsMsg                = "Logged in as \"%s\".\n"
	LoggedInUsingEnvMsg          = "Using environment \"%s\" (\"%s\").\n"
	LoggedOutMsg                 = "You are now logged out."
	WrittenCredentialsToNetrcMsg = "Written credentials to netrc file \"%s\"\n"
	KafkaClusterDeletedMsg       = "The Kafka cluster \"%s\" has been deleted.\n"

	// confluent cluster command
	UnregisteredClusterMsg = "Successfully unregistered the cluster %s from the Cluster Registry.\n"

	// connector commands
	CreatedConnectorMsg = "Created connector %s %s\n"
	UpdatedConnectorMsg = "Updated connector %s\n"
	DeletedConnectorMsg = "Successfully deleted connector %s\n"
	PausedConnectorMsg  = "Successfully paused connector %s\n"
	ResumedConnectorMsg = "Successfully resumed connector %s\n"

	// environment commands
	UsingEnvMsg = "Now using \"%s\" as the default (active) environment.\n"

	// feedback commands
	ThanksForFeedbackMsg = "Thanks for your feedback."

	// kafka cluster commands
	ConfirmAuthorizedKeyMsg = "Please confirm you have authorized the key for these accounts: %s"

	// kafka topic commands
	StartingProducerMsg = "Starting Kafka Producer. ^C or ^D to exit"
	StoppingConsumer    = "Stopping Consumer."
	StartingConsumerMsg = "Starting Kafka Consumer. ^C or ^D to exit"

	// ksql commands
	EndPointNotPopulatedMsg   = "Endpoint not yet populated. To obtain the endpoint, use `ccloud ksql app describe`."
	KsqlDBDeletedMsg          = "ksqlDB app \"%s\" has been deleted.\n"
	KsqlDBNotBackedByKafkaMsg = "The ksqlDB cluster \"%s\" is not backed by \"%s\" which is not the current Kafka cluster \"%s\".\n"

	// local commands
	AvailableServicesMsg       = "Available Services:\n%s\n"
	UsingConfluentCurrentMsg   = "Using CONFLUENT_CURRENT: %s\n"
	AvailableConnectPluginsMsg = "Available Connect Plugins:\n%s\n"
	StartingServiceMsg         = "Starting %s\n"
	StoppingServiceMsg         = "Stopping %s\n"
	ServiceStatusMsg           = "%s is [%s]\n"
	DestroyDeletingMsg         = "Deleting: %s\n"

	// schema-registry commands
	UpdatedToLevelCompatibilityMsg      = "Successfully updated Top Level compatibility to \"%s\"\n"
	UpdatedTopLevelModeMsg              = "Successfully updated Top Level mode to \"%s\"\n"
	RegisteredSchemaMsg                 = "Successfully registered schema with ID %v"
	DeletedAllSubjectVersionMsg         = "Successfully %s deleted all versions for subject\n"
	DeletedSubjectVersionMsg            = "Successfully %s deleted version \"%s\" for subject\n"
	UpdatedSubjectLevelCompatibilityMsg = "Successfully updated Subject Level compatibility to \"%s\" for subject \"%s\"\n"
	UpdatedSubjectLevelModeMsg          = "Successfully updated Subject level Mode to \"%s\" for subject \"%s\"\n"
	NoSubjectsMsg                       = "No subjects"
	SRCredsValidationFailedMsg          = "Failed to validate Schema Registry API key and secret."

	// update command
	CheckingForUpdatesMsg = "Checking for updates..."
	UpToDateMsg           = "Already up to date."
	UpdateAutocompleteMsg = "Update your autocomplete scripts as instructed by `%s help completion`.\n"

	// cmd package
	TokenExpiredMsg        = "Your token has expired. You are now logged out."
	NotifyUpdateMsg        = "Updates are available for %s from (current: %s, latest: %s).\nTo view release notes and install them, please run:\n$ %s update\n\n"
	LocalCommandDevOnlyMsg = "The local commands are intended for a single-node development environment only,\n" +
		"NOT for production usage. https://docs.confluent.io/current/cli/index.html\n"

	// config package
	APIKeyMissingMsg     = "API key missing"
	KeyPairMismatchMsg   = "key of the dictionary does not match API key of the pair"
	APISecretMissingMsg  = "API secret missing"
	APIKeysMapAutofixMsg = "There are malformed API key secret pair entries in the dictionary for cluster \"%s\" under context \"%s\".\n" +
		"The issues are the following: %s.\n" +
		"Deleting the malformed entries.\n" +
		"You can re-add the API key secret pair with `ccloud api-key store --resource %s`\n"
	CurrentAPIKeyAutofixMsg = "Current API key \"%s\" of resource \"%s\" under context \"%s\" is not found.\n" +
		"Removing current API key setting for the resource.\n" +
		"You can re-add the API key with `ccloud api-key store --resource %s'` and then set current API key with `ccloud api-key use`.\n"

	// feedback package
	FeedbackNudgeMsg = "\nDid you know you can use the `ccloud feedback` command to send the team feedback?\n" +
		"Let us know if the CLI is meeting your needs, or what we can do to improve it.\n"

	// sso package
	NoBrowserSSOInstructionsMsg = "Navigate to the following link in your browser to authenticate:\n" +
		"%s\n" +
		"\n" +
		"After authenticating in your browser, paste the code here:\n"

	// update package
	PromptToDownloadDescriptionMsg = "New version of %s is available\n" +
		"Current Version: %s\n" +
		"Latest Version:  %s\n" +
		"%s\n\n\n"
	PromptToDownloadQuestionMsg = "Do you want to download and install this update? (y/n): "
	InvalidChoiceMsg            = "%s is not a valid choice\n"
)
