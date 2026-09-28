package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// jsonTagNames returns the json tag name of every exported field of t (a struct type)
// whose tag is not "-". Embedded/unexported fields are not expected on any of these types.
func jsonTagNames(t reflect.Type) map[string]bool {
	names := map[string]bool{}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("json")
		if tag == "" || tag == "-" {
			continue
		}
		name := strings.Split(tag, ",")[0]
		names[name] = true
	}
	return names
}

func TestConfigStores_PartitionCoversEveryPersistedField(t *testing.T) {
	persisted := jsonTagNames(reflect.TypeOf(Config{}))
	settings := jsonTagNames(reflect.TypeOf(settingsFile{}))
	contexts := jsonTagNames(reflect.TypeOf(contextsFile{}))

	for name := range settings {
		require.False(t, contexts[name], "%q appears in both settingsFile and contextsFile", name)
	}

	union := map[string]bool{}
	for name := range settings {
		union[name] = true
	}
	for name := range contexts {
		union[name] = true
	}
	require.Equal(t, persisted, union, "settingsFile+contextsFile must cover exactly Config's persisted fields")
}

func TestConfigStores_RoundTrip(t *testing.T) {
	setTestHome(t, t.TempDir())

	platform := &Platform{Name: "https://example.com", Server: "https://example.com"}
	credential := &Credential{
		Name:           "api-key-abc",
		CredentialType: APIKey,
		APIKeyPair:     &APIKeyPair{Key: "abc"},
	}
	ctx := &Context{
		Name:           "ctx",
		PlatformName:   platform.Name,
		CredentialName: credential.Name,
	}

	c := New()
	c.EnableColor = false
	c.DisableUpdateCheck = true
	c.Platforms[platform.Name] = platform
	c.LocalPorts = &LocalPorts{KafkaRestPort: "8082"}
	c.Credentials[credential.Name] = credential
	c.CurrentContext = ctx.Name
	c.Contexts[ctx.Name] = ctx
	c.ContextStates[ctx.Name] = &ContextState{Auth: &AuthConfig{}}
	c.SavedCredentials["shell"] = &LoginCredential{Username: "shell-user"}

	require.NoError(t, c.saveConfigStores())

	settingsRaw, err := os.ReadFile(SettingsFilename())
	require.NoError(t, err)
	require.Contains(t, string(settingsRaw), `"enable_color"`)
	require.Contains(t, string(settingsRaw), `"platforms"`)
	require.NotContains(t, string(settingsRaw), `"contexts"`)
	require.NotContains(t, string(settingsRaw), `"credentials"`)

	contextsRaw, err := os.ReadFile(ContextsFilename())
	require.NoError(t, err)
	require.Contains(t, string(contextsRaw), `"current_context"`)
	require.Contains(t, string(contextsRaw), `"contexts"`)
	require.Contains(t, string(contextsRaw), `"context_states"`)
	require.Contains(t, string(contextsRaw), `"saved_credentials"`)
	require.NotContains(t, string(contextsRaw), `"enable_color"`)
	require.NotContains(t, string(contextsRaw), `"platforms"`)

	reloaded := New()
	found, err := reloaded.loadConfigStores()
	require.NoError(t, err)
	require.True(t, found)

	require.Equal(t, c.DisableFeatureFlags, reloaded.DisableFeatureFlags)
	require.Equal(t, c.DisablePlugins, reloaded.DisablePlugins)
	require.Equal(t, c.DisablePluginsOnceWindows, reloaded.DisablePluginsOnceWindows)
	require.Equal(t, c.DisableUpdateCheck, reloaded.DisableUpdateCheck)
	require.Equal(t, c.EnableColor, reloaded.EnableColor)
	require.Equal(t, c.Platforms, reloaded.Platforms)
	require.Equal(t, c.LocalPorts, reloaded.LocalPorts)
	require.Equal(t, c.CurrentContext, reloaded.CurrentContext)
	require.Equal(t, c.Contexts, reloaded.Contexts)
	require.Equal(t, c.Credentials, reloaded.Credentials)
	require.Equal(t, c.ContextStates, reloaded.ContextStates)
	require.Equal(t, c.SavedCredentials, reloaded.SavedCredentials)
}

func TestConfigStores_MissingFiles(t *testing.T) {
	setTestHome(t, t.TempDir())

	c := New()
	found, err := c.loadConfigStores()
	require.NoError(t, err)
	require.False(t, found)
	require.NotNil(t, c.Contexts)
	require.Empty(t, c.Contexts)

	require.NoError(t, os.MkdirAll(filepath.Dir(SettingsFilename()), 0700))
	require.NoError(t, os.WriteFile(SettingsFilename(), []byte(`{"enable_color":true}`), 0600))

	c2 := New()
	found2, err := c2.loadConfigStores()
	require.NoError(t, err)
	require.True(t, found2)
	require.True(t, c2.EnableColor)
	require.NotNil(t, c2.Contexts)
	require.Empty(t, c2.Contexts)

	require.NoError(t, os.WriteFile(ContextsFilename(), []byte(``), 0600))

	c3 := New()
	found3, err := c3.loadConfigStores()
	require.NoError(t, err)
	require.True(t, found3)
	require.NotNil(t, c3.Contexts)
	require.Empty(t, c3.Contexts)
}

func TestConfigStores_MalformedIsHardError(t *testing.T) {
	setTestHome(t, t.TempDir())

	require.NoError(t, os.MkdirAll(filepath.Dir(ContextsFilename()), 0700))
	require.NoError(t, os.WriteFile(ContextsFilename(), []byte("not json"), 0600))

	c := New()
	_, err := c.loadConfigStores()
	require.Error(t, err)
	require.Contains(t, err.Error(), ContextsFilename())

	setTestHome(t, t.TempDir())
	require.NoError(t, os.MkdirAll(filepath.Dir(SettingsFilename()), 0700))
	require.NoError(t, os.WriteFile(SettingsFilename(), []byte("not json"), 0600))

	c2 := New()
	_, err = c2.loadConfigStores()
	require.Error(t, err)
	require.Contains(t, err.Error(), SettingsFilename())
}

func TestConfigStores_WritesSettingsBeforeContexts(t *testing.T) {
	setTestHome(t, t.TempDir())

	require.NoError(t, os.MkdirAll(ContextsFilename(), 0700))

	c := New()
	err := c.saveConfigStores()
	require.Error(t, err)

	settingsRaw, readErr := os.ReadFile(SettingsFilename())
	require.NoError(t, readErr, "settings.json must be written before contexts.json fails")

	var settings settingsFile
	require.NoError(t, json.Unmarshal(settingsRaw, &settings))
}
