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
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	authv1 "github.com/confluentinc/ccloudapis/auth/v1"
	corev1 "github.com/confluentinc/ccloudapis/core/v1"
	kafkav1 "github.com/confluentinc/ccloudapis/kafka/v1"
	orgv1 "github.com/confluentinc/ccloudapis/org/v1"

	"github.com/confluentinc/cli/internal/pkg/config"
)

var (
	binaryName = "ccloud"
	noRebuild  = flag.Bool("no-rebuild", false, "skip rebuilding CLI if it already exists")
	update     = flag.Bool("update", false, "update golden files")
	debug      = flag.Bool("debug", false, "enable verbose output")
)

// CLITest represents a test configuration
type CLITest struct {
	// Name to show in go test output; defaults to args if not set
	name string
	// The CLI command being tested; this is a string of args and flags passed to the binary
	args string
	// The set of environment variables to be set when the CLI is run
	env []string
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
	// An optional function that allows you to specify other calls
	wantFunc func(t *testing.T)
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

	for _, binary := range []string{"ccloud", "confluent"} {
		if _, err = os.Stat(binaryPath(s.T(), binary)); os.IsNotExist(err) || !*noRebuild {
			makeCmd := exec.Command("make", "build")
			output, err := makeCmd.CombinedOutput()
			if err != nil {
				s.T().Log(string(output))
				req.NoError(err)
			}
		}
	}
}

func (s *CLITestSuite) Test_Confluent_Help() {
	tests := []CLITest{
		{name: "no args", fixture: "confluent-help-flag.golden"},
		{args: "help", fixture: "confluent-help.golden"},
		{args: "--help", fixture: "confluent-help-flag.golden"},
		{args: "version", fixture: "confluent-version.golden"},
	}
	for _, tt := range tests {
		s.runConfluentTest(tt)
	}
}

