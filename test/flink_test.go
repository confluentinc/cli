package test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bradleyjkemp/cupaloy/v2"
	"github.com/stretchr/testify/require"
)

const (
	vhsTapeInputFile    = "test/fixtures/input/flink/flinkshell.tape"
	vhsTapeOutputTxt    = "flinkshell.txt"
	vhsTapeOutputMp4    = "flinkshell.mp4"
	fixtureOutputFolder = "test/fixtures/output/flink/shell"
)

var vhsFrameSeparator = strings.Repeat("─", 80)

func (s *CLITestSuite) TestFlinkComputePool() {
	tests := []CLITest{
		{args: "flink compute-pool create my-compute-pool --cloud aws --region us-west-2", fixture: "flink/compute-pool/create.golden"},
		{args: "flink compute-pool describe lfcp-123456", fixture: "flink/compute-pool/describe.golden"},
		{args: "flink compute-pool list", fixture: "flink/compute-pool/list.golden"},
		{args: "flink compute-pool list --region us-west-2", fixture: "flink/compute-pool/list-region.golden"},
		{args: "flink compute-pool update lfcp-123456 --max-cfu 5", fixture: "flink/compute-pool/update.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestFlinkComputePoolDelete() {
	tests := []CLITest{
		{args: "flink compute-pool delete lfcp-123456 --force", fixture: "flink/compute-pool/delete.golden"},
		{args: "flink compute-pool delete lfcp-123456 lfcp-222222", input: "n\n", fixture: "flink/compute-pool/delete-multiple-refuse.golden"},
		{args: "flink compute-pool delete lfcp-123456 lfcp-222222", input: "y\n", fixture: "flink/compute-pool/delete-multiple-success.golden"},
		{args: "flink compute-pool delete lfcp-123456 lfcp-654321", fixture: "flink/compute-pool/delete-multiple-fail.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestFlinkComputePoolUnset() {
	tests := []CLITest{
		{args: "flink compute-pool unset", login: "cloud", fixture: "flink/compute-pool/unset-before-use.golden"},
		{args: "flink compute-pool use lfcp-123456", login: "cloud", fixture: "flink/compute-pool/use-before-unset.golden"},
		{args: "flink compute-pool unset", login: "cloud", fixture: "flink/compute-pool/unset.golden"},
	}

	for _, test := range tests {
		test.workflow = true
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestFlinkComputePoolUse() {
	tests := []CLITest{
		{args: "flink compute-pool use lfcp-999999", login: "cloud", fixture: "flink/compute-pool/use-fail.golden", exitCode: 1},
		{args: "flink compute-pool use lfcp-123456", login: "cloud", fixture: "flink/compute-pool/use.golden"},
		{args: "flink compute-pool describe", fixture: "flink/compute-pool/describe-after-use.golden"},
		{args: "flink compute-pool list", fixture: "flink/compute-pool/list-after-use.golden"},
		{args: "flink compute-pool update --max-cfu 5", fixture: "flink/compute-pool/update-after-use.golden"},
	}

	for _, test := range tests {
		test.workflow = true
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestFlinkRegion() {
	tests := []CLITest{
		{args: "flink region use --cloud aws --region eu-west-1", fixture: "flink/region/use.golden"},
		{args: "flink region list", fixture: "flink/region/list.golden"},
		{args: "flink region list -o json", fixture: "flink/region/list-json.golden"},
		{args: "flink region list --cloud aws", fixture: "flink/region/list-cloud.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		test.workflow = true
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestFlinkStatement() {
	tests := []CLITest{
		{args: "flink statement delete my-statement --force --cloud aws --region eu-west-1", fixture: "flink/statement/delete.golden"},
		{args: "flink statement list --cloud aws --region eu-west-1", fixture: "flink/statement/list.golden"},
		{args: "flink statement list --cloud aws --region eu-west-1 -o yaml", fixture: "flink/statement/list-yaml.golden"},
		{args: "flink statement list --cloud aws --region eu-west-1 --status completed", fixture: "flink/statement/list-completed.golden"},
		{args: "flink statement list --cloud aws --region eu-west-1 --status pending", fixture: "flink/statement/list-pending.golden"},
		{args: "flink statement describe my-statement --cloud aws --region eu-west-1", fixture: "flink/statement/describe.golden"},
		{args: "flink statement describe my-statement --cloud aws --region eu-west-1 -o yaml", fixture: "flink/statement/describe-yaml.golden"},
		{args: "flink statement stop my-statement --region eu-west-1 --cloud aws", fixture: "flink/statement/stop.golden"},
		{args: "flink statement exception list my-statement --cloud aws --region eu-west-1", fixture: "flink/statement/exception/list.golden"},
		{args: "flink statement exception list my-statement --cloud aws --region eu-west-1 -o yaml", fixture: "flink/statement/exception/list-yaml.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestFlinkStatementCreate() {
	tests := []CLITest{
		{args: `flink statement create my-statement --sql "INSERT * INTO table;" --compute-pool lfcp-123456 --service-account sa-123456`, fixture: "flink/statement/create.golden"},
		{args: `flink statement create my-statement --sql "INSERT * INTO table;" --compute-pool lfcp-123456`, fixture: "flink/statement/create-service-account-warning.golden"},
		{args: `flink statement create my-statement --sql "INSERT * INTO table;" --compute-pool lfcp-123456 --service-account sa-123456 --wait`, fixture: "flink/statement/create-wait.golden"},
		{args: `flink statement create --sql "INSERT * INTO table;" --compute-pool lfcp-123456 --service-account sa-123456 -o yaml`, fixture: "flink/statement/create-no-name-yaml.golden", regex: true},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestFlink_Autocomplete() {
	tests := []CLITest{
		{args: `__complete flink compute-pool create my-compute-pool --cloud ""`, fixture: "flink/compute-pool/create-cloud-autocomplete.golden"},
		{args: `__complete flink compute-pool create my-compute-pool --cloud aws --region ""`, fixture: "flink/compute-pool/create-region-autocomplete.golden"},
		{args: `__complete flink compute-pool delete ""`, fixture: "flink/compute-pool/delete-autocomplete.golden"},
		{args: `__complete flink compute-pool list --region ""`, fixture: "flink/compute-pool/list-region-autocomplete.golden"},
		{args: `__complete flink statement create my-statement --database ""`, fixture: "flink/statement/create-database-autocomplete.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestFlinkShell() {
	// HACK: empty test to log in, before running the actual integration test with charm/VHS
	test := CLITest{
		login: "cloud",
	}
	s.runIntegrationTest(test)

	dir, err := os.Getwd()
	require.NoError(s.T(), err)

	cmd := exec.Command("vhs", filepath.Join(dir, vhsTapeInputFile))
	defer cleanup()

	out, err := cmd.CombinedOutput()
	require.NoError(s.T(), err, fmt.Sprintf("%s error: %v", string(out), err))

	content, err := os.ReadFile(vhsTapeOutputTxt)
	require.NoError(s.T(), err)

	// take snapshot of last frame
	lastFrame := getLastFrame(string(content))
	snapshotConfig := cupaloy.New(
		cupaloy.SnapshotSubdirectory(filepath.Join(dir, fixtureOutputFolder)),
		// update snapshot if update flag was set
		cupaloy.ShouldUpdate(func() bool {
			return *update
		}),
	)
	require.NoError(s.T(), snapshotConfig.SnapshotWithName("shell.golden", lastFrame), fmt.Sprintf("Got: %s", string(content)))
}

// deletes the tmp files generated by VHS
func cleanup() {
	// ignore errors in case files weren't generated
	_ = os.Remove(vhsTapeOutputTxt)
	_ = os.Remove(vhsTapeOutputMp4)
}

func getLastFrame(s string) string {
	lastFrameEnd := strings.LastIndex(s, vhsFrameSeparator)
	if lastFrameEnd == -1 {
		return ""
	}
	lastFrameStart := strings.LastIndex(s[:lastFrameEnd], vhsFrameSeparator)
	if lastFrameStart == -1 {
		return ""
	}
	return strings.TrimSpace(s[lastFrameStart+len(vhsFrameSeparator) : lastFrameEnd])
}

func TestGetLastFrame(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "",
			expected: "",
		},
		{
			input:    fmt.Sprintf("%s\ntest", vhsFrameSeparator),
			expected: "",
		},
		{
			input:    fmt.Sprintf("test\n%s", vhsFrameSeparator),
			expected: "",
		},
		{
			input:    fmt.Sprintf("%stest%s", vhsFrameSeparator, vhsFrameSeparator),
			expected: "test",
		},
		{
			input:    fmt.Sprintf("%s\ntest\n%s", vhsFrameSeparator, vhsFrameSeparator),
			expected: "test",
		},
	}

	for _, test := range tests {
		require.Equal(t, test.expected, getLastFrame(test.input))
	}
}
