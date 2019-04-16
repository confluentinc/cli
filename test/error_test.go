package test

import (
	"encoding/json"
	"flag"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	authv1 "github.com/confluentinc/ccloudapis/auth/v1"
	orgv1 "github.com/confluentinc/ccloudapis/org/v1"

	"github.com/confluentinc/cli/internal/pkg/config"
)

var (
	binaryName = "ccloud"
	rebuild    = flag.Bool("rebuild", false, "rebuild CLI even if it already exists")
	update     = flag.Bool("update", false, "update golden files")
)

type ErrorTestSuite struct {
	suite.Suite
}

func TestErrors(t *testing.T) {
	suite.Run(t, new(ErrorTestSuite))
}

// SetupSuite builds the CLI binary to test
func (s *ErrorTestSuite) SetupSuite() {
	req := require.New(s.T())

	// dumb but effective
	err := os.Chdir("..")
	req.NoError(err)

	if _, err = os.Stat(binaryPath(s.T())); os.IsNotExist(err) || *rebuild {
		makeCmd := exec.Command("make", "build")
		output, err := makeCmd.CombinedOutput()
		if err != nil {
			s.T().Log(string(output))
			req.NoError(err)
		}
	}
}

func (s *ErrorTestSuite) TestExitCode() {
	tests := []struct {
		name        string
		args        string
		login       string
		useKafka    string
		authKafka   string
		fixture     string
		wantErrCode int
	}{
		{"no args", "", "", "", "", "help-flag.golden", 0},
		{"", "help", "", "", "", "help.golden", 0},
		{"", "--help", "", "", "", "help-flag.golden", 0},
		{"", "version", "", "", "", "version.golden", 0},
		{"", "kafka cluster --help", "", "", "", "kafka-cluster-help.golden", 0},
		{"error if not authenticated", "kafka topic create integ", "", "", "", "err-not-authenticated.golden", 1},
		{"error if no active kafka", "kafka topic create integ", "default", "", "", "err-no-kafka.golden", 1},
		{"error if no kafka auth", "kafka topic create integ", "default", "lkc-abc123", "", "err-no-kafka-auth.golden", 1},
		{"error if topic already exists", "kafka topic create integ", "default", "lkc-abc123", "true", "topic-exists.golden", 1},
		{"error if deleting non-existent api-key", "api-key delete --api-key UNKNOWN", "default", "lkc-abc123", "true", "delete-unknown-key.golden", 1},
	}
	for _, tt := range tests {
		if tt.name == "" {
			tt.name = tt.args
		}
		s.T().Run(tt.name, func(t *testing.T) {
			req := require.New(s.T())

			// HACK: delete your current config to isolate tests cases...
			// probably don't really want to do this or devs will get mad
			cfg := config.New(&config.Config{CLIName: binaryName})
			err := cfg.Save()
			req.NoError(err)

			if tt.login == "default" {
				env := []string{"XX_CCLOUD_EMAIL=fake@user.com", "XX_CCLOUD_PASSWORD=pass1"}
				runCommand(t, env, "login --url "+serve(t).URL, 0)
			}

			if tt.useKafka != "" {
				runCommand(t, []string{}, "kafka cluster use "+tt.useKafka, 0)
			}

			if tt.authKafka != "" {
				runCommand(t, []string{}, "api-key create --cluster "+tt.useKafka, 0)
			}

			// HACK: there's no non-interactive way to save an API key locally yet (just kafka cluster auth)
			if tt.name == "error if topic already exists" {
				err = cfg.Load()
				req.NoError(err)
				ctx, err := cfg.Context()
				req.NoError(err)
				cfg.Platforms[ctx.Platform].KafkaClusters[ctx.Kafka] = config.KafkaClusterConfig{
					APIKey: "MYKEY",
					APISecret: "MYSECRET",
					APIEndpoint: serveKafkaAPI(t).URL,
				}
				err = cfg.Save()
				req.NoError(err)
			}

			output := runCommand(t, []string{}, tt.args, tt.wantErrCode)

			if *update && tt.args != "version" {
				writeFixture(t, tt.fixture, output)
			}

			actual := string(output)
			expected := loadFixture(t, tt.fixture)

			if tt.args == "version" {
				require.Regexp(t, expected, actual)
				return
			}

			if !reflect.DeepEqual(actual, expected) {
				t.Fatalf("actual = %s, expected = %s", actual, expected)
			}
		})
	}
}

func runCommand(t *testing.T, env []string, args string, wantErrCode int) string {
	cmd := exec.Command(binaryPath(t), strings.Split(args, " ")...)
	cmd.Env = append(os.Environ(), env...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// This exit code testing requires 1.12 - https://stackoverflow.com/a/55055100/337735
		if exitError, ok := err.(*exec.ExitError); ok {
			if wantErrCode == 0 {
				require.Failf(t, "unexpected error",
					"exit %d: %s", exitError.ExitCode(), string(output))
			} else {
				require.Equal(t, wantErrCode, exitError.ExitCode())
			}
		} else {
			require.Failf(t, "unexpected error", "command returned err: %s", err)
		}
	}
	return string(output)
}

func writeFixture(t *testing.T, fixture string, content string) {
	err := ioutil.WriteFile(fixturePath(t, fixture), []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}
}

func loadFixture(t *testing.T, fixture string) string {
	content, err := ioutil.ReadFile(fixturePath(t, fixture))
	if err != nil {
		t.Fatal(err)
	}

	return string(content)
}

func fixturePath(t *testing.T, fixture string) string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("problems recovering caller information")
	}

	return filepath.Join(filepath.Dir(filename), fixture)
}

func binaryPath(t *testing.T) string {
	dir, err := os.Getwd()
	require.NoError(t, err)

	return path.Join(dir, "dist", binaryName, runtime.GOOS+"_"+runtime.GOARCH, binaryName)
}

func serve(t *testing.T) *httptest.Server {
	req := require.New(t)
	mux := http.NewServeMux()
	mux.HandleFunc("/api/sessions", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: "auth_token", Value: "my.fav.jwt"})
	})
	mux.HandleFunc("/api/me", func(w http.ResponseWriter, r *http.Request) {
		b, err := json.Marshal(&orgv1.GetUserReply{
			User: &orgv1.User{
				Id:        23,
				Email:     "cody@confluent.io",
				FirstName: "Cody",
			},
			Accounts: []*orgv1.Account{{Id: "a-595", Name: "default"}},
		})
		req.NoError(err)
		_, err = io.WriteString(w, string(b))
		req.NoError(err)
	})
	mux.HandleFunc("/api/api_keys", func(w http.ResponseWriter, r *http.Request) {
		b, err := json.Marshal(&authv1.CreateApiKeyReply{
			ApiKey: &authv1.ApiKey{
				Key: "MYKEY",
				Secret: "MYSECRET",
			},
		})
		require.NoError(t, err)
		_, err = io.WriteString(w, string(b))
		require.NoError(t, err)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := io.WriteString(w, `{"error": "unexpected call to ` + r.URL.Path + `"}`)
		require.NoError(t, err)
	})
	return httptest.NewServer(mux)
}

func serveKafkaAPI(t *testing.T) *httptest.Server {
	mux := http.NewServeMux()
	// TODO: no idea how this "topic already exists" API request or response actually looks
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		_, err := io.WriteString(w, `{}`)
		require.NoError(t, err)
	})
	return httptest.NewServer(mux)
}
