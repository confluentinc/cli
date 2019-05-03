package secret

import (
	"bytes"
	"fmt"
	"io/ioutil"

	"os"
	"strconv"
	"strings"
	"time"
	"regexp"

	"github.com/magiconair/properties"
	"github.com/confluentinc/cli/internal/pkg/log"
)

/**
* Password Protection is a security plugin to securely store and add passwords to a properties file.
* Passwords in property file are encrypted and stored in security config file.
*/

type PasswordProtection interface {
	GenerateEncryptionKeys (passphrase string, pathType string, path string) error
	EncryptConfigFileSecrets (configFilePath string) error
	DecryptConfigFileSecrets (configFilePath string, outputFilePath string) error
	AddEncryptedPasswords (configFilePath string, modifiedFilePath string) error
	RotateMasterKey (oldPassphrase string, newPassphrase string) error
	RotateDataKey (masterPassphrase string) error
}

type PasswordProtectionSuite struct {
    SecureConfigFilePath string
	PasswordProtectionDir string
	Logger *log.Logger
}

func NewPasswordProtectionPlugin(logger *log.Logger) *PasswordProtectionSuite {
	// Set Confluent and Password Protection Home Directory
	confluentHome := os.Getenv(CONFLUENT_HOME)
	configFilePath :=  confluentHome + SECURE_CONFIG_FILE_PATH
	ppDir := confluentHome + SECURITY_DIR_PATH
	return &PasswordProtectionSuite{SecureConfigFilePath: configFilePath, PasswordProtectionDir: ppDir, Logger: logger}
}

/**
 * This function generates a new data key for encryption/decryption of secrets. The DEK is wrapped using the master key and saved in the secrets file
 * along with other metadata.
 *
 * @param passphrase New Master Key passphrase.
 * @param pathType path type for storing the master key: environment-variable/confluent-home/user-defined
 * @param path user defined path for storing the master key. Empty string for environment-variable and confluent-home
 * @return error: Failure nil: Success
 */
func(c *PasswordProtectionSuite) GenerateEncryptionKeys(passphrase string, pathType string, path string) error {
	// Verify if master key is not previously set.
	if c.isMasterKeySet() {
		return fmt.Errorf("master key is already set, you can change the key by rotate-key command")
	}

	// Create password protection home dir.
	if _, err := os.Stat(c.PasswordProtectionDir); os.IsNotExist(err) {
		err := os.MkdirAll(c.PasswordProtectionDir, os.ModePerm)
		if err != nil {
			return err
		}
	}

	// Save the master key based on path type.
	switch pathType {
	case "environment-variable":
		// Make sure Master Key is set in environment variable.
		masterKey := os.Getenv(CONFLUENT_KEY_ENVVAR)

		if len(masterKey) == 0 {
			return fmt.Errorf("master key not set in environment variable. set the key and execute this command again.")
		}

		if strings.Compare(masterKey, passphrase) != 0 {
			return fmt.Errorf("master key does not match the master key set in environment variable.")
		}
	case "confluent-home":
		// TO-DO Decide on default path for saving the secret file and secret
		path = os.Getenv(CONFLUENT_HOME) + CONFLUENT_MASTER_KEY_FILE_PATH
		// Save the master key in confluent home dir.
		masterKeyFileErr := c.generateMasterKeyFile(passphrase, path)
		if masterKeyFileErr != nil {
			return masterKeyFileErr
		}
	case "user-defined":
		// Save the master key in user defined path.
		masterKeyFileErr := c.generateMasterKeyFile(passphrase, path)
		if masterKeyFileErr != nil {
			return masterKeyFileErr
		}
	default:
		return fmt.Errorf("Invalid Option. Please choose from environment-variable/confluent-home/user-defined")
	}

	// Generate the metadata for encryption keys
	cipherSuite := NewDefaultCipherSuite()
	cipherSuite.SetMasterKeyPath(path)
	engine := NewEncryptionEngine(cipherSuite, c.Logger)

	// Generate a new data key. This data key will be used for encrypting the secrets.
	dataKey, err := engine.GenerateRandomDataKey(METADATA_KEY_DEFAULT_LENGTH_BYTES)
	if err != nil {
		return err
	}

	// Wrap data key with master key
	encodedDataKey, err := c.wrapDataKey(engine, dataKey, passphrase)
	if err != nil {
		return err
	}
	cipherSuite.SetDataKey(encodedDataKey)

	// Generate the secrets file and save the metadata for encryption key.
	secureConfigErr := c.generateSecureConfigFile(cipherSuite)

	return secureConfigErr
}

