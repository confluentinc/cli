//go:build cuj_test

package cuj

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"text/template"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

var cujBin = "test/cuj/bin/confluent"

// CUJ is a critical user journey: a plain-English task handed to an AI agent,
// judged afterwards by checks against live Confluent Cloud state.
type CUJ struct {
	// Name identifies the CUJ in run artifacts (test/cuj/runs/<Name>/...).
	Name string
	// Task is the only thing the agent sees. It is a text/template rendered with
	// {{.cluster_id}}, {{.environment_id}}, plus any vars passed to runCUJ.
	Task string
	// Checks run after the agent finishes, each as its own subtest.
	Checks []CUJCheck
}

// CUJCheck is one pass/fail criterion. It should judge from live state, using
// the transcript only for things the agent must report.
type CUJCheck struct {
	ID    string
	Check func(res agentResult) (pass bool, detail string)
}

type checkResult struct {
	ID     string `json:"id"`
	Pass   bool   `json:"pass"`
	Detail string `json:"detail"`
}

// runRecord is written to run.json for every run.
type runRecord struct {
	CUJ           string        `json:"cuj"`
	ClaudeVersion string        `json:"claude_version"`
	Model         string        `json:"model"`
	EnvironmentID string        `json:"environment_id"`
	ClusterID     string        `json:"cluster_id"`
	StartedAt     string        `json:"started_at"`
	DurationS     int           `json:"duration_s"`
	ExitCode      int           `json:"exit_code"`
	TimedOut      bool          `json:"timed_out"`
	Passed        int           `json:"passed"`
	Total         int           `json:"total"`
	Checks        []checkResult `json:"checks"`
}

// CUJTestSuite runs agent-driven CUJ tests against real Confluent Cloud.
type CUJTestSuite struct {
	suite.Suite
	binPath       string
	baseHome      string // login and selected environment/cluster, copied for each CUJ
	home          string // isolated HOME for every confluent process in the current CUJ
	wrapperDir    string // holds the confluent wrapper Claude finds on PATH
	claudePath    string
	claudeVersion string
	model         string
	envID         string
	clusterID     string
}

func TestCUJ(t *testing.T) {
	if os.Getenv("CLI_CUJ_TEST") != "1" {
		t.Skip("Skipping CUJ test. Set CLI_CUJ_TEST=1 to run.")
	}
	suite.Run(t, new(CUJTestSuite))
}

func (s *CUJTestSuite) SetupSuite() {
	req := require.New(s.T())

	// Navigate to repo root (test/cuj -> repo root is ../..)
	req.NoError(os.Chdir("../.."))
	dir, err := os.Getwd()
	req.NoError(err)

	bin := cujBin
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	s.binPath = filepath.Join(dir, bin)
	_, err = os.Stat(s.binPath)
	req.NoError(err, "CLI binary not found at %s — run `make build-for-cuj-test` first", s.binPath)

	s.setupClaude()
	s.setupBaseHome()
	s.checkOrganization()
	s.usePinnedCluster()
	s.T().Logf("claude=%s cluster=%s environment=%s", s.claudeVersion, s.clusterID, s.envID)

	// The task promises the agent an environment and cluster are already selected.
	// This is saved in the base HOME that each CUJ's HOME is copied from.
	out, err := s.confluent("environment", "use", s.envID)
	req.NoError(err, out)
	out, err = s.confluent("kafka", "cluster", "use", s.clusterID)
	req.NoError(err, out)
}

// usePinnedCluster reads the pinned environment and Kafka cluster from
// CLI_CUJ_ENVIRONMENT_ID and CLI_CUJ_CLUSTER_ID, and checks the cluster is UP.
// Every run uses the same long-lived cluster, so results are comparable.
func (s *CUJTestSuite) usePinnedCluster() {
	t := s.T()
	s.envID = os.Getenv("CLI_CUJ_ENVIRONMENT_ID")
	s.clusterID = os.Getenv("CLI_CUJ_CLUSTER_ID")
	require.True(t, s.envID != "" && s.clusterID != "",
		"set CLI_CUJ_ENVIRONMENT_ID and CLI_CUJ_CLUSTER_ID to the pinned environment and Kafka cluster in the CUJ organization")
	var cluster struct {
		Status string `json:"status"`
	}
	require.NoError(t, s.confluentJSON(&cluster, "kafka", "cluster", "describe", s.clusterID, "--environment", s.envID))
	require.Equal(t, "UP", cluster.Status, "pinned cluster %s is not UP", s.clusterID)
}

// runCUJ renders the task, runs the agent, runs each check as a subtest, and
// writes task.md, the agent's output, and run.json to test/cuj/runs/.
func (s *CUJTestSuite) runCUJ(t *testing.T, cuj CUJ, vars map[string]string) {
	t.Helper()

	data := map[string]string{"cluster_id": s.clusterID, "environment_id": s.envID}
	for k, v := range vars {
		data[k] = v
	}
	task := renderTask(t, cuj.Task, data)

	startedAt := time.Now()
	runDir, err := filepath.Abs(filepath.Join("test", "cuj", "runs", cuj.Name, startedAt.Format("20060102-150405")))
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(runDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(runDir, "task.md"), []byte(task), 0o644))

	res := s.runAgent(t, task, runDir)

	record := runRecord{
		CUJ:           cuj.Name,
		ClaudeVersion: s.claudeVersion,
		Model:         s.model,
		EnvironmentID: s.envID,
		ClusterID:     s.clusterID,
		StartedAt:     startedAt.Format(time.RFC3339),
		DurationS:     res.DurationS,
		ExitCode:      res.ExitCode,
		TimedOut:      res.TimedOut,
		Total:         len(cuj.Checks),
	}
	for _, c := range cuj.Checks {
		t.Run(c.ID, func(t *testing.T) {
			pass, detail := c.Check(res)
			record.Checks = append(record.Checks, checkResult{ID: c.ID, Pass: pass, Detail: detail})
			if pass {
				record.Passed++
				t.Log(detail)
			} else {
				t.Error(detail)
			}
		})
	}

	out, err := json.MarshalIndent(record, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(runDir, "run.json"), out, 0o644))
	t.Logf("score %d/%d (timed_out=%t) — artifacts in %s", record.Passed, record.Total, record.TimedOut, runDir)
}

func renderTask(t *testing.T, text string, data map[string]string) string {
	t.Helper()
	tmpl, err := template.New("task").Option("missingkey=error").Parse(text)
	require.NoError(t, err)
	var buf bytes.Buffer
	require.NoError(t, tmpl.Execute(&buf, data))
	return buf.String()
}
