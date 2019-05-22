package secret

import (
	"bytes"
	"fmt"
	"io/ioutil"

	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/magiconair/properties"
)

/**
* Password Protection is a security plugin to securely store and add passwords to a properties file.
* Passwords in property file are encrypted and stored in security config file.
 */

type PasswordProtection interface {
	CreateMasterKey(passphrase string) (string, error)
	EncryptConfigFileSecrets(configFilePath string, localSecureConfigPath string, remoteSecureConfigPath string) error
	DecryptConfigFileSecrets(configFilePath string, localSecureConfigPath string, outputFilePath string) error
	AddEncryptedPasswords(configFilePath string, localSecureConfigPath string, remoteSecureConfigPath string, newConfigs string) error
	UpdateEncryptedPasswords(configFilePath string, localSecureConfigPath string, remoteSecureConfigPath string, newConfigs string) error
	RemoveEncryptedPasswords(configFilePath string, localSecureConfigPath string, removeConfigs string) error
	RotateMasterKey(oldPassphrase string, newPassphrase string, localSecureConfigPath string) (string, error)
	RotateDataKey(passphrase string, localSecureConfigPath string) error
}

type PasswordProtectionSuite struct {
	Logger *log.Logger
}

func NewPasswordProtectionPlugin(logger *log.Logger) *PasswordProtectionSuite {
	return &PasswordProtectionSuite{Logger: logger}
}

/**
 * This function generates a new data key for encryption/decryption of secrets. The DEK is wrapped using the master key and saved in the secrets file
 * along with other metadata.
 *
 * @param passphrase New Master Key passphrase.
 * @return error: Failure nil: Success
 */
func (c *PasswordProtectionSuite) CreateMasterKey(passphrase string) (string, error) {

	// Generate the metadata for master key
	cipherSuite := NewDefaultCipherSuite()
	return c.generateMasterKey(passphrase, cipherSuite)
}

func (c *PasswordProtectionSuite) generateMasterKey(passphrase string, cipherSuite *CipherSuite) (string, error) {
	engine := NewEncryptionEngine(cipherSuite, c.Logger)

	// Generate a new master key from passphrase
	masterKey, err := engine.GenerateMasterKey(passphrase)
	if err != nil {
		return "", err
	}

	return masterKey, nil
}

func (c *PasswordProtectionSuite) generateNewDataKey(masterKey string) (*CipherSuite, error) {
	// Generate the metadata for encryption keys
	cipherSuite := NewDefaultCipherSuite()
	engine := NewEncryptionEngine(cipherSuite, c.Logger)

	// Generate a new data key. This data key will be used for encrypting the secrets.
	dataKey, err := engine.GenerateRandomDataKey(METADATA_KEY_DEFAULT_LENGTH_BYTES)
	if err != nil {
		return nil, err
	}

	// Wrap data key with master key
	encodedDataKey, err := c.wrapDataKey(engine, dataKey, masterKey)
	if err != nil {
		return nil, err
	}
	cipherSuite.SetDataKey(encodedDataKey)

	// Generate the secrets file and save the metadata for encryption key.
	//err = c.generateSecureConfigFile(cipherSuite)
	return cipherSuite, err
}

/**
 * This function encrypts all the passwords in configFilePath properties files. It searches for keys with keyword 'password'
 * in properties file configFilePath and encrypts the password using the encryption engine. The original password in the
 * properties file is replaced with tuple ${providerName:[path:]key} and encrypted password is stored in secureConfigPath
 * properties file with key as configFilePath:key and value as encrypted password.
 * We also add the properties to instantiate the SecurePass provider to the config properties file.
 *
 */
func (c *PasswordProtectionSuite) EncryptConfigFileSecrets(configFilePath string, localSecureConfigPath string, remoteSecureConfigPath string) error {

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

	// Encrypt the secrets with DEK. Save the encrypted secrets in secure config file.
	return c.encryptConfigValues(matchProps, localSecureConfigPath, configProps, configFilePath, remoteSecureConfigPath)
}

