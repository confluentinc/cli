package test

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/confluentinc/mds-sdk-go"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/confluentinc/ccloud-sdk-go"
	authv1 "github.com/confluentinc/ccloudapis/auth/v1"
	corev1 "github.com/confluentinc/ccloudapis/core/v1"
	kafkav1 "github.com/confluentinc/ccloudapis/kafka/v1"
	orgv1 "github.com/confluentinc/ccloudapis/org/v1"
	srv1 "github.com/confluentinc/ccloudapis/schemaregistry/v1"
	utilv1 "github.com/confluentinc/ccloudapis/util/v1"

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
		{name: "no args", fixture: "confluent-help-flag.golden", wantErrCode: 1},
		{args: "help", fixture: "confluent-help.golden"},
		{args: "--help", fixture: "confluent-help-flag.golden"},
		{args: "version", fixture: "confluent-version.golden"},
	}
	for _, tt := range tests {
		kafkaAPIURL := serveKafkaAPI(s.T()).URL
		s.runConfluentTest(tt, serveMds(s.T(), kafkaAPIURL).URL)
	}
}

func (s *CLITestSuite) Test_Confluent_Iam_Rolebinding_List() {
	tests := []CLITest{
		{
			name:        "confluent iam rolebinding list, no principal nor role",
			args:        "iam rolebinding list --kafka-cluster-id CID",
			fixture:     "confluent-iam-rolebinding-list-no-principal-nor-role.golden",
			login:       "default",
			wantErrCode: 1,
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --principal User:frodo",
			fixture: "confluent-iam-rolebinding-list-user.golden",
			login:   "default",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --principal User:frodo --role DeveloperRead",
			fixture: "confluent-iam-rolebinding-list-user-and-role-with-multiple-resources-from-one-group.golden",
			login:   "default",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --principal User:frodo --role DeveloperWrite",
			fixture: "confluent-iam-rolebinding-list-user-and-role-with-resources-from-multiple-groups.golden",
			login:   "default",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --principal User:frodo --role SecurityAdmin",
			fixture: "confluent-iam-rolebinding-list-user-and-role-with-cluster-resource.golden",
			login:   "default",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --principal User:frodo --role SystemAdmin",
			fixture: "confluent-iam-rolebinding-list-user-and-role-with-no-matches.golden",
			login:   "default",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --principal Group:hobbits --role DeveloperRead",
			fixture: "confluent-iam-rolebinding-list-group-and-role-with-multiple-resources.golden",
			login:   "default",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --principal Group:hobbits --role DeveloperWrite",
			fixture: "confluent-iam-rolebinding-list-group-and-role-with-one-resource.golden",
			login:   "default",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --principal Group:hobbits --role SecurityAdmin",
			fixture: "confluent-iam-rolebinding-list-group-and-role-with-no-matches.golden",
			login:   "default",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --role DeveloperRead",
			fixture: "confluent-iam-rolebinding-list-role-with-multiple-bindings-to-one-group.golden",
			login:   "default",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --role DeveloperWrite",
			fixture: "confluent-iam-rolebinding-list-role-with-bindings-to-multiple-groups.golden",
			login:   "default",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --role SecurityAdmin",
			fixture: "confluent-iam-rolebinding-list-role-on-cluster-bound-to-user.golden",
			login:   "default",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --role SystemAdmin",
			fixture: "confluent-iam-rolebinding-list-role-with-no-matches.golden",
			login:   "default",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --role DeveloperRead --resource Topic:food",
			fixture: "confluent-iam-rolebinding-list-role-and-resource-with-exact-match.golden",
			login:   "default",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --role DeveloperRead --resource Topic:shire-parties",
			fixture: "confluent-iam-rolebinding-list-role-and-resource-with-no-match.golden",
			login:   "default",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --role DeveloperWrite --resource Topic:shire-parties",
			fixture: "confluent-iam-rolebinding-list-role-and-resource-with-prefix-match.golden",
			login:   "default",
		},
	}
	for _, tt := range tests {
		kafkaAPIURL := serveKafkaAPI(s.T()).URL
		s.runConfluentTest(tt, serveMds(s.T(), kafkaAPIURL).URL)
	}
}

