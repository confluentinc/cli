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
	LoggedInAsMsgWithOrg       = "Logged in as \"%s\" for organization \"%s\" (\"%s\").\n"
	LoggedInUsingEnvMsg        = "Using environment \"%s\" (\"%s\").\n"
	LoggedOutMsg               = "You are now logged out."
	WroteCredentialsToNetrcMsg = "Wrote credentials to netrc file \"%s\"\n"
	RemoveNetrcCredentialsMsg  = "Removed credentials for user \"%s\" from netrc file \"%s\"\n"
	KafkaClusterDeletedMsg     = "Deleted Kafka cluster \"%s\".\n"
	StopNonInteractiveMsg      = "(remove these credentials or use the `--prompt` flag to bypass non-interactive login)"
	FoundEnvCredMsg            = "Found credentials for user \"%s\" from environment variables \"%s\" and \"%s\" " +
		StopNonInteractiveMsg + ".\n"
	FoundNetrcCredMsg = "Found credentials for user \"%s\" from netrc file \"%s\" " +
		StopNonInteractiveMsg + ".\n"
	FoundOrganizationIdMsg = "Found default organization id for user \"%s\" from environment variable \"%s\".\n"

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

	// kafka cluster commands
	UseKafkaClusterMsg              = "Set Kafka cluster \"%s\" as the active cluster for environment \"%s\".\n"
	CopyBYOKAWSPermissionsHeaderMsg = "Copy and append these permissions to the existing \"Statements\" array field in the key policy of your ARN to authorize access for Confluent:"

	// kafka consumer-group commands
	RestProxyNotAvailable = "Operation not supported: REST proxy is not available.\n"

	// kafka topic commands
	StartingProducerMsg  = "Starting Kafka Producer. Use Ctrl-C or Ctrl-D to exit."
	StoppingConsumer     = "Stopping Consumer."
	StartingConsumerMsg  = "Starting Kafka Consumer. Use Ctrl-C to exit."
	CreatedTopicMsg      = "Created topic \"%s\".\n"
	DeletedTopicMsg      = "Deleted topic \"%s\".\n"
	UpdateTopicConfigMsg = "Updated the following configs for topic \"%s\":\n"

	// kafka link commands
	DeletedLinkMsg = "Deleted cluster link \"%s\".\n"
	CreatedLinkMsg = "Created cluster link \"%s\".\n"
	UpdatedLinkMsg = "Updated cluster link \"%s\".\n"

	// kafka mirror commands
	RestProxyNotAvailableMsg = "Kafka REST is not enabled: the operation is only support with kafka rest proxy."
	CreatedMirrorMsg         = "Created mirror topic \"%s\".\n"

	// kafka acl commands
	DeletedServiceAccountMsg = "Deleted service account \"%s\".\n"
	DeletedACLsMsg           = "Deleted ACLs.\n"
	DeletedACLsCountMsg      = "Deleted %d ACLs.\n"
	ACLsNotFoundMsg          = "ACL not found; ACL may have been misspelled or already deleted.\n"

	// ksql commands
	EndPointNotPopulatedMsg   = "Endpoint not yet populated. To obtain the endpoint, use `confluent ksql app describe`."
	KsqlDBDeletedMsg          = "ksqlDB app \"%s\" has been deleted.\n"
	KsqlDBNotBackedByKafkaMsg = "The ksqlDB cluster \"%s\" is backed by \"%s\" which is not the current Kafka cluster \"%s\".\nTo switch to the correct cluster, use `confluent kafka cluster use %s`.\n"

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
	ExporterActionMsg                   = "%s schema exporter \"%s\".\n"
	NoSubjectsMsg                       = "No subjects."
	NoExporterMsg                       = "No exporters."
	SRCredsValidationFailedMsg          = "Failed to validate Schema Registry API key and secret."

	// secret commands
	UpdateSecretFileMsg = "Updated the encrypted secrets."

	// update command
	CheckingForUpdatesMsg   = "Checking for updates..."
	UpToDateMsg             = "Already up to date."
	MajorVersionUpdateMsg   = "The only available update is a major version update. Use `%s update --major` to accept the update.\n"
	NoMajorVersionUpdateMsg = "No major version updates are available.\n"
	UpdateAutocompleteMsg   = "Update your autocomplete scripts as instructed by `%s help completion`.\n"

	// cmd package
	TokenExpiredMsg        = "Your token has expired. You are now logged out."
	NotifyMajorUpdateMsg   = "A major version update is available for %s from (current: %s, latest: %s).\nTo view release notes and install the update, please run `%s update --major`.\n\n"
	NotifyMinorUpdateMsg   = "A minor version update is available for %s from (current: %s, latest: %s).\nTo view release notes and install the update, please run `%s update`.\n\n"
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
		"You can re-add the API key secret pair with `confluent api-key store --resource %s`\n"
	CurrentAPIKeyAutofixMsg = "Current API key \"%s\" of resource \"%s\" under context \"%s\" is not found.\n" +
		"Removing current API key setting for the resource.\n" +
		"You can re-add the API key with `confluent api-key store --resource %s'` and then set current API key with `confluent api-key use`.\n"

	// feedback package
	FeedbackNudgeMsg = "\nDid you know you can use the `confluent feedback` command to send the team feedback?\n" +
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
	InvalidChoiceMsg = "%s is not a valid choice"

	// General
	UpdateSuccessMsg = "Updated the %s of %s \"%s\" to \"%s\".\n"
)