func (s *CLITestSuite) Test_Ccloud_Errors() {
	t := s.T()
	type errorer interface {
		GetError() *corev1.Error
	}
	serveErrors := func(t *testing.T) string {
		req := require.New(t)
		write := func(w http.ResponseWriter, resp interface{}) {
			if r, ok := resp.(errorer); ok {
				w.WriteHeader(int(r.GetError().Code))
			}
			b, err := json.Marshal(resp)
			req.NoError(err)
			_, err = io.WriteString(w, string(b))
			req.NoError(err)
		}
		mux := http.NewServeMux()
		mux.HandleFunc("/api/sessions", func(w http.ResponseWriter, r *http.Request) {
			b, err := ioutil.ReadAll(r.Body)
			req.NoError(err)
			// TODO: mark AuthenticateRequest as not internal so its in CCloudAPIs
			// https://github.com/confluentinc/cc-structs/blob/ce0ea5a6670d21a4b5c4f4f6ebd3d30b44cbb9f1/kafka/flow/v1/flow.proto#L41
			auth := &struct {
				Email    string
				Password string
			}{}
			err = json.Unmarshal(b, auth)
			req.NoError(err)
			switch auth.Email {
			case "incorrect@user.com":
				w.WriteHeader(http.StatusForbidden)
			case "expired@user.com":
				http.SetCookie(w, &http.Cookie{Name: "auth_token", Value: "expired"})
			case "malformed@user.com":
				http.SetCookie(w, &http.Cookie{Name: "auth_token", Value: "malformed"})
			case "invalid@user.com":
				http.SetCookie(w, &http.Cookie{Name: "auth_token", Value: "invalid"})
			default:
				http.SetCookie(w, &http.Cookie{Name: "auth_token", Value: "good"})
			}
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
		mux.HandleFunc("/api/clusters", func(w http.ResponseWriter, r *http.Request) {
			switch r.Header.Get("Authorization") {
			// TODO: these assume the upstream doesn't change its error responses. Fragile, fragile, fragile. :(
			// https://github.com/confluentinc/cc-auth-service/blob/06db0bebb13fb64c9bc3c6e2cf0b67709b966632/jwt/token.go#L23
			case "Bearer expired":
				write(w, &kafkav1.GetKafkaClustersReply{Error: &corev1.Error{Message: "token is expired", Code: http.StatusUnauthorized}})
			case "Bearer malformed":
				write(w, &kafkav1.GetKafkaClustersReply{Error: &corev1.Error{Message: "malformed token", Code: http.StatusBadRequest}})
			case "Bearer invalid":
				// TODO: The response for an invalid token should be 4xx, not 500 (e.g., if you take a working token from devel and try in stag)
				write(w, &kafkav1.GetKafkaClustersReply{Error: &corev1.Error{Message: "Token parsing error: crypto/rsa: verification error", Code: http.StatusInternalServerError}})
			default:
				req.Fail("reached the unreachable", "auth=%s", r.Header.Get("Authorization"))
			}
		})
		server := httptest.NewServer(mux)
		return server.URL
	}

	t.Run("invalid user or pass", func(tt *testing.T) {
		loginURL := serveErrors(tt)
		env := []string{"XX_CCLOUD_EMAIL=incorrect@user.com", "XX_CCLOUD_PASSWORD=pass1"}
		output := runCommand(tt, "ccloud", env, "login --url "+loginURL, 1)
		require.Equal(tt, "Error: You have entered an incorrect username or password. Please try again.\n", output)
	})

	t.Run("expired token", func(tt *testing.T) {
		loginURL := serveErrors(tt)
		env := []string{"XX_CCLOUD_EMAIL=expired@user.com", "XX_CCLOUD_PASSWORD=pass1"}
		output := runCommand(tt, "ccloud", env, "login --url "+loginURL, 1)
		require.Equal(tt, "Logged in as expired@user.com\nUsing environment a-595 (\"default\")\n", output)

		output = runCommand(t, "ccloud", []string{}, "kafka cluster list", 1)
		require.Equal(tt, "Error: Your access to Confluent Cloud has expired. Please login again.\n", output)
	})

	t.Run("malformed token", func(tt *testing.T) {
		loginURL := serveErrors(tt)
		env := []string{"XX_CCLOUD_EMAIL=malformed@user.com", "XX_CCLOUD_PASSWORD=pass1"}
		output := runCommand(tt, "ccloud", env, "login --url "+loginURL, 1)
		require.Equal(tt, "Logged in as malformed@user.com\nUsing environment a-595 (\"default\")\n", output)

		output = runCommand(t, "ccloud", []string{}, "kafka cluster list", 1)
		require.Equal(tt, "Error: Your auth token has been corrupted. Please login again.\n", output)
	})

	t.Run("invalid jwt", func(tt *testing.T) {
		loginURL := serveErrors(tt)
		env := []string{"XX_CCLOUD_EMAIL=invalid@user.com", "XX_CCLOUD_PASSWORD=pass1"}
		output := runCommand(tt, "ccloud", env, "login --url "+loginURL, 1)
		require.Equal(tt, "Logged in as invalid@user.com\nUsing environment a-595 (\"default\")\n", output)

		output = runCommand(t, "ccloud", []string{}, "kafka cluster list", 1)
		require.Equal(tt, "Error: Your auth token has been corrupted. Please login again.\n", output)
	})
}

func (s *CLITestSuite) Test_Ccloud_Help() {
	tests := []CLITest{
		{name: "no args", fixture: "help-flag.golden"},
		{args: "help", fixture: "help.golden"},
		{args: "--help", fixture: "help-flag.golden"},
		{args: "version", fixture: "version.golden"},
	}
	for _, tt := range tests {
		kafkaAPIURL := serveKafkaAPI(s.T()).URL
		s.runCcloudTest(tt, serve(s.T(), kafkaAPIURL).URL, kafkaAPIURL)
	}
}

func (s *CLITestSuite) Test_Ccloud_Login_UseKafka_AuthKafka_Errors() {
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
			name:     "error if no api key used",
			args:     "kafka topic produce integ",
			fixture:  "err-no-api-key.golden",
			login:    "default",
			useKafka: "lkc-abc123",
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
			name:    "error if using unknown kafka",
			args:    "kafka cluster use lkc-unknown",
			fixture: "err-use-unknown-kafka.golden",
			login:   "default",
		},
	}
	for _, tt := range tests {
		if strings.HasPrefix(tt.name, "error") {
			tt.wantErrCode = 1
		}
		kafkaAPIURL := serveKafkaAPI(s.T()).URL
		s.runCcloudTest(tt, serve(s.T(), kafkaAPIURL).URL, kafkaAPIURL)
	}
}

