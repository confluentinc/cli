package test

import (
	"encoding/json"
	"flag"
	"fmt"
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
	kafkav1 "github.com/confluentinc/ccloudapis/kafka/v1"
	orgv1 "github.com/confluentinc/ccloudapis/org/v1"

	"github.com/confluentinc/cli/internal/pkg/config"
)

var (
	binaryName = "ccloud"
	noRebuild  = flag.Bool("no-rebuild", false, "skip rebuilding CLI if it already exists")
	update     = flag.Bool("update", false, "update golden files")
)

// CLITest represents a test configuration
type CLITest struct {
	// Name to show in go test output; defaults to args if not set
	name string
	// The CLI command being tested; this is a string of args and flags passed to the binary
	args string
	// "default" if you need to login, or "" otherwise
	login string
	// The kafka cluster ID to "use"
	useKafka string
	// The API Key to set as Kafka credentials
	authKafka string
	// Name of a golden output fixture containing expected output
	fixture string
	// Expected exit code (e.g., 0 for success or 1 for failure)
	wantErrCode int
	// If true, don't reset the config/state between tests to enable testing CLI workflows
	workflow bool
}

// CLITestSuite is the CLI integration tests.
type CLITestSuite struct {
	suite.Suite
}

// TestCLI runs the CLI integration test suite.
func TestCLI(t *testing.T) {
	suite.Run(t, new(CLITestSuite))
}

// SetupSuite builds the CLI binary to test
func (s *CLITestSuite) SetupSuite() {
	req := require.New(s.T())

	// dumb but effective
	err := os.Chdir("..")
	req.NoError(err)

	if _, err = os.Stat(binaryPath(s.T())); os.IsNotExist(err) || !*noRebuild {
		makeCmd := exec.Command("make", "build")
		output, err := makeCmd.CombinedOutput()
		if err != nil {
			s.T().Log(string(output))
			req.NoError(err)
		}
	}
}

func (s *CLITestSuite) Test_Help() {
	tests := []CLITest{
		{name: "no args", fixture: "help-flag.golden"},
		{args: "help", fixture: "help.golden"},
		{args: "--help", fixture: "help-flag.golden"},
		{args: "version", fixture: "version.golden"},
	}
	for _, tt := range tests {
		s.runTest(tt, serve(s.T()).URL, serveKafkaAPI(s.T()).URL)
	}
}

func (s *CLITestSuite) Test_Login_UseKafka_AuthKafka_Errors() {
	tests := []CLITest{
		{
			name:    "error if not authenticated",
			args:    "kafka topic create integ",
			fixture: "err-not-authenticated.golden",
		},
		{
			name:    "error if no active kafka",
			args:    "kafka topic create integ",
			fixture: "err-no-kafka.golden",
			login:   "default",
		},
		{
			name:      "error if topic already exists",
			args:      "kafka topic create integ",
			fixture:   "topic-exists.golden",
			login:     "default",
			useKafka:  "lkc-abc123",
			authKafka: "true",
		},
		{
			name:      "error if deleting non-existent api-key",
			args:      "api-key delete UNKNOWN",
			fixture:   "delete-unknown-key.golden",
			login:     "default",
			useKafka:  "lkc-abc123",
			authKafka: "true",
		},
		{
			name:     "error if using unknown kafka",
			args:     "kafka cluster use lkc-unknown",
			fixture:  "err-use-unknown-kafka.golden",
			login:    "default",
		},
	}
	for _, tt := range tests {
		if strings.HasPrefix(tt.name, "error") {
			tt.wantErrCode = 1
		}
		s.runTest(tt, serve(s.T()).URL, serveKafkaAPI(s.T()).URL)
	}
}