/**
 * This function encrypts all the passwords in configFilePath properties files. It searches for keys with keyword 'password'
 * in properties file configFilePath and encrypts the password using the encryption engine. The original password in the
 * properties file is replaced with tuple ${providerName:[path:]key} and encrypted password is stored in secureConfigPath
 * properties file with key as configFilePath:key and value as encrypted password.
 * We also add the properties to instantiate the SecurePass provider to the config properties file.
 *
 * @param configFilePath properties file path whose passwords need to be encrypted.
 * @return error: Failure nil: Success
 */
func(c *PasswordProtectionSuite) EncryptConfigFileSecrets(configFilePath string) error {
	// Check if config file path is valid.
	if !c.isPathValid(configFilePath) {
		return fmt.Errorf("Invalid File Path" + configFilePath)
	}

	// Load the configs.
	configProps, err := c.loadPropertiesFile(configFilePath)
	if err != nil {
		return err
	}

	// Filter the properties which have keyword 'password' in the key.
	//To-do should we accept the list of keywords from user?
	matchProps, err := configProps.Filter("(?i)password")
	if err != nil {
		return err
	}
	secureConfigProps, err := c.loadPropertiesFile(c.SecureConfigFilePath)
	if err != nil {
		return err
	}

	// Encrypt the secrets with DEK. Save the encrypted secrets in secure config file.
	err = c.encryptConfigValues(matchProps, secureConfigProps, configProps, configFilePath)

	return err
}

/**
 * This function decrypts all the passwords in configFilePath properties files and stores the decrypted passwords in outputFilePath.
 * It searches for the encrypted secrets by comparing it with the tuple ${providerName:[path:]key}. If encrypted secrets are found it fetches
 * the encrypted value from the file secureConfigPath, decrypts it using the data key and stores the output at outputFilePath.
 *
 * @param configFilePath properties file path whose passwords need to be encrypted.
 * @param outputFilePath properties file path where encrypted passwords are stored.
 * @return error: Failure nil: Success
 */
func(c *PasswordProtectionSuite) DecryptConfigFileSecrets(configFilePath string, outputFilePath string) error {
	// Check if config file path is valid
	if !c.isPathValid(configFilePath) {
		return fmt.Errorf("Invalid File Path" + configFilePath)
	}

	// Load the config values.
	configProps, err := c.loadPropertiesFile(configFilePath)
	if err != nil {
		return err
	}

	// Load the encrypted config value.
	secureConfigProps, err := c.loadPropertiesFile(c.SecureConfigFilePath)
	if err != nil {
		return err
	}
	decryptedSecrets := properties.NewProperties()
	cipherSuite, err := c.loadCipherSuite()
	if err != nil {
		return err
	}
	engine := NewEncryptionEngine(&cipherSuite, c.Logger)
	// Unwrap DEK with MEK
	dataKey, err := c.unwrapDataKey(cipherSuite.EncryptedDataKey, cipherSuite.MasterKeyPath, engine)
	if err != nil {
		return err
	}

	for key, value := range configProps.Map() {
		// If config value is encrypted, decrypt it with DEK.
		if c.isPasswordEncrypted(value) {
			pathKey := configFilePath + ":" + key
			cipher := secureConfigProps.GetString(pathKey, "")
			if cipher != "" {
				data, iv, algo := c.parseCipherValue(cipher)
				plainSecret, err := engine.AESDecrypt(data, iv, algo, dataKey)
				if err == nil {
					_,_,_ = decryptedSecrets.Set(key, plainSecret)
				}
			}
		}

	}

	// Save the decrypted ciphers to output file.
	err = c.writePropertiesFile(outputFilePath, decryptedSecrets)
	if err != nil {
		return err
	}
	return nil
}

/**
 * This function generates a new data key and re-encrypts the values in the secureConfigPath properties file with the new data key.
 *
 * @param masterPassphrase master key passphrase for verification
 * @return Failure nil: Success
 */