func (s *CLITestSuite) runCcloudTest(tt CLITest, loginURL, kafkaAPIEndpoint string) {
	if tt.name == "" {
		tt.name = tt.args
	}
	if strings.HasPrefix(tt.name, "error") {
		tt.wantErrCode = 1
	}
	s.T().Run(tt.name, func(t *testing.T) {
		if !tt.workflow {
			resetConfiguration(t, "ccloud")
		}

		if tt.login == "default" {
			env := []string{"XX_CCLOUD_EMAIL=fake@user.com", "XX_CCLOUD_PASSWORD=pass1"}
			output := runCommand(t, "ccloud", env, "login --url "+loginURL, 0)
			if *debug {
				fmt.Println(output)
			}
		}

		if tt.useKafka != "" {
			output := runCommand(t, "ccloud", []string{}, "kafka cluster use "+tt.useKafka, 0)
			if *debug {
				fmt.Println(output)
			}
		}

		if tt.authKafka != "" {
			output := runCommand(t, "ccloud", []string{}, "api-key create --cluster "+tt.useKafka, 0)
			if *debug {
				fmt.Println(output)
			}
			// HACK: we don't have scriptable output yet so we parse it from the table
			key := strings.TrimSpace(strings.Split(strings.Split(output, "\n")[2], "|")[2])
			output = runCommand(t, "ccloud", []string{}, fmt.Sprintf("api-key use %s --cluster %s", key, tt.useKafka), 0)
			if *debug {
				fmt.Println(output)
			}
		}

		output := runCommand(t, "ccloud", tt.env, tt.args, tt.wantErrCode)
		if *debug {
			fmt.Println(output)
		}

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

		if tt.wantFunc != nil {
			tt.wantFunc(t)
		}
	})
}

