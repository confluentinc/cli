package secret

// Config Provider Configs
const (
	ConfigProviderKey            = "config.providers"
	SecureConfigProviderClassKey = "config.providers.securepass.class"
	SecureConfigProvider         = "securepass"
	SecureConfigProviderClass    = "io.confluent.kafka.security.config.provider.SecurePassConfigProvider"
)

// Encryption Keys Metadata
const (
	MetadataKeyEnvvar             = "_metadata.symmetric_key.0.envvar"
	MetadataKeyTimestamp          = "_metadata.symmetric_key.0.created_at"
	MetadataKeyLength             = "_metadata.symmetric_key.0.length"
	MetadataDEKSalt               = "_metadata.symmetric_key.0.salt"
	MetadataMEKSalt               = "_metadata.master_key.0.salt"
	MetadataKeyIterations         = "_metadata.symmetric_key.0.iterations"
	MetadataDataKey               = "_metadata.symmetric_key.0.enc"
	MetadataKeyDefaultLengthBytes = 32
	MetadataKeyDefaultIterations  = 1000
	MetadataPrefix                = "_metadata"
)

const (
	MetadataEncAlgorithm = "AES/CBC/PKCS5Padding"
	DataPattern          = "data\\:(.*?)\\,"
	IVPattern            = "iv\\:(.*?)\\,"
	EncPattern           = "ENC\\[(.*?)\\,"
	PasswordPattern      = "^\\$\\{(.*?):((.*?):)?(.*?)\\}$"
	CipherPattern        = "ENC\\[(.*?)\\]"
)

// Password Protection File Metadata
const (
	ConfluentKeyEnvvar = "CONFLUENT_SECURITY_MASTER_KEY"
)

// JAAS Configuration Const
const (
	JAASValuePattern      = "\\s*?=\\s*?(?P<value>\\S+)"
	JAASKeyPattern        = "(.*?)/(.*?)/(.*?)"
	ControlFlagRequired   = "required"
	ControlFlagRequisite  = "requisite"
	ControlFlagSufficient = "sufficient"
	ControlFlagOptional   = "optional"
	ClassId               = 0
	ParentId              = 1
	KeyId                 = 2
	Delete                = "delete"
	Update                = "update"
	Space                 = " "
	KeySeparator          = "/"
)