func(c *PasswordProtectionSuite) RotateDataKey(masterPassphrase string) error {
	cipherSuite, err := c.loadCipherSuite()
	if err != nil {
		return err
	}

	// Load MEK
	masterKey, err := c.loadMasterKey(cipherSuite.MasterKeyPath)
	if err != nil {
		return err
	}

	// Verify master key passphrase
	if masterKey != masterPassphrase {
		return fmt.Errorf("Authentication Failure: Invalid master key passphrase.")
	}

	secureConfigProps, err := c.loadPropertiesFile(c.SecureConfigFilePath)
	if err != nil {
		return err
	}

	// Unwrap old DEK using the MEK
	rotatedSecrets := properties.NewProperties()
	engine := NewEncryptionEngine(&cipherSuite, c.Logger)
	dataKey, err := c.unwrapDataKey(cipherSuite.EncryptedDataKey, cipherSuite.MasterKeyPath, engine)
	if err != nil {
		return err
	}

	// Generate a new DEK
	newDataKey, err := engine.GenerateRandomDataKey(METADATA_KEY_DEFAULT_LENGTH_BYTES)
	if err != nil {
		return err
	}

	// Re-encrypt the ciphers with new DEK
	for key, value := range secureConfigProps.Map() {
		if c.isCipher(value) {
			data, iv, algo := c.parseCipherValue(value)
			plainSecret, err := engine.AESDecrypt(data, iv, algo, dataKey)
			if err == nil {
				cipher, iv, err := engine.AESEncrypt(plainSecret, newDataKey)
				if err == nil {
					formattedCipher := c.formatCipherValue(cipher, iv)
					_,_,_ = rotatedSecrets.Set(key, formattedCipher)
				}
			}

		}

	}

	// Wrap new DEK with MEK
	wrappedNewDK, err := c.wrapDataKey(engine, newDataKey, masterKey)
	if err != nil {
		return err
	}

	// Save new DEK and re-encrypted ciphers.
	now := time.Now()
	_,_,_ = secureConfigProps.Set(METADATA_KEY_TIMESTAMP, now.String())
	_,_,_ = secureConfigProps.Set(METADATA_DATA_KEY, wrappedNewDK)
	err = c.writePropertiesFile(c.SecureConfigFilePath, rotatedSecrets)
	if err != nil {
		return err
	}

	return nil
}

/**
 * This function is used to change the master key. It wraps the data key with newly set master key.
 *
 * @param masterPassphrase master key passphrase for verification
 * @param newPassphrase new master key passphrase
 * @return Failure nil: Success
 */
func(c *PasswordProtectionSuite) RotateMasterKey(oldPassphrase string, newPassphrase string) error {

	cipherSuite, err := c.loadCipherSuite()
	if err != nil {
		return err
	}

	// Load MEK
	masterKey, err := c.loadMasterKey(cipherSuite.MasterKeyPath)
	if err != nil {
		return err
	}

	// Verify master key passphrase
	if masterKey != oldPassphrase {
		return fmt.Errorf("Authentication Failure: Invalid master key passphrase.")
	}

	// Unwrap DEK using the MEK
	engine := NewEncryptionEngine(&cipherSuite, c.Logger)
	dataKey, err := c.unwrapDataKey(cipherSuite.EncryptedDataKey, cipherSuite.MasterKeyPath, engine)
	if err != nil {
		return err
	}

	// Wrap DEK using the new MEK
	wrappedDataKey, iv, err := engine.WrapDataKey(dataKey, newPassphrase)
	if err != nil {
		return err
	}
	newEncodedDataKey := c.formatCipherValue(wrappedDataKey, iv)


	secureConfigProps, err := c.loadPropertiesFile(c.SecureConfigFilePath)
	if err != nil {
		return err
	}

	// Save DEK
	now := time.Now()
	_,_,_ = secureConfigProps.Set(METADATA_KEY_TIMESTAMP, now.String())
	_,_,_ = secureConfigProps.Set(METADATA_DATA_KEY, newEncodedDataKey)

	// To-do: Can we make following 2 operations Atomic
	if cipherSuite.MasterKeyPath != "" {
		masterKeyFileErr := c.generateMasterKeyFile(newPassphrase, cipherSuite.MasterKeyPath)
		if masterKeyFileErr != nil {
			return masterKeyFileErr
		}
	}

	err = c.writePropertiesFile(c.SecureConfigFilePath, secureConfigProps)

	return err
}

/* To-do Make it Edit */
/**
 * This function adds a new key value pair to the configFilePath property file. The original 'value' is
 * encrypted value and stored in the secureConfigPath properties file with key as
 * configFilePath:key and value as encrypted password.
 * We also add the properties to instantiate the SecurePass provider to the config properties file.
 *
 * @param configFilePath properties file path where new password needs to be added.
 * @param modifiedFilePath properties file path with newly added passwords.
 * @return
 */
func(c *PasswordProtectionSuite) AddEncryptedPasswords(configFilePath string, modifiedFilePath string) error {
	newConfigProps := properties.MustLoadString(modifiedFilePath)
	configProps, err := c.loadPropertiesFile(configFilePath)
	if err != nil {
		return err
	}

	secureConfigProps, err := c.loadPropertiesFile(c.SecureConfigFilePath)
	if err != nil {
		return err
	}
	err = c.encryptConfigValues(newConfigProps, secureConfigProps, configProps, configFilePath)

	return err
}

