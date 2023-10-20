package errors

/*
	Error message and suggestions message associated with them
*/

const (
	ApiKeyNotFoundSuggestions         = "Ensure the API key exists and has not been deleted, or create a new API key via `confluent api-key create`."
	BadServiceAccountIdErrorMsg       = `failed to parse service account id: ensure service account id begins with "sa-"`
	ByokKeyNotFoundSuggestions        = "Ensure the self-managed key exists and has not been deleted, or register a new key via `confluent byok register`."
	DeleteResourceErrorMsg            = `failed to delete %s "%s": %w`
	EndOfFreeTrialErrorMsg            = `organization "%s" has been suspended because your free trial has ended`
	EndOfFreeTrialSuggestions         = "To continue using Confluent Cloud, please enter a credit card with `confluent admin payment update` or claim a promo code with `confluent admin promo add`. To enter payment via the UI, please go to https://confluent.cloud/login."
	EnsureCpSixPlusSuggestions        = "Ensure that you are running against MDS with CP 6.0+."
	ExactlyOneSetErrorMsg             = "exactly one of %v must be set"
	ListResourceSuggestions           = "List available %ss with `%s list`."
	MoreThanOneNonKafkaErrorMsg       = "cannot specify more than one non-Kafka cluster ID for a scope"
	MustSetAllowOrDenyErrorMsg        = "`--allow` or `--deny` must be set when adding or deleting an ACL"
	MustSetResourceTypeErrorMsg       = "exactly one resource type (%s) must be set"
	ServiceAccountNotFoundErrorMsg    = `service account "%s" not found`
	ServiceAccountNotFoundSuggestions = "List service accounts with `confluent service-account list`."
	SpecifyKafkaIdErrorMsg            = "must specify `--kafka-cluster` to uniquely identify the scope"
	UnknownConnectorIdErrorMsg        = `unknown connector ID "%s"`
	InvalidApiKeyErrorMsg             = `invalid API key "%s" for resource "%s"`
	InvalidApiKeySuggestions          = "To list API keys that belong to resource \"%[1]s\", use `confluent api-key list --resource %[1]s`.\nTo create new API key for resource \"%[1]s\", use `confluent api-key create --resource %[1]s`."

	// kafka cluster commands
	CkuMoreThanZeroErrorMsg                          = "`--cku` value must be greater than 0"
	TopicDoesNotExistOrMissingPermissionsErrorMsg    = `topic "%s" does not exist or user does not have the ACLs or role bindings required to describe it`
	TopicDoesNotExistOrMissingPermissionsSuggestions = "To list topics for Kafka cluster \"%[1]s\", use `confluent kafka topic list --cluster %[1]s`.\nTo list ACLs use `confluent kafka acl list --cluster %[1]s`.\nTo list role bindings use `confluent iam rbac role-binding list`."
	KafkaClusterNotFoundErrorMsg                     = `Kafka cluster "%s" not found`
	ChooseRightEnvironmentSuggestions                = "Ensure the cluster ID you entered is valid.\n" +
		"Ensure the cluster you are specifying belongs to the currently selected environment with `confluent kafka cluster list`, `confluent environment list`, and `confluent environment use`."
	UnknownTopicErrorMsg              = `unknown topic "%s"`
	KafkaClusterMissingPrefixErrorMsg = `Kafka cluster "%s" is missing required prefix "lkc-"`

	// kafka topic commands
	FailedToCreateProducerErrorMsg    = "failed to create producer: %v"
	FailedToCreateConsumerErrorMsg    = "failed to create consumer: %v"
	FailedToCreateAdminClientErrorMsg = "failed to create confluent-kafka-go admin client: %w"
	FailedToProduceErrorMsg           = "failed to produce offset %d: %s\n"
	UnknownValueFormatErrorMsg        = "unknown value schema format"
	ExceedPartitionLimitSuggestions   = "The total partition limit for a dedicated cluster may be increased by expanding its CKU count using `confluent kafka cluster update <id> --cku <count>`."

	// serialization/deserialization commands
	JsonDocumentInvalidErrorMsg       = "the JSON document is invalid"
	AvroReferenceNotSupportedErrorMsg = "avro reference not supported in cloud CLI"
	ProtoSchemaInvalidErrorMsg        = "the protobuf schema is invalid"
	ProtoDocumentInvalidErrorMsg      = "the protobuf document is invalid"

	// ksql commands
	KsqldbNoServiceAccountErrorMsg = `ACLs do not need to be configured for the ksqlDB cluster, "%s", because it was created with user-level access to the Kafka cluster`

	// local commands
	FailedToStartErrorMsg        = "%s failed to start"
	FailedToReadPortsErrorMsg    = "failed to read local ports from config"
	FailedToReadPortsSuggestions = "Restart Confluent Local with `confluent local kafka stop` and `confluent local kafka start`"

	// schema-registry commands
	CompatibilityOrModeErrorMsg = "must pass either `--compatibility` or `--mode` flag"

	// auth package
	NoCredentialsFoundErrorMsg = "no credentials found"
	NoUrlEnvVarErrorMsg        = "no URL env var"
	InvalidInputFormatErrorMsg = `"%s" is not of valid format for field "%s"`

	// config package
	UnableToReadConfigurationFileErrorMsg = `unable to read configuration file "%s": %w`
	NoNameContextErrorMsg                 = "one of the existing contexts has no name"
	ContextDoesNotExistErrorMsg           = `context "%s" does not exist`
	ContextAlreadyExistsErrorMsg          = `context "%s" already exists`
	UnspecifiedPlatformErrorMsg           = `context "%s" has corrupted platform`
	UnspecifiedCredentialErrorMsg         = `context "%s" has corrupted credentials`

	// secret package
	ConfigNotInJaasErrorMsg         = `the configuration "%s" not present in JAAS configuration`
	InvalidJaasConfigErrorMsg       = "invalid JAAS configuration: %s"
	ExpectedConfigNameErrorMsg      = `expected a configuration name but received "%s"`
	LoginModuleControlFlagErrorMsg  = "login module control flag is not specified"
	EmptyPassphraseErrorMsg         = "master key passphrase cannot be empty"
	InvalidConfigFilePathErrorMsg   = `invalid config file path "%s"`
	UnwrapDataKeyErrorMsg           = "failed to unwrap the data key: invalid master key or corrupted data key"
	DecryptConfigErrorMsg           = `failed to decrypt config "%s": corrupted data`
	IncorrectPassphraseErrorMsg     = "authentication failure: incorrect master key passphrase"
	SamePassphraseErrorMsg          = "new master key passphrase may not be the same as the previous passphrase"
	ConfigKeyNotEncryptedErrorMsg   = `configuration key "%s" is not encrypted`
	ConfigKeyNotInJsonErrorMsg      = `configuration key "%s" not present in JSON configuration file`
	MasterKeyNotExportedErrorMsg    = "master key is not exported in `%s` environment variable"
	MasterKeyNotExportedSuggestions = "Set the environment variable `%s` to the master key and execute this command again."
	ConfigKeyNotPresentErrorMsg     = `configuration key "%s" not present in the configuration file`
	InvalidJsonFileFormatErrorMsg   = "invalid json file format"
	InvalidFilePathErrorMsg         = `invalid file path "%s"`
	UnsupportedFileFormatErrorMsg   = `unsupported file format for file "%s"`
	IncorrectNonceLengthErrorMsg    = `incorrect nonce length from ~/.confluent/config.json passed into encryption`

	// sso package
	BrowserAuthTimedOutErrorMsg    = "timed out while waiting for browser authentication to occur"
	BrowserAuthTimedOutSuggestions = "Try logging in again."
	FmtMissingOauthFieldErrorMsg   = `oauth token response body did not contain field "%s"`

	// update package
	ParseVersionErrorMsg      = "unable to parse %s version %s"
	NoVersionsErrorMsg        = "no versions found"
	GetBinaryVersionsErrorMsg = "unable to get available binary versions"

	// plugin package
	NoVersionFoundErrorMsg = "no version found in plugin manifest"

	// catcher
	ResourceNotFoundErrorMsg    = `resource "%s" not found`
	ResourceNotFoundSuggestions = "Check that the resource \"%s\" exists.\n" +
		"To list Kafka clusters, use `confluent kafka cluster list`.\n" +
		"To check Schema Registry cluster information, use `confluent schema-registry cluster describe`.\n" +
		"To list KSQL clusters, use `confluent ksql cluster list`."
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
	AvoidTimeoutSuggestions = "To avoid session timeouts, non-SSO users can save their credentials with `confluent login --save`."
	NotLoggedInErrorMsg     = "not logged in"
	AuthTokenSuggestions    = "You must be logged in to retrieve an oauthbearer token.\n" +
		"An oauthbearer token is required to authenticate OAUTHBEARER mechanism and Schema Registry."
	OnPremConfigGuideSuggestions   = "See configuration and produce/consume command guide: https://docs.confluent.io/confluent-cli/current/cp-produce-consume.html ."
	SRNotAuthenticatedErrorMsg     = "not logged in, or no Schema Registry endpoint specified"
	SREndpointNotSpecifiedErrorMsg = "no Schema Registry endpoint specified"
	SRClientNotValidatedErrorMsg   = "failed to validate Schema Registry client with token"
	CorruptedTokenErrorMsg         = "corrupted auth token"
	CorruptedTokenSuggestions      = "Please log in again.\n" +
		AvoidTimeoutSuggestions
	ExpiredTokenErrorMsg    = "expired token"
	ExpiredTokenSuggestions = "Your session has timed out, you need to log in again.\n" +
		AvoidTimeoutSuggestions
	InvalidLoginURLErrorMsg      = "invalid URL value, see structure: http(s)://<domain/hostname/ip>:<port>/"
	InvalidLoginErrorMsg         = "incorrect email, password, or organization ID"
	InvalidLoginErrorSuggestions = "To log into an organization other than the default organization, use the `--organization-id` flag.\n" +
		AvoidTimeoutSuggestions
	SuspendedOrganizationSuggestions = "Your organization has been suspended, please contact support if you want to unsuspend it."
	FailedToReadInputErrorMsg        = "failed to read input"

	// Special error types
	GenericOpenApiErrorMsg = "metadata service backend error: %s: %s"

	// Network commands
	CorruptedNetworkResponseErrorMsg = "corrupted %s in response"
)
