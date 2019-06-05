package secret

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"

	"github.com/jonboulle/clockwork"
	"github.com/magiconair/properties"

	"github.com/confluentinc/cli/internal/pkg/log"
)

// Password Protection is a security plugin to securely store and add passwords to a properties file.
// Passwords in property file are encrypted and stored in security config file.

type PasswordProtection interface {
	CreateMasterKey(passphrase string, localSecureConfigPath string) (string, error)
	EncryptConfigFileSecrets(configFilePath string, localSecureConfigPath string, remoteSecureConfigPath string, encryptConfigKeys string) error
	DecryptConfigFileSecrets(configFilePath string, localSecureConfigPath string, outputFilePath string) error
	AddEncryptedPasswords(configFilePath string, localSecureConfigPath string, remoteSecureConfigPath string, newConfigs string) error
	UpdateEncryptedPasswords(configFilePath string, localSecureConfigPath string, remoteSecureConfigPath string, newConfigs string) error
	RemoveEncryptedPasswords(configFilePath string, localSecureConfigPath string, removeConfigs string) error
	RotateMasterKey(oldPassphrase string, newPassphrase string, localSecureConfigPath string) (string, error)
	RotateDataKey(passphrase string, localSecureConfigPath string) error
}

type PasswordProtectionSuite struct {
	Logger      *log.Logger
	Clock       clockwork.Clock
	RandSource  rand.Source
	CipherSuite Cipher
}

func NewPasswordProtectionPlugin(logger *log.Logger) *PasswordProtectionSuite {
	return &PasswordProtectionSuite{Logger: logger, Clock: clockwork.NewRealClock(), RandSource: &cryptoSource{Logger: logger}}
}

// This function generates a new data key for encryption/decryption of secrets. The DEK is wrapped using the master key and saved in the secrets file
// along with other metadata.
func (c *PasswordProtectionSuite) CreateMasterKey(passphrase string, localSecureConfigPath string) (string, error) {
	if len(strings.TrimSpace(passphrase)) == 0 {
		return "", fmt.Errorf("master key passphrase cannot be empty")
	}

	secureConfigProps := properties.NewProperties()
	// Check if secure config properties file exists and DEK is generated
	if DoesPathExist(localSecureConfigPath) {
		secureConfigProps, err := LoadPropertiesFile(localSecureConfigPath)
		if err != nil {
			return "", err
		}
		cipherSuite, err := c.loadCipherSuiteFromSecureProps(secureConfigProps)
		if err != nil {
			return "", err
		}
		// Data Key is already created
		if cipherSuite.EncryptedDataKey != "" {
			return "", fmt.Errorf("master key is already generated, to change the key invoke the rotate command")
		}
	}

	// Generate the master key from passphrase
	cipherSuite := NewDefaultCipher()
	engine := NewEncryptionEngine(cipherSuite, c.Logger, c.RandSource)
	newMasterKey, salt, err := engine.GenerateMasterKey(passphrase, "")
	if err != nil {
		return "", err
	}

	// save the master key salt
	_, _, err = secureConfigProps.Set(METADATA_MEK_SALT, salt)
	if err != nil {
		return "", err
	}

	now := c.Clock.Now()
	_, _, err = secureConfigProps.Set(METADATA_KEY_TIMESTAMP, now.String())
	if err != nil {
		return "", err
	}

	err = WritePropertiesFile(localSecureConfigPath, secureConfigProps, true)
	if err != nil {
		return "", err
	}

	return newMasterKey, nil
}

func (c *PasswordProtectionSuite) generateNewDataKey(masterKey string) (*Cipher, error) {
	// Generate the metadata for encryption keys
	cipherSuite := NewDefaultCipher()
	engine := NewEncryptionEngine(cipherSuite, c.Logger, c.RandSource)

	// Generate a new data key. This data key will be used for encrypting the secrets.
	dataKey, salt, err := engine.GenerateRandomDataKey(METADATA_KEY_DEFAULT_LENGTH_BYTES)
	if err != nil {
		return nil, err
	}

	// Wrap data key with master key
	encodedDataKey, err := c.wrapDataKey(engine, dataKey, masterKey)
	if err != nil {
		return nil, err
	}
	cipherSuite.EncryptedDataKey = encodedDataKey
	cipherSuite.SaltDEK = salt

	return cipherSuite, err
}

