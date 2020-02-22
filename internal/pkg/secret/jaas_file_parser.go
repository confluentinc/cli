package secret

import (
	"fmt"
	"github.com/confluentinc/properties"
	"io/ioutil"
	"regexp"
	"strings"
	"text/scanner"
	"unicode"
)

type JAASParserInterface interface {
	Load(path string) (*properties.Properties, error)
	Write(props *properties.Properties, op string) error
}

// Result represents a jaas value that is returned from Parse().
type JAASParser struct {
	// Raw is the raw jaas
	Raw string
	JaasOriginalConfigKeys *properties.Properties
	JaasProps              *properties.Properties
	JaasClassConfig        *properties.Properties
	Index                  int
	Path                   string
	WhitespaceKey          string
	tokenizer              scanner.Scanner
}

func NewJAASParser() *JAASParser {
	return &JAASParser{
		JaasOriginalConfigKeys: properties.NewProperties(),
		JaasProps:              properties.NewProperties(),
		JaasClassConfig:        properties.NewProperties(),
		Raw:                    "",
		WhitespaceKey:          "",
	}
}

func (j *JAASParser) Load(path string) (*properties.Properties, error) {
	jaasFile, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	j.Path = path
	j.Raw = string(jaasFile)
	j.Raw = strings.TrimSpace(j.Raw)
	parsedConfigs, origConfig, classConfig, err := j.parseJAASFile(j.Raw)
	if err != nil {
		return nil, err
	}

	j.JaasProps = parsedConfigs
	j.JaasProps.DisableExpansion = true
	j.JaasOriginalConfigKeys = origConfig
	j.JaasOriginalConfigKeys.DisableExpansion = true
	j.JaasClassConfig = classConfig
	j.JaasClassConfig.DisableExpansion = true
	return j.JaasProps, nil
}

func (j *JAASParser) addSecureConfigProps(config string) string {
	secureConfigProvider := NEW_LINE + CONFIG_PROVIDER_KEY + j.WhitespaceKey + "=" + j.WhitespaceKey + SECURE_CONFIG_PROVIDER_CLASS_KEY + NEW_LINE +

		SECURE_CONFIG_PROVIDER + j.WhitespaceKey + "=" + j.WhitespaceKey + SECURE_CONFIG_PROVIDER_CLASS
	return strings.TrimSuffix(config, ";") + secureConfigProvider + ";"
}

func (j *JAASParser) updateJAASConfig(op string, key string, value string, config string, addSecretsConfig bool) (string, error) {
	switch op {
	case DELETE:
		keyValuePattern := key + JAAS_VALUE_PATTERN
		pattern := regexp.MustCompile(keyValuePattern)
		delete := ""
		// check if value is in JAAS format
		if pattern.MatchString(config) {
			config = pattern.ReplaceAllString(config, delete)
		} else {
			keyValuePattern := key + PASSWORD_PATTERN // check if value is in Secrets format
			pattern := regexp.MustCompile(keyValuePattern)
			config = pattern.ReplaceAllString(config, delete)
		}
		break
	case UPDATE:
		keyValuePattern := key + JAAS_VALUE_PATTERN
		pattern := regexp.MustCompile(keyValuePattern)
		if pattern.MatchString(config) {
			replaceVal := key + j.WhitespaceKey + "=" + j.WhitespaceKey + value
			matched := pattern.FindString(config)
			config = pattern.ReplaceAllString(config, replaceVal)
			if strings.HasSuffix(matched, ";") {
				config = config + ";"
			}
			if addSecretsConfig && !strings.Contains(config, CONFIG_PROVIDER_KEY) {
				config = j.addSecureConfigProps(config)
			}
		} else {
			add := NEW_LINE + key + j.WhitespaceKey + "=" + j.WhitespaceKey + value
			config = strings.TrimSuffix(config, ";") + add + ";"
			if addSecretsConfig && !strings.Contains(config, CONFIG_PROVIDER_KEY) {
				config = j.addSecureConfigProps(config)
			}
		}
		break
	default:
		return "", fmt.Errorf("operation not supported")
	}

	return config, nil
}