/**
 * This function decrypts all the passwords in configFilePath properties files and stores the decrypted passwords in outputFilePath.
 * It searches for the encrypted secrets by comparing it with the tuple ${providerName:[path:]key}. If encrypted secrets are found it fetches
 * the encrypted value from the file secureConfigPath, decrypts it using the data key and stores the output at outputFilePath.
 *
 */
func (c *PasswordProtectionSuite) DecryptConfigFileSecrets(configFilePath string, localSecureConfigPath string, outputFilePath string) error {
	// Check if config file path is valid
	if !c.isPathValid(configFilePath) {
		return fmt.Errorf("Invalid File Path" + configFilePath)
	}

	// Check if secure config file path is valid
	if !c.isPathValid(localSecureConfigPath) {
		return fmt.Errorf("Invalid File Path" + localSecureConfigPath)
	}
	// Load the config values.
	configProps, err := c.loadPropertiesFile(configFilePath)
	if err != nil {
		return err
	}

	// Load the encrypted config value.
	secureConfigProps, err := c.loadPropertiesFile(localSecureConfigPath)
	if err != nil {
		return err
	}
	decryptedSecrets := properties.NewProperties()
	cipherSuite, err := c.loadCipherSuiteFromLocalFile(localSecureConfigPath)
	if err != nil {
		return err
	}

	engine := NewEncryptionEngine(&cipherSuite, c.Logger)
	// Unwrap DEK with MEK
	dataKey, err := c.unwrapDataKey(cipherSuite.EncryptedDataKey, engine)
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
				if err != nil {
					return fmt.Errorf("unable to decrypt %s: %s", key, err)
				}
				_, _, _ = decryptedSecrets.Set(key, plainSecret)
			}
		}

	}

	// Save the decrypted ciphers to output file.
	return c.writePropertiesFile(outputFilePath, decryptedSecrets)
}

/**
 * This function generates a new data key and re-encrypts the values in the secureConfigPath properties file with the new data key.
 *
 * @param masterPassphrase master key passphrase for verification
 * @return Failure nil: Success
 */