func (s *CLITestSuite) runTest(tt CLITest, loginURL, kafkaAPIEndpoint string) {
	if tt.name == "" {
		tt.name = tt.args
	}
	s.T().Run(tt.name, func(t *testing.T) {
		req := require.New(t)

		if !tt.workflow {
			resetConfiguration(t)
		}

		if tt.login == "default" {
			env := []string{"XX_CCLOUD_EMAIL=fake@user.com", "XX_CCLOUD_PASSWORD=pass1"}
			runCommand(t, env, "login --url "+loginURL, 0)
		}

		if tt.useKafka != "" {
			runCommand(t, []string{}, "kafka cluster use "+tt.useKafka, 0)
		}

		if tt.authKafka != "" {
			runCommand(t, []string{}, "api-key create --cluster "+tt.useKafka, 0)
		}

		// HACK: there's no non-interactive way to save an API key locally yet (just kafka cluster auth)
		if tt.name == "error if topic already exists" {
			cfg := config.New(&config.Config{CLIName: binaryName})
			err := cfg.Load()
			req.NoError(err)
			ctx, err := cfg.Context()
			req.NoError(err)
			cfg.Platforms[ctx.Platform].KafkaClusters[ctx.Kafka] = config.KafkaClusterConfig{
				APIKey:      "MYKEY",
				APISecret:   "MYSECRET",
				APIEndpoint: kafkaAPIEndpoint,
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

func runCommand(t *testing.T, env []string, args string, wantErrCode int) string {
	_, _ = fmt.Println(binaryPath(t), args)
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

func resetConfiguration(t *testing.T) {
	// HACK: delete your current config to isolate tests cases for non-workflow tests...
	// probably don't really want to do this or devs will get mad
	cfg := config.New(&config.Config{CLIName: binaryName})
	err := cfg.Save()
	require.NoError(t, err)
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

	return filepath.Join(filepath.Dir(filename), "fixtures", "output", fixture)
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
		if r.Method == "POST" {
			b, err := json.Marshal(&authv1.CreateApiKeyReply{
				ApiKey: &authv1.ApiKey{
					Key:    "MYKEY",
					Secret: "MYSECRET",
					LogicalClusters: []*authv1.ApiKey_Cluster{
						&authv1.ApiKey_Cluster{Id: "bob"},
					},
					UserId: 23,
				},
			})
			require.NoError(t, err)
			_, err = io.WriteString(w, string(b))
			require.NoError(t, err)
		} else if r.Method == "GET" {
			b, err := json.Marshal(&authv1.GetApiKeysReply{
				ApiKeys: []*authv1.ApiKey{
					&authv1.ApiKey{
						Key:    "MYKEY",
						Secret: "MYSECRET",
						LogicalClusters: []*authv1.ApiKey_Cluster{
							&authv1.ApiKey_Cluster{Id: "bob"},
						},
						UserId: 23,
					},
					&authv1.ApiKey{
						Key:    "MYKEY2",
						Secret: "MYSECRET2",
						LogicalClusters: []*authv1.ApiKey_Cluster{
							&authv1.ApiKey_Cluster{Id: "abc"},
						},
						UserId: 23,
					},
				}})
			require.NoError(t, err)
			_, err = io.WriteString(w, string(b))
			require.NoError(t, err)
		}
	})
	mux.HandleFunc("/api/clusters/", func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.URL.Path, "/")
		id := parts[len(parts)-1]
		if id == "lkc-unknown" {
			_, err := io.WriteString(w, `{"error":{"code":404,"message":"resource not found","nested_errors":{},"details":[],"stack":null},"cluster":null}`)
			require.NoError(t, err)
			return
		}
		b, err := json.Marshal(&kafkav1.GetKafkaClusterReply{
			Cluster: &kafkav1.KafkaCluster{
				Id:          id,
				Endpoint:    "SASL_SSL://kafka-endpoint",
				ApiEndpoint: "https://kafka-api-endpoint",
			},
		})
		require.NoError(t, err)
		_, err = io.WriteString(w, string(b))
		require.NoError(t, err)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := io.WriteString(w, `{"error": "unexpected call to `+r.URL.Path+`"}`)
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
