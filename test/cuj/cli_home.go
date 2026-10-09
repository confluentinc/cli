//go:build cuj_test

package cuj

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/stretchr/testify/require"
)

// The agent keeps its own login under the real HOME, so the agent process runs
// with it. Every confluent process gets an isolated HOME instead, so runs never
// read or write the developer's ~/.confluent: the harness passes it directly,
// and the agent reaches the CLI through a wrapper script first on its PATH.

// setupBaseHome creates the suite's base HOME with a CLI login to the CUJ
// organization. It logs in with CONFLUENT_CLOUD_EMAIL and CONFLUENT_CLOUD_PASSWORD,
// at CLI_CUJ_LOGIN_URL if set (for example https://stag.cpdev.cloud). With
// CLI_CUJ_CONFIG_DIR set, it copies that CLI config instead.
func (s *CUJTestSuite) setupBaseHome() {
	t := s.T()
	home, err := os.MkdirTemp("", "cli-cuj-home-*")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(home) })
	s.baseHome = home
	s.home = home

	if dir := os.Getenv("CLI_CUJ_CONFIG_DIR"); dir != "" {
		src := dir
		if filepath.Base(src) != ".confluent" {
			src = filepath.Join(src, ".confluent")
		}
		_, err := os.Stat(filepath.Join(src, "config.json"))
		require.NoError(t, err, "no CLI config found in %s (CLI_CUJ_CONFIG_DIR)", src)
		require.NoError(t, copyDir(src, filepath.Join(home, ".confluent")))
		t.Logf("using a copy of %s; the original is not modified", src)
		return
	}

	require.True(t, os.Getenv("CONFLUENT_CLOUD_EMAIL") != "" && os.Getenv("CONFLUENT_CLOUD_PASSWORD") != "",
		"set CONFLUENT_CLOUD_EMAIL and CONFLUENT_CLOUD_PASSWORD to the CUJ account, or CLI_CUJ_CONFIG_DIR to copy an existing CLI login")
	args := []string{"login"}
	if url := os.Getenv("CLI_CUJ_LOGIN_URL"); url != "" {
		args = append(args, "--url", url)
	}
	out, err := s.confluent(args...)
	require.NoError(t, err, "`confluent %s` failed:\n%s", strings.Join(args, " "), out)
	t.Log("logged in with CONFLUENT_CLOUD_EMAIL in an isolated HOME")
}

// checkOrganization stops the suite unless the CLI is logged in to the
// organization in CLI_CUJ_ORGANIZATION_ID, so an agent never runs in the wrong org.
func (s *CUJTestSuite) checkOrganization() {
	t := s.T()
	want := os.Getenv("CLI_CUJ_ORGANIZATION_ID")
	require.NotEmpty(t, want, "set CLI_CUJ_ORGANIZATION_ID to the CUJ organization's ID")
	var org struct {
		ID string `json:"id"`
	}
	require.NoError(t, s.confluentJSON(&org, "organization", "describe"))
	require.Equal(t, want, org.ID, "logged in to organization %s, but CLI_CUJ_ORGANIZATION_ID is %s", org.ID, want)
}

// SetupTest gives each CUJ a fresh copy of the base HOME, so no CUJ sees
// changes the agent made in an earlier one, and writes the agent's wrapper.
func (s *CUJTestSuite) SetupTest() {
	t := s.T()
	s.home = t.TempDir()
	require.NoError(t, copyDir(filepath.Join(s.baseHome, ".confluent"), filepath.Join(s.home, ".confluent")))

	s.wrapperDir = t.TempDir()
	script := fmt.Sprintf("#!/bin/sh\nexport HOME=%s\nexec %s \"$@\"\n", shellQuote(s.home), shellQuote(s.binPath))
	require.NoError(t, os.WriteFile(filepath.Join(s.wrapperDir, "confluent"), []byte(script), 0o755))
}

// cliCommand returns a command that runs the built CLI with the isolated HOME.
func (s *CUJTestSuite) cliCommand(ctx context.Context, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, s.binPath, args...)
	cmd.Env = withEnv(os.Environ(), "HOME="+s.home)
	return cmd
}

// withEnv returns env with each KEY=value in overrides replacing any existing KEY.
func withEnv(env []string, overrides ...string) []string {
	keys := make(map[string]bool, len(overrides))
	for _, o := range overrides {
		keys[strings.SplitN(o, "=", 2)[0]] = true
	}
	var result []string
	for _, e := range env {
		if !keys[strings.SplitN(e, "=", 2)[0]] {
			result = append(result, e)
		}
	}
	return append(result, overrides...)
}

// withoutEnv returns env without the given keys.
func withoutEnv(env []string, keys ...string) []string {
	drop := make(map[string]bool, len(keys))
	for _, k := range keys {
		drop[k] = true
	}
	var result []string
	for _, e := range env {
		if !drop[strings.SplitN(e, "=", 2)[0]] {
			result = append(result, e)
		}
	}
	return result
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// copyDir recursively copies a directory tree from src to dst, keeping file modes.
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, info.Mode())
	})
}
