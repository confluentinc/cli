//go:build eval

package scenarios

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/confluentinc/cli/v4/test/eval"
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

func TestLoginStepCarriesURL(t *testing.T) {
	if got := loginStep("http://127.0.0.1:9000"); got != "login --url http://127.0.0.1:9000" {
		t.Errorf("loginStep built %q", got)
	}
}

func TestObserveActiveEnvironmentReadsSessionHome(t *testing.T) {
	home := t.TempDir()
	writeConfig(t, home, `{"current_context":"ctx","contexts":{"ctx":{"current_environment":"env-596"}}}`)

	got, err := observeActiveEnvironment(eval.SessionResult{HomeDir: home})

	if err != nil || got != "env-596" {
		t.Errorf("observeActiveEnvironment = %q, %v; want env-596", got, err)
	}
}

func TestCreatedGlobalKeyParsesStdout(t *testing.T) {
	r := eval.SessionResult{Invocations: []eval.Invocation{
		{Command: "login --url x"},
		{Command: "api-key create --resource global", Stdout: "It may take a couple of minutes...\nAPI Key: MYKEY2\nAPI Secret: MYSECRET2\n"},
	}}
	if got := createdGlobalKey(r); got != "MYKEY2" {
		t.Errorf("parsed %q", got)
	}
}

func TestCrudEmptyKeyIsNotSilentlyOK(t *testing.T) {
	home := t.TempDir()
	writeConfig(t, home, `{"current_context":"ctx","contexts":{"ctx":{"global_api_keys":{}}}}`)
	r := eval.SessionResult{Session: 0, HomeDir: home, Invocations: []eval.Invocation{{Command: "api-key create --resource global", Stdout: "no key here", ExitCode: 0}}}

	out := eval.GradeSession(r, "", observeGlobalKeyPresent(""))

	if out.Verdict == eval.VerdictOK {
		t.Fatalf("empty parsed key must not grade OK (would mask a dropped key); got %s", out.Verdict)
	}
}
