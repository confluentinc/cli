package secret

// Config Provider Configs
const (
	CONFIG_PROVIDER_KEY = "config.providers"

	SECURE_CONFIG_PROVIDER_CLASS_KEY = "config.providers.securepass.class"

	SECURE_CONFIG_PROVIDER = "securepass"

	SECURE_CONFIG_PROVIDER_CLASS = "io.confluent.kafka.security.config.provider.SecurePassConfigProvider"

	/* The properties file writer associates comments with the next property, so if the comment is the last line in the config file
	   the comment is deleted after performing encryption/add/update operation on the file. In order to retain the last comment we add delimiter
	   SECURE_CONFIG_PROVIDER_DELIMITER at the end of the config file before performing the encryption/add/update operation. This delimiter is removed
	   after the operations are completed.
	*/
	SECURE_CONFIG_PROVIDER_DELIMITER = "\nconfig.providers.securepass.delimiter = delimiter"
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

	PASSWORD_PATTERN = "\\$\\{(.*?):((.*?):)?(.*?)\\}"

	CIPHER_PATTERN = "ENC\\[(.*?)\\]"
)

// Password Protection File Metadata
const (
	CONFLUENT_KEY_ENVVAR = "CONFLUENT_SECURITY_MASTER_KEY"
)
