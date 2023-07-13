package store

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
