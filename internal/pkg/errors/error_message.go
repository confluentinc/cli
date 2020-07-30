package errors

/*
	Error message and suggestions message associated with them
*/

const (
	// format
	prefixFormat = "%s: %s"

	// api-key commands
	UnableToStoreAPIKeyErrorMsg       = "unable to store API key locally"
	NonKafkaNotImplementedErrorMsg    = "command not yet available for non-Kafka cluster resources"
	RefuseToOverrideSecretErrorMsg    = "refusing to overwrite existing secret for API Key \"%s\""
	RefuseToOverrideSecretSuggestions = "If you would like to override the existing secret stored for API key \"%s\", use `--force` flag."
	APIKeyUseFailedErrorMsg           = "unable to set active API key"

	// audit-log command
	EnsureCPSixPlusSuggestions        = "Ensure that you are running against MDS with CP 6.0+."
	UnableToAccessEndpointErrorMsg    = "unable to access endpoint"
	UnableToAccessEndpointSuggestions = EnsureCPSixPlusSuggestions

	// login command
	UnableToSaveUserAuthErrorMsg     = "unable to save user authentication"
	NoEnvironmentFoundErrorMsg       = "no environment found for authenticated user"
	NotUsernameAuthenticatedErrorMsg = "user not username authenticated has no access to ccloud client"

	// confluent cluster commands
	FetchClusterMetadataErrorMsg     = "unable to fetch cluster metadata: %s - %s"
	AccessClusterRegistryErrorMsg    = "unable to access Cluster Registry"
	AccessClusterRegistrySuggestions = EnsureCPSixPlusSuggestions
	MustSpecifyOneClusterIDErrorMsg  = "must specify at least one cluster ID"
	ProtocolNotSupportedErrorMsg     = "protocol %s is currently not supported"

	// completion command
	UnsupportedShellErrorMsg = "unsupported shell type \"%s\""

	// connect and connector-catalog commands
	EmptyConfigFileErrorMsg        = "connector config file \"%s\" is empty"
	MissingRequiredConfigsErrorMsg = "required configs \"name\" and \"connector.class\" missing from connector config file \"%s\""
	PluginNameNotPassedErrorMsg    = "plugin name must be passed"
	InvalidCloudErrorMsg           = "error defining plugin on given Kafka cluster"

	// environment command
	EnvNotFoundErrorMsg    = "environment \"%s\" not found"
	EnvNotFoundSuggestions = "List available environments with `ccloud environment list`."
	EnvSwitchErrorMsg      = "failed to switch environment: failed to save config"
	EnvRefreshErrorMsg     = "unable to save user auth while refreshing environment list"

	// iam acl & kafka acl commands
	UnableToPerformAclErrorMsg    = "unable to %s ACLs: %s"
	UnableToPerformAclSuggestions = "Ensure that you're running against MDS with CP 5.4+."
	MustSetAllowOrDenyErrorMsg    = "--allow or --deny must be set when adding or deleting an ACL"
	MustSetResourceTypeErrorMsg   = "exactly one resource type (%v) must be set"
	InvalidOperationValueErrorMsg = "invalid operation value: %s"
	ExactlyOneSetErrorMsg         = "exactly one of %v must be set"

	// iam role commands
	UnknownRoleErrorMsg    = "unknown role \"%s\""
	UnknownRoleSuggestions = "The available roles are: %s"

	// iam role-binding commands
	PrincipalFormatErrorMsg         = "incorrect principal format specified"
	PrincipalFormatSuggestions      = "Principal must be specified in this format: `<Principal Type>:<Principal Name>`."
	ResourceFormatErrorMsg          = "incorrect resource format specified"
	ResourceFormatSuggestions       = "Resource must be specified in this format: `<Resource Type>:<Resource Name>`."
	LookUpRoleErrorMsg              = "failed to lookup role \"%s\""
	LookUpRoleSuggestions           = "To check for valid roles, use `confluent role list`."
	InvalidResourceTypeErrorMsg     = "invalid resource type \"%s\""
	InvalidResourceTypeSuggestions  = "The available resource types are: %s"
	SpecifyKafkaIDErrorMsg          = "must also specify a --kafka-cluster-id to uniquely identify the scope"
	SpecifyCloudClusterErrorMsg     = "must specify cloud-cluster flag to indicate role binding scope"
	SpecifyEnvironmentErrorMsg      = "must specify environment flag to indicate role binding scope"
	BothClusterNameAndScopeErrorMsg = "cannot specify both cluster name and cluster scope"
	SpecifyClusterErrorMsg          = "must specify either cluster ID to indicate role binding scope or the cluster name"
	MoreThanOneNonKafkaErrorMsg     = "cannot specify more than one non-Kafka cluster ID for a scope"
	PrincipalOrRoleRequiredErrorMsg = "must specify either principal or role"
	HTTPStatusCodeErrorMsg          = "no error but received HTTP status code %d"
	HTTPStatusCodeSuggestions       = "Please file a support ticket with details."

	// init command
	CannotBeEmptyErrorMsg         = "%s cannot be empty"
	OnlyKafkaAuthErrorMsg         = "only `kafka-auth` is currently supported"
	UnknownCredentialTypeErrorMsg = "credential type %d unknown"

	// kafka cluster commands
	ListTopicSuggestions                 = "To list topics for the cluster \"%s\", use `ccloud kafka topic list --cluster %s`."
	FailedToRenderKeyPolicyErrorMsg      = "BYOK error: failed to render key policy"
	FailedToReadConfirmationErrorMsg     = "BYOK error: failed to read your confirmation"
	AuthorizeAccountsErrorMsg            = "BYOK error: please authorize the accounts (%s) for the key"
	CKUOnlyForDedicatedErrorMsg          = "specifying `--cku` flag is valid only for dedicated Kafka cluster creation"
	CKUMoreThanZeroErrorMsg              = "`--cku` value must be greater than 0"
	CloudRegionNotAvailableErrorMsg      = "\"%s\" is not an available region for \"%s\""
	CloudRegionNotAvailableSuggestions   = "To view a list of available regions for \"%s\", use `ccloud kafka region list --cloud %s`."
	CloudProviderNotAvailableErrorMsg    = "\"%s\" is not an available cloud provider"
	CloudProviderNotAvailableSuggestions = "To view a list of available cloud providers and regions, use `ccloud kafka region list`."
	TopicNotExistsErrorMsg               = "topic \"%s\" does not exist"
	TopicNotExistsSuggestions            = ListTopicSuggestions
	InvalidAvailableFlagErrorMsg         = "invalid value \"%s\" for `--availability` flag"
	InvalidAvailableFlagSuggestions      = "Allowed values for `--availability` flag are: %s, %s."
	InvalidTypeFlagErrorMsg              = "invalid value \"%s\" for `--type` flag"
	InvalidTypeFlagSuggestions           = "Allowed values for `--type` flag are: %s, %s, %s."
	NameOrCKUFlagErrorMsg                = "must either specify --name with non-empty value or --cku (for dedicated clusters) with positive integer"
	NonEmptyNameErrorMsg                 = "`--name` flag value must not be emtpy"

	// kafka topic commands
	FailedToProduceErrorMsg    = "failed to produce offset %d: %s\n"
	ConfigurationFormErrorMsg  = "configuration must be in the form of key=value"
	MissingKeyErrorMsg         = "missing key in message"
	UnknownValueFormatErrorMsg = "unknown value schema format"
	TopicExistsErrorMsg        = "topic \"%s\" already exists for Kafka cluster \"%s\""
	TopicExistsSuggestions     = ListTopicSuggestions

	// serialization/deserialization commands
	JsonSchemaInvalidErrorMsg    = "the json schema is invalid"
	JsonDocumentInvalidErrorMsg  = "the json document is invalid"
	ProtoSchemaInvalidErrorMsg   = "the protobuf schema is invalid"
	ProtoDocumentInvalidErrorMsg = "the protobuf document is invalid"

	// ksql commands
	NoServiceAccountErrorMsg = "no service account found for KSQL cluster \"%s\""

	// local commands
	NoServicesRunningErrorMsg = "no services running"
	TopNotAvailableErrorMsg   = "top command not available on platform: %s"
	InvalidConnectorErrorMsg  = "invalid connector: %s"
	FailedToStartErrorMsg     = "%s failed to start"
	FailedToStopErrorMsg      = "%s failed to stop"
	JavaRequirementErrorMsg   = "the Confluent CLI requires Java version 1.8 or 1.11.\n" +
		"See https://docs.confluent.io/current/installation/versions-interoperability.html\n" +
		"If you have multiple versions of Java installed, you may need to set JAVA_HOME to the version you want Confluent to use."
	NoLogFoundErrorMsg       = "no log found: to run %s, use \"confluent local services %s start\""
	MacVersionErrorMsg       = "macOS version >= %s is required (detected: %s)"
	JavaExecNotFondErrorMsg  = "could not find java executable, please install java or set JAVA_HOME"
	NothingToDestroyErrorMsg = "nothing to destroy"

	// prompt command
	ParseTimeOutErrorMsg      = "invalid value \"%s\" for `-t, --timeout` flag: unable to parse %s as duration or milliseconds"
	ParsePromptFormatErrorMsg = "error parsing prompt format string \"%s\""

	// schema-registry commands
	CompatibilityOrModeErrorMsg  = "must pass either `--compatibility` or `--mode` flag"
	BothSchemaAndSubjectErrorMsg = "cannot specify both schema ID and subject/version"
	SchemaOrSubjectErrorMsg      = "must specify either schema ID or subject/version"
	SchemaIntegerErrorMsg        = "invalid schema ID \"%s\""
	SchemaIntegerSuggestions     = "Schema ID must be an integer."

	// secret commands
	EnterInputTypeErrorMsg    = "enter %s"
	PipeInputTypeErrorMsg     = "pipe %s over stdin"
	SpecifyPassphraseErrorMsg = "specify `--passphrase -` if you intend to pipe your passphrase over stdin"
	PipePassphraseErrorMsg    = "pipe your passphrase over stdin"

	// update command
	UpdateClientFailurePrefix      = "update client failure"
	UpdateClientFailureSuggestions = "Please submit a support ticket.\n" +
		"In the meantime, see link for other ways to download the latest CLI version:\n" +
		"%s"
	ReadingYesFlagErrorMsg              = "error reading `--yes` flag as bool"
	CheckingForUpdateErrorMsg           = "error checking for updates"
	UpdateBinaryErrorMsg                = "error updating CLI binary"
	ObtainingReleaseNotesErrorMsg       = "error obtaining release notes: %s"
	ReleaseNotesVersionCheckErrorMsg    = "unable to perform release notes and binary version check: %s"
	ReleaseNotesVersionMismatchErrorMsg = "binary version (v%s) and latest release notes version (v%s) mismatch"

	// auth package
	NoReaderForCustomCertErrorMsg    = "no reader specified for reading custom certificates"
	ReadCertErrorMsg                 = "failed to read certificate"
	NoCertsAppendedErrorMsg          = "no certs appended, using system certs only"
	WriteToNetrcFileErrorMsg         = "unable to write to netrc file \"%s\""
	ResolvingNetrcFilepathErrorMsg   = "unable to resolve netrc filepath at \"%s\""
	GetNetrcCredentialsErrorMsg      = "unable to get credentials from netrc file \"%s\""
	NetrcCredentialsNotFoundErrorMsg = "login credentials not found in netrc file \"%s\""
	CreateNetrcFileErrorMsg          = "unable to create netrc file \"%s\""

	// cmd package
	FindKafkaNoClientErrorMsg = "unable to obtain Kafka cluster information for cluster \"%s\": no client"
	InvalidAPIKeyErrorMsg     = "invalid API key \"%s\" for resource \"%s\""
	InvalidAPIKeySuggestions  = "To list API key that belongs to resource \"%s\", use `ccloud api-key list --resource %s`.\n" +
		"To create new API key for resource \"%s\", use `ccloud api-key create --resource %s`."
	SRNotEnabledErrorMsg    = "Schema Registry not enabled"
	SRNotEnabledSuggestions = "Schema Registry must be enabled for the environment in order to run the command.\n" +
		"You can enable Schema Registry for this environment with `ccloud schema-registry cluster enable`."
	EnvironmentNotFoundErrorMsg = "environment \"%s\" not found in context \"%s\""
	MalformedJWTNoExprErrorMsg  = "malformed JWT claims: no expiration"

	// config package
	CorruptedConfigErrorPrefix = "corrupted CLI config"
	CorruptedConfigSuggestions = "Your CLI config file \"%s\" is corrupted.\n" +
		"Remove config file, and run `%s login` or `%s init`.\n" +
		"Unfortunately, your active CLI state will be lost as a result.\n" +
		"Please file a support ticket with details about your config file to help us address this issue.\n" +
		"Please rerun the command with the verbosity flag `-vvvv` and attach the output with the support ticket."
	UnableToCreateConfigErrorMsg       = "unable to create config"
	UnableToReadConfigErrorMsg         = "unable to read config file \"%s\""
	ConfigNotUpToDateErrorMsg          = "config version v%s not up to date with the latest version v%s"
	InvalidConfigVersionErrorMsg       = "invalid config version v%s"
	ParseConfigErrorMsg                = "unable to parse config file \"%s\""
	NoNameContextErrorMsg              = "one of the existing contexts has no name"
	MissingKafkaClusterContextErrorMsg = "context \"%s\" missing KafkaClusterContext"
	MarshalConfigErrorMsg              = "unable to marshal config"
	CreateConfigDirectoryErrorMsg      = "unable to create config directory: %s"
	CreateConfigFileErrorMsg           = "unable to write config to file: %s"
	CurrentContextNotExistErrorMsg     = "the current context \"%s\" does not exist"
	ContextNotExistErrorMsg            = "context \"%s\" does not exist"
	ContextNameExistsErrorMsg          = "cannot create context \"%s\": context with this name already exists"
	CredentialNotFoundErrorMsg         = "credential \"%s\" not found"
	PlatformNotFoundErrorMsg           = "platform \"%s\" not found"
	NoNameCredentialErrorMsg           = "credential must have a name"
	NoNamePlatformErrorMsg             = "platform must have a name"
	ResolvingConfigPathErrorMsg        = "error resolving the config filepath at \"%s\" has occurred"
	ResolvingConfigPathSuggestions     = "Try moving the config file to a different location."
	UnspecifiedPlatformErrorMsg        = "context \"%s\" has corrupted platform"
	UnspecifiedCredentialErrorMsg      = "context \"%s\" has corrupted credentials"
	ContextStateMismatchErrorMsg       = "context state mismatch for context \"%s\""
	ContextStateNotMappedErrorMsg      = "context state mapping error for context \"%s\""
	ClearInvalidAPIFailErrorMsg        = "unable to clear invalid API key pairs"
	DeleteUserAuthErrorMsg             = "unable to delete user auth"
	ResetInvalidAPIKeyErrorMsg         = "unable to reset invalid active API key"
	NoIDClusterErrorMsg                = "Kafka cluster under context \"%s\" has no ID"

	// local package
	ConfluentHomeNotFoundErrorMsg         = "could not find %s in CONFLUENT_HOME"
	SetConfluentHomeErrorMsg              = "set environment variable CONFLUENT_HOME"
	KafkaScriptFormatNotSupportedErrorMsg = "format %s is not supported in this version"
	KafkaScriptInvalidFormatErrorMsg      = "invalid format: %s"

	// secret package
	EncryptPlainTextErrorMsg           = "failed to encrypt the plain text"
	DecryptCypherErrorMsg              = "failed to decrypt the cipher"
	DataCorruptedErrorMsg              = "failed to decrypt the cipher: data is corrupted"
	ConfigNotInJAASErrorMsg            = "the configuration \"%s\" not present in JAAS configuration"
	OperationNotSupportedErrorMsg      = "the operation \"%s\" is not supported"
	InvalidJAASConfigErrorMsg          = "invalid JAAS configuration: %s"
	ExpectedConfigNameErrorMsg         = "expected a configuration name but received \"%s\""
	LoginModuleControlFlagErrorMsg     = "login module control flag is not specified"
	ConvertPropertiesToJAASErrorMsg    = "failed to convert the properties to a JAAS configuration"
	ValueNotSpecifiedForKeyErrorMsg    = "value is not specified for the key \"%s\""
	MissSemicolonErrorMsg              = "configuration not terminated with a ';'"
	EmptyPassphraseErrorMsg            = "master key passphrase cannot be empty"
	AlreadyGeneratedErrorMsg           = "master key is already generated"
	AlreadyGeneratedSuggestions        = "You can rotate the key with `confluent secret file rotate`."
	InvalidConfigFilePathErrorMsg      = "invalid config file path \"%s\""
	InvalidSecretFilePathErrorMsg      = "invalid secrets file path \"%s\""
	UnwrapDataKeyErrorMsg              = "failed to unwrap the data key: invalid master key or corrupted data key"
	DecryptConfigErrorMsg              = "failed to decrypt config \"%s\": corrupted data"
	SecretConfigFileMissingKeyErrorMsg = "missing config key \"%s\" in secret config file"
	IncorrectPassphraseErrorMsg        = "authentication failure: incorrect master key passphrase"
	SamePassphraseErrorMsg             = "new master key passphrase may not be the same as the previous passphrase"
	EmptyNewConfigListErrorMsg         = "add failed: empty list of new configs"
	EmptyUpdateConfigListErrorMsg      = "update failed: empty list of update configs"
	ConfigKeyNotEncryptedErrorMsg      = "configuration key \"%s\" is not encrypted"
	FileTypeNotSupportedErrorMsg       = "file type \"%s\" currently not supported"
	ConfigKeyNotInJSONErrorMsg         = "configuration key \"%s\" not present in JSON configuration file"
	MasterKeyNotExportedErrorMsg       = "master key is not exported in `%s` environment variable"
	MasterKeyNotExportedSuggestions    = "Set the environment variable `%s` to the master key and execute this command again."
	ConfigKeyNotPresentErrorMsg        = "configuration key \"%s\" not present in the configuration file"
	InvalidJSONFileFormatErrorMsg      = "invalid json file format"
	InvalidFilePathErrorMsg            = "invalid file path \"%s\""
	UnsupportedFileFormatErrorMsg      = "unsupported file format for file \"%s\""

	// sso package
	StartHTTPServerErrorMsg            = "unable to start HTTP server"
	AuthServerRunningErrorMsg          = "CLI HTTP auth server encountered error while running: %s\n"
	AuthServerShutdownErrorMsg         = "CLI HTTP auth server encountered error while shutting down: %s\n"
	BrowserAuthTimedOutErrorMsg        = "timed out while waiting for browser authentication to occur"
	BrowserAuthTimedOutSuggestions     = "Try logging in again."
	LoginFailedCallbackURLErrorMsg     = "authentication callback URL either did not contain a state parameter in query string, or the state parameter was invalid; login will fail"
	LoginFailedQueryStringErrorMsg     = "authentication callback URL did not contain code parameter in query string; login will fail"
	ReadCallbackPageTemplateErrorMsg   = "could not read callback page template"
	PastedInputErrorMsg                = "pasted input had invalid format"
	LoginFailedStateParamErrorMsg      = "authentication code either did not contain a state parameter or the state parameter was invalid; login will fail"
	OpenWebBrowserErrorMsg             = "unable to open web browser for authorization"
	GenerateRandomSSOProviderErrorMsg  = "unable to generate random bytes for SSO provider state"
	GenerateRandomCodeVerifierErrorMsg = "unable to generate random bytes for code verifier"
	ComputeHashErrorMsg                = "unable to compute hash for code challenge"
	MissingIDTokenFieldErrorMsg        = "oauth token response body did not contain id_token field"
	ConstructOAuthRequestErrorMsg      = "failed to construct oauth token request"
	UnmarshalOAuthTokenErrorMsg        = "failed to unmarshal response body in oauth token request"

	// update package
	ParseVersionErrorMsg            = "unable to parse %s version %s"
	TouchLastCheckFileErrorMsg      = "unable to touch last check file"
	GetTempDirErrorMsg              = "unable to get temp dir for %s"
	DownloadVersionErrorMsg         = "unable to download %s version %s to %s"
	MoveFileErrorMsg                = "unable to move %s to %s"
	MoveRestoreErrorMsg             = "unable to move (restore) %s to %s"
	CopyErrorMsg                    = "unable to copy %s to %s"
	ChmodErrorMsg                   = "unable to chmod 0755 %s"
	SepNonEmptyErrorMsg             = "sep must be a non-empty string"
	NoVersionsErrorMsg              = "no versions found"
	GetBinaryVersionsErrorMsg       = "unable to get available binary versions"
	GetReleaseNotesVersionsErrorMsg = "unable to get available release notes versions"
	UnexpectedS3ResponseErrorMsg    = "received unexpected response from S3: %s"
	MissingRequiredParamErrorMsg    = "missing required parameter: %s"
	ListingS3BucketErrorMsg         = "error listing s3 bucket"
	FindingCredsErrorMsg            = "error while finding credentials"
	EmptyAccessKeyIDErrorMsg        = "access key id is empty for %s"
	AWSCredsExpiredErrorMsg         = "AWS credentials in profile %s are expired"
	FindAWSCredsErrorMsg            = "failed to find AWS credentials in profiles: %s"

	// Flag Errors
	ProhibitedFlagCombinationErrorMsg = "cannot use `--%s` and `--%s` flags at the same time"
	InvalidFlagValueErrorMsg          = "invalid value \"%s\" for flag `--%s`"
	InvalidFlagValueSuggestions       = "The possible values for flag `%s` are: %s."

	// catcher
	CCloudBackendErrorPrefix           = "CCloud backend error"
	UnexpectedBackendOutputPrefix      = "unexpected CCloud backend output"
	UnexpectedBackendOutputSuggestions = "Please submit a support ticket."
	BackendUnmarshallingErrorMsg       = "protobuf unmarshalling error"
	ResourceNotFoundErrorMsg           = "resource \"%s\" not found"
	ResourceNotFoundSuggestions        = "Check that the resource \"%s\" exists.\n" +
		"To list Kafka clusters, use `ccloud kafka cluster list`.\n" +
		"To check schema-registry cluster info, use `ccloud schema-registry cluster describe`.\n" +
		"To list KSQL clusters, use `ccloud ksql app list`."
	KafkaNotFoundErrorMsg             = "Kafka cluster \"%s\" not found"
	KafkaNotFoundSuggestions          = "To list Kafka clusters, use `ccloud kafka cluster list`."
	KSQLNotFoundSuggestions           = "To list KSQL clusters, use `ccloud ksql app list`."
	SRNotFoundSuggestions             = "Check the schema-registry cluster ID with `ccloud schema-registry cluster describe`."
	KafkaNotReadyErrorMsg             = "Kafka cluster \"%s\" not ready"
	KafkaNotReadySuggestions          = "It may take up to 5 minutes for a recently created Kafka cluster to be ready."
	NoKafkaSelectedErrorMsg           = "no Kafka cluster selected"
	NoKafkaSelectedSuggestions        = "You must pass `--cluster` flag with the command or set an active kafka in your context with `ccloud kafka cluster use`."
	UnableToConnectToKafkaErrorMsg    = "unable to connect to Kafka cluster"
	UnableToConnectToKafkaSuggestions = "For recently created Kafka clusters and API keys, it may take a few minutes before the resources are ready.\n" +
		"Otherwise, verify that for Kafka cluster \"%s\" the active API key \"%s\" used is the right one.\n" +
		"Also verify that the correct API secret is stored for the API key.\n" +
		"If the API secret is incorrect, override with `ccloud api-key store %s --resource %s --force`."
	NoAPISecretStoredErrorMsg    = "no API secret for API key \"%s\" of resource \"%s\" stored in local CLI state"
	NoAPISecretStoredSuggestions = "Store the API secret with `ccloud api-key store %s --resource %s`."

	// Special error handling
	avoidTimeoutWithCLINameSuggestion = "To avoid session timeouts, you can save credentials to netrc file with `%s login --save`."
	ccloudAvoidTimeoutSuggestion      = "To avoid session timeouts, you can save credentials to netrc file with `ccloud login --save`."
	avoidTimeoutGeneralSuggestion     = "To avoid session timeouts, you can save credentials to netrc file by logging in with `--save` flag."
	NotLoggedInErrorMsg               = "not logged in"
	NotLoggedInSuggestions            = "You must be logged in to run this command.\n" +
		avoidTimeoutWithCLINameSuggestion
	SRNotAuthenticatedErrorMsg    = "not logged in, and no Schema Registry endpoint specified"
	SRNotAuthenticatedSuggestions = "You must specify the endpoint for a Schema Registry cluster (--sr-endpoint) or be logged in using `ccloud login` to run this command.\n" +
		avoidTimeoutWithCLINameSuggestion
	CorruptedTokenErrorMsg    = "corrupted auth token"
	CorruptedTokenSuggestions = "Please log in again.\n" +
		avoidTimeoutGeneralSuggestion
	ExpiredTokenErrorMsg    = "expired token"
	ExpiredTokenSuggestions = "Your session has timed out, you need to log in again.\n" +
		avoidTimeoutGeneralSuggestion
	InvalidEmailErrorMsg    = "user \"%s\" not found"
	InvalidEmailSuggestions = "Check the email credential.\n" +
		"If the email is correct, check that you have successfully verified your email.\n" +
		"If the problem persists, please submit a support ticket.\n" +
		ccloudAvoidTimeoutSuggestion
	InvalidLoginErrorMsg          = "incorrect email or password"
	CCloudInvalidLoginSuggestions = ccloudAvoidTimeoutSuggestion
	NoAPIKeySelectedErrorMsg      = "no API key selected for resource \"%s\""
	NoAPIKeySelectedSuggestions   = "Select an API key for resource \"%s\" with `ccloud api-key use <API_KEY> --resource %s`.\n" +
		"To do so, you must have either already created or stored an API key for the resource.\n" +
		"To create an API key, use `ccloud api-key create --resource %s`.\n" +
		"To store an existing API key, use `ccloud api-key store --resource %s`."

	// Special error types
	GenericOpenAPIErrorMsg = "Metadata Service backend error: %s: %s"
)