func (c *PasswordProtectionSuite) RotateDataKey(masterPassphrase string, localSecureConfigPath string) error {
	cipherSuite, err := c.loadCipherSuiteFromLocalFile(localSecureConfigPath)
	if err != nil {
		return err
	}

	// Load MEK
	masterKey, err := c.loadMasterKey()
	if err != nil {
		return err
	}

	userMasterKey, err := c.generateMasterKey(masterPassphrase, &cipherSuite)
	if err != nil {
		return err
	}

	// Verify master key passphrase
	if masterKey != userMasterKey {
		return fmt.Errorf("Authentication Failure: Invalid master key passphrase.")
	}

	secureConfigProps, err := c.loadPropertiesFile(localSecureConfigPath)
	if err != nil {
		return err
	}

	// Unwrap old DEK using the MEK
	engine := NewEncryptionEngine(&cipherSuite, c.Logger)
	dataKey, err := c.unwrapDataKey(cipherSuite.EncryptedDataKey, engine)
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
		if c.isCipher(value) && !strings.HasPrefix(key, "metadata") {
			data, iv, algo := c.parseCipherValue(value)
			plainSecret, err := engine.AESDecrypt(data, iv, algo, dataKey)
			if err != nil {
				return err
			}
			cipher, iv, err := engine.AESEncrypt(plainSecret, newDataKey)
			if err != nil {
				return err
			}
			formattedCipher := c.formatCipherValue(cipher, iv)
			_, _, err = secureConfigProps.Set(key, formattedCipher)
			if err != nil {
				return err
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
	_, _, _ = secureConfigProps.Set(METADATA_KEY_TIMESTAMP, now.String())
	_, _, _ = secureConfigProps.Set(METADATA_DATA_KEY, wrappedNewDK)
	err = c.writePropertiesFile(localSecureConfigPath, secureConfigProps)
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
func (c *PasswordProtectionSuite) RotateMasterKey(oldPassphrase string, newPassphrase string, localSecureConfigPath string) (string, error) {

	cipherSuite, err := c.loadCipherSuiteFromLocalFile(localSecureConfigPath)
	if err != nil {
		return "", err
	}

	// Load MEK
	masterKey, err := c.loadMasterKey()
	if err != nil {
		return "", err
	}

	userMasterKey, err := c.generateMasterKey(oldPassphrase, &cipherSuite)
	if err != nil {
		return "", err
	}

	// Verify master key passphrase
	if masterKey != userMasterKey {
		return "", fmt.Errorf("Authentication Failure: Invalid master key passphrase.")
	}

	// Unwrap DEK using the MEK
	engine := NewEncryptionEngine(&cipherSuite, c.Logger)
	dataKey, err := c.unwrapDataKey(cipherSuite.EncryptedDataKey, engine)
	if err != nil {
		return "", err
	}

	newMasterKey, err := c.generateMasterKey(newPassphrase, &cipherSuite)
	if err != nil {
		return "", err
	}

	// Wrap DEK using the new MEK
	wrappedDataKey, iv, err := engine.WrapDataKey(dataKey, newMasterKey)
	if err != nil {
		return "", err
	}
	newEncodedDataKey := c.formatCipherValue(wrappedDataKey, iv)

	secureConfigProps, err := c.loadPropertiesFile(localSecureConfigPath)
	if err != nil {
		return "", err
	}

	// Save DEK
	now := time.Now()
	_, _, _ = secureConfigProps.Set(METADATA_KEY_TIMESTAMP, now.String())
	_, _, _ = secureConfigProps.Set(METADATA_DATA_KEY, newEncodedDataKey)

	err = c.writePropertiesFile(localSecureConfigPath, secureConfigProps)

	return newMasterKey, err
}

/**
 * This function adds a new key value pair to the configFilePath property file. The original 'value' is
 * encrypted value and stored in the secureConfigPath properties file with key as
 * configFilePath:key and value as encrypted password.
 * We also add the properties to instantiate the SecurePass provider to the config properties file.
 */
func (c *PasswordProtectionSuite) AddEncryptedPasswords(configFilePath string, localSecureConfigPath string, remoteSecureConfigPath string, newConfigs string) error {
	newConfigProps := properties.MustLoadString(newConfigs)
	configProps, err := c.loadPropertiesFile(configFilePath)

	if err != nil {
		return err
	}

	if newConfigProps.Len() == 0 {
		return fmt.Errorf("Configs file is empty !!!")
	}

	err = c.encryptConfigValues(newConfigProps, localSecureConfigPath, configProps, configFilePath, remoteSecureConfigPath)

	return err
}

/**
 * This function updates a the value of existing keys in the configFilePath property file. The original 'value' is
 * encrypted value and stored in the secureConfigPath properties file with key as
 * configFilePath:key and value as encrypted password.
 * We also add the properties to instantiate the SecurePass provider to the config properties file.
 */
func (c *PasswordProtectionSuite) UpdateEncryptedPasswords(configFilePath string, localSecureConfigPath string, remoteSecureConfigPath string, newConfigs string) error {
	newConfigProps := properties.MustLoadString(newConfigs)
	configProps, err := c.loadPropertiesFile(configFilePath)

	if err != nil {
		return err
	}

	// Verify if key is already present in config file.
	for key := range newConfigProps.Map() {
		_, exists := configProps.Get(key)
		if exists == false {
			newConfigProps.Delete(key)
		}
	}

	if newConfigProps.Len() == 0 {
		return fmt.Errorf("Update Failed: Keys not present in config file.!!!")
	}

	err = c.encryptConfigValues(newConfigProps, localSecureConfigPath, configProps, configFilePath, remoteSecureConfigPath)

	return err
}

func (c *PasswordProtectionSuite) RemoveEncryptedPasswords(configFilePath string, localSecureConfigPath string, removeConfigs string) error {

	configProps, err := c.loadPropertiesFile(configFilePath)
	if err != nil {
		return err
	}

	secureConfigProps, err := c.loadPropertiesFile(localSecureConfigPath)
	if err != nil {
		return err
	}

	configs := strings.Split(removeConfigs, ",")

	for _, key := range configs {
		pathKey := configFilePath + ":" + key
		configProps.Delete(key)
		secureConfigProps.Delete(pathKey)
	}

	err = c.writePropertiesFile(configFilePath, configProps)
	if err != nil {
		return err
	}

	err = c.writePropertiesFile(localSecureConfigPath, secureConfigProps)

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

func (c *PasswordProtectionSuite) loadCipherSuiteFromLocalFile(localSecureConfigPath string) (CipherSuite, error) {
	var cipher CipherSuite
	secureConfigProps, err := c.loadPropertiesFile(localSecureConfigPath)
	if err != nil {
		return cipher, err
	}

	return c.loadCipherSuiteFromSecureProps(secureConfigProps)
}

func (c *PasswordProtectionSuite) loadCipherSuiteFromSecureProps(secureConfigProps *properties.Properties) (CipherSuite, error) {
	var cipher CipherSuite
	matchProps, err := secureConfigProps.Filter("(?i)metadata")
	if err != nil {
		return cipher, err
	}
	cipher.Iterations = matchProps.GetInt(METADATA_KEY_ITERATIONS, METADATA_KEY_DEFAULT_ITERATIONS)
	cipher.KeyLength = matchProps.GetInt(METADATA_KEY_LENGTH, METADATA_KEY_DEFAULT_LENGTH_BYTES)
	cipher.Salt = matchProps.GetString(METADATA_KEY_SALT, METADATA_KEY_DEFAULT_SALT)
	cipher.EncryptedDataKey = matchProps.GetString(METADATA_DATA_KEY, "")
	return cipher, nil
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

func (c *PasswordProtectionSuite) writePropertiesFile(path string, property *properties.Properties) error {
	buf := new(bytes.Buffer)
	_, _ = property.WriteComment(buf, "# ", properties.ISO_8859_1)
	err := ioutil.WriteFile(path, buf.Bytes(), 0644)
	return err
}

func (c *PasswordProtectionSuite) loadPropertiesFile(path string) (*properties.Properties, error) {
	if !c.isPathValid(path) {
		return properties.NewProperties(), fmt.Errorf("Invalid file path.")
	}

	property := properties.MustLoadFile(path, properties.ISO_8859_1)
	return property, nil
}

func (c *PasswordProtectionSuite) isPasswordEncrypted(config string) bool {
	regex, _ := regexp.Compile("\\$\\{(.*?):((.*?):)?(.*?)\\}")
	return regex.MatchString(config)
}

func (c *PasswordProtectionSuite) generateConfigValue(key string, path string) string {
	return "${securePass:" + path + ":" + key + "}"
}

func (c *PasswordProtectionSuite) formatCipherValue(cipher string, iv string) string {
	return "ENC[" + METADATA_ENC_ALGORITHM + ",data:" + cipher + ",iv:" + iv + ",type:str]"
}

func (c *PasswordProtectionSuite) isCipher(config string) bool {
	regex, _ := regexp.Compile("ENC\\[(.*?)\\]")
	return regex.MatchString(config)
}

func (c *PasswordProtectionSuite) addSecureConfigProviderProperty(property *properties.Properties) *properties.Properties {
	configProviders := property.GetString(CONFIG_PROVIDER_KEY, "")
	if configProviders == "" {
		configProviders = SECURE_CONFIG_PROVIDER
	} else if strings.Contains(configProviders, SECURE_CONFIG_PROVIDER) {
		configProviders = configProviders + "," + SECURE_CONFIG_PROVIDER
	}

	_, _, _ = property.Set(CONFIG_PROVIDER_KEY, configProviders)
	_, _, _ = property.Set(SECURE_CONFIG_PROVIDER_CLASS_KEY, SECURE_CONFIG_PROVIDER_CLASS)
	return property
}

func (c *PasswordProtectionSuite) unwrapDataKey(key string, engine EncryptionEngine) ([]byte, error) {
	masterKey, err := c.loadMasterKey()
	if err != nil {
		return []byte{}, err
	}
	data, iv, algo := c.parseCipherValue(key)
	return engine.UnWrapDataKey(data, iv, algo, masterKey)
}

func (c *PasswordProtectionSuite) fetchSecureConfigProps(localSecureConfigPath string, masterKey string) (*properties.Properties, *CipherSuite, error) {

	secureConfigProps, _ := c.loadPropertiesFile(localSecureConfigPath)

	// Check if secure config properties file exists and DEK is generated
	if c.isPathValid(localSecureConfigPath) {
		cipherSuite, err := c.loadCipherSuiteFromSecureProps(secureConfigProps)
		// Data Key is already created
		if cipherSuite.EncryptedDataKey != "" {
			return secureConfigProps, &cipherSuite, err
		}
	}

	// Generate a new DEK
	cipherSuites, err := c.generateNewDataKey(masterKey)

	// Add DEK Metadata to secureConfigProps
	now := time.Now()
	_, _, _ = secureConfigProps.Set(METADATA_KEY_TIMESTAMP, now.String())

	_, _, _ = secureConfigProps.Set(METADATA_KEY_ENVVAR, CONFLUENT_KEY_ENVVAR)

	_, _, _ = secureConfigProps.Set(METADATA_KEY_LENGTH, strconv.Itoa(cipherSuites.KeyLength))
	_, _, _ = secureConfigProps.Set(METADATA_KEY_ITERATIONS, strconv.Itoa(cipherSuites.Iterations))
	_, _, _ = secureConfigProps.Set(METADATA_KEY_SALT, cipherSuites.Salt)
	_, _, _ = secureConfigProps.Set(METADATA_DATA_KEY, cipherSuites.EncryptedDataKey)

	return secureConfigProps, cipherSuites, err
}

func (c *PasswordProtectionSuite) loadMasterKey() (string, error) {
	// Check if master key is created and set in the environment variable
	masterKey, found := os.LookupEnv(CONFLUENT_KEY_ENVVAR)
	if !found {
		return "", fmt.Errorf("master key is not exported in %s environment variable; export the key and execute this command again", CONFLUENT_KEY_ENVVAR)
	}
	return masterKey, nil
}

func (c *PasswordProtectionSuite) encryptConfigValues(matchProps *properties.Properties, localSecureConfigPath string, configProps *properties.Properties,
	configFilePath string, remoteConfigFilePath string) error {

	// Load master Key
	masterKey, err := c.loadMasterKey()
	if err != nil {
		return err
	}

	// Fetch secure config props and cipher suite
	secureConfigProps, cipherSuite, err := c.fetchSecureConfigProps(localSecureConfigPath, masterKey)
	if err != nil {
		return err
	}

	// Unwrap DEK
	engine := NewEncryptionEngine(cipherSuite, c.Logger)
	dataKey, err := c.unwrapDataKey(cipherSuite.EncryptedDataKey, engine)
	if err != nil {
		return err
	}

	for key, value := range matchProps.Map() {
		if !c.isPasswordEncrypted(value) {
			// Generate tuple ${providerName:[path:]key}
			pathKey := configFilePath + ":" + key
			newConfigVal := c.generateConfigValue(pathKey, remoteConfigFilePath)
			_, _, _ = configProps.Set(key, newConfigVal)
			cipher, iv, err := engine.AESEncrypt(value, dataKey)

			if err != nil {
				return err
			}
			formattedCipher := c.formatCipherValue(cipher, iv)
			_, _, err = secureConfigProps.Set(pathKey, formattedCipher)
			if err != nil {
				return err
			}
		}

	}

	configProps = c.addSecureConfigProviderProperty(configProps)

	err = c.writePropertiesFile(configFilePath, configProps)
	if err != nil {
		return err
	}

	err = c.writePropertiesFile(localSecureConfigPath, secureConfigProps)
	if err != nil {
		return err
	}

	return nil
}

func (c *PasswordProtectionSuite) findMatchTrim(original string, pattern string, prefix string, suffix string) string {
	re := regexp.MustCompile(pattern)
	match := re.FindStringSubmatch(original)
	substring := ""
	if len(match) != 0 {
		substring = strings.TrimPrefix(strings.TrimSuffix(match[0], suffix), prefix)
	}
	return substring
}

func (c *PasswordProtectionSuite) parseCipherValue(cipher string) (string, string, string) {
	data := c.findMatchTrim(cipher, "data\\:(.*?)\\,", "data:", ",")
	iv := c.findMatchTrim(cipher, "iv\\:(.*?)\\,", "iv:", ",")
	algo := c.findMatchTrim(cipher, "ENC\\[(.*?)\\,", "ENC[", ",")
	return data, iv, algo
}