// This function encrypts all the passwords in configFilePath properties files. It searches for keys with keyword 'password'
// in properties file configFilePath and encrypts the password using the encryption engine. The original password in the
// properties file is replaced with tuple ${providerName:[path:]key} and encrypted password is stored in secureConfigPath
// properties file with key as configFilePath:key and value as encrypted password.
// We also add the properties to instantiate the SecurePass provider to the config properties file.
func (c *PasswordProtectionSuite) EncryptConfigFileSecrets(configFilePath string, localSecureConfigPath string, remoteSecureConfigPath string, encryptConfigKeys string) error {
	// Check if config file path is valid.
	if !DoesPathExist(configFilePath) {
		return fmt.Errorf("invalid config file path: %s", configFilePath)
	}
	// Load the configs.
	configProps, err := LoadPropertiesFile(configFilePath)
	configProps.DisableExpansion = true
	if err != nil {
		return err
	}

	matchProps := properties.NewProperties()
	if encryptConfigKeys != "" {
		configs := strings.Split(encryptConfigKeys, ",")
		for _, key := range configs {
			key := strings.TrimSpace(key)
			value, ok := configProps.Get(key)
			// If key present in config file
			if ok {
				_, _, err = matchProps.Set(key, value)
				if err != nil {
					return err
				}
			} else {
				return fmt.Errorf("Key " + key + " not present in config file.")
			}
		}
	} else {
		// Filter the properties which have keyword 'password' in the key.
		matchProps, err = configProps.Filter("(?i).password")
		if err != nil {
			return err
		}
	}
	// Encrypt the secrets with DEK. Save the encrypted secrets in secure config file.
	return c.encryptConfigValues(matchProps, localSecureConfigPath, configProps, configFilePath, remoteSecureConfigPath)
}

// This function decrypts all the passwords in configFilePath properties files and stores the decrypted passwords in outputFilePath.
// It searches for the encrypted secrets by comparing it with the tuple ${providerName:[path:]key}. If encrypted secrets are found it fetches
// the encrypted value from the file secureConfigPath, decrypts it using the data key and stores the output at outputFilePath.
func (c *PasswordProtectionSuite) DecryptConfigFileSecrets(configFilePath string, localSecureConfigPath string, outputFilePath string) error {
	// Check if config file path is valid
	if !DoesPathExist(configFilePath) {
		return fmt.Errorf("invalid config file path:" + configFilePath)
	}

	// Check if secure config file path is valid
	if !DoesPathExist(localSecureConfigPath) {
		return fmt.Errorf("invalid secrets file path:" + localSecureConfigPath)
	}
	// Load the config values.
	configProps, err := LoadPropertiesFile(configFilePath)
	if err != nil {
		return err
	}

	// Load the encrypted config value.
	secureConfigProps, err := LoadPropertiesFile(localSecureConfigPath)
	if err != nil {
		return err
	}
	//decryptedSecrets := properties.NewProperties()
	cipherSuite, err := c.loadCipherSuiteFromLocalFile(localSecureConfigPath)
	if err != nil {
		return err
	}

	engine := NewEncryptionEngine(cipherSuite, c.Logger, c.RandSource)
	// Unwrap DEK with MEK
	dataKey, err := c.unwrapDataKey(cipherSuite.EncryptedDataKey, engine)
	if err != nil {
		c.Logger.Debug(err)
		return fmt.Errorf("failed to unwrap the data key due to invalid master key or corrupted data key.")
	}

	for key, value := range configProps.Map() {
		// If config value is encrypted, decrypt it with DEK.
		encryptedPass, err := c.isPasswordEncrypted(value)
		if err != nil {
			return err
		}
		if encryptedPass {
			pathKey := GenerateConfigKey(configFilePath, key)
			cipher := secureConfigProps.GetString(pathKey, "")
			if cipher != "" {
				data, iv, algo := ParseCipherValue(cipher)
				plainSecret, err := engine.Decrypt(data, iv, algo, dataKey)
				if err != nil {
					c.Logger.Debug(err)
					return fmt.Errorf("failed to decrypt config %s due to corrupted data.", key)
				}
				_, _, err = configProps.Set(key, plainSecret)
				if err != nil {
					return err
				}
			} else {
				return fmt.Errorf("missing config key in secret config file.")
			}
		} else {
			configProps.Delete(key)
		}

	}

	// Save the decrypted ciphers to output file.
	return WritePropertiesFile(outputFilePath, configProps, false)
}

