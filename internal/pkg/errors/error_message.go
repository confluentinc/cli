package errors

/*
	Error message and suggestions message associated with them
*/

const (
	// format
	prefixFormat = "%s: %s"

	// admin commands
	BadResourceIDErrorMsg  = `failed parsing resource ID: missing prefix "%s-" is required`
	BadEmailFormatErrorMsg = "invalid email structure"

	// api-key commands
	BadServiceAccountIDErrorMsg         = `failed to parse service account id: ensure service account id begins with "sa-"`
	UnableToStoreAPIKeyErrorMsg         = "unable to store API key locally"
	NonKafkaNotImplementedErrorMsg      = "functionality not yet available for non-Kafka cluster resources"
	RefuseToOverrideSecretErrorMsg      = `refusing to overwrite existing secret for API Key "%s"`
	RefuseToOverrideSecretSuggestions   = "If you would like to override the existing secret stored for API key \"%s\", use `--force` flag."
	APIKeyUseFailedErrorMsg             = "unable to set active API key"
	APIKeyUseFailedSuggestions          = "If you did not create this API key with the CLI or created it on another computer, you must first store the API key and secret locally with `confluent api-key store %s <secret>`."
	APIKeyNotValidForClusterErrorMsg    = "the provided API key does not belong to the target cluster"
	APIKeyNotValidForClusterSuggestions = "Specify the cluster this API key belongs to using the `--resource` flag. Alternatively, first execute the `confluent kafka cluster use` command to set the context to the proper cluster for this key and retry the `confluent api-key store` command."
	APIKeyNotFoundErrorMsg              = "unknown API key %s"
	APIKeyNotFoundSuggestions           = "Ensure the API key exists and has not been deleted, or create a new API key via `confluent api-key create`."
	ServiceAccountNotFoundErrorMsg      = `service account "%s" not found`
	ServiceAccountNotFoundSuggestions   = "List service accounts with `confluent service-account list`."

	// audit-log command
	EnsureCPSixPlusSuggestions        = "Ensure that you are running against MDS with CP 6.0+."
	UnableToAccessEndpointErrorMsg    = "unable to access endpoint"
	UnableToAccessEndpointSuggestions = EnsureCPSixPlusSuggestions
	AuditLogsNotEnabledErrorMsg       = "Audit Logs are not enabled for this organization"
	MalformedConfigErrorMsg           = "bad input file: the audit log configuration for cluster %q uses invalid JSON: %v"

	// byok commands
	ByokKeyNotFoundSuggestions = "Ensure the self-managed key exists and has not been deleted, or register a new key via `confluent byok register`."
	ByokUnknownKeyTypeErrorMsg = "unknown byok key type"

	// login command
	UnneccessaryUrlFlagForCloudLoginErrorMsg         = "there is no need to pass the url flag if you are logging in to Confluent Cloud"
	UnneccessaryUrlFlagForCloudLoginSuggestions      = "Log in to Confluent Cloud with `confluent login`."
	SSOCredentialsDoNotMatchLoginCredentialsErrorMsg = "expected SSO credentials for %s but got credentials for %s"
	SSOCredentialsDoNotMatchSuggestions              = "Please re-login and use the same email at the prompt and in the SSO portal."
	EndOfFreeTrialErrorMsg                           = `organization "%s" has been suspended because your free trial has ended`
	EndOfFreeTrialSuggestions                        = `To continue using Confluent Cloud, please enter a credit card with "confluent admin payment update" or claim a promo code with "confluent admin promo add". To enter payment via the UI, please go to confluent.cloud/login .`

	// confluent cluster commands
	FetchClusterMetadataErrorMsg     = "unable to fetch cluster metadata: %s - %s"
	AccessClusterRegistryErrorMsg    = "unable to access Cluster Registry"
	AccessClusterRegistrySuggestions = EnsureCPSixPlusSuggestions
	MustSpecifyOneClusterIDErrorMsg  = "must specify at least one cluster ID"
	ProtocolNotSupportedErrorMsg     = "protocol %s is currently not supported"
	UnknownClusterErrorMsg           = `unknown cluster "%s"`

	// connect and connector-catalog commands
	UnknownConnectorIdErrorMsg         = `unknown connector ID "%s"`
	EmptyConfigFileErrorMsg            = `connector config file "%s" is empty`
	MissingRequiredConfigsErrorMsg     = `required configs "name" and "connector.class" missing from connector config file "%s"`
	InvalidCloudErrorMsg               = "error defining plugin on given Kafka cluster"
	InvalidCloudSuggestions            = "To list available connector plugin types, use `confluent connect plugin list`."
	ConnectLogEventsNotEnabledErrorMsg = "Connect Log Events are not enabled for this organization"

	// environment & organization command
	EnvNotFoundErrorMsg            = `environment "%s" not found`
	OrgResourceNotFoundSuggestions = "List available %[1]ss with `confluent %[1]s list`."
	EnvSwitchErrorMsg              = "failed to switch environment: failed to save config"
	NoEnvironmentFoundErrorMsg     = "no environment found"
	NoEnvironmentFoundSuggestions  = "This issue may occur if this user has no valid role bindings. Contact an Organization Admin to create a role binding for this user."

	// iam acl & kafka acl commands
	UnableToPerformAclErrorMsg    = "unable to %s ACLs: %s"
	UnableToPerformAclSuggestions = "Ensure that you're running against MDS with CP 5.4+."
	MustSetAllowOrDenyErrorMsg    = "`--allow` or `--deny` must be set when adding or deleting an ACL"
	OnlySetAllowOrDenyErrorMsg    = "only `--allow` or `--deny` may be set when adding or deleting an ACL"
	MustSetResourceTypeErrorMsg   = "exactly one resource type (%v) must be set"
	InvalidOperationValueErrorMsg = "invalid operation value: %s"
	ExactlyOneSetErrorMsg         = "exactly one of %v must be set"
	UserIdNotValidErrorMsg        = "can't map user id to a valid service account"
	PrincipalNotFoundErrorMsg     = `user or service account "%s" not found`

	// iam rbac role commands
	UnknownRoleErrorMsg    = `unknown role "%s"`
	UnknownRoleSuggestions = "The available roles are: %s."

	// iam rbac role-binding commands
	PrincipalFormatErrorMsg         = "incorrect principal format specified"
	PrincipalFormatSuggestions      = "Principal must be specified in this format: \"<Principal Type>:<Principal Name>\".\nFor example, \"User:u-xxxxxx\" or \"User:sa-xxxxxx\"."
	ResourceFormatErrorMsg          = "incorrect resource format specified"
	ResourceFormatSuggestions       = "Resource must be specified in this format: `<Resource Type>:<Resource Name>`."
	LookUpRoleErrorMsg              = `failed to look up role "%s"`
	LookUpRoleSuggestions           = "To check for valid roles, use `confluent iam rbac role list`."
	InvalidResourceTypeErrorMsg     = `invalid resource type "%s"`
	InvalidResourceTypeSuggestions  = "The available resource types are: %s."
	SpecifyKafkaIDErrorMsg          = "must specify `--kafka-cluster` to uniquely identify the scope"
	SpecifyCloudClusterErrorMsg     = "must specify `--cloud-cluster` to indicate role binding scope"
	SpecifyEnvironmentErrorMsg      = "must specify `--environment` to indicate role binding scope"
	BothClusterNameAndScopeErrorMsg = "cannot specify both cluster name and cluster scope"
	SpecifyClusterErrorMsg          = "must specify either cluster ID to indicate role binding scope or the cluster name"
	MoreThanOneNonKafkaErrorMsg     = "cannot specify more than one non-Kafka cluster ID for a scope"
	PrincipalOrRoleRequiredErrorMsg = "must specify either principal or role"
	HTTPStatusCodeErrorMsg          = "no error but received HTTP status code %d"
	HTTPStatusCodeSuggestions       = "Please file a support ticket with details."
	UnauthorizedErrorMsg            = "user is unauthorized to perform this action"
	UnauthorizedSuggestions         = "Check the user's privileges by running `confluent iam rbac role-binding list`.\nGive the user the appropriate permissions using `confluent iam rbac role-binding create`."
	RoleBindingNotFoundErrorMsg     = "failed to look up matching role binding"
	RoleBindingNotFoundSuggestions  = "To list role bindings, use `confluent iam rbac role-binding list`."

	// iam service-account commands
	ServiceNameInUseErrorMsg    = `service name "%s" is already in use`
	ServiceNameInUseSuggestions = "To list all service account, use `confluent iam service-account list`."

	// iam provider commands
	IdentityProviderNoOpUpdateErrorMsg = "one of `--description` or `--name` must be set"

	// iam pool commands
	IdentityPoolNoOpUpdateErrorMsg = "one of `--description`, `--filter`, `--identity-claim`, or `--name` must be set"

	// init command
	CannotBeEmptyErrorMsg         = "%s cannot be empty"
	UnknownCredentialTypeErrorMsg = "credential type %d unknown"

	// kafka client-config package
	FetchConfigFileErrorMsg               = "failed to get config file: error code %d"
	KafkaCredsValidationFailedErrorMsg    = "failed to validate Kafka API credential"
	KafkaCredsValidationFailedSuggestions = "Verify that the correct Kafka API credential is used.\n" +
		"If you are using the stored Kafka API credential, verify that the secret is correct. If incorrect, override with `confluent api-key store -f`.\n" +
		"If you are using the flags, verify that the correct Kafka API credential is passed to `--api-key` and `--api-secret`."
	SRCredsValidationFailedErrorMsg    = "failed to validate Schema Registry API credential"
	SRCredsValidationFailedSuggestions = "Verify that the correct Schema Registry API credential is passed to `--schema-registry-api-key` and --schema-registry-api-secret`."

	// kafka cluster commands
	ListTopicSuggestions                             = "To list topics for the cluster \"%s\", use `confluent kafka topic list --cluster %s`."
	FailedToRenderKeyPolicyErrorMsg                  = "BYOK error: failed to render key policy"
	FailedToReadConfirmationErrorMsg                 = "BYOK error: failed to read your confirmation"
	FailedToReadClusterResizeConfirmationErrorMsg    = "cluster resize error: failed to read your confirmation"
	AuthorizeAccountsErrorMsg                        = "BYOK error: please authorize the key for the accounts (%s)x"
	AuthorizeIdentityErrorMsg                        = "BYOK error: please authorize the key for the identity (%s)"
	CKUOnlyForDedicatedErrorMsg                      = "specifying `--cku` flag is valid only for dedicated Kafka cluster creation"
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
	InvalidTypeFlagSuggestions                       = "Allowed values for `--type` flag are: %s, %s, %s."
	NameOrCKUFlagErrorMsg                            = "must either specify --name with non-empty value or --cku (for dedicated clusters) with positive integer"
	NonEmptyNameErrorMsg                             = "`--name` flag value must not be empty"
	KafkaClusterNotFoundErrorMsg                     = `Kafka cluster "%s" not found`
	KafkaClusterStillProvisioningErrorMsg            = "your cluster is still provisioning, so it can't be updated yet; please retry in a few minutes"
	KafkaClusterUpdateFailedSuggestions              = "A cluster can't be updated while still provisioning. If you just created this cluster, retry in a few minutes."
	KafkaClusterExpandingErrorMsg                    = "your cluster is expanding; please wait for that operation to complete before updating again"
	KafkaClusterShrinkingErrorMsg                    = "your cluster is shrinking; Please wait for that operation to complete before updating again"
	KafkaClusterInaccessibleErrorMsg                 = `Kafka cluster "%s" not found or access forbidden`
	KafkaClusterInaccessibleSuggestions              = ChooseRightEnvironmentSuggestions + "\n" +
		"The active Kafka cluster may have been deleted. Set a new active cluster with `confluent kafka cluster use`."
	KafkaClusterDeletingSuggestions = KafkaClusterInaccessibleSuggestions + "\n" +
		"Ensure the cluster is not associated with any active Connect clusters."
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
	FailedToFindSchemaIDErrorMsg         = "failed to find schema ID in topic data"
	MissingKeyErrorMsg                   = "missing key in message"
	UnknownValueFormatErrorMsg           = "unknown value schema format"
	TopicExistsErrorMsg                  = `topic "%s" already exists for Kafka cluster "%s"`
	TopicExistsSuggestions               = ListTopicSuggestions
	NoAPISecretStoredOrPassedErrorMsg    = `no API secret for API key "%s" of resource "%s" passed via flag or stored in local CLI state`
	NoAPISecretStoredOrPassedSuggestions = "Pass the API secret with flag \"--api-secret\" or store with `confluent api-key store %s --resource %s`."
	PassedSecretButNotKeyErrorMsg        = "no API key specified"
	PassedSecretButNotKeySuggestions     = `Use the "api-key" flag to specify an API key.`
	ProducingToCompactedTopicErrorMsg    = "producer has detected an INVALID_RECORD error for topic %s"
	ProducingToCompactedTopicSuggestions = "If the topic has schema validation enabled, ensure you are producing with a schema-enabled producer.\n" +
		"If your topic is compacted, ensure you are producing a record with a key."
	FailedToLoadSchemaSuggestions   = "Specify a schema by passing the path to a schema file to the `--schema` flag, or by passing a registered schema ID to the `--schema-id` flag."
	ExceedPartitionLimitSuggestions = "The total partition limit for a dedicated cluster may be increased by expanding its CKU count using `confluent kafka cluster update <id> --cku <count>`."

	// Cluster Link commands
	EmptyConfigErrorMsg = "config file name is empty or config file is empty"

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
	NoLogFoundErrorMsg       = `no log found: to run %s, use "confluent local services %s start"`
	MacVersionErrorMsg       = "macOS version >= %s is required (detected: %s)"
	JavaExecNotFondErrorMsg  = "could not find java executable, please install java or set JAVA_HOME"
	NothingToDestroyErrorMsg = "nothing to destroy"

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
	NoSubjectLevelConfigErrorMsg = `subject "%s" does not have subject-level compatibility configured`
	SRInvalidPackageTypeErrorMsg = `"%s" is an invalid package type`
	SRInvalidPackageSuggestions  = "Allowed values for `--package` flag are: %s."
	SRInvalidPackageUpgrade      = "Environment \"%s\" is already using the Stream Governance \"%s\" package.\n"

	// secret commands
	EnterInputTypeErrorMsg    = "enter %s"
	PipeInputTypeErrorMsg     = "pipe %s over stdin"
	SpecifyPassphraseErrorMsg = "specify `--passphrase -` if you intend to pipe your passphrase over stdin"
	PipePassphraseErrorMsg    = "pipe your passphrase over stdin"

	// update command
	UpdateClientFailurePrefix      = "update client failure"
	UpdateClientFailureSuggestions = "Please submit a support ticket.\n" +
		"In the meantime, see link for other ways to download the latest CLI version:\n" +
		"https://docs.confluent.io/current/cli/installing.html ."
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
	CreateNetrcFileErrorMsg          = `unable to create netrc file "%s"`
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
	EnvironmentNotFoundErrorMsg = `environment "%s" not found in context "%s"`
	MalformedJWTNoExprErrorMsg  = "malformed JWT claims: no expiration"

	// config package
	CorruptedConfigErrorPrefix = "corrupted CLI config"
	CorruptedConfigSuggestions = `Your CLI config file "%s" is corrupted.\n` +
		"Remove config file, and run `confluent login` or `confluent context create`.\n" +
		"Unfortunately, your active CLI state will be lost as a result.\n" +
		"Please file a support ticket with details about your config file to help us address this issue.\n" +
		"Please rerun the command with the verbosity flag `-vvvv` and attach the output with the support ticket."
	UnableToCreateConfigErrorMsg       = "unable to create config"
	UnableToReadConfigErrorMsg         = `unable to read config file "%s"`
	ConfigNotUpToDateErrorMsg          = "config version v%s not up to date with the latest version v%s"
	InvalidConfigVersionErrorMsg       = "invalid config version v%s"
	ParseConfigErrorMsg                = `unable to parse config file "%s"`
	NoNameContextErrorMsg              = "one of the existing contexts has no name"
	MissingKafkaClusterContextErrorMsg = `context "%s" missing KafkaClusterContext`
	MarshalConfigErrorMsg              = "unable to marshal config"
	CreateConfigDirectoryErrorMsg      = "unable to create config directory: %s"
	CreateConfigFileErrorMsg           = "unable to write config to file: %s"
	CurrentContextNotExistErrorMsg     = `the current context "%s" does not exist`
	ContextDoesNotExistErrorMsg        = `context "%s" does not exist`
	ContextAlreadyExistsErrorMsg       = `context "%s" already exists`
	CredentialNotFoundErrorMsg         = `credential "%s" not found`
	PlatformNotFoundErrorMsg           = `platform "%s" not found`
	NoNameCredentialErrorMsg           = "credential must have a name"
	SavedCredentialNoContextErrorMsg   = "saved credential must match a context"
	KeychainNotAvailableErrorMsg       = "keychain not available on platforms other than darwin"
	NoValidKeychainCredentialErrorMsg  = "no matching credentials found in keychain"
	NoNamePlatformErrorMsg             = "platform must have a name"
	UnspecifiedPlatformErrorMsg        = `context "%s" has corrupted platform`
	UnspecifiedCredentialErrorMsg      = `context "%s" has corrupted credentials`
	ContextStateMismatchErrorMsg       = `context state mismatch for context "%s"`
	ContextStateNotMappedErrorMsg      = `context state mapping error for context "%s"`
	DeleteUserAuthErrorMsg             = "unable to delete user auth"

	// local package
	ConfluentHomeNotFoundErrorMsg         = "could not find %s in CONFLUENT_HOME"
	SetConfluentHomeErrorMsg              = "set environment variable CONFLUENT_HOME"
	KafkaScriptFormatNotSupportedErrorMsg = "format %s is not supported in this version"
	KafkaScriptInvalidFormatErrorMsg      = "invalid format: %s"

	// secret package
	EncryptPlainTextErrorMsg           = "failed to encrypt the plain text"
	DecryptCypherErrorMsg              = "failed to decrypt the cipher"
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

	// catcher
	CCloudBackendErrorPrefix           = "Confluent Cloud backend error"
	UnexpectedBackendOutputPrefix      = "unexpected CCloud backend output"
	UnexpectedBackendOutputSuggestions = "Please submit a support ticket."
	BackendUnmarshallingErrorMsg       = "protobuf unmarshalling error"
	ResourceNotFoundErrorMsg           = `resource "%s" not found`
	ResourceNotFoundSuggestions        = "Check that the resource \"%s\" exists.\n" +
		"To list Kafka clusters, use `confluent kafka cluster list`.\n" +
		"To check schema-registry cluster info, use `confluent schema-registry cluster describe`.\n" +
		"To list KSQL clusters, use `confluent ksql cluster list`."
	KafkaNotFoundErrorMsg         = `Kafka cluster "%s" not found`
	KafkaNotFoundSuggestions      = "To list Kafka clusters, use `confluent kafka cluster list`."
	KSQLNotFoundSuggestions       = "To list KSQL clusters, use `confluent ksql cluster list`."
	KafkaNotReadyErrorMsg         = `Kafka cluster "%s" not ready`
	KafkaNotReadySuggestions      = "It may take up to 5 minutes for a recently created Kafka cluster to be ready."
	NoKafkaSelectedErrorMsg       = "no Kafka cluster selected"
	NoKafkaSelectedSuggestions    = "You must pass `--cluster` with the command or set an active Kafka cluster in your context with `confluent kafka cluster use`."
	NoKafkaForDescribeSuggestions = "You must provide the cluster ID argument or set an active Kafka cluster in your context with `ccloud kafka cluster use`."
	NoAPISecretStoredErrorMsg     = `no API secret for API key "%s" of resource "%s" stored in local CLI state`
	NoAPISecretStoredSuggestions  = "Store the API secret with `confluent api-key store %s --resource %s`."
	InvalidCkuErrorMsg            = "cku must be greater than 1 for multi-zone dedicated cluster"

	// Kafka REST Proxy errors
	InternalServerErrorMsg            = "internal server error"
	UnknownErrorMsg                   = "unknown error"
	InternalServerErrorSuggestions    = "Please check the status of your Kafka cluster or submit a support ticket."
	EmptyResponseErrorMsg             = "empty server response"
	KafkaRestErrorMsg                 = "Kafka REST request failed: %s %s: %s"
	KafkaRestConnectionErrorMsg       = "unable to establish Kafka REST connection: %s: %s"
	KafkaRestUnexpectedStatusErrorMsg = "Kafka REST request failed: %s: unexpected HTTP Status: %d"
	KafkaRestCertErrorSuggestions     = `To specify a CA certificate, please use the "ca-cert-path" flag or set "CONFLUENT_PLATFORM_CA_CERT_PATH".`
	KafkaRestUrlNotFoundErrorMsg      = "Kafka REST URL not found"
	KafkaRestUrlNotFoundSuggestions   = "Use the `--url` flag or set CONFLUENT_REST_URL."
	KafkaRestProvisioningErrorMsg     = `Kafka REST unavailable: Kafka cluster "%s" is still provisioning`
	NoClustersFoundErrorMsg           = "no clusters found"
	NoClustersFoundSuggestions        = "Please check the status of your cluster and the Kafka REST bootstrap.servers configuration."
	NeedClientCertAndKeyPathsErrorMsg = `must set "client-cert-path" and "client-key-path" flags together`
	InvalidMDSTokenErrorMsg           = "Invalid MDS token"
	InvalidMDSTokenSuggestions        = `Re-login with "confluent login".`

	// Special error handling
	QuotaExceededSuggestions = `Look up Confluent Cloud service quota limits with "confluent service-quota list".`
	AvoidTimeoutSuggestions  = "To avoid session timeouts, non-SSO users can save their credentials to the netrc file with `confluent login --save`."
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
	NoAPIKeySelectedErrorMsg         = `no API key selected for resource "%s"`
	NoAPIKeySelectedSuggestions      = "Select an API key for resource \"%s\" with `confluent api-key use <API_KEY> --resource %s`.\n" +
		"To do so, you must have either already created or stored an API key for the resource.\n" +
		"To create an API key, use `confluent api-key create --resource %s`.\n" +
		"To store an existing API key, use `confluent api-key store --resource %s`."
	FailedToReadInputErrorMsg = "failed to read input"

	// Flag parsing errors
	EnvironmentFlagWithApiLoginErrorMsg = `"environment" flag should not be passed for API key context`
	ClusterFlagWithApiLoginErrorMsg     = `"cluster" flag should not be passed for API key context, cluster is inferred`

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
	DeleteResourceErrorMsg        = `failed to delete %s "%s": %v`
	DeleteResourceConfirmErrorMsg = `input does not match "%s"`
	UpdateResourceErrorMsg        = `failed to update %s "%s": %v`
	MustSpecifyBothFlagsErrorMsg  = "must specify both `--%s` and `--%s`"
)