func (j *JAASParser) Write(path string, props *properties.Properties, op string, addSecretsConfig bool) error {
	if j.Raw == "" {
		_, err := j.Load(path)
		if err != nil {
			return err
		}
	}
	rawCopy := j.Raw
	for key, value := range props.Map() {
		keys := strings.Split(key, KEY_SEPARATOR)
		config, ok := j.JaasOriginalConfigKeys.Get(keys[CLASS_ID] + KEY_SEPARATOR + keys[PARENT_ID])
		if !ok {
			return fmt.Errorf("config " + key + " not present in the file")
		}
		newConfig, err := j.updateJAASConfig(op, keys[KEY_ID], value, config, addSecretsConfig)
		if err != nil {
			return err
		}

		originalClass, ok := j.JaasClassConfig.Get(keys[CLASS_ID])
		if !ok {
			return fmt.Errorf("config " + keys[CLASS_ID] + " not present in the file")
		}
		if len(originalClass) == 0 {
			return fmt.Errorf("invalid file format")
		}
		replaceString := strings.Replace(originalClass, config, newConfig, -1)
		rawCopy = strings.Replace(rawCopy, originalClass, replaceString, -1)

		_, _, err = j.JaasClassConfig.Set(keys[CLASS_ID], replaceString)
		if err != nil {
			return err
		}
		_, _, err = j.JaasOriginalConfigKeys.Set(keys[CLASS_ID]+KEY_SEPARATOR+keys[PARENT_ID], config)
		if err != nil {
			return err
		}
	}

	return WriteFile(j.Path, []byte(rawCopy))
}

func (j *JAASParser) parseConfig(specialChar rune) (string, int, error) {
	configName := ""
	offset := -1
	if unicode.IsSpace(j.tokenizer.Peek()) {
		j.tokenizer.Scan()
		configName = j.tokenizer.TokenText()
		offset = j.tokenizer.Pos().Offset
	}

	for j.tokenizer.Peek() != scanner.EOF && !unicode.IsSpace(j.tokenizer.Peek()) && j.tokenizer.Peek() != specialChar {
		j.tokenizer.Scan()
		configName = configName + j.tokenizer.TokenText()
		if offset == -1 {
			offset = j.tokenizer.Pos().Offset
		}
	}
	err := validateConfig(configName)
	if err != nil {
		return "", offset, err
	}
	return configName, offset, nil
}

func validateConfig(config string) error {
	if config == "}" || config == "{" || config == ";" || config == "=" || config == "};" || config == "" || config == " " {
		return fmt.Errorf("invalid jaas file: expected a configuration name received " + config)
	}

	return nil
}

func (j *JAASParser) ignoreBackslash() {
	tokenizer := j.tokenizer
	tokenizer.Scan()
	if tokenizer.TokenText() == "\\" {
		j.tokenizer.Scan()
	}
}

func (j *JAASParser) isClosingBracket() bool {
	// If its whitespace move ahead
	tokenizer := j.tokenizer
	if unicode.IsSpace(tokenizer.Peek()) {
		tokenizer.Scan()
		if tokenizer.TokenText() == "}" {
			j.tokenizer.Scan()
			return true
		}
	} else if tokenizer.Peek() == '}' {
		j.tokenizer.Scan()
		return true
	}

	return false
}

func (j *JAASParser) parseControlFlag() error {
	j.tokenizer.Scan()
	val := j.tokenizer.TokenText()
	switch val {
	case CONTROL_FLAG_REQUIRED, CONTROL_FLAG_REQUISITE, CONTROL_FLAG_OPTIONAL, CONTROL_FLAG_SUFFICIENT:
		j.ignoreBackslash()
		return nil
	default:
		return fmt.Errorf("invalid jaas file: login module control flag not specified")
	}
}

func (j *JAASParser) ParseJAASConfigurationEntry(jaasConfig string, key string) (*properties.Properties, error) {
	j.tokenizer.Init(strings.NewReader(jaasConfig))
	_, _, parsedToken, parentKey, err := j.parseConfigurationEntry(key)
	if err != nil {
		return nil, err
	}
	j.JaasOriginalConfigKeys.DisableExpansion = true
	_, _, err = j.JaasOriginalConfigKeys.Set(key+KEY_SEPARATOR+parentKey, jaasConfig)
	if err != nil {
		return nil, err
	}

	return parsedToken, nil
}