func (s *CLITestSuite) runConfluentTest(tt CLITest) {
	if tt.name == "" {
		tt.name = tt.args
	}
	if strings.HasPrefix(tt.name, "error") {
		tt.wantErrCode = 1
	}
	s.T().Run(tt.name, func(t *testing.T) {
		if !tt.workflow {
			resetConfiguration(t, "confluent")
		}
		output := runCommand(t, "confluent", []string{}, tt.args, tt.wantErrCode)

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

func runCommand(t *testing.T, binaryName string, env []string, args string, wantErrCode int) string {
	path := binaryPath(t, binaryName)
	_, _ = fmt.Println(path, args)
	cmd := exec.Command(path, strings.Split(args, " ")...)
	cmd.Env = append(os.Environ(), env...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// This exit code testing requires 1.12 - https://stackoverflow.com/a/55055100/337735
		if exitError, ok := err.(*exec.ExitError); ok {
			if wantErrCode == 0 {
				require.Failf(t, "unexpected error",
					"exit %d: %s\n%s", exitError.ExitCode(), args, string(output))
			} else {
				require.Equal(t, wantErrCode, exitError.ExitCode())
			}
		} else {
			require.Failf(t, "unexpected error", "command returned err: %s", err)
		}
	}
	return string(output)
}

func resetConfiguration(t *testing.T, cliName string) {
	// HACK: delete your current config to isolate tests cases for non-workflow tests...
	// probably don't really want to do this or devs will get mad
	cfg := config.New(&config.Config{CLIName: cliName})
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

func binaryPath(t *testing.T, binaryName string) string {
	dir, err := os.Getwd()
	require.NoError(t, err)

	return path.Join(dir, "dist", binaryName, runtime.GOOS+"_"+runtime.GOARCH, binaryName)
}

var KEY_STORE = map[int32]*authv1.ApiKey{}
var KEY_INDEX = int32(1)

type ApiKeyList []*authv1.ApiKey

// Len is part of sort.Interface.
func (d ApiKeyList) Len() int {
	return len(d)
}

// Swap is part of sort.Interface.
func (d ApiKeyList) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}

// Less is part of sort.Interface. We use Key as the value to sort by
func (d ApiKeyList) Less(i, j int) bool {
	return d[i].Key < d[j].Key
}

func init() {
	KEY_STORE[KEY_INDEX] = &authv1.ApiKey{
		Key:    "MYKEY1",
		Secret: "MYSECRET1",
		LogicalClusters: []*authv1.ApiKey_Cluster{
			{Id: "bob"},
		},
		UserId: 12,
	}
	KEY_INDEX += 1
	KEY_STORE[KEY_INDEX] = &authv1.ApiKey{
		Key:    "MYKEY2",
		Secret: "MYSECRET2",
		LogicalClusters: []*authv1.ApiKey_Cluster{
			{Id: "abc"},
		},
		UserId: 18,
	}
	KEY_INDEX += 1
	KEY_STORE[100] = &authv1.ApiKey{
		Key:    "UIAPIKEY100",
		Secret: "UIAPISECRET100",
		LogicalClusters: []*authv1.ApiKey_Cluster{
			{Id: "lkc-cool1"},
		},
		UserId: 25,
	}
	KEY_STORE[101] = &authv1.ApiKey{
		Key:    "UIAPIKEY101",
		Secret: "UIAPISECRET101",
		LogicalClusters: []*authv1.ApiKey_Cluster{
			{Id: "lkc-other1"},
		},
		UserId: 25,
	}
	KEY_STORE[102] = &authv1.ApiKey{
		Key:    "UIAPIKEY102",
		Secret: "UIAPISECRET102",
		LogicalClusters: []*authv1.ApiKey_Cluster{
			{Id: "lksqlc-ksql1"},
		},
		UserId: 25,
	}
	KEY_STORE[103] = &authv1.ApiKey{
		Key:    "UIAPIKEY103",
		Secret: "UIAPISECRET103",
		LogicalClusters: []*authv1.ApiKey_Cluster{
			{Id: "lkc-cool1"},
		},
		UserId: 25,
	}
}

func serve(t *testing.T, kafkaAPIURL string) *httptest.Server {
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
			b, err := ioutil.ReadAll(r.Body)
			require.NoError(t, err)
			req := &authv1.CreateApiKeyRequest{}
			err = json.Unmarshal(b, req)
			require.NoError(t, err)
			require.NotEmpty(t, req.ApiKey.AccountId)
			apiKey := req.ApiKey
			apiKey.Id = int32(KEY_INDEX)
			apiKey.Key = fmt.Sprintf("MYKEY%d", KEY_INDEX)
			apiKey.Secret = fmt.Sprintf("MYSECRET%d", KEY_INDEX)
			apiKey.UserId = 23
			KEY_INDEX += 1
			KEY_STORE[apiKey.Id] = apiKey
			b, err = json.Marshal(&authv1.CreateApiKeyReply{ApiKey: apiKey})
			require.NoError(t, err)
			_, err = io.WriteString(w, string(b))
			require.NoError(t, err)
		} else if r.Method == "GET" {
			require.NotEmpty(t, r.URL.Query().Get("account_id"))
			var apiKeys []*authv1.ApiKey
			for _, a := range KEY_STORE {
				apiKeys = append(apiKeys, a)
			}
			// Return sorted data or the test output will not be stable
			sort.Sort(ApiKeyList(apiKeys))
			b, err := json.Marshal(&authv1.GetApiKeysReply{ApiKeys: apiKeys})
			require.NoError(t, err)
			_, err = io.WriteString(w, string(b))
			require.NoError(t, err)
		}
	})
	mux.HandleFunc("/api/clusters/", func(w http.ResponseWriter, r *http.Request) {
		require.NotEmpty(t, r.URL.Query().Get("account_id"))
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
				ApiEndpoint: kafkaAPIURL,
			},
		})
		require.NoError(t, err)
		_, err = io.WriteString(w, string(b))
		require.NoError(t, err)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := io.WriteString(w, `{"error": {"message": "unexpected call to `+r.URL.Path+`"}}`)
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