func (s *CLITestSuite) Test_Ccloud_Help() {
	tests := []CLITest{
		{name: "no args", fixture: "help-flag.golden", wantErrCode: 1},
		{args: "help", fixture: "help.golden"},
		{args: "--help", fixture: "help-flag.golden"},
		{args: "version", fixture: "version.golden"},
	}
	for _, tt := range tests {
		kafkaAPIURL := serveKafkaAPI(s.T()).URL
		s.runCcloudTest(tt, serve(s.T(), kafkaAPIURL).URL, kafkaAPIURL)
	}
}

func assertUserAgent(t *testing.T, expected string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		require.Regexp(t, expected, r.Header.Get("User-Agent"))
	}
}

func (s *CLITestSuite) Test_UserAgent() {
	t := s.T()

	checkUserAgent := func(t *testing.T, expected string) string {
		kafkaApiRouter := http.NewServeMux()
		kafkaApiRouter.HandleFunc("/", assertUserAgent(t, expected))
		kafkaApiServer := httptest.NewServer(kafkaApiRouter)
		cloudRouter := http.NewServeMux()
		cloudRouter.HandleFunc("/api/sessions", compose(assertUserAgent(t, expected), handleLogin(t)))
		cloudRouter.HandleFunc("/api/me", compose(assertUserAgent(t, expected), handleMe(t)))
		cloudRouter.HandleFunc("/api/check_email/", compose(assertUserAgent(t, expected), handleCheckEmail(t)))
		cloudRouter.HandleFunc("/api/clusters/", compose(assertUserAgent(t, expected), handleKafkaClusterGetListDelete(t, kafkaApiServer.URL)))
		return httptest.NewServer(cloudRouter).URL
	}

	serverURL := checkUserAgent(t, fmt.Sprintf("Confluent-Cloud-CLI/v(?:[0-9]\\.?){3}([^ ]*) \\(https://confluent.cloud; support@confluent.io\\) "+
		"ccloud-sdk-go/%s \\(%s/%s; go[^ ]*\\)", ccloud.SDKVersion, runtime.GOOS, runtime.GOARCH))
	env := []string{"XX_CCLOUD_EMAIL=valid@user.com", "XX_CCLOUD_PASSWORD=pass1"}

	t.Run("ccloud login", func(tt *testing.T) {
		_ = runCommand(tt, "ccloud", env, "login --url "+serverURL, 0)
	})
	t.Run("ccloud cluster list", func(tt *testing.T) {
		_ = runCommand(tt, "ccloud", env, "kafka cluster list", 0)
	})
	t.Run("ccloud topic list", func(tt *testing.T) {
		_ = runCommand(tt, "ccloud", env, "kafka topic list --cluster lkc-abc123", 0)
	})
}

