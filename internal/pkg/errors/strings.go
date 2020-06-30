package errors

const (
	// auth commands
    LoggedInAsMsg = "Logged in as %s."
    LoginUsingEnvMsg    = "Using environment %s (%s)."
    LoggedOutMsg = "You are now logged out."

	WrittenCredentialsToNetrcMsg = "Written credentials to netrc file \"%s\""
	KafkaClusterDeletedMsg       = "The Kafka cluster \"%s\" has been deleted."

	// connector commands
	CreatedConnectorMsg = "Created connector %s %s"
	UpdatedConnectorMsg = "Updated connector %s"
	DeletedConnectorMsg = "Successfully deleted connector %s"
	PausedConnectorMsg  = "Successfully paused connector %s"
	ResumedConnectorMsg = "Successfully resumed connector %s"

	// environment commands
	UsingEnvMsg = "Now using \"%s\" as the default (active) environment."

	// feedback command
	ThanksForFeedbackMsg = "Thanks for your feedback."

	// kafka cluster command
	ConfirmAuthorizedKeyMsg = "Please confirm you have authorized the key for these accounts: %s"

	// kafka topic command
	StartingProducerMsg = "Starting Kafka Producer. ^C or ^D to exit"
	StoppingConsumer = "Stopping Consumer."
	StartingConsumerMsg = "Starting Kafka Consumer. ^C or ^D to exit"

	// ksql command
	EndPointNotPopulatedMsg = "Endpoint not yet populated. To obtain the endpoint please use `ccloud ksql app describe`."
	KSQLDeletedMsg = "KSQL app \"%s\" has been deleted."
	KSQLNotBackedByKafkaMsg = "The KSQL cluster \"%s\" is not backed by \"%s\" which is not the current Kafka cluster \"%s\"."

	// schema-registry commands
	UpdatedToLevelCompatibilityMsg = "Successfully updated Top Level compatibility to \"%s\""
	UpdatedTopLevelModeMsg         = "Successfully updated Top Level mode to \"%s\""
	RegisteredSchemaMsg            = "Successfully registered schema with ID %v"
	DeletedAllSubjectVersionMsg    = "Successfully deleted all versions for subject"
	DeletedSubjectVersionMsg       = "Successfully deleted version \"%s\" for subject"
	UpdatedSubjectLevelCompatibilityMsg = "Successfully updated Subject Level compatibility to \"%s\" for subject \"%s\""
	UpdatedSubjectLevelModeMsg = "Successfully updated Subject level Mode to \"%s\" for subject \"%s\""
	NoSubjectsMsg = "No subjects"

	// secret commands
	SaveTheMasterKeyMsg = "Save the master key. It cannot be retrieved later."

	// update command
	CheckingForUpdatesMsg = "Checking for updates..."
	UpToDateMsg           = "Already up to date."
	UpdateAutocompleteMsg = "Update your autocomplete scripts as instructed by `%s help completion`."

	// Packages

	// cmd package
	TokenExpiredMsg = "Your token has expired. You are now logged out."
	NotifyUpdateMsg = "Updates are available for %s from (current: %s, latest: %s).\nTo view release notes and install them, please run:\n$ %s update\n"
	LocalCommandDevOnlyMsg = "The local commands are intended for a single-node development environment only,\n" +
		"NOT for production usage. https://docs.confluent.io/current/cli/index.html"


	// config package
	APIKeyMissingMsg = "API key missing"
	KeyPairMismatchMsg = "key of the dictionary does not match API key of the pair"
	APISecretMissingMsg = "API secret missing"
	APIKeysMapAutofixMsg = "There are malformed API key secret pair entries in the dictionary for cluster \"%s\" under context \"%s\".\n"+
		"The issues are the following: %s.\n"+
		"Deleting the malformed entries.\n"+
		"You can re-add the API key secret pair with `ccloud api-key store --resource %s`\n"
	CurrentAPIKeyAutofixMsg = "Current API key \"%s'\" of resource \"%s\" under context \"%s\" is not found.\n"+
		"Removing current API key setting for the resource.\n"+
		"You can re-add the API key with 'ccloud api-key store --resource %s' and set current API key with 'ccloud api-key use'.\n"


	// feedback package
	FeedbackNudgeMsg = "\nDid you know you can use the \"ccloud feedback\" command to send the team feedback?\nLet us know if the ccloud CLI is meeting your needs, or what we can do to improve it."


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
