package errors

/*
	Error message and suggestions message associated with them
*/

const (
	ApiKeyNotFoundSuggestions         = "Ensure the API key exists and has not been deleted, or create a new API key via `confluent api-key create`."
	BadServiceAccountIdErrorMsg       = `failed to parse service account id: ensure service account id begins with "sa-"`
	ByokKeyNotFoundSuggestions        = "Ensure the self-managed key exists and has not been deleted, or register a new key via `confluent byok register`."
	EndOfFreeTrialErrorMsg            = `organization "%s" has been suspended because your free trial has ended`
	EndOfFreeTrialSuggestions         = "To continue using Confluent Cloud, please enter a credit card with `confluent admin payment update` or claim a promo code with `confluent admin promo add`. To enter payment via the UI, please go to https://confluent.cloud/login."
	EnsureCpSixPlusSuggestions        = "Ensure that you are running against MDS with CP 6.0+."
	ExactlyOneSetErrorMsg             = "exactly one of %v must be set"
	MoreThanOneNonKafkaErrorMsg       = "cannot specify more than one non-Kafka cluster ID for a scope"
	MustSetAllowOrDenyErrorMsg        = "`--allow` or `--deny` must be set when adding or deleting an ACL"
	MustSetResourceTypeErrorMsg       = "exactly one resource type (%s) must be set"
	ServiceAccountNotFoundErrorMsg    = `service account "%s" not found`
	ServiceAccountNotFoundSuggestions = "List service accounts with `confluent service-account list`."
	SpecifyKafkaIDErrorMsg            = "must specify `--kafka-cluster` to uniquely identify the scope"
	UnknownConnectorIdErrorMsg        = `unknown connector ID "%s"`

	// kafka cluster commands
	ListTopicSuggestions                             = "To list topics for the cluster \"%s\", use `confluent kafka topic list --cluster %s`."
	FailedToReadConfirmationErrorMsg                 = "BYOK error: failed to read your confirmation"
	FailedToReadClusterResizeConfirmationErrorMsg    = "cluster resize error: failed to read your confirmation"
	AuthorizeAccountsErrorMsg                        = "BYOK error: please authorize the key for the accounts (%s)x"
	AuthorizeIdentityErrorMsg                        = "BYOK error: please authorize the key for the identity (%s)"
	EncryptionKeySupportErrorMsg                     = "BYOK via `--encryption-key` is only available for GCP. Use `confluent byok create` to register AWS and Azure keys."
	CKUMoreThanZeroErrorMsg                          = "`--cku` value must be greater than 0"
	CKUMoreThanOneErrorMsg                           = "`--cku` value must be greater than 1 for High Durability"
	ClusterResizeNotSupportedErrorMsg                = "failed to update kafka cluster: cluster resize is only supported on dedicated clusters"
	CloudRegionNotAvailableErrorMsg                  = `"%s" is not an available region for "%s"`
	CloudRegionNotAvailableSuggestions               = "To view a list of available regions for \"%s\", use `confluent kafka region list --cloud %s`."
	CloudProviderNotAvailableErrorMsg                = `"%s" is not an available cloud provider`
	CloudProviderNotAvailableSuggestions             = "To view a list of available cloud providers and regions, use `confluent kafka region list`."
	TopicDoesNotExistOrMissingPermissionsErrorMsg    = `topic "%s" does not exist or user does not have the ACLs or role bindings required to describe it`
	TopicDoesNotExistOrMissingPermissionsSuggestions = "To list topics for Kafka cluster \"%s\", use `confluent kafka topic list --cluster %s`.\nTo list ACLs use `confluent kafka acl list --cluster %s`.\nTo list role bindings use `confluent iam rbac role-binding list`."
	InvalidAvailableFlagErrorMsg                     = "invalid value \"%s\" for `--availability` flag"
	InvalidAvailableFlagSuggestions                  = "Allowed values for `--availability` flag are: %s, %s."
	InvalidTypeFlagErrorMsg                          = "invalid value \"%s\" for `--type` flag"
	NonEmptyNameErrorMsg                             = "`--name` flag value must not be empty"
	KafkaClusterNotFoundErrorMsg                     = `Kafka cluster "%s" not found`
	KafkaClusterStillProvisioningErrorMsg            = "your cluster is still provisioning, so it can't be updated yet; please retry in a few minutes"
	KafkaClusterUpdateFailedSuggestions              = "A cluster can't be updated while still provisioning. If you just created this cluster, retry in a few minutes."
	KafkaClusterExpandingErrorMsg                    = "your cluster is expanding; please wait for that operation to complete before updating again"
	KafkaClusterShrinkingErrorMsg                    = "your cluster is shrinking; Please wait for that operation to complete before updating again"
	KafkaClusterInaccessibleErrorMsg                 = `Kafka cluster "%s" not found or access forbidden`
	KafkaClusterInaccessibleSuggestions              = ChooseRightEnvironmentSuggestions + "\n" +
		"The active Kafka cluster may have been deleted. Set a new active cluster with `confluent kafka cluster use`."
	ChooseRightEnvironmentSuggestions = "Ensure the cluster ID you entered is valid.\n" +
		"Ensure the cluster you are specifying belongs to the currently selected environment with `confluent kafka cluster list`, `confluent environment list`, and `confluent environment use`."
	UnknownTopicErrorMsg              = `unknown topic "%s"`
	MdsUrlNotFoundSuggestions         = "Pass the `--url` flag or set the `CONFLUENT_PLATFORM_MDS_URL` environment variable."
	KafkaClusterMissingPrefixErrorMsg = `Kafka cluster "%s" is missing required prefix "lkc-"`

	// kafka topic commands
	FailedToCreateProducerErrorMsg       = "failed to create producer: %v"
	FailedToCreateConsumerErrorMsg       = "failed to create consumer: %v"
	FailedToCreateAdminClientErrorMsg    = "failed to create confluent-kafka-go admin client: %v"
	InvalidOffsetErrorMsg                = "offset value must be a non-negative integer"
	InvalidSecurityProtocolErrorMsg      = "security protocol not supported: %v"
	TopicExistsOnPremErrorMsg            = `topic "%s" already exists for the Kafka cluster`
	TopicExistsOnPremSuggestions         = "To list topics for the cluster, use `confluent kafka topic list --url <url>`."
	FailedToProduceErrorMsg              = "failed to produce offset %d: %s\n"
	MissingKeyErrorMsg                   = "missing key in message"
	UnknownValueFormatErrorMsg           = "unknown value schema format"
	TopicExistsErrorMsg                  = `topic "%s" already exists for Kafka cluster "%s"`
	TopicExistsSuggestions               = "To list topics for the cluster \"%s\", use `confluent kafka topic list --cluster %s`."
	NoAPISecretStoredOrPassedErrorMsg    = `no API secret for API key "%s" of resource "%s" passed via flag or stored in local CLI state`
	NoAPISecretStoredOrPassedSuggestions = "Pass the API secret with flag `--api-secret` or store with `confluent api-key store %s --resource %s`."
	PassedSecretButNotKeyErrorMsg        = "no API key specified"
	PassedSecretButNotKeySuggestions     = "Use the `--api-key` flag to specify an API key."
	ProducingToCompactedTopicErrorMsg    = "producer has detected an INVALID_RECORD error for topic %s"
	ProducingToCompactedTopicSuggestions = "If the topic has schema validation enabled, ensure you are producing with a schema-enabled producer.\n" +
		"If your topic is compacted, ensure you are producing a record with a key."
	ExceedPartitionLimitSuggestions = "The total partition limit for a dedicated cluster may be increased by expanding its CKU count using `confluent kafka cluster update <id> --cku <count>`."

	// serialization/deserialization commands
	JsonSchemaInvalidErrorMsg         = "the JSON schema is invalid"
	JsonDocumentInvalidErrorMsg       = "the JSON document is invalid"
	AvroReferenceNotSupportedErrorMsg = "avro reference not supported in cloud CLI"
	ProtoSchemaInvalidErrorMsg        = "the protobuf schema is invalid"
	ProtoDocumentInvalidErrorMsg      = "the protobuf document is invalid"

	// ksql commands
	KsqlDBNoServiceAccountErrorMsg = `ACLs do not need to be configured for the ksqlDB cluster, "%s", because it was created with user-level access to the Kafka cluster`
	KsqlDBTerminateClusterErrorMsg = `failed to terminate ksqlDB cluster "%s" due to "%s"`

	// local commands
	NoServicesRunningErrorMsg = "no services running"
	TopNotAvailableErrorMsg   = "top command not available on platform: %s"
	InvalidConnectorErrorMsg  = "invalid connector: %s"
	FailedToStartErrorMsg     = "%s failed to start"
	FailedToStopErrorMsg      = "%s failed to stop"
	JavaRequirementErrorMsg   = "the Confluent CLI requires Java version 1.8 or 1.11.\n" +
		"See https://docs.confluent.io/current/installation/versions-interoperability.html .\n" +
		"If you have multiple versions of Java installed, you may need to set JAVA_HOME to the version you want Confluent to use."
	NoLogFoundErrorMsg              = "no log found: to run %s, use `confluent local services %s start`"
	MacVersionErrorMsg              = "macOS version >= %s is required (detected: %s)"
	JavaExecNotFondErrorMsg         = "could not find java executable, please install java or set JAVA_HOME"
	NothingToDestroyErrorMsg        = "nothing to destroy"
	ComputePoolNotFoundErrorMsg     = `Flink compute pool "%s" not found or access forbidden`
	ComputePoolNotFoundSuggestions  = "List available Flink compute pools with `confluent flink compute-pool list`.\nMake sure you have selected the compute pool's environment with `confluent environment use`."
	FailedToReadPortsErrorMsg       = "failed to read local ports from config"
	FailedToReadPortsSuggestions    = "Restart Confluent Local with `confluent local kafka stop` and `confluent local kafka start`"
	InstallAndStartDockerSuggestion = "Make sure Docker is installed following the guide: `https://docs.docker.com/engine/install/` and Docker daemon is running."

	// schema-registry commands
	InvalidSchemaRegistryLocationErrorMsg    = "invalid input for flag `--geo`"
	InvalidSchemaRegistryLocationSuggestions = `Geo must be either "us", "eu", or "apac".`
	CompatibilityOrModeErrorMsg              = "must pass either `--compatibility` or `--mode` flag"
	BothSchemaAndSubjectErrorMsg             = "cannot specify both schema ID and subject/version"
	SchemaOrSubjectErrorMsg                  = "must specify either schema ID or subject/version"
	SchemaIntegerErrorMsg                    = `invalid schema ID "%s"`
	SchemaIntegerSuggestions                 = "Schema ID must be an integer."
	SchemaNotFoundErrorMsg                   = "Schema Registry subject or version not found"
	SchemaNotFoundSuggestions                = "List available subjects with `confluent schema-registry subject list`.\n" +
		"List available versions with `confluent schema-registry subject describe`."
	SRInvalidPackageTypeErrorMsg = `"%s" is an invalid package type`
	SRInvalidPackageSuggestions  = "Allowed values for `--package` flag are: %s."
	SRInvalidPackageUpgrade      = "Environment \"%s\" is already using the Stream Governance \"%s\" package.\n"

	// secret commands
	EnterInputTypeErrorMsg    = "enter %s"
	PipeInputTypeErrorMsg     = "pipe %s over stdin"
	SpecifyPassphraseErrorMsg = "specify `--passphrase -` if you intend to pipe your passphrase over stdin"
	PipePassphraseErrorMsg    = "pipe your passphrase over stdin"

	// update command
	ReadingYesFlagErrorMsg              = "error reading `--yes` flag as bool"
	CheckingForUpdateErrorMsg           = "error checking for updates"
	UpdateBinaryErrorMsg                = "error updating CLI binary"
	ObtainingReleaseNotesErrorMsg       = "error obtaining release notes: %s"
	ReleaseNotesVersionCheckErrorMsg    = "unable to perform release notes and binary version check: %s"
	ReleaseNotesVersionMismatchErrorMsg = "binary version (v%s) and latest release notes version (v%s) mismatch"

	// auth package
	NoReaderForCustomCertErrorMsg    = "no reader specified for reading custom certificates"
	ReadCertErrorMsg                 = "failed to read certificate"
	CaCertNotSpecifiedErrorMsg       = "no CA certificate specified"
	SRCaCertSuggestions              = "Please specify `--ca-location` to enable Schema Registry client."
	NoCertsAppendedErrorMsg          = "no certs appended, using system certs only"
	WriteToNetrcFileErrorMsg         = `unable to write to netrc file "%s"`
	NetrcCredentialsNotFoundErrorMsg = `login credentials not found in netrc file "%s"`
	FailedToObtainedUserSSOErrorMsg  = `unable to obtain SSO info for user "%s"`
	NonSSOUserErrorMsg               = `tried to obtain SSO token for non SSO user "%s"`
	NoCredentialsFoundErrorMsg       = "no credentials found"
	NoURLEnvVarErrorMsg              = "no URL env var"
	InvalidInputFormatErrorMsg       = `"%s" is not of valid format for field "%s"`
	ParseKeychainCredentialsErrorMsg = "unable to parse credentials in keychain access"

	// cmd package
	InvalidAPIKeyErrorMsg    = `invalid API key "%s" for resource "%s"`
	InvalidAPIKeySuggestions = "To list API key that belongs to resource \"%s\", use `confluent api-key list --resource %s`.\n" +
		"To create new API key for resource \"%s\", use `confluent api-key create --resource %s`."
	SRNotEnabledErrorMsg    = "Schema Registry not enabled"
	SRNotEnabledSuggestions = "Schema Registry must be enabled for the environment in order to run the command.\n" +
		"You can enable Schema Registry for this environment with `confluent schema-registry cluster enable`."

	// config package
	CorruptedConfigErrorPrefix = "corrupted CLI config"
	CorruptedConfigSuggestions = "Your configuration file \"%s\" is corrupted.\n" +
		"Remove config file, and run `confluent login` or `confluent context create`.\n" +
		"Unfortunately, your active CLI state will be lost as a result.\n" +
		"Please file a support ticket with details about your config file to help us address this issue.\n" +
		"Please rerun the command with the verbosity flag `-vvvv` and attach the output with the support ticket."
	UnableToReadConfigurationFileErrorMsg = `unable to read configuration file "%s"`
	NoNameContextErrorMsg                 = "one of the existing contexts has no name"
	MissingKafkaClusterContextErrorMsg    = `context "%s" missing KafkaClusterContext`
	MarshalConfigErrorMsg                 = "unable to marshal config"
	CreateConfigDirectoryErrorMsg         = "unable to create config directory: %s"
	CreateConfigFileErrorMsg              = "unable to write config to file: %s"
	CurrentContextNotExistErrorMsg        = `the current context "%s" does not exist`
	ContextDoesNotExistErrorMsg           = `context "%s" does not exist`
	ContextAlreadyExistsErrorMsg          = `context "%s" already exists`
	CredentialNotFoundErrorMsg            = `credential "%s" not found`
	PlatformNotFoundErrorMsg              = `platform "%s" not found`
	NoNameCredentialErrorMsg              = "credential must have a name"
	SavedCredentialNoContextErrorMsg      = "saved credential must match a context"
	KeychainNotAvailableErrorMsg          = "keychain not available on platforms other than darwin"
	NoValidKeychainCredentialErrorMsg     = "no matching credentials found in keychain"
	NoNamePlatformErrorMsg                = "platform must have a name"
	UnspecifiedPlatformErrorMsg           = `context "%s" has corrupted platform`
	UnspecifiedCredentialErrorMsg         = `context "%s" has corrupted credentials`
	ContextStateMismatchErrorMsg          = `context state mismatch for context "%s"`
	ContextStateNotMappedErrorMsg         = `context state mapping error for context "%s"`
	DeleteUserAuthErrorMsg                = "unable to delete user auth"

	// local package
	ConfluentHomeNotFoundErrorMsg         = "could not find %s in CONFLUENT_HOME"
	SetConfluentHomeErrorMsg              = "set environment variable CONFLUENT_HOME"
	KafkaScriptFormatNotSupportedErrorMsg = "format %s is not supported in this version"
	KafkaScriptInvalidFormatErrorMsg      = "invalid format: %s"

	// secret package
	DataCorruptedErrorMsg              = "failed to decrypt the cipher: data is corrupted"
	ConfigNotInJAASErrorMsg            = `the configuration "%s" not present in JAAS configuration`
	OperationNotSupportedErrorMsg      = `the operation "%s" is not supported`
	InvalidJAASConfigErrorMsg          = "invalid JAAS configuration: %s"
	ExpectedConfigNameErrorMsg         = `expected a configuration name but received "%s"`
	LoginModuleControlFlagErrorMsg     = "login module control flag is not specified"
	ConvertPropertiesToJAASErrorMsg    = "failed to convert the properties to a JAAS configuration"
	ValueNotSpecifiedForKeyErrorMsg    = `value is not specified for the key "%s"`
	MissSemicolonErrorMsg              = "configuration not terminated with a ';'"
	EmptyPassphraseErrorMsg            = "master key passphrase cannot be empty"
	AlreadyGeneratedErrorMsg           = "master key is already generated"
	AlreadyGeneratedSuggestions        = "You can rotate the key with `confluent secret file rotate`."
	InvalidConfigFilePathErrorMsg      = `invalid config file path "%s"`
	InvalidSecretFilePathErrorMsg      = `invalid secrets file path "%s"`
	UnwrapDataKeyErrorMsg              = "failed to unwrap the data key: invalid master key or corrupted data key"
	DecryptConfigErrorMsg              = `failed to decrypt config "%s": corrupted data`
	SecretConfigFileMissingKeyErrorMsg = `missing config key "%s" in secret config file`
	IncorrectPassphraseErrorMsg        = "authentication failure: incorrect master key passphrase"
	SamePassphraseErrorMsg             = "new master key passphrase may not be the same as the previous passphrase"
	EmptyNewConfigListErrorMsg         = "add failed: empty list of new configs"
	EmptyUpdateConfigListErrorMsg      = "update failed: empty list of update configs"
	ConfigKeyNotEncryptedErrorMsg      = `configuration key "%s" is not encrypted`
	FileTypeNotSupportedErrorMsg       = `file type "%s" currently not supported`
	ConfigKeyNotInJSONErrorMsg         = `configuration key "%s" not present in JSON configuration file`
	MasterKeyNotExportedErrorMsg       = "master key is not exported in `%s` environment variable"
	MasterKeyNotExportedSuggestions    = "Set the environment variable `%s` to the master key and execute this command again."
	ConfigKeyNotPresentErrorMsg        = `configuration key "%s" not present in the configuration file`
	InvalidJSONFileFormatErrorMsg      = "invalid json file format"
	InvalidFilePathErrorMsg            = `invalid file path "%s"`
	UnsupportedFileFormatErrorMsg      = `unsupported file format for file "%s"`
	InvalidAlgorithmErrorMsg           = `invalid algorithm "%s"`
	IncorrectNonceLengthErrorMsg       = `incorrect nonce length from ~/.confluent/config.json passed into encryption`

	// sso package
	StartHTTPServerErrorMsg            = "unable to start HTTP server"
	AuthServerRunningErrorMsg          = "CLI HTTP auth server encountered error while running: %s\n"
	AuthServerShutdownErrorMsg         = "CLI HTTP auth server encountered error while shutting down: %s\n"
	BrowserAuthTimedOutErrorMsg        = "timed out while waiting for browser authentication to occur"
	BrowserAuthTimedOutSuggestions     = "Try logging in again."
	LoginFailedCallbackURLErrorMsg     = "authentication callback URL either did not contain a state parameter in query string, or the state parameter was invalid; login will fail"
	LoginFailedQueryStringErrorMsg     = "authentication callback URL did not contain code parameter in query string; login will fail"
	PastedInputErrorMsg                = "pasted input had invalid format"
	LoginFailedStateParamErrorMsg      = "authentication code either did not contain a state parameter or the state parameter was invalid; login will fail"
	OpenWebBrowserErrorMsg             = "unable to open web browser for authorization"
	GenerateRandomSSOProviderErrorMsg  = "unable to generate random bytes for SSO provider state"
	GenerateRandomCodeVerifierErrorMsg = "unable to generate random bytes for code verifier"
	ComputeHashErrorMsg                = "unable to compute hash for code challenge"
	FmtMissingOAuthFieldErrorMsg       = `oauth token response body did not contain field "%s"`
	ConstructOAuthRequestErrorMsg      = "failed to construct oauth token request"
	UnmarshalOAuthTokenErrorMsg        = "failed to unmarshal response body in oauth token request"

	// update package
	ParseVersionErrorMsg            = "unable to parse %s version %s"
	TouchLastCheckFileErrorMsg      = "unable to touch last check file"
	GetTempDirErrorMsg              = "unable to get temp dir for %s"
	DownloadVersionErrorMsg         = "unable to download %s version %s to %s"
	SepNonEmptyErrorMsg             = "sep must be a non-empty string"
	NoVersionsErrorMsg              = "no versions found"
	GetBinaryVersionsErrorMsg       = "unable to get available binary versions"
	GetReleaseNotesVersionsErrorMsg = "unable to get available release notes versions"
	UnexpectedS3ResponseErrorMsg    = "received unexpected response from S3: %s"

	// plugin package
	NoVersionFoundErrorMsg = "no version found in plugin manifest"

	// catcher
	CCloudBackendErrorPrefix      = "Confluent Cloud backend error"
	UnexpectedBackendOutputPrefix = "unexpected CCloud backend output"
	ResourceNotFoundErrorMsg      = `resource "%s" not found`
	ResourceNotFoundSuggestions   = "Check that the resource \"%s\" exists.\n" +
		"To list Kafka clusters, use `confluent kafka cluster list`.\n" +
		"To check Schema Registry cluster information, use `confluent schema-registry cluster describe`.\n" +
		"To list KSQL clusters, use `confluent ksql cluster list`."
	KafkaNotFoundErrorMsg         = `Kafka cluster "%s" not found`
	KafkaNotFoundSuggestions      = "To list Kafka clusters, use `confluent kafka cluster list`."
	KSQLNotFoundSuggestions       = "To list KSQL clusters, use `confluent ksql cluster list`."
	NoKafkaSelectedErrorMsg       = "no Kafka cluster selected"
	NoKafkaSelectedSuggestions    = "You must pass `--cluster` with the command or set an active Kafka cluster in your context with `confluent kafka cluster use`."
	NoKafkaForDescribeSuggestions = "You must provide the cluster ID argument or set an active Kafka cluster in your context with `confluent kafka cluster use`."
	NoAPISecretStoredErrorMsg     = `no API secret for API key "%s" of resource "%s" stored in local CLI state`
	NoAPISecretStoredSuggestions  = "Store the API secret with `confluent api-key store %s --resource %s`."

	// Kafka REST Proxy errors
	InternalServerErrorMsg            = "internal server error"
	UnknownErrorMsg                   = "unknown error"
	InternalServerErrorSuggestions    = "Please check the status of your Kafka cluster or submit a support ticket."
	EmptyResponseErrorMsg             = "empty server response"
	KafkaRestErrorMsg                 = "Kafka REST request failed: %s %s: %s"
	KafkaRestConnectionErrorMsg       = "unable to establish Kafka REST connection: %s: %s"
	KafkaRestCertErrorSuggestions     = "To specify a CA certificate, please use the `--ca-cert-path` flag or set `CONFLUENT_PLATFORM_CA_CERT_PATH`."
	KafkaRestUrlNotFoundErrorMsg      = "Kafka REST URL not found"
	KafkaRestUrlNotFoundSuggestions   = "Use the `--url` flag or set `CONFLUENT_REST_URL`."
	KafkaRestProvisioningErrorMsg     = `Kafka REST unavailable: Kafka cluster "%s" is still provisioning`
	NoClustersFoundErrorMsg           = "no clusters found"
	NoClustersFoundSuggestions        = "Please check the status of your cluster and the Kafka REST bootstrap.servers configuration."
	NeedClientCertAndKeyPathsErrorMsg = "must set `--client-cert-path` and `--client-key-path` flags together"
	InvalidMDSTokenErrorMsg           = "Invalid MDS token"
	InvalidMDSTokenSuggestions        = "Re-login with `confluent login`."

	// Special error handling
	QuotaExceededSuggestions = "Look up Confluent Cloud service quota limits with `confluent service-quota list`."
	AvoidTimeoutSuggestions  = "To avoid session timeouts, non-SSO users can save their credentials with `confluent login --save`."
	NotLoggedInErrorMsg      = "not logged in"
	AuthTokenSuggestions     = "You must be logged in to retrieve an oauthbearer token.\n" +
		"An oauthbearer token is required to authenticate OAUTHBEARER mechanism and Schema Registry."
	OnPremConfigGuideSuggestions = "See configuration and produce/consume command guide: https://docs.confluent.io/confluent-cli/current/cp-produce-consume.html ."
	NotLoggedInSuggestions       = "You must be logged in to run this command.\n" +
		AvoidTimeoutSuggestions
	SRNotAuthenticatedErrorMsg     = "not logged in, or no Schema Registry endpoint specified"
	SREndpointNotSpecifiedErrorMsg = "no Schema Registry endpoint specified"
	SRClientNotValidatedErrorMsg   = "failed to validate Schema Registry client with token"
	SRNotAuthenticatedSuggestions  = "You must specify the endpoint for a Schema Registry cluster (--schema-registry-endpoint) or be logged in using `confluent login` to run this command.\n" +
		AvoidTimeoutSuggestions
	CorruptedTokenErrorMsg    = "corrupted auth token"
	CorruptedTokenSuggestions = "Please log in again.\n" +
		AvoidTimeoutSuggestions
	ExpiredTokenErrorMsg    = "expired token"
	ExpiredTokenSuggestions = "Your session has timed out, you need to log in again.\n" +
		AvoidTimeoutSuggestions
	InvalidEmailErrorMsg         = `user "%s" not found`
	InvalidLoginURLErrorMsg      = "invalid URL value, see structure: http(s)://<domain/hostname/ip>:<port>/"
	InvalidLoginErrorMsg         = "incorrect email, password, or organization ID"
	InvalidLoginErrorSuggestions = "To log into an organization other than the default organization, use the `--organization-id` flag.\n" +
		AvoidTimeoutSuggestions
	SuspendedOrganizationSuggestions = "Your organization has been suspended, please contact support if you want to unsuspend it."
	FailedToReadInputErrorMsg        = "failed to read input"

	// Partition command errors
	SpecifyPartitionIdWithTopicErrorMsg = "must specify topic along with partition ID"

	// Broker commands
	MustSpecifyAllOrBrokerIDErrorMsg = "must pass broker ID argument or specify `--all` flag"
	OnlySpecifyAllOrBrokerIDErrorMsg = "only specify broker ID argument OR `--all` flag"
	InvalidBrokerTaskTypeErrorMsg    = "invalid broker task type"
	InvalidBrokerTaskTypeSuggestions = "Valid broker task types are `remove-broker` and `add-broker`."

	// Special error types
	GenericOpenAPIErrorMsg       = "metadata service backend error: %s: %s"
	ParsedGenericOpenAPIErrorMsg = "metadata service backend error: %s"

	// FeatureFlags errors
	UnsupportedCustomAttributeErrorMsg = `attribute "%s" is not one of the supported FeatureFlags targeting values`

	// General
	DeleteResourceErrorMsg  = `failed to delete %s "%s": %v`
	ListResourceSuggestions = "List available %ss with `%s list`."
	UpdateResourceErrorMsg  = `failed to update %s "%s": %v`
)