func (s *CLITestSuite) Test_Ccloud_Errors() {
	t := s.T()
	type errorer interface {
		GetError() *corev1.Error
	}
	serveErrors := func(t *testing.T) string {
		req := require.New(t)
		write := func(w http.ResponseWriter, resp proto.Message) {
			if r, ok := resp.(errorer); ok {
				w.WriteHeader(int(r.GetError().Code))
			}
			b, err := utilv1.MarshalJSONToBytes(resp)
			req.NoError(err)
			_, err = io.WriteString(w, string(b))
			req.NoError(err)
		}
		router := http.NewServeMux()
		router.HandleFunc("/api/sessions", handleLogin(t))
		router.HandleFunc("/api/me", handleMe(t))
		router.HandleFunc("/api/check_email/", handleCheckEmail(t))
		router.HandleFunc("/api/clusters", func(w http.ResponseWriter, r *http.Request) {
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
		server := httptest.NewServer(router)
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
		output := runCommand(tt, "ccloud", env, "login --url "+loginURL, 0)
		require.Equal(tt, "Logged in as expired@user.com\nUsing environment a-595 (\"default\")\n", output)

		output = runCommand(tt, "ccloud", []string{}, "kafka cluster list", 1)
		require.Equal(tt, "Error: Your session has expired. Please login again.\n", output)
	})

	t.Run("malformed token", func(tt *testing.T) {
		loginURL := serveErrors(tt)
		env := []string{"XX_CCLOUD_EMAIL=malformed@user.com", "XX_CCLOUD_PASSWORD=pass1"}
		output := runCommand(tt, "ccloud", env, "login --url "+loginURL, 0)
		require.Equal(tt, "Logged in as malformed@user.com\nUsing environment a-595 (\"default\")\n", output)

		output = runCommand(tt, "ccloud", []string{}, "kafka cluster list", 1)
		require.Equal(tt, "Error: Your auth token has been corrupted. Please login again.\n", output)
	})

	t.Run("invalid jwt", func(tt *testing.T) {
		loginURL := serveErrors(tt)
		env := []string{"XX_CCLOUD_EMAIL=invalid@user.com", "XX_CCLOUD_PASSWORD=pass1"}
		output := runCommand(tt, "ccloud", env, "login --url "+loginURL, 0)
		require.Equal(tt, "Logged in as invalid@user.com\nUsing environment a-595 (\"default\")\n", output)

		output = runCommand(tt, "ccloud", []string{}, "kafka cluster list", 1)
		require.Equal(tt, "Error: Your auth token has been corrupted. Please login again.\n", output)
	})
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
			output := runCommand(t, "ccloud", []string{}, "api-key create --resource "+tt.useKafka, 0)
			if *debug {
				fmt.Println(output)
			}
			// HACK: we don't have scriptable output yet so we parse it from the table
			key := strings.TrimSpace(strings.Split(strings.Split(output, "\n")[2], "|")[2])
			output = runCommand(t, "ccloud", []string{}, fmt.Sprintf("api-key use %s --resource %s", key, tt.useKafka), 0)
			if *debug {
				fmt.Println(output)
			}
		}

		output := runCommand(t, "ccloud", tt.env, tt.args, tt.wantErrCode)
		if *debug {
			fmt.Println(output)
		}

		if *update && tt.args != "version" {
			if strings.HasPrefix(tt.args, "kafka cluster create") {
				re := regexp.MustCompile("https?://127.0.0.1:[0-9]+")
				output = re.ReplaceAllString(output, "http://127.0.0.1:12345")
			}
			writeFixture(t, tt.fixture, output)
		}

		actual := string(output)
		expected := loadFixture(t, tt.fixture)

		if tt.args == "version" {
			require.Regexp(t, expected, actual)
			return
		} else if strings.HasPrefix(tt.args, "kafka cluster create") {
			fmt.Println(tt.args, actual)
			re := regexp.MustCompile("https?://127.0.0.1:[0-9]+")
			actual = re.ReplaceAllString(actual, "http://127.0.0.1:12345")
		}

		if !reflect.DeepEqual(actual, expected) {
			t.Fatalf("actual = %s, expected = %s", actual, expected)
		}

		if tt.wantFunc != nil {
			tt.wantFunc(t)
		}
	})
}