// This function generates a new data key and re-encrypts the values in the secureConfigPath properties file with the new data key.
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

	engine := NewEncryptionEngine(cipherSuite, c.Logger, c.RandSource)

	// Generate a master key from passphrase
	userMasterKey, _, err := engine.GenerateMasterKey(masterPassphrase, cipherSuite.SaltMEK)
	if err != nil {
		return err
	}

	// Verify master key passphrase
	if masterKey != userMasterKey {
		return fmt.Errorf("authentication failure: incorrect master key passphrase.")
	}

	secureConfigProps, err := LoadPropertiesFile(localSecureConfigPath)
	if err != nil {
		return err
	}

	// Unwrap old DEK using the MEK
	dataKey, err := c.unwrapDataKey(cipherSuite.EncryptedDataKey, engine)
	if err != nil {
		c.Logger.Debug(err)
		return fmt.Errorf("failed to unwrap the data key due to invalid master key or corrupted data key.")
	}

	// Generate a new DEK
	newDataKey, salt, err := engine.GenerateRandomDataKey(METADATA_KEY_DEFAULT_LENGTH_BYTES)
	if err != nil {
		return err
	}

	// Re-encrypt the ciphers with new DEK
	for key, value := range secureConfigProps.Map() {
		encrypted, err := c.isCipher(value)
		if err != nil {
			return err
		}
		if encrypted && !strings.HasPrefix(key, METADATA_PREFIX) {
			data, iv, algo := ParseCipherValue(value)
			plainSecret, err := engine.Decrypt(data, iv, algo, dataKey)
			if err != nil {
				return err
			}
			cipher, iv, err := engine.Encrypt(plainSecret, newDataKey)
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
	now := c.Clock.Now()
	_, _, err = secureConfigProps.Set(METADATA_KEY_TIMESTAMP, now.String())
	if err != nil {
		return err
	}
	_, _, err = secureConfigProps.Set(METADATA_DATA_KEY, wrappedNewDK)
	if err != nil {
		return err
	}
	_, _, err = secureConfigProps.Set(METADATA_DEK_SALT, salt)
	if err != nil {
		return err
	}
	err = WritePropertiesFile(localSecureConfigPath, secureConfigProps, true)
	if err != nil {
		return err
	}

	return nil
}

// This function is used to change the master key. It wraps the data key with newly set master key.
func (c *PasswordProtectionSuite) RotateMasterKey(oldPassphrase string, newPassphrase string, localSecureConfigPath string) (string, error) {
	if len(strings.TrimSpace(oldPassphrase)) == 0 || len(strings.TrimSpace(newPassphrase)) == 0 {
		return "", fmt.Errorf("master key passphrase cannot be empty.")
	}

	if strings.Compare(oldPassphrase, newPassphrase) == 0 {
		return "", fmt.Errorf("new master key passphrase may not be the same as the previous passphrase.")
	}

	cipherSuite, err := c.loadCipherSuiteFromLocalFile(localSecureConfigPath)
	if err != nil {
		return "", err
	}

	// Load MEK
	masterKey, err := c.loadMasterKey()
	if err != nil {
		return "", err
	}

	engine := NewEncryptionEngine(cipherSuite, c.Logger, c.RandSource)

	// Generate a master key from passphrase
	userMasterKey, _, err := engine.GenerateMasterKey(oldPassphrase, cipherSuite.SaltMEK)
	if err != nil {
		return "", err
	}

	// Verify master key passphrase
	if masterKey != userMasterKey {
		return "", fmt.Errorf("authentication failure: incorrect master key passphrase.")
	}

	// Unwrap DEK using the MEK
	dataKey, err := c.unwrapDataKey(cipherSuite.EncryptedDataKey, engine)
	if err != nil {
		c.Logger.Debug(err)
		return "", fmt.Errorf("Failed to unwrap the Data Key due to invalid master key.")
	}

	newMasterKey, salt, err := engine.GenerateMasterKey(newPassphrase, "")
	if err != nil {
		return "", err
	}

	// Wrap DEK using the new MEK
	wrappedDataKey, iv, err := engine.WrapDataKey(dataKey, newMasterKey)
	if err != nil {
		return "", err
	}
	newEncodedDataKey := c.formatCipherValue(wrappedDataKey, iv)

	secureConfigProps, err := LoadPropertiesFile(localSecureConfigPath)
	if err != nil {
		return "", err
	}

	// Save DEK
	now := c.Clock.Now()
	_, _, err = secureConfigProps.Set(METADATA_KEY_TIMESTAMP, now.String())
	if err != nil {
		return "", err
	}
	_, _, err = secureConfigProps.Set(METADATA_DATA_KEY, newEncodedDataKey)
	if err != nil {
		return "", err
	}

	// save the master key salt
	_, _, err = secureConfigProps.Set(METADATA_MEK_SALT, salt)
	if err != nil {
		return "", err
	}

	err = WritePropertiesFile(localSecureConfigPath, secureConfigProps, true)
	if err != nil {
		return "", err
	}

	return newMasterKey, nil
}

// This function adds a new key value pair to the configFilePath property file. The original 'value' is
// encrypted value and stored in the secureConfigPath properties file with key as
// configFilePath:key and value as encrypted password.
// We also add the properties to instantiate the SecurePass provider to the config properties file.
func (c *PasswordProtectionSuite) AddEncryptedPasswords(configFilePath string, localSecureConfigPath string, remoteSecureConfigPath string, newConfigs string) error {
	newConfigs = strings.Replace(newConfigs, `\n`, "\n", -1)
	newConfigProps, err := properties.LoadString(newConfigs)
	if err != nil {
		return err
	}

	configProps, err := LoadPropertiesFile(configFilePath)
	if err != nil {
		return err
	}

	if newConfigProps.Len() == 0 {
		return fmt.Errorf("add failed: empty list of new configs")
	}

	err = c.encryptConfigValues(newConfigProps, localSecureConfigPath, configProps, configFilePath, remoteSecureConfigPath)

	return err
}

// This function updates a the value of existing keys in the configFilePath property file. The original 'value' is
// encrypted value and stored in the secureConfigPath properties file with key as
// configFilePath:key and value as encrypted password.
// We also add the properties to instantiate the SecurePass provider to the config properties file.
func (c *PasswordProtectionSuite) UpdateEncryptedPasswords(configFilePath string, localSecureConfigPath string, remoteSecureConfigPath string, newConfigs string) error {
	newConfigs = strings.Replace(newConfigs, `\n`, "\n", -1)
	newConfigProps, err := properties.LoadString(newConfigs)
	if err != nil {
		return err
	}
	configProps, err := LoadPropertiesFile(configFilePath)
	configProps.DisableExpansion = true

	if err != nil {
		return err
	}

	// Verify if key is already present in config file.
	for key := range newConfigProps.Map() {
		_, exists := configProps.Get(key)
		if !exists {
			return fmt.Errorf("config " + key + " not present in config file.")
		}
	}

	if newConfigProps.Len() == 0 {
		return fmt.Errorf("update failed: empty list of update configs")
	}

	err = c.encryptConfigValues(newConfigProps, localSecureConfigPath, configProps, configFilePath, remoteSecureConfigPath)

	return err
}

func (c *PasswordProtectionSuite) RemoveEncryptedPasswords(configFilePath string, localSecureConfigPath string, removeConfigs string) error {
	configProps, err := LoadPropertiesFile(configFilePath)
	configProps.DisableExpansion = true
	if err != nil {
		return err
	}

	secureConfigProps, err := LoadPropertiesFile(localSecureConfigPath)
	secureConfigProps.DisableExpansion = true
	if err != nil {
		return err
	}

	configs := strings.Split(removeConfigs, ",")

	for _, key := range configs {
		pathKey := GenerateConfigKey(configFilePath, key)

		//Check if config is present
		_, ok := configProps.Get(key)
		if !ok {
			return fmt.Errorf("config " + key + " not present in config file.")
		}

		// Check if config is removed from secrets files
		_, ok = secureConfigProps.Get(pathKey)
		if !ok {
			return fmt.Errorf("config " + key + " not present in secrets file.")
		}
		configProps.Delete(key)
		secureConfigProps.Delete(pathKey)
	}

	err = WritePropertiesFile(configFilePath, configProps, true)
	if err != nil {
		return err
	}

	err = WritePropertiesFile(localSecureConfigPath, secureConfigProps, true)
	return err
}

// Helper Functions

func (c *PasswordProtectionSuite) wrapDataKey(engine EncryptionEngine, dataKey []byte, masterKey string) (string, error) {
	wrappedDataKey, iv, err := engine.WrapDataKey(dataKey, masterKey)
	if err != nil {
		return "", err
	}

	encodedDataKey := c.formatCipherValue(wrappedDataKey, iv)

	return encodedDataKey, nil
}

func (c *PasswordProtectionSuite) loadCipherSuiteFromLocalFile(localSecureConfigPath string) (*Cipher, error) {
	secureConfigProps, err := LoadPropertiesFile(localSecureConfigPath)
	if err != nil {
		return nil, err
	}

	return c.loadCipherSuiteFromSecureProps(secureConfigProps)
}

func (c *PasswordProtectionSuite) loadCipherSuiteFromSecureProps(secureConfigProps *properties.Properties) (*Cipher, error) {
	matchProps := secureConfigProps.FilterPrefix("_metadata")
	matchProps.DisableExpansion = true
	cipher := NewDefaultCipher()
	cipher.Iterations = matchProps.GetInt(METADATA_KEY_ITERATIONS, METADATA_KEY_DEFAULT_ITERATIONS)
	cipher.KeyLength = matchProps.GetInt(METADATA_KEY_LENGTH, METADATA_KEY_DEFAULT_LENGTH_BYTES)
	cipher.SaltDEK = matchProps.GetString(METADATA_DEK_SALT, "")
	cipher.SaltMEK = matchProps.GetString(METADATA_MEK_SALT, "")
	cipher.EncryptedDataKey = matchProps.GetString(METADATA_DATA_KEY, "")
	return cipher, nil
}

func (c *PasswordProtectionSuite) isPasswordEncrypted(config string) (bool, error) {
	return passwordRegex.MatchString(config), nil
}

func (c *PasswordProtectionSuite) formatCipherValue(cipher string, iv string) string {
	return "ENC[" + METADATA_ENC_ALGORITHM + ",data:" + cipher + ",iv:" + iv + ",type:str]"
}

func (c *PasswordProtectionSuite) isCipher(config string) (bool, error) {
	return cipherRegex.MatchString(config), nil
}

func (c *PasswordProtectionSuite) addSecureConfigProviderProperty(property *properties.Properties) (*properties.Properties, error) {
	property.DisableExpansion = true
	configProviders := property.GetString(CONFIG_PROVIDER_KEY, "")
	if configProviders == "" {
		configProviders = SECURE_CONFIG_PROVIDER
	} else if !strings.Contains(configProviders, SECURE_CONFIG_PROVIDER) {
		configProviders = configProviders + "," + SECURE_CONFIG_PROVIDER
	}

	_, _, err := property.Set(CONFIG_PROVIDER_KEY, configProviders)
	if err != nil {
		return nil, err
	}
	_, _, err = property.Set(SECURE_CONFIG_PROVIDER_CLASS_KEY, SECURE_CONFIG_PROVIDER_CLASS)
	if err != nil {
		return nil, err
	}
	return property, nil
}

func (c *PasswordProtectionSuite) unwrapDataKey(key string, engine EncryptionEngine) ([]byte, error) {
	masterKey, err := c.loadMasterKey()
	if err != nil {
		return []byte{}, err
	}
	data, iv, algo := ParseCipherValue(key)
	return engine.UnwrapDataKey(data, iv, algo, masterKey)
}

func (c *PasswordProtectionSuite) fetchSecureConfigProps(localSecureConfigPath string, masterKey string) (*properties.Properties, *Cipher, error) {
	secureConfigProps, err := LoadPropertiesFile(localSecureConfigPath)
	if err != nil {
		secureConfigProps = properties.NewProperties()
	}

	// Check if secure config properties file exists and DEK is generated
	if DoesPathExist(localSecureConfigPath) {
		cipherSuite, err := c.loadCipherSuiteFromSecureProps(secureConfigProps)
		if err != nil {
			return nil, nil, err
		}
		// Data Key is already created
		if cipherSuite.EncryptedDataKey != "" {
			return secureConfigProps, cipherSuite, err
		}
	}

	// Generate a new DEK
	cipherSuites, err := c.generateNewDataKey(masterKey)
	if err != nil {
		return nil, nil, err
	}

	// Add DEK Metadata to secureConfigProps
	now := c.Clock.Now()
	_, _, err = secureConfigProps.Set(METADATA_KEY_TIMESTAMP, now.String())
	if err != nil {
		return nil, nil, err
	}
	_, _, err = secureConfigProps.Set(METADATA_KEY_ENVVAR, CONFLUENT_KEY_ENVVAR)
	if err != nil {
		return nil, nil, err
	}
	_, _, err = secureConfigProps.Set(METADATA_KEY_LENGTH, strconv.Itoa(cipherSuites.KeyLength))
	if err != nil {
		return nil, nil, err
	}
	_, _, err = secureConfigProps.Set(METADATA_KEY_ITERATIONS, strconv.Itoa(cipherSuites.Iterations))
	if err != nil {
		return nil, nil, err
	}
	_, _, err = secureConfigProps.Set(METADATA_DEK_SALT, cipherSuites.SaltDEK)
	if err != nil {
		return nil, nil, err
	}
	_, _, err = secureConfigProps.Set(METADATA_DATA_KEY, cipherSuites.EncryptedDataKey)
	if err != nil {
		return nil, nil, err
	}
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
	engine := NewEncryptionEngine(cipherSuite, c.Logger, c.RandSource)
	dataKey, err := c.unwrapDataKey(cipherSuite.EncryptedDataKey, engine)
	if err != nil {
		c.Logger.Debug(err)
		return fmt.Errorf("failed to unwrap the data key due to invalid master key or corrupted data key.")
	}

	for key, value := range matchProps.Map() {
		encryptedPass, err := c.isPasswordEncrypted(value)
		if err != nil {
			return err
		}
		if !encryptedPass {
			// Generate tuple ${providerName:[path:]key}
			pathKey := GenerateConfigKey(configFilePath, key)
			newConfigVal := GenerateConfigValue(pathKey, remoteConfigFilePath)
			_, _, err = configProps.Set(key, newConfigVal)
			if err != nil {
				return err
			}
			cipher, iv, err := engine.Encrypt(value, dataKey)

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

	configProps, err = c.addSecureConfigProviderProperty(configProps)
	if err != nil {
		return err
	}

	err = WritePropertiesFile(configFilePath, configProps, true)
	if err != nil {
		return err
	}

	err = WritePropertiesFile(localSecureConfigPath, secureConfigProps, true)
	if err != nil {
		return err
	}

	return nil
}
