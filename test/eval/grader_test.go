//go:build eval

package eval

import (
	"os"
	"path/filepath"
	"testing"
)

func writeConfig(t *testing.T, home, body string) {
	t.Helper()
	dir := filepath.Join(home, ".confluent")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

const goodConfig = `{
  "current_context": "ctx-1",
  "contexts": {"ctx-1": {"current_environment": "env-596"}}
}`

func TestReadActiveEnvironmentReturnsCurrentEnv(t *testing.T) {
	home := t.TempDir()
	writeConfig(t, home, goodConfig)

	got, err := ReadActiveEnvironment(home)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "env-596" {
		t.Fatalf("active env = %q, want env-596", got)
	}
}

func TestGradeConfigIntegrityFailsOnTruncatedFile(t *testing.T) {
	home := t.TempDir()
	writeConfig(t, home, `{"current_context": "ctx-1", "contexts": {`) // torn write

	err := GradeConfigIntegrity(home)

	if err == nil {
		t.Fatalf("expected integrity error on truncated config")
	}
}
