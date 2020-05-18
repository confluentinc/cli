package secret

// Config Provider Configs
const (
	CONFIG_PROVIDER_KEY = "config.providers"

	SECURE_CONFIG_PROVIDER_CLASS_KEY = "config.providers.securepass.class"

	SECURE_CONFIG_PROVIDER = "securepass"

	SECURE_CONFIG_PROVIDER_CLASS = "io.confluent.kafka.security.config.provider.SecurePassConfigProvider"
)

// Encryption Keys Metadata
const (
	METADATA_KEY_ENVVAR = "_metadata.symmetric_key.0.envvar"

	METADATA_KEY_TIMESTAMP = "_metadata.symmetric_key.0.created_at"

	METADATA_KEY_LENGTH = "_metadata.symmetric_key.0.length"

	METADATA_DEK_SALT = "_metadata.symmetric_key.0.salt"

	METADATA_MEK_SALT = "_metadata.master_key.0.salt"

	METADATA_KEY_ITERATIONS = "_metadata.symmetric_key.0.iterations"

	METADATA_DATA_KEY = "_metadata.symmetric_key.0.enc"

	METADATA_KEY_DEFAULT_LENGTH_BYTES = 32

	METADATA_KEY_DEFAULT_ITERATIONS = 1000

	METADATA_PREFIX = "_metadata"
)

const (
	METADATA_ENC_ALGORITHM = "AES/CBC/PKCS5Padding"

	DATA_PATTERN = "data\\:(.*?)\\,"

	IV_PATTERN = "iv\\:(.*?)\\,"

	ENC_PATTERN = "ENC\\[(.*?)\\,"

	PASSWORD_PATTERN = "^\\$\\{(.*?):((.*?):)?(.*?)\\}$"

	CIPHER_PATTERN = "ENC\\[(.*?)\\]"
)

// Password Protection File Metadata
const (
	CONFLUENT_KEY_ENVVAR = "CONFLUENT_SECURITY_MASTER_KEY"
)

// JAAS Configuration Const
const (
	JAAS_VALUE_PATTERN      = "\\s*?=\\s*?(?P<value>\\S+)"
	JAAS_KEY_PATTERN        = "(.*?)/(.*?)/(.*?)"
	CONTROL_FLAG_REQUIRED   = "required"
	CONTROL_FLAG_REQUISITE  = "requisite"
	CONTROL_FLAG_SUFFICIENT = "sufficient"
	CONTROL_FLAG_OPTIONAL   = "optional"
	CLASS_ID                = 0
	PARENT_ID               = 1
	KEY_ID                  = 2
	ADD                     = "add"
	DELETE                  = "delete"
	UPDATE                  = "update"
	SPACE                   = " "
	KEY_SEPARATOR           = "/"
)
