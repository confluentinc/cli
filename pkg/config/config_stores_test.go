package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"

	"github.com/confluentinc/cli/v4/pkg/errors"
)

// readConfigurationFileErrorPrefix is UnableToReadConfigurationFileErrorMsg's text before the path.
var readConfigurationFileErrorPrefix = strings.SplitN(errors.UnableToReadConfigurationFileErrorMsg, "%s", 2)[0]

// jsonTags maps every persisted field of t (exported, tag not "-") to its full json tag, keyed by
// the name encoding/json persists it under.
func jsonTags(t reflect.Type) map[string]string {
	tags := map[string]string{}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}
		// encoding/json persists an untagged exported field under its Go name.
		tag := field.Tag.Get("json")
		if tag == "-" {
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

// A key absent from settings.json keeps New()'s default (e.g. EnableColor) instead of zeroing it.
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

// A zero-byte contexts.json with no settings.json contributes nothing, same as a missing one.
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

// The settings.json mirror of TestConfigStores_OnlyZeroByteContextsFile.
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

// A malformed contexts.json must not leave c half-applied with the valid settings.json side.
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

// nonDefaultPersisted returns a Config whose every persisted field (exported, json tag not "-")
// differs from New()'s value, so a field dropped from any copy in the store code shows up as a
// diff. It fails on a field kind it does not know, forcing an update when Config grows one.
func nonDefaultPersisted(t *testing.T) *Config {
	t.Helper()
	c := New()
	v := reflect.ValueOf(c).Elem()
	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		if !field.IsExported() || field.Tag.Get("json") == "-" {
			continue
		}
		fv := v.Field(i)
		switch fv.Kind() {
		case reflect.Bool:
			fv.SetBool(!fv.Bool())
		case reflect.String:
			fv.SetString("value-" + field.Name)
		case reflect.Map:
			m := reflect.MakeMap(fv.Type())
			m.SetMapIndex(reflect.ValueOf("key-"+field.Name), reflect.New(fv.Type().Elem().Elem()))
			fv.Set(m)
		case reflect.Pointer:
			fv.Set(reflect.New(fv.Type().Elem()))
		default:
			t.Fatalf("persisted field %s has unhandled kind %s", field.Name, fv.Kind())
		}
	}
	return c
}

// Every persisted field must survive saveConfigStores + loadConfigStores, and a key absent from
// both stores must keep the loading config's value (the seed copies).
func TestConfigStores_ReflectiveRoundTrip(t *testing.T) {
	setTestHome(t, t.TempDir())
	want := nonDefaultPersisted(t)
	wantJSON, err := json.Marshal(want)
	require.NoError(t, err)

	require.NoError(t, want.saveConfigStores())
	reloaded := New()
	found, err := reloaded.loadConfigStores()

	require.NoError(t, err)
	require.True(t, found)
	gotJSON, err := json.Marshal(reloaded)
	require.NoError(t, err)
	require.JSONEq(t, string(wantJSON), string(gotJSON), "a persisted field was lost in the store round trip")

	require.NoError(t, os.WriteFile(SettingsFilename(), []byte(`{}`), 0600))
	require.NoError(t, os.WriteFile(ContextsFilename(), []byte(`{}`), 0600))
	seeded := nonDefaultPersisted(t)
	found, err = seeded.loadConfigStores()
	require.NoError(t, err)
	require.True(t, found)
	seededJSON, err := json.Marshal(seeded)
	require.NoError(t, err)
	require.JSONEq(t, string(wantJSON), string(seededJSON), "a key absent from the stores must keep the config's value")
}