func (s *CLITestSuite) runConfluentTest(tt CLITest, loginURL string) {
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

		if tt.login == "default" {
			env := []string{"XX_CONFLUENT_USERNAME=fake@user.com", "XX_CONFLUENT_PASSWORD=pass1"}
			output := runCommand(t, "confluent", env, "login --url "+loginURL, 0)
			if *debug {
				fmt.Println(output)
			}
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
				require.Equal(t, wantErrCode, exitError.ExitCode(), string(output))
			}
		} else {
			require.Failf(t, "unexpected error", "command returned err: %s", err)
		}
	} else {
		require.Equal(t, wantErrCode, 0)
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

func serveMds(t *testing.T, mdsURL string) *httptest.Server {
	req := require.New(t)
	router := http.NewServeMux()
	router.HandleFunc("/security/1.0/authenticate", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/json")
		reply := &mds.AuthenticationResponse{
			AuthToken:"eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJPbmxpbmUgSldUIEJ1aWxkZXIiLCJpYXQiOjE1NjE2NjA4NTcsImV4cCI6MjUzMzg2MDM4NDU3LCJhdWQiOiJ3d3cuZXhhbXBsZS5jb20iLCJzdWIiOiJqcm9ja2V0QGV4YW1wbGUuY29tIn0.G6IgrFm5i0mN7Lz9tkZQ2tZvuZ2U7HKnvxMuZAooPmE",
			TokenType:"dunno",
			ExpiresIn:9999999999,
		}
		b, err := json.Marshal(&reply)
		req.NoError(err)
		_, err = io.WriteString(w, string(b))
		req.NoError(err)
	})
	routesAndReplies := map[string]string {
		"/security/1.0/principals/User:frodo/groups": `[
                       "hobbits",
                       "ringBearers"]`,
		"/security/1.0/principals/User:frodo/roleNames": `[
                       "DeveloperRead",
                       "DeveloperWrite",
                       "SecurityAdmin"]`,
		"/security/1.0/principals/User:frodo/roles/DeveloperRead/resources": `[]`,
		"/security/1.0/principals/User:frodo/roles/DeveloperWrite/resources": `[]`,
		"/security/1.0/principals/User:frodo/roles/SecurityAdmin/resources": `[]`,
		"/security/1.0/principals/Group:hobbits/roles/DeveloperRead/resources": `[
                       {"resourceType":"Topic","name":"drink","patternType":"LITERAL"},
                       {"resourceType":"Topic","name":"food","patternType":"LITERAL"}]`,
		"/security/1.0/principals/Group:hobbits/roles/DeveloperWrite/resources": `[
                       {"resourceType":"Topic","name":"shire-","patternType":"PREFIXED"}]`,
		"/security/1.0/principals/Group:hobbits/roles/SecurityAdmin/resources": `[]`,
		"/security/1.0/principals/Group:ringBearers/roles/DeveloperRead/resources": `[]`,
		"/security/1.0/principals/Group:ringBearers/roles/DeveloperWrite/resources": `[
                       {"resourceType":"Topic","name":"ring-","patternType":"PREFIXED"}]`,
		"/security/1.0/principals/Group:ringBearers/roles/SecurityAdmin/resources": `[]`,
		"/security/1.0/lookup/principal/User:frodo/resources": `{
                       "Group:hobbits":{
                               "DeveloperWrite":[
                                       {"resourceType":"Topic","name":"shire-","patternType":"PREFIXED"}],
                               "DeveloperRead":[
                                       {"resourceType":"Topic","name":"drink","patternType":"LITERAL"},
                                       {"resourceType":"Topic","name":"food","patternType":"LITERAL"}]},
                       "Group:ringBearers":{
                               "DeveloperWrite":[
                                       {"resourceType":"Topic","name":"ring-","patternType":"PREFIXED"}]},
                       "User:frodo":{
                               "SecurityAdmin": []}}`,
		"/security/1.0/lookup/principal/Group:hobbits/resources": `{
                       "Group:hobbits":{
                               "DeveloperWrite":[
                                       {"resourceType":"Topic","name":"shire-","patternType":"PREFIXED"}],
                               "DeveloperRead":[
                                       {"resourceType":"Topic","name":"drink","patternType":"LITERAL"},
                                       {"resourceType":"Topic","name":"food","patternType":"LITERAL"}]}}`,
		"/security/1.0/lookup/role/DeveloperRead": `["Group:hobbits"]`,
		"/security/1.0/lookup/role/DeveloperWrite": `["Group:hobbits","Group:ringBearers"]`,
		"/security/1.0/lookup/role/SecurityAdmin": `["User:frodo"]`,
		"/security/1.0/lookup/role/SystemAdmin": `[]`,
		"/security/1.0/lookup/role/DeveloperRead/resource/Topic/name/food": `["Group:hobbits"]`,
		"/security/1.0/lookup/role/DeveloperRead/resource/Topic/name/shire-parties": `[]`,
		"/security/1.0/lookup/role/DeveloperWrite/resource/Topic/name/shire-parties": `["Group:hobbits"]`,
		"/security/1.0/roles/DeveloperRead": `{
                       "name":"DeveloperRead",
                       "accessPolicy":{
                               "scopeType":"Resource",
                               "allowedOperations":[
                                       {"resourceType":"Cluster","operations":[]},
                                       {"resourceType":"TransactionalId","operations":["Describe"]},
                                       {"resourceType":"Group","operations":["Read","Describe"]},
                                       {"resourceType":"Subject","operations":["Read","ReadCompatibility"]},
                                       {"resourceType":"Connector","operations":["ReadStatus","ReadConfig"]},
                                       {"resourceType":"Topic","operations":["Read","Describe"]}]}}`,
		"/security/1.0/roles/DeveloperWrite": `{
                       "name":"DeveloperWrite",
                       "accessPolicy":{
                               "scopeType":"Resource",
                               "allowedOperations":[
                                       {"resourceType":"Subject","operations":["Write"]},
                                       {"resourceType":"Group","operations":[]},
                                       {"resourceType":"Topic","operations":["Write","Describe"]},
                                       {"resourceType":"Cluster","operations":["IdempotentWrite"]},
                                       {"resourceType":"KsqlCluster","operations":["Contribute"]},
                                       {"resourceType":"Connector","operations":["ReadStatus","Configure"]},
                                       {"resourceType":"TransactionalId","operations":["Write","Describe"]}]}}`,
		"/security/1.0/roles/SecurityAdmin": `{
                       "name":"SecurityAdmin",
                       "accessPolicy":{
                               "scopeType":"Cluster",
                               "allowedOperations":[
                                       {"resourceType":"All","operations":["DescribeAccess"]}]}}`,
		"/security/1.0/roles/SystemAdmin": `{
                       "name":"SystemAdmin",
                       "accessPolicy":{
                               "scopeType":"Cluster",
                               "allowedOperations":[
                                       {"resourceType":"All","operations":["All"]}]}}`,
	}
	for route, reply := range routesAndReplies {
		s := reply
		router.HandleFunc(route, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/json")
			_, err := io.WriteString(w, s)
			req.NoError(err)
		})
	}
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := io.WriteString(w, `{"error": {"message": "unexpected call to `+r.URL.Path+`"}}`)
		require.NoError(t, err)
	})
	return httptest.NewServer(router)
}

