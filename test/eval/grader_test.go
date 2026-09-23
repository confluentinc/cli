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

func TestReadCrossFieldSelections(t *testing.T) {
	home := t.TempDir()
	writeConfig(t, home, `{"current_context":"ctx","contexts":{"ctx":{
		"current_environment":"env-596",
		"environments":{"env-596":{"current_flink_compute_pool":"lfcp-123456"}},
		"kafka_cluster_context":{"kafka_environment_contexts":{"env-596":{"active_kafka":"lkc-12345"}}}
	}}}`)

	if got, err := ReadActiveKafkaCluster(home, "env-596"); err != nil || got != "lkc-12345" {
		t.Errorf("kafka: got %q err %v", got, err)
	}
	if got, err := ReadCurrentFlinkComputePool(home, "env-596"); err != nil || got != "lfcp-123456" {
		t.Errorf("flink: got %q err %v", got, err)
	}
}

func TestGlobalAPIKeyPresent(t *testing.T) {
	home := t.TempDir()
	writeConfig(t, home, `{"current_context":"ctx","contexts":{"ctx":{"global_api_keys":{"MYKEY1":{}}}}}`)

	if ok, err := GlobalAPIKeyPresent(home, "MYKEY1"); err != nil || !ok {
		t.Errorf("expected MYKEY1 present (ok=%v err=%v)", ok, err)
	}
	if ok, _ := GlobalAPIKeyPresent(home, "MYKEY2"); ok {
		t.Error("expected MYKEY2 absent")
	}
}

func TestCredsCleared(t *testing.T) {
	home := t.TempDir()
	writeConfig(t, home, `{"current_context":"","contexts":{"ctx":{}},"context_states":{"ctx":{"auth_token":""}}}`)

	if ok, err := CredsCleared(home); err != nil || !ok {
		t.Errorf("expected creds cleared (ok=%v err=%v)", ok, err)
	}

	home2 := t.TempDir()
	writeConfig(t, home2, `{"current_context":"ctx","contexts":{"ctx":{}},"context_states":{"ctx":{"auth_token":"tok"}}}`)
	if ok, _ := CredsCleared(home2); ok {
		t.Error("expected creds NOT cleared when auth_token is set")
	}
}

func TestCredsClearedErrorsWhenContextStateMissing(t *testing.T) {
	home := t.TempDir()
	writeConfig(t, home, `{"current_context":"","contexts":{"ctx":{}},"context_states":{}}`)

	_, err := CredsCleared(home)

	if err == nil {
		t.Error("expected an error when the context's state entry is missing, not a clean logout")
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
