package store

import (
	"fmt"
	"sort"

	"github.com/samber/lo"
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

func (p *UserProperties) GetProperties() map[string]string {
	return p.properties
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
	p.properties = map[string]string{}
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
