package store

import (
	"fmt"
	"sort"
	"strings"

	"github.com/samber/lo"

	"github.com/confluentinc/cli/v4/pkg/flink/config"
	"github.com/confluentinc/cli/v4/pkg/flink/types"
)

const emptyStringTag = "<unset>"

type UserProperties struct {
	defaultProperties map[string]string
	properties        map[string]string
}

func getDefaultProperties(appOptions *types.ApplicationOptions) map[string]string {
	properties := map[string]string{config.KeyOutputFormat: string(config.OutputFormatStandard)}
	if appOptions.Cloud {
		properties[config.KeyServiceAccount] = appOptions.GetServiceAccountId()
		properties[config.KeyLocalTimeZone] = getLocalTimezone()
	}

	return properties
}

func getInitialProperties(appOptions *types.ApplicationOptions) map[string]string {
	properties := map[string]string{}

	if appOptions.GetEnvironmentName() != "" {
		properties[config.KeyCatalog] = appOptions.GetEnvironmentName()
	}
	if appOptions.GetDatabase() != "" {
		properties[config.KeyDatabase] = appOptions.GetDatabase()
	}

	return properties
}

func NewUserProperties(appOptions *types.ApplicationOptions) types.UserPropertiesInterface {
	return NewUserPropertiesWithDefaults(getDefaultProperties(appOptions), getInitialProperties(appOptions))
}

// NewUserPropertiesWithDefaults add initial props
func NewUserPropertiesWithDefaults(defaultProperties map[string]string, initialProperties map[string]string) types.UserPropertiesInterface {
	userProperties := UserProperties{
		defaultProperties: defaultProperties,
		properties:        initialProperties,
	}
	userProperties.addDefaultProperties()
	return &userProperties
}

func (p *UserProperties) addDefaultProperties() {
	for key, value := range p.defaultProperties {
		p.properties[key] = value
	}
}

func (p *UserProperties) Set(key, value string) {
	p.properties[key] = value
}

func (p *UserProperties) Get(key string) string {
	return p.GetOrDefault(key, "")
}

func (p *UserProperties) GetOrDefault(key, defaultValue string) string {
	val, keyExists := p.properties[key]
	if keyExists {
		return val
	}
	return defaultValue
}

func (p *UserProperties) HasKey(key string) bool {
	_, keyExists := p.properties[key]
	return keyExists
}

// GetProperties returns all properties
func (p *UserProperties) GetProperties() map[string]string {
	return p.properties
}

// GetNonLocalProperties returns only the properties that should be sent when creating a statement (identified by not having the 'client.' prefix)
func (p *UserProperties) GetNonLocalProperties() map[string]string {
	nonLocalProperties := map[string]string{}
	for key, value := range p.properties {
		if !strings.HasPrefix(key, config.NamespaceClient) {
			nonLocalProperties[key] = value
		}
	}
	return nonLocalProperties
}

// GetMaskedNonLocalProperties returns the same as GetNonLocalProperties but with sensitive values masked
func (p *UserProperties) GetMaskedNonLocalProperties() map[string]string {
	maskedProperties := p.GetNonLocalProperties()
	for key := range maskedProperties {
		if !strings.HasPrefix(key, config.NamespaceClient) {
			if hasSensitiveKey(key) {
				maskedProperties[key] = "hidden"
			}
		}
	}
	return maskedProperties
}

func (p *UserProperties) Delete(key string) {
	defaultValue, isDefaultKey := p.defaultProperties[key]
	if isDefaultKey {
		p.Set(key, defaultValue)
		return
	}

	delete(p.properties, key)
}

func (p *UserProperties) Clear() {
	clear(p.properties)
	p.addDefaultProperties()
}

func (p *UserProperties) ToSortedSlice(annotateDefaultValues bool) [][]string {
	props := lo.MapToSlice(p.properties, func(key, val string) []string {
		return p.createKeyValuePair(key, val, annotateDefaultValues)
	})
	sort.Slice(props, func(i, j int) bool {
		return props[i][0] < props[j][0]
	})
	return props
}

func (p *UserProperties) GetOutputFormat() config.OutputFormat {
	return config.OutputFormat(p.Get(config.KeyOutputFormat))
}

func (p *UserProperties) createKeyValuePair(key, val string, annotateDefaultValues bool) []string {
	defaultVal, isDefaultKey := p.defaultProperties[key]
	if annotateDefaultValues && isDefaultKey && defaultVal == val {
		return []string{key, annotateDefaultValue(val)}
	}
	return []string{key, val}
}

func annotateDefaultValue(val string) string {
	if val == "" {
		val = emptyStringTag
	}
	return fmt.Sprintf("%s (default)", val)
}