func serve(t *testing.T, kafkaAPIURL string) *httptest.Server {
	router := http.NewServeMux()
	router.HandleFunc("/api/sessions", handleLogin(t))
	router.HandleFunc("/api/check_email/", handleCheckEmail(t))
	router.HandleFunc("/api/me", handleMe(t))
	router.HandleFunc("/api/api_keys", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			req := &authv1.CreateApiKeyRequest{}
			err := utilv1.UnmarshalJSON(r.Body, req)
			require.NoError(t, err)
			require.NotEmpty(t, req.ApiKey.AccountId)
			apiKey := req.ApiKey
			apiKey.Id = int32(KEY_INDEX)
			apiKey.Key = fmt.Sprintf("MYKEY%d", KEY_INDEX)
			apiKey.Secret = fmt.Sprintf("MYSECRET%d", KEY_INDEX)
			apiKey.UserId = 23
			KEY_INDEX++
			KEY_STORE[apiKey.Id] = apiKey
			b, err := utilv1.MarshalJSONToBytes(&authv1.CreateApiKeyReply{ApiKey: apiKey})
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
			b, err := utilv1.MarshalJSONToBytes(&authv1.GetApiKeysReply{ApiKeys: apiKeys})
			require.NoError(t, err)
			_, err = io.WriteString(w, string(b))
			require.NoError(t, err)
		}
	})
	router.HandleFunc("/api/accounts", func(w http.ResponseWriter, r *http.Request) {
		b, err := utilv1.MarshalJSONToBytes(&orgv1.ListAccountsReply{Accounts: []*orgv1.Account{
			{Id: "a-595", Name: "default"}, {Id: "not-595", Name: "other"},
		}})
		require.NoError(t, err)
		_, err = io.WriteString(w, string(b))
		require.NoError(t, err)
	})
	router.HandleFunc("/api/clusters/", handleKafkaClusterGetListDelete(t, kafkaAPIURL))
	router.HandleFunc("/api/clusters", handleKafkaClusterCreate(t, kafkaAPIURL))
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := io.WriteString(w, `{"error": {"message": "unexpected call to `+r.URL.Path+`"}}`)
		require.NoError(t, err)
	})
	router.HandleFunc("/api/schema_registries/", func(w http.ResponseWriter, r *http.Request) {
		srCluster := &srv1.SchemaRegistryCluster{
			Id:        "lsrc-1",
			AccountId: "23",
			Name:      "account schema-registry",
			Endpoint:  "SASL_SSL://sr-endpoint",
		}
		b, err := utilv1.MarshalJSONToBytes(&srv1.GetSchemaRegistryClusterReply{
			Cluster: srCluster,
		})
		require.NoError(t, err)
		_, err = io.WriteString(w, string(b))
		require.NoError(t, err)
	})
	return httptest.NewServer(router)
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

