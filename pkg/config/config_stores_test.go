package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"
	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/v4/pkg/errors"
)

// readConfigurationFileErrorPrefix is the literal text preceding the "%s" path placeholder
// in errors.UnableToReadConfigurationFileErrorMsg, derived rather than hardcoded so it stays
// coupled to that constant.
var readConfigurationFileErrorPrefix = strings.SplitN(errors.UnableToReadConfigurationFileErrorMsg, "%s", 2)[0]

// jsonTags returns, for every exported field of t whose json tag is not "-", the full tag
// string (name plus options such as omitempty) keyed by the tag's name - or by the field's Go
// name when the tag omits one (e.g. a bare ",omitempty"), since that's what encoding/json
// itself persists under.
func jsonTags(t reflect.Type) map[string]string {
	tags := map[string]string{}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}
		tag, ok := field.Tag.Lookup("json")
		if !ok || tag == "-" {
			continue
		}
		name := strings.Split(tag, ",")[0]
		if name == "" {
			name = field.Name
		}
		tags[name] = tag
	}
	return tags
}

func TestConfigStores_PartitionCoversEveryPersistedField(t *testing.T) {
	persisted := jsonTags(reflect.TypeOf(Config{}))
	settings := jsonTags(reflect.TypeOf(settingsFile{}))
	contexts := jsonTags(reflect.TypeOf(contextsFile{}))

	union := map[string]string{}
	for name, tag := range settings {
		union[name] = tag
	}
	for name, tag := range contexts {
		_, dup := union[name]
		require.False(t, dup, "%q appears in both settingsFile and contextsFile", name)
		union[name] = tag
	}

	require.Equal(t, persisted, union, "settingsFile+contextsFile tags must cover exactly Config's persisted fields, verbatim")
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
	c.DisableFeatureFlags = true
	c.DisablePlugins = true
	c.DisablePluginsOnceWindows = true
	c.DisablePluginsOnce = true
	c.EnableColor = false
	c.DisableUpdateCheck = true
	c.Platforms[platform.Name] = platform
	c.LocalPorts = &LocalPorts{KafkaRestPort: "8082"}
	c.Credentials[credential.Name] = credential
	c.CurrentContext = ctx.Name
	c.Contexts[ctx.Name] = ctx
	c.ContextStates[ctx.Name] = &ContextState{Auth: &AuthConfig{
		User:         &ccloudv1.User{Id: 123, Email: "test-user@example.com"},
		Organization: &ccloudv1.Organization{Id: 456, ResourceId: "org-abc123"},
	}}
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
	require.Equal(t, c.DisablePluginsOnce, reloaded.DisablePluginsOnce)
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

// TestConfigStores_MissingKeyKeepsDefault is the regression test for the bug where decoding
// settings.json onto a fresh zero-value settingsFile clobbered New()'s defaults (EnableColor)
// for any key the file happened to omit.
func TestConfigStores_MissingKeyKeepsDefault(t *testing.T) {
	setTestHome(t, t.TempDir())
	require.NoError(t, os.MkdirAll(filepath.Dir(SettingsFilename()), 0700))
	require.NoError(t, os.WriteFile(SettingsFilename(), []byte(`{"disable_update_check":true}`), 0600))

	c := New()
	found, err := c.loadConfigStores()

	require.NoError(t, err)
	require.True(t, found)
	require.True(t, c.DisableUpdateCheck)
	require.True(t, c.EnableColor, "a key absent from settings.json must keep New()'s default, not zero out")
}

func TestConfigStores_MissingFiles(t *testing.T) {
	setTestHome(t, t.TempDir())

	c := New()
	found, err := c.loadConfigStores()
	require.NoError(t, err)
	require.False(t, found)
	requireNonNilMaps(t, c)

	require.NoError(t, os.MkdirAll(filepath.Dir(SettingsFilename()), 0700))
	require.NoError(t, os.WriteFile(SettingsFilename(), []byte(`{"enable_color":true}`), 0600))

	c2 := New()
	found2, err := c2.loadConfigStores()
	require.NoError(t, err)
	require.True(t, found2)
	require.True(t, c2.EnableColor)
	requireNonNilMaps(t, c2)

	require.NoError(t, os.WriteFile(ContextsFilename(), []byte(``), 0600))

	c3 := New()
	found3, err := c3.loadConfigStores()
	require.NoError(t, err)
	require.True(t, found3)
	requireNonNilMaps(t, c3)
}

// TestConfigStores_OnlyZeroByteContextsFile covers a zero-byte contexts.json with no
// settings.json at all: found must still be false, since a zero-byte file contributes
// nothing, same as a missing one.
func TestConfigStores_OnlyZeroByteContextsFile(t *testing.T) {
	setTestHome(t, t.TempDir())
	require.NoError(t, os.MkdirAll(filepath.Dir(ContextsFilename()), 0700))
	require.NoError(t, os.WriteFile(ContextsFilename(), []byte(``), 0600))

	c := New()
	found, err := c.loadConfigStores()

	require.NoError(t, err)
	require.False(t, found)
	requireNonNilMaps(t, c)
}

// TestConfigStores_OnlyZeroByteSettingsFile is the settings.json mirror of
// TestConfigStores_OnlyZeroByteContextsFile.
func TestConfigStores_OnlyZeroByteSettingsFile(t *testing.T) {
	setTestHome(t, t.TempDir())
	require.NoError(t, os.MkdirAll(filepath.Dir(SettingsFilename()), 0700))
	require.NoError(t, os.WriteFile(SettingsFilename(), []byte(``), 0600))

	c := New()
	found, err := c.loadConfigStores()

	require.NoError(t, err)
	require.False(t, found)
	requireNonNilMaps(t, c)
}

// requireNonNilMaps asserts New()'s initialized empty maps survive loadConfigStores when a
// store contributes nothing for them.
func requireNonNilMaps(t *testing.T, c *Config) {
	t.Helper()
	require.NotNil(t, c.Platforms)
	require.Empty(t, c.Platforms)
	require.NotNil(t, c.Credentials)
	require.Empty(t, c.Credentials)
	require.NotNil(t, c.Contexts)
	require.Empty(t, c.Contexts)
	require.NotNil(t, c.ContextStates)
	require.Empty(t, c.ContextStates)
	require.NotNil(t, c.SavedCredentials)
	require.Empty(t, c.SavedCredentials)
}

func TestConfigStores_MalformedIsHardError(t *testing.T) {
	setTestHome(t, t.TempDir())

	require.NoError(t, os.MkdirAll(filepath.Dir(ContextsFilename()), 0700))
	require.NoError(t, os.WriteFile(ContextsFilename(), []byte("not json"), 0600))

	c := New()
	_, err := c.loadConfigStores()
	require.Error(t, err)
	require.Contains(t, err.Error(), readConfigurationFileErrorPrefix)
	require.Contains(t, err.Error(), ContextsFilename())

	setTestHome(t, t.TempDir())
	require.NoError(t, os.MkdirAll(filepath.Dir(SettingsFilename()), 0700))
	require.NoError(t, os.WriteFile(SettingsFilename(), []byte("not json"), 0600))

	c2 := New()
	_, err = c2.loadConfigStores()
	require.Error(t, err)
	require.Contains(t, err.Error(), readConfigurationFileErrorPrefix)
	require.Contains(t, err.Error(), SettingsFilename())
}

// TestConfigStores_MalformedLeavesConfigUntouched guards the read-both-before-applying-either
// ordering: a valid settings.json paired with a malformed contexts.json must not leave c
// half-applied with only the settings side written.
func TestConfigStores_MalformedLeavesConfigUntouched(t *testing.T) {
	setTestHome(t, t.TempDir())
	require.NoError(t, os.MkdirAll(filepath.Dir(SettingsFilename()), 0700))
	require.NoError(t, os.WriteFile(SettingsFilename(), []byte(`{"disable_update_check":true,"enable_color":false}`), 0600))
	require.NoError(t, os.WriteFile(ContextsFilename(), []byte("not json"), 0600))

	c := New()
	_, err := c.loadConfigStores()

	require.Error(t, err)
	require.False(t, c.DisableUpdateCheck, "settings.json must not be applied when contexts.json fails to decode")
	require.True(t, c.EnableColor, "settings.json must not be applied when contexts.json fails to decode")
}

func TestConfigStores_WritesSettingsBeforeContexts(t *testing.T) {
	setTestHome(t, t.TempDir())

	require.NoError(t, os.MkdirAll(ContextsFilename(), 0700))

	c := New()
	err := c.saveConfigStores()
	require.Error(t, err)
	require.Contains(t, err.Error(), ContextsFilename())

	settingsRaw, readErr := os.ReadFile(SettingsFilename())
	require.NoError(t, readErr, "settings.json must be written before contexts.json fails")

	var settings settingsFile
	require.NoError(t, json.Unmarshal(settingsRaw, &settings))
}