func (j *JAASParser) ConvertPropertiesToJAAS(props *properties.Properties, op string, addSecretsConfig bool) (*properties.Properties, error) {
	configKey := ""
	result := properties.NewProperties()
	result.DisableExpansion = true
	for key, value := range props.Map() {
		keys := strings.Split(key, KEY_SEPARATOR)
		configKey = keys[CLASS_ID] + KEY_SEPARATOR + keys[PARENT_ID]
		jaas, ok := j.JaasOriginalConfigKeys.Get(configKey)
		if !ok {
			return nil, fmt.Errorf("failed to convert properties to JAAS configuration")
		}
		jaas, err := j.updateJAASConfig(op, keys[KEY_ID], value, jaas, addSecretsConfig)
		if err != nil {
			return nil, err
		}
		_, _, err = j.JaasOriginalConfigKeys.Set(configKey, jaas)
		if err != nil {
			return nil, err
		}
		_, _, err = result.Set(keys[CLASS_ID], jaas)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

func (j *JAASParser) parseConfigurationEntry(prefixKey string) (int, int, *properties.Properties, string, error) {
	// Parse Parent Key
	parsedConfigs := properties.NewProperties()

	parentKey, startIndex, err := j.parseConfig('=')
	if err != nil {
		return 0, 0, nil, "", err
	}

	// Parse Control Flag
	err = j.parseControlFlag()
	if err != nil {
		return 0, 0, nil, "", err
	}

	key := ""
	for j.tokenizer.Peek() != scanner.EOF && j.tokenizer.Peek() != ';' {
		// Parse Key
		key, _, err = j.parseConfig('=')
		if err != nil {
			return 0, 0, nil, "", err
		}

		if j.tokenizer.Peek() == ' ' {
			j.WhitespaceKey = " "
		}

		// Parse =
		if j.tokenizer.Peek() == scanner.EOF || j.tokenizer.Scan() != '=' || j.tokenizer.TokenText() == "" {
			return 0, 0, nil, "", fmt.Errorf("invalid jaas file: value not specified for the key " + key)
		}

		// Parse Value
		value := ""
		value, _, err = j.parseConfig(';')
		if err != nil {
			return 0, 0, nil, "", err
		}
		newKey := prefixKey + KEY_SEPARATOR + parentKey + KEY_SEPARATOR + key
		_, _, err := parsedConfigs.Set(newKey, value)
		if err != nil {
			return 0, 0, nil, "", err
		}
		j.ignoreBackslash()
	}
	if j.tokenizer.Scan() != ';' {
		return 0, 0, nil, "", fmt.Errorf("invalid jaas file: not terminated with ';'")
	}
	endIndex := j.tokenizer.Pos().Offset

	return startIndex, endIndex, parsedConfigs, parentKey, nil
}

func (j *JAASParser) parseJAASFile(jaasConfig string) (*properties.Properties, *properties.Properties, *properties.Properties, error) {

	jaasConfigOriginalMap := properties.NewProperties()
	jaasConfigOriginalMap.DisableExpansion = true
	jaasClassConfigMap := properties.NewProperties()
	jaasClassConfigMap.DisableExpansion = true
	parsedConfigs := properties.NewProperties()
	parsedConfigs.DisableExpansion = true

	j.tokenizer.Init(strings.NewReader(jaasConfig))
	tokenClass := j.tokenizer.Peek()
	for tokenClass != scanner.EOF {
		// Parse Class Name
		classKey, start, err := j.parseConfig('{')
		if err != nil {
			return nil, nil, nil, err
		}
		// Parse {
		if j.tokenizer.Scan() != '{' {
			return nil, nil, nil, fmt.Errorf("invalid jaas file: '{' is missing")
		}

		tok := j.tokenizer.Peek()
		for tok != scanner.EOF && !j.isClosingBracket() {
			startIndex, endIndex, parsedToken, parentKey, err := j.parseConfigurationEntry(classKey)
			if err != nil {
				return nil, nil, nil, err
			}
			parsedConfigs.Merge(parsedToken)
			_, _, err = jaasConfigOriginalMap.Set(classKey+KEY_SEPARATOR+parentKey, jaasConfig[startIndex:endIndex])
			if err != nil {
				return nil, nil, nil, err
			}
			tok = j.tokenizer.Peek()
		}
		if j.tokenizer.TokenText() != "}" {
			return nil, nil, nil, fmt.Errorf("invalid jaas file: not terminated with '}'")
		}
		if j.tokenizer.Scan() != ';' {
			return nil, nil, nil, fmt.Errorf("invalid jaas file: not terminated with ';'")
		}
		end := j.tokenizer.Pos().Offset
		tokenClass = j.tokenizer.Peek()
		_, _, err = jaasClassConfigMap.Set(classKey, jaasConfig[start:end])
		if err != nil {
			return nil, nil, nil, err
		}
	}
	return parsedConfigs, jaasConfigOriginalMap, jaasClassConfigMap, nil
}
