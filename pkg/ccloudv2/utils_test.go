package ccloudv2

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"

	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/log"
)

func TestIsCCloudURL_True(t *testing.T) {
	for _, url := range []string{
		"confluent.cloud",
		"https://confluent.cloud",
		"https://devel.cpdev.cloud/",
		"devel.cpdev.cloud",
		"stag.cpdev.cloud",
	} {
		isCCloud := IsCCloudURL(url, false)
		require.True(t, isCCloud, url+" should return true")
	}
}

func TestIsCCloudURL_False(t *testing.T) {
	for _, url := range []string{
		"example.com",
		"example.com:8090",
		"https://example.com",
	} {
		isCCloud := IsCCloudURL(url, false)
		require.False(t, isCCloud, url+" should return false")
	}
}

func TestGetServerUrl(t *testing.T) {
	m := map[string]string{
		"https://confluent.cloud":   "https://api.confluent.cloud",
		"https://devel.cpdev.cloud": "https://api.devel.cpdev.cloud",
		"https://stag.cpdev.cloud":  "https://api.stag.cpdev.cloud",
		"https://stag.cpdev.cloud/": "https://api.stag.cpdev.cloud",
	}

	for baseUrl, serverUrl := range m {
		assert.Equal(t, serverUrl, getServerUrl(baseUrl))
	}
}

func TestNewRetryableHttpClient_TestModeSkipsBackoff(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		attempts++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()
	client := NewRetryableHttpClient(&config.Config{IsTest: true}, false)

	start := time.Now()
	_, err := client.Get(server.URL)
	elapsed := time.Since(start)

	require.Error(t, err)
	require.Equal(t, 5, attempts)
	require.Less(t, elapsed, time.Second)
}

func TestToLower(t *testing.T) {
	require.Equal(t, "sasl-ssl", ToLower("SASL_SSL"))
}

func TestToUpper(t *testing.T) {
	require.Equal(t, "SASL_SSL", ToUpper("sasl-ssl"))
}

func TestExtractPageToken(t *testing.T) {
	token, err := ExtractPageToken("https://example.com/results?page_token=20")
	require.NoError(t, err)
	require.Equal(t, "20", token)
}

func TestExtractPageToken_MissingToken(t *testing.T) {
	_, err := ExtractPageToken("https://example.com/results")
	require.ErrorContains(t, err, `could not parse the value for query parameter "page_token"`)
}

// captureWarnings swaps log.CliLogger for one that writes to a buffer at WARN
// verbosity for the duration of fn, then returns everything logged.
func captureWarnings(t *testing.T, fn func()) string {
	t.Helper()

	var buf bytes.Buffer
	original := log.CliLogger
	log.CliLogger = log.New(log.WARN, &buf)
	defer func() { log.CliLogger = original }()

	fn()
	return buf.String()
}

func newTestV1Client(baseURL string) *ccloudv1.Client {
	return ccloudv1.NewClient(&ccloudv1.Params{
		BaseURL:    baseURL,
		HttpClient: ccloudv1.BaseClient,
		Logger:     log.CliLogger,
	})
}

// newTestConfig returns a mock cloud config whose Save() writes under a fresh
// temp directory rather than the real user config path.
func newTestConfig(t *testing.T) *config.Config {
	t.Helper()

	cfg := config.AuthenticatedCloudConfigMock()
	cfg.Filename = filepath.Join(t.TempDir(), "config.json")
	return cfg
}

func TestRefreshAndSave_LogsRefreshFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	cfg := newTestConfig(t)

	logs := captureWarnings(t, func() {
		refreshAndSave(cfg, newTestV1Client(server.URL))
	})

	require.Contains(t, logs, "failed to refresh session")
}

func TestRefreshAndSave_LogsSaveFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		require.NoError(t, json.NewEncoder(w).Encode(ccloudv1.AuthenticateReply{Token: "new-token"}))
	}))
	defer server.Close()

	cfg := newTestConfig(t)
	// Put a regular file where Save() needs the config directory, so its MkdirAll
	// fails with ENOTDIR. A merely-absent parent no longer forces a failure: Save()
	// now creates the directory itself for fresh-machine first runs.
	notADir := filepath.Join(t.TempDir(), "not-a-dir")
	require.NoError(t, os.WriteFile(notADir, nil, 0600))
	cfg.Filename = filepath.Join(notADir, "config.json")

	logs := captureWarnings(t, func() {
		refreshAndSave(cfg, newTestV1Client(server.URL))
	})

	require.Contains(t, logs, "failed to save config")
	require.NotContains(t, logs, "failed to refresh session")
}

func TestRefreshAndSave_LogsNothingOnSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		require.NoError(t, json.NewEncoder(w).Encode(ccloudv1.AuthenticateReply{Token: "new-token"}))
	}))
	defer server.Close()

	cfg := newTestConfig(t)

	logs := captureWarnings(t, func() {
		refreshAndSave(cfg, newTestV1Client(server.URL))
	})

	require.Empty(t, logs)
}
