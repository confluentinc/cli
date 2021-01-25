package errors

const (
	//admin commands
	DeletedUserMsg     = "Successfully deleted user %s."
	EmailInviteSentMsg = "An email invitation has been sent to %s"
	// api-key command
	DeletedAPIKeyMsg = "Deleted API key \"%s\".\n"
	StoredAPIKeyMsg  = "Stored API secret for API key \"%s\".\n"
	UseAPIKeyMsg     = "Set API Key \"%s\" as the active API key for \"%s\".\n"

	// auth commands
	LoggedInAsMsg              = "Logged in as \"%s\".\n"
	LoggedInUsingEnvMsg        = "Using environment \"%s\" (\"%s\").\n"
	LoggedOutMsg               = "You are now logged out."
	WroteCredentialsToNetrcMsg = "Wrote credentials to netrc file \"%s\"\n"
	KafkaClusterDeletedMsg     = "Deleted Kafka cluster \"%s\".\n"
	InvalidInputFormatMsg      = "\"%s\" is not of valid format for field \"%s\"."
	StopNonInteractiveMsg      = "(remove these credentials or use the `--prompt` flag to bypass non-interactive login)"
	FoundEnvCredMsg            = "Found credentials for user \"%s\" from environment variables \"%s\" and \"%s " +
		StopNonInteractiveMsg + ".\n"
	EnvLoginFailedMsg = "Non-interactive login with environment variables failed: %s\n"
	FoundNetrcCredMsg = "Found credentials for user \"%s\" from netrc file \"%s " +
		StopNonInteractiveMsg + ".\n"
	NetrcLoginFailedMsg = "Non-interactive login with netrc failed: %s\n"

	// confluent cluster command
	UnregisteredClusterMsg = "Successfully unregistered the cluster %s from the Cluster Registry.\n"

	// connector commands
	CreatedConnectorMsg = "Created connector %s %s\n"
	UpdatedConnectorMsg = "Updated connector %s\n"
	DeletedConnectorMsg = "Deleted connector \"%s\".\n"
	PausedConnectorMsg  = "Paused connector \"%s\".\n"
	ResumedConnectorMsg = "Resumed connector \"%s\".\n"

	// environment commands
	UsingEnvMsg   = "Now using \"%s\" as the default (active) environment.\n"
	DeletedEnvMsg = "Deleted environment \"%s\".\n"

	// feedback commands
	ThanksForFeedbackMsg = "Thanks for your feedback."

	// init command
	InitContextMsg = "Context \"%s\" initialized and set as current context.\n"

	// kafka cluster commands
	UseKafkaClusterMsg           = "Set Kafka cluster \"%s\" as the active cluster for environment \"%s\".\n"
	CopyBYOKPermissionsHeaderMsg = "Copy and append these permissions to the existing \"Statements\" array field in the key policy of your ARN to authorize access for Confluent:"

	// kafka topic commands
	StartingProducerMsg    = "Starting Kafka Producer. ^C or ^D to exit"
	StoppingConsumer       = "Stopping Consumer."
	StartingConsumerMsg    = "Starting Kafka Consumer. ^C or ^D to exit"
	CreatedTopicMsg        = "Created topic \"%s\".\n"
	DeletedTopicMsg        = "Deleted topic \"%s\".\n"
	UnknownTopicMsg        = "Unknown topic: \"%s\".\n"
	UpdateTopicConfigMsg   = "Updated the following configs for topic \"%s\":\n"
	InvalidMirrorActionMsg = "Invalid mirror action \"%s\" provided.\n"
	StoppedTopicMirrorMsg  = "Stopped mirroring for topic \"%s\".\n"

	// kafka link commands
	DryRunPrefix   = "[DRY RUN] "
	DeletedLinkMsg = "Deleted cluster link \"%s\".\n"
	CreatedLinkMsg = "Created cluster link \"%s\".\n"
	UpdatedLinkMsg = "Updated cluster link \"%s\".\n"

	// kafka acl commands
	DeletedACLsMsg      = "Deleted ACLs.\n"
	DeletedACLsCountMsg = "Deleted %d ACLs.\n"
	ACLsNotFoundMsg     = "ACL not found; ACL may have been misspelled or already deleted.\n"

	// ksql commands
	EndPointNotPopulatedMsg   = "Endpoint not yet populated. To obtain the endpoint, use `ccloud ksql app describe`."
	KsqlDBDeletedMsg          = "ksqlDB app \"%s\" has been deleted.\n"
	KsqlDBNotBackedByKafkaMsg = "The ksqlDB cluster \"%s\" is backed by \"%s\" which is not the current Kafka cluster \"%s\".\nTo switch to the correct cluster, use `ccloud kafka cluster use %s`.\n"

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
	RegisteredSchemaMsg                 = "Successfully registered schema with ID %v\n"
	DeletedAllSubjectVersionMsg         = "Successfully %s deleted all versions for subject \"%s\"\n"
	DeletedSubjectVersionMsg            = "Successfully %s deleted version \"%s\" for subject \"%s\".\n"
	UpdatedSubjectLevelCompatibilityMsg = "Successfully updated Subject Level compatibility to \"%s\" for subject \"%s\"\n"
	UpdatedSubjectLevelModeMsg          = "Successfully updated Subject level Mode to \"%s\" for subject \"%s\"\n"
	NoSubjectsMsg                       = "No subjects"
	SRCredsValidationFailedMsg          = "Failed to validate Schema Registry API key and secret."

	// secret commands
	UpdateSecretFileMsg = "Updated the encrypted secrets."

	// update command
	CheckingForUpdatesMsg = "Checking for updates..."
	UpToDateMsg           = "Already up to date."
	UpdateAutocompleteMsg = "Update your autocomplete scripts as instructed by `%s help completion`.\n"

	// cmd package
	TokenExpiredMsg        = "Your token has expired. You are now logged out."
	NotifyUpdateMsg        = "Updates are available for %s from (current: %s, latest: %s).\nTo view release notes and install them, please run:\n$ %s update\n\n"
	LocalCommandDevOnlyMsg = "The local commands are intended for a single-node development environment only,\n" +
		"NOT for production usage. https://docs.confluent.io/current/cli/index.html\n"
	AutoLoginMsg = "Successful auto log in with non-interactive credentials.\n"

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
	PromptToDownloadQuestionMsg = "Do you want to download and install this update?"
	InvalidChoiceMsg            = "%s is not a valid choice"

	// General
	UpdateSuccessMsg = "Updated the %s of %s \"%s\" to \"%s\".\n"

	// shell package
	AlreadyAuthenticatedMsg      = "You are already authenticated."
	CurrentlyNotAuthenticatedMsg = "You are currently not authenticated."
	ShellWelcomeMsg              = "Welcome to the %s shell! %s\n"
	ShellExitInstructionsMsg     = "Please press `Ctrl-D` or type `quit` to exit."
)
