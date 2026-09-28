package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/confluentinc/cli/v4/pkg/errors"
)

// settingsFile is settings.json: machine-managed preferences.
type settingsFile struct {
	DisableFeatureFlags       bool                 `json:"disable_feature_flags"`
	DisablePlugins            bool                 `json:"disable_plugins"`
	DisablePluginsOnceWindows bool                 `json:"disable_plugins_once_windows,omitempty"`
	DisableUpdateCheck        bool                 `json:"disable_update_check"`
	EnableColor               bool                 `json:"enable_color"`
	Platforms                 map[string]*Platform `json:"platforms,omitempty"`
	LocalPorts                *LocalPorts          `json:"local_ports,omitempty"`
	// Deprecated
	DisablePluginsOnce bool `json:"disable_plugins_once,omitempty"`
}

// contextsFile is contexts.json: the secret-free selection state.
type contextsFile struct {
	CurrentContext   string                      `json:"current_context"`
	Contexts         map[string]*Context         `json:"contexts,omitempty"`
	Credentials      map[string]*Credential      `json:"credentials,omitempty"`
	ContextStates    map[string]*ContextState    `json:"context_states,omitempty"`
	SavedCredentials map[string]*LoginCredential `json:"saved_credentials,omitempty"`
}

// readStoreFile reads and unmarshals path into v. A missing or zero-byte file contributes
// nothing (found is false, err is nil); malformed JSON is a hard error.
func readStoreFile(path string, v any) (bool, error) {
	input, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf(errors.UnableToReadConfigurationFileErrorMsg, path, err)
	}
	if len(input) == 0 {
		return false, nil
	}
	if err := json.Unmarshal(input, v); err != nil {
		return false, fmt.Errorf(errors.UnableToReadConfigurationFileErrorMsg, path, err)
	}
	return true, nil
}

// writeStoreFile marshals v and atomically writes it to path, creating the parent
// directory first. Shared by the config stores below and by secretStore.write.
func writeStoreFile(path string, v any) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("unable to create store directory %s: %w", dir, err)
	}

	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("unable to marshal store %s: %w", path, err)
	}

	return writeFileAtomic(path, data)
}

// loadConfigStores reads settings.json and contexts.json into c. A missing or zero-byte
// file contributes nothing; found is false only when neither file holds data. Both files
// are read and decoded before anything is applied to c, so a malformed file leaves c
// untouched rather than half-applied. Each shape is seeded from c's current values before
// decoding, so a key absent from the file (an older or hand-edited store) keeps c's
// existing value - notably New()'s defaults - instead of silently zeroing it.
func (c *Config) loadConfigStores() (bool, error) {
	// Only scalars are seeded from c (see the doc comment above); map/pointer fields stay at
	// their zero value, since seeding those would alias c's own map/struct and let
	// json.Unmarshal merge into it instead of replacing it wholesale.
	settings := &settingsFile{
		DisableFeatureFlags:       c.DisableFeatureFlags,
		DisablePlugins:            c.DisablePlugins,
		DisablePluginsOnceWindows: c.DisablePluginsOnceWindows,
		DisableUpdateCheck:        c.DisableUpdateCheck,
		EnableColor:               c.EnableColor,
		DisablePluginsOnce:        c.DisablePluginsOnce,
	}
	settingsFound, err := readStoreFile(SettingsFilename(), settings)
	if err != nil {
		return false, err
	}

	contexts := &contextsFile{
		CurrentContext: c.CurrentContext,
	}
	contextsFound, err := readStoreFile(ContextsFilename(), contexts)
	if err != nil {
		return false, err
	}

	if settingsFound {
		c.DisableFeatureFlags = settings.DisableFeatureFlags
		c.DisablePlugins = settings.DisablePlugins
		c.DisablePluginsOnceWindows = settings.DisablePluginsOnceWindows
		c.DisableUpdateCheck = settings.DisableUpdateCheck
		c.EnableColor = settings.EnableColor
		c.DisablePluginsOnce = settings.DisablePluginsOnce
		if settings.Platforms != nil {
			c.Platforms = settings.Platforms
		}
		if settings.LocalPorts != nil {
			c.LocalPorts = settings.LocalPorts
		}
	}

	if contextsFound {
		c.CurrentContext = contexts.CurrentContext
		if contexts.Contexts != nil {
			c.Contexts = contexts.Contexts
		}
		if contexts.Credentials != nil {
			c.Credentials = contexts.Credentials
		}
		if contexts.ContextStates != nil {
			c.ContextStates = contexts.ContextStates
		}
		if contexts.SavedCredentials != nil {
			c.SavedCredentials = contexts.SavedCredentials
		}
	}

	return settingsFound || contextsFound, nil
}

// saveConfigStores writes c's persisted fields to settings.json, then contexts.json. Settings
// is written first: a crash between the two writes then leaves only an orphan platform on
// disk, never a context pointing at a platform that was never written (Validate() rejects
// the latter).
func (c *Config) saveConfigStores() error {
	settings := &settingsFile{
		DisableFeatureFlags:       c.DisableFeatureFlags,
		DisablePlugins:            c.DisablePlugins,
		DisablePluginsOnceWindows: c.DisablePluginsOnceWindows,
		DisableUpdateCheck:        c.DisableUpdateCheck,
		EnableColor:               c.EnableColor,
		Platforms:                 c.Platforms,
		LocalPorts:                c.LocalPorts,
		DisablePluginsOnce:        c.DisablePluginsOnce,
	}
	if err := writeStoreFile(SettingsFilename(), settings); err != nil {
		return err
	}

	contexts := &contextsFile{
		CurrentContext:   c.CurrentContext,
		Contexts:         c.Contexts,
		Credentials:      c.Credentials,
		ContextStates:    c.ContextStates,
		SavedCredentials: c.SavedCredentials,
	}
	return writeStoreFile(ContextsFilename(), contexts)
}
