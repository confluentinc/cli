package secret

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/confluentinc/properties"
	"github.com/tidwall/gjson"
	"github.com/tidwall/pretty"
	"github.com/tidwall/sjson"
)

var dataRegex = regexp.MustCompile(DataPattern)
var ivRegex = regexp.MustCompile(IVPattern)
var algoRegex = regexp.MustCompile(EncPattern)
var passwordRegex = regexp.MustCompile(PasswordPattern)
var cipherRegex = regexp.MustCompile(CipherPattern)

func GenerateConfigValue(key string, path string) string {
	return "${securepass:" + path + ":" + key + "}"
}

func ParseCipherValue(cipher string) (string, string, string) {
	data := findMatchTrim(cipher, dataRegex, "data:", ",")
	iv := findMatchTrim(cipher, ivRegex, "iv:", ",")
	algo := findMatchTrim(cipher, algoRegex, "ENC[", ",")
	return data, iv, algo
}

func findMatchTrim(original string, re *regexp.Regexp, prefix string, suffix string) string {
	match := re.FindStringSubmatch(original)
	substring := ""
	if len(match) != 0 {
		substring = strings.TrimPrefix(strings.TrimSuffix(match[0], suffix), prefix)
	}
	return substring
}

func SaveConfiguration(path string, configuration *properties.Properties, addSecureConfig bool) error {
	switch filepath.Ext(path) {
	case ".properties":
		return writePropertiesConfig(path, configuration, addSecureConfig)
	case ".json":
		return writeJSONConfig(path, configuration, addSecureConfig)
	default:
		return fmt.Errorf("The file format is currently not supported.")
	}
}

func WritePropertiesFile(path string, property *properties.Properties, writeComments bool) error {
	buf := new(bytes.Buffer)
	if writeComments {
		_, err := property.WriteFormattedComment(buf, properties.UTF8)
		if err != nil {
			return err
		}
	} else {
		_, err := property.Write(buf, properties.UTF8)
		if err != nil {
			return err
		}

	}

	err := WriteFile(path, buf.Bytes())
	return err
}