/** Helper Functions **/

func (c *PasswordProtectionSuite) wrapDataKey(engine EncryptionEngine, dataKey []byte, masterKey string) (string, error) {
	wrappedDataKey, iv, err := engine.WrapDataKey(dataKey, masterKey)
	if err != nil {
		return "", err
	}

	encodedDataKey := c.formatCipherValue(wrappedDataKey, iv)

	return encodedDataKey, nil
}

func (c *PasswordProtectionSuite) loadCipherSuite() (CipherSuite, error) {
	var cipher CipherSuite
	secureConfigProps, err := c.loadPropertiesFile(c.SecureConfigFilePath)
	if err != nil {
		return cipher, err
	}
	matchProps, err := secureConfigProps.Filter("(?i)metadata")
	if err != nil {
		return cipher, err
	}
	cipher.Iterations = matchProps.GetInt(METADATA_KEY_ITERATIONS, METADATA_KEY_DEFAULT_ITERATIONS)
	cipher.KeyLength = matchProps.GetInt(METADATA_KEY_LENGTH, METADATA_KEY_DEFAULT_LENGTH_BYTES)
	cipher.MasterKeyPath = matchProps.GetString(METADATA_KEY_PATH, "")
	cipher.Salt = matchProps.GetString(METADATA_KEY_SALT, METADATA_KEY_DEFAULT_SALT)
	cipher.EncryptedDataKey = matchProps.GetString(METADATA_DATA_KEY, "")
	return cipher, nil
}

func (c *PasswordProtectionSuite) isMasterKeySet() bool {
	if _, err := os.Stat(c.SecureConfigFilePath); os.IsExist(err) {
		return true
	}
	return false
}

