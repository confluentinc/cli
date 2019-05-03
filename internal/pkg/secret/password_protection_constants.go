package secret

/* Config Provider Configs*/
const (
	ENCRYPTION_CONFIG_PATH = "encryptionConfigPath"

	CONFIG_PROVIDER_KEY = "config.providers"

	SECURE_CONFIG_PROVIDER_CLASS_KEY = "config.providers.securePass.class"

	SECURE_CONFIG_PROVIDER = "securePass"

	SECURE_CONFIG_PROVIDER_CLASS = "org.apache.kafka.common.config.provider.SecurePassConfigProvider"

	SECURE_CONFIG_PROVIDER_ARG1 = "config.providers.securePass.encryptionConfigPath"
)

/* Encryption Keys Metadata */
const (
	METADATA_KEY_PATH = "metadata.symmetric_key.0.file"

	METADATA_KEY_ENVVAR = "metadata.symmetric_key.0.envvar"

	METADATA_KEY_TIMESTAMP = "metadata.symmetric_key.0.created_at"

	METADATA_KEY_LENGTH = "metadata.symmetric_key.0.length"

	METADATA_KEY_SALT = "metadata.symmetric_key.0.salt"

	METADATA_KEY_ITERATIONS = "metadata.symmetric_key.0.iterations"

	METADATA_DATA_KEY = "metadata.symmetric_key.0.enc"

	METADATA_KEY_DEFAULT_SALT = "727155B85D4E2C207F1BBA12681A5D5F"

	METADATA_KEY_DEFAULT_LENGTH_BYTES = 32

	METADATA_KEY_DEFAULT_ITERATIONS = 1000

)

/* Encryption Algorithm Metadata */
const (
	METADATA_ENC_ALGORITHM = "AES/CBC/PKCS5Padding"
)

/* Password Protection File Metadata */
const (
	SECURITY_DIR_PATH = "/etc/secure/"

	CONFLUENT_MASTER_KEY_FILE_PATH =  SECURITY_DIR_PATH + "confluent_master_key.properties"

	SECURE_CONFIG_FILE_PATH = SECURITY_DIR_PATH + "confluent_secure_config.properties"

	CONFLUENT_HOME = "CONFLUENT_HOME"

	CONFLUENT_KEY_ENVVAR = "CONFLUENT_SECURITY_MASTER_KEY"
)