func handleLogin(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		req := require.New(t)
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
			http.SetCookie(w, &http.Cookie{Name: "auth_token", Value: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJPbmxpbmUgSldUIEJ1aWxkZXIiLCJpYXQiOjE1MzAxMjQ4NTcsImV4cCI6MTUzMDAzODQ1NywiYXVkIjoid3d3LmV4YW1wbGUuY29tIiwic3ViIjoianJvY2tldEBleGFtcGxlLmNvbSJ9.Y2ui08GPxxuV9edXUBq-JKr1VPpMSnhjSFySczCby7Y"})
		case "malformed@user.com":
			http.SetCookie(w, &http.Cookie{Name: "auth_token", Value: "malformed"})
		case "invalid@user.com":
			http.SetCookie(w, &http.Cookie{Name: "auth_token", Value: "invalid"})
		default:
			http.SetCookie(w, &http.Cookie{Name: "auth_token", Value: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJPbmxpbmUgSldUIEJ1aWxkZXIiLCJpYXQiOjE1NjE2NjA4NTcsImV4cCI6MjUzMzg2MDM4NDU3LCJhdWQiOiJ3d3cuZXhhbXBsZS5jb20iLCJzdWIiOiJqcm9ja2V0QGV4YW1wbGUuY29tIn0.G6IgrFm5i0mN7Lz9tkZQ2tZvuZ2U7HKnvxMuZAooPmE"})
		}
	}
}

func handleMe(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		b, err := utilv1.MarshalJSONToBytes(&orgv1.GetUserReply{
			User: &orgv1.User{
				Id:        23,
				Email:     "cody@confluent.io",
				FirstName: "Cody",
			},
			Accounts: []*orgv1.Account{{Id: "a-595", Name: "default"}},
		})
		require.NoError(t, err)
		_, err = io.WriteString(w, string(b))
		require.NoError(t, err)
	}
}

func handleCheckEmail(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		req := require.New(t)
		email := strings.Replace(r.URL.String(), "/api/check_email/", "", 1)
		reply := &orgv1.GetUserReply{}
		switch email {
		case "cody@confluent.io":
			reply.User = &orgv1.User{
				Email: "cody@confluent.io",
			}
		}
		b, err := utilv1.MarshalJSONToBytes(reply)
		req.NoError(err)
		_, err = io.WriteString(w, string(b))
		req.NoError(err)
	}
}

func handleKafkaClusterGetListDelete(t *testing.T, kafkaAPIURL string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.URL.Path, "/")
		id := parts[len(parts)-1]
		if id == "lkc-unknown" {
			_, err := io.WriteString(w, `{"error":{"code":404,"message":"resource not found","nested_errors":{},"details":[],"stack":null},"cluster":null}`)
			require.NoError(t, err)
			return
		}
		if r.Method == "DELETE" {
			w.WriteHeader(http.StatusNoContent)
			return
		} else {
			// this is in the body of delete requests
			require.NotEmpty(t, r.URL.Query().Get("account_id"))
		}
		// Now return the KafkaCluster with updated ApiEndpoint
		b, err := utilv1.MarshalJSONToBytes(&kafkav1.GetKafkaClusterReply{
			Cluster: &kafkav1.KafkaCluster{
				Id:              id,
				Name:            "kafka-cluster",
				NetworkIngress:  100,
				NetworkEgress:   100,
				Storage:         500,
				ServiceProvider: "aws",
				Region:          "us-west-2",
				Endpoint:        "SASL_SSL://kafka-endpoint",
				ApiEndpoint:     kafkaAPIURL,
			},
		})
		require.NoError(t, err)
		_, err = io.WriteString(w, string(b))
		require.NoError(t, err)
	}
}

func handleKafkaClusterCreate(t *testing.T, kafkaAPIURL string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		b, err := utilv1.MarshalJSONToBytes(&kafkav1.GetKafkaClusterReply{
			Cluster: &kafkav1.KafkaCluster{
				Id:              "lkc-def963",
				Name:            "kafka-cluster",
				NetworkIngress:  100,
				NetworkEgress:   100,
				Storage:         500,
				ServiceProvider: "aws",
				Region:          "us-west-2",
				Endpoint:        "SASL_SSL://kafka-endpoint",
				ApiEndpoint:     kafkaAPIURL,
			},
		})
		require.NoError(t, err)
		_, err = io.WriteString(w, string(b))
		require.NoError(t, err)
	}
}

func compose(funcs ...func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		for _, f := range funcs {
			f(w, r)
		}
	}
}