func (c *PasswordProtectionSuite) isPathValid(path string) bool {

	if path == "" {
		return false
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func(c *PasswordProtectionSuite) writePropertiesFile(path string, property *properties.Properties) error {
	buf := new(bytes.Buffer)
	_, _ = property.WriteComment(buf, "", properties.ISO_8859_1)
	err := ioutil.WriteFile(path, buf.Bytes(), 0644)
	return err
}

func(c *PasswordProtectionSuite) loadPropertiesFile(path string) (*properties.Properties, error) {
	if !c.isPathValid(path) {
		return properties.NewProperties(), fmt.Errorf("Invalid file path.")
	}

	property := properties.MustLoadFile(path, properties.ISO_8859_1)
	return property, nil
}

func(c *PasswordProtectionSuite) generateSecureConfigFile(suite *CipherSuite) error {
	// Check if File exists
	if _, err := os.Stat(c.SecureConfigFilePath); os.IsExist(err) {
		return fmt.Errorf("secure config file already exists.")
	}

	secureConfigProps := properties.NewProperties()
	now := time.Now()
	_,_,_ = secureConfigProps.Set(METADATA_KEY_TIMESTAMP, now.String())

	if suite.MasterKeyPath == "" {
		_,_,_ = secureConfigProps.Set(METADATA_KEY_ENVVAR, CONFLUENT_KEY_ENVVAR)
	} else {
		_,_,_ = secureConfigProps.Set(METADATA_KEY_PATH, suite.MasterKeyPath)
	}

	_,_,_ = secureConfigProps.Set(METADATA_KEY_LENGTH, strconv.Itoa(suite.KeyLength))
	_,_,_ = secureConfigProps.Set(METADATA_KEY_ITERATIONS, strconv.Itoa(suite.Iterations))
	_,_,_ = secureConfigProps.Set(METADATA_KEY_SALT, suite.Salt)
	_,_,_ = secureConfigProps.Set(METADATA_DATA_KEY, suite.EncryptedDataKey)

	if _, err := os.Stat(c.SecureConfigFilePath); os.IsExist(err) {
		return fmt.Errorf("secure config file already exists.")
	}
	err := c.writePropertiesFile(c.SecureConfigFilePath, secureConfigProps)
	return err
}

func(c *PasswordProtectionSuite) generateMasterKeyFile(passphrase string, path string) error {
	masterKeyProps := properties.NewProperties()
	_,_,_ = masterKeyProps.Set(CONFLUENT_KEY_ENVVAR, passphrase)
	err := c.writePropertiesFile(path, masterKeyProps)

	return err
}

func(c *PasswordProtectionSuite) isPasswordEncrypted(config string) bool {
	regex, _ := regexp.Compile("\\$\\{(.*?):((.*?):)?(.*?)\\}")
	return regex.MatchString(config)
}

func(c *PasswordProtectionSuite) generateConfigValue(key string, path string) string {
	return "${securePass:" + path + ":" + key + "}";
}

func(c *PasswordProtectionSuite) formatCipherValue(cipher string, iv string) string {
	return "ENC[" + METADATA_ENC_ALGORITHM + ",data:" + cipher + ",iv:" + iv + ",type:str]";
}

func(c *PasswordProtectionSuite) isCipher(config string) bool {
	regex, _ := regexp.Compile("ENC\\[(.*?)\\]")
	return regex.MatchString(config)
}

func(c *PasswordProtectionSuite) addSecureConfigProviderProperty(property *properties.Properties) *properties.Properties {
	configProviders := property.GetString(CONFIG_PROVIDER_KEY, "")
	if configProviders == "" {
		configProviders = SECURE_CONFIG_PROVIDER
	} else if strings.Contains(configProviders, SECURE_CONFIG_PROVIDER) {
		configProviders = configProviders + "," + SECURE_CONFIG_PROVIDER
	}

	_,_,_ = property.Set(CONFIG_PROVIDER_KEY, configProviders);
	_,_,_ = property.Set(SECURE_CONFIG_PROVIDER_CLASS_KEY, SECURE_CONFIG_PROVIDER_CLASS);
	return property
}

func(c *PasswordProtectionSuite) unwrapDataKey(key string, masterKeyPath string, engine EncryptionEngine) ([]byte, error) {
	masterKey, err := c.loadMasterKey(masterKeyPath)
	if err != nil {
		return []byte{}, err
	}
	data, iv, algo := c.parseCipherValue(key)
	return engine.UnWrapDataKey(data, iv, algo, masterKey)
}

func(c *PasswordProtectionSuite) encryptConfigValues(matchProps *properties.Properties, secureConfigProps *properties.Properties, configProps *properties.Properties,
	configFilePath string) error {

	cipherSuite, err := c.loadCipherSuite()
	if err != nil {
		return err
	}
	engine := NewEncryptionEngine(&cipherSuite, c.Logger)
	dataKey, err := c.unwrapDataKey(cipherSuite.EncryptedDataKey, cipherSuite.MasterKeyPath, engine)
	if err != nil {
		return err
	}

	for key, value := range matchProps.Map() {
		if !c.isPasswordEncrypted(value) {
			// Generate tuple ${providerName:[path:]key}
			newConfigVal := c.generateConfigValue(key, c.SecureConfigFilePath)
			_,_,_ = configProps.Set(key, newConfigVal)
			cipher, iv, err := engine.AESEncrypt(value, dataKey)
			if err == nil {
				formattedCipher := c.formatCipherValue(cipher, iv)
				pathKey := configFilePath + ":" + key
				_,_,_ = secureConfigProps.Set(pathKey, formattedCipher)
			}
		}

	}

	configProps = c.addSecureConfigProviderProperty(configProps)

	err = c.writePropertiesFile(c.SecureConfigFilePath, secureConfigProps)
	if err != nil {
		return err
	}

	err = c.writePropertiesFile(configFilePath, configProps)
	if err != nil {
		return err
	}

	return nil
}

func (c *PasswordProtectionSuite) loadMasterKey(masterKeyPath string) (string, error) {
	masterKey := ""
	if masterKeyPath == "" {
		masterKey = os.Getenv(CONFLUENT_KEY_ENVVAR)
	} else {
		property := properties.MustLoadFile(masterKeyPath, properties.ISO_8859_1)
		masterKey = property.GetString(CONFLUENT_KEY_ENVVAR, "")
	}

	if len(masterKey) == 0 {
		return "", fmt.Errorf("master Key not set")
	}

	return masterKey, nil
}

func(c *PasswordProtectionSuite) findMatchTrim(original string, pattern string, prefix string, suffix string) string {
	re := regexp.MustCompile(pattern)
	match := re.FindStringSubmatch(original)
	substring := ""
	if len(match) != 0 {
		substring = strings.TrimPrefix( strings.TrimSuffix(match[0],suffix), prefix )
	}

	return substring
}

func(c *PasswordProtectionSuite) parseCipherValue(cipher string) (string, string, string) {
	data := c.findMatchTrim(cipher, "data\\:(.*?)\\,", "data:", ",")
	iv := c.findMatchTrim(cipher, "iv\\:(.*?)\\,", "iv:", ",")
	algo := c.findMatchTrim(cipher, "ENC\\[(.*?)\\,", "ENC[", ",")

	return data, iv, algo
}
