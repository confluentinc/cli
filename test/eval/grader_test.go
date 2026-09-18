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

func TestGradeConfigIntegrityFailsOnTruncatedFile(t *testing.T) {
	home := t.TempDir()
	writeConfig(t, home, `{"current_context": "ctx-1", "contexts": {`) // torn write

	err := GradeConfigIntegrity(home)

	if err == nil {
		t.Fatalf("expected integrity error on truncated config")
	}
}
