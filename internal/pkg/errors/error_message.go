package errors

// Error message and suggestions message associated with them

const (
	// Flag Errors
	ProhibitedFlagCombinationErrorMsg = "cannot use \"--%s\" and \"--%s\" flags at the same time"
	InvalidFlagValueErrorMsg        = "invalid value \"%s\" for flag \"--%s\""
	InvalidFlagValueSuggestions     = "The possible values for flag \"%s\" are: %s."

	// API key commands
	UnableToStoreAPIKeyErrorMsg       = "unable to store API key locally"
	NonKafkaNotImplementedErrorMsg    = "command not yet available for non-Kafka cluster resources"
	RefuseToOverrideSecretErrorMsg    = "refusing to overwrite existing secret for API Key \"%s\""
	RefuseToOverrideSecretSuggestions = "If you would like to override the existing secret stored for API key \"%s\", please use `--force` flag."
	APIKeyUseFailedErrorMsg           = "unable to set active API key"

	// Login
	UnableToSaveUserAuthErrorMsg = "unable to save user authentication"
	NoEnvironmentFoundErrorMsg   = "no environment found for authenticated user"

	// Confluent Cluster commands
	FetchClusterMetadataErrorMsg     = "unable to fetch cluster metadata: %s - %s"
	AccessClusterRegistryErrorMsg    = "unable to access Cluster Registry"
	AccessClusterRegistrySuggestions = "Ensure that you're running against MDS with CP 6.0+."

	// ccloud config

	// ccloud connect and connector-catalog
	EmptyConfigFileErrorMsg        = "connector config file \"%s\" is empty"
	MissingRequiredConfigsErrorMsg = "required configs \"name\" and \"connector.class\" missing from connector config file \"%s\""
	PluginNameNotPassedErrorMsg    = "plugin name must be passed"
	InvalidCloudErrorMsg           = "error defining plugin on given kafka cluster"

	// environment command
	EnvNotFoundErrorMsg    = "environment \"%s\" not found"
	EnvNotFoundSuggestions = "List available environments with `ccloud environment list`."
	EnvSwitchErrorMsg      = "failed to switch environment: failed to save config"
	EnvRefreshErrorMsg      = "unable to save user auth while refreshing environment list"

	// iam acl & kafka acl
	UnableToPerformAclErrorMsg = "unable to %s ACLs (%s)"
	UnableToPerformAclSuggestions = "Ensure that you're running against MDS with CP 5.4+."
	MustSetAllowOrDenyErrorMsg = "--allow or --deny must be set when adding or deleting an ACL"
	MustSetResourceTypeErrorMsg = "exactly one resource type (%v) must be set"
	InvalidOperationValueErrorMsg = "invalid operation value: %s"

	// iam role
	UnknownRoleErrorMsg = "unknown role \"%s\""
	UnknownRoleSuggestions = "The available roles are: %s"

	// iam role-binding
	PrincipalFormatErrorMsg = "incorrect principal format specified"
	PrincipalFormatSuggestions = "Principal must be specified in this format: <Principal Type>:<Principal Name>."
	ResourceFormatErrorMsg = "incorrect resource format specified"
	ResourceFormatSuggestions = "Resource must be specified in this format: <Resource Type>:<Resource Name>."
	LookUpRoleErrorMsg     = "failed to lookup role \"%s\""
	LookUpRoleSuggestions  = "Check that the role name is valid with `confluent role list`"
	InvalidResourceTypeErrorMsg = "invalid resource type \"%s\""
	InvalidResourceTypeSuggestions = "The available resource types are: %s"
	SpecifyKafkaIDErrorMsg = "must also specify a --kafka-cluster-id to uniquely identify the scope"
	SpecifyClusterIDErrorMsg = "must specify at least one cluster ID flag to indicate role binding scope"
	MoreThanOneNonKafkaErrorMsg = "Cannot specify more than one non-Kafka cluster ID for a scope"
	PrincipalOrRoleRequiredErrorMsg = "must specify either principal or role"
	HTTPStatusCodeErrorMsg = "no error but received HTTP status code %d"
	HTTPStatusCodeSuggestions = "Please file a support ticket with details."

	// init command
	CannotBeEmptyErrorMsg = "%s cannot be empty"
	OnlyKafkaAuthErrorMsg = "only kafka-auth is currently supported"
	UnknownCredentialTypeErrorMsg = "credential type %d unknown"

	// ccloud kafka cluster command
	FailedToReadConfirmationErrorMsg = "BYOK error: failed to read your confirmation"
	AuthorizeAccountsErrorMsg = "BYOK error: please authorize the accounts (%s) for the key"
	CKUOnlyForDedicatedErrorMsg = "specifying \"--cku\" flag is valid only for dedicated Kafka cluster creation"
	CKUMoreThanZeroErrorMsg = "\"--cku\" valaue must be greater than 0"
	CloudRegionNotAvailableErrorMsg      = "\"%s\" is not an available region for \"%s\""
	CloudRegionNotAvailableSuggestions   = "You can view a list of available regions for \"%s\" with `ccloud kafka region list --cloud %s` command."
	CloudProviderNotAvailableErrorMsg    = "\"%s\" is not an available cloud provider"
	CloudProviderNotAvailableSuggestions = "You can view a list of available cloud providers and regions with the `ccloud kafka region list` command."
	TopicNotExistsErrorMsg               = "topic \"%s\" does not exist"
	TopicNotExistsSuggestions            = "Check the available topics for Kafka cluster \"%s\" with `ccloud kafka topic list --cluster %s`."
	InvalidAvailableFlagErrorMsg = "invalid value \"%s\" for \"--availability\" flag"
	InvalidAvailableFlagSuggesions = "Allowed values for \"--availablility\" flag are: %s, %s."
	InvalidTypeFlagErrorMsg = "invalid vale \"%s\" for \"--type\" flag"
	InvalidTypeFlagSuggestions = "Allowed values for \"--type\" flag are: %s, %s, %s."
	NameOrCKUFlagErrorMsg = "must either specify --name with non-empty value or --cku (for dedicated clusters) with positive integer"
	NonEmptyNameErrorMsg = "\"--name\" flag value must not be emtpy"

	// kafka topic command
	FailedToProduceErrorMsg = "Failed to produce offset %d: %s\n"
	ConfigurationFormErrorMsg = "configuration must be in the form of key=value"

	// ksql command
	NoServiceAccountErrorMsg = "no service account found for KSQL cluster \"%s\""

	// prompt command
	ParseTimeOutErrorMsg = "invalid value \"%s\" for \"-t, --timeout\" flag: unable to parse %s as duration or milliseconds"
	ParsePromptFormatErrorMsg = "error parsing prompt format string \"%s\""

	// schema-registry commands
	CompatibilityOrModeErrorMsg  = "must pass either \"--compatibility\" or \"--mode\" flag"
	BothSchemaAndSubjectErrorMsg = "cannot specify both schema ID and subject/version"
	SchemaOrSubjectErrorMsg      = "must specify either schema ID or subject/version"
	SchemaIntegerErrorMsg        = "invalid schema ID \"%s\""
	SchemaIntegerSuggestions     = "Schema ID must be an integer."
	ValidateSRAPIErrorMsg = "failed Schema Registry API key and secret validation"

	// secret commands
	EnterInputTypeErrorMsg = "please enter %s"
	PipeInputTypeErrorMsg = "please pipe %s over stdin"
	SpecifyPassphraseErrorMsg = "please specify '--passphrase -' if you intend to pipe your passphrase over stdin"
	PipePassphraseErrorMsg = "please pipe your passphrase over stdin"

	// update command
	UpdateClientFailurePrefix = "update client failure"
	UpdateClientFailureSuggestions = "Please submit a support ticket.\n" +
		"In the meantime, see link for other ways to download the latest CLI version:\n" +
		"%s"
	ReadingYesFlagErrorMsg = "error reading \"--yes\" flag as bool"
	CheckingForUpdateErrorMsg = "error checking for updates"
	UpdateBinaryErrorMsg = "error updating CLI binary"
	ObtainingReleaseNotesErrorMsg = "error obtaining release notes: %s"
	ReleaseNotesVersionCheckErrorMsg = "unable to perform release notes and binary version check: %s"
	ReleaseNotesVersionMismatchErrorMsg = "binary version (v%s) and latest release notes version (v%s) mismatch"

	// package
	// auth package
	NoReaderForCustomCertErrorMsg = "no reader specified for reading custom certificates"
	ReadCertErrorMsg                 = "failed to read certificate"
	NoCertsAppendedErrorMsg          = "no certs appended, using system certs only"
	WriteToNetrcFileErrorMsg         = "unable to write to netrc file \"%s\""
	ResolvingNetrcFilepathErrorMsg   = "unable to resolve netrc filepath at \"%s\""
	GetNetrcCredentialsErrorMsg      = "unable to get credentials from netrc file \"%s\""
	NetrcCredentialsNotFoundErrorMsg = "login credentials not found in netrc file \"%s\""
	CreateNetrcFileErrorMsg          = "unable to create netrc file \"%s\""

	// cmd package
	InvalidAPIKeyErrorMsg      = "invalid API key \"%s\" for resource \"%s\""
	InvalidAPIKeySuggestions   = "List API key that belongs to resource \"%s\" with `ccloud api-key list --resource %s`.\n" +
		"Create new API key for resource \"%s\" with `ccloud api-key create --resource %s`."
	SRNotEnabledErrorMsg       = "Schema Registry not enabled"
	SRNotEnabledSuggestions     = "Schema Registry must be enabled for the environment in order to run the command.\n" +
		"You can enable Schema Registry for this environment with `ccloud schema-registry cluster enable`."
	EnvironmentNotFoundErrorMsg = "environment \"%s\" not found in context \"%s\""
	MalformedJWTNoExprErrorMsg = "malformed JWT claims: no expiration"

	// config package
	UnableToCreateConfigErrorMsg = "unable to create config"
	UnableToReadConfigErrorMsg   = "unable to read config file \"%s\""
	ConfigNotUpToDateErrorMsg    = "config version v%s not up to date with the latest version v%s"
	InvalidConfigVersionErrorMsg = "invalid config version V%s"
	ParseConfigErrorMsg = "unable to parse config file \"%s\""
	NoNameContextErrorMsg = "one of the existing contexts has no name"
	MissingKafkaClusterContextErrorMsg = "context \"%s\" missing KafkaClusterContext"
	MarshalConfigErrorMsg = "unable to marshal config"
	CreateConfigDirectoryErrorMsg = "unable to create config directory: %s"
	CreateConfigFileErrorMsg = "unable to write config to file: %s"
	CurrentContextNotExistErrorMsg = "the current context \"%s\" does not exist"
	ContextNotExistErrorMsg = "context \"%s\" does not exist"
	ContextNameExistsErrorMsg = "cannot create context \"%s\": context with this name already exists"
	CredentialNotFoundErrorMsg = "credential \"%s\" not found"
	PlatformNotFoundErrorMsg = "platform \"%s\" not found"
	NoNameCredentialErrorMsg = "credential must have a name"
	NoNamePlatformErrorMsg = "platform must have a name"
	ResolvingConfigPathErrorMsg = "error resolving the config filepath at %s has occurred"
	ResolvingConfigPathSuggestions = "Please try moving the config file to a different location."
	UnspecifiedPlatformErrorMsg   = "corrupted config: context \"%s\" has corrupted platform"
	UnspecifiedCredentialErrorMsg = "corrupted config: context \"%s\" has corrupted credentials"
	ContextStateMismatchErrorMsg = "corrupted config: context state mismatch for context \"%s\""
	ContextStateNotMappedErrorMsg = "corrupted config: context state mapping error for context \"%s\""
	CorruptedConfigErrorPrefix = "corrupted CLI config"
	CorruptedConfigSuggestions = "Your CLI config file \"%s\" is corrupted. Please remove the file, and run `%s login` or `%s init`.\n" +
		"Unfortunately, your CLI state will be lost as a result.\n" +
		"Please file a support ticket with details about your config file to help us address this issue."
	ClearInvalidAPIFailErrorMsg = "unable to clear invalid API key pairs"
	DeleteUserAuthErrorMsg = "unable to delete user auth"
	ResetInvalidAPIKeyErrorMsg = "unable to reset invalid active API key"
	NoIDClusterErrorMsg = "Kafka cluster under context '%s' has no ID"

	// secret package
	EncryptPlainTextErrorMsg       = "failed to encrypt the plain text"
	DecryptCypherErrorMsg            = "failed to decrypt the cipher"
	DataCorruptedErrorMsg            = "failed to decrypt the cipher: data is corrupted"
	ConfigNotPresentInJAASErrorMsg   = "the configuration \"%s\" not present in JAAS configuration"
	OperationNotSupportedErrorMsg   = "the operation \"%s\" is not supported"
	InvalidJAASConfigErrorMsg       = "invalid JAAS configuration: %s"
	ExpectedConfigNameErrorMsg      = "expected a configuration name but received \"%s\""
	LoginModuleControlFlagErrorMsg  = "login module control flag is not specified"
	ConvertPropertiesToJAASErrorMsg = "failed to convert the properties to a JAAS configuration"
	ValueNotSpecifiedForKeyErrorMsg = "value is not specified for the key \"%s\""
	MissSemicolonErrorMsg           = "configuration not terminated with a ';'"
	EmptyPassphraseErrorMsg         = "master key passphrase cannot be empty"
	AlreadyGeneratedErrorMsg = "master key is already generated"
	AlreadyGeneratedSuggestions = "You can rotate the key with `confluent secret file rotate`."
	InvalidConfigFilePathErrorMsg = "invalid config file path \"%s\""
	InvalidSecretFilePathErrorMsg = "invalid secrets file path \"%s\""
	UnwrapDataKeyErrorMsg = "failed to unwrap the data key: invalid master key or corrupted data key"
	DecryptConfigErrorMsg = "failed to decrypt config \"%s\": corrupted data"
	SecretConfigFileMissingKeyErrorMsg = "missing config key \"%s\" in secret config file"
	IncorrectPassphraseErrorMsg = "authentication failure: incorrect master key passphrase"
	SamePassphraseErrorMsg = "new master key passphrase may not be the same as the previous passphrase"
	EmptyNewConfigListErrorMsg = "add failed: empty list of new configs"
	EmptyUpdateConfigListErrorMsg = "update failed: empty list of update configs"
	ConfigKeyNotEncryptedErrorMsg   = "configuration key \"%s\" is not encrypted"
	FileTypeNotSupportedErrorMsg    = "file type \"%s\" currently not supported"
	ConfigKeyNotInJSONErrorMsg      = "configuration key \"%s\" not present in JSON configuration file"
	MasterKeyNotExportedErrorMsg    = "master key is not exported in '%s' environment variable"
	MasterKeyNotExportedSuggestions = "Please set the environment variable '%s' to the master key and execute this command again."
	ConfigKeyNotPresentErrorMsg     = "configuration key \"%s\" not present in the configuration file"
	InvalidJSONFileFormatErrorMsg   = "invalid json file format"
	InvalidFilePathErrorMsg         = "invalid file path \"%s\""
	UnsupportedFileFormatErrorMsg   = "unsupported file format for file \"%s\""

	// sso package
	StartHTTPServerErrorMsg = "unable to start HTTP server"
	AuthServerRunningErrorMsg = "CLI HTTP auth server encountered error while running: %s\n"
	AuthServerShutdownErrorMsg = "CLI HTTP auth server encountered error while shutting down: %s\n"
	BrowserAuthTimedOutErrorMsg = "timed out while waiting for browser authentication to occur"
	BrowserAuthTimedOutSuggestions = "Please try logging in again."
	LoginFailedCallbackURLErrorMsg = "authentication callback URL either did not contain a state parameter in query string, or the state parameter was invalid; login will fail"
	LoginFailedQueryStringErrorMsg = "authentication callback URL did not contain code parameter in query string; login will fail"
	ReadCallbackPageTemplateErrorMsg = "could not read callback page template"
	PastedInputErrorMsg = "pasted input had invalid format"
	LoginFailedStateParamErrorMsg      = "authentication code either did not contain a state parameter or the state parameter was invalid; login will fail"
	OpenWebBrowserErrorMsg             = "unable to open web browser for authorization"
	GenerateRandomSSOProviderErrorMsg  = "unable to generate random bytes for SSO provider state"
	GenerateRandomCodeVerifierErrorMsg = "unable to generate random bytes for code verifier"
	ComputeHashErrorMsg                = "unable to compute hash for code challenge"
	MissingIDTokenFieldErrorMsg        = "oauth token response body did not contain id_token field"
	ConstructOAuthRequestErrorMsg      = "failed to construct oauth token request"
	UnmarshalOAuthTokenErrorMsg        = "failed to unmarshal response body in oauth token request"

	// update package
	ParseVersionErrorMsg = "unable to parse %s version %s"
	TouchLastCheckFileErrorMsg = "unable to touch last check file"
	GetTempDirErrorMsg = "unable to get temp dir for %s"
	DownloadVersionErrorMsg = "unable to download %s version %s to %s"
	MoveFileErrorMsg = "unable to move %s to %s"
	MoveRestoreErrorMsg = "unable to move (restore) %s to %s"
	CopyErrorMsg = "unable to copy %s to %s"
	ChmodErrorMsg = "unable to chmod 0755 %s"
	SepNonEmptyErrorMsg = "sep must be a non-empty string"
	NoVersionsErrorMsg = "no versions found"
	GetBinaryVersionsErrorMsg = "unable to get available binary versions"
	GetReleaseNotesVersionsErrorMsg = "unable to get available release notes versions"
	UnexpectedS3ResponseErrorMsg = "received unexpected response from S3: %s"
	MissingRequiredParamErrorMsg = "missing required parameter: %s"
	ListingS3BucketErrorMsg = "error listing s3 bucket"
	FindingCredsErrorMsg = "error while finding credentials"
	MissingAccessKeyIDErrorMsg = "access key id is empty for %s"
	AWSCredsExpiredErrorMsg = "AWS credentials in profile %s are expired"
	FindAWSCredsErrorMsg = "failed to find aws credentials in profiles: %s"

	//
	FindKafkaNoClientErrorMsg   = "unable to obtain Kafka cluster information for cluster \"%s\": no client"
	ResourceNotFoundErrorMsg    = "resource \"%s\" not found"
	ResourceNotFoundSuggestions = "Please check that the resource \"%s\" exists.\n" +
		" To list Kafka clusters use `ccloud kafka cluster list`\n" +
		" To check schema-registry cluster info use `ccloud schema-registry cluster describe`\n" +
		" To list KSQL clusters use `ccloud ksql app list`."
	KafkaNotFoundSuggestions   = "List Kafka clusters with `ccloud kafka cluster list`."
	KSQLNotFoundSuggestions    = "List KSQL clusters with `ccloud ksql app list`."
	SRNotFoundSuggestions      = "Check the schema-registry cluster ID with `ccloud schema-registry cluster describe`."
	KafkaNotReadyErrorMsg      = "Kafka cluster \"%s\" not ready"
	KafkaNotReadySuggestions   = "It may take up to 5 minutes for a recently created Kafka cluster to be ready."
	NoKafkaSelectedErrorMsg    = "no Kafka cluster selected"
	NoKafkaSelectedSuggestions = "You must pass \"--cluster\" flag with the command or set an active kafka in your context with `ccloud kafka cluster use`"
	NoAPIKeySelectedErrorMsg   = "no API key selected for resource \"%s\""
	NoAPIKeySelectedSuggestions = "Select an API key for resource \"%s\" with `ccloud api-key use <API_KEY> --resource %s`.\n" +
		"If the resource does not have an API key stored in local CLI state, you must first either create an API key or store an existing key in the CLI.\n" +
		"To create an API key use `ccloud api-key create --resource %s`.\n" +
		"To store an existing API key use `ccluod api-key store --resource %s`."

	UnableToConnectToKafkaErrorMsg = "unable to connect to Kafka cluster"
	UnableToConnectToKafkaSuggestions = "For recently created API keys, it may take a couple of minutes before the keys are ready." +
		"Otherwise, verify that for Kafka cluster \"%s\", the active API key \"%s\" has the correct API secret stored.\n" +
		"If incorrect, override the API secret with `ccloud api-key store %s --resource %s --force`."
	NoAPISecretStoredErrorMsg   = "no API secret for API key \"%s\" of resource \"%s\" stored in local CLI state"
	NoAPISecretStoredSuggestions = "Store the API secret with `ccloud api-key store %s --resource %s`."



	//
	NotLoggedInErrorMsg    = "not logged in"
	NotLoggedInSuggestions = "You must be logged in to run this command.\n" +
		"To avoid session timeouts, you can save credentials to netrc with `%s login --save`."
	CorruptedTokenErrorMsg = "corrupted auth token"
	CorruptedTokenSuggestions = "Please log in again.\n" +
		"To automatically recover from corrupted token, you can save credentials to netrc with \"--save\" flag during `login`."
	ExpiredTokenErrorMsg = "expired token"
	ExpiredTokenSuggestions = "Your session has timed out, please log in again.\n" +
		"To avoid session timeouts, you can save credentials to netrc with \"--save\" flag during `login`."
	InvalidLoginErrorMsg = "incorrect username or password"
	InvalidLoginSUggestions = "Please login again.\n" +
		"To avoid session timeouts, you can save credentials to netrc with \"--save\" flag."


	// Special error types
	GenericOpenAPIErrorMsg = "metadata service backend error: %s: %s"
)