func DoesPathExist(path string) bool {
	if path == "" {
		return false
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func LoadPropertiesFile(path string) (*properties.Properties, error) {
	if !DoesPathExist(path) {
		return nil, fmt.Errorf("Invalid file path.")
	}
	loader := new(properties.Loader)
	loader.Encoding = properties.UTF8
	loader.PreserveFormatting = true
	//property.DisableExpansion = true
	property, err := loader.LoadFile(path)
	if err != nil {
		return nil, err
	}
	property.DisableExpansion = true
	return property, nil
}

func addSecureConfigProviderProperty(property *properties.Properties) (*properties.Properties, error) {
	property.DisableExpansion = true
	configProviders := property.GetString(ConfigProviderKey, "")
	if configProviders == "" {
		configProviders = SecureConfigProvider
	} else if !strings.Contains(configProviders, SecureConfigProvider) {
		configProviders = configProviders + "," + SecureConfigProvider
	}

	_, _, err := property.Set(ConfigProviderKey, configProviders)
	if err != nil {
		return nil, err
	}
	_, _, err = property.Set(SecureConfigProviderClassKey, SecureConfigProviderClass)
	if err != nil {
		return nil, err
	}
	return property, nil
}

func LoadConfiguration(path string, configKeys []string, filter bool) (*properties.Properties, error) {
	if !DoesPathExist(path) {
		return nil, fmt.Errorf("Invalid file path.")
	}
	fileType := filepath.Ext(path)
	switch fileType {
	case ".properties":
		return loadPropertiesConfig(path, configKeys, filter)
	case ".json":
		return loadJSONConfig(path, configKeys)
	default:
		return nil, fmt.Errorf("The file format is currently not supported.")
	}
}

func filterProperties(configProps *properties.Properties, configKeys []string, filterPassword bool) (*properties.Properties, error) {
	configProps.DisableExpansion = true
	matchProps := properties.NewProperties()
	matchProps.DisableExpansion = true
	if len(configKeys) > 0 {
		for _, key := range configKeys {
			key := strings.TrimSpace(key)
			value, ok := configProps.Get(key)
			// If key present in config file
			if ok {
				_, _, err := matchProps.Set(key, value)
				if err != nil {
					return nil, err
				}
			} else {
				return nil, fmt.Errorf("Configuration key " + key + " is not present in the configuration file.")
			}
		}
		return matchProps, nil
	} else if filterPassword {
		// Filter the properties which have keyword 'password' in the key.
		matchProps, err := configProps.Filter("(?i).password")
		if err != nil {
			return nil, err
		}

		return matchProps, nil
	}

	return configProps, nil
}

func loadPropertiesConfig(path string, configKeys []string, filter bool) (*properties.Properties, error) {
	loader := new(properties.Loader)
	loader.Encoding = properties.UTF8
	loader.PreserveFormatting = true
	configProps, err := loader.LoadFile(path)
	if err != nil {
		return nil, err
	}
	// convert embedded jaas to props
	configProps = parseJAASProperties(configProps)

	return filterProperties(configProps, configKeys, filter)
}

func parseJAASProperties(props *properties.Properties) *properties.Properties {
	parser := NewJAASParser()
	matchProps, err := props.Filter("(?i).jaas")
	if err != nil {
		return props
	}
	matchProps.DisableExpansion = true
	for key, value := range matchProps.Map() {
		jaasProps, err := parser.ParseJAASConfigurationEntry(value, key)
		if err == nil {
			props.Merge(jaasProps)
		}

	}
	return props
}

func convertPropertiesJAAS(props *properties.Properties, originalConfigs *properties.Properties, op string) (*properties.Properties, error) {
	parser := NewJAASParser()
	matchProps, err := props.Filter("(?i).jaas")
	if err != nil {
		return props, err
	}
	matchProps.DisableExpansion = true

	pattern := regexp.MustCompile(JAASKeyPattern)

	jaasProps := properties.NewProperties()
	jaasProps.DisableExpansion = true
	jaasOriginal := properties.NewProperties()
	jaasOriginal.DisableExpansion = true

	for key, value := range matchProps.Map() {
		if pattern.MatchString(key) {
			parentKeys := strings.Split(key, KeySeparator)
			origKey := parentKeys[ClassId]
			origVal, ok := originalConfigs.Get(origKey)
			if ok {
				_, _, err = jaasProps.Set(key, value)
				if err != nil {
					return props, nil
				}
				_, _, err = jaasOriginal.Set(parentKeys[ClassId]+KeySeparator+parentKeys[ParentId], origVal)
				if err != nil {
					return props, nil
				}
				props.Delete(key)
			}
		}

	}

	parser.SetOriginalConfigKeys(jaasOriginal)
	jaasConf, err := parser.ConvertPropertiesToJAAS(jaasProps, op)
	if err == nil {
		props.Merge(jaasConf)
	}

	return props, err
}

func LoadJSONFile(path string) (string, error) {
	jsonFile, err := os.Open(path)
	// if we os.Open returns an error then handle it
	if err != nil {
		return "", err
	}

	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	// read our opened jsonFile as a byte array.
	jsonByteArr, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return "", err
	}

	jsonConfig := string(jsonByteArr)
	if !gjson.Valid(jsonConfig) {
		return "", fmt.Errorf("Invalid json file format.")
	}

	return jsonConfig, nil
}

func loadJSONConfig(path string, configKeys []string) (*properties.Properties, error) {
	jsonConfig, err := LoadJSONFile(path)
	if err != nil {
		return nil, err
	}

	matchProps := properties.NewProperties()
	for _, key := range configKeys {
		key := strings.TrimSpace(key)

		// If key present in config file
		if gjson.Get(jsonConfig, key).Exists() {
			configValue := gjson.Get(jsonConfig, key)
			_, _, err = matchProps.Set(key, configValue.String())
			if err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("Configuration key " + key + " is not present in JSON configuration file.")
		}
	}

	return matchProps, nil
}

func writePropertiesConfig(path string, configs *properties.Properties, addSecureConfig bool) error {
	configProps, err := LoadPropertiesFile(path)
	if err != nil {
		return err
	}
	configProps.DisableExpansion = true
	configs, err = convertPropertiesJAAS(configs, configProps, Update)

	if err != nil {
		return err
	}

	for key, value := range configs.Map() {
		_, _, err = configProps.Set(key, value)
		if err != nil {
			return err
		}

	}

	if addSecureConfig {
		configProps, err = addSecureConfigProviderProperty(configProps)
		if err != nil {
			return err
		}
	}

	err = WritePropertiesFile(path, configProps, true)
	return err
}

func RemovePropertiesConfig(removeConfigs []string, path string) error {
	configProps, err := LoadPropertiesFile(path)
	pattern := regexp.MustCompile(JAASKeyPattern)
	if err != nil {
		return err
	}
	configProps.DisableExpansion = true
	removeJAASConfig := properties.NewProperties()
	removeJAASConfig.DisableExpansion = true
	for _, key := range removeConfigs {
		//Check if config is present
		if pattern.MatchString(key) {
			_, _, err = removeJAASConfig.Set(key, "")
			if err != nil {
				return err
			}
		} else {
			_, ok := configProps.Get(key)
			if !ok {
				return fmt.Errorf("Configuration key " + key + " is not present in the configuration file.")
			}
			configProps.Delete(key)
		}
	}

	configs, err := convertPropertiesJAAS(removeJAASConfig, configProps, Delete)

	if err != nil {
		return err
	}

	for key, value := range configs.Map() {
		_, _, err = configProps.Set(key, value)
		if err != nil {
			return err
		}

	}

	err = WritePropertiesFile(path, configProps, true)
	return err
}

func writeJSONConfig(path string, configs *properties.Properties, addSecureConfig bool) error {
	jsonConfig, err := LoadJSONFile(path)
	if err != nil {
		return err
	}

	if gjson.Get(jsonConfig, ConfigProviderKey).Exists() {
		configValue := gjson.Get(jsonConfig, ConfigProviderKey)
		_, _, err = configs.Set(ConfigProviderKey, configValue.String())
		if err != nil {
			return err
		}
	}

	for key, value := range configs.Map() {
		jsonConfig, err = sjson.Set(jsonConfig, key, value)
		if err != nil {
			return err
		}
	}

	if addSecureConfig {
		configs, err = addSecureConfigProviderProperty(configs)
		if err != nil {
			return err
		}

		providerKeyJson := strings.ReplaceAll(ConfigProviderKey, ".", "\\.")
		providerClassKeyJson := strings.ReplaceAll(SecureConfigProviderClassKey, ".", "\\.")

		value, _ := configs.Get(ConfigProviderKey)
		jsonConfig, err = sjson.Set(jsonConfig, providerKeyJson, value)
		if err != nil {
			return err
		}
		value, _ = configs.Get(SecureConfigProviderClassKey)
		jsonConfig, err = sjson.Set(jsonConfig, providerClassKeyJson, value)
		if err != nil {
			return err
		}
	}

	result := pretty.Pretty([]byte(jsonConfig))
	err = WriteFile(path, result)
	return err
}

func WriteFile(path string, data []byte) error {
	return ioutil.WriteFile(path, data, 0644)
}

func GenerateConfigKey(path string, key string) string {
	fileName := filepath.Base(path)
	// Intentionally not using the filepath.Join(fileName, key), because even if this CLI is run on Windows we know that
	// the server-side version will be running on a *nix variant and will thus have forward slashes to lookup the correct path
	return fileName + "/" + key
}
