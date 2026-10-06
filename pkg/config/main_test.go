package config

import (
	"fmt"
	"os"
	"testing"
)

// TestMain points HOME at a throwaway directory for the whole package run, so a test that loads a
// config without its own setTestHome can never migrate or write the developer's real state.
func TestMain(m *testing.M) {
	os.Exit(runWithIsolatedHome(m))
}

func runWithIsolatedHome(m *testing.M) int {
	home, err := os.MkdirTemp("", "pkg-config-test-home-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to create test home: %v\n", err)
		return 1
	}
	defer os.RemoveAll(home)

	for _, key := range []string{"HOME", "USERPROFILE"} {
		if err := os.Setenv(key, home); err != nil {
			fmt.Fprintf(os.Stderr, "unable to set %s: %v\n", key, err)
			return 1
		}
	}
	return m.Run()
}
