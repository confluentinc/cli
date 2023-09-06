package store

import (
	"fmt"
	"sort"
	"strings"

	"github.com/samber/lo"

	"github.com/confluentinc/cli/v3/pkg/flink/config"
)

const emptyStringTag = "<unset>"

type UserProperties struct {
	defaultProperties map[string]string
	properties        map[string]string
}

func NewUserProperties(defaultProperties map[string]string) UserProperties {
	userProperties := UserProperties{
		defaultProperties: defaultProperties,
		properties:        map[string]string{},
	}
	userProperties.addDefaultProperties()
	return userProperties
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

// GetSqlProperties returns only the properties that should be sent when creating a statement (identified by the 'sql.' prefix)
func (p *UserProperties) GetSqlProperties() map[string]string {
	sqlProperties := map[string]string{}
	for key, value := range p.properties {
		if strings.HasPrefix(key, config.ConfigNamespaceSql) {
			sqlProperties[key] = value
		}
	}
	return sqlProperties
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
